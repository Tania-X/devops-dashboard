package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/Tania-X/devops-dashboard/backend/internal/monitor"
	"github.com/gin-gonic/gin"
)

func GetDashboardMetrics(c *gin.Context) {
	snapshot, err := monitor.Collect()
	if err != nil {
		// 采集失败时回退到假数据，保证前端不白屏
		c.JSON(http.StatusOK, model.DashboardMetrics{
			CPU:        model.MetricValue{Current: 0, Status: "normal"},
			Memory:     model.MetricValue{Current: 0, Status: "normal"},
			Disk:       model.MetricValue{Current: 0, Status: "normal"},
			AlertCount: 0,
		})
		return
	}

	c.JSON(http.StatusOK, model.DashboardMetrics{
		CPU:        model.MetricValue{Current: snapshot.CPUPercent, Status: monitor.Status(snapshot.CPUPercent)},
		Memory:     model.MetricValue{Current: snapshot.MemoryPercent, Status: monitor.Status(snapshot.MemoryPercent)},
		Disk:       model.MetricValue{Current: snapshot.DiskPercent, Status: monitor.Status(snapshot.DiskPercent)},
		AlertCount: 0, // 告警数暂未接入真实数据源
	})
}

func GetDashboardTrend(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "6"))
	if hours < 1 || hours > 24 {
		hours = 6
	}

	// 如果 history 未初始化，返回空数据
	if trendHistory == nil {
		c.JSON(http.StatusOK, model.DashboardTrend{
			TimeLabels: []string{},
			CpuData:    []float64{},
			MemoryData: []float64{},
		})
		return
	}

	labels, cpuData, memoryData := trendHistory.Query(hours)
	c.JSON(http.StatusOK, model.DashboardTrend{
		TimeLabels: labels,
		CpuData:    cpuData,
		MemoryData: memoryData,
	})
}

func GetDashboardAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit > 20 {
		limit = 20
	}

	alerts := []model.AlertItem{
		{
			ID:      "alert-001",
			Level:   "critical",
			Message: "服务器 srv-012 磁盘使用率超过 90%，当前 93%",
			Source:  "srv-012 (192.168.1.45)",
			Time:    time.Now().Add(-15 * time.Minute).Format("01-02 15:04"),
		},
		{
			ID:      "alert-002",
			Level:   "warning",
			Message: "api-gateway 服务响应时间 P99 超过 500ms",
			Source:  "api-gateway",
			Time:    time.Now().Add(-42 * time.Minute).Format("01-02 15:04"),
		},
		{
			ID:      "alert-003",
			Level:   "warning",
			Message: "支付服务 payment-service 内存使用率持续上升",
			Source:  "payment-service",
			Time:    time.Now().Add(-78 * time.Minute).Format("01-02 15:04"),
		},
		{
			ID:      "alert-004",
			Level:   "info",
			Message: "每日备份任务已完成，耗时 4m32s",
			Source:  "backup-agent",
			Time:    time.Now().Add(-120 * time.Minute).Format("01-02 15:04"),
		},
		{
			ID:      "alert-005",
			Level:   "critical",
			Message: "数据库主从延迟超过 5 秒，当前 7.3s",
			Source:  "db-master (192.168.1.10)",
			Time:    time.Now().Add(-180 * time.Minute).Format("01-02 15:04"),
		},
	}

	if len(alerts) > limit {
		alerts = alerts[:limit]
	}

	c.JSON(http.StatusOK, alerts)
}
