package api

import (
	"net/http"
	"strconv"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// GetLogList 获取日志列表
// @Summary     获取日志列表
// @Description 按分页、级别、服务和关键词筛选返回日志列表
// @Tags        Log
// @Param       page     query int    false "页码"      default(1)
// @Param       pageSize query int    false "每页条数"   default(10)
// @Param       level    query string false "日志级别"   Enums(INFO, WARN, ERROR)
// @Param       service  query string false "服务名称"
// @Param       keyword  query string false "关键词搜索"
// @Success     200 {object} model.PagedResultLogItem
// @Router      /logs [get]
func (h *Handler) GetLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	level := c.Query("level")
	service := c.Query("service")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	logs, total, err := h.services.LogService.List(page, pageSize, level, service, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.PagedResultLogItem{
		List:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
