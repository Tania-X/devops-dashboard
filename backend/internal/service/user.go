package service

import (
	"errors"
	"time"

	"github.com/Tania-X/devops-dashboard/backend/internal/authz"
	"github.com/Tania-X/devops-dashboard/backend/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

func (s *UserService) Create(user model.User) (*model.User, error) {
	// 校验角色存在（防止创建绑定不存在角色的用户）
	if err := s.validateRole(user.Role); err != nil {
		return nil, err
	}
	user.ID = uuid.New().String()
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashed)
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}
	user.Password = ""
	return &user, nil
}

func (s *UserService) Update(id string, user model.User) (*model.User, error) {
	// 校验角色存在（防止将用户绑定到不存在的角色）
	if err := s.validateRole(user.Role); err != nil {
		return nil, err
	}
	var existing model.User
	if err := s.db.First(&existing, "id = ?", id).Error; err != nil {
		return nil, err
	}
	// 用户名不可修改（前端编辑表单 username 为 disabled，请求 DTO 不含该字段）
	existing.Role = user.Role
	existing.UpdatedAt = time.Now()
	if user.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		existing.Password = string(hashed)
	}
	if err := s.db.Save(&existing).Error; err != nil {
		return nil, err
	}
	existing.Password = ""
	return &existing, nil
}

// validateRole 校验角色是否存在于 roles 表（数据库错误时保守拒绝，避免绑定未知角色）。
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
