package api

import (
	"net/http"
	"strconv"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/Tania-X/devops-dashboard/backend/internal/repository"
	"github.com/gin-gonic/gin"
)

func GetServerList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	var servers []model.Server
	var total int64

	query := repository.DB.Model(&model.Server{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	query.Order("CASE status WHEN 'running' THEN 1 WHEN 'maintenance' THEN 2 WHEN 'stopped' THEN 3 END, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&servers)

	c.JSON(http.StatusOK, model.PagedResultServerItem{
		List:     servers,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func GetServerDetail(c *gin.Context) {
	id := c.Param("id")

	var server model.Server
	result := repository.DB.Preload("DiskPartitions").Preload("NetworkInterfaces").First(&server, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	c.JSON(http.StatusOK, server)
}
