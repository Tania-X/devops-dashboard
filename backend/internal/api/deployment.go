package api

import (
	"net/http"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/Tania-X/devops-dashboard/backend/internal/repository"
	"github.com/gin-gonic/gin"
)

func GetDeploymentList(c *gin.Context) {
	var deployments []model.Deployment
	repository.DB.Order("last_deployed_at DESC").Find(&deployments)
	c.JSON(http.StatusOK, deployments)
}

func GetDeploymentHistory(c *gin.Context) {
	id := c.Param("id")

	var histories []model.DeploymentHistory
	repository.DB.Where("deployment_id = ?", id).Order("deployed_at DESC").Find(&histories)

	c.JSON(http.StatusOK, histories)
}
