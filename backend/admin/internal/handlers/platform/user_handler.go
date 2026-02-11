package platform

import (
	"strconv"

	"balance/backend/internal/consts"
	"balance/backend/internal/services/platform"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理处理器（平台专用）
type UserHandler struct {
	userService *platform.UserService
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: platform.NewUserService(),
	}
}

// ListUsers 获取用户列表（平台可查看所有用户）
// GET /api/v1/balance/admin/platform/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userType := c.Query("user_type")
	keyword := c.Query("keyword")

	if page < 1 {
		page = consts.DefaultPage
	}
	if pageSize < 1 || pageSize > consts.MaxPageSize {
		pageSize = consts.DefaultPageSize
	}

	list, total, err := h.userService.ListUsers(c.Request.Context(), userType, keyword, page, pageSize)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// GetUser 获取用户详情
// GET /api/v1/balance/admin/platform/users/:user_id
func (h *UserHandler) GetUser(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "用户ID格式错误")
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, user)
}

// UpdateUserStatus 更新用户状态
// PUT /api/v1/balance/admin/platform/users/:user_id/status
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "用户ID格式错误")
		return
	}

	var req struct {
		Status int `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.userService.UpdateUserStatus(c.Request.Context(), userID, req.Status); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, nil)
}
