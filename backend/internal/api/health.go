package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) HealthCheck(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ok", "db": "disconnected"})
		return
	}
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ok", "db": "disconnected"})
		return
	}
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ok", "db": "disconnected"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "connected"})
	slog.Info("健康检查", "db_status", "connected")
}
