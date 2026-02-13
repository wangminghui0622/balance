package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/models"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	statsSchedulerLock = "stats:scheduler:lock"
	statsLockTTL       = 10 * time.Minute
)

// StatsService 统计服务
type StatsService struct {
	db  *gorm.DB
	rdb *redis.Client
}

// NewStatsService 创建统计服务
func NewStatsService() *StatsService {
	return &StatsService{
		db:  database.GetDB(),
		rdb: database.GetRedis(),
	}
}

// GenerateDailyStats 生成每日统计（凌晨执行，统计前一天数据）
func (s *StatsService) GenerateDailyStats(ctx context.Context) error {
	// 获取分布式锁
	ok, err := s.rdb.SetNX(ctx, statsSchedulerLock, "1", statsLockTTL).Result()
	if err != nil {
		return fmt.Errorf("获取统计锁失败: %w", err)
	}
	if !ok {
		log.Println("[Stats] 其他节点正在生成统计，跳过")
		return nil
	}
	defer s.rdb.Del(ctx, statsSchedulerLock)

	// 统计前一天的数据
	yesterday := time.Now().AddDate(0, 0, -1)
	statDate := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local)

	log.Printf("[Stats] 开始生成 %s 的每日统计...", statDate.Format("2006-01-02"))

	// 1. 生成订单统计
	if err := s.generateOrderDailyStats(ctx, statDate); err != nil {
		log.Printf("[Stats] 生成订单统计失败: %v", err)
	}

	// 2. 生成财务统计
	if err := s.generateFinanceDailyStats(ctx, statDate); err != nil {
		log.Printf("[Stats] 生成财务统计失败: %v", err)
	}

	// 3. 生成平台统计
	if err := s.generatePlatformDailyStats(ctx, statDate); err != nil {
		log.Printf("[Stats] 生成平台统计失败: %v", err)
	}

	log.Printf("[Stats] %s 每日统计生成完成", statDate.Format("2006-01-02"))
	return nil
}

// generateOrderDailyStats 生成订单每日统计（按店铺）
func (s *StatsService) generateOrderDailyStats(ctx context.Context, statDate time.Time) error {
	startTime := statDate
	endTime := statDate.AddDate(0, 0, 1)

	// 获取所有店铺
	var shops []models.Shop
	if err := s.db.Where("status = 1").Find(&shops).Error; err != nil {
		return err
	}

	for _, shop := range shops {
		orderTable := database.GetOrderTableName(shop.ShopID)
		settlementTable := database.GetOrderSettlementTableName(shop.ShopID)

		var stats struct {
			OrderCount    int64
			TotalAmount   decimal.Decimal
			ShippedCount  int64
			SettledCount  int64
			SettledAmount decimal.Decimal
		}

		// 订单统计
		s.db.Table(orderTable).
			Where("shop_id = ? AND create_time >= ? AND create_time < ?", shop.ShopID, startTime, endTime).
			Select("COUNT(*) as order_count, COALESCE(SUM(total_amount), 0) as total_amount").
			Scan(&stats)

		// 发货统计
		s.db.Table(orderTable).
			Where("shop_id = ? AND create_time >= ? AND create_time < ? AND order_status IN ?", 
				shop.ShopID, startTime, endTime, []string{"SHIPPED", "COMPLETED"}).
			Count(&stats.ShippedCount)

		// 结算统计
		s.db.Table(settlementTable).
			Where("shop_id = ? AND created_at >= ? AND created_at < ? AND status = ?", 
				shop.ShopID, startTime, endTime, models.OrderSettlementCompleted).
			Select("COUNT(*) as settled_count, COALESCE(SUM(total_amount), 0) as settled_amount").
			Scan(&stats)

		// 保存统计
		dailyStat := models.OrderDailyStat{
			StatDate:      statDate,
			ShopID:        shop.ShopID,
			OrderCount:    stats.OrderCount,
			TotalAmount:   stats.TotalAmount,
			ShippedCount:  stats.ShippedCount,
			SettledCount:  stats.SettledCount,
			SettledAmount: stats.SettledAmount,
		}

		// 使用 upsert
		s.db.Where("stat_date = ? AND shop_id = ?", statDate, shop.ShopID).
			Assign(dailyStat).FirstOrCreate(&dailyStat)
	}

	return nil
}

// generateFinanceDailyStats 生成财务每日统计（按店铺）
func (s *StatsService) generateFinanceDailyStats(ctx context.Context, statDate time.Time) error {
	startTime := statDate
	endTime := statDate.AddDate(0, 0, 1)

	var shops []models.Shop
	if err := s.db.Where("status = 1").Find(&shops).Error; err != nil {
		return err
	}

	for _, shop := range shops {
		financeTable := database.GetFinanceIncomeTableName(shop.ShopID)

		var stats struct {
			IncomeCount  int64
			IncomeAmount decimal.Decimal
		}

		s.db.Table(financeTable).
			Where("shop_id = ? AND created_at >= ? AND created_at < ?", shop.ShopID, startTime, endTime).
			Select("COUNT(*) as income_count, COALESCE(SUM(amount), 0) as income_amount").
			Scan(&stats)

		dailyStat := models.FinanceDailyStat{
			StatDate:     statDate,
			ShopID:       shop.ShopID,
			IncomeCount:  stats.IncomeCount,
			IncomeAmount: stats.IncomeAmount,
		}

		s.db.Where("stat_date = ? AND shop_id = ?", statDate, shop.ShopID).
			Assign(dailyStat).FirstOrCreate(&dailyStat)
	}

	return nil
}

// generatePlatformDailyStats 生成平台每日统计（汇总）
func (s *StatsService) generatePlatformDailyStats(ctx context.Context, statDate time.Time) error {
	startTime := statDate
	endTime := statDate.AddDate(0, 0, 1)

	var stats models.PlatformDailyStat
	stats.StatDate = statDate

	// 遍历所有分表统计
	for i := 0; i < database.ShardCount; i++ {
		orderTable := fmt.Sprintf("orders_%d", i)
		settlementTable := fmt.Sprintf("order_settlements_%d", i)
		financeTable := fmt.Sprintf("finance_incomes_%d", i)

		var orderCount int64
		var totalAmount, settledAmount, platformShare decimal.Decimal

		s.db.Table(orderTable).
			Where("create_time >= ? AND create_time < ?", startTime, endTime).
			Count(&orderCount)
		stats.TotalOrders += orderCount

		s.db.Table(orderTable).
			Where("create_time >= ? AND create_time < ?", startTime, endTime).
			Select("COALESCE(SUM(total_amount), 0)").Scan(&totalAmount)
		stats.TotalAmount = stats.TotalAmount.Add(totalAmount)

		s.db.Table(settlementTable).
			Where("created_at >= ? AND created_at < ? AND status = ?", startTime, endTime, models.OrderSettlementCompleted).
			Select("COALESCE(SUM(total_amount), 0)").Scan(&settledAmount)
		stats.SettledAmount = stats.SettledAmount.Add(settledAmount)

		s.db.Table(settlementTable).
			Where("created_at >= ? AND created_at < ? AND status = ?", startTime, endTime, models.OrderSettlementCompleted).
			Select("COALESCE(SUM(platform_share), 0)").Scan(&platformShare)
		stats.PlatformShare = stats.PlatformShare.Add(platformShare)

		var incomeAmount decimal.Decimal
		s.db.Table(financeTable).
			Where("created_at >= ? AND created_at < ?", startTime, endTime).
			Select("COALESCE(SUM(amount), 0)").Scan(&incomeAmount)
		stats.TotalIncome = stats.TotalIncome.Add(incomeAmount)
	}

	// 活跃店铺数
	s.db.Model(&models.Shop{}).Where("status = 1").Count(&stats.ActiveShops)

	// 保存
	s.db.Where("stat_date = ?", statDate).Assign(stats).FirstOrCreate(&stats)

	return nil
}

// GetPlatformStats 获取平台统计（使用汇总表，高性能）
func (s *StatsService) GetPlatformStats(ctx context.Context, startDate, endDate time.Time) (*models.PlatformStatsResult, error) {
	var result models.PlatformStatsResult

	// 从汇总表查询
	s.db.Model(&models.PlatformDailyStat{}).
		Where("stat_date >= ? AND stat_date <= ?", startDate, endDate).
		Select(`
			SUM(total_orders) as total_orders,
			SUM(total_amount) as total_amount,
			SUM(settled_amount) as settled_amount,
			SUM(platform_share) as platform_share,
			SUM(total_income) as total_income
		`).Scan(&result)

	return &result, nil
}

// GetShopStats 获取店铺统计（使用汇总表）
func (s *StatsService) GetShopStats(ctx context.Context, shopID uint64, startDate, endDate time.Time) (*models.ShopStatsResult, error) {
	var result models.ShopStatsResult

	s.db.Model(&models.OrderDailyStat{}).
		Where("shop_id = ? AND stat_date >= ? AND stat_date <= ?", shopID, startDate, endDate).
		Select(`
			SUM(order_count) as total_orders,
			SUM(total_amount) as total_amount,
			SUM(shipped_count) as shipped_count,
			SUM(settled_count) as settled_count,
			SUM(settled_amount) as settled_amount
		`).Scan(&result)

	return &result, nil
}

// GetDailyTrend 获取每日趋势数据
func (s *StatsService) GetDailyTrend(ctx context.Context, days int) ([]models.PlatformDailyStat, error) {
	var stats []models.PlatformDailyStat
	startDate := time.Now().AddDate(0, 0, -days)

	err := s.db.Where("stat_date >= ?", startDate).
		Order("stat_date ASC").
		Find(&stats).Error

	return stats, err
}
