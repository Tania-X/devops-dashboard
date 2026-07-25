package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetProcessList 获取本机进程列表
func (h *Handler) GetProcessList(c *gin.Context) {
	sortBy := c.DefaultQuery("sortBy", "cpu")
	order := c.DefaultQuery("order", "desc")
	keyword := strings.ToLower(c.DefaultQuery("keyword", ""))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	items, err := h.services.MonitorService.GetProcessList(sortBy, order, keyword, limit)
	if err != nil {
		slog.Warn("获取进程列表失败", "err", err)
	}
	c.JSON(http.StatusOK, items)
}

// GetProcessDetail 获取单个进程详情
func (h *Handler) GetProcessDetail(c *gin.Context) {
	pid, err := strconv.Atoi(c.Param("pid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 PID"})
		return
	}

	processDetail, err := h.services.MonitorService.GetProcessDetail(pid)
	if processDetail == nil {
		slog.Warn("获取进程详情失败", "err", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "进程不存在或已退出"})
		return
	}
	if err != nil {
		slog.Warn("获取进程详情异常", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, processDetail)
	}
}

// GetHostInfo 获取主机信息
func (h *Handler) GetHostInfo(c *gin.Context) {
	hostInfo, err := h.services.MonitorService.GetHostInfo()
	if err != nil {
		slog.Warn("获取主机信息失败", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hostInfo)
}
