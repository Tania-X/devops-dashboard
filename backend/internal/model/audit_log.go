package model

import "time"

// AuditLog 审计日志实体：记录"谁在什么时候对什么做了什么"。
// 只记录敏感管理操作（角色/权限/用户增删改），不记录登录等噪音。
type AuditLog struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Actor     string    `json:"actor"`  // 操作人用户名
	Action    string    `json:"action"` // 动作枚举：role.create / role.update / role.delete / permission.update / user.create / user.update / user.delete
	Target    string    `json:"target"` // 操作对象描述，如 "角色 operator" / "用户 zhangsan"
	Detail    string    `json:"detail"` // JSON 详情（变更前后内容）
	CreatedAt time.Time `json:"createdAt"`
}
