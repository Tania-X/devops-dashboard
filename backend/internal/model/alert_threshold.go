package model

import "time"

// AlertThreshold 告警阈值配置(按指标),热生效,无需重启。
// 阈值默认值在 monitor.Alerter(启动兜底);DB 有记录时以 DB 为准。
type AlertThreshold struct {
	Metric        string    `gorm:"primaryKey" json:"metric"`
	WarnThreshold float64   `json:"warnThreshold"`
	CritThreshold float64   `json:"critThreshold"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// UpdateAlertThresholdRequest 更新告警阈值入参
type UpdateAlertThresholdRequest struct {
	Metric        string  `json:"metric" binding:"required"`
	WarnThreshold float64 `json:"warnThreshold" binding:"required"`
	CritThreshold float64 `json:"critThreshold" binding:"required"`
}
