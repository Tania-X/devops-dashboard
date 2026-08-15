package service

import (
	"errors"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/authz"
	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) List() ([]model.User, error) {
	var users []model.User
	if err := s.db.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Create 创建用户:校验角色存在性 → 工厂创建实体(内部完成密码哈希)→ 落库。
// 编排者角色,业务规则(密码哈希/角色校验)收敛在实体与工厂。
func (s *UserService) Create(req model.CreateUserRequest) (*model.User, error) {
	// 校验角色存在(防止创建绑定不存在角色的用户)
	if err := s.validateRole(req.Role); err != nil {
		return nil, err
	}
	u, err := model.NewUser(req.Username, req.Password, model.UserRole(req.Role))
	if err != nil {
		return nil, err
	}
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	if err := s.db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

// Update 更新用户:角色/密码留空表示不修改;变更走实体方法,业务规则内聚。
func (s *UserService) Update(id string, req model.UpdateUserRequest) (*model.User, error) {
	var existing model.User
	if err := s.db.First(&existing, "id = ?", id).Error; err != nil {
		return nil, err
	}
	// 角色变更:先查证存在性(DB),再应用实体方法(格式/内置校验)
	if req.Role != "" {
		if err := s.validateRole(req.Role); err != nil {
			return nil, err
		}
		if err := existing.ChangeRole(model.UserRole(req.Role)); err != nil {
			return nil, err
		}
	}
	// 密码变更:留空表示不修改
	if req.Password != "" {
		if err := existing.SetPassword(req.Password); err != nil {
			return nil, err
		}
	}
	existing.UpdatedAt = time.Now()
	if err := s.db.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

// validateRole 校验角色是否存在于 roles 表(数据库错误时保守拒绝,避免绑定未知角色)。
func (s *UserService) validateRole(role string) error {
	exists, err := authz.RoleExists(role)
	if err != nil {
		return errors.New("校验角色失败: " + err.Error())
	}
	if !exists {
		return errors.New("角色不存在: " + role)
	}
	return nil
}

func (s *UserService) Delete(id string) error {
	return s.db.Delete(&model.User{}, "id = ?", id).Error
}
