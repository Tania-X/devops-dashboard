package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRole 用户角色值对象。
// 内置角色有预置权限(见 authz seed);自定义角色由角色管理功能维护。
// 注意:model.Role 是角色表实体(角色管理用),此处 UserRole 是用户角色值的类型约束。
type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleViewer   UserRole = "viewer"
	UserRoleOperator UserRole = "operator"
)

// Valid 校验角色是否合法:内置角色直接通过;自定义角色要求非空且长度 <= 32。
// 注意:仅校验"形式合法",角色是否存在于 roles 表由应用服务层查证(需访问 DB)。
func (r UserRole) Valid() bool {
	switch r {
	case UserRoleAdmin, UserRoleViewer, UserRoleOperator:
		return true
	}
	s := string(r)
	return s != "" && len(s) <= 32
}

// User 用户实体(充血模型)。
// 行为内聚:密码的哈希/校验、角色变更都收敛到实体方法,Service 只做编排。
// 敏感字段 Password 用 json:"-" 永不序列化,任何接口返回 User 都不会泄露密码。
type User struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SetPassword 设置密码:bcrypt 哈希后存入 Password。
// 密码策略(最小长度等)后续可在此收紧。
func (u *User) SetPassword(plain string) error {
	if plain == "" {
		return errors.New("密码不能为空")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

// VerifyPassword 校验密码是否正确(bcrypt 比对)。
func (u *User) VerifyPassword(plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plain)) == nil
}

// ChangeRole 变更角色:校验角色合法性(内置/格式)。
// 角色是否存在由应用服务层调用 authz.RoleExists 查证后,再调用本方法。
func (u *User) ChangeRole(r UserRole) error {
	if !r.Valid() {
		return fmt.Errorf("非法角色: %q", string(r))
	}
	u.Role = string(r)
	return nil
}

// NewUser 工厂:创建用户实体,内部完成 ID 生成与密码哈希。
// 返回的实体尚未落库,持久化由应用服务负责。
func NewUser(username, plain string, role UserRole) (*User, error) {
	if username == "" {
		return nil, errors.New("用户名不能为空")
	}
	if !role.Valid() {
		return nil, fmt.Errorf("非法角色: %q", string(role))
	}
	u := &User{
		ID:       uuid.New().String(),
		Username: username,
		Role:     string(role),
	}
	if err := u.SetPassword(plain); err != nil {
		return nil, err
	}
	return u, nil
}

// CreateUserRequest 创建用户入参。
// 设计原则:敏感字段(Password)只出现在请求 DTO 中,实体 User.Password 用
// json:"-" 永不序列化,任何接口返回 User 都不会泄露密码,无需手动清空。
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

// UpdateUserRequest 更新用户入参(密码留空表示不修改;用户名不可修改)
type UpdateUserRequest struct {
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token       string   `json:"token"`
	User        User     `json:"user"`
	Permissions []string `json:"permissions"` // 当前用户拥有的权限点(前端按钮级控制)
}

// MeResponse 当前登录用户信息 + 权限点
type MeResponse struct {
	User        User     `json:"user"`
	Permissions []string `json:"permissions"`
}
