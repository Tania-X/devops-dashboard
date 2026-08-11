package service

import "testing"

// TestValidateWebhookURL 验证 Webhook 地址校验逻辑（合法/非法/内网拦截）
func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool // true = 允许
	}{
		{"企微官方域名", "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=abc", true},
		{"钉钉官方域名", "https://oapi.dingtalk.com/robot/send?access_token=abc", true},
		{"普通公网域名", "https://example.com/webhook", true},
		{"本机回环地址", "http://127.0.0.1:9999/webhook", false},
		{"localhost", "http://localhost:9999/webhook", false},
		{"私网 10.x", "http://10.0.0.1/webhook", false},
		{"私网 192.168", "http://192.168.1.100/webhook", false},
		{"私网 172.16", "http://172.16.0.1/webhook", false},
		{"非 http 协议", "ftp://example.com/x", false},
		{"缺少主机名", "http://", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebhookURL(tt.url)
			if tt.want && err != nil {
				t.Errorf("应允许 %q，实际报错: %v", tt.url, err)
			}
			if !tt.want && err == nil {
				t.Errorf("应拒绝 %q，实际通过了", tt.url)
			}
		})
	}
}
