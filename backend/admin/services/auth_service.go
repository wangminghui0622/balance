package services

import (
	shareUtils "balance/internal/utils"
	"balance/models"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	adminRepo     *models.AdminRepository
	idGenerator   *shareUtils.IDGenerator
	jwtSecret     []byte
	jwtExpiration time.Duration
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, redisClient *redis.Client, jwtSecret []byte, jwtExpiration time.Duration) *AuthService {
	return &AuthService{
		adminRepo:     models.NewAdminRepository(db),
		idGenerator:   shareUtils.NewIDGenerator(redisClient),
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

// Login 登录
func (s *AuthService) Login(ctx context.Context, username, password, clientIP string) (*models.Admin, string, error) {
	// 查询用户
	admin, err := s.adminRepo.GetByUserName(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", errors.New("用户名或密码错误")
		}
		return nil, "", err
	}

	// 验证密码
	if !shareUtils.VerifyPassword(password, admin.Salt, admin.Hash) {
		return nil, "", errors.New("用户名或密码错误")
	}

	// 检查状态
	if admin.Status != 1 {
		return nil, "", errors.New("账号已被禁用")
	}

	// 更新登录信息
	s.adminRepo.UpdateLoginInfo(admin.ID, clientIP)

	// 生成JWT token
	token, err := s.generateToken(admin.ID)
	if err != nil {
		return nil, "", errors.New("生成token失败")
	}

	return admin, token, nil
}

// Register 注册
func (s *AuthService) Register(ctx context.Context, username, password, email string, userType int8) (*models.Admin, error) {
	// 验证用户类型
	if userType != 1 && userType != 5 {
		return nil, errors.New("用户类型错误，1=店铺，5=运营")
	}

	// 检查用户名是否存在
	exists, err := s.adminRepo.ExistsByUserName(username)
	if err != nil {
		return nil, errors.New("检查用户名失败")
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	if email != "" {
		exists, err = s.adminRepo.ExistsByEmail(email)
		if err != nil {
			return nil, errors.New("检查邮箱失败")
		}
		if exists {
			return nil, errors.New("邮箱已存在")
		}
	}

	// 生成ID
	var userID int64
	if userType == 1 {
		// 店铺
		userID, err = s.idGenerator.GenerateShopOwnerID(ctx)
	} else if userType == 5 {
		// 运营
		userID, err = s.idGenerator.GenerateOperatorID(ctx)
	}
	if err != nil {
		return nil, errors.New("生成用户ID失败")
	}

	// 生成密码盐和hash
	salt, err := shareUtils.GenerateSalt()
	if err != nil {
		return nil, errors.New("生成密码盐失败")
	}
	hash := shareUtils.HashPassword(password, salt)

	// 生成用户编号
	userNo := shareUtils.GenerateUserNo(userID)

	// 创建用户
	admin := &models.Admin{
		ID:       userID,
		UserNo:   userNo,
		UserType: userType,
		UserName: username,
		Salt:     salt,
		Hash:     hash,
		Email:    email,
		Status:   1,
		Language: "zh",
	}

	err = s.adminRepo.Create(admin)
	if err != nil {
		return nil, errors.New("创建用户失败: " + err.Error())
	}

	return admin, nil
}

// generateToken 生成JWT Token
func (s *AuthService) generateToken(userID int64) (string, error) {
	return shareUtils.GenerateToken(userID, s.jwtSecret, s.jwtExpiration)
}
