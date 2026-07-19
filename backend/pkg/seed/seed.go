package seed

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"gorm.io/gorm"
)

func SeedIfNeeded(db *gorm.DB) {
	var count int64
	db.Model(&model.Server{}).Count(&count)
	if count > 0 {
		slog.Info("数据库已有数据，跳过初始化")
		return
	}

	slog.Info("开始初始化数据...")

	seedServers(db)
	seedLogs(db)
	seedDeployments(db)

	slog.Info("数据初始化完成", "servers", 35, "logs", 200, "deployments", 15)
}

func seedServers(db *gorm.DB) {
	osList := []string{"CentOS 7.9", "Ubuntu 22.04", "Debian 12", "Rocky Linux 9", "AlmaLinux 8"}
	statusList := []string{"running", "stopped", "maintenance"}

	servers := make([]model.Server, 35)
	for i := 0; i < 35; i++ {
		id := fmt.Sprintf("srv-%03d", i+1)
		servers[i] = model.Server{
			ID:       id,
			Hostname: fmt.Sprintf("%s-%s-%02d", randomWord(), randomWord(), i+1),
			IP:       fmt.Sprintf("192.168.1.%d", 10+i),
			OS:       osList[rand.Intn(len(osList))],
			CpuCores: []int{4, 8, 16, 32, 64}[rand.Intn(5)],
			MemoryGb: []int{8, 16, 32, 64, 128}[rand.Intn(5)],
			Status:   statusList[rand.Intn(len(statusList))],
			Uptime:   fmt.Sprintf("%dd %dh %dm", rand.Intn(365), rand.Intn(24), rand.Intn(60)),
		}

		mounts := []string{"/", "/data", "/var/log", "/backup"}
		for _, mount := range mounts {
			totalGb := []int{100, 250, 500, 1000, 2000}[rand.Intn(5)]
			usedGb := int(float64(totalGb) * (0.2 + rand.Float64()*0.75))
			servers[i].DiskPartitions = append(servers[i].DiskPartitions, model.DiskPartition{
				ServerID: id,
				Mount:    mount,
				TotalGb:  totalGb,
				UsedGb:   usedGb,
			})
		}

		names := []string{"eth0", "eth1", "lo"}
		for _, name := range names {
			ip := "127.0.0.1"
			if name != "lo" {
				ip = fmt.Sprintf("192.168.%d.%d", rand.Intn(256), rand.Intn(256))
			}
			servers[i].NetworkInterfaces = append(servers[i].NetworkInterfaces, model.NetworkInterface{
				ServerID: id,
				Name:     name,
				IP:       ip,
				Mac:      fmt.Sprintf("00:16:3e:%02x:%02x:%02x", rand.Intn(256), rand.Intn(256), rand.Intn(256)),
			})
		}
	}

	if err := db.Create(&servers).Error; err != nil {
		slog.Error("seed servers failed", "error", err)
	}
}

func seedLogs(db *gorm.DB) {
	levels := []string{"INFO", "WARN", "ERROR"}
	services := []string{"api-gateway", "user-service", "order-service", "payment-service", "notification-service", "auth-service", "log-collector", "monitor-agent"}
	templates := map[string][]string{
		"INFO": {
			"Request processed successfully",
			"User login: user_%s",
			"Health check passed",
			"Scheduled task completed",
			"Config reloaded from remote",
		},
		"WARN": {
			"High memory usage detected: %d%%",
			"Slow query detected: %dms",
			"Connection pool approaching limit",
			"Disk usage above 80%%",
		},
		"ERROR": {
			"Database connection timeout after %dms",
			"NullPointerException in %s",
			"Failed to connect to upstream",
			"Message queue consumer crashed",
		},
	}

	logs := make([]model.Log, 200)
	for i := 0; i < 200; i++ {
		level := levels[rand.Intn(len(levels))]
		service := services[rand.Intn(len(services))]
		template := templates[level][rand.Intn(len(templates[level]))]
		content := fmt.Sprintf(template, rand.Intn(100), service)

		logs[i] = model.Log{
			ID:         fmt.Sprintf("log-%05d", i+1),
			Time:       time.Now().Add(-time.Duration(rand.Intn(7*24)) * time.Hour).Format("2006-01-02 15:04:05"),
			Level:      level,
			Service:    service,
			Content:    content,
			SourceHost: fmt.Sprintf("srv-%03d", rand.Intn(35)+1),
			LogPath:    fmt.Sprintf("/var/log/%s/app.log", service),
			TraceID:    fmt.Sprintf("%016x", rand.Int63()),
		}
	}

	if err := db.Create(&logs).Error; err != nil {
		slog.Error("seed logs failed", "error", err)
	}
}

func seedDeployments(db *gorm.DB) {
	appNames := []string{"api-gateway", "user-service", "order-service", "payment-service", "notification-service", "auth-service", "log-collector", "monitor-agent", "config-server", "cache-proxy", "search-engine", "report-generator", "data-sync", "file-storage", "web-frontend"}
	envs := []string{"dev", "test", "prod"}
	statuses := []string{"pending", "deploying", "success", "failed"}
	historyStatuses := []string{"success", "failed"}

	for i := 0; i < 15; i++ {
		id := fmt.Sprintf("app-%03d", i+1)
		appName := appNames[i]
		version := fmt.Sprintf("v2.%d.%d", rand.Intn(5), rand.Intn(10))

		deployment := model.Deployment{
			ID:             id,
			AppName:        appName,
			Version:        version,
			Env:            envs[rand.Intn(len(envs))],
			Status:         statuses[rand.Intn(len(statuses))],
			LastDeployedAt: time.Now().Add(-time.Duration(rand.Intn(72)) * time.Hour),
		}

		if err := db.Create(&deployment).Error; err != nil {
			slog.Error("seed deployment failed", "error", err)
			continue
		}

		historyCount := 3 + rand.Intn(6)
		histories := make([]model.DeploymentHistory, historyCount)
		for j := 0; j < historyCount; j++ {
			histories[j] = model.DeploymentHistory{
				DeploymentID: id,
				Version:      fmt.Sprintf("v2.%d.%d", rand.Intn(5), rand.Intn(10)),
				Operator:     fmt.Sprintf("operator-%d", rand.Intn(10)+1),
				DurationSec:  30 + rand.Intn(570),
				Status:       historyStatuses[rand.Intn(len(historyStatuses))],
				DeployedAt:   time.Now().Add(-time.Duration(j*rand.Intn(24)+rand.Intn(24)) * time.Hour),
			}
		}

		if err := db.Create(&histories).Error; err != nil {
			slog.Error("seed deployment history failed", "error", err)
		}
	}
}

func randomWord() string {
	words := []string{"alpha", "beta", "gamma", "delta", "prod", "web", "api", "db", "cache", "queue"}
	return words[rand.Intn(len(words))]
}
