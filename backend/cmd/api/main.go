package main

// @title          DevOps Dashboard API
// @version        1.0
// @description    运维监控仪表盘后端接口
// @termsOfService http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  your-email@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

import (
	"log/slog"
	"os"

	"github.com/Tania-X/devops-dashboard/backend/internal/app"
	"github.com/Tania-X/devops-dashboard/backend/internal/config"
)

func main() {
	cfg := config.Load()

	application := app.New(cfg)
	if err := application.Init(); err != nil {
		slog.Error("应用初始化失败", "error", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		slog.Error("应用运行失败", "error", err)
		os.Exit(1)
	}
}
