package service

import (
	serversvc "github.com/Tania-X/devops-dashboard/backend/internal/dashboard/server/service"
	usersvc "github.com/Tania-X/devops-dashboard/backend/internal/dashboard/user/service"
	"github.com/Tania-X/devops-dashboard/backend/internal/logs"
	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"
	"github.com/Tania-X/devops-dashboard/backend/internal/notify"
	"gorm.io/gorm"
)

// Services 聚合所有 Service，方便 Handler 一次性获取依赖
type Services struct {
	db *gorm.DB

	ServerService     *serversvc.ServerService
	DeploymentService *DeploymentService
	LogService        *LogService
	DashboardService  *DashboardService
	MonitorService    *MonitorService
	AgentService      *AgentService
	AuthService       *AuthService
	UserService       *usersvc.UserService
	WebhookManager    *WebhookManager
	AuditService      *AuditService
}

// HealthCheck 检查数据库连通性
func (s *Services) HealthCheck() (bool, error) {
	sqlDB, err := s.db.DB()
	if err != nil {
		return false, err
	}
	if err := sqlDB.Ping(); err != nil {
		return false, err
	}
	return true, nil
}

func NewServices(db *gorm.DB, history *monitor.History, rc *monitor.RemoteCollector, alerter *monitor.Alerter, bus *notify.AlertBus, jwtSecret string, agentSecretKey string, agentBinPath string) *Services {
	logReader := logs.NewReader("storage/logs/app.log")
	return &Services{
		db:                db,
		ServerService:     serversvc.NewServerService(db),
		DeploymentService: NewDeploymentService(db),
		LogService:        NewLogService(logReader),
		DashboardService:  NewDashboardService(db, history, rc, alerter),
		MonitorService:    NewMonitorService(db),
		AgentService:      NewAgentService(db, agentSecretKey, agentBinPath),
		AuthService:       NewAuthService(db, jwtSecret),
		UserService:       usersvc.NewUserService(db),
		WebhookManager:    NewWebhookManager(db, bus),
		AuditService:      NewAuditService(db),
	}
}
