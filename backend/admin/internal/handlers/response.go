package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResponse 分页响应结构
type PageResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	List     interface{} `json:"list"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithPage 分页成功响应
func SuccessWithPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PageResponse{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			List:     list,
		},
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 参数错误
func BadRequest(c *gin.Context, message string) {
	Error(c, 400, message)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	Error(c, 404, message)
}

// InternalError 内部错误
func InternalError(c *gin.Context, message string) {
	Error(c, 500, message)
}

// SuccessWithCode 成功响应（指定code）
func SuccessWithCode(c *gin.Context, code int, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: "success",
		Data:    data,
	})
}
