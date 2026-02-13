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

// CommissionHandler 平台佣金管理处理器
type CommissionHandler struct {
	db *gorm.DB
}

// NewCommissionHandler 创建平台佣金管理处理器
func NewCommissionHandler() *CommissionHandler {
	return &CommissionHandler{
		db: database.GetDB(),
	}
}

// GetCommissionStats 获取佣金统计 - 遍历所有分表
// GET /platform/commission/stats
func (h *CommissionHandler) GetCommissionStats(c *gin.Context) {
	var totalCommission, withdrawable, pending decimal.Decimal

	// 遍历所有分表统计
	for i := 0; i < database.ShardCount; i++ {
		settlementTable := fmt.Sprintf("order_settlements_%d", i)
		shipmentRecordTable := fmt.Sprintf("order_shipment_records_%d", i)

		var tc, p decimal.Decimal
		h.db.Table(settlementTable).
			Where("status = ?", models.OrderSettlementCompleted).
			Select("COALESCE(SUM(platform_share), 0)").Scan(&tc)
		totalCommission = totalCommission.Add(tc)

		h.db.Table(shipmentRecordTable).
			Where("status = ?", models.ShipmentRecordStatusShipped).
			Select("COALESCE(SUM(total_cost * 0.05), 0)").Scan(&p)
		pending = pending.Add(p)
	}

	withdrawable = totalCommission

	utils.Success(c, gin.H{
		"total_commission": totalCommission,
		"withdrawable":     withdrawable,
		"pending":          pending,
	})
}

// GetCommissionList 获取佣金列表
// GET /platform/commission/list?type=all&page=1&page_size=20
func (h *CommissionHandler) GetCommissionList(c *gin.Context) {
	transType := c.DefaultQuery("type", "all")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	if transType == "withdraw" {
		utils.SuccessWithPage(c, []interface{}{}, 0, page, pageSize)
		return
	} else if transType == "adjustment" {
		utils.SuccessWithPage(c, []interface{}{}, 0, page, pageSize)
		return
	}

	var allSettlements []models.OrderSettlement
	var total int64

	// 遍历所有分表
	for i := 0; i < database.ShardCount; i++ {
		settlementTable := fmt.Sprintf("order_settlements_%d", i)
		query := h.db.Table(settlementTable).Where("status = ?", models.OrderSettlementCompleted)

		var count int64
		query.Count(&count)
		total += count

		var settlements []models.OrderSettlement
		query.Order("created_at DESC").Find(&settlements)
		allSettlements = append(allSettlements, settlements...)
	}

	// 内存分页
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset >= len(allSettlements) {
		allSettlements = []models.OrderSettlement{}
	} else {
		if end > len(allSettlements) {
			end = len(allSettlements)
		}
		allSettlements = allSettlements[offset:end]
	}
	settlements := allSettlements

	// 转换为前端需要的格式
	list := make([]gin.H, 0, len(settlements))
	for _, s := range settlements {
		list = append(list, gin.H{
			"date":     s.CreatedAt.Format("2006-01-02 15:04:05"),
			"type":     "佣金",
			"store_id": s.ShopID,
			"order_no": s.OrderSN,
			"amount":   s.PlatformShare,
			"balance":  s.PlatformShare, // 简化处理
			"status":   "已结算",
		})
	}

	utils.SuccessWithPage(c, list, total, page, pageSize)
}
