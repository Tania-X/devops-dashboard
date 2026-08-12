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
	"sort"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
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

// Init 初始化授权引擎：加载 model.conf + GORM 策略适配器（自动建 casbin_rule 表），
// 并在策略表为空时写入预置角色策略。
func Init(db *gorm.DB) error {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return err
	}
	// NewModelFromString：从内嵌字符串加载模型（NewEnforcer 第一参是文件路径，不能直接传内容）
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return err
	}
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return err
	}
	enforcer = e
	return seedIfNeeded(db)
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
// 角色清单来自代码元数据（roleMetas），权限点实时从策略引擎读取（admin 通配展开全量）。
func ListRoles() ([]RoleInfo, error) {
	if enforcer == nil {
		return nil, errNotInitialized
	}
	out := make([]RoleInfo, 0, len(roleMetas))
	for _, meta := range roleMetas {
		perms, err := PermissionsOf(meta.Name)
		if err != nil {
			return nil, err
		}
		info := meta
		info.Permissions = perms
		out = append(out, info)
	}
	return out, nil
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

	// 仅允许修改已知角色
	known := false
	for _, meta := range roleMetas {
		if meta.Name == role {
			known = true
			if meta.Locked {
				return errors.New("admin 角色为通配策略，权限不可修改")
			}
			break
		}
	}
	if !known {
		return errors.New("未知角色: " + role)
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
