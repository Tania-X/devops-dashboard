package service

import (
	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"

	"gorm.io/gorm"
)

type DashboardService struct {
	db              *gorm.DB
	history         *monitor.History
	remoteCollector *monitor.RemoteCollector
	alerter         *monitor.Alerter
}

func NewDashboardService(db *gorm.DB, history *monitor.History, rc *monitor.RemoteCollector, alerter *monitor.Alerter) *DashboardService {
	return &DashboardService{
		db:              db,
		history:         history,
		remoteCollector: rc,
		alerter:         alerter,
	}
}

func (d *DashboardService) GetMetrics() (model.DashboardMetrics, error) {
	var snapshot *monitor.MetricSnapshot
	var err error

	if d.remoteCollector != nil {
		// 有远程 Agent：从 Agent 拉取
		snapshot, err = d.remoteCollector.GetMetrics()
	} else {
		// 无 Agent：本地采集（降级）
		snapshot, err = monitor.Collect()
	}
	if err != nil {
		// 采集失败时回退到假数据，保证前端不白屏
		return model.DashboardMetrics{
			CPU:        model.MetricValue{Current: 0, Status: "normal"},
			Memory:     model.MetricValue{Current: 0, Status: "normal"},
			Disk:       model.MetricValue{Current: 0, Status: "normal"},
			AlertCount: 0,
		}, err
	}
	return model.DashboardMetrics{
		CPU:        model.MetricValue{Current: snapshot.CPUPercent, Status: monitor.Status(snapshot.CPUPercent)},
		Memory:     model.MetricValue{Current: snapshot.MemoryPercent, Status: monitor.Status(snapshot.MemoryPercent)},
		Disk:       model.MetricValue{Current: snapshot.DiskPercent, Status: monitor.Status(snapshot.DiskPercent)},
		AlertCount: d.alerterCount(),
	}, nil
}

// alerterCount 获取当前活跃告警数量
func (d *DashboardService) alerterCount() int {
	if d.alerter == nil {
		return 0
	}
	return len(d.alerter.GetAlerts(100))
}

// GetAlerts 获取告警列表
func (d *DashboardService) GetAlerts(limit int) []model.AlertItem {
	if d.alerter == nil {
		return []model.AlertItem{}
	}
	return d.alerter.GetAlerts(limit)
}

func (d *DashboardService) GetTrend(hours int) (model.DashboardTrend, error) {
	if d.history == nil {
		return model.DashboardTrend{}, nil
	}
	labels, cpuData, memoryData := d.history.Query(hours)
	return model.DashboardTrend{
		TimeLabels: labels,
		CpuData:    cpuData,
		MemoryData: memoryData,
	}, nil
}
