package platform

import (
	"context"
	"fmt"

	"balance/backend/internal/database"
	"balance/backend/internal/models"

	"gorm.io/gorm"
)

// UserService 用户服务（平台专用）
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		db: database.GetDB(),
	}
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, userType, keyword string, page, pageSize int) ([]models.Admin, int64, error) {
	var users []models.Admin
	var total int64

	query := s.db.Model(&models.Admin{})
	if userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetUser 获取用户详情
func (s *UserService) GetUser(ctx context.Context, userID int64) (*models.Admin, error) {
	var user models.Admin
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return &user, nil
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(ctx context.Context, userID int64, status int) error {
	result := s.db.Model(&models.Admin{}).Where("id = ?", userID).Update("status", status)
	if result.RowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}
	return result.Error
}
