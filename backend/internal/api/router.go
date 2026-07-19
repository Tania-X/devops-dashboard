package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"log/slog"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(requestLogger())

	api := r.Group("/api")
	{
		api.GET("/servers", GetServerList)
		api.GET("/servers/:id", GetServerDetail)
		api.GET("/dashboard/metrics", GetDashboardMetrics)
		api.GET("/dashboard/trend", GetDashboardTrend)
		api.GET("/dashboard/alerts", GetDashboardAlerts)
		api.GET("/logs", GetLogList)
		api.GET("/deployments", GetDeploymentList)
		api.GET("/deployments/:id/history", GetDeploymentHistory)
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
