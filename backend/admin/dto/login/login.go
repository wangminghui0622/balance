package login

import "balance/internal/constants"

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	UserType int8   `json:"userType" binding:"required"` // constants.UserTypeShopOwner=店铺, constants.UserTypeOperator=运营
}

// GetUserTypeName 获取用户类型名称
func (r *RegisterRequest) GetUserTypeName() string {
	switch r.UserType {
	case constants.UserTypeShopOwner:
		return "店铺"
	case constants.UserTypeOperator:
		return "运营"
	case constants.UserTypePlatform:
		return "平台"
	default:
		return "未知"
	}
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    LoginData `json:"data"`
}

// LoginData 登录数据
type LoginData struct {
	UserType int8   `json:"userType" binding:"required"`
	Token    string `json:"token"`
	UserID   int64  `json:"userId"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Response 通用响应
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
