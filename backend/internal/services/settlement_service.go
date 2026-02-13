package services

import (
	"context"
	"fmt"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SettlementService 结算服务
type SettlementService struct {
	db             *gorm.DB
	accountService *AccountService
	shardedDB      *database.ShardedDB
}

// NewSettlementService 创建结算服务
func NewSettlementService() *SettlementService {
	db := database.GetDB()
	return &SettlementService{
		db:             db,
		accountService: NewAccountService(),
		shardedDB:      database.NewShardedDB(db),
	}
}

// GenerateSettlementNo 生成结算单号
func (s *SettlementService) GenerateSettlementNo() string {
	return fmt.Sprintf("ST%d%d", time.Now().UnixNano(), time.Now().UnixMicro()%1000)
}

// SettleOrder 结算订单 (Shopee 结算后调用) - 需要shopID来定位分表
func (s *SettlementService) SettleOrder(ctx context.Context, shopID uint64, orderSN string, escrowAmount decimal.Decimal) (*models.OrderSettlement, error) {
	shipmentRecordTable := database.GetOrderShipmentRecordTableName(shopID)
	settlementTable := database.GetOrderSettlementTableName(shopID)

	// 1. 获取发货记录
	var shipmentRecord models.OrderShipmentRecord
	if err := s.db.Table(shipmentRecordTable).Where("order_sn = ? AND status = ?", orderSN, models.ShipmentRecordStatusShipped).First(&shipmentRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到发货记录或订单未发货")
		}
		return nil, err
	}

	// 2. 检查是否已结算
	var existingSettlement models.OrderSettlement
	if err := s.db.Table(settlementTable).Where("order_sn = ?", orderSN).First(&existingSettlement).Error; err == nil {
		if existingSettlement.Status == models.OrderSettlementCompleted {
			return nil, fmt.Errorf("订单已结算")
		}
	}

	// 3. 获取分成配置
	config, err := s.getProfitShareConfig(shipmentRecord.ShopID, shipmentRecord.OperatorID)
	if err != nil {
		return nil, fmt.Errorf("获取分成配置失败: %w", err)
	}

	// 4. 计算利润和分成
	totalCost := shipmentRecord.TotalCost
	profit := escrowAmount.Sub(totalCost)

	// 分成计算
	hundred := decimal.NewFromInt(100)
	platformShare := profit.Mul(config.PlatformShareRate).Div(hundred).Round(2)
	operatorShare := profit.Mul(config.OperatorShareRate).Div(hundred).Round(2)
	shopOwnerShare := profit.Sub(platformShare).Sub(operatorShare) // 剩余给店主，避免精度问题

	// 运营实际收入 = 成本 + 运营分成
	operatorIncome := totalCost.Add(operatorShare)

	// 5. 创建结算记录
	settlement := &models.OrderSettlement{
		SettlementNo:       s.GenerateSettlementNo(),
		ShopID:             shipmentRecord.ShopID,
		OrderSN:            orderSN,
		OrderID:            shipmentRecord.OrderID,
		ShopOwnerID:        shipmentRecord.ShopOwnerID,
		OperatorID:         shipmentRecord.OperatorID,
		Currency:           shipmentRecord.Currency,
		EscrowAmount:       escrowAmount,
		GoodsCost:          shipmentRecord.GoodsCost,
		ShippingCost:       shipmentRecord.ShippingCost,
		TotalCost:          totalCost,
		Profit:             profit,
		PlatformShareRate:  config.PlatformShareRate,
		OperatorShareRate:  config.OperatorShareRate,
		ShopOwnerShareRate: config.ShopOwnerShareRate,
		PlatformShare:      platformShare,
		OperatorShare:      operatorShare,
		ShopOwnerShare:     shopOwnerShare,
		OperatorIncome:     operatorIncome,
		Status:             models.OrderSettlementPending,
	}

	if err := s.db.Table(settlementTable).Create(settlement).Error; err != nil {
		return nil, fmt.Errorf("创建结算记录失败: %w", err)
	}

	// 6. 执行资金划转
	err = s.executeSettlement(ctx, settlement, &shipmentRecord)
	if err != nil {
		settlement.Remark = fmt.Sprintf("结算失败: %s", err.Error())
		s.db.Save(settlement)
		return nil, fmt.Errorf("执行结算失败: %w", err)
	}

	// 7. 更新状态
	now := time.Now()
	settlement.Status = models.OrderSettlementCompleted
	settlement.SettledAt = &now
	s.db.Table(settlementTable).Where("id = ?", settlement.ID).Updates(map[string]interface{}{
		"status":     settlement.Status,
		"settled_at": settlement.SettledAt,
	})

	// 8. 更新发货记录状态
	shipmentRecord.Status = models.ShipmentRecordStatusCompleted
	shipmentRecord.SettlementID = settlement.ID
	s.db.Table(shipmentRecordTable).Where("id = ?", shipmentRecord.ID).Updates(map[string]interface{}{
		"status":        shipmentRecord.Status,
		"settlement_id": shipmentRecord.SettlementID,
	})

	return settlement, nil
}

// executeSettlement 执行结算资金划转
func (s *SettlementService) executeSettlement(ctx context.Context, settlement *models.OrderSettlement, shipmentRecord *models.OrderShipmentRecord) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 从店铺老板冻结金额中扣除 (结算给运营)
		// 冻结金额 = 成本，结算时从冻结金额扣除
		_, err := s.accountService.SettlePrepayment(ctx, settlement.ShopOwnerID, shipmentRecord.FrozenAmount, settlement.OrderSN,
			fmt.Sprintf("订单结算-成本%s", settlement.TotalCost.String()))
		if err != nil {
			return fmt.Errorf("扣除店铺老板预付款失败: %w", err)
		}

		// 2. 从托管账户转出 (发货时已转入托管)
		err = s.accountService.TransferFromEscrow(ctx, shipmentRecord.FrozenAmount, settlement.OrderSN, "订单结算转出")
		if err != nil {
			return fmt.Errorf("托管账户转出失败: %w", err)
		}

		// 3. 给运营账户增加收入 (成本 + 运营分成)
		_, err = s.accountService.AddOperatorIncome(ctx, settlement.OperatorID, settlement.OperatorIncome, settlement.OrderSN,
			fmt.Sprintf("订单结算-成本%s+分成%s", settlement.TotalCost.String(), settlement.OperatorShare.String()))
		if err != nil {
			return fmt.Errorf("增加运营收入失败: %w", err)
		}

		// 4. 给店主佣金账户增加收入 (店主分成)
		if settlement.ShopOwnerShare.GreaterThan(decimal.Zero) {
			_, err = s.accountService.AddShopOwnerCommission(ctx, settlement.ShopOwnerID, settlement.ShopOwnerShare, settlement.OrderSN,
				fmt.Sprintf("订单结算-利润分成%s%%", settlement.ShopOwnerShareRate.String()))
			if err != nil {
				return fmt.Errorf("增加店主佣金失败: %w", err)
			}
		}

		// 5. 给平台佣金账户增加收入 (平台分成)
		if settlement.PlatformShare.GreaterThan(decimal.Zero) {
			_, err = s.accountService.AddPlatformCommission(ctx, settlement.PlatformShare, settlement.OrderSN,
				fmt.Sprintf("订单结算-平台分成%s%%", settlement.PlatformShareRate.String()))
			if err != nil {
				return fmt.Errorf("增加平台佣金失败: %w", err)
			}
		}

		return nil
	})
}

// getProfitShareConfig 获取分成配置
func (s *SettlementService) getProfitShareConfig(shopID uint64, operatorID int64) (*models.ProfitShareConfig, error) {
	var config models.ProfitShareConfig
	err := s.db.Where("shop_id = ? AND operator_id = ? AND status = 1", shopID, operatorID).
		Where("effective_from <= ? AND (effective_to IS NULL OR effective_to > ?)", time.Now(), time.Now()).
		First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// 返回默认配置
		return &models.ProfitShareConfig{
			PlatformShareRate:  decimal.NewFromFloat(5),   // 平台 5%
			OperatorShareRate:  decimal.NewFromFloat(45),  // 运营 45%
			ShopOwnerShareRate: decimal.NewFromFloat(50),  // 店主 50%
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateProfitShareConfig 创建分成配置
func (s *SettlementService) CreateProfitShareConfig(ctx context.Context, config *models.ProfitShareConfig) error {
	// 验证比例总和为 100%
	total := config.PlatformShareRate.Add(config.OperatorShareRate).Add(config.ShopOwnerShareRate)
	if !total.Equal(decimal.NewFromInt(100)) {
		return fmt.Errorf("分成比例总和必须为100%%，当前为%s%%", total.String())
	}

	config.Status = 1
	config.EffectiveFrom = time.Now()
	return s.db.Create(config).Error
}

// ProcessShopeeSettlement 处理 Shopee 结算 (定时任务调用) - 遍历所有分表
func (s *SettlementService) ProcessShopeeSettlement(ctx context.Context) (int, error) {
	settledCount := 0

	// 遍历所有分表
	for i := 0; i < database.ShardCount; i++ {
		shipmentRecordTable := fmt.Sprintf("order_shipment_records_%d", i)

		// 查找已发货但未结算的订单
		var records []models.OrderShipmentRecord
		err := s.db.Table(shipmentRecordTable).Where("status = ?", models.ShipmentRecordStatusShipped).Find(&records).Error
		if err != nil {
			continue
		}

		for _, record := range records {
			// 检查 Shopee 是否已结算 (通过 finance_incomes 分表)
			financeTable := database.GetFinanceIncomeTableName(record.ShopID)
			var income models.FinanceIncome
			err := s.db.Table(financeTable).Where("order_sn = ? AND transaction_type = ?", record.OrderSN, models.TransactionTypeEscrowVerifiedAdd).First(&income).Error
			if err != nil {
				continue // Shopee 还未结算
			}

			// 检查是否已处理
			if income.SettlementHandleStatus == models.SettlementStatusCompleted {
				continue
			}

			// 执行结算
			_, err = s.SettleOrder(ctx, record.ShopID, record.OrderSN, income.Amount)
			if err != nil {
				continue
			}

			// 标记已处理
			s.db.Table(financeTable).Where("id = ?", income.ID).Update("settlement_handle_status", models.SettlementStatusCompleted)

			settledCount++
		}
	}

	return settledCount, nil
}

// GetPendingSettlements 获取待结算订单 - 遍历所有分表
func (s *SettlementService) GetPendingSettlements(ctx context.Context, page, pageSize int) ([]models.OrderShipmentRecord, int64, error) {
	var allRecords []models.OrderShipmentRecord
	var total int64

	// 遍历所有分表统计和获取数据
	for i := 0; i < database.ShardCount; i++ {
		shipmentRecordTable := fmt.Sprintf("order_shipment_records_%d", i)

		var count int64
		s.db.Table(shipmentRecordTable).Where("status = ?", models.ShipmentRecordStatusShipped).Count(&count)
		total += count

		var records []models.OrderShipmentRecord
		s.db.Table(shipmentRecordTable).Where("status = ?", models.ShipmentRecordStatusShipped).Order("created_at DESC").Find(&records)
		allRecords = append(allRecords, records...)
	}

	// 内存分页
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset >= len(allRecords) {
		return []models.OrderShipmentRecord{}, total, nil
	}
	if end > len(allRecords) {
		end = len(allRecords)
	}

	return allRecords[offset:end], total, nil
}

// GetSettlements 获取结算记录 - 遍历所有分表
func (s *SettlementService) GetSettlements(ctx context.Context, adminID int64, role string, page, pageSize int) ([]models.OrderSettlement, int64, error) {
	var allSettlements []models.OrderSettlement
	var total int64

	// 遍历所有分表
	for i := 0; i < database.ShardCount; i++ {
		settlementTable := fmt.Sprintf("order_settlements_%d", i)

		query := s.db.Table(settlementTable)
		switch role {
		case "shop_owner":
			query = query.Where("shop_owner_id = ?", adminID)
		case "operator":
			query = query.Where("operator_id = ?", adminID)
		}

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
		return []models.OrderSettlement{}, total, nil
	}
	if end > len(allSettlements) {
		end = len(allSettlements)
	}

	return allSettlements[offset:end], total, nil
}

// GetSettlementStats 获取结算统计 - 遍历所有分表
func (s *SettlementService) GetSettlementStats(ctx context.Context, adminID int64, role string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalSettled decimal.Decimal
	var totalPending int64
	var totalProfit decimal.Decimal

	// 遍历所有分表统计
	for i := 0; i < database.ShardCount; i++ {
		settlementTable := fmt.Sprintf("order_settlements_%d", i)
		shipmentRecordTable := fmt.Sprintf("order_shipment_records_%d", i)

		var settled, profit decimal.Decimal
		query := s.db.Table(settlementTable)

		switch role {
		case "shop_owner":
			query.Where("shop_owner_id = ? AND status = ?", adminID, models.SettlementStatusCompleted).Select("COALESCE(SUM(shop_owner_share), 0)").Scan(&settled)
			query.Where("shop_owner_id = ? AND status = ?", adminID, models.SettlementStatusCompleted).Select("COALESCE(SUM(profit), 0)").Scan(&profit)
		case "operator":
			query.Where("operator_id = ? AND status = ?", adminID, models.SettlementStatusCompleted).Select("COALESCE(SUM(operator_income), 0)").Scan(&settled)
			query.Where("operator_id = ? AND status = ?", adminID, models.SettlementStatusCompleted).Select("COALESCE(SUM(operator_share), 0)").Scan(&profit)
		}

		totalSettled = totalSettled.Add(settled)
		totalProfit = totalProfit.Add(profit)

		var pending int64
		s.db.Table(shipmentRecordTable).Where("status = ?", models.ShipmentRecordStatusShipped).Count(&pending)
		totalPending += pending
	}

	stats["total_settled"] = totalSettled
	stats["total_pending"] = totalPending
	stats["total_profit"] = totalProfit

	return stats, nil
}
