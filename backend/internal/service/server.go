package service

import (
	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"gorm.io/gorm"
)

type ServerService struct {
	db *gorm.DB
}

func NewServerService(db *gorm.DB) *ServerService {
	return &ServerService{db: db}
}

func (s *ServerService) List(page, pageSize int, status string) ([]model.Server, int64, error) {

	var servers []model.Server
	var total int64

	query := s.db.Model(&servers)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if result := query.Count(&total); result.Error != nil {
		return servers, 0, result.Error
	}
	if result := query.Order("CASE status WHEN 'running' THEN 1 WHEN 'maintenance' THEN 2 WHEN 'stopped' THEN 3 END, id ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&servers); result.Error != nil {
		return servers, 0, result.Error
	}
	return servers, total, nil
}

func (s *ServerService) GetByID(id string) (model.Server, error) {
	var server model.Server
	if result := s.db.Preload("DiskPartitions").Preload("NetworkInterfaces").
		First(&server, "id = ?", id); result.Error != nil {
		return server, result.Error
	}
	return server, nil
}
