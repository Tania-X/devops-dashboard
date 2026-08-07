package service

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/Tania-X/devops-dashboard/backend/internal/notify"
	"gorm.io/gorm"
)

// nowStr 返回当前时间（与 Alerter 告警时间格式保持一致）
func nowStr() string {
	return time.Now().Format("01-02 15:04")
}

// WebhookManager 管理当前生效的 Webhook 通知器，支持热更新
//
// 设计：前端改配置 → 存 DB → Rebuild() 用新配置创建新的 WebhookNotifier
// 并原子替换进 AlertBus。配置未启用/未配置时 notifier 为 nil（零开销）。
type WebhookManager struct {
	mu       sync.RWMutex
	bus      *notify.AlertBus
	db       *gorm.DB
	notifier notify.Notifier // 当前生效的通知器（可能为 nil）
}

// NewWebhookManager 创建 Webhook 管理器并注册进告警总线
func NewWebhookManager(db *gorm.DB, bus *notify.AlertBus) *WebhookManager {
	m := &WebhookManager{db: db, bus: bus}
	m.Rebuild()
	return m
}

// Get 返回当前配置（secret 已脱敏，不回传前端）
func (m *WebhookManager) Get() *model.WebhookConfig {
	var cfg model.WebhookConfig
	err := m.db.First(&cfg, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &model.WebhookConfig{ID: 1, Enabled: false, Kind: "dingtalk"}
	}
	if err != nil {
		slog.Error("读取 Webhook 配置失败", "err", err)
		return &model.WebhookConfig{ID: 1, Enabled: false, Kind: "dingtalk"}
	}
	return &cfg
}

// Update 保存配置并热更新通知器
// secret 为空表示不修改；URL/Kind 变化时重建 Notifier
func (m *WebhookManager) Update(input model.WebhookConfig) (*model.WebhookConfig, error) {
	// 校验
	if input.Kind != "dingtalk" && input.Kind != "wecom" {
		return nil, errors.New("kind 仅支持 dingtalk 或 wecom")
	}
	if input.URL == "" {
		return nil, errors.New("URL 不能为空")
	}

	var cfg model.WebhookConfig
	err := m.db.First(&cfg, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cfg = model.WebhookConfig{ID: 1}
	} else if err != nil {
		return nil, err
	}

	cfg.Enabled = input.Enabled
	cfg.Kind = input.Kind
	cfg.URL = input.URL
	if input.Secret != "" {
		cfg.Secret = input.Secret
	}

	if err := m.db.Save(&cfg).Error; err != nil {
		return nil, err
	}

	m.Rebuild()
	return &cfg, nil
}

// Test 向当前配置的 URL 发送一条测试告警，验证渠道连通
func (m *WebhookManager) Test() (string, error) {
	m.mu.RLock()
	n := m.notifier
	m.mu.RUnlock()

	if n == nil {
		return "", errors.New("Webhook 未启用或未配置")
	}

	entry := model.AlertItem{
		ID:      "test-001",
		Level:   "info",
		Message: "这是一条测试告警，用于验证 Webhook 通道连通性",
		Source:  "devops-dashboard",
		Time:    nowStr(),
	}
	if err := n.Send(entry); err != nil {
		return "", err
	}
	return "测试消息已发送", nil
}

// Rebuild 根据 DB 配置重建通知器并替换进总线（配置变更后调用）
func (m *WebhookManager) Rebuild() {
	cfg := m.Get()

	m.mu.Lock()
	defer m.mu.Unlock()

	if !cfg.Enabled || cfg.URL == "" {
		m.notifier = nil
		m.bus.SetNotifiers(nil)
		slog.Info("Webhook 推送未启用")
		return
	}

	n := notify.NewWebhookNotifier(cfg.Kind, cfg.URL)
	m.notifier = n
	m.bus.SetNotifiers([]notify.Notifier{n})
	slog.Info("Webhook 推送已启用", "kind", cfg.Kind, "url", cfg.URL)
}
