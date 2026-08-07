package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
)

func sampleAlert() model.AlertItem {
	return model.AlertItem{
		ID:      "alert-001",
		Level:   "critical",
		Message: "CPU 使用率 87.5% — 超过 critical 阈值 (80%)",
		Source:  "localhost",
		Time:    "08-07 15:30",
	}
}

// TestBuildPayload_Dingtalk 验证钉钉消息格式
func TestBuildPayload_Dingtalk(t *testing.T) {
	n := NewWebhookNotifier("dingtalk", "http://example.com")
	payload, err := n.buildPayload(sampleAlert())
	if err != nil {
		t.Fatalf("buildPayload 报错: %v", err)
	}

	md, ok := payload["markdown"].(map[string]any)
	if !ok {
		t.Fatal("payload 缺少 markdown 字段")
	}
	if payload["msgtype"] != "markdown" {
		t.Errorf("msgtype = %v, want markdown", payload["msgtype"])
	}
	text := md["text"].(string)
	if !contains(text, "🔴") {
		t.Errorf("critical 级别应映射 🔴 emoji，实际 text: %s", text)
	}
	if !contains(text, "87.5%") {
		t.Errorf("text 应包含告警消息，实际: %s", text)
	}
}

// TestBuildPayload_Wecom 验证企业微信消息格式
func TestBuildPayload_Wecom(t *testing.T) {
	n := NewWebhookNotifier("wecom", "http://example.com")
	payload, err := n.buildPayload(sampleAlert())
	if err != nil {
		t.Fatalf("buildPayload 报错: %v", err)
	}

	md, ok := payload["markdown"].(map[string]any)
	if !ok {
		t.Fatal("payload 缺少 markdown 字段")
	}
	content := md["content"].(string)
	if !contains(content, "【DevOps 告警】") {
		t.Errorf("企微 content 应含标题，实际: %s", content)
	}
	if !contains(content, "🔴") {
		t.Errorf("critical 级别应映射 🔴 emoji，实际: %s", content)
	}
}

// TestBuildPayload_UnsupportedKind 不支持的渠道类型应报错
func TestBuildPayload_UnsupportedKind(t *testing.T) {
	n := NewWebhookNotifier("slack", "http://example.com")
	_, err := n.buildPayload(sampleAlert())
	if err == nil {
		t.Fatal("slack 渠道应返回错误")
	}
}

// TestSend_Success 正常发送（200）应成功
func TestSend_Success(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := NewWebhookNotifier("dingtalk", server.URL)
	if err := n.Send(sampleAlert()); err != nil {
		t.Fatalf("Send 应成功，实际报错: %v", err)
	}
	if gotBody["msgtype"] != "markdown" {
		t.Errorf("收到 msgtype = %v", gotBody["msgtype"])
	}
}

// TestSend_Retry 服务端一直 500，应重试后失败（耗时约 7s）
func TestSend_Retry(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := NewWebhookNotifier("wecom", server.URL)
	start := time.Now()
	err := n.Send(sampleAlert())
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("500 响应应返回错误")
	}
	if attempts != 3 {
		t.Errorf("应重试 3 次，实际 %d 次", attempts)
	}
	if elapsed < 3*time.Second {
		t.Errorf("重试退避时间不足，实际耗时 %v", elapsed)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
