package utils

// 业务状态码定义
const (
	// 成功
	CodeSuccess = 0

	// 通用错误 (400-499)
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404

	// 服务器错误 (500-599)
	CodeInternalError = 500
)

// 状态码对应的默认消息
var CodeMessages = map[int]string{
	CodeSuccess:       "success",
	CodeBadRequest:    "参数错误",
	CodeUnauthorized:  "未授权",
	CodeForbidden:     "禁止访问",
	CodeNotFound:      "资源不存在",
	CodeInternalError: "内部错误",
}

// GetCodeMessage 获取状态码对应的默认消息
func GetCodeMessage(code int) string {
	if msg, ok := CodeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
