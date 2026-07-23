package api

import (
	"net/http"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log/slog"
)

// Handler 聚合所有 handler 所需的依赖
// 通过 NewHandler 注入，避免使用全局变量
type Handler struct {
	db      *gorm.DB
	history *monitor.History
}

// NewHandler 创建 Handler 实例
func NewHandler(db *gorm.DB, history *monitor.History) *Handler {
	return &Handler{
		db:      db,
		history: history,
	}
}

// SetupRouter 配置并返回 Gin 路由引擎
func (h *Handler) SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(requestLogger())

	api := r.Group("/api")
	{
		api.GET("/servers", h.GetServerList)
		api.GET("/servers/:id", h.GetServerDetail)
		api.GET("/dashboard/metrics", h.GetDashboardMetrics)
		api.GET("/dashboard/trend", h.GetDashboardTrend)
		api.GET("/dashboard/alerts", h.GetDashboardAlerts)
		api.GET("/logs", h.GetLogList)
		api.GET("/deployments", h.GetDeploymentList)
		api.GET("/deployments/:id/history", h.GetDeploymentHistory)
		api.GET("/monitor/processes", h.GetProcessList)
		api.GET("/monitor/processes/:pid", h.GetProcessDetail)
		api.GET("/monitor/host", h.GetHostInfo)
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		slog.Info("HTTP",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}
