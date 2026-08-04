package config

import (
	"os"
	"strings"
)

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
	AgentHosts      []string // Agent 目标地址列表
	AgentSecretKey  string   // Agent 密码加密密钥
	AgentBinPath    string   // Agent 二进制文件路径
	JwtSecret       string   // JWT 签名密钥
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

	hosts := os.Getenv("AGENT_HOSTS")
	if hosts != "" {
		cfg.AgentHosts = strings.Split(hosts, ",")
	}

	cfg.AgentSecretKey = os.Getenv("AGENT_SECRET_KEY")
	if cfg.AgentSecretKey == "" {
		cfg.AgentSecretKey = "devops-dashboard-secret-key-32byte!"
	}

	cfg.AgentBinPath = os.Getenv("AGENT_BIN_PATH")
	if cfg.AgentBinPath == "" {
		cfg.AgentBinPath = "bin/agent-linux-amd64"
	}

	cfg.JwtSecret = os.Getenv("JWT_SECRET")
	if cfg.JwtSecret == "" {
		cfg.JwtSecret = "devops-dashboard-jwt-secret-key"
	}

	return cfg
}
