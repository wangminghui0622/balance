package platform

import (
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// SyncHandler 同步管理处理器（平台管理员专用）
type SyncHandler struct{}

// NewSyncHandler 创建同步管理处理器
func NewSyncHandler() *SyncHandler {
	return &SyncHandler{}
}

// GetSyncStats 获取同步统计
// GET /sync/stats
func (h *SyncHandler) GetSyncStats(c *gin.Context) {
	db := database.GetDB()

	var totalShops int64
	var enabledShops int64
	var pausedShops int64
	var totalSynced int64

	db.Model(&models.ShopSyncRecord{}).Where("sync_type = ?", models.SyncTypeFinanceIncome).Count(&totalShops)
	db.Model(&models.ShopSyncRecord{}).Where("sync_type = ? AND status = ?", models.SyncTypeFinanceIncome, models.SyncStatusEnabled).Count(&enabledShops)
	db.Model(&models.ShopSyncRecord{}).Where("sync_type = ? AND status = ?", models.SyncTypeFinanceIncome, models.SyncStatusPaused).Count(&pausedShops)
	db.Model(&models.ShopSyncRecord{}).Where("sync_type = ?", models.SyncTypeFinanceIncome).Select("COALESCE(SUM(total_synced_count), 0)").Scan(&totalSynced)

	// 获取最近失败的店铺
	var failedRecords []models.ShopSyncRecord
	db.Where("sync_type = ? AND consecutive_fail_count > 0", models.SyncTypeFinanceIncome).
		Order("consecutive_fail_count DESC").
		Limit(10).
		Find(&failedRecords)

	utils.Success(c, gin.H{
		"total_shops":    totalShops,
		"enabled_shops":  enabledShops,
		"paused_shops":   pausedShops,
		"total_synced":   totalSynced,
		"failed_records": failedRecords,
	})
}

// ListSyncRecords 获取同步记录列表
// GET /sync/records?sync_type=finance_income&status=1&page=1&page_size=20
func (h *SyncHandler) ListSyncRecords(c *gin.Context) {
	db := database.GetDB()

	syncType := c.DefaultQuery("sync_type", models.SyncTypeFinanceIncome)
	status := c.Query("status")
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if v, err := parseInt(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := parseInt(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	query := db.Model(&models.ShopSyncRecord{}).Where("sync_type = ?", syncType)
	if status != "" {
		if s, err := parseInt(status); err == nil {
			query = query.Where("status = ?", s)
		}
	}

	var total int64
	query.Count(&total)

	var records []models.ShopSyncRecord
	offset := (page - 1) * pageSize
	query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&records)

	utils.SuccessWithPage(c, records, total, page, pageSize)
}

// ResetSyncRecord 重置同步记录（恢复暂停的店铺）
// POST /sync/records/:shop_id/reset
func (h *SyncHandler) ResetSyncRecord(c *gin.Context) {
	db := database.GetDB()

	shopID := c.Param("shop_id")
	syncType := c.DefaultQuery("sync_type", models.SyncTypeFinanceIncome)

	result := db.Model(&models.ShopSyncRecord{}).
		Where("shop_id = ? AND sync_type = ?", shopID, syncType).
		Updates(map[string]interface{}{
			"status":                 models.SyncStatusEnabled,
			"consecutive_fail_count": 0,
			"last_error":             "",
		})

	if result.RowsAffected == 0 {
		utils.Error(c, 404, "同步记录不存在")
		return
	}

	utils.Success(c, gin.H{"message": "重置成功"})
}

// parseInt 解析整数
func parseInt(s string) (int, error) {
	var v int
	_, err := parseIntHelper(s, &v)
	return v, err
}

func parseIntHelper(s string, v *int) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	*v = n
	return n, nil
}
