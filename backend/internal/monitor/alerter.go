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
	prevStatus map[string]string // key: "cpu"|"memory"|"disk", value: 已确认状态 "normal"|"warning"|"critical"
	nextID     int
	thresholds map[string]MetricThreshold
	streak     map[string]int // 连续异常周期计数(达到 confirmPeriods 才确认告警,防瞬时抖动)
	confirmPeriods int        // 连续超阈值确认周期数(>=1)

	// OnAlert 可选回调：每条新告警产生时触发（用于 Webhook 推送等外部通知）
	// 在 addAlert 内同步调用；回调应快速返回（如只做 channel 投递），不得阻塞
	OnAlert func(model.AlertItem)
}

// defaultConfirmPeriods 告警确认周期:连续 N 个采集周期超阈值才真正告警,
// 防瞬时抖动(如 CPU 偶尔飙一下)误报。采集间隔约 10s → 3 周期 ≈ 30s 确认。
const defaultConfirmPeriods = 3

func NewAlerter() *Alerter {
	return NewAlerterWithConfirm(defaultConfirmPeriods)
}

// NewAlerterWithConfirm 指定确认周期创建告警评估器(测试/配置用;periods=1 即立即告警)
func NewAlerterWithConfirm(periods int) *Alerter {
	thresholds := make(map[string]MetricThreshold, len(defaultThresholds))
	for k, v := range defaultThresholds {
		thresholds[k] = v
	}
	if periods < 1 {
		periods = 1
	}
	return &Alerter{
		maxAlerts:  20,
		prevStatus: make(map[string]string),
		thresholds: thresholds,
		streak:     make(map[string]int),
		confirmPeriods: periods,
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

	prevStatus := a.prevStatus[name] // 已确认状态

	// ── 异常状态:连续计数,达到确认周期才动作(防瞬时抖动) ──
	if currentStatus != "normal" {
		a.streak[name]++
		// 未达确认周期 → 不告警(仍在确认中,prevStatus 不变)
		if a.streak[name] < a.confirmPeriods {
			return
		}
		// 已达确认周期:若与已确认状态不同 → 触发/升级
		if currentStatus != prevStatus {
			a.prevStatus[name] = currentStatus
			threshold := t.Warn
			label := "warning"
			if currentStatus == "critical" {
				threshold = t.Crit
				label = "critical"
			}
			a.addAlert(currentStatus, fmt.Sprintf("%s 使用率 %.1f%% — 超过 %s 阈值 (%.0f%%)",
				metricLabel(name), value, label, threshold), name, value)
		}
		// 状态相同且已达确认周期 → 不重复(去重)
		return
	}

	// ── 正常状态:重置计数;之前有确认异常 → 恢复通知 ──
	a.streak[name] = 0
	if prevStatus != "" && prevStatus != "normal" {
		a.prevStatus[name] = "normal"
		a.addAlert("info", fmt.Sprintf("%s 使用率已恢复至 %.1f%%", metricLabel(name), value), name, value)
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
