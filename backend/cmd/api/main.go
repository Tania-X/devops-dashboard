package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/api"
	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"
	"github.com/Tania-X/devops-dashboard/backend/internal/repository"
	"github.com/Tania-X/devops-dashboard/backend/pkg/seed"
)

func main() {
	// 初始化结构化日志
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	// 初始化数据库
	db := repository.InitDB("storage/devops.db")

	// 如果数据库为空，自动填充假数据
	seed.SeedIfNeeded(db)

	// 初始化并启动历史数据采集器（每10秒采一次，保留24小时）
	history := monitor.NewHistory(24*time.Hour, 10*time.Second)
	stopCh := history.StartCollector(10 * time.Second)
	api.InitMonitorHistory(history)
	defer close(stopCh)

	// 设置路由
	r := api.SetupRouter()

	// 启动服务
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("服务启动", "address", "http://localhost:"+port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("服务启动失败", "error", err)
		os.Exit(1)
	}
}
