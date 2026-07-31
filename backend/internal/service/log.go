package service

import (
	"github.com/Tania-X/devops-dashboard/backend/internal/logs"
	"github.com/Tania-X/devops-dashboard/backend/internal/model"
)

type LogService struct {
	reader *logs.Reader
}

func NewLogService(reader *logs.Reader) *LogService {
	return &LogService{reader: reader}
}

func (l *LogService) List(page, pageSize int, level, service, keyword string) ([]model.Log, int64, error) {
	// service 参数在文件日志场景下对应日志的 service 字段
	// 当前实现：service 过滤合并到 keyword 搜索中
	if service != "" && keyword == "" {
		keyword = service
	}
	return l.reader.List(page, pageSize, level, keyword)
}
