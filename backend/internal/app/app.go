package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/api"
	"github.com/Tania-X/devops-dashboard/backend/internal/authz"
	"github.com/Tania-X/devops-dashboard/backend/internal/config"
	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"
	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/Tania-X/devops-dashboard/backend/internal/notify"
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
	recorder *service.AlertRecorder

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

	// Casbin 授权引擎初始化：加载 PERM 模型 + 策略 seed（admin/viewer/operator 预置）
	if err := authz.Init(a.db); err != nil {
		return fmt.Errorf("init authz failed: %w", err)
	}

	retain, err := time.ParseDuration(a.cfg.HistoryRetain)
	if err != nil {
		return fmt.Errorf("invalid HISTORY_RETAIN %q: %w", a.cfg.HistoryRetain, err)
	}
	interval, err := time.ParseDuration(a.cfg.HistoryInterval)
	if err != nil {
		return fmt.Errorf("invalid HISTORY_INTERVAL %q: %w", a.cfg.HistoryInterval, err)
	}

	// 创建告警评估器
	alerter := monitor.NewAlerter()
	slog.Info("告警引擎已启动", "maxAlerts", 20)

	// 告警总线：Alerter 产生告警 → channel → Webhook 通知器（异步，不阻塞采集）
	bus := notify.NewAlertBus()
	bus.Run()

	// 告警历史落库器（异步）:告警同时进总线(推送)与落库(历史查询)
	recorder := service.NewAlertRecorder(a.db)
	a.recorder = recorder
	alerter.OnAlert = func(e model.AlertItem) {
		bus.Publish(e)
		recorder.Record(e)
	}

	a.history = monitor.NewHistory(retain, interval, alerter)
	a.stopCh = a.history.StartCollector(interval)

	// 如果配置了 AGENT_HOSTS，创建远程采集器
	var rc *monitor.RemoteCollector
	if len(a.cfg.AgentHosts) > 0 {
		rc = monitor.NewRemoteCollector(a.cfg.AgentHosts[0])
		slog.Info("使用远程采集模式", "agent", a.cfg.AgentHosts[0])
	} else {
		slog.Info("使用本地采集模式（未配置 AGENT_HOSTS）")
	}

	a.services = service.NewServices(a.db, a.history, rc, alerter, bus, recorder, a.cfg.JwtSecret, a.cfg.AgentSecretKey, a.cfg.AgentBinPath)

	handler := api.NewHandler(a.services)
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

	// 停止告警落库器并等待消费完成(防 goroutine 泄漏)
	if a.recorder != nil {
		a.recorder.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	slog.Info("服务已关闭")
	return nil
}

// setupLogger 根据配置初始化 slog，同时写入 stdout 和文件
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

	// 双写：stdout + 日志文件（供 /api/logs 读取）
	logDir := "storage/logs"
	logFile := logDir + "/app.log"
	os.MkdirAll(logDir, 0755)
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	multiWriter := io.MultiWriter(os.Stdout)
	if err == nil {
		multiWriter = io.MultiWriter(os.Stdout, f)
	} else {
		slog.Warn("无法创建日志文件，仅输出到 stdout", "err", err)
	}

	var handler slog.Handler
	switch a.cfg.LogFormat {
	case "json":
		handler = slog.NewJSONHandler(multiWriter, opts).WithAttrs(attrs)
	default:
		handler = slog.NewTextHandler(multiWriter, opts).WithAttrs(attrs)
	}

	slog.SetDefault(slog.New(handler))
}
