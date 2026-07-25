package service

import (
	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"gorm.io/gorm"
)

type DeploymentService struct {
	db *gorm.DB
}

func NewDeploymentService(db *gorm.DB) *DeploymentService {
	return &DeploymentService{db: db}
}

func (d *DeploymentService) List() ([]model.Deployment, error) {

	var deployments []model.Deployment

	query := d.db.Model(&deployments)

	if result := query.Order("last_deployed_at DESC").Find(&deployments); result.Error != nil {
		return deployments, result.Error
	}
	return deployments, nil
}

func (d *DeploymentService) GetDeploymentHistory(deploymentID string) ([]model.DeploymentHistory, error) {
	var historyList []model.DeploymentHistory
	if result := d.db.Where("deployment_id = ?", deploymentID).
		Order("deployed_at DESC").Find(&historyList); result.Error != nil {
		return nil, result.Error
	}
	return historyList, nil
}
