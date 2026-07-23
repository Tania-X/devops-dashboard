package api

import (
	"net/http"
	"strconv"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/gin-gonic/gin"
)

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

	var logs []model.Log
	var total int64

	query := h.db.Model(&model.Log{})
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if service != "" {
		query = query.Where("service = ?", service)
	}
	if keyword != "" {
		query = query.Where("content LIKE ? OR service LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)
	query.Order("time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)

	c.JSON(http.StatusOK, model.PagedResultLogItem{
		List:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
