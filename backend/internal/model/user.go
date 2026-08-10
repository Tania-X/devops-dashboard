package model

import "time"

type User struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateUserRequest 创建用户入参。
// 设计原则：敏感字段（Password）只出现在请求 DTO 中，实体 User.Password 用
// json:"-" 永不序列化，任何接口返回 User 都不会泄露密码，无需手动清空。
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

// UpdateUserRequest 更新用户入参（密码留空表示不修改；用户名不可修改）
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
	Permissions []string `json:"permissions"` // 当前用户拥有的权限点（前端按钮级控制）
}

// MeResponse 当前登录用户信息 + 权限点
type MeResponse struct {
	User        User     `json:"user"`
	Permissions []string `json:"permissions"`
}
