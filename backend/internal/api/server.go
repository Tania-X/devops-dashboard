package api

import (
	"net/http"
	"strconv"

	serverdomain "github.com/Tania-X/devops-dashboard/backend/internal/dashboard/server/domain"
	"github.com/gin-gonic/gin"
)

// GetServerList 获取服务器列表
// @Summary     获取服务器列表
// @Description 按分页、状态筛选返回服务器列表
// @Tags        Server
// @Param       page     query int    false "页码"  default(1)
// @Param       pageSize query int    false "每页条数" default(10)
// @Param       status   query string false "状态筛选" Enums(running, stopped, maintenance)
// @Success     200 {object} serverdomain.PagedResultServerItem
// @Router      /servers [get]
func (h *Handler) GetServerList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	servers, total, err := h.services.ServerService.List(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, serverdomain.PagedResultServerItem{
		List:     servers,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetServerDetail 获取服务器详情
// @Summary     获取服务器详情
// @Description 按 ID 获取单台服务器详情（含磁盘分区和网络接口）
// @Tags        Server
// @Param       id path string true "服务器 ID"
// @Success     200 {object} serverdomain.Server
// @Failure     404 {object} map[string]string
// @Router      /servers/{id} [get]
func (h *Handler) GetServerDetail(c *gin.Context) {
	id := c.Param("id")

	server, err := h.services.ServerService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	c.JSON(http.StatusOK, server)
}
