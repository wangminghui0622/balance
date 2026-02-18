package shopower

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"balance/backend/internal/models"
	"balance/backend/internal/services"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// AccountHandler 账户处理器
type AccountHandler struct {
	accountService          *services.AccountService
	settlementService       *services.SettlementService
	prepaymentCheckService  *services.PrepaymentCheckService
}

// NewAccountHandler 创建账户处理器
func NewAccountHandler() *AccountHandler {
	return &AccountHandler{
		accountService:         services.NewAccountService(),
		settlementService:      services.NewSettlementService(),
		prepaymentCheckService: services.NewPrepaymentCheckService(),
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
// GET /shopower/account/prepayment/transactions?page=1&page_size=20&transaction_type=recharge
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

	// 支持按交易类型过滤，多个用逗号分隔，如 transaction_type=recharge 或 transaction_type=profit_share,cost_settle
	var txTypes []string
	if t := c.Query("transaction_type"); t != "" {
		txTypes = strings.Split(t, ",")
	}

	transactions, total, err := h.accountService.GetAccountTransactions(c.Request.Context(), models.AccountTypePrepayment, adminID, page, pageSize, txTypes...)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, transactions, total, page, pageSize)
}

// GetDepositTransactions 获取保证金流水
// GET /shopower/account/deposit/transactions?page=1&page_size=20&transaction_type=deposit_pay
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

	var txTypes []string
	if t := c.Query("transaction_type"); t != "" {
		txTypes = strings.Split(t, ",")
	}

	transactions, total, err := h.accountService.GetAccountTransactions(c.Request.Context(), models.AccountTypeDeposit, adminID, page, pageSize, txTypes...)
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

// Recharge 充值（直接到账，无需审核）
// POST /shopower/recharge
func (h *AccountHandler) Recharge(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	var req struct {
		AccountType   string  `json:"account_type" binding:"required"`
		Amount        float64 `json:"amount" binding:"required,gt=0"`
		PaymentMethod string  `json:"payment_method"`
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
	remark := req.Remark
	if remark == "" {
		remark = "店主充值"
	}

	// 充值无需审核，直接入账
	var transaction interface{}
	var err error
	switch req.AccountType {
	case models.AccountTypePrepayment:
		transaction, err = h.accountService.RechargePrepayment(c.Request.Context(), adminID, amount, remark, adminID)
	case models.AccountTypeDeposit:
		transaction, err = h.accountService.PayDeposit(c.Request.Context(), adminID, amount, remark, adminID)
	}
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	// 充值预付款成功后，异步补扣历史「预付款不足」的订单
	if req.AccountType == models.AccountTypePrepayment {
		go func(ownerID int64) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			okCount, failCount, backfillErr := h.prepaymentCheckService.BackfillInsufficientOrders(bgCtx, ownerID)
			if backfillErr != nil {
				fmt.Printf("[Recharge] 预付款补扣异常: adminID=%d, err=%v\n", ownerID, backfillErr)
			}
			if okCount > 0 || failCount > 0 {
				fmt.Printf("[Recharge] 预付款补扣完成: adminID=%d, 成功=%d, 失败=%d\n", ownerID, okCount, failCount)
			}
		}(adminID)
	}

	utils.Success(c, transaction)
}

// GetRechargeRecords 获取充值申请列表
// GET /shopower/recharge/list?status=-1&page=1&page_size=20
func (h *AccountHandler) GetRechargeRecords(c *gin.Context) {
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

	applications, total, err := h.accountService.GetRechargeRecords(c.Request.Context(), adminID, int8(status), page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, applications, total, page, pageSize)
}
