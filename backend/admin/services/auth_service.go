package services

import (
	"balance/internal/models"
	"errors"
	"gorm.io/gorm"
)

type AuthService struct {
	authRepo *models.AuthRepository
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		authRepo: models.NewAuthRepository(db),
	}
}
func (s *AuthService) GetByPartnerId() (*models.AuthConfig, error) {
	auth, err := s.authRepo.GetByPartnerId()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}
	return auth, nil
}
