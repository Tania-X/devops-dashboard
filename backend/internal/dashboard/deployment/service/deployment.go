package service

import (
	deploymentdomain "github.com/Tania-X/devops-dashboard/backend/internal/dashboard/deployment/domain"
	"gorm.io/gorm"
)

// DeploymentService Deployment 应用服务（编排者）。
// 写操作：查聚合根 → 调聚合方法（AddHistory/ChangeStatus）→ 落库，
// 业务规则（状态机、不变式）收敛在 domain，Service 不做业务判断。
type DeploymentService struct {
	db *gorm.DB
}

func NewDeploymentService(db *gorm.DB) *DeploymentService {
	return &DeploymentService{db: db}
}

// List 部署列表（列表页只展示概要，不预加载历史）
func (s *DeploymentService) List() ([]deploymentdomain.Deployment, error) {
	var deployments []deploymentdomain.Deployment
	if err := s.db.Order("last_deployed_at DESC").Find(&deployments).Error; err != nil {
		return nil, err
	}
	return deployments, nil
}

// GetDeploymentHistory 按应用 ID 获取部署历史。
// 读路径：子实体直接按外键查询（子实体的持久化职责在聚合的读模型上，写路径仍经聚合根）。
func (s *DeploymentService) GetDeploymentHistory(deploymentID string) ([]deploymentdomain.DeploymentHistory, error) {
	var historyList []deploymentdomain.DeploymentHistory
	if err := s.db.Where("deployment_id = ?", deploymentID).
		Order("deployed_at DESC").Find(&historyList).Error; err != nil {
		return nil, err
	}
	return historyList, nil
}
