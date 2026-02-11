package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"balance/backend/internal/config"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthService 认证服务
type AuthService struct {
	db          *gorm.DB
	idGenerator *utils.IDGenerator
}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{
		db:          database.GetDB(),
		idGenerator: utils.NewIDGenerator(database.GetRedis()),
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=6,max=16"`
	Password  string `json:"password" binding:"required,min=6"`
	Email     string `json:"email" binding:"required,email"`
	EmailCode string `json:"emailCode" binding:"required,len=6"`
	UserType  int8   `json:"userType" binding:"required,oneof=1 5"`
	RealName  string `json:"realName"`
	Phone     string `json:"phone"`
	LineID    string `json:"lineId"`
	Wechat    string `json:"wechat"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	EmailCode   string `json:"emailCode" binding:"required,len=6"`
	NewPassword string `json:"newPassword" binding:"required,min=8,max=16"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	UserID   int64  `json:"userId"`
	UserType int8   `json:"userType"`
}

// CurrentUserResponse 当前用户响应
type CurrentUserResponse struct {
	ID       int64  `json:"id"`
	UserNo   string `json:"userNo"`
	UserType int8   `json:"userType"`
	UserName string `json:"userName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) error {
	// 验证邮箱验证码
	emailService := NewEmailService()
	if err := emailService.VerifyCode(ctx, req.Email, req.EmailCode); err != nil {
		return err
	}

	var existing models.Admin
	if err := s.db.Where("user_name = ?", req.Username).First(&existing).Error; err == nil {
		return errors.New("用户名已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询用户失败: %w", err)
	}

	if err := s.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return errors.New("邮箱已被注册")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询邮箱失败: %w", err)
	}

	var userID int64
	var err error
	switch req.UserType {
	case models.UserTypeShopOwner:
		userID, err = s.idGenerator.GenerateShopOwnerID(ctx)
	case models.UserTypeOperator:
		userID, err = s.idGenerator.GenerateOperatorID(ctx)
	case models.UserTypePlatform:
		userID, err = s.idGenerator.GeneratePlatformID(ctx)
	default:
		return errors.New("无效的用户类型")
	}
	if err != nil {
		return fmt.Errorf("生成用户ID失败: %w", err)
	}

	salt := generateSalt(8)
	hash := hashPassword(req.Password, salt)
	userNo := utils.GenerateUserNo(userID)

	admin := models.Admin{
		ID:       userID,
		UserNo:   userNo,
		UserType: req.UserType,
		UserName: req.Username,
		RealName: req.RealName,
		Salt:     salt,
		Hash:     hash,
		Email:    req.Email,
		Phone:    req.Phone,
		LineID:   req.LineID,
		Wechat:   req.Wechat,
		Status:   models.UserStatusNormal,
		Language: "zh",
	}

	if err := s.db.Create(&admin).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	return nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *LoginRequest, clientIP string) (*LoginResponse, error) {
	var admin models.Admin
	if err := s.db.Where("user_name = ?", req.Username).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	if admin.IsDisabled() {
		return nil, errors.New("账户已被禁用")
	}

	if !verifyPassword(req.Password, admin.Salt, admin.Hash) {
		return nil, errors.New("用户名或密码错误")
	}

	now := time.Now()
	s.db.Model(&admin).Updates(map[string]interface{}{
		"login_ip":   clientIP,
		"login_date": now,
	})

	token, err := generateToken(admin.ID, admin.UserType)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	return &LoginResponse{
		Token:    token,
		UserID:   admin.ID,
		UserType: admin.UserType,
	}, nil
}

// GetCurrentUser 获取当前用户信息
func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*CurrentUserResponse, error) {
	var admin models.Admin
	if err := s.db.First(&admin, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &CurrentUserResponse{
		ID:       admin.ID,
		UserNo:   admin.UserNo,
		UserType: admin.UserType,
		UserName: admin.UserName,
		Email:    admin.Email,
		Phone:    admin.Phone,
		Avatar:   admin.Avatar,
	}, nil
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	// 验证邮箱验证码
	emailService := NewEmailService()
	if err := emailService.VerifyCode(ctx, req.Email, req.EmailCode); err != nil {
		return err
	}

	// 查找用户
	var admin models.Admin
	if err := s.db.Where("email = ?", req.Email).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("该邮箱未注册")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 生成新的盐和密码哈希
	salt := generateSalt(8)
	hash := hashPassword(req.NewPassword, salt)

	// 更新密码
	if err := s.db.Model(&admin).Updates(map[string]interface{}{
		"salt": salt,
		"hash": hash,
	}).Error; err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

func generateSalt(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

func hashPassword(password, salt string) string {
	h := sha256.New()
	h.Write([]byte(password + salt))
	return hex.EncodeToString(h.Sum(nil))
}

func verifyPassword(password, salt, hash string) bool {
	return hashPassword(password, salt) == hash
}

// Claims JWT Claims
type Claims struct {
	UserID   int64 `json:"user_id"`
	UserType int8  `json:"user_type"`
	jwt.RegisteredClaims
}

func generateToken(userID int64, userType int8) (string, error) {
	cfg := config.Get()
	jwtSecret := cfg.JWT.Secret
	expireHours := cfg.JWT.ExpireHours

	claims := Claims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "sheepx",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.Get()
	jwtSecret := cfg.JWT.Secret

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
