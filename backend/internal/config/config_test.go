package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	cases := []struct {
		name     string
		envVars  map[string]string // 测试前设置的 env
		want     Config
	}{
		{
			name:    "all_defaults",
			envVars: map[string]string{},
			want: Config{
				Port:            "8080",
				DBPath:          "storage/devops.db",
				LogLevel:        "info",
				LogFormat:       "text",
				Env:             "dev",
				HistoryRetain:   "24h",
				HistoryInterval: "10s",
			},
		},
		{
			name: "custom_values",
			envVars: map[string]string{
				"PORT":       "3000",
				"DB_PATH":    "/data/devops.db",
				"LOG_LEVEL":  "debug",
				"LOG_FORMAT": "json",
				"ENV":        "prod",
			},
			want: Config{
				Port:            "3000",
				DBPath:          "/data/devops.db",
				LogLevel:        "debug",
				LogFormat:       "json",
				Env:             "prod",
				HistoryRetain:   "24h",
				HistoryInterval: "10s",
			},
		},
		{
			name: "partial_overrides",
			envVars: map[string]string{
				"PORT":            "9090",
				"HISTORY_INTERVAL": "30s",
			},
			want: Config{
				Port:            "9090",
				DBPath:          "storage/devops.db",
				LogLevel:        "info",
				LogFormat:       "text",
				Env:             "dev",
				HistoryRetain:   "24h",
				HistoryInterval: "30s",
			},
		},
		{
			name: "agent_hosts_single",
			envVars: map[string]string{
				"AGENT_HOSTS": "192.168.1.100:9100",
			},
			want: Config{
				Port:            "8080",
				DBPath:          "storage/devops.db",
				LogLevel:        "info",
				LogFormat:       "text",
				Env:             "dev",
				HistoryRetain:   "24h",
				HistoryInterval: "10s",
				AgentHosts:      []string{"192.168.1.100:9100"},
			},
		},
		{
			name: "agent_hosts_multiple",
			envVars: map[string]string{
				"AGENT_HOSTS": "10.0.0.1:9100,10.0.0.2:9100,10.0.0.3:9101",
			},
			want: Config{
				Port:            "8080",
				DBPath:          "storage/devops.db",
				LogLevel:        "info",
				LogFormat:       "text",
				Env:             "dev",
				HistoryRetain:   "24h",
				HistoryInterval: "10s",
				AgentHosts:      []string{"10.0.0.1:9100", "10.0.0.2:9100", "10.0.0.3:9101"},
			},
		},
		{
			name: "agent_hosts_empty",
			envVars: map[string]string{
				"AGENT_HOSTS": "",
			},
			want: Config{
				Port:            "8080",
				DBPath:          "storage/devops.db",
				LogLevel:        "info",
				LogFormat:       "text",
				Env:             "dev",
				HistoryRetain:   "24h",
				HistoryInterval: "10s",
				AgentHosts:      nil, // 空字符串不应产生空切片
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 清理所有相关 env 再逐个设置（当前 case 需要的）
			for _, key := range []string{
				"PORT", "DB_PATH", "LOG_LEVEL", "LOG_FORMAT", "ENV",
				"HISTORY_RETAIN", "HISTORY_INTERVAL", "AGENT_HOSTS",
			} {
				os.Unsetenv(key)
			}
			for k, v := range tc.envVars {
				os.Setenv(k, v)
			}

			got := Load()

			if got.Port != tc.want.Port {
				t.Errorf("Port = %q, want %q", got.Port, tc.want.Port)
			}
			if got.DBPath != tc.want.DBPath {
				t.Errorf("DBPath = %q, want %q", got.DBPath, tc.want.DBPath)
			}
			if got.LogLevel != tc.want.LogLevel {
				t.Errorf("LogLevel = %q, want %q", got.LogLevel, tc.want.LogLevel)
			}
			if got.LogFormat != tc.want.LogFormat {
				t.Errorf("LogFormat = %q, want %q", got.LogFormat, tc.want.LogFormat)
			}
			if got.Env != tc.want.Env {
				t.Errorf("Env = %q, want %q", got.Env, tc.want.Env)
			}
			if got.HistoryRetain != tc.want.HistoryRetain {
				t.Errorf("HistoryRetain = %q, want %q", got.HistoryRetain, tc.want.HistoryRetain)
			}
			if got.HistoryInterval != tc.want.HistoryInterval {
				t.Errorf("HistoryInterval = %q, want %q", got.HistoryInterval, tc.want.HistoryInterval)
			}
			if len(got.AgentHosts) != len(tc.want.AgentHosts) {
				t.Errorf("AgentHosts len = %d, want %d", len(got.AgentHosts), len(tc.want.AgentHosts))
			} else {
				for i := range got.AgentHosts {
					if got.AgentHosts[i] != tc.want.AgentHosts[i] {
						t.Errorf("AgentHosts[%d] = %q, want %q", i, got.AgentHosts[i], tc.want.AgentHosts[i])
					}
				}
			}
		})
	}
}

func TestLoad_AgentHosts_LeadingTrailingSpace(t *testing.T) {
	// 清理
	for _, key := range []string{
		"PORT", "DB_PATH", "LOG_LEVEL", "LOG_FORMAT", "ENV",
		"HISTORY_RETAIN", "HISTORY_INTERVAL", "AGENT_HOSTS",
	} {
		os.Unsetenv(key)
	}
	os.Setenv("AGENT_HOSTS", " 10.0.0.1:9100 , 10.0.0.2:9100 ")

	got := Load()

	// strings.Split(",") 会保留空格，本测试记录当前行为
	// 如果后续代码加了 TrimSpace，这个测试应该随实现一起改
	expected := []string{" 10.0.0.1:9100 ", " 10.0.0.2:9100 "}
	if strings.Join(got.AgentHosts, "|") != strings.Join(expected, "|") {
		t.Errorf("AgentHosts = %v, want %v", got.AgentHosts, expected)
	}
}
