package api

import (
	"net/http"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetDeploymentList(c *gin.Context) {
	var deployments []model.Deployment
	h.db.Order("last_deployed_at DESC").Find(&deployments)
	c.JSON(http.StatusOK, deployments)
}

func (h *Handler) GetDeploymentHistory(c *gin.Context) {
	id := c.Param("id")

	var histories []model.DeploymentHistory
	h.db.Where("deployment_id = ?", id).Order("deployed_at DESC").Find(&histories)

	c.JSON(http.StatusOK, histories)
}
