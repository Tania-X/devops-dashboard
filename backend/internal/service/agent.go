package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

type AgentService struct {
	db        *gorm.DB
	secretKey []byte
	agentBin  string
}

func NewAgentService(db *gorm.DB, secretKey string, agentBin string) *AgentService {
	if agentBin == "" {
		agentBin = "bin/agent-linux-amd64"
	}
	return &AgentService{db: db, secretKey: padOrTrimKey([]byte(secretKey)), agentBin: agentBin}
}

func (s *AgentService) List() ([]model.AgentTarget, error) {
	var targets []model.AgentTarget
	if err := s.db.Order("created_at DESC").Find(&targets).Error; err != nil {
		return nil, err
	}
	return targets, nil
}

func (s *AgentService) Create(target model.AgentTarget) (*model.AgentTarget, error) {
	target.ID = uuid.New().String()
	target.Status = "unknown"
	target.CreatedAt = time.Now()
	target.UpdatedAt = target.CreatedAt
	if target.Password != "" {
		encrypted, err := s.encrypt(target.Password)
		if err != nil {
			return nil, err
		}
		target.Password = encrypted
	}
	if err := s.db.Create(&target).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

func (s *AgentService) Update(id string, target model.AgentTarget) (*model.AgentTarget, error) {
	var existing model.AgentTarget
	if err := s.db.First(&existing, "id = ?", id).Error; err != nil {
		return nil, err
	}
	existing.Name = target.Name
	existing.Host = target.Host
	existing.Port = target.Port
	existing.Username = target.Username
	existing.AuthType = target.AuthType
	existing.DeployDir = target.DeployDir
	existing.AgentPort = target.AgentPort
	existing.UpdatedAt = time.Now()
	if target.Password != "" {
		encrypted, err := s.encrypt(target.Password)
		if err != nil {
			return nil, err
		}
		existing.Password = encrypted
	}
	if err := s.db.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *AgentService) Delete(id string) error {
	return s.db.Delete(&model.AgentTarget{}, "id = ?", id).Error
}

func (s *AgentService) GetByID(id string) (*model.AgentTarget, error) {
	var target model.AgentTarget
	if err := s.db.First(&target, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

func (s *AgentService) Deploy(id string) error {
	target, err := s.GetByID(id)
	if err != nil {
		return fmt.Errorf("未找到目标: %w", err)
	}
	password, err := s.decrypt(target.Password)
	if err != nil {
		return fmt.Errorf("密码解密失败: %w", err)
	}
	client, err := s.sshConnect(target.Host, target.Port, target.Username, password)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer client.Close()
	if err := s.sshExec(client, fmt.Sprintf("mkdir -p %s", target.DeployDir)); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	s.sshExec(client, fmt.Sprintf("pkill -f %s/agent || true", target.DeployDir))
	if err := s.scpUpload(client, s.agentBin, filepath.Join(target.DeployDir, "agent")); err != nil {
		return fmt.Errorf("上传 agent 失败: %w", err)
	}
	if err := s.sshExec(client, fmt.Sprintf("chmod +x %s/agent", target.DeployDir)); err != nil {
		return fmt.Errorf("设置权限失败: %w", err)
	}
	startCmd := fmt.Sprintf("cd %s && PORT=%d nohup ./agent > agent.log 2>&1 &", target.DeployDir, target.AgentPort)
	if err := s.sshExec(client, startCmd); err != nil {
		return fmt.Errorf("启动 agent 失败: %w", err)
	}
	s.db.Model(target).Updates(map[string]interface{}{"status": "unknown", "updated_at": time.Now()})
	return nil
}

func (s *AgentService) Stop(id string) error {
	target, err := s.GetByID(id)
	if err != nil {
		return err
	}
	password, err := s.decrypt(target.Password)
	if err != nil {
		return err
	}
	client, err := s.sshConnect(target.Host, target.Port, target.Username, password)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer client.Close()
	if err := s.sshExec(client, fmt.Sprintf("pkill -f %s/agent || true", target.DeployDir)); err != nil {
		return fmt.Errorf("停止 agent 失败: %w", err)
	}
	s.db.Model(target).Updates(map[string]interface{}{"status": "offline", "updated_at": time.Now()})
	return nil
}

func (s *AgentService) StatusCheck(id string) (string, error) {
	target, err := s.GetByID(id)
	if err != nil {
		return "unknown", err
	}
	url := fmt.Sprintf("http://%s:%d/api/health", target.Host, target.AgentPort)
	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Get(url)
	status := "offline"
	if err == nil && resp != nil && resp.StatusCode == 200 {
		status = "online"
	}
	s.db.Model(target).Updates(map[string]interface{}{"status": status, "updated_at": time.Now()})
	return status, nil
}

func (s *AgentService) sshConnect(host string, port int, user, password string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
}

func (s *AgentService) sshExec(client *ssh.Client, cmd string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	var stderr bytes.Buffer
	session.Stderr = &stderr
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("%s: %w", stderr.String(), err)
	}
	return nil
}

func (s *AgentService) scpUpload(client *ssh.Client, localPath, remotePath string) error {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("读取本地文件失败: %w", err)
	}
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%#o %d %s\n", 0755, len(data), filepath.Base(remotePath))
		w.Write(data)
		fmt.Fprint(w, "\x00")
	}()
	return session.Run(fmt.Sprintf("scp -t %s", remotePath))
}

func (s *AgentService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.secretKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *AgentService) decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", errors.New("空密码")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.secretKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("密文太短")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func padOrTrimKey(key []byte) []byte {
	const keySize = 32
	if len(key) >= keySize {
		return key[:keySize]
	}
	padded := make([]byte, keySize)
	copy(padded, key)
	return padded
}

type AgentDeployResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Status  string `json:"status,omitempty"`
}

func (r AgentDeployResult) JSON() []byte {
	b, _ := json.Marshal(r)
	return b
}
