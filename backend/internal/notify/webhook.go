package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
)

// WebhookNotifier 通过 Webhook 推送告警到钉钉/企业微信机器人
type WebhookNotifier struct {
	kind string // dingtalk | wecom
	url  string // 机器人 Webhook 地址
	http *http.Client
}

// NewWebhookNotifier 创建 Webhook 通知器
// kind: "dingtalk" | "wecom"；url: 机器人 Webhook 完整地址
func NewWebhookNotifier(kind, url string) *WebhookNotifier {
	return &WebhookNotifier{
		kind: kind,
		url:  url,
		http: &http.Client{Timeout: 5 * time.Second},
	}
}

// Name 返回渠道名（用于日志标识）
func (n *WebhookNotifier) Name() string {
	return "webhook-" + n.kind
}

// Send 推送一条告警：格式化 → HTTP POST → 失败重试 3 次（指数退避）
func (n *WebhookNotifier) Send(entry model.AlertItem) error {
	payload, err := n.buildPayload(entry)
	if err != nil {
		return fmt.Errorf("构建消息失败: %w", err)
	}

	body, _ := json.Marshal(payload)
	var lastErr error

	// 重试 3 次：1s / 2s / 4s 退避
	for attempt := 1; attempt <= 3; attempt++ {
		lastErr = n.post(body)
		if lastErr == nil {
			return nil
		}
		slog.Warn("Webhook 推送失败，准备重试", "kind", n.kind, "attempt", attempt, "err", lastErr)
		if attempt < 3 {
			time.Sleep(time.Duration(1<<(attempt-1)) * time.Second)
		}
	}
	return fmt.Errorf("推送失败（已重试 3 次）: %w", lastErr)
}

// post 发送一次 HTTP POST 请求，返回 nil 表示成功
func (n *WebhookNotifier) post(body []byte) error {
	resp, err := n.http.Post(n.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// buildPayload 按渠道类型构建机器人消息格式
func (n *WebhookNotifier) buildPayload(entry model.AlertItem) (map[string]any, error) {
	switch n.kind {
	case "dingtalk":
		return map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]any{
				"title": fmt.Sprintf("【DevOps 告警】%s", entry.Message),
				"text":  n.dingtalkText(entry),
			},
		}, nil
	case "wecom":
		return map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]any{
				"content": n.wecomText(entry),
			},
		}, nil
	default:
		return nil, fmt.Errorf("不支持的渠道类型: %s", n.kind)
	}
}

// levelEmoji 告警级别 → emoji 映射：critical 红 / warning 黄 / info（恢复）绿
func levelEmoji(level string) string {
	switch level {
	case "critical":
		return "🔴"
	case "warning":
		return "🟡"
	case "info":
		return "🟢"
	}
	return "⚪"
}

// dingtalkText 钉钉 markdown 文本
func (n *WebhookNotifier) dingtalkText(entry model.AlertItem) string {
	return fmt.Sprintf(
		"### DevOps Dashboard 告警\n\n- **级别**: %s %s\n- **来源**: %s\n- **时间**: %s\n\n> %s",
		levelEmoji(entry.Level), entry.Level, entry.Source, entry.Time, entry.Message,
	)
}

// wecomText 企业微信 markdown 文本
func (n *WebhookNotifier) wecomText(entry model.AlertItem) string {
	return fmt.Sprintf(
		"**【DevOps 告警】%s**\n> 级别: %s %s\n> 来源: %s\n> 时间: %s\n> 详情: %s",
		entry.Message,
		levelEmoji(entry.Level), entry.Level,
		entry.Source, entry.Time, entry.Message,
	)
}
