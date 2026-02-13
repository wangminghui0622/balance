package shopower

import (
	"strconv"

	"balance/backend/internal/models"
	"balance/backend/internal/services"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// AccountHandler 账户处理器
type AccountHandler struct {
	accountService    *services.AccountService
	settlementService *services.SettlementService
}

// NewAccountHandler 创建账户处理器
func NewAccountHandler() *AccountHandler {
	return &AccountHandler{
		accountService:    services.NewAccountService(),
		settlementService: services.NewSettlementService(),
	}
}

// GetPrepaymentAccount 获取预付款账户信息
// GET /shopower/account/prepayment
func (h *AccountHandler) GetPrepaymentAccount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	account, err := h.accountService.GetOrCreatePrepaymentAccount(c.Request.Context(), adminID)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, account)
}

// GetDepositAccount 获取保证金账户信息
// GET /shopower/account/deposit
func (h *AccountHandler) GetDepositAccount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	account, err := h.accountService.GetOrCreateDepositAccount(c.Request.Context(), adminID)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, account)
}

// GetPrepaymentTransactions 获取预付款流水
// GET /shopower/account/prepayment/transactions?page=1&page_size=20
func (h *AccountHandler) GetPrepaymentTransactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	transactions, total, err := h.accountService.GetAccountTransactions(c.Request.Context(), models.AccountTypePrepayment, adminID, page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, transactions, total, page, pageSize)
}

// GetDepositTransactions 获取保证金流水
// GET /shopower/account/deposit/transactions?page=1&page_size=20
func (h *AccountHandler) GetDepositTransactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	transactions, total, err := h.accountService.GetAccountTransactions(c.Request.Context(), models.AccountTypeDeposit, adminID, page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, transactions, total, page, pageSize)
}

// GetSettlements 获取结算记录
// GET /shopower/settlements?page=1&page_size=20
func (h *AccountHandler) GetSettlements(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	settlements, total, err := h.settlementService.GetSettlements(c.Request.Context(), adminID, "shop_owner", page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, settlements, total, page, pageSize)
}

// GetSettlementStats 获取结算统计
// GET /shopower/settlements/stats
func (h *AccountHandler) GetSettlementStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	stats, err := h.settlementService.GetSettlementStats(c.Request.Context(), adminID, "shop_owner")
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetCommissionAccount 获取佣金账户信息
// GET /shopower/account/commission
func (h *AccountHandler) GetCommissionAccount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	account, err := h.accountService.GetOrCreateShopOwnerCommissionAccount(c.Request.Context(), adminID)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, account)
}

// GetCommissionTransactions 获取佣金流水
// GET /shopower/account/commission/transactions?page=1&page_size=20
func (h *AccountHandler) GetCommissionTransactions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	transactions, total, err := h.accountService.GetAccountTransactions(c.Request.Context(), models.AccountTypeShopOwnerCommission, adminID, page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, transactions, total, page, pageSize)
}

// GetAllAccounts 获取所有账户汇总
// GET /shopower/account/summary
func (h *AccountHandler) GetAllAccounts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	prepayment, _ := h.accountService.GetOrCreatePrepaymentAccount(c.Request.Context(), adminID)
	deposit, _ := h.accountService.GetOrCreateDepositAccount(c.Request.Context(), adminID)
	commission, _ := h.accountService.GetOrCreateShopOwnerCommissionAccount(c.Request.Context(), adminID)

	utils.Success(c, gin.H{
		"prepayment": prepayment,
		"deposit":    deposit,
		"commission": commission,
	})
}

// ApplyWithdraw 申请提现
// POST /shopower/withdraw/apply
func (h *AccountHandler) ApplyWithdraw(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	var req struct {
		AccountType         string  `json:"account_type" binding:"required"`
		Amount              float64 `json:"amount" binding:"required,gt=0"`
		CollectionAccountID uint64  `json:"collection_account_id" binding:"required"`
		Remark              string  `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 验证账户类型
	if req.AccountType != models.AccountTypeShopOwnerCommission && req.AccountType != models.AccountTypeDeposit {
		utils.BadRequest(c, "店主只能从佣金账户或保证金账户提现")
		return
	}

	amount := utils.ToDecimal(req.Amount)
	application, err := h.accountService.ApplyWithdraw(c.Request.Context(), adminID, req.AccountType, amount, req.CollectionAccountID, req.Remark)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, application)
}

// GetWithdrawApplications 获取提现申请列表
// GET /shopower/withdraw/list?status=-1&page=1&page_size=20
func (h *AccountHandler) GetWithdrawApplications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	applications, total, err := h.accountService.GetWithdrawApplications(c.Request.Context(), adminID, int8(status), page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, applications, total, page, pageSize)
}

// ApplyRecharge 申请充值 (线下充值)
// POST /shopower/recharge/apply
func (h *AccountHandler) ApplyRecharge(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	var req struct {
		AccountType   string  `json:"account_type" binding:"required"`
		Amount        float64 `json:"amount" binding:"required,gt=0"`
		PaymentMethod string  `json:"payment_method" binding:"required"`
		PaymentProof  string  `json:"payment_proof"`
		Remark        string  `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// 验证账户类型
	if req.AccountType != models.AccountTypePrepayment && req.AccountType != models.AccountTypeDeposit {
		utils.BadRequest(c, "只能充值预付款账户或保证金账户")
		return
	}

	amount := utils.ToDecimal(req.Amount)
	application, err := h.accountService.ApplyRecharge(c.Request.Context(), adminID, req.AccountType, amount, req.PaymentMethod, req.PaymentProof, req.Remark)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, application)
}

// GetRechargeApplications 获取充值申请列表
// GET /shopower/recharge/list?status=-1&page=1&page_size=20
func (h *AccountHandler) GetRechargeApplications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	applications, total, err := h.accountService.GetRechargeApplications(c.Request.Context(), adminID, int8(status), page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, applications, total, page, pageSize)
}
