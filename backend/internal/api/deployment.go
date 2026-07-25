package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetDeploymentList(c *gin.Context) {
	deployments, err := h.services.DeploymentService.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deployments)
}

func (h *Handler) GetDeploymentHistory(c *gin.Context) {
	deployments, err := h.services.DeploymentService.GetDeploymentHistory(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployments)
}
