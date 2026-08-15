package authz

import (
	"testing"

	userdomain "github.com/Tania-X/devops-dashboard/backend/internal/dashboard/user/domain"
	"github.com/Tania-X/devops-dashboard/backend/internal/model"
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
	// 角色 CRUD 需要 roles 表（与生产 AutoMigrate 对齐）
	if err := db.AutoMigrate(&model.Role{}, &userdomain.User{}); err != nil {
		t.Fatalf("迁移测试表失败: %v", err)
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

// TestRoleCRUD 角色管理 CRUD 测试（RBAC 三期）
func TestRoleCRUD(t *testing.T) {
	db := newTestDB(t)
	if err := Init(db); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	t.Run("seed 后内置角色存在", func(t *testing.T) {
		for _, name := range []string{RoleAdmin, RoleOperator, RoleViewer} {
			ok, err := RoleExists(name)
			if err != nil {
				t.Fatalf("RoleExists 出错: %v", err)
			}
			if !ok {
				t.Errorf("内置角色 %s 应存在", name)
			}
		}
	})

	t.Run("创建自定义角色", func(t *testing.T) {
		err := CreateRole(model.Role{Name: "auditor", Label: "审计员", Description: "只读+审计"})
		if err != nil {
			t.Fatalf("创建角色失败: %v", err)
		}
		ok, err := RoleExists("auditor")
		if err != nil {
			t.Fatalf("RoleExists 出错: %v", err)
		}
		if !ok {
			t.Error("auditor 应存在")
		}
	})

	t.Run("重复创建被拒绝", func(t *testing.T) {
		err := CreateRole(model.Role{Name: "auditor", Label: "审计员2"})
		if err == nil {
			t.Error("重复创建应返回错误")
		}
	})

	t.Run("内置角色名被保留", func(t *testing.T) {
		err := CreateRole(model.Role{Name: "admin", Label: "伪造"})
		if err == nil {
			t.Error("内置角色名不可重复创建")
		}
	})

	t.Run("非法角色名被拒绝", func(t *testing.T) {
		for _, bad := range []string{"AUDITOR", "a b", "-bad", "bad-", "a"} {
			if err := CreateRole(model.Role{Name: bad, Label: "x"}); err == nil {
				t.Errorf("非法角色名 %q 应被拒绝", bad)
			}
		}
	})

	t.Run("更新角色描述", func(t *testing.T) {
		err := UpdateRole("auditor", model.UpdateRoleRequest{Description: "审计日志查看"})
		if err != nil {
			t.Fatalf("更新角色失败: %v", err)
		}
		roles, _ := ListRoles()
		for _, r := range roles {
			if r.Name == "auditor" && r.Description != "审计日志查看" {
				t.Error("描述应更新")
			}
		}
	})

	t.Run("内置角色显示名不可改", func(t *testing.T) {
		err := UpdateRole(RoleViewer, model.UpdateRoleRequest{Label: "观察者X"})
		if err == nil {
			t.Error("内置角色显示名应不可修改")
		}
	})

	t.Run("有用户绑定的角色不可删", func(t *testing.T) {
		// 创建用户绑定 auditor 角色
		u := userdomain.User{ID: "u1", Username: "tester", Password: "x", Role: "auditor"}
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("创建测试用户失败: %v", err)
		}
		err := DeleteRole("auditor")
		if err == nil {
			t.Error("有用户绑定的角色删除应被拒绝")
		}
		// 删除用户后可删角色
		db.Delete(&u)
		if err := DeleteRole("auditor"); err != nil {
			t.Errorf("无用户绑定后删除应成功: %v", err)
		}
		ok, err := RoleExists("auditor")
		if err != nil {
			t.Fatalf("RoleExists 出错: %v", err)
		}
		if ok {
			t.Error("auditor 应已被删除")
		}
	})

	t.Run("内置角色不可删", func(t *testing.T) {
		err := DeleteRole(RoleViewer)
		if err == nil {
			t.Error("内置角色删除应被拒绝")
		}
	})

	t.Run("自定义角色权限可配置", func(t *testing.T) {
		if err := CreateRole(model.Role{Name: "ops2", Label: "运维2"}); err != nil {
			t.Fatalf("创建角色失败: %v", err)
		}
		if err := UpdateRolePermissions("ops2", []string{PermDashboardView}); err != nil {
			t.Fatalf("配置权限失败: %v", err)
		}
		if ok, _ := HasPermission("ops2", "dashboard", "view"); !ok {
			t.Error("ops2 应有 dashboard:view")
		}
	})

	t.Run("删除角色后 Casbin 策略同步清理", func(t *testing.T) {
		// 准备:创建角色并配置权限
		if err := CreateRole(model.Role{Name: "temp-role", Label: "临时角色"}); err != nil {
			t.Fatalf("创建角色失败: %v", err)
		}
		if err := UpdateRolePermissions("temp-role", []string{PermDashboardView, PermServerRead}); err != nil {
			t.Fatalf("配置权限失败: %v", err)
		}
		if ok, _ := HasPermission("temp-role", "dashboard", "view"); !ok {
			t.Fatal("删除前 temp-role 应有权限")
		}

		// 执行删除
		if err := DeleteRole("temp-role"); err != nil {
			t.Fatalf("删除角色失败: %v", err)
		}

		// DB 记录已删
		ok, err := RoleExists("temp-role")
		if err != nil {
			t.Fatalf("RoleExists 出错: %v", err)
		}
		if ok {
			t.Error("temp-role 应已从 DB 删除")
		}
		// 策略同步清理（角色已删，残留策略为孤儿数据；此处验证 Casbin 引擎中该角色不再有生效策略）
		policies, _ := enforcer.GetFilteredPolicy(0, "temp-role")
		if len(policies) > 0 {
			t.Errorf("删除后 temp-role 仍残留 %d 条策略", len(policies))
		}
	})

	t.Run("删除后重建同名角色:权限干净无残留(越权防护)", func(t *testing.T) {
		// 准备:创建带权限的角色
		if err := CreateRole(model.Role{Name: "rebuild-role", Label: "重建角色"}); err != nil {
			t.Fatalf("创建角色失败: %v", err)
		}
		if err := UpdateRolePermissions("rebuild-role", []string{PermDashboardView, PermServerRead}); err != nil {
			t.Fatalf("配置权限失败: %v", err)
		}

		// 删除
		if err := DeleteRole("rebuild-role"); err != nil {
			t.Fatalf("删除角色失败: %v", err)
		}

		// 模拟极端场景:直接向 casbin_rule 注入残留策略(等价于 DeleteRole 清理失败)
		if _, err := enforcer.AddPolicy("rebuild-role", "dashboard", "view"); err != nil {
			t.Fatalf("注入残留策略失败: %v", err)
		}
		if err := enforcer.LoadPolicy(); err != nil {
			t.Fatalf("LoadPolicy 失败: %v", err)
		}
		// 此时同名角色不存在但残留策略命中 → 模拟越权条件
		if ok, _ := HasPermission("rebuild-role", "dashboard", "view"); !ok {
			t.Fatal("前置条件:残留策略应命中(模拟 DeleteRole 清理失败的越权隐患)")
		}

		// 重建同名角色(CreateRole 兜底清理残留策略)
		if err := CreateRole(model.Role{Name: "rebuild-role", Label: "重建角色2"}); err != nil {
			t.Fatalf("重建角色失败: %v", err)
		}

		// 验证:新角色无任何旧权限(越权被阻断)
		if ok, _ := HasPermission("rebuild-role", "dashboard", "view"); ok {
			t.Error("重建后不应继承残留的 dashboard:view(越权)")
		}
		if ok, _ := HasPermission("rebuild-role", "server", "read"); ok {
			t.Error("重建后不应继承残留的 server:read(越权)")
		}
	})
}
