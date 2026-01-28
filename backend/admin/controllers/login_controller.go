package controllers

import (
	"balance/admin/dto/login"
	"balance/admin/services"
	"balance/internal/constants"
	shareUtils "balance/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type LoginController struct {
	loginService *services.LoginService
}

func NewLoginController(loginService *services.LoginService) *LoginController {
	return &LoginController{
		loginService: loginService,
	}
}

// Login 登录
func (ctrl *LoginController) Login(c *gin.Context) {
	var req login.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, login.Response{
			Code:    constants.HTTPStatusBadRequest,
			Message: "参数错误: " + err.Error(),
		})
		return
	}
	admin, token, err := ctrl.loginService.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP())
	if err != nil {
		code := constants.HTTPStatusUnauthorized
		if err.Error() == "账号已被禁用" {
			code = constants.HTTPStatusForbidden
		}
		c.JSON(code, login.Response{
			Code:    code,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, login.LoginResponse{
		Code:    constants.ResponseCodeSuccess,
		Message: "success",
		Data: login.LoginData{
			UserType: admin.UserType,
			Token:    token,
			UserID:   admin.ID,
		},
	})
}

// GetCurrentUser 获取当前登录用户信息
func (ctrl *LoginController) GetCurrentUser(c *gin.Context) {
	// 从请求头获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, login.Response{
			Code:    constants.HTTPStatusUnauthorized,
			Message: "缺少Authorization头",
		})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.JSON(http.StatusUnauthorized, login.Response{
			Code:    constants.HTTPStatusUnauthorized,
			Message: "Authorization格式错误",
		})
		return
	}

	tokenStr := parts[1]
	userID, err := shareUtils.ParseToken(tokenStr, ctrl.loginService.GetJWTSecret())
	if err != nil {
		c.JSON(http.StatusUnauthorized, login.Response{
			Code:    constants.HTTPStatusUnauthorized,
			Message: "无效或过期的token",
		})
		return
	}

	// 查询用户信息
	admin, err := ctrl.loginService.GetAdminByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, login.Response{
			Code:    constants.HTTPStatusInternalServerError,
			Message: "获取用户信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    constants.ResponseCodeSuccess,
		"message": "success",
		"data": gin.H{
			"id":       admin.ID,
			"userNo":   admin.UserNo,
			"userType": admin.UserType,
			"userName": admin.UserName,
			"email":    admin.Email,
			"phone":    admin.Phone,
			"avatar":   admin.Avatar,
		},
	})
}

// Register 注册
func (ctrl *LoginController) Register(c *gin.Context) {
	var req login.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, login.Response{
			Code:    constants.HTTPStatusBadRequest,
			Message: "参数错误: " + err.Error(),
		})
		return
	}
	_, err := ctrl.loginService.Register(c.Request.Context(), req.Username, req.Password, req.Email, req.UserType)
	if err != nil {
		code := constants.HTTPStatusBadRequest
		if err.Error() == "生成用户ID失败" || err.Error() == "创建用户失败" {
			code = constants.HTTPStatusInternalServerError
		}
		c.JSON(code, login.Response{
			Code:    code,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, login.RegisterResponse{
		Code:    constants.ResponseCodeSuccess,
		Message: "注册成功",
	})
}
