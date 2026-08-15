package domain

import (
	"strings"
	"testing"
)

func TestRoleValid(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{"内置 admin", UserRoleAdmin, true},
		{"内置 viewer", UserRoleViewer, true},
		{"内置 operator", UserRoleOperator, true},
		{"自定义角色", UserRole("ops-team"), true},
		{"空角色", UserRole(""), false},
		{"超长角色", UserRole(strings.Repeat("a", 33)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.Valid(); got != tt.want {
				t.Errorf("Role(%q).Valid() = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestNewUser(t *testing.T) {
	t.Run("创建成功且密码已哈希", func(t *testing.T) {
		u, err := NewUser("alice", "secret123", UserRoleViewer)
		if err != nil {
			t.Fatalf("NewUser 失败: %v", err)
		}
		if u.Username != "alice" {
			t.Errorf("username = %q, want alice", u.Username)
		}
		if u.Role != string(UserRoleViewer) {
			t.Errorf("role = %q, want viewer", u.Role)
		}
		if u.ID == "" {
			t.Error("ID 不应为空(工厂应生成)")
		}
		// 密码必须是 bcrypt 哈希(不是明文)
		if u.Password == "secret123" {
			t.Error("Password 不应是明文")
		}
		if !u.VerifyPassword("secret123") {
			t.Error("正确密码应校验通过")
		}
		if u.VerifyPassword("wrong") {
			t.Error("错误密码不应校验通过")
		}
	})

	t.Run("空用户名拒绝", func(t *testing.T) {
		if _, err := NewUser("", "secret", UserRoleViewer); err == nil {
			t.Error("空用户名应报错")
		}
	})

	t.Run("空密码拒绝", func(t *testing.T) {
		if _, err := NewUser("bob", "", UserRoleViewer); err == nil {
			t.Error("空密码应报错")
		}
	})

	t.Run("非法角色拒绝", func(t *testing.T) {
		if _, err := NewUser("bob", "secret", UserRole("")); err == nil {
			t.Error("空角色应报错")
		}
	})
}

func TestUserChangeRole(t *testing.T) {
	u, _ := NewUser("alice", "secret", UserRoleViewer)

	if err := u.ChangeRole(UserRoleOperator); err != nil {
		t.Fatalf("合法角色变更失败: %v", err)
	}
	if u.Role != string(UserRoleOperator) {
		t.Errorf("role = %q, want operator", u.Role)
	}

	if err := u.ChangeRole(UserRole("")); err == nil {
		t.Error("空角色应报错")
	}
	if err := u.ChangeRole(UserRole(strings.Repeat("x", 33))); err == nil {
		t.Error("超长角色应报错")
	}
}

func TestUserSetPassword(t *testing.T) {
	u, _ := NewUser("alice", "old", UserRoleViewer)

	if err := u.SetPassword("new-pass"); err != nil {
		t.Fatalf("SetPassword 失败: %v", err)
	}
	if !u.VerifyPassword("new-pass") {
		t.Error("新密码应校验通过")
	}
	if u.VerifyPassword("old") {
		t.Error("旧密码不应再通过")
	}
	if err := u.SetPassword(""); err == nil {
		t.Error("空密码应报错")
	}
}
