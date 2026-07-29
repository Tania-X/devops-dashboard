package service

import (
	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"
	"gorm.io/gorm"
)

// Services 聚合所有 Service，方便 Handler 一次性获取依赖
type Services struct {
	ServerService     *ServerService
	DeploymentService *DeploymentService
	LogService        *LogService
	DashboardService  *DashboardService
	MonitorService    *MonitorService
}

func NewServices(db *gorm.DB, history *monitor.History, rc *monitor.RemoteCollector, alerter *monitor.Alerter) *Services {
	return &Services{
		ServerService:     NewServerService(db),
		DeploymentService: NewDeploymentService(db),
		LogService:        NewLogService(db),
		DashboardService:  NewDashboardService(db, history, rc, alerter),
		MonitorService:    NewMonitorService(db),
	}
}
