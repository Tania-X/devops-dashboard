package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
// @Summary     健康检查
// @Description 检查服务运行状态及数据库连通性
// @Tags        System
// @Success     200 {object} map[string]string
// @Router      /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	ok, err := h.services.HealthCheck()
	if !ok || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ok", "db": "disconnected"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "connected"})
	slog.Info("健康检查", "db_status", "connected")
}
