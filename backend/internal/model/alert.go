package model

import "time"

// Alert 告警历史记录(落库,支持分页查询)。
// 与 AlertItem(内存环形缓冲,仪表盘"最近告警")互补:AlertItem 看最近,
// 本表查历史(重启不丢)。
type Alert struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Level     string    `gorm:"index" json:"level"`    // 级别筛选走索引
	Message   string    `json:"message"`
	Source    string    `json:"source"`                // localhost / agent host
	Time      string    `json:"time"`                  // 展示格式 "01-02 15:04"(与 AlertItem 一致)
	CreatedAt time.Time `gorm:"index" json:"createdAt"` // 排序 + 过期清理走索引
}
