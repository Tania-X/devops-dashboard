package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetDeploymentList 获取部署列表
// @Summary     获取部署列表
// @Description 返回所有应用的部署状态列表
// @Tags        Deployment
// @Success     200 {array} model.Deployment
// @Router      /deployments [get]
func (h *Handler) GetDeploymentList(c *gin.Context) {
	deployments, err := h.services.DeploymentService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deployments)
}

// GetDeploymentHistory 获取部署历史
// @Summary     获取部署历史
// @Description 按应用 ID 获取该应用的所有历史部署记录
// @Tags        Deployment
// @Param       id path string true "应用 ID"
// @Success     200 {array} model.DeploymentHistory
// @Router      /deployments/{id}/history [get]
func (h *Handler) GetDeploymentHistory(c *gin.Context) {
	deployments, err := h.services.DeploymentService.GetDeploymentHistory(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployments)
}
