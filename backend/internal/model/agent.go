package model

import "time"

type AgentTarget struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	AuthType  string    `json:"authType"`
	Password  string    `json:"-"`
	DeployDir string    `json:"deployDir"`
	AgentPort int       `json:"agentPort"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateAgentRequest 创建 Agent 入参（密码仅入站，实体 Password json:"-" 不序列化）
type CreateAgentRequest struct {
	Name      string `json:"name" binding:"required"`
	Host      string `json:"host" binding:"required"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	AuthType  string `json:"authType"`
	Password  string `json:"password"`
	DeployDir string `json:"deployDir"`
	AgentPort int    `json:"agentPort"`
}

// UpdateAgentRequest 更新 Agent 入参（密码留空表示不修改）
type UpdateAgentRequest struct {
	Name      string `json:"name"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	AuthType  string `json:"authType"`
	Password  string `json:"password"`
	DeployDir string `json:"deployDir"`
	AgentPort int    `json:"agentPort"`
}
