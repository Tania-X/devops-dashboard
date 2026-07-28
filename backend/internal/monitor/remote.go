package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RemoteCollector 从远程 Agent 拉取指标数据
type RemoteCollector struct {
	baseURL string
	client  *http.Client
}

// NewRemoteCollector 创建远程采集器
// agentAddr 格式为 "ip:port"，如 "192.168.1.100:9100"
func NewRemoteCollector(agentAddr string) *RemoteCollector {
	return &RemoteCollector{
		baseURL: fmt.Sprintf("http://%s", agentAddr),
		client:  &http.Client{},
	}
}

// GetMetrics 从 Agent 拉取系统指标
func (rc *RemoteCollector) GetMetrics() (*MetricSnapshot, error) {
	resp, err := rc.client.Get(rc.baseURL + "/api/metrics")
	if err != nil {
		return nil, fmt.Errorf("连接 Agent 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Agent 返回异常状态码: %d", resp.StatusCode)
	}

	var snapshot MetricSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("解析 Agent 响应失败: %w", err)
	}
	return &snapshot, nil
}

// HealthCheck 检查 Agent 是否存活
func (rc *RemoteCollector) HealthCheck() error {
	resp, err := rc.client.Get(rc.baseURL + "/api/health")
	if err != nil {
		return fmt.Errorf("Agent 健康检查失败: %w", err)
	}
	resp.Body.Close()
	return nil
}
