package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Agent 状态值(集中定义,替代散落的魔法字符串)
const (
	AgentStatusUnknown = "unknown" // 初始/部署后待确认
	AgentStatusOnline  = "online"  // 健康检查通过
	AgentStatusOffline = "offline" // 停止后或健康检查失败
)

// AgentTarget Agent 实体(充血模型)。
// 状态转换规则收敛到实体方法:CheckStoppable(停止前校验)、
// MarkDeployed/MarkOnline/MarkOffline(部署/健康检查结果落状态)。
// 敏感字段 Password json:"-" 永不序列化;SSH/加密等基础设施不在 domain。
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

// NewAgent 工厂:生成 ID + 初始状态 unknown(密码加密由应用服务调用 infra 完成)
func NewAgent(name, host string, port int, username, authType, deployDir string, agentPort int) *AgentTarget {
	return &AgentTarget{
		ID:        uuid.New().String(),
		Name:      name,
		Host:      host,
		Port:      port,
		Username:  username,
		AuthType:  authType,
		DeployDir: deployDir,
		AgentPort: agentPort,
		Status:    AgentStatusUnknown,
	}
}

// CheckStoppable 停止前状态校验:已离线则拒绝重复停止。
// 部署(Deploy)是幂等更新操作,任意状态可部署,无需前置校验。
func (a *AgentTarget) CheckStoppable() error {
	if a.Status == AgentStatusOffline {
		return errors.New("Agent 已离线,无需停止")
	}
	return nil
}

// MarkDeployed 部署完成:状态置 unknown(等待健康检查确认在线/离线)
func (a *AgentTarget) MarkDeployed() {
	a.Status = AgentStatusUnknown
}

// MarkOnline 健康检查通过
func (a *AgentTarget) MarkOnline() {
	a.Status = AgentStatusOnline
}

// MarkOffline 健康检查失败/已停止
func (a *AgentTarget) MarkOffline() {
	a.Status = AgentStatusOffline
}

// CreateAgentRequest 创建 Agent 入参(密码仅入站,实体 Password json:"-" 不序列化)
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

// UpdateAgentRequest 更新 Agent 入参(密码留空表示不修改)
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
