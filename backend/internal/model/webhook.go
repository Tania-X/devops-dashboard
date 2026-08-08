package model

import "time"

// WebhookConfig 告警 Webhook 推送配置（单条记录，ID 恒为 1）
// 存于 settings 表，前端可配置，保存后热生效
type WebhookConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Enabled   bool      `json:"enabled"`  // 推送开关
	Kind      string    `json:"kind"`     // 渠道类型：dingtalk | wecom
	URL       string    `json:"url"`      // 机器人 Webhook 地址
	Secret    string    `json:"-"`        // 加签密钥（钉钉安全设置），不回传前端
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UpdateWebhookConfigRequest 更新 Webhook 配置入参（secret 留空表示不修改）
type UpdateWebhookConfigRequest struct {
	Enabled bool   `json:"enabled"`
	Kind    string `json:"kind"`
	URL     string `json:"url"`
	Secret  string `json:"secret"`
}
