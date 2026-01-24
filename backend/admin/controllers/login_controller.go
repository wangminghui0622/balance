package controllers

import (
	"balance/admin/dto/login"
	"balance/admin/services"
	"balance/internal/constants"
	"net/http"

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
			Token:  token,
			UserID: admin.ID,
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
