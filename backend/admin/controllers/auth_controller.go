package controllers

import (
	"balance/admin/dto"
	"balance/admin/services"
	"balance/internal/constants"
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
			Code:    constants.HTTPStatusBadRequest,
			Message: "参数错误: " + err.Error(),
		})
		return
	}
	admin, token, err := ctrl.authService.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP())
	if err != nil {
		code := constants.HTTPStatusUnauthorized
		if err.Error() == "账号已被禁用" {
			code = constants.HTTPStatusForbidden
		}
		c.JSON(code, dto.Response{
			Code:    code,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Code:    constants.ResponseCodeSuccess,
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
			Code:    constants.HTTPStatusBadRequest,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	_, err := ctrl.authService.Register(c.Request.Context(), req.Username, req.Password, req.Email, req.UserType)
	if err != nil {
		code := constants.HTTPStatusBadRequest
		if err.Error() == "生成用户ID失败" || err.Error() == "创建用户失败" {
			code = constants.HTTPStatusInternalServerError
		}
		c.JSON(code, dto.Response{
			Code:    code,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.RegisterResponse{
		Code:    constants.ResponseCodeSuccess,
		Message: "注册成功",
	})
}
