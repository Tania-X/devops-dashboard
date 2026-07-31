package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Tania-X/devops-dashboard/backend/internal/logs"
	"github.com/Tania-X/devops-dashboard/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func TestGetLogList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 造一个临时日志文件
	dir := t.TempDir()
	logPath := dir + "/app.log"

	content := `time=2026-07-31T12:00:00+08:00 level=INFO msg=服务启动 service=svc
time=2026-07-31T12:01:00+08:00 level=WARN msg=内存使用率偏高 service=svc
time=2026-07-31T12:05:00+08:00 level=ERROR msg=数据库查询超时 service=myapp
`
	os.WriteFile(logPath, []byte(content), 0644)

	// 组装 Handler（只注入 LogService，其他 nil）
	svcs := &service.Services{
		LogService: service.NewLogService(logs.NewReader(logPath)),
	}
	h := NewHandler(nil, nil, svcs)
	router := h.SetupRouter()

	cases := []struct {
		name       string
		path       string
		wantStatus int
		wantTotal  int64
	}{
		{
			name:       "all_logs",
			path:       "/api/logs?page=1&pageSize=10",
			wantStatus: 200,
			wantTotal:  3,
		},
		{
			name:       "filter_error",
			path:       "/api/logs?level=ERROR",
			wantStatus: 200,
			wantTotal:  1,
		},
		{
			name:       "keyword_search",
			path:       "/api/logs?keyword=内存",
			wantStatus: 200,
			wantTotal:  1,
		},
		{
			name:       "no_match",
			path:       "/api/logs?keyword=nonexistent",
			wantStatus: 200,
			wantTotal:  0,
		},
		{
			name:       "page2",
			path:       "/api/logs?page=2&pageSize=2",
			wantStatus: 200,
			wantTotal:  3, // total 不变，但 list 只 1 条
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("状态码 = %d, want %d", w.Code, tc.wantStatus)
			}

			var body struct {
				List     []map[string]any `json:"list"`
				Total    int64            `json:"total"`
				Page     int              `json:"page"`
				PageSize int              `json:"pageSize"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("JSON 解析失败: %v", err)
			}

			if body.Total != tc.wantTotal {
				t.Errorf("Total = %d, want %d", body.Total, tc.wantTotal)
			}
		})
	}
}

// 注意：NewHandler(nil, nil, svcs) 中 db 和 history 为 nil，但只要只访问
// /api/logs 就不会触发 nil dereference——GetLogList 只用 h.services.LogService。
