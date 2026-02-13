package platform

import (
	"fmt"
	"strconv"

	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PenaltyHandler 罚补账户处理器
type PenaltyHandler struct {
	db *gorm.DB
}

// NewPenaltyHandler 创建罚补账户处理器
func NewPenaltyHandler() *PenaltyHandler {
	return &PenaltyHandler{
		db: database.GetDB(),
	}
}

// GetPenaltyStats 获取罚补账户统计 - 遍历所有分表
// GET /platform/penalty/stats
func (h *PenaltyHandler) GetPenaltyStats(c *gin.Context) {
	var balance decimal.Decimal

	// 遍历所有分表统计
	for i := 0; i < database.ShardCount; i++ {
		settlementTable := fmt.Sprintf("order_settlements_%d", i)
		var b decimal.Decimal
		h.db.Table(settlementTable).
			Where("status = ?", models.OrderSettlementCompleted).
			Select("COALESCE(SUM(platform_share), 0)").Scan(&b)
		balance = balance.Add(b)
	}

	utils.Success(c, gin.H{
		"balance": balance,
	})
}

// GetPenaltyList 获取罚补交易列表
// GET /platform/penalty/list?type=all&page=1&page_size=20
func (h *PenaltyHandler) GetPenaltyList(c *gin.Context) {
	transType := c.DefaultQuery("type", "all")
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
		query := h.db.Table(txTable)

		switch transType {
		case "recharge":
			query = query.Where("transaction_type = ?", "recharge")
		case "withdraw":
			query = query.Where("transaction_type = ?", "withdraw")
		case "penalty":
			query = query.Where("transaction_type = ?", "penalty")
		case "subsidy":
			query = query.Where("transaction_type = ?", "subsidy")
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

		role := "店主"
		if admin.UserType == 5 {
			role = "运营"
		} else if admin.UserType == 9 {
			role = "平台"
		}

		list = append(list, gin.H{
			"date":       tx.CreatedAt.Format("2006-01-02 15:04:05"),
			"role":       role,
			"name":       admin.UserName,
			"type":       tx.TransactionType,
			"channel":    "系统",
			"order_no":   tx.TransactionNo,
			"amount":     tx.Amount,
			"balance":    tx.BalanceAfter,
			"status":     "已完成",
		})
	}

	utils.SuccessWithPage(c, list, total, page, pageSize)
}

// CreatePenaltyRequest 创建罚款/补贴请求
type CreatePenaltyRequest struct {
	AdminID int64   `json:"admin_id" binding:"required"`
	Type    string  `json:"type" binding:"required"` // penalty/subsidy
	Amount  float64 `json:"amount" binding:"required,gt=0"`
	Remark  string  `json:"remark"`
}

// CreatePenalty 创建罚款/补贴
// POST /platform/penalty/create
func (h *PenaltyHandler) CreatePenalty(c *gin.Context) {
	var req CreatePenaltyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	amount := decimal.NewFromFloat(req.Amount)
	if req.Type == "penalty" {
		amount = amount.Neg() // 罚款为负数
	}

	// 获取用户账户
	var account models.PrepaymentAccount
	if err := h.db.Where("admin_id = ?", req.AdminID).First(&account).Error; err != nil {
		utils.Error(c, 400, "用户账户不存在")
		return
	}

	// 创建交易记录
	tx := &models.AccountTransaction{
		TransactionNo:   generateTransactionNo("PN"),
		AccountType:     models.AccountTypePrepayment,
		AdminID:         req.AdminID,
		TransactionType: req.Type,
		Amount:          amount,
		BalanceBefore:   account.Balance,
		BalanceAfter:    account.Balance.Add(amount),
		Remark:          req.Remark,
		Status:          1,
	}

	txTable := database.GetAccountTransactionTableName(req.AdminID)
	err := h.db.Transaction(func(dbTx *gorm.DB) error {
		if err := dbTx.Table(txTable).Create(tx).Error; err != nil {
			return err
		}
		account.Balance = account.Balance.Add(amount)
		return dbTx.Save(&account).Error
	})

	if err != nil {
		utils.Error(c, 500, "操作失败: "+err.Error())
		return
	}

	utils.Success(c, tx)
}
