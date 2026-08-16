package service

import (
	"fmt"
	"net/http"
	"time"

	agentdomain "github.com/Tania-X/devops-dashboard/backend/internal/dashboard/agent/domain"
	"github.com/Tania-X/devops-dashboard/backend/internal/dashboard/agent/infra"
	"gorm.io/gorm"
)

// AgentService Agent 应用服务(编排者):查实体 → 校验/状态机 → infra(SSH/加密)→ 落库。
// SSH/加密细节已抽离到 infra,domain 不依赖基础设施。
type AgentService struct {
	db       *gorm.DB
	crypto   *infra.Crypto
	ssh      *infra.SSH
	agentBin string
}

func NewAgentService(db *gorm.DB, secretKey string, agentBin string) *AgentService {
	if agentBin == "" {
		agentBin = "bin/agent-linux-amd64"
	}
	return &AgentService{db: db, crypto: infra.NewCrypto(secretKey), ssh: &infra.SSH{}, agentBin: agentBin}
}

func (s *AgentService) List() ([]agentdomain.AgentTarget, error) {
	var targets []agentdomain.AgentTarget
	if err := s.db.Order("created_at DESC").Find(&targets).Error; err != nil {
		return nil, err
	}
	return targets, nil
}

// Create 创建 Agent:工厂生成实体(ID + 初始状态)→ 密码加密(infra)→ 落库
func (s *AgentService) Create(req agentdomain.CreateAgentRequest) (*agentdomain.AgentTarget, error) {
	target := agentdomain.NewAgent(req.Name, req.Host, req.Port, req.Username, req.AuthType, req.DeployDir, req.AgentPort)
	if req.Password != "" {
		encrypted, err := s.crypto.Encrypt(req.Password)
		if err != nil {
			return nil, err
		}
		target.Password = encrypted
	}
	now := time.Now()
	target.CreatedAt = now
	target.UpdatedAt = now
	if err := s.db.Create(target).Error; err != nil {
		return nil, err
	}
	return target, nil
}

// Update 更新 Agent:密码留空表示不修改;字段更新后落库
func (s *AgentService) Update(id string, req agentdomain.UpdateAgentRequest) (*agentdomain.AgentTarget, error) {
	var existing agentdomain.AgentTarget
	if err := s.db.First(&existing, "id = ?", id).Error; err != nil {
		return nil, err
	}
	// 普通字段无条件覆盖(与旧版一致,支持清空;前端总是提交全量表单)。
	// 仅密码特殊:留空表示不修改。
	existing.Name = req.Name
	existing.Host = req.Host
	existing.Port = req.Port
	existing.Username = req.Username
	existing.AuthType = req.AuthType
	existing.DeployDir = req.DeployDir
	existing.AgentPort = req.AgentPort
	if req.Password != "" {
		encrypted, err := s.crypto.Encrypt(req.Password)
		if err != nil {
			return nil, err
		}
		existing.Password = encrypted
	}
	existing.UpdatedAt = time.Now()
	if err := s.db.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *AgentService) Delete(id string) error {
	return s.db.Delete(&agentdomain.AgentTarget{}, "id = ?", id).Error
}

func (s *AgentService) GetByID(id string) (*agentdomain.AgentTarget, error) {
	var target agentdomain.AgentTarget
	if err := s.db.First(&target, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

// Deploy 部署 Agent(幂等,任意状态可重部署=更新):
// 解密 → SSH 传文件/启动 → 实体 MarkDeployed(状态置 unknown,待健康检查确认)
func (s *AgentService) Deploy(id string) error {
	target, err := s.GetByID(id)
	if err != nil {
		return fmt.Errorf("未找到目标: %w", err)
	}
	password, err := s.crypto.Decrypt(target.Password)
	if err != nil {
		return fmt.Errorf("密码解密失败: %w", err)
	}
	client, err := s.ssh.Connect(target.Host, target.Port, target.Username, password)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer client.Close()
	if err := s.ssh.Exec(client, fmt.Sprintf("mkdir -p %s", target.DeployDir)); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	// 停止旧 agent(与 Stop 一致检查错误;否则旧进程可能占用端口导致新 agent 启动失败)
	if err := s.ssh.Exec(client, fmt.Sprintf("pkill -f %s/agent || true", target.DeployDir)); err != nil {
		return fmt.Errorf("停止旧 agent 失败: %w", err)
	}
	if err := s.ssh.Upload(client, s.agentBin, fmt.Sprintf("%s/agent", target.DeployDir)); err != nil {
		return fmt.Errorf("上传 agent 失败: %w", err)
	}
	if err := s.ssh.Exec(client, fmt.Sprintf("chmod +x %s/agent", target.DeployDir)); err != nil {
		return fmt.Errorf("设置权限失败: %w", err)
	}
	startCmd := fmt.Sprintf("cd %s && PORT=%d nohup ./agent > agent.log 2>&1 &", target.DeployDir, target.AgentPort)
	if err := s.ssh.Exec(client, startCmd); err != nil {
		return fmt.Errorf("启动 agent 失败: %w", err)
	}

	// 状态机:部署完成 → unknown(待健康检查确认)
	target.MarkDeployed()
	target.UpdatedAt = time.Now()
	return s.db.Save(target).Error
}

// Stop 停止 Agent:实体校验(offline 拒绝重复停止)→ SSH 结束进程 → MarkOffline
func (s *AgentService) Stop(id string) error {
	target, err := s.GetByID(id)
	if err != nil {
		return err
	}
	if err := target.CheckStoppable(); err != nil {
		return err
	}
	password, err := s.crypto.Decrypt(target.Password)
	if err != nil {
		return err
	}
	client, err := s.ssh.Connect(target.Host, target.Port, target.Username, password)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer client.Close()
	if err := s.ssh.Exec(client, fmt.Sprintf("pkill -f %s/agent || true", target.DeployDir)); err != nil {
		return fmt.Errorf("停止 agent 失败: %w", err)
	}

	target.MarkOffline()
	target.UpdatedAt = time.Now()
	return s.db.Save(target).Error
}

// StatusCheck 健康检查:HTTP 探活 → 实体 MarkOnline/MarkOffline → 落库
func (s *AgentService) StatusCheck(id string) (string, error) {
	target, err := s.GetByID(id)
	if err != nil {
		return agentdomain.AgentStatusUnknown, err
	}
	url := fmt.Sprintf("http://%s:%d/api/health", target.Host, target.AgentPort)
	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Get(url)
	if err == nil && resp != nil {
		// 关闭响应体归还连接池(仅成功响应分支注册 defer,err 非 nil 时无 Body 可关)
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			target.MarkOnline()
		} else {
			target.MarkOffline()
		}
	} else {
		target.MarkOffline()
	}
	target.UpdatedAt = time.Now()
	if err := s.db.Save(target).Error; err != nil {
		return target.Status, err
	}
	return target.Status, nil
}

// AgentDeployResult 部署操作结果(前端展示)
type AgentDeployResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
