// Package authz 基于 Casbin 的 RBAC 授权模块。
//
// 设计说明（对应 docs/rbac-casbin-design.md）：
//   - model.conf 定义 PERM 元模型（r = sub, obj, act），策略持久化到 SQLite 的 casbin_rule 表
//   - 启动时 Init(db) 完成引擎初始化 + 策略 seed（admin/viewer/operator 三个预置角色）
//   - 判断入口 HasPermission(role, obj, act)，供 HTTP 中间件 RequirePermission 调用
//   - PermissionsOf(role) 返回角色拥有的全部权限点，供 /api/auth/me 返回给前端做按钮级控制
package authz

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

//go:embed model.conf
var modelText string

// 权限点常量：obj:act 组合，单一事实源。
// 新增权限时需同步：① 此处常量 ② rolePolicies() seed ③ 路由标注 ④ 前端按钮。
const (
	PermDashboardView  = "dashboard:view"
	PermServerRead     = "server:read"
	PermLogRead        = "log:read"
	PermDeploymentRead = "deployment:read"
	PermMonitorRead    = "monitor:read"

	PermAgentRead   = "agent:read"
	PermAgentCreate = "agent:create"
	PermAgentUpdate = "agent:update"
	PermAgentDelete = "agent:delete"
	PermAgentDeploy = "agent:deploy"
	PermAgentStop   = "agent:stop"

	PermUserRead   = "user:read"
	PermUserCreate = "user:create"
	PermUserUpdate = "user:update"
	PermUserDelete = "user:delete"

	PermWebhookRead   = "webhook:read"
	PermWebhookUpdate = "webhook:update"

	PermSettingsManage = "settings:manage"
)

// 角色名常量（与 roleMetas 一致，供代码引用避免裸字符串）。
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

// allPermissions 全部权限点列表（admin 通配策略展开时使用）。
var allPermissions = []string{
	PermDashboardView, PermServerRead, PermLogRead, PermDeploymentRead, PermMonitorRead,
	PermAgentRead, PermAgentCreate, PermAgentUpdate, PermAgentDelete, PermAgentDeploy, PermAgentStop,
	PermUserRead, PermUserCreate, PermUserUpdate, PermUserDelete,
	PermWebhookRead, PermWebhookUpdate,
	PermSettingsManage,
}

// RoleInfo 角色信息 + 当前权限点（供 /api/settings/roles 返回前端渲染）。
type RoleInfo struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Builtin     bool     `json:"builtin"` // 内置角色，不可删除
	Locked      bool     `json:"locked"`  // 锁定角色，权限不可修改（admin 通配）
	Permissions []string `json:"permissions"`
}

// roleMetas 角色元数据清单（单一事实源：角色下拉/权限配置页/后端校验共用）。
var roleMetas = []RoleInfo{
	{Name: RoleAdmin, Label: "管理员", Description: "全部权限，通配策略锁定", Builtin: true, Locked: true},
	{Name: RoleOperator, Label: "运维", Description: "只读 + Agent 管理", Builtin: true, Locked: false},
	{Name: RoleViewer, Label: "观察者", Description: "只读权限", Builtin: true, Locked: false},
}

// PermissionGroup 权限点分组（前端矩阵按 obj 分组渲染，组内权限点用中文标签展示）。
// Requires 记录组内每个权限点的前置依赖（键=权限点，值=被依赖点），供前端联动勾选。
type PermissionGroup struct {
	Obj         string            `json:"obj"`
	Label       string            `json:"label"`
	Permissions []string          `json:"permissions"`
	Requires    map[string]string `json:"requires,omitempty"`
}

// permissionGroups 权限点分组清单（obj → 中文名 → 该组权限点）。
var permissionGroups = []PermissionGroup{
	{Obj: "dashboard", Label: "系统概览", Permissions: []string{PermDashboardView}},
	{Obj: "server", Label: "服务器", Permissions: []string{PermServerRead}},
	{Obj: "log", Label: "日志", Permissions: []string{PermLogRead}},
	{Obj: "deployment", Label: "部署", Permissions: []string{PermDeploymentRead}},
	{Obj: "monitor", Label: "监控", Permissions: []string{PermMonitorRead}},
	{Obj: "agent", Label: "Agent", Permissions: []string{
		PermAgentRead, PermAgentCreate, PermAgentUpdate, PermAgentDelete, PermAgentDeploy, PermAgentStop,
	}},
	{Obj: "user", Label: "用户", Permissions: []string{
		PermUserRead, PermUserCreate, PermUserUpdate, PermUserDelete,
	}},
	{Obj: "webhook", Label: "告警 Webhook", Permissions: []string{
		PermWebhookRead, PermWebhookUpdate,
	}},
	{Obj: "settings", Label: "系统设置", Permissions: []string{PermSettingsManage}},
}

// permissionRequires 权限点前置依赖声明（Prerequisite Permission）。
//
// 语义：若角色拥有 X，则必须同时拥有 requires(X)——"操作权限依赖入口权限"。
// 例：webhook:update 依赖 webhook:read（无 read 则设置页不可达，update 无意义）。
// 约束：依赖必须在同一 obj 分组内；被依赖点通常是该组的 read 权限。
// 新增权限点时：若它是"操作类"权限点，在此声明其入口依赖。
var permissionRequires = map[string]string{
	// webhook: 配置依赖查看
	PermWebhookUpdate: PermWebhookRead,
	// agent: 各操作依赖查看
	PermAgentCreate: PermAgentRead,
	PermAgentUpdate: PermAgentRead,
	PermAgentDelete: PermAgentRead,
	PermAgentDeploy: PermAgentRead,
	PermAgentStop:   PermAgentRead,
	// user: 各操作依赖查看
	PermUserCreate: PermUserRead,
	PermUserUpdate: PermUserRead,
	PermUserDelete: PermUserRead,
}

// RequiresOf 返回权限点的前置依赖（无依赖返回空串）。
func RequiresOf(perm string) string {
	return permissionRequires[perm]
}

var errNotInitialized = errors.New("authz: 未初始化，请先调用 Init")

// enforcer 全局唯一授权引擎（进程内单例，Casbin 内部线程安全）。
var enforcer *casbin.Enforcer

// db 全局数据库句柄（Init 时注入），角色 CRUD 读写 roles 表用。
var db *gorm.DB

// Init 初始化授权引擎：加载 model.conf + GORM 策略适配器（自动建 casbin_rule 表），
// 并在策略表为空时写入预置角色策略。
func Init(d *gorm.DB) error {
	db = d
	adapter, err := gormadapter.NewAdapterByDB(d)
	if err != nil {
		return err
	}
	// NewModelFromString：从内嵌字符串加载模型（NewEnforcer 第一参是文件路径，不能直接传内容）
	m, err := casbinmodel.NewModelFromString(modelText)
	if err != nil {
		return err
	}
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return err
	}
	enforcer = e
	if err := seedRoles(d); err != nil {
		return err
	}
	return seedIfNeeded(d)
}

// seedRoles 幂等 seed 内置角色到 roles 表（首次启动时写入；已存在则跳过）。
// 角色元数据以代码 roleMetas 为单一事实源，数据库仅作运行时扩展（自定义角色）。
func seedRoles(d *gorm.DB) error {
	var count int64
	if err := d.Table("roles").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	roles := make([]model.Role, 0, len(roleMetas))
	for _, meta := range roleMetas {
		roles = append(roles, model.Role{
			Name:        meta.Name,
			Label:       meta.Label,
			Description: meta.Description,
			Builtin:     meta.Builtin,
			Locked:      meta.Locked,
		})
	}
	return d.Create(&roles).Error
}

// HasPermission 判断角色 role 是否拥有对资源 obj 执行 act 的权限。
// 由 HTTP 中间件 RequirePermission 调用；权限变更后调用 Reload 即时生效。
func HasPermission(role, obj, act string) (bool, error) {
	if enforcer == nil {
		return false, errNotInitialized
	}
	return enforcer.Enforce(role, obj, act)
}

// PermissionsOf 返回角色拥有的全部权限点（"obj:act" 字符串数组，已排序）。
// admin 命中通配策略（*, *）时展开为全部权限点常量。
// 供 /api/auth/me 与登录响应返回给前端，用于按钮级显隐控制。
func PermissionsOf(role string) ([]string, error) {
	if enforcer == nil {
		return nil, errNotInitialized
	}
	policies, err := enforcer.GetFilteredPolicy(0, role)
	if err != nil {
		return nil, err
	}

	set := make(map[string]bool)
	for _, p := range policies {
		if len(p) < 3 {
			continue
		}
		obj, act := p[1], p[2]
		if obj == "*" || act == "*" {
			// 通配策略：展开为全部权限点
			for _, perm := range allPermissions {
				set[perm] = true
			}
			continue
		}
		set[obj+":"+act] = true
	}

	out := make([]string, 0, len(set))
	for perm := range set {
		out = append(out, perm)
	}
	sort.Strings(out)
	return out, nil
}

// Reload 从策略存储重新加载全部策略（权限变更后调用，无需重新登录）。
func Reload() error {
	if enforcer == nil {
		return errNotInitialized
	}
	return enforcer.LoadPolicy()
}

// ListRoles 返回全部角色及其当前权限点，供 /api/settings/roles 使用。
// 角色清单来自数据库 roles 表（内置角色由 seedRoles 写入），权限点实时从策略引擎读取（admin 通配展开全量）。
func ListRoles() ([]RoleInfo, error) {
	if enforcer == nil {
		return nil, errNotInitialized
	}
	var roles []model.Role
	if err := db.Order("builtin desc, name asc").Find(&roles).Error; err != nil {
		return nil, err
	}
	out := make([]RoleInfo, 0, len(roles))
	for _, r := range roles {
		perms, err := PermissionsOf(r.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, RoleInfo{
			Name:        r.Name,
			Label:       r.Label,
			Description: r.Description,
			Builtin:     r.Builtin,
			Locked:      r.Locked,
			Permissions: perms,
		})
	}
	return out, nil
}

// CreateRole 创建自定义角色并初始化权限（默认空权限，可后续配置）。
// 约束：name 需符合 [a-z0-9-] 且不与现有角色冲突；builtin 角色名保留不可创建。
func CreateRole(role model.Role) error {
	if enforcer == nil || db == nil {
		return errNotInitialized
	}
	if role.Name == "" || role.Label == "" {
		return errors.New("角色名称和显示名不能为空")
	}
	if !validRoleName(role.Name) {
		return errors.New("角色名称仅允许小写字母/数字/连字符，且不以连字符开头结尾")
	}
	// 内置角色名保留
	for _, meta := range roleMetas {
		if meta.Name == role.Name {
			return errors.New("内置角色名不可重复创建: " + role.Name)
		}
	}
	var count int64
	if err := db.Model(&model.Role{}).Where("name = ?", role.Name).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("角色已存在: " + role.Name)
	}
	// 兜底：清理该角色名可能残留的孤儿 Casbin 策略（DeleteRole 清理失败时遗留）。
	// 若不清理，同名角色重建后 Enforce 会命中残留策略 → 新角色凭空获得旧权限（越权）。
	// 兜底清理失败视为数据库异常，拒绝创建（fail-closed），避免带病上线。
	if _, err := enforcer.RemoveFilteredPolicy(0, role.Name); err != nil {
		return fmt.Errorf("创建角色前清理残留策略失败: %w", err)
	}
	if err := db.Create(&role).Error; err != nil {
		return err
	}
	return enforcer.LoadPolicy()
}

// UpdateRole 更新角色的显示名/描述（名称不可修改）。
// 约束：内置角色（admin/operator/viewer）仅允许改描述，不允许改显示名（保持系统一致性）。
func UpdateRole(name string, upd model.UpdateRoleRequest) error {
	if db == nil {
		return errNotInitialized
	}
	var role model.Role
	if err := db.First(&role, "name = ?", name).Error; err != nil {
		return errors.New("未知角色: " + name)
	}
	if role.Builtin && upd.Label != "" && upd.Label != role.Label {
		return errors.New("内置角色显示名不可修改")
	}
	updates := map[string]interface{}{}
	if upd.Label != "" {
		updates["label"] = upd.Label
	}
	if upd.Description != "" {
		updates["description"] = upd.Description
	}
	if len(updates) == 0 {
		return errors.New("没有可更新的字段")
	}
	return db.Model(&role).Updates(updates).Error
}

// DeleteRole 删除自定义角色并清理其 Casbin 策略。
// 约束：内置角色不可删；有用户绑定的角色不可删（需先转移用户）。
func DeleteRole(name string) error {
	if enforcer == nil || db == nil {
		return errNotInitialized
	}
	var role model.Role
	if err := db.First(&role, "name = ?", name).Error; err != nil {
		return errors.New("未知角色: " + name)
	}
	if role.Builtin {
		return errors.New("内置角色不可删除")
	}
	var userCount int64
	if err := db.Model(&model.User{}).Where("role = ?", name).Count(&userCount).Error; err != nil {
		return err
	}
	if userCount > 0 {
		return errors.New("该角色下存在用户，请先转移用户再删除")
	}
	// 先删 DB 角色记录：失败则策略未动，DB 与策略保持一致（不会出现"角色在但权限被撤"的破坏性状态）
	if err := db.Delete(&role).Error; err != nil {
		return err
	}
	// 再清理 Casbin 策略。清理失败仅记日志不回滚：
	//   ① DB 角色已删，返回错误会让调用方误以为"删除失败"（实际已删），状态更混乱
	//   ② 残留孤儿策略无害：同名角色重建时 CreateRole 兜底清理，且兜底失败会拒绝创建（fail-closed）
	if _, err := enforcer.RemoveFilteredPolicy(0, name); err != nil {
		slog.Warn("删除角色后清理策略失败", "role", name, "err", err)
	}
	return enforcer.LoadPolicy()
}

// RoleExists 判断角色是否存在（用户创建/更新时校验角色合法性用）。
// 返回 (bool, error)：数据库查询失败时返回错误，避免调用方误判角色不存在。
func RoleExists(name string) (bool, error) {
	if db == nil {
		return false, errNotInitialized
	}
	var count int64
	if err := db.Model(&model.Role{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// validRoleName 校验自定义角色名格式：小写字母/数字/连字符，长度 2-32。
func validRoleName(name string) bool {
	if len(name) < 2 || len(name) > 32 {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			continue
		}
		if c == '-' && i > 0 && i < len(name)-1 {
			continue
		}
		return false
	}
	return true
}

// PermissionGroups 返回权限点分组清单（obj 分组 + 中文标签），供前端矩阵渲染。
// 权限点定义是代码契约（§9.6 决策），不落入数据库。
func PermissionGroups() []PermissionGroup {
	// 附上各权限点的前置依赖（仅返回本组内的依赖，避免跨组引用）
	out := make([]PermissionGroup, len(permissionGroups))
	copy(out, permissionGroups)
	for i, g := range out {
		reqs := make(map[string]string)
		for _, perm := range g.Permissions {
			if req := RequiresOf(perm); req != "" {
				reqs[perm] = req
			}
		}
		if len(reqs) > 0 {
			out[i].Requires = reqs
		}
	}
	return out
}

// ValidPermission 判断权限点是否在系统清单内（防止配置页提交未注册权限点）。
func ValidPermission(perm string) bool {
	for _, p := range allPermissions {
		if p == perm {
			return true
		}
	}
	return false
}

// UpdateRolePermissions 更新角色权限并热生效（无需重新登录）。
//
// 实现：RemoveFilteredPolicy 清空该角色全部策略 → AddPolicies 重建 → LoadPolicy 重载。
// 约束：
//   - admin 为通配策略（*, *）锁定，不可修改（返回错误）
//   - 传入的权限点必须全部在系统清单内（ValidPermission 校验）
//   - 未在 roleMetas 中声明的角色名拒绝修改
func UpdateRolePermissions(role string, perms []string) error {
	if enforcer == nil {
		return errNotInitialized
	}

	// 仅允许修改已知角色（查库，兼容内置+自定义）
	var r model.Role
	if err := db.First(&r, "name = ?", role).Error; err != nil {
		return errors.New("未知角色: " + role)
	}
	if r.Locked {
		return errors.New("该角色为通配策略，权限不可修改")
	}

	// 校验权限点合法性（去重，保持传入顺序）
	seen := make(map[string]bool)
	valid := make([]string, 0, len(perms))
	for _, perm := range perms {
		if seen[perm] {
			continue
		}
		seen[perm] = true
		if !ValidPermission(perm) {
			return errors.New("未知权限点: " + perm)
		}
		valid = append(valid, perm)
	}

	// 前置依赖自动补全：若含 X 且缺 requires(X)，则补上被依赖点（隐式继承）。
	// 保证"仅 update 无 read"之类的死配置在数据层不可能存在。
	for _, perm := range valid {
		if req := RequiresOf(perm); req != "" && !seen[req] {
			seen[req] = true
			valid = append(valid, req)
		}
	}

	// 清空该角色现有策略
	if _, err := enforcer.RemoveFilteredPolicy(0, role); err != nil {
		return err
	}

	// 重建策略：权限点 "obj:act" → (role, obj, act)
	policies := make([][]string, 0, len(valid))
	for _, perm := range valid {
		obj, act := splitPermission(perm)
		policies = append(policies, []string{role, obj, act})
	}
	if len(policies) > 0 {
		if _, err := enforcer.AddPolicies(policies); err != nil {
			return err
		}
	}
	return enforcer.LoadPolicy()
}

// splitPermission 拆分 "obj:act" 为 (obj, act)。
func splitPermission(perm string) (string, string) {
	for i := 0; i < len(perm); i++ {
		if perm[i] == ':' {
			return perm[:i], perm[i+1:]
		}
	}
	return perm, ""
}

// seedIfNeeded 策略表为空时写入预置角色策略（admin/viewer/operator）。
func seedIfNeeded(db *gorm.DB) error {
	var count int64
	if err := db.Table("casbin_rule").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := enforcer.AddPolicies(rolePolicies())
	return err
}

// rolePolicies 预置角色-权限策略矩阵（与权限点清单保持一致）。
func rolePolicies() [][]string {
	return [][]string{
		// admin 全通（通配 obj/act）
		{"admin", "*", "*"},
		// viewer 只读
		{"viewer", "dashboard", "view"},
		{"viewer", "server", "read"},
		{"viewer", "log", "read"},
		{"viewer", "deployment", "read"},
		{"viewer", "monitor", "read"},
		// operator：只读 + Agent 管理
		{"operator", "dashboard", "view"},
		{"operator", "server", "read"},
		{"operator", "log", "read"},
		{"operator", "deployment", "read"},
		{"operator", "monitor", "read"},
		{"operator", "agent", "read"},
		{"operator", "agent", "create"},
		{"operator", "agent", "update"},
		{"operator", "agent", "delete"},
		{"operator", "agent", "deploy"},
		{"operator", "agent", "stop"},
	}
}
