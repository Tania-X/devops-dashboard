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
	PermWebhookTest   = "webhook:test"
)

// allPermissions 全部权限点列表（admin 通配策略展开时使用）。
var allPermissions = []string{
	PermDashboardView, PermServerRead, PermLogRead, PermDeploymentRead, PermMonitorRead,
	PermAgentRead, PermAgentCreate, PermAgentUpdate, PermAgentDelete, PermAgentDeploy, PermAgentStop,
	PermUserRead, PermUserCreate, PermUserUpdate, PermUserDelete,
	PermWebhookRead, PermWebhookUpdate, PermWebhookTest,
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
