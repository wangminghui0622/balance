package platform

import (
	"strconv"

	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CollectionHandler 收款账户处理器
type CollectionHandler struct {
	db *gorm.DB
}

// NewCollectionHandler 创建收款账户处理器
func NewCollectionHandler() *CollectionHandler {
	return &CollectionHandler{
		db: database.GetDB(),
	}
}


// GetCollectionAccounts 获取收款账户列表
// GET /platform/collection/accounts?admin_id=1
func (h *CollectionHandler) GetCollectionAccounts(c *gin.Context) {
	adminIDStr := c.Query("admin_id")
	
	var adminID int64
	if adminIDStr != "" {
		adminID, _ = strconv.ParseInt(adminIDStr, 10, 64)
	} else {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.Unauthorized(c, "未登录")
			return
		}
		adminID = userID.(int64)
	}

	var wallets, banks []models.CollectionAccount

	h.db.Where("admin_id = ? AND account_type = ?", adminID, "wallet").Find(&wallets)
	h.db.Where("admin_id = ? AND account_type = ?", adminID, "bank").Find(&banks)

	// 转换格式
	walletList := make([]gin.H, 0, len(wallets))
	for _, w := range wallets {
		status := "已激活"
		if w.Status == 2 {
			status = "未激活"
		}
		walletList = append(walletList, gin.H{
			"id":         w.ID,
			"name":       w.AccountName,
			"account":    w.AccountNo,
			"payee":      w.Payee,
			"status":     status,
			"is_default": w.IsDefault,
		})
	}

	bankList := make([]gin.H, 0, len(banks))
	for _, b := range banks {
		status := "已激活"
		if b.Status == 2 {
			status = "未激活"
		}
		bankList = append(bankList, gin.H{
			"id":          b.ID,
			"name":        b.BankName,
			"account":     b.AccountNo,
			"payee":       b.Payee,
			"bank_branch": b.BankBranch,
			"status":      status,
			"is_default":  b.IsDefault,
		})
	}

	utils.Success(c, gin.H{
		"wallets": walletList,
		"banks":   bankList,
	})
}

// CreateCollectionAccountRequest 创建收款账户请求
type CreateCollectionAccountRequest struct {
	AccountType string `json:"account_type" binding:"required"` // wallet/bank
	AccountName string `json:"account_name" binding:"required"`
	AccountNo   string `json:"account_no" binding:"required"`
	BankName    string `json:"bank_name"`
	BankBranch  string `json:"bank_branch"`
	Payee       string `json:"payee" binding:"required"`
	IsDefault   bool   `json:"is_default"`
}

// CreateCollectionAccount 创建收款账户
// POST /platform/collection/accounts
func (h *CollectionHandler) CreateCollectionAccount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Unauthorized(c, "未登录")
		return
	}
	adminID := userID.(int64)

	var req CreateCollectionAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	account := &models.CollectionAccount{
		AdminID:     adminID,
		AccountType: req.AccountType,
		AccountName: req.AccountName,
		AccountNo:   req.AccountNo,
		BankName:    req.BankName,
		BankBranch:  req.BankBranch,
		Payee:       req.Payee,
		IsDefault:   req.IsDefault,
		Status:      1,
	}

	// 如果设为默认，取消其他默认
	if req.IsDefault {
		h.db.Model(&models.CollectionAccount{}).
			Where("admin_id = ? AND account_type = ?", adminID, req.AccountType).
			Update("is_default", false)
	}

	if err := h.db.Create(account).Error; err != nil {
		utils.Error(c, 500, "创建失败: "+err.Error())
		return
	}

	utils.Success(c, account)
}

// UpdateCollectionAccount 更新收款账户
// PUT /platform/collection/accounts/:id
func (h *CollectionHandler) UpdateCollectionAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, "无效的ID")
		return
	}

	var account models.CollectionAccount
	if err := h.db.First(&account, id).Error; err != nil {
		utils.Error(c, 404, "账户不存在")
		return
	}

	var req CreateCollectionAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	account.AccountName = req.AccountName
	account.AccountNo = req.AccountNo
	account.BankName = req.BankName
	account.BankBranch = req.BankBranch
	account.Payee = req.Payee

	if req.IsDefault && !account.IsDefault {
		h.db.Model(&models.CollectionAccount{}).
			Where("admin_id = ? AND account_type = ?", account.AdminID, account.AccountType).
			Update("is_default", false)
		account.IsDefault = true
	}

	h.db.Save(&account)

	utils.Success(c, account)
}

// DeleteCollectionAccount 删除收款账户
// DELETE /platform/collection/accounts/:id
func (h *CollectionHandler) DeleteCollectionAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, "无效的ID")
		return
	}

	if err := h.db.Delete(&models.CollectionAccount{}, id).Error; err != nil {
		utils.Error(c, 500, "删除失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// SetDefaultAccount 设置默认账户
// POST /platform/collection/accounts/:id/default
func (h *CollectionHandler) SetDefaultAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, "无效的ID")
		return
	}

	var account models.CollectionAccount
	if err := h.db.First(&account, id).Error; err != nil {
		utils.Error(c, 404, "账户不存在")
		return
	}

	// 取消其他默认
	h.db.Model(&models.CollectionAccount{}).
		Where("admin_id = ? AND account_type = ?", account.AdminID, account.AccountType).
		Update("is_default", false)

	// 设置当前为默认
	account.IsDefault = true
	h.db.Save(&account)

	utils.Success(c, gin.H{"message": "设置成功"})
}
