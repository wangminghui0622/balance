package platform

import (
	"fmt"
	"strconv"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// FinanceAuditHandler 财务审核处理器
type FinanceAuditHandler struct {
	db *gorm.DB
}

// NewFinanceAuditHandler 创建财务审核处理器
func NewFinanceAuditHandler() *FinanceAuditHandler {
	return &FinanceAuditHandler{
		db: database.GetDB(),
	}
}

// GetAuditStats 获取审核统计 - 遍历所有分表
// GET /platform/finance/audit/stats?type=withdraw
func (h *FinanceAuditHandler) GetAuditStats(c *gin.Context) {
	auditType := c.DefaultQuery("type", "withdraw")

	var pending, approved int64

	// 遍历所有分表统计
	for i := 0; i < database.ShardCount; i++ {
		txTable := fmt.Sprintf("account_transactions_%d", i)
		var p, a int64

		switch auditType {
		case "withdraw":
			h.db.Table(txTable).Where("transaction_type = ? AND status = 0", "withdraw").Count(&p)
			h.db.Table(txTable).Where("transaction_type = ? AND status = 1", "withdraw").Count(&a)
		case "recharge":
			h.db.Table(txTable).Where("transaction_type = ? AND status = 0", "recharge").Count(&p)
			h.db.Table(txTable).Where("transaction_type = ? AND status = 1", "recharge").Count(&a)
		}
		pending += p
		approved += a
	}

	utils.Success(c, gin.H{
		"pending":      pending,
		"approved":     approved,
		"avg_time":     "02:30",
		"time_percent": 15,
	})
}

// GetWithdrawAuditList 获取提现审核列表 - 遍历所有分表
// GET /platform/finance/audit/withdraw?status=pending&page=1&page_size=20
func (h *FinanceAuditHandler) GetWithdrawAuditList(c *gin.Context) {
	statusParam := c.DefaultQuery("status", "")
	_ = c.Query("keyword") // 暂未使用
	withdrawType := c.Query("withdraw_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var allTransactions []models.AccountTransaction
	var total int64

	// 遍历所有分表
	for i := 0; i < database.ShardCount; i++ {
		txTable := fmt.Sprintf("account_transactions_%d", i)
		query := h.db.Table(txTable).Where("transaction_type = ?", "withdraw")

		if statusParam == "pending" {
			query = query.Where("status = 0")
		} else if statusParam == "approved" {
			query = query.Where("status = 1")
		}

		if withdrawType != "" {
			query = query.Where("account_type = ?", withdrawType)
		}

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
	transactions := allTransactions

	// 补充用户信息
	list := make([]gin.H, 0, len(transactions))
	for _, tx := range transactions {
		var admin models.Admin
		h.db.Where("id = ?", tx.AdminID).First(&admin)

		statusText := "待审批"
		if tx.Status == 1 {
			statusText = "已审批"
		}

		list = append(list, gin.H{
			"id":              tx.ID,
			"transaction_no":  tx.TransactionNo,
			"admin_id":        tx.AdminID,
			"username":        admin.UserName,
			"account_type":    tx.AccountType,
			"amount":          tx.Amount,
			"status":          tx.Status,
			"status_text":     statusText,
			"remark":          tx.Remark,
			"created_at":      tx.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	utils.SuccessWithPage(c, list, total, page, pageSize)
}

// GetRechargeAuditList 获取充值审核列表 - 遍历所有分表
// GET /platform/finance/audit/recharge?status=pending&page=1&page_size=20
func (h *FinanceAuditHandler) GetRechargeAuditList(c *gin.Context) {
	statusParam := c.DefaultQuery("status", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var allTransactions []models.AccountTransaction
	var total int64

	// 遍历所有分表
	for i := 0; i < database.ShardCount; i++ {
		txTable := fmt.Sprintf("account_transactions_%d", i)
		query := h.db.Table(txTable).Where("transaction_type = ?", "recharge")

		if statusParam == "pending" {
			query = query.Where("status = 0")
		} else if statusParam == "approved" {
			query = query.Where("status = 1")
		}

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
	transactions := allTransactions

	list := make([]gin.H, 0, len(transactions))
	for _, tx := range transactions {
		var admin models.Admin
		h.db.Where("id = ?", tx.AdminID).First(&admin)

		statusText := "待审批"
		if tx.Status == 1 {
			statusText = "已审批"
		}

		list = append(list, gin.H{
			"id":              tx.ID,
			"transaction_no":  tx.TransactionNo,
			"admin_id":        tx.AdminID,
			"username":        admin.UserName,
			"account_type":    tx.AccountType,
			"amount":          tx.Amount,
			"status":          tx.Status,
			"status_text":     statusText,
			"remark":          tx.Remark,
			"created_at":      tx.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	utils.SuccessWithPage(c, list, total, page, pageSize)
}

// ApproveAuditRequest 审批请求
type ApproveAuditRequest struct {
	TransactionID uint64 `json:"transaction_id" binding:"required"`
	Action        string `json:"action" binding:"required"` // approve/reject
	Remark        string `json:"remark"`
}

// ApproveAudit 审批操作
// POST /platform/finance/audit/approve
func (h *FinanceAuditHandler) ApproveAudit(c *gin.Context) {
	var req ApproveAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 遍历所有分表查找交易记录
	var tx models.AccountTransaction
	var foundTable string
	for i := 0; i < database.ShardCount; i++ {
		txTable := fmt.Sprintf("account_transactions_%d", i)
		if err := h.db.Table(txTable).First(&tx, req.TransactionID).Error; err == nil {
			foundTable = txTable
			break
		}
	}
	if foundTable == "" {
		utils.Error(c, 404, "交易记录不存在")
		return
	}

	if tx.Status != 0 {
		utils.Error(c, 400, "该记录已审批")
		return
	}

	newStatus := int8(0)
	if req.Action == "approve" {
		newStatus = 1
	} else if req.Action == "reject" {
		newStatus = 2
	} else {
		utils.Error(c, 400, "无效的操作")
		return
	}

	updates := map[string]interface{}{"status": newStatus}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}

	h.db.Table(foundTable).Where("id = ?", tx.ID).Updates(updates)

	utils.Success(c, gin.H{"message": "审批成功"})
}

// CreateWithdrawRequest 创建提现申请请求
type CreateWithdrawRequest struct {
	AccountType string  `json:"account_type" binding:"required"` // prepayment/deposit/operator
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Remark      string  `json:"remark"`
}

// CreateWithdrawApplication 创建提现申请
// POST /platform/finance/withdraw/apply
func (h *FinanceAuditHandler) CreateWithdrawApplication(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	var req CreateWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	amount := decimal.NewFromFloat(req.Amount)

	// 检查余额
	var balance decimal.Decimal
	switch req.AccountType {
	case "prepayment":
		var account models.PrepaymentAccount
		if err := h.db.Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			utils.Error(c, 400, "账户不存在")
			return
		}
		balance = account.Balance
	case "deposit":
		var account models.DepositAccount
		if err := h.db.Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			utils.Error(c, 400, "账户不存在")
			return
		}
		balance = account.Balance
	case "operator":
		var account models.OperatorAccount
		if err := h.db.Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			utils.Error(c, 400, "账户不存在")
			return
		}
		balance = account.Balance
	default:
		utils.Error(c, 400, "无效的账户类型")
		return
	}

	if balance.LessThan(amount) {
		utils.Error(c, 400, "余额不足")
		return
	}

	// 创建提现申请
	tx := &models.AccountTransaction{
		TransactionNo:   generateTransactionNo("WD"),
		AccountType:     req.AccountType,
		AdminID:         adminID,
		TransactionType: "withdraw",
		Amount:          amount.Neg(), // 负数表示支出
		BalanceBefore:   balance,
		BalanceAfter:    balance.Sub(amount),
		Remark:          req.Remark,
		Status:          0, // 待审批
	}

	txTable := database.GetAccountTransactionTableName(adminID)
	if err := h.db.Table(txTable).Create(tx).Error; err != nil {
		utils.Error(c, 500, "创建提现申请失败: "+err.Error())
		return
	}

	utils.Success(c, tx)
}

func generateTransactionNo(prefix string) string {
	return prefix + time.Now().Format("20060102150405") + strconv.FormatInt(time.Now().UnixNano()%10000, 10)
}
