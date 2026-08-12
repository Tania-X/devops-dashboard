package authz

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// newTestDB 创建内存 SQLite（与项目生产驱动一致：glebarez/sqlite，纯 Go）
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:authz_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试库失败: %v", err)
	}
	return db
}

// TestHasPermission 表驱动验证角色权限矩阵
func TestHasPermission(t *testing.T) {
	db := newTestDB(t)
	if err := Init(db); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	tests := []struct {
		name string
		role string
		obj  string
		act  string
		want bool
	}{
		// admin 通配全通
		{"admin 全通-用户删除", "admin", "user", "delete", true},
		{"admin 全通-任意资源", "admin", "anything", "anything", true},
		// viewer 只读
		{"viewer 可看仪表盘", "viewer", "dashboard", "view", true},
		{"viewer 可读服务器", "viewer", "server", "read", true},
		{"viewer 不可读用户", "viewer", "user", "read", false},
		{"viewer 不可管理 Agent", "viewer", "agent", "manage", false},
		// operator 中间角色
		{"operator 可部署 Agent", "operator", "agent", "deploy", true},
		{"operator 可停止 Agent", "operator", "agent", "stop", true},
		{"operator 不可读用户", "operator", "user", "read", false},
		{"operator 不可配置 Webhook", "operator", "webhook", "update", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HasPermission(tt.role, tt.obj, tt.act)
			if err != nil {
				t.Fatalf("HasPermission 出错: %v", err)
			}
			if got != tt.want {
				t.Errorf("HasPermission(%q, %q, %q) = %v, want %v", tt.role, tt.obj, tt.act, got, tt.want)
			}
		})
	}
}

// TestPermissionsOf 验证角色权限点列表展开
func TestPermissionsOf(t *testing.T) {
	db := newTestDB(t)
	if err := Init(db); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	contains := func(list []string, target string) bool {
		for _, v := range list {
			if v == target {
				return true
			}
		}
		return false
	}

	t.Run("viewer 只读权限", func(t *testing.T) {
		perms, err := PermissionsOf("viewer")
		if err != nil {
			t.Fatalf("PermissionsOf 出错: %v", err)
		}
		if !contains(perms, PermDashboardView) || !contains(perms, PermServerRead) {
			t.Errorf("viewer 应包含 dashboard:view/server:read, got %v", perms)
		}
		if contains(perms, PermUserRead) {
			t.Errorf("viewer 不应包含 user:read, got %v", perms)
		}
	})

	t.Run("operator 含 Agent 管理", func(t *testing.T) {
		perms, err := PermissionsOf("operator")
		if err != nil {
			t.Fatalf("PermissionsOf 出错: %v", err)
		}
		if !contains(perms, PermAgentDeploy) || !contains(perms, PermAgentStop) {
			t.Errorf("operator 应包含 agent:deploy/agent:stop, got %v", perms)
		}
		if contains(perms, PermUserRead) {
			t.Errorf("operator 不应包含 user:read, got %v", perms)
		}
	})

	t.Run("admin 通配展开全量", func(t *testing.T) {
		perms, err := PermissionsOf("admin")
		if err != nil {
			t.Fatalf("PermissionsOf 出错: %v", err)
		}
		if len(perms) != len(allPermissions) {
			t.Errorf("admin 权限点数量 = %d, want %d, got %v", len(perms), len(allPermissions), perms)
		}
	})
}

// TestSeedIdempotent 验证重复初始化不重复写入策略
func TestSeedIdempotent(t *testing.T) {
	db := newTestDB(t)
	if err := Init(db); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	if err := Init(db); err != nil {
		t.Fatalf("二次 Init 失败: %v", err)
	}

	var count int64
	if err := db.Table("casbin_rule").Count(&count).Error; err != nil {
		t.Fatalf("查询策略表失败: %v", err)
	}
	if count != int64(len(rolePolicies())) {
		t.Errorf("策略条数 = %d, want %d（seed 应幂等）", count, len(rolePolicies()))
	}
}

// TestListRoles 验证角色清单包含三个预置角色且权限点正确
func TestListRoles(t *testing.T) {
	db := newTestDB(t)
	if err := Init(db); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	roles, err := ListRoles()
	if err != nil {
		t.Fatalf("ListRoles 出错: %v", err)
	}
	if len(roles) != 3 {
		t.Fatalf("角色数 = %d, want 3", len(roles))
	}

	byName := make(map[string]RoleInfo)
	for _, r := range roles {
		byName[r.Name] = r
	}

	if !byName[RoleAdmin].Locked {
		t.Error("admin 应标记 Locked=true")
	}
	if byName[RoleOperator].Locked || byName[RoleViewer].Locked {
		t.Error("operator/viewer 不应标记 Locked")
	}

	contains := func(list []string, target string) bool {
		for _, v := range list {
			if v == target {
				return true
			}
		}
		return false
	}
	if !contains(byName[RoleOperator].Permissions, PermAgentDeploy) {
		t.Error("operator 权限应包含 agent:deploy")
	}
	if contains(byName[RoleViewer].Permissions, PermUserRead) {
		t.Error("viewer 权限不应包含 user:read")
	}
}

// TestPermissionGroups 验证权限点分组覆盖全部权限点
func TestPermissionGroups(t *testing.T) {
	groups := PermissionGroups()
	if len(groups) == 0 {
		t.Fatal("权限点分组不应为空")
	}

	// 组内权限点集合应与 allPermissions 完全一致
	set := make(map[string]bool)
	for _, g := range groups {
		for _, p := range g.Permissions {
			set[p] = true
		}
	}
	if len(set) != len(allPermissions) {
		t.Errorf("分组权限点数 = %d, allPermissions = %d，分组遗漏或多余", len(set), len(allPermissions))
	}
	for _, p := range allPermissions {
		if !set[p] {
			t.Errorf("权限点 %s 未出现在分组中", p)
		}
	}
}

// TestUpdateRolePermissions 验证热生效更新
func TestUpdateRolePermissions(t *testing.T) {
	db := newTestDB(t)
	if err := Init(db); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	t.Run("viewer 更新后即时生效", func(t *testing.T) {
		// viewer 原本不可读用户，更新后授予 user:read
		if ok, _ := HasPermission(RoleViewer, "user", "read"); ok {
			t.Fatal("前置条件不成立：viewer 初始不应有 user:read")
		}
		err := UpdateRolePermissions(RoleViewer, []string{
			PermDashboardView, PermServerRead, PermLogRead,
			PermDeploymentRead, PermMonitorRead, PermUserRead,
		})
		if err != nil {
			t.Fatalf("UpdateRolePermissions 出错: %v", err)
		}
		if ok, _ := HasPermission(RoleViewer, "user", "read"); !ok {
			t.Error("更新后 viewer 应拥有 user:read")
		}
		if ok, _ := HasPermission(RoleViewer, "agent", "deploy"); ok {
			t.Error("更新后 viewer 不应拥有 agent:deploy")
		}
	})

	t.Run("admin 锁定不可修改", func(t *testing.T) {
		err := UpdateRolePermissions(RoleAdmin, []string{PermDashboardView})
		if err == nil {
			t.Fatal("修改 admin 应返回错误")
		}
	})

	t.Run("非法权限点被拒绝", func(t *testing.T) {
		err := UpdateRolePermissions(RoleViewer, []string{"nonexistent:read"})
		if err == nil {
			t.Fatal("非法权限点应返回错误")
		}
	})

	t.Run("未知角色被拒绝", func(t *testing.T) {
		err := UpdateRolePermissions("hacker", []string{PermDashboardView})
		if err == nil {
			t.Fatal("未知角色应返回错误")
		}
	})

	t.Run("前置依赖自动补全:仅 update 无 read 时自动补 read", func(t *testing.T) {
		// 只提交 webhook:update（不提交 webhook:read），依赖规则应自动补全 read
		err := UpdateRolePermissions(RoleViewer, []string{PermWebhookUpdate})
		if err != nil {
			t.Fatalf("UpdateRolePermissions 出错: %v", err)
		}
		// update 本身生效
		if ok, _ := HasPermission(RoleViewer, "webhook", "update"); !ok {
			t.Error("viewer 应拥有 webhook:update")
		}
		// 依赖的 read 被自动补上（隐式继承）
		if ok, _ := HasPermission(RoleViewer, "webhook", "read"); !ok {
			t.Error("自动补全:viewer 应同时拥有 webhook:read")
		}
	})

	t.Run("前置依赖不重复:已含 read 时不重复添加", func(t *testing.T) {
		perms, err := PermissionsOf(RoleViewer)
		if err != nil {
			t.Fatalf("PermissionsOf 出错: %v", err)
		}
		// 只应出现一次 webhook:read
		count := 0
		for _, p := range perms {
			if p == PermWebhookRead {
				count++
			}
		}
		if count != 1 {
			t.Errorf("webhook:read 应恰好 1 个,实际 %d 个: %v", count, perms)
		}
	})

	t.Run("agent 操作类权限同样补全依赖", func(t *testing.T) {
		err := UpdateRolePermissions(RoleViewer, []string{PermAgentDeploy})
		if err != nil {
			t.Fatalf("UpdateRolePermissions 出错: %v", err)
		}
		if ok, _ := HasPermission(RoleViewer, "agent", "read"); !ok {
			t.Error("agent:deploy 应自动补全 agent:read")
		}
	})
}
