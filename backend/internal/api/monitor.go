package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetProcessList 获取进程列表
// @Summary     获取进程列表
// @Description 返回本机所有进程列表，支持按字段排序和关键词搜索
// @Tags        Monitor
// @Param       sortBy  query string false "排序字段" Enums(cpu, memory, pid, name) default(cpu)
// @Param       order   query string false "排序方向" Enums(desc, asc) default(desc)
// @Param       keyword query string false "进程名搜索"
// @Param       limit   query int    false "返回条数" default(50)
// @Success     200 {array} model.ProcessItem
// @Router      /monitor/processes [get]
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

// GetProcessDetail 获取进程详情
// @Summary     获取进程详情
// @Description 按 PID 获取单个进程的详细信息
// @Tags        Monitor
// @Param       pid path int true "进程 PID"
// @Success     200 {object} model.ProcessDetail
// @Failure     404 {object} map[string]string
// @Router      /monitor/processes/{pid} [get]
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
// @Summary     获取主机信息
// @Description 返回本机的主机名、操作系统、CPU 和内存等详细信息
// @Tags        Monitor
// @Success     200 {object} model.HostInfo
// @Failure     500 {object} map[string]string
// @Router      /monitor/host [get]
func (h *Handler) GetHostInfo(c *gin.Context) {
	hostInfo, err := h.services.MonitorService.GetHostInfo()
	if err != nil {
		slog.Warn("获取主机信息失败", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hostInfo)
}
