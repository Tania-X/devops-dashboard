package model

import "time"

// Role 角色实体（RBAC 二期扩展：角色从代码配置改为数据库实体）。
// builtin=true 的内置角色（admin/operator/viewer）不可删除、不可改名；
// locked=true 的角色权限不可修改（admin 通配策略锁定）。
type Role struct {
	Name        string    `gorm:"primaryKey" json:"name"`
	Label       string    `gorm:"not null" json:"label"`
	Description string    `json:"description"`
	Builtin     bool      `json:"builtin"`
	Locked      bool      `json:"locked"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateRoleRequest 创建角色入参
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Label       string `json:"label" binding:"required"`
	Description string `json:"description"`
}

// UpdateRoleRequest 更新角色入参（名称不可修改，仅改标签/描述）
type UpdateRoleRequest struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}
