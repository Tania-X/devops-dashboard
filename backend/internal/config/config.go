package config

import "os"

// Config 聚合所有应用配置
// 从环境变量读取，并提供合理的本地开发默认值
type Config struct {
	Port            string
	DBPath          string
	LogLevel        string
	LogFormat       string // 日志格式：text（人类可读）或 json（给 Loki）
	Env             string // 运行环境：dev / prod
	HistoryRetain   string // 历史数据保留时长，如 "24h"
	HistoryInterval string // 历史数据采集间隔，如 "10s"
}

// Load 从环境变量加载配置
func Load() Config {
	cfg := Config{
		Port:            os.Getenv("PORT"),
		DBPath:          os.Getenv("DB_PATH"),
		LogLevel:        os.Getenv("LOG_LEVEL"),
		LogFormat:       os.Getenv("LOG_FORMAT"),
		Env:             os.Getenv("ENV"),
		HistoryRetain:   os.Getenv("HISTORY_RETAIN"),
		HistoryInterval: os.Getenv("HISTORY_INTERVAL"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "storage/devops.db"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.LogFormat == "" {
		cfg.LogFormat = "text"
	}
	if cfg.Env == "" {
		cfg.Env = "dev"
	}
	if cfg.HistoryRetain == "" {
		cfg.HistoryRetain = "24h"
	}
	if cfg.HistoryInterval == "" {
		cfg.HistoryInterval = "10s"
	}

	return cfg
}
