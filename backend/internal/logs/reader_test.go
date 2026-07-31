package logs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLine(t *testing.T) {
	cases := []struct {
		name string
		line string
		want *struct {
			level   string
			content string
			service string
			host    string
		}
		wantNil bool
	}{
		{
			name: "full_http_log",
			line: `time=2026-07-31T12:34:56.789+08:00 level=INFO msg=HTTP service=devops-dashboard env=dev version=dev pid=15840 hostname=DESKTOP-EHF777O method=GET path=/api/metrics status=200 duration_ms=2`,
			want: &struct {
				level   string
				content string
				service string
				host    string
			}{level: "INFO", content: "HTTP", service: "devops-dashboard", host: "DESKTOP-EHF777O"},
		},
		{
			name: "error_log",
			line: `time=2026-07-31T12:35:00.000+08:00 level=ERROR msg=数据库连接失败 service=devops-dashboard`,
			want: &struct {
				level   string
				content string
				service string
				host    string
			}{level: "ERROR", content: "数据库连接失败", service: "devops-dashboard", host: ""},
		},
		{
			name: "warn_log",
			line: `time=2026-07-31T12:36:00.000+08:00 level=WARN msg=采集指标失败 service=devops-dashboard hostname=my-pc`,
			want: &struct {
				level   string
				content string
				service string
				host    string
			}{level: "WARN", content: "采集指标失败", service: "devops-dashboard", host: "my-pc"},
		},
		{
			name: "startup_log",
			line: `time=2026-07-31T12:00:00.000+08:00 level=INFO msg=服务启动 service=devops-dashboard env=dev version=dev pid=1 hostname=server-01`,
			want: &struct {
				level   string
				content string
				service string
				host    string
			}{level: "INFO", content: "服务启动", service: "devops-dashboard", host: "server-01"},
		},
		{
			name:    "empty_msg",
			line:    `time=2026-07-31T12:00:00+08:00 level=INFO msg= service=devops-dashboard`,
			wantNil: true, // msg 为空字符串时应返回 nil
		},
		{
			name:    "no_msg_field",
			line:    `time=2026-07-31T12:00:00+08:00 level=INFO`,
			wantNil: true, // 没有 msg 字段时应返回 nil
		},
		{
			name:    "empty_line",
			line:    "",
			wantNil: true,
		},
		{
			name: "malformed_line",
			line: `some random text without equals`,
			want: &struct {
				level   string
				content string
				service string
				host    string
			}{level: "INFO", content: "", service: "", host: ""},
			wantNil: true, // 解析不到 msg，返回 nil
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseLine(tc.line)

			if tc.wantNil {
				if got != nil {
					t.Errorf("期望 nil，得到 %+v", got)
				}
				return
			}

			if got.Level != tc.want.level {
				t.Errorf("Level = %q, want %q", got.Level, tc.want.level)
			}
			if got.Content != tc.want.content {
				t.Errorf("Content = %q, want %q", got.Content, tc.want.content)
			}
			if got.Service != tc.want.service {
				t.Errorf("Service = %q, want %q", got.Service, tc.want.service)
			}
			if got.SourceHost != tc.want.host {
				t.Errorf("SourceHost = %q, want %q", got.SourceHost, tc.want.host)
			}
		})
	}
}

func TestReader_List(t *testing.T) {
	// 造一个临时日志文件
	dir := t.TempDir()
	logPath := filepath.Join(dir, "app.log")

	content := `time=2026-07-31T12:00:00+08:00 level=INFO msg=服务启动 service=svc
time=2026-07-31T12:01:00+08:00 level=WARN msg=内存使用率偏高 service=svc
time=2026-07-31T12:02:00+08:00 level=ERROR msg=数据库查询超时 service=svc
time=2026-07-31T12:03:00+08:00 level=INFO msg=HTTP service=svc method=GET path=/api/metrics
time=2026-07-31T12:04:00+08:00 level=INFO msg=健康检查通过 service=svc
`
	if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
		t.Fatalf("写测试文件失败: %v", err)
	}

	reader := NewReader(logPath)

	t.Run("full_page", func(t *testing.T) {
		logs, total, err := reader.List(1, 10, "", "")
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 5 {
			t.Fatalf("期望 5 条日志，得到 %d 条", total)
		}
		// 最新的一条是最后一个（12:04 的健康检查）
		if logs[0].Content != "健康检查通过" {
			t.Errorf("第一条日志应为'健康检查通过'，得到 %q", logs[0].Content)
		}
	})

	t.Run("pagination_page1", func(t *testing.T) {
		logs, total, err := reader.List(1, 2, "", "")
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 5 {
			t.Errorf("总数应仍为 5，得到 %d", total)
		}
		if len(logs) != 2 {
			t.Fatalf("pageSize=2 应返回 2 条，得到 %d 条", len(logs))
		}
		// 第一页：最新 2 条
		if logs[0].Content != "健康检查通过" {
			t.Errorf("第一条应为'健康检查通过'，得到 %q", logs[0].Content)
		}
		if logs[1].Content != "HTTP" {
			t.Errorf("第二条应为'HTTP'，得到 %q", logs[1].Content)
		}
	})

	t.Run("pagination_page2", func(t *testing.T) {
		logs, total, err := reader.List(2, 2, "", "")
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 5 {
			t.Errorf("总数应为 5，得到 %d", total)
		}
		if len(logs) != 2 {
			t.Fatalf("pageSize=2 应返回 2 条，得到 %d 条", len(logs))
		}
		// 第二页：接下来 2 条
		if logs[0].Content != "数据库查询超时" {
			t.Errorf("第一条应为'数据库查询超时'，得到 %q", logs[0].Content)
		}
	})

	t.Run("pagination_out_of_range", func(t *testing.T) {
		logs, total, err := reader.List(10, 5, "", "")
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 5 {
			t.Errorf("总数应为 5，得到 %d", total)
		}
		if len(logs) != 0 {
			t.Errorf("超范围页应返回 0 条，得到 %d 条", len(logs))
		}
	})

	t.Run("level_filter", func(t *testing.T) {
		logs, total, err := reader.List(1, 10, "ERROR", "")
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 1 {
			t.Fatalf("ERROR 级别过滤后应为 1 条，得到 %d 条", total)
		}
		if logs[0].Content != "数据库查询超时" {
			t.Errorf("ERROR 日志应为'数据库查询超时'，得到 %q", logs[0].Content)
		}
	})

	t.Run("keyword_search", func(t *testing.T) {
		logs, total, err := reader.List(1, 10, "", "内存")
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 1 {
			t.Fatalf("关键词'内存'应匹配 1 条，得到 %d 条", total)
		}
		if logs[0].Content != "内存使用率偏高" {
			t.Errorf("应为'内存使用率偏高'，得到 %q", logs[0].Content)
		}
	})

	t.Run("no_match", func(t *testing.T) {
		logs, total, err := reader.List(1, 10, "", "nonexistent")
		if err != nil {
			t.Fatalf("List 失败: %v", err)
		}
		if total != 0 {
			t.Errorf("不应匹配任何日志，得到 %d 条", total)
		}
		if len(logs) != 0 {
			t.Errorf("结果应为空，得到 %d 条", len(logs))
		}
	})

	t.Run("file_not_found", func(t *testing.T) {
		reader := NewReader("nonexistent.log")
		logs, total, err := reader.List(1, 10, "", "")
		if err != nil {
			t.Fatalf("文件不存在应不报错，得到: %v", err)
		}
		if total != 0 {
			t.Errorf("文件不存在时总数应为 0，得到 %d", total)
		}
		if len(logs) != 0 {
			t.Errorf("文件不存在时结果应为空，得到 %d 条", len(logs))
		}
	})
}
