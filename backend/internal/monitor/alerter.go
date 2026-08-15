package monitor

import (
	"fmt"
	"sync"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
)

// MetricThreshold 单个指标的两档阈值
type MetricThreshold struct {
	Warn float64
	Crit float64
}

// defaultThresholds 默认阈值(启动兜底;可通过 AlertThresholdManager 热更新)
var defaultThresholds = map[string]MetricThreshold{
	"cpu":    {Warn: 60, Crit: 80},
	"memory": {Warn: 70, Crit: 85},
	"disk":   {Warn: 75, Crit: 90},
}

// Alerter 告警评估器
// 根据采集数据判断阈值，生成告警条目并缓存在内存中
type Alerter struct {
	mu         sync.RWMutex
	alerts     []model.AlertItem
	maxAlerts  int
	prevStatus map[string]string // key: "cpu"|"memory"|"disk", value: "normal"|"warning"|"critical"
	nextID     int
	thresholds map[string]MetricThreshold

	// OnAlert 可选回调：每条新告警产生时触发（用于 Webhook 推送等外部通知）
	// 在 addAlert 内同步调用；回调应快速返回（如只做 channel 投递），不得阻塞
	OnAlert func(model.AlertItem)
}

func NewAlerter() *Alerter {
	thresholds := make(map[string]MetricThreshold, len(defaultThresholds))
	for k, v := range defaultThresholds {
		thresholds[k] = v
	}
	return &Alerter{
		maxAlerts:  20,
		prevStatus: make(map[string]string),
		thresholds: thresholds,
	}
}

// SetThreshold 热更新某指标阈值(并发安全;下次 Evaluate 即生效)
func (a *Alerter) SetThreshold(metric string, warn, crit float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.thresholds[metric] = MetricThreshold{Warn: warn, Crit: crit}
}

// GetThresholds 返回当前生效的阈值快照(含默认值)
func (a *Alerter) GetThresholds() map[string]MetricThreshold {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]MetricThreshold, len(a.thresholds))
	for k, v := range a.thresholds {
		out[k] = v
	}
	return out
}

// thresholdFor 读取单指标阈值(读锁,防与 SetThreshold 并发)
func (a *Alerter) thresholdFor(metric string) MetricThreshold {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.thresholds[metric]
}

// Evaluate 根据采集快照评估告警规则
// 每条规则：值超过 critical 阈值 → critical 告警；超过 warning 阈值 → warning 告警
// 和上一次状态相同 → 不重复生成（去重）；异常恢复正常 → 生成"已恢复"信息
func (a *Alerter) Evaluate(snapshot *MetricSnapshot) {
	a.evaluateMetric("cpu", snapshot.CPUPercent, a.thresholdFor("cpu"))
	a.evaluateMetric("memory", snapshot.MemoryPercent, a.thresholdFor("memory"))
	a.evaluateMetric("disk", snapshot.DiskPercent, a.thresholdFor("disk"))
}

func (a *Alerter) evaluateMetric(name string, value float64, t MetricThreshold) {
	a.mu.Lock()
	defer a.mu.Unlock()

	currentStatus := "normal"
	if value >= t.Crit {
		currentStatus = "critical"
	} else if value >= t.Warn {
		currentStatus = "warning"
	}

	prevStatus := a.prevStatus[name]

	// 状态没变 → 不重复告警
	if currentStatus == prevStatus {
		return
	}

	a.prevStatus[name] = currentStatus

	// 从异常恢复到正常
	if currentStatus == "normal" && prevStatus != "" && prevStatus != "normal" {
		a.addAlert("info", fmt.Sprintf("%s 使用率已恢复至 %.1f%%", metricLabel(name), value), name, value)
		return
	}

	// 触发新告警
	if currentStatus != "normal" {
		threshold := t.Warn
		label := "warning"
		if currentStatus == "critical" {
			threshold = t.Crit
			label = "critical"
		}
		a.addAlert(currentStatus, fmt.Sprintf("%s 使用率 %.1f%% — 超过 %s 阈值 (%.0f%%)",
			metricLabel(name), value, label, threshold), name, value)
	}
}

func (a *Alerter) addAlert(level, message, metric string, value float64) {
	a.nextID++
	entry := model.AlertItem{
		ID:      fmt.Sprintf("alert-%03d", a.nextID),
		Level:   level,
		Message: message,
		Source:  "localhost",
		Time:    time.Now().Format("01-02 15:04"),
	}
	a.alerts = append(a.alerts, entry)
	if len(a.alerts) > a.maxAlerts {
		a.alerts = a.alerts[len(a.alerts)-a.maxAlerts:]
	}

	// 触发外部通知回调（如 Webhook 推送）
	if a.OnAlert != nil {
		a.OnAlert(entry)
	}
}

// GetAlerts 返回最近的告警列表（按时间倒序，最新的在前）
func (a *Alerter) GetAlerts(limit int) []model.AlertItem {
	a.mu.RLock()
	defer a.mu.RUnlock()

	n := len(a.alerts)
	if n == 0 {
		return []model.AlertItem{}
	}
	if limit > n {
		limit = n
	}

	result := make([]model.AlertItem, limit)
	for i := 0; i < limit; i++ {
		result[i] = a.alerts[n-1-i]
	}
	return result
}

func metricLabel(name string) string {
	switch name {
	case "cpu":
		return "CPU"
	case "memory":
		return "内存"
	case "disk":
		return "磁盘"
	}
	return name
}
