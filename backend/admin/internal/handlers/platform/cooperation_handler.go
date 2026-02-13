package platform

import (
	"strconv"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CooperationHandler 合作管理处理器（店铺-运营分配）
type CooperationHandler struct {
	db *gorm.DB
}

// NewCooperationHandler 创建合作管理处理器
func NewCooperationHandler() *CooperationHandler {
	return &CooperationHandler{
		db: database.GetDB(),
	}
}

// CooperationListResponse 合作列表响应
type CooperationListResponse struct {
	ID                uint64    `json:"id"`
	ShopID            uint64    `json:"shop_id"`
	ShopName          string    `json:"shop_name"`
	ShopOwnerID       int64     `json:"shop_owner_id"`
	ShopOwnerName     string    `json:"shop_owner_name"`
	OperatorID        int64     `json:"operator_id"`
	OperatorName      string    `json:"operator_name"`
	Status            int8      `json:"status"`
	StatusText        string    `json:"status_text"`
	PlatformShareRate string    `json:"platform_share_rate"`
	OperatorShareRate string    `json:"operator_share_rate"`
	ShopOwnerShareRate string   `json:"shop_owner_share_rate"`
	AssignedAt        time.Time `json:"assigned_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// ListCooperations 获取合作列表
// GET /platform/cooperations?status=1&keyword=xxx&page=1&page_size=20
func (h *CooperationHandler) ListCooperations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var relations []models.ShopOperatorRelation
	var total int64

	query := h.db.Model(&models.ShopOperatorRelation{})
	if status != "" {
		s, _ := strconv.Atoi(status)
		query = query.Where("status = ?", s)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&relations).Error; err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	// 构建响应
	list := make([]CooperationListResponse, 0, len(relations))
	for _, r := range relations {
		item := CooperationListResponse{
			ID:          r.ID,
			ShopID:      r.ShopID,
			ShopOwnerID: r.ShopOwnerID,
			OperatorID:  r.OperatorID,
			Status:      r.Status,
			AssignedAt:  r.AssignedAt,
			CreatedAt:   r.CreatedAt,
		}

		// 获取店铺名称
		var shop models.Shop
		if h.db.Where("shop_id = ?", r.ShopID).First(&shop).Error == nil {
			item.ShopName = shop.ShopName
		}

		// 获取店主名称
		var shopOwner models.Admin
		if h.db.Where("id = ?", r.ShopOwnerID).First(&shopOwner).Error == nil {
			item.ShopOwnerName = shopOwner.UserName
		}

		// 获取运营名称
		var operator models.Admin
		if h.db.Where("id = ?", r.OperatorID).First(&operator).Error == nil {
			item.OperatorName = operator.UserName
		}

		// 获取分成配置
		var config models.ProfitShareConfig
		if h.db.Where("shop_id = ? AND operator_id = ? AND status = 1", r.ShopID, r.OperatorID).First(&config).Error == nil {
			item.PlatformShareRate = config.PlatformShareRate.String()
			item.OperatorShareRate = config.OperatorShareRate.String()
			item.ShopOwnerShareRate = config.ShopOwnerShareRate.String()
		} else {
			item.PlatformShareRate = "5.00"
			item.OperatorShareRate = "45.00"
			item.ShopOwnerShareRate = "50.00"
		}

		// 状态文本
		switch r.Status {
		case models.RelationStatusActive:
			item.StatusText = "合作中"
		case models.RelationStatusReleased:
			item.StatusText = "已取消"
		default:
			item.StatusText = "未知"
		}

		// 关键词过滤
		if keyword != "" {
			if !containsKeyword(item.ShopName, keyword) && !containsKeyword(item.ShopOwnerName, keyword) && !containsKeyword(item.OperatorName, keyword) {
				continue
			}
		}

		list = append(list, item)
	}

	utils.SuccessWithPage(c, list, total, page, pageSize)
}

func containsKeyword(s, keyword string) bool {
	return len(keyword) > 0 && len(s) > 0 && (s == keyword || len(s) >= len(keyword) && (s[:len(keyword)] == keyword || s[len(s)-len(keyword):] == keyword))
}

// CreateCooperationRequest 创建合作请求
type CreateCooperationRequest struct {
	ShopID             uint64  `json:"shop_id" binding:"required"`
	OperatorID         int64   `json:"operator_id" binding:"required"`
	PlatformShareRate  float64 `json:"platform_share_rate"`
	OperatorShareRate  float64 `json:"operator_share_rate"`
	ShopOwnerShareRate float64 `json:"shop_owner_share_rate"`
}

// CreateCooperation 创建合作关系
// POST /platform/cooperations
func (h *CooperationHandler) CreateCooperation(c *gin.Context) {
	var req CreateCooperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 检查店铺是否存在
	var shop models.Shop
	if err := h.db.Where("shop_id = ?", req.ShopID).First(&shop).Error; err != nil {
		utils.Error(c, 400, "店铺不存在")
		return
	}

	// 检查运营是否存在
	var operator models.Admin
	if err := h.db.Where("id = ? AND user_type = 5", req.OperatorID).First(&operator).Error; err != nil {
		utils.Error(c, 400, "运营不存在")
		return
	}

	// 检查是否已存在合作关系
	var existing models.ShopOperatorRelation
	if err := h.db.Where("shop_id = ? AND operator_id = ?", req.ShopID, req.OperatorID).First(&existing).Error; err == nil {
		if existing.Status == models.RelationStatusActive {
			utils.Error(c, 400, "合作关系已存在")
			return
		}
		// 重新激活
		existing.Status = models.RelationStatusActive
		existing.AssignedAt = time.Now()
		h.db.Save(&existing)
		utils.Success(c, existing)
		return
	}

	// 创建合作关系
	relation := &models.ShopOperatorRelation{
		ShopID:      req.ShopID,
		ShopOwnerID: shop.AdminID,
		OperatorID:  req.OperatorID,
		Status:      models.RelationStatusActive,
		AssignedAt:  time.Now(),
	}
	if err := h.db.Create(relation).Error; err != nil {
		utils.Error(c, 500, "创建合作关系失败: "+err.Error())
		return
	}

	// 创建分成配置
	platformRate := decimal.NewFromFloat(5.00)
	operatorRate := decimal.NewFromFloat(45.00)
	shopOwnerRate := decimal.NewFromFloat(50.00)

	if req.PlatformShareRate > 0 {
		platformRate = decimal.NewFromFloat(req.PlatformShareRate)
	}
	if req.OperatorShareRate > 0 {
		operatorRate = decimal.NewFromFloat(req.OperatorShareRate)
	}
	if req.ShopOwnerShareRate > 0 {
		shopOwnerRate = decimal.NewFromFloat(req.ShopOwnerShareRate)
	}

	config := &models.ProfitShareConfig{
		ShopID:             req.ShopID,
		OperatorID:         req.OperatorID,
		PlatformShareRate:  platformRate,
		OperatorShareRate:  operatorRate,
		ShopOwnerShareRate: shopOwnerRate,
		Status:             1,
		EffectiveFrom:      time.Now(),
	}
	h.db.Create(config)

	utils.Success(c, relation)
}

// UpdateCooperationRequest 更新合作请求
type UpdateCooperationRequest struct {
	Status             *int8   `json:"status"`
	PlatformShareRate  float64 `json:"platform_share_rate"`
	OperatorShareRate  float64 `json:"operator_share_rate"`
	ShopOwnerShareRate float64 `json:"shop_owner_share_rate"`
}

// UpdateCooperation 更新合作关系
// PUT /platform/cooperations/:id
func (h *CooperationHandler) UpdateCooperation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, "无效的ID")
		return
	}

	var req UpdateCooperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	var relation models.ShopOperatorRelation
	if err := h.db.First(&relation, id).Error; err != nil {
		utils.Error(c, 404, "合作关系不存在")
		return
	}

	// 更新状态
	if req.Status != nil {
		relation.Status = *req.Status
	}
	h.db.Save(&relation)

	// 更新分成配置
	if req.PlatformShareRate > 0 || req.OperatorShareRate > 0 || req.ShopOwnerShareRate > 0 {
		var config models.ProfitShareConfig
		if h.db.Where("shop_id = ? AND operator_id = ? AND status = 1", relation.ShopID, relation.OperatorID).First(&config).Error == nil {
			if req.PlatformShareRate > 0 {
				config.PlatformShareRate = decimal.NewFromFloat(req.PlatformShareRate)
			}
			if req.OperatorShareRate > 0 {
				config.OperatorShareRate = decimal.NewFromFloat(req.OperatorShareRate)
			}
			if req.ShopOwnerShareRate > 0 {
				config.ShopOwnerShareRate = decimal.NewFromFloat(req.ShopOwnerShareRate)
			}
			h.db.Save(&config)
		}
	}

	utils.Success(c, relation)
}

// CancelCooperation 取消合作关系
// DELETE /platform/cooperations/:id
func (h *CooperationHandler) CancelCooperation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, 400, "无效的ID")
		return
	}

	var relation models.ShopOperatorRelation
	if err := h.db.First(&relation, id).Error; err != nil {
		utils.Error(c, 404, "合作关系不存在")
		return
	}

	relation.Status = models.RelationStatusReleased
	h.db.Save(&relation)

	utils.Success(c, gin.H{"message": "合作关系已取消"})
}

// GetCooperationStats 获取合作统计
// GET /platform/cooperations/stats
func (h *CooperationHandler) GetCooperationStats(c *gin.Context) {
	var total, active, cancelled int64

	h.db.Model(&models.ShopOperatorRelation{}).Count(&total)
	h.db.Model(&models.ShopOperatorRelation{}).Where("status = ?", models.RelationStatusActive).Count(&active)
	h.db.Model(&models.ShopOperatorRelation{}).Where("status = ?", models.RelationStatusReleased).Count(&cancelled)

	utils.Success(c, gin.H{
		"total":     total,
		"active":    active,
		"cancelled": cancelled,
	})
}

// GetOperatorList 获取运营列表（用于下拉选择）
// GET /platform/operators
func (h *CooperationHandler) GetOperatorList(c *gin.Context) {
	var operators []models.Admin
	if err := h.db.Where("user_type = 5 AND status = 1").Find(&operators).Error; err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	list := make([]gin.H, 0, len(operators))
	for _, op := range operators {
		list = append(list, gin.H{
			"id":       op.ID,
			"username": op.UserName,
			"email":    op.Email,
		})
	}

	utils.Success(c, list)
}

// GetShopOwnerList 获取店主列表（用于下拉选择）
// GET /platform/shop-owners
func (h *CooperationHandler) GetShopOwnerList(c *gin.Context) {
	var owners []models.Admin
	if err := h.db.Where("user_type = 1 AND status = 1").Find(&owners).Error; err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	list := make([]gin.H, 0, len(owners))
	for _, owner := range owners {
		list = append(list, gin.H{
			"id":       owner.ID,
			"username": owner.UserName,
			"email":    owner.Email,
		})
	}

	utils.Success(c, list)
}
