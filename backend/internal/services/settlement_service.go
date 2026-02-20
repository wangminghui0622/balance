package services

import (
	"context"
	"fmt"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
//
// 整个流程在一个事务中完成：
//  1. FOR UPDATE 锁定发货记录 → 防止并发结算
//  2. 检查是否已结算（幂等）
//  3. 创建结算记录
//  4. 执行资金划转（内部各账户操作有自己的事务和行锁）
//  5. 更新结算状态 + 发货记录状态
func (s *SettlementService) SettleOrder(ctx context.Context, shopID uint64, orderSN string, escrowAmount decimal.Decimal) (*models.OrderSettlement, error) {
	shipmentRecordTable := database.GetOrderShipmentRecordTableName(shopID)
	settlementTable := database.GetOrderSettlementTableName(shopID)

	var settlement *models.OrderSettlement

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 获取发货记录（FOR UPDATE 行锁防止并发结算同一订单）
		var shipmentRecord models.OrderShipmentRecord
		if err := tx.Table(shipmentRecordTable).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_sn = ? AND status = ?", orderSN, models.ShipmentRecordStatusShipped).
			First(&shipmentRecord).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("未找到发货记录或订单未发货")
			}
			return err
		}

		// 2. 检查是否已结算（幂等）
		var existingSettlement models.OrderSettlement
		if err := tx.Table(settlementTable).Where("order_sn = ?", orderSN).First(&existingSettlement).Error; err == nil {
			if existingSettlement.Status == models.OrderSettlementCompleted {
				return fmt.Errorf("订单已结算")
			}
		}

		// 3. 获取分成配置
		config, err := s.getProfitShareConfig(shipmentRecord.ShopID, shipmentRecord.OperatorID)
		if err != nil {
			return fmt.Errorf("获取分成配置失败: %w", err)
		}

		// 4. 计算利润和分成
		totalCost := shipmentRecord.TotalCost
		profit := escrowAmount.Sub(totalCost)
		hundred := decimal.NewFromInt(100)
		platformShare := profit.Mul(config.PlatformShareRate).Div(hundred).Round(2)
		operatorShare := profit.Mul(config.OperatorShareRate).Div(hundred).Round(2)
		shopOwnerShare := profit.Sub(platformShare).Sub(operatorShare)
		operatorIncome := totalCost.Add(operatorShare)

		// 5. 创建结算记录
		settlement = &models.OrderSettlement{
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
		if err := tx.Table(settlementTable).Create(settlement).Error; err != nil {
			return fmt.Errorf("创建结算记录失败: %w", err)
		}

		// 6. 执行资金划转（使用 InTx 版本，复用同一事务连接，避免嵌套事务连接池死锁）
		if err := s.executeSettlementInTx(tx, ctx, settlement, &shipmentRecord); err != nil {
			return fmt.Errorf("执行结算失败: %w", err)
		}

		// 7. 更新结算状态为已完成
		now := time.Now()
		if err := tx.Table(settlementTable).Where("id = ?", settlement.ID).Updates(map[string]interface{}{
			"status":     models.OrderSettlementCompleted,
			"settled_at": &now,
		}).Error; err != nil {
			return err
		}
		settlement.Status = models.OrderSettlementCompleted
		settlement.SettledAt = &now

		// 8. 更新发货记录状态为已结算
		return tx.Table(shipmentRecordTable).Where("id = ?", shipmentRecord.ID).Updates(map[string]interface{}{
			"status":        models.ShipmentRecordStatusCompleted,
			"settlement_id": settlement.ID,
		}).Error
	})

	if err != nil {
		return nil, err
	}
	return settlement, nil
}

// executeSettlementInTx 执行结算资金划转（复用调用方事务，避免嵌套独立事务导致连接池死锁）
// 预付款在订单 READY_TO_SHIP 时已冻结，发货不转托管，结算直接按 orders_x/发货记录分账，不操作托管账户
func (s *SettlementService) executeSettlementInTx(outerTx *gorm.DB, ctx context.Context, settlement *models.OrderSettlement, shipmentRecord *models.OrderShipmentRecord) error {
	// 1. 从店铺老板冻结金额中扣除 (结算预付款消耗，与 orders_x 一致)
	_, err := s.accountService.SettlePrepaymentInTx(outerTx, ctx, settlement.ShopOwnerID, shipmentRecord.PrepaymentAmount, settlement.OrderSN,
		fmt.Sprintf("订单结算-成本%s", settlement.TotalCost.String()))
	if err != nil {
		return fmt.Errorf("扣除店铺老板预付款失败: %w", err)
	}

	// 2. 给运营账户增加收入 (成本 + 运营分成)
	_, err = s.accountService.AddOperatorIncomeInTx(outerTx, ctx, settlement.OperatorID, settlement.OperatorIncome, settlement.OrderSN,
		fmt.Sprintf("订单结算-成本%s+分成%s", settlement.TotalCost.String(), settlement.OperatorShare.String()))
	if err != nil {
		return fmt.Errorf("增加运营收入失败: %w", err)
	}

	// 3. 给店主佣金账户增加收入 (店主分成)
	if settlement.ShopOwnerShare.GreaterThan(decimal.Zero) {
		_, err = s.accountService.AddShopOwnerCommissionInTx(outerTx, ctx, settlement.ShopOwnerID, settlement.ShopOwnerShare, settlement.OrderSN,
			fmt.Sprintf("订单结算-利润分成%s%%", settlement.ShopOwnerShareRate.String()))
		if err != nil {
			return fmt.Errorf("增加店主佣金失败: %w", err)
		}
	}

	// 4. 给平台佣金账户增加收入 (平台分成)
	if settlement.PlatformShare.GreaterThan(decimal.Zero) {
		_, err = s.accountService.AddPlatformCommissionInTx(outerTx, ctx, settlement.PlatformShare, settlement.OrderSN,
			fmt.Sprintf("订单结算-平台分成%s%%", settlement.PlatformShareRate.String()))
		if err != nil {
			return fmt.Errorf("增加平台佣金失败: %w", err)
		}
	}

	return nil
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
			// 检查 Shopee 是否已结算 (通过 finance_incomes 分表，仅作触发条件)
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

			// 结算金额以 orders_x.prepayment_amount 为准（订单入系统时已扣的预付款）
			orderTable := database.GetOrderTableName(record.ShopID)
			var order models.Order
			if err := s.db.Table(orderTable).Where("shop_id = ? AND order_sn = ?", record.ShopID, record.OrderSN).
				Select("prepayment_amount").First(&order).Error; err != nil {
				continue
			}
			escrowAmount := order.PrepaymentAmount
			if !escrowAmount.IsPositive() {
				escrowAmount = record.PrepaymentAmount
			}
			if !escrowAmount.IsPositive() {
				escrowAmount = income.Amount
			}

			// 执行结算（按 prepayment_amount 分账）
			_, err = s.SettleOrder(ctx, record.ShopID, record.OrderSN, escrowAmount)
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

// ProcessShopeeAdjustments 处理 Shopee 调账 (定时任务调用) - 遍历所有分表
// 当虾皮发生退款、扣款等调账时，需要反向分账
func (s *SettlementService) ProcessShopeeAdjustments(ctx context.Context) (int, error) {
	adjustedCount := 0

	// 遍历所有分表
	for i := 0; i < database.ShardCount; i++ {
		financeTable := fmt.Sprintf("finance_incomes_%d", i)

		// 查找未处理的调账记录（退款、调账等）
		var incomes []models.FinanceIncome
		err := s.db.Table(financeTable).
			Where("settlement_handle_status = ?", models.SettlementStatusPending).
			Where("order_sn != ''").
			Where("transaction_type IN ?", []string{
				models.TransactionTypeRefund,
				models.TransactionTypeEscrowAdjustment,
				models.TransactionTypeSellerAdjustment,
				models.TransactionTypeCommissionAdjust,
				models.TransactionTypeServiceFeeAdjust,
				models.TransactionTypeShippingFeeAdjust,
			}).
			Find(&incomes).Error
		if err != nil {
			continue
		}

		for _, income := range incomes {
			// 处理调账
			err := s.handleAdjustment(ctx, &income)
			if err != nil {
				fmt.Printf("[Settlement] 处理调账失败 order_sn=%s: %v\n", income.OrderSN, err)
				continue
			}

			// 标记已处理
			s.db.Table(financeTable).Where("id = ?", income.ID).Update("settlement_handle_status", models.SettlementStatusCompleted)
			adjustedCount++
		}
	}

	return adjustedCount, nil
}

// handleAdjustment 处理单个调账记录（写入原结算单的 adj1/adj2/adj3 预留字段，最多3次）
func (s *SettlementService) handleAdjustment(ctx context.Context, income *models.FinanceIncome) error {
	settlementTable := database.GetOrderSettlementTableName(income.ShopID)

	adjustAmount := income.Amount
	if adjustAmount.IsZero() {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查找原结算记录并加锁
		var original models.OrderSettlement
		if err := tx.Table(settlementTable).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_sn = ? AND status = ?", income.OrderSN, models.OrderSettlementCompleted).
			First(&original).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil // 无原结算则跳过
			}
			return err
		}

		if original.AdjustmentCount >= 3 {
			return fmt.Errorf("该订单调账已达3次上限")
		}

		hundred := decimal.NewFromInt(100)
		platformAdjust := adjustAmount.Mul(original.PlatformShareRate).Div(hundred).Round(2)
		operatorAdjust := adjustAmount.Mul(original.OperatorShareRate).Div(hundred).Round(2)
		shopOwnerAdjust := adjustAmount.Sub(platformAdjust).Sub(operatorAdjust)
		adjRemark := fmt.Sprintf("虾皮调账: %s", income.TransactionType)

		// 2. 执行调账资金划转
		adjSettlement := &models.OrderSettlement{
			OrderSN:     income.OrderSN,
			ShopOwnerID: original.ShopOwnerID,
			OperatorID:  original.OperatorID,
			EscrowAmount: adjustAmount,
			PlatformShare: platformAdjust,
			OperatorShare: operatorAdjust,
			ShopOwnerShare: shopOwnerAdjust,
		}
		if err := s.executeAdjustmentInTx(tx, ctx, adjSettlement); err != nil {
			return fmt.Errorf("执行调账: %w", err)
		}

		// 3. 写入对应的 adjN 预留字段
		now := time.Now()
		slot := original.AdjustmentCount + 1
		updates := map[string]interface{}{
			"adjustment_count": slot,
		}
		switch slot {
		case 1:
			updates["adj1_amount"] = adjustAmount
			updates["adj1_platform_share"] = platformAdjust
			updates["adj1_operator_share"] = operatorAdjust
			updates["adj1_shop_owner_share"] = shopOwnerAdjust
			updates["adj1_at"] = now
			updates["adj1_remark"] = adjRemark
		case 2:
			updates["adj2_amount"] = adjustAmount
			updates["adj2_platform_share"] = platformAdjust
			updates["adj2_operator_share"] = operatorAdjust
			updates["adj2_shop_owner_share"] = shopOwnerAdjust
			updates["adj2_at"] = now
			updates["adj2_remark"] = adjRemark
		case 3:
			updates["adj3_amount"] = adjustAmount
			updates["adj3_platform_share"] = platformAdjust
			updates["adj3_operator_share"] = operatorAdjust
			updates["adj3_shop_owner_share"] = shopOwnerAdjust
			updates["adj3_at"] = now
			updates["adj3_remark"] = adjRemark
		}

		return tx.Table(settlementTable).Where("id = ?", original.ID).Updates(updates).Error
	})
}

// executeAdjustmentInTx 执行调账资金划转（复用调用方事务）
// 当调账为扣款（负数）时，需将扣款金额返还给店铺老板预付款账户
func (s *SettlementService) executeAdjustmentInTx(outerTx *gorm.DB, ctx context.Context, settlement *models.OrderSettlement) error {
	// 0. 调账扣款时：返还金额到店铺老板预付款（虾皮多扣了，需退还给店主）
	if settlement.EscrowAmount.LessThan(decimal.Zero) {
		refundAmount := settlement.EscrowAmount.Abs()
		if refundAmount.GreaterThan(decimal.Zero) {
			_, err := s.accountService.RefundPrepaymentAdjustmentInTx(outerTx, ctx, settlement.ShopOwnerID, refundAmount, settlement.OrderSN,
				fmt.Sprintf("虾皮调账返还-退%s至预付款", refundAmount.String()))
			if err != nil {
				return fmt.Errorf("返还预付款失败: %w", err)
			}
		}
	}

	// 1. 调整运营账户
	if !settlement.OperatorShare.IsZero() {
		if settlement.OperatorShare.GreaterThan(decimal.Zero) {
			_, err := s.accountService.AddOperatorIncomeInTx(outerTx, ctx, settlement.OperatorID, settlement.OperatorShare, settlement.OrderSN,
				fmt.Sprintf("虾皮调账补款-运营分成%s", settlement.OperatorShare.String()))
			if err != nil {
				return fmt.Errorf("增加运营调账收入失败: %w", err)
			}
		} else {
			_, err := s.accountService.DeductOperatorBalanceInTx(outerTx, ctx, settlement.OperatorID, settlement.OperatorShare.Abs(), settlement.OrderSN,
				fmt.Sprintf("虾皮调账扣款-运营分成%s", settlement.OperatorShare.Abs().String()))
			if err != nil {
				return fmt.Errorf("扣除运营调账金额失败: %w", err)
			}
		}
	}

	// 2. 调整店主佣金账户
	if !settlement.ShopOwnerShare.IsZero() {
		if settlement.ShopOwnerShare.GreaterThan(decimal.Zero) {
			_, err := s.accountService.AddShopOwnerCommissionInTx(outerTx, ctx, settlement.ShopOwnerID, settlement.ShopOwnerShare, settlement.OrderSN,
				fmt.Sprintf("虾皮调账补款-店主分成%s", settlement.ShopOwnerShare.String()))
			if err != nil {
				return fmt.Errorf("增加店主调账佣金失败: %w", err)
			}
		} else {
			_, err := s.accountService.DeductShopOwnerCommissionInTx(outerTx, ctx, settlement.ShopOwnerID, settlement.ShopOwnerShare.Abs(), settlement.OrderSN,
				fmt.Sprintf("虾皮调账扣款-店主分成%s", settlement.ShopOwnerShare.Abs().String()))
			if err != nil {
				return fmt.Errorf("扣除店主调账佣金失败: %w", err)
			}
		}
	}

	// 3. 调整平台佣金账户
	if !settlement.PlatformShare.IsZero() {
		if settlement.PlatformShare.GreaterThan(decimal.Zero) {
			_, err := s.accountService.AddPlatformCommissionInTx(outerTx, ctx, settlement.PlatformShare, settlement.OrderSN,
				fmt.Sprintf("虾皮调账补款-平台分成%s", settlement.PlatformShare.String()))
			if err != nil {
				return fmt.Errorf("增加平台调账佣金失败: %w", err)
			}
		} else {
			_, err := s.accountService.DeductPlatformCommissionInTx(outerTx, ctx, settlement.PlatformShare.Abs(), settlement.OrderSN,
				fmt.Sprintf("虾皮调账扣款-平台分成%s", settlement.PlatformShare.Abs().String()))
			if err != nil {
				return fmt.Errorf("扣除平台调账佣金失败: %w", err)
			}
		}
	}

	return nil
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
