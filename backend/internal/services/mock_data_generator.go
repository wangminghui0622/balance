package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MockDataGenerator 模拟数据生成器（仅用于沙箱环境测试）
type MockDataGenerator struct {
	db          *gorm.DB
	shardedDB   *database.ShardedDB
	idGenerator *utils.IDGenerator
}

// NewMockDataGenerator 创建模拟数据生成器
func NewMockDataGenerator() *MockDataGenerator {
	db := database.GetDB()
	return &MockDataGenerator{
		db:          db,
		shardedDB:   database.NewShardedDB(db),
		idGenerator: utils.NewIDGenerator(database.GetRedis()),
	}
}

// MockOrderRequest 模拟订单请求
type MockOrderRequest struct {
	ShopID     uint64 `json:"shop_id"`
	Count      int    `json:"count"`       // 生成订单数量
	MinAmount  int    `json:"min_amount"`  // 最小金额（分）
	MaxAmount  int    `json:"max_amount"`  // 最大金额（分）
	WithSettle bool   `json:"with_settle"` // 是否生成结算数据
	WithEscrow bool   `json:"with_escrow"` // 是否生成托管数据
	WithIncome bool   `json:"with_income"` // 是否生成收入数据
}

// MockOrderResult 模拟订单结果
type MockOrderResult struct {
	OrderCount  int      `json:"order_count"`
	SettleCount int      `json:"settle_count"`
	EscrowCount int      `json:"escrow_count"`
	IncomeCount int      `json:"income_count"`
	OrderSNs    []string `json:"order_sns"`
	TotalAmount string   `json:"total_amount"`
}

// GenerateMockOrders 生成模拟订单数据
func (g *MockDataGenerator) GenerateMockOrders(ctx context.Context, req *MockOrderRequest) (*MockOrderResult, error) {
	if req.Count <= 0 || req.Count > 100 {
		req.Count = 10
	}
	if req.MinAmount <= 0 {
		req.MinAmount = 1000 // 10元
	}
	if req.MaxAmount <= 0 || req.MaxAmount < req.MinAmount {
		req.MaxAmount = 100000 // 1000元
	}

	result := &MockOrderResult{
		OrderSNs: make([]string, 0, req.Count),
	}
	totalAmount := decimal.Zero

	// 获取店铺信息
	var shop models.Shop
	if err := g.db.Where("shop_id = ?", req.ShopID).First(&shop).Error; err != nil {
		return nil, fmt.Errorf("店铺不存在: %w", err)
	}

	// 获取分润配置
	var profitConfig models.ProfitShareConfig
	g.db.Where("shop_id = ?", req.ShopID).First(&profitConfig)
	if profitConfig.ID == 0 {
		// 使用默认分润配置
		profitConfig.PlatformShareRate = decimal.NewFromFloat(5.00)   // 5%
		profitConfig.ShopOwnerShareRate = decimal.NewFromFloat(50.00) // 50%
		profitConfig.OperatorShareRate = decimal.NewFromFloat(45.00)  // 45%
	}

	now := time.Now()
	orderTable := database.GetOrderTableName(req.ShopID)
	itemTable := database.GetOrderItemTableName(req.ShopID)

	for i := 0; i < req.Count; i++ {
		// 生成订单号
		orderSN := fmt.Sprintf("MOCK%d%d%04d", req.ShopID, now.Unix(), i)

		// 随机金额
		amount := req.MinAmount + rand.Intn(req.MaxAmount-req.MinAmount)
		orderAmount := decimal.NewFromInt(int64(amount)).Div(decimal.NewFromInt(100))
		totalAmount = totalAmount.Add(orderAmount)

		// 创建订单
		createTime := now.Add(-time.Duration(rand.Intn(7*24)) * time.Hour)
		order := models.Order{
			ShopID:        req.ShopID,
			OrderSN:       orderSN,
			OrderStatus:   consts.OrderStatusCompleted,
			Region:        shop.Region,
			Currency:      "SGD",
			TotalAmount:   orderAmount,
			BuyerUsername: fmt.Sprintf("test_buyer_%d", i),
			PayTime:       &now,
			CreateTime:    &createTime,
		}

		orderID, _ := g.idGenerator.GenerateOrderID(ctx)
		order.ID = uint64(orderID)
		if err := g.db.Table(orderTable).Create(&order).Error; err != nil {
			log.Printf("[MockData] 创建订单失败: %v", err)
			continue
		}

		// 创建订单商品
		itemCount := 1 + rand.Intn(3)
		itemAmount := orderAmount.Div(decimal.NewFromInt(int64(itemCount)))
		for j := 0; j < itemCount; j++ {
			itemID, _ := g.idGenerator.GenerateOrderItemID(ctx)
			item := models.OrderItem{
				ID:        uint64(itemID),
				OrderID:   order.ID,
				ShopID:    req.ShopID,
				OrderSN:   orderSN,
				ItemID:    uint64(1000000 + rand.Intn(999999)),
				ItemName:  fmt.Sprintf("测试商品_%d_%d", i, j),
				ItemSKU:   fmt.Sprintf("SKU%d%d", i, j),
				ModelID:   uint64(rand.Intn(10000)),
				ModelName: "默认规格",
				ModelSKU:  fmt.Sprintf("MSKU%d%d", i, j),
				Quantity:  1,
				ItemPrice: itemAmount,
			}
			g.db.Table(itemTable).Create(&item)
		}

		result.OrderSNs = append(result.OrderSNs, orderSN)
		result.OrderCount++

		// 生成结算数据
		if req.WithSettle {
			if err := g.generateMockSettlement(ctx, req.ShopID, order.ID, orderSN, orderAmount, &profitConfig); err != nil {
				log.Printf("[MockData] 生成结算数据失败: %v", err)
			} else {
				result.SettleCount++
			}
		}

		// 生成托管数据
		if req.WithEscrow {
			if err := g.generateMockEscrow(ctx, req.ShopID, orderSN, orderAmount); err != nil {
				log.Printf("[MockData] 生成托管数据失败: %v", err)
			} else {
				result.EscrowCount++
			}
		}

		// 生成收入数据
		if req.WithIncome {
			if err := g.generateMockIncome(ctx, req.ShopID, orderSN, orderAmount); err != nil {
				log.Printf("[MockData] 生成收入数据失败: %v", err)
			} else {
				result.IncomeCount++
			}
		}
	}

	result.TotalAmount = totalAmount.StringFixed(2)
	log.Printf("[MockData] 生成模拟数据完成: 订单=%d, 结算=%d, 托管=%d, 收入=%d",
		result.OrderCount, result.SettleCount, result.EscrowCount, result.IncomeCount)

	return result, nil
}

// generateMockSettlement 生成模拟结算数据
func (g *MockDataGenerator) generateMockSettlement(ctx context.Context, shopID uint64, orderID uint64, orderSN string, amount decimal.Decimal, config *models.ProfitShareConfig) error {
	now := time.Now()
	settlementTable := database.GetOrderSettlementTableName(shopID)

	// 计算利润和分成（配置中是百分比）
	escrowAmount := amount.Mul(decimal.NewFromFloat(0.98)) // 扣除2%手续费
	profit := escrowAmount                                  // 简化：利润=结算金额

	platformShare := profit.Mul(config.PlatformShareRate).Div(decimal.NewFromInt(100))
	ownerShare := profit.Mul(config.ShopOwnerShareRate).Div(decimal.NewFromInt(100))
	operatorShare := profit.Mul(config.OperatorShareRate).Div(decimal.NewFromInt(100))

	settlement := models.OrderSettlement{
		SettlementNo:       fmt.Sprintf("SET%d%d", shopID, now.UnixNano()),
		ShopID:             shopID,
		OrderSN:            orderSN,
		OrderID:            orderID,
		ShopOwnerID:        0,
		OperatorID:         0,
		Currency:           "SGD",
		EscrowAmount:       escrowAmount,
		GoodsCost:          decimal.Zero,
		ShippingCost:       decimal.Zero,
		TotalCost:          decimal.Zero,
		Profit:             profit,
		PlatformShareRate:  config.PlatformShareRate,
		OperatorShareRate:  config.OperatorShareRate,
		ShopOwnerShareRate: config.ShopOwnerShareRate,
		PlatformShare:      platformShare,
		OperatorShare:      operatorShare,
		ShopOwnerShare:     ownerShare,
		OperatorIncome:     operatorShare,
		Status:             models.OrderSettlementCompleted,
		SettledAt:          &now,
	}

	settlementID, _ := g.idGenerator.GenerateOrderSettlementID(ctx)
	settlement.ID = uint64(settlementID)
	return g.db.Table(settlementTable).Create(&settlement).Error
}

// generateMockEscrow 生成模拟托管数据
func (g *MockDataGenerator) generateMockEscrow(ctx context.Context, shopID uint64, orderSN string, amount decimal.Decimal) error {
	escrowTable := database.GetOrderEscrowTableName(shopID)

	escrowAmount := amount.Mul(decimal.NewFromFloat(0.98)) // 扣除2%手续费

	escrow := models.OrderEscrow{
		ShopID:               shopID,
		OrderSN:              orderSN,
		Currency:             "SGD",
		BuyerTotalAmount:     amount,
		OriginalPrice:        amount,
		EscrowAmount:         escrowAmount,
		BuyerPaidShippingFee: decimal.NewFromFloat(5.00),
		FinalShippingFee:     decimal.NewFromFloat(5.00),
		ActualShippingFee:    decimal.NewFromFloat(5.00),
		EstimatedShippingFee: decimal.NewFromFloat(5.00),
		ServiceFee:           amount.Mul(decimal.NewFromFloat(0.02)),
		SyncStatus:           1,
	}

	escrowID, _ := g.idGenerator.GenerateOrderEscrowID(ctx)
	escrow.ID = uint64(escrowID)
	return g.db.Table(escrowTable).Create(&escrow).Error
}

// generateMockIncome 生成模拟收入数据
func (g *MockDataGenerator) generateMockIncome(ctx context.Context, shopID uint64, orderSN string, amount decimal.Decimal) error {
	incomeTable := database.GetFinanceIncomeTableName(shopID)

	income := models.FinanceIncome{
		ShopID:          shopID,
		TransactionID:   time.Now().UnixNano(),
		OrderSN:         orderSN,
		Amount:          amount.Mul(decimal.NewFromFloat(0.98)),
		TransactionType: models.TransactionTypeEscrowVerifiedAdd,
		Status:          "COMPLETED",
		Description:     "模拟订单收入",
		TransactionTime: time.Now().Unix(),
	}

	incomeID, _ := g.idGenerator.GenerateFinanceIncomeID(ctx)
	income.ID = uint64(incomeID)
	return g.db.Table(incomeTable).Create(&income).Error
}

// CleanMockData 清理模拟数据
func (g *MockDataGenerator) CleanMockData(ctx context.Context, shopID uint64) (int64, error) {
	var totalDeleted int64

	orderTable := database.GetOrderTableName(shopID)
	itemTable := database.GetOrderItemTableName(shopID)
	settlementTable := database.GetOrderSettlementTableName(shopID)
	escrowTable := database.GetOrderEscrowTableName(shopID)
	incomeTable := database.GetFinanceIncomeTableName(shopID)

	// 删除模拟订单相关数据（以MOCK开头的订单号）
	result := g.db.Table(itemTable).Where("shop_id = ? AND order_sn LIKE 'MOCK%'", shopID).Delete(&models.OrderItem{})
	totalDeleted += result.RowsAffected

	result = g.db.Table(settlementTable).Where("shop_id = ? AND order_sn LIKE 'MOCK%'", shopID).Delete(&models.OrderSettlement{})
	totalDeleted += result.RowsAffected

	result = g.db.Table(escrowTable).Where("shop_id = ? AND order_sn LIKE 'MOCK%'", shopID).Delete(&models.OrderEscrow{})
	totalDeleted += result.RowsAffected

	result = g.db.Table(incomeTable).Where("shop_id = ? AND order_sn LIKE 'MOCK%'", shopID).Delete(&models.FinanceIncome{})
	totalDeleted += result.RowsAffected

	result = g.db.Table(orderTable).Where("shop_id = ? AND order_sn LIKE 'MOCK%'", shopID).Delete(&models.Order{})
	totalDeleted += result.RowsAffected

	log.Printf("[MockData] 清理模拟数据完成: 店铺=%d, 删除记录=%d", shopID, totalDeleted)

	return totalDeleted, nil
}
