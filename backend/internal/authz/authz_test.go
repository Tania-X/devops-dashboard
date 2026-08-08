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
