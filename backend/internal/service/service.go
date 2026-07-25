package service

import "gorm.io/gorm"

// Services 聚合所有 Service，方便 Handler 一次性获取依赖
type Services struct {
	ServerService     *ServerService
	DeploymentService *DeploymentService
	LogService        *LogService
}

func NewServices(db *gorm.DB) *Services {
	return &Services{
		ServerService:     NewServerService(db),
		DeploymentService: NewDeploymentService(db),
		LogService:        NewLogService(db),
	}
}
