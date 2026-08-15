package model

import "time"

// Alert 告警历史记录(落库,支持分页查询)。
// 与 AlertItem(内存环形缓冲,仪表盘"最近告警")互补:AlertItem 看最近,
// 本表查历史(重启不丢)。
type Alert struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Level     string    `json:"level"`     // info | warning | critical
	Message   string    `json:"message"`
	Source    string    `json:"source"`    // localhost / agent host
	Time      string    `json:"time"`      // 展示格式 "01-02 15:04"(与 AlertItem 一致)
	CreatedAt time.Time `json:"createdAt"` // 落库时间,查询排序依据
}
