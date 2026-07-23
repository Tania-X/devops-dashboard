package main

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
