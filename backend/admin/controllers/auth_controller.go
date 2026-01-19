package controllers

import (
	"balance/admin/dto"
	"balance/admin/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
type AuthController struct {
	authService *services.AuthService
}

// NewAuthController 创建认证控制器
func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Login 登录
func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	admin, token, err := ctrl.authService.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP())
	if err != nil {
		code := http.StatusUnauthorized
		if err.Error() == "账号已被禁用" {
			code = http.StatusForbidden
		}
		c.JSON(code, dto.Response{
			Code:    code,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Code:    200,
		Message: "success",
		Data: dto.LoginData{
			Token:  token,
			UserID: admin.ID,
		},
	})
}

// Register 注册
func (ctrl *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Code:    400,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	_, err := ctrl.authService.Register(c.Request.Context(), req.Username, req.Password, req.Email, req.UserType)
	if err != nil {
		code := http.StatusBadRequest
		if err.Error() == "生成用户ID失败" || err.Error() == "创建用户失败" {
			code = http.StatusInternalServerError
		}
		c.JSON(code, dto.Response{
			Code:    code,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.RegisterResponse{
		Code:    200,
		Message: "注册成功",
	})
}
