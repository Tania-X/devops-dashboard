package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/api"
	"github.com/Tania-X/devops-dashboard/backend/internal/config"
	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"
	"github.com/Tania-X/devops-dashboard/backend/internal/repository"
	"github.com/Tania-X/devops-dashboard/backend/internal/service"
	"github.com/Tania-X/devops-dashboard/backend/pkg/seed"
	"gorm.io/gorm"
)

// App 聚合应用生命周期所需的所有依赖
// 负责初始化、启动、优雅关闭
type App struct {
	cfg      config.Config
	db       *gorm.DB
	history  *monitor.History
	services *service.Services

	server *http.Server
	stopCh chan struct{}
}

// New 创建 App 实例（此时还未初始化任何依赖）
func New(cfg config.Config) *App {
	return &App{
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
}

// Init 按顺序初始化 logger、db、seed、monitor history、http server
func (a *App) Init() error {
	a.setupLogger()

	db, err := repository.InitDB(a.cfg.DBPath)
	if err != nil {
		return fmt.Errorf("init db failed: %w", err)
	}
	a.db = db

	seed.SeedIfNeeded(a.db)

	retain, err := time.ParseDuration(a.cfg.HistoryRetain)
	if err != nil {
		return fmt.Errorf("invalid HISTORY_RETAIN %q: %w", a.cfg.HistoryRetain, err)
	}
	interval, err := time.ParseDuration(a.cfg.HistoryInterval)
	if err != nil {
		return fmt.Errorf("invalid HISTORY_INTERVAL %q: %w", a.cfg.HistoryInterval, err)
	}

	a.history = monitor.NewHistory(retain, interval)
	a.stopCh = a.history.StartCollector(interval)

	handler := api.NewHandler(a.db, a.history, a.services)
	a.server = &http.Server{
		Addr:    ":" + a.cfg.Port,
		Handler: handler.SetupRouter(),
	}

	return nil
}

// Run 阻塞运行 HTTP 服务，并监听系统信号实现优雅关闭
func (a *App) Run() error {
	if a.server == nil {
		return errors.New("app not initialized, call Init() first")
	}

	slog.Info("服务启动", "address", "http://localhost:"+a.cfg.Port)

	errCh := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
		return a.shutdown()
	}
}

// shutdown 关闭后台采集器并优雅停止 HTTP 服务
func (a *App) shutdown() error {
	slog.Info("服务正在关闭...")

	close(a.stopCh)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	slog.Info("服务已关闭")
	return nil
}

// setupLogger 根据配置初始化 slog
//
//	text 格式便于本地开发调试阅读；json 格式便于接入 Loki 等日志系统
func (a *App) setupLogger() {
	level := slog.LevelInfo
	switch a.cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}

	// 全局字段：每条日志都会带上，用于区分服务、环境、进程
	version := os.Getenv("VERSION")
	if version == "" {
		version = "dev"
	}
	attrs := []slog.Attr{
		slog.String("service", "devops-dashboard"),
		slog.String("env", a.cfg.Env),
		slog.String("version", version),
		slog.Int("pid", os.Getpid()),
	}
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		attrs = append(attrs, slog.String("hostname", hostname))
	}

	var handler slog.Handler
	switch a.cfg.LogFormat {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts).WithAttrs(attrs)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts).WithAttrs(attrs)
	}

	slog.SetDefault(slog.New(handler))
}
