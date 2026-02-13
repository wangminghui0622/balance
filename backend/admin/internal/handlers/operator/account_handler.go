package operator

import (
	"strconv"

	"balance/backend/internal/models"
	"balance/backend/internal/services"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// AccountHandler 运营账户处理器
type AccountHandler struct {
	accountService *services.AccountService
}

// NewAccountHandler 创建运营账户处理器
func NewAccountHandler() *AccountHandler {
	return &AccountHandler{
		accountService: services.NewAccountService(),
	}
}

// GetAccount 获取运营账户信息
// GET /operator/account
func (h *AccountHandler) GetAccount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	operatorID := userID.(int64)

	account, err := h.accountService.GetOrCreateOperatorAccount(c.Request.Context(), operatorID)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, account)
}

// GetTransactions 获取运营账户流水
// GET /operator/account/transactions?page=1&page_size=20
func (h *AccountHandler) GetTransactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	operatorID := userID.(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	transactions, total, err := h.accountService.GetAccountTransactions(c.Request.Context(), models.AccountTypeOperator, operatorID, page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, transactions, total, page, pageSize)
}

// ApplyWithdraw 申请提现
// POST /operator/withdraw/apply
func (h *AccountHandler) ApplyWithdraw(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	operatorID := userID.(int64)

	var req struct {
		Amount              float64 `json:"amount" binding:"required,gt=0"`
		CollectionAccountID uint64  `json:"collection_account_id" binding:"required"`
		Remark              string  `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	amount := utils.ToDecimal(req.Amount)
	application, err := h.accountService.ApplyWithdraw(c.Request.Context(), operatorID, models.AccountTypeOperator, amount, req.CollectionAccountID, req.Remark)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, application)
}

// GetWithdrawApplications 获取提现申请列表
// GET /operator/withdraw/list?status=-1&page=1&page_size=20
func (h *AccountHandler) GetWithdrawApplications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	operatorID := userID.(int64)

	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	applications, total, err := h.accountService.GetWithdrawApplications(c.Request.Context(), operatorID, int8(status), page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, applications, total, page, pageSize)
}
