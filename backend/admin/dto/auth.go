package dto

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
	UserType int8   `json:"userType" binding:"required"` // 1=店铺, 5=运营
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    LoginData `json:"data"`
}

// LoginData 登录数据
type LoginData struct {
	Token  string `json:"token"`
	UserID int64  `json:"userId"`
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
