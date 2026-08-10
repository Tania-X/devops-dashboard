package api

import (

	// ... 现有 import
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// 导入自动生成的 docs 包
	_ "github.com/Tania-X/devops-dashboard/backend/docs"

	"net/http"
	"time"

	"log/slog"

	"github.com/Tania-X/devops-dashboard/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// Handler 聚合所有 handler 所需的依赖
// 通过 NewHandler 注入，避免使用全局变量
type Handler struct {
	services *service.Services
}

// NewHandler 创建 Handler 实例
func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

// SetupRouter 配置并返回 Gin 路由引擎
func (h *Handler) SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(RecoveryMiddleware())
	r.Use(corsMiddleware())
	r.Use(requestLogger())

	// Swagger UI（只在非生产环境注册）
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		// 公开路由（无需认证）
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
		api.GET("/health", h.HealthCheck)

		// 认证路由（无需认证）
		api.POST("/auth/login", h.Login)
		api.POST("/auth/logout", h.Logout)

		// 需要认证的路由（权限点驱动：每个路由标注 RequirePermission(obj, act)）
		auth := api.Group("")
		auth.Use(h.AuthMiddleware())
		{
			auth.GET("/auth/me", h.GetMe)

			// Agent 管理路由（admin / operator）
			auth.GET("/agents", RequirePermission("agent", "read"), h.GetAgentList)
			auth.POST("/agents", RequirePermission("agent", "create"), h.CreateAgent)
			auth.PUT("/agents/:id", RequirePermission("agent", "update"), h.UpdateAgent)
			auth.DELETE("/agents/:id", RequirePermission("agent", "delete"), h.DeleteAgent)
			auth.POST("/agents/:id/deploy", RequirePermission("agent", "deploy"), h.DeployAgent)
			auth.POST("/agents/:id/stop", RequirePermission("agent", "stop"), h.StopAgent)
			auth.GET("/agents/:id/status", RequirePermission("agent", "read"), h.CheckAgentStatus)

			// 用户管理路由（仅 admin）
			auth.GET("/users", RequirePermission("user", "read"), h.GetUserList)
			auth.POST("/users", RequirePermission("user", "create"), h.CreateUser)
			auth.PUT("/users/:id", RequirePermission("user", "update"), h.UpdateUser)
			auth.DELETE("/users/:id", RequirePermission("user", "delete"), h.DeleteUser)

			// 告警 Webhook 配置（仅 admin）
			auth.GET("/settings/webhook", RequirePermission("webhook", "read"), h.GetWebhookConfig)
			auth.PUT("/settings/webhook", RequirePermission("webhook", "update"), h.UpdateWebhookConfig)
			auth.POST("/settings/webhook/test", RequirePermission("webhook", "test"), h.TestWebhookConfig)
		}
	}
	r.NoRoute(func(c *gin.Context) {
		ErrorJSON(c, http.StatusNotFound, "resource not exist")
	})

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
