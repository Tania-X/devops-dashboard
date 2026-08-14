package service

import (
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditService 审计日志服务：记录"谁在什么时候对什么做了什么"。
// 只记录敏感管理操作（角色/权限/用户增删改），不记录登录等噪音。
type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// 动作枚举（与前端展示一一对应）
const (
	ActionRoleCreate       = "role.create"
	ActionRoleUpdate       = "role.update"
	ActionRoleDelete       = "role.delete"
	ActionPermissionUpdate = "permission.update"
	ActionUserCreate       = "user.create"
	ActionUserUpdate       = "user.update"
	ActionUserDelete       = "user.delete"
)

// Record 写入一条审计日志（幂等，不因审计失败影响主操作）。
func (s *AuditService) Record(actor, action, target, detail string) {
	if s == nil || s.db == nil {
		return
	}
	log := model.AuditLog{
		ID:        uuid.New().String(),
		Actor:     actor,
		Action:    action,
		Target:    target,
		Detail:    detail,
		CreatedAt: time.Now(),
	}
	// 审计失败只记日志不向上抛错（审计是旁路，不影响主流程）
	_ = s.db.Create(&log).Error
}

// List 分页查询审计日志（新→旧）。
type AuditPage struct {
	Items []model.AuditLog `json:"items"`
	Total int64            `json:"total"`
}

func (s *AuditService) List(page, size int) (*AuditPage, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	var total int64
	if err := s.db.Model(&model.AuditLog{}).Count(&total).Error; err != nil {
		return nil, err
	}
	var items []model.AuditLog
	if err := s.db.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}
	return &AuditPage{Items: items, Total: total}, nil
}
