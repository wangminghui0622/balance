package handlers

import (
	"strings"

	"balance/backend/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService  *services.AuthService
	emailService *services.EmailService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService:  services.NewAuthService(),
		emailService: services.NewEmailService(),
	}
}

// SendEmailCode 发送邮箱验证码
// POST /api/v1/balance/app/auth/send-code
func (h *AuthHandler) SendEmailCode(c *gin.Context) {
	var req services.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.emailService.SendVerificationCode(c.Request.Context(), &req); err != nil {
		Error(c, 400, err.Error())
		return
	}

	Success(c, gin.H{"message": "验证码已发送"})
}

// Register 用户注册
// POST /api/v1/balance/admin/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.authService.Register(c.Request.Context(), &req); err != nil {
		Error(c, 400, err.Error())
		return
	}

	Success(c, nil)
}

// Login 用户登录
// POST /api/v1/balance/admin/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	clientIP := c.ClientIP()
	resp, err := h.authService.Login(c.Request.Context(), &req, clientIP)
	if err != nil {
		Unauthorized(c, err.Error())
		return
	}

	Success(c, resp)
}

// GetCurrentUser 获取当前用户信息
// GET /api/v1/balance/admin/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// 从上下文获取用户ID（由JWT中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		Unauthorized(c, "未登录")
		return
	}

	resp, err := h.authService.GetCurrentUser(c.Request.Context(), userID.(int64))
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	SuccessWithCode(c, 200, resp)
}

// ResetPassword 重置密码
// POST /api/v1/balance/app/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req services.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), &req); err != nil {
		Error(c, 400, err.Error())
		return
	}

	Success(c, gin.H{"message": "密码重置成功"})
}

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			Unauthorized(c, "未提供认证信息")
			c.Abort()
			return
		}

		// Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := services.ParseToken(tokenString)
		if err != nil {
			Unauthorized(c, "认证失败: "+err.Error())
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("user_type", claims.UserType)
		c.Next()
	}
}
