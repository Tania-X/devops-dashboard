package service

import (
	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"gorm.io/gorm"
)

type LogService struct {
	db *gorm.DB
}

func NewLogService(db *gorm.DB) *LogService {
	return &LogService{db: db}
}

func (l *LogService) List(page, pageSize int, level, service, keyword string) ([]model.Log, int64, error) {
	var logs []model.Log
	var total int64

	query := l.db.Model(&logs)
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if service != "" {
		query = query.Where("service = ?", service)
	}
	if keyword != "" {
		query = query.Where("content LIKE ? OR service LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if result := query.Count(&total); result.Error != nil {
		return logs, 0, result.Error
	}
	if result := query.Order("time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs); result.Error != nil {
		return logs, 0, result.Error
	}

	return logs, total, nil

}
