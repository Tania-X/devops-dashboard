package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetDashboardMetrics 获取仪表盘核心指标
// @Summary     获取仪表盘核心指标
// @Description 返回本机实时的 CPU、内存、磁盘使用率及当前告警数
// @Tags        Dashboard
// @Success     200 {object} model.DashboardMetrics
// @Router      /dashboard/metrics [get]
func (h *Handler) GetDashboardMetrics(c *gin.Context) {
	metrics, err := h.services.DashboardService.GetMetrics()
	if err != nil {
		slog.Warn("采集系统指标失败，使用降级数据", "err", err)
	}
	c.JSON(http.StatusOK, metrics)
}

// GetDashboardTrend 获取仪表盘趋势数据
// @Summary     获取仪表盘趋势数据
// @Description 返回最近 N 小时内本机 CPU 和内存使用率的时序数据
// @Tags        Dashboard
// @Param       hours query int false "查询小时范围" default(6)
// @Success     200 {object} model.DashboardTrend
// @Router      /dashboard/trend [get]
func (h *Handler) GetDashboardTrend(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "6"))
	if hours < 1 || hours > 24 {
		hours = 6
	}
	trend, err := h.services.DashboardService.GetTrend(hours)
	if err != nil {
		slog.Warn("获取趋势数据失败", "hours", hours, "err", err)
	}
	c.JSON(http.StatusOK, trend)
}

// GetDashboardAlerts 获取告警列表
// @Summary     获取告警列表
// @Description 返回最近的告警条目
// @Tags        Dashboard
// @Param       limit query int false "返回条数" default(5)
// @Success     200 {array} model.AlertItem
// @Router      /dashboard/alerts [get]
func (h *Handler) GetDashboardAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit > 20 {
		limit = 20
	}

	alerts := h.services.DashboardService.GetAlerts(limit)
	c.JSON(http.StatusOK, alerts)
}
