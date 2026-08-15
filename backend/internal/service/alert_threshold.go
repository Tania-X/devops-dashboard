package service

import (
	"errors"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"
	"gorm.io/gorm"
)

// alertMetrics 支持的告警指标
var alertMetrics = []string{"cpu", "memory", "disk"}

// AlertThresholdManager 管理告警阈值配置,支持热生效(无需重启)。
// 设计:启动时从 DB 加载已配置阈值 → alerter.SetThreshold;
//      更新时校验(warn < crit)→ 存 DB → 热设置 alerter,下次采集即生效。
type AlertThresholdManager struct {
	db      *gorm.DB
	alerter *monitor.Alerter
}

func NewAlertThresholdManager(db *gorm.DB, alerter *monitor.Alerter) *AlertThresholdManager {
	m := &AlertThresholdManager{db: db, alerter: alerter}
	m.load()
	return m
}

// load 启动时把 DB 中已配置的阈值加载到 alerter(未配置的指标用默认阈值)
func (m *AlertThresholdManager) load() {
	var rows []model.AlertThreshold
	if err := m.db.Find(&rows).Error; err != nil {
		return
	}
	for _, r := range rows {
		m.alerter.SetThreshold(r.Metric, r.WarnThreshold, r.CritThreshold)
	}
}

// List 返回全部指标当前生效的阈值(DB 有则返回,无则默认值)
func (m *AlertThresholdManager) List() []model.AlertThreshold {
	thresholds := m.alerter.GetThresholds()
	out := make([]model.AlertThreshold, 0, len(alertMetrics))
	for _, metric := range alertMetrics {
		t, ok := thresholds[metric]
		if !ok {
			continue
		}
		out = append(out, model.AlertThreshold{
			Metric:        metric,
			WarnThreshold: t.Warn,
			CritThreshold: t.Crit,
		})
	}
	return out
}

// Update 更新某指标阈值:校验(warn < crit)→ 存 DB → 热设置 alerter
func (m *AlertThresholdManager) Update(metric string, warn, crit float64) error {
	if !m.validMetric(metric) {
		return errors.New("不支持的指标: " + metric)
	}
	if warn <= 0 || crit <= 0 {
		return errors.New("阈值必须大于 0")
	}
	if warn >= crit {
		return errors.New("warning 阈值必须小于 critical 阈值")
	}

	row := model.AlertThreshold{
		Metric:        metric,
		WarnThreshold: warn,
		CritThreshold: crit,
		UpdatedAt:     time.Now(),
	}
	if err := m.db.Save(&row).Error; err != nil {
		return err
	}
	// 热生效:立即更新 alerter,下次采集即用新阈值
	m.alerter.SetThreshold(metric, warn, crit)
	return nil
}

func (m *AlertThresholdManager) validMetric(metric string) bool {
	for _, m := range alertMetrics {
		if m == metric {
			return true
		}
	}
	return false
}
