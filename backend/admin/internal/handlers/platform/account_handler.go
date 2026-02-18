package platform

import (
	"fmt"
	"strconv"

	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/services"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AccountHandler 平台账户管理处理器
type AccountHandler struct {
	db             *gorm.DB
	accountService *services.AccountService
}

// NewAccountHandler 创建平台账户管理处理器
func NewAccountHandler() *AccountHandler {
	return &AccountHandler{
		db:             database.GetDB(),
		accountService: services.NewAccountService(),
	}
}

// ListPrepaymentAccounts 获取预付款账户列表
// GET /platform/accounts/prepayment?page=1&page_size=20
func (h *AccountHandler) ListPrepaymentAccounts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var accounts []models.PrepaymentAccount
	var total int64

	h.db.Model(&models.PrepaymentAccount{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := h.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&accounts).Error; err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	// 补充用户信息
	type AccountWithUser struct {
		models.PrepaymentAccount
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	list := make([]AccountWithUser, 0, len(accounts))
	for _, acc := range accounts {
		item := AccountWithUser{PrepaymentAccount: acc}
		var admin models.Admin
		if h.db.Where("id = ?", acc.AdminID).First(&admin).Error == nil {
			item.Username = admin.UserName
			item.Email = admin.Email
		}
		list = append(list, item)
	}

	utils.SuccessWithPage(c, list, total, page, pageSize)
}

// ListDepositAccounts 获取保证金账户列表
// GET /platform/accounts/deposit?page=1&page_size=20
func (h *AccountHandler) ListDepositAccounts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var accounts []models.DepositAccount
	var total int64

	h.db.Model(&models.DepositAccount{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := h.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&accounts).Error; err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	// 补充用户信息
	type AccountWithUser struct {
		models.DepositAccount
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	list := make([]AccountWithUser, 0, len(accounts))
	for _, acc := range accounts {
		item := AccountWithUser{DepositAccount: acc}
		var admin models.Admin
		if h.db.Where("id = ?", acc.AdminID).First(&admin).Error == nil {
			item.Username = admin.UserName
			item.Email = admin.Email
		}
		list = append(list, item)
	}

	utils.SuccessWithPage(c, list, total, page, pageSize)
}

// ListOperatorAccounts 获取运营账户列表
// GET /platform/accounts/operator?page=1&page_size=20
func (h *AccountHandler) ListOperatorAccounts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var accounts []models.OperatorAccount
	var total int64

	h.db.Model(&models.OperatorAccount{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := h.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&accounts).Error; err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	// 补充用户信息
	type AccountWithUser struct {
		models.OperatorAccount
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	list := make([]AccountWithUser, 0, len(accounts))
	for _, acc := range accounts {
		item := AccountWithUser{OperatorAccount: acc}
		var admin models.Admin
		if h.db.Where("id = ?", acc.AdminID).First(&admin).Error == nil {
			item.Username = admin.UserName
			item.Email = admin.Email
		}
		list = append(list, item)
	}

	utils.SuccessWithPage(c, list, total, page, pageSize)
}

// RechargeRequest 充值请求
type RechargeRequest struct {
	AdminID int64   `json:"admin_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
	Remark  string  `json:"remark"`
}

// RechargePrepayment 预付款充值
// POST /platform/accounts/prepayment/recharge
func (h *AccountHandler) RechargePrepayment(c *gin.Context) {
	var req RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	amount := decimal.NewFromFloat(req.Amount)
	remark := req.Remark
	if remark == "" {
		remark = "平台充值"
	}

	tx, err := h.accountService.RechargePrepayment(c.Request.Context(), req.AdminID, amount, remark, 0)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, tx)
}

// PayDepositRequest 缴纳保证金请求
type PayDepositRequest struct {
	AdminID int64   `json:"admin_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
	Remark  string  `json:"remark"`
}

// PayDeposit 缴纳保证金
// POST /platform/accounts/deposit/pay
func (h *AccountHandler) PayDeposit(c *gin.Context) {
	var req PayDepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	amount := decimal.NewFromFloat(req.Amount)
	remark := req.Remark
	if remark == "" {
		remark = "保证金缴纳"
	}

	tx, err := h.accountService.PayDeposit(c.Request.Context(), req.AdminID, amount, remark, 0)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, tx)
}

// GetAccountTransactions 获取账户流水 - 使用分表
// GET /platform/accounts/transactions?account_type=prepayment&admin_id=1&page=1&page_size=20
func (h *AccountHandler) GetAccountTransactions(c *gin.Context) {
	accountType := c.DefaultQuery("account_type", "prepayment")
	adminIDStr := c.Query("admin_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 如果指定了admin_id，直接查询对应分表
	if adminIDStr != "" {
		adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
		transactions, total, err := h.accountService.GetAccountTransactions(c.Request.Context(), accountType, adminID, page, pageSize)
		if err != nil {
			utils.Error(c, 500, err.Error())
			return
		}
		utils.SuccessWithPage(c, transactions, total, page, pageSize)
		return
	}

	// 未指定admin_id，遍历所有分表
	var allTransactions []models.AccountTransaction
	var total int64

	for i := 0; i < database.ShardCount; i++ {
		txTable := fmt.Sprintf("account_transactions_%d", i)
		query := h.db.Table(txTable).Where("account_type = ?", accountType)

		var count int64
		query.Count(&count)
		total += count

		var transactions []models.AccountTransaction
		query.Order("created_at DESC").Find(&transactions)
		allTransactions = append(allTransactions, transactions...)
	}

	// 内存分页
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset >= len(allTransactions) {
		allTransactions = []models.AccountTransaction{}
	} else {
		if end > len(allTransactions) {
			end = len(allTransactions)
		}
		allTransactions = allTransactions[offset:end]
	}

	utils.SuccessWithPage(c, allTransactions, total, page, pageSize)
}

// GetAccountStats 获取账户统计
// GET /platform/accounts/stats
func (h *AccountHandler) GetAccountStats(c *gin.Context) {
	var prepaymentTotal, depositTotal, operatorTotal decimal.Decimal

	h.db.Model(&models.PrepaymentAccount{}).Select("COALESCE(SUM(balance), 0)").Scan(&prepaymentTotal)
	h.db.Model(&models.DepositAccount{}).Select("COALESCE(SUM(balance), 0)").Scan(&depositTotal)
	h.db.Model(&models.OperatorAccount{}).Select("COALESCE(SUM(balance), 0)").Scan(&operatorTotal)

	var prepaymentCount, depositCount, operatorCount int64
	h.db.Model(&models.PrepaymentAccount{}).Count(&prepaymentCount)
	h.db.Model(&models.DepositAccount{}).Count(&depositCount)
	h.db.Model(&models.OperatorAccount{}).Count(&operatorCount)

	// 获取平台佣金账户
	platformCommission, _ := h.accountService.GetOrCreatePlatformCommissionAccount(c.Request.Context())
	// 获取店主佣金账户总额
	var shopOwnerCommissionTotal decimal.Decimal
	var shopOwnerCommissionCount int64
	h.db.Model(&models.ShopOwnerCommissionAccount{}).Select("COALESCE(SUM(balance), 0)").Scan(&shopOwnerCommissionTotal)
	h.db.Model(&models.ShopOwnerCommissionAccount{}).Count(&shopOwnerCommissionCount)

	utils.Success(c, gin.H{
		"prepayment": gin.H{
			"total_balance": prepaymentTotal,
			"account_count": prepaymentCount,
		},
		"deposit": gin.H{
			"total_balance": depositTotal,
			"account_count": depositCount,
		},
		"operator": gin.H{
			"total_balance": operatorTotal,
			"account_count": operatorCount,
		},
		"shop_owner_commission": gin.H{
			"total_balance": shopOwnerCommissionTotal,
			"account_count": shopOwnerCommissionCount,
		},
		"platform_commission": gin.H{
			"balance":        platformCommission.Balance,
			"total_earnings": platformCommission.TotalEarnings,
		},
	})
}

// GetWithdrawApplications 获取提现申请列表 (平台审核)
// GET /platform/withdraw/list?status=0&page=1&page_size=20
func (h *AccountHandler) GetWithdrawApplications(c *gin.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// adminID = 0 表示查询所有用户的提现申请
	applications, total, err := h.accountService.GetWithdrawApplications(c.Request.Context(), 0, int8(status), page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, applications, total, page, pageSize)
}

// ApproveWithdraw 审批通过提现
// POST /platform/withdraw/approve
func (h *AccountHandler) ApproveWithdraw(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	auditBy := userID.(int64)

	var req struct {
		ApplicationID uint64 `json:"application_id" binding:"required"`
		AuditRemark   string `json:"audit_remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	err := h.accountService.ApproveWithdraw(c.Request.Context(), req.ApplicationID, auditBy, req.AuditRemark)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}

// RejectWithdraw 拒绝提现
// POST /platform/withdraw/reject
func (h *AccountHandler) RejectWithdraw(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	auditBy := userID.(int64)

	var req struct {
		ApplicationID uint64 `json:"application_id" binding:"required"`
		AuditRemark   string `json:"audit_remark" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	err := h.accountService.RejectWithdraw(c.Request.Context(), req.ApplicationID, auditBy, req.AuditRemark)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}

// ConfirmWithdrawPaid 确认提现已打款
// POST /platform/withdraw/confirm_paid
func (h *AccountHandler) ConfirmWithdrawPaid(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	operatorID := userID.(int64)

	var req struct {
		ApplicationID uint64 `json:"application_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	err := h.accountService.ConfirmWithdrawPaid(c.Request.Context(), req.ApplicationID, operatorID)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetPlatformCommissionAccount 获取平台佣金账户
// GET /platform/account/commission
func (h *AccountHandler) GetPlatformCommissionAccount(c *gin.Context) {
	account, err := h.accountService.GetOrCreatePlatformCommissionAccount(c.Request.Context())
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, account)
}

// GetPlatformCommissionTransactions 获取平台佣金流水
// GET /platform/account/commission/transactions?page=1&page_size=20
func (h *AccountHandler) GetPlatformCommissionTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 平台佣金账户 adminID = 0
	transactions, total, err := h.accountService.GetAccountTransactions(c.Request.Context(), models.AccountTypePlatformCommission, 0, page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, transactions, total, page, pageSize)
}

// GetRechargeRecords 获取充值申请列表 (平台审核)
// GET /platform/recharge/list?status=0&page=1&page_size=20
func (h *AccountHandler) GetRechargeRecords(c *gin.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// adminID = 0 表示查询所有用户的充值申请
	applications, total, err := h.accountService.GetRechargeRecords(c.Request.Context(), 0, int8(status), page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.SuccessWithPage(c, applications, total, page, pageSize)
}

// ApproveRecharge 审批通过充值
// POST /platform/recharge/approve
func (h *AccountHandler) ApproveRecharge(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	auditBy := userID.(int64)

	var req struct {
		ApplicationID uint64 `json:"application_id" binding:"required"`
		AuditRemark   string `json:"audit_remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	err := h.accountService.ApproveRecharge(c.Request.Context(), req.ApplicationID, auditBy, req.AuditRemark)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}

// RejectRecharge 拒绝充值
// POST /platform/recharge/reject
func (h *AccountHandler) RejectRecharge(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	auditBy := userID.(int64)

	var req struct {
		ApplicationID uint64 `json:"application_id" binding:"required"`
		AuditRemark   string `json:"audit_remark" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	err := h.accountService.RejectRecharge(c.Request.Context(), req.ApplicationID, auditBy, req.AuditRemark)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}
