package service

import (
	"time"

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
	for i := range users {
		users[i].Password = ""
	}
	return users, nil
}

func (s *UserService) Create(user model.User) (*model.User, error) {
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
	var existing model.User
	if err := s.db.First(&existing, "id = ?", id).Error; err != nil {
		return nil, err
	}
	existing.Username = user.Username
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

func (s *UserService) Delete(id string) error {
	return s.db.Delete(&model.User{}, "id = ?", id).Error
}
