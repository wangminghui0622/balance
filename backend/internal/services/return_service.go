package services

import (
	"context"
	"fmt"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/shopee"
	"balance/backend/internal/utils"

	"github.com/go-redsync/redsync/v4"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReturnService 退货退款服务
// 负责：
//  1. 从 Shopee 拉取退货详情并保存到 returns_X 分表
//  2. 当退货状态确认退款后（ACCEPTED / REFUND_PAID），自动解冻店铺老板的预付款
//  3. 巡检时批量同步近期退货记录
type ReturnService struct {
	db             *gorm.DB
	rs             *redsync.Redsync
	idGenerator    *utils.IDGenerator
	accountService *AccountService
}

// NewReturnService 创建退货退款服务
func NewReturnService() *ReturnService {
	return &ReturnService{
		db:             database.GetDB(),
		rs:             database.GetRedsync(),
		idGenerator:    utils.NewIDGenerator(database.GetRedis()),
		accountService: NewAccountService(),
	}
}

// UpsertReturn 创建或更新退货退款记录（幂等），并在退款确认时自动解冻预付款
//
// 流程：
//  1. 从 Shopee API 响应构建 Return 对象
//  2. 在事务中：查找已有记录 → 不存在则创建，存在则更新状态/金额
//  3. 如果退货状态变为已确认退款（ACCEPTED/REFUND_PAID），处理预付款解冻
//
// 并发安全：
//   - 使用分布式锁保证同一退货单不会被并发处理
//   - DB 事务内 SELECT ... FOR UPDATE 行锁防止重复写入
func (s *ReturnService) UpsertReturn(ctx context.Context, shopID uint64, detail *shopee.ReturnDetailResponse) error {
	resp := &detail.Response
	if resp.ReturnSN == "" {
		return nil
	}

	// 获取退货级分布式锁
	lockKey := fmt.Sprintf(consts.KeyReturnLock, shopID, resp.ReturnSN)
	mutex := s.rs.NewMutex(lockKey,
		redsync.WithExpiry(consts.ReturnLockExpire),
		redsync.WithTries(3),
		redsync.WithRetryDelay(200*time.Millisecond),
	)
	unlockFunc, acquired := utils.TryLockWithAutoExtend(ctx, mutex, consts.ReturnLockExpire/3)
	if !acquired {
		return nil // 有其他协程在处理，放弃
	}
	defer unlockFunc()

	returnTable := database.GetReturnTableName(shopID)
	refundAmount := decimal.NewFromFloat(resp.RefundAmount)
	amountBeforeDisc := decimal.NewFromFloat(resp.AmountBeforeDisc)

	var shopeeCreateTime, shopeeUpdateTime, dueDate *time.Time
	if resp.CreateTime > 0 {
		t := time.Unix(resp.CreateTime, 0)
		shopeeCreateTime = &t
	}
	if resp.UpdateTime > 0 {
		t := time.Unix(resp.UpdateTime, 0)
		shopeeUpdateTime = &t
	}
	if resp.DueDate > 0 {
		t := time.Unix(resp.DueDate, 0)
		dueDate = &t
	}

	var needProcessRefund bool

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 查找已有记录（行锁防止并发写入）
		var existing models.Return
		err := tx.Table(returnTable).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("shop_id = ? AND return_sn = ?", shopID, resp.ReturnSN).
			First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// 新记录：创建
			returnID, idErr := s.idGenerator.GenerateReturnID(ctx)
			if idErr != nil {
				return fmt.Errorf("生成退货ID失败: %w", idErr)
			}

			// 如果退款已确认，直接标记为 PROCESSING（防止另一台机器重复处理）
			r := &models.Return{Status: resp.Status}
			refundStatus := int8(models.ReturnRefundUnprocessed)
			if r.IsRefundConfirmed() {
				refundStatus = models.ReturnRefundProcessing
				needProcessRefund = true
			}

			newReturn := models.Return{
				ID:               uint64(returnID),
				ShopID:           shopID,
				ReturnSN:         resp.ReturnSN,
				OrderSN:          resp.OrderSN,
				Reason:           resp.Reason,
				TextReason:       resp.TextReason,
				RefundAmount:     refundAmount,
				AmountBeforeDisc: amountBeforeDisc,
				Currency:         resp.Currency,
				Status:           resp.Status,
				NeedsLogistics:   resp.NeedsLogistics,
				TrackingNumber:   resp.TrackingNumber,
				LogisticsStatus:  resp.LogisticsStatus,
				BuyerUsername:     resp.User.Username,
				ShopeeCreateTime: shopeeCreateTime,
				ShopeeUpdateTime: shopeeUpdateTime,
				DueDate:          dueDate,
				RefundStatus:     refundStatus,
			}

			if err := tx.Table(returnTable).Create(&newReturn).Error; err != nil {
				return fmt.Errorf("创建退货记录失败: %w", err)
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("查询退货记录失败: %w", err)
		}

		// 已有记录：更新字段
		updates := map[string]interface{}{
			"reason":             resp.Reason,
			"text_reason":        resp.TextReason,
			"refund_amount":      refundAmount,
			"amount_before_disc": amountBeforeDisc,
			"currency":           resp.Currency,
			"status":             resp.Status,
			"needs_logistics":    resp.NeedsLogistics,
			"tracking_number":    resp.TrackingNumber,
			"logistics_status":   resp.LogisticsStatus,
			"buyer_username":     resp.User.Username,
			"shopee_create_time": shopeeCreateTime,
			"shopee_update_time": shopeeUpdateTime,
			"due_date":           dueDate,
		}

		// 状态变为已确认退款 且 尚未处理退款 → 标记为 PROCESSING 并处理
		r := &models.Return{Status: resp.Status}
		if r.IsRefundConfirmed() && existing.RefundStatus == models.ReturnRefundUnprocessed {
			updates["refund_status"] = models.ReturnRefundProcessing
			needProcessRefund = true
		}

		if err := tx.Table(returnTable).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新退货记录失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 在事务外处理退款（refund_status 已是 PROCESSING，其他协程不会重复处理）
	if needProcessRefund {
		s.processRefund(ctx, shopID, detail)
	}

	return nil
}

// processRefund 处理退货退款 — 返还预付款
//
// 前置条件：refund_status 已在 UpsertReturn 事务内被设为 PROCESSING（防止并发重入）
//
// 逻辑：
//  - 发货前全额退款：主单+子单 CANCELLED_BEFORE_SHIP，返还主单 prepayment_amount
//  - 发货前部分退款：退款子单 CANCELLED_BEFORE_SHIP，从回调/拉取数据填充子单 prepayment_amount，返还其和
//  - 已发货：按 refund_amount 比例返还
func (s *ReturnService) processRefund(ctx context.Context, shopID uint64, detail *shopee.ReturnDetailResponse) {
	resp := &detail.Response
	returnSN, orderSN := resp.ReturnSN, resp.OrderSN
	refundAmount := decimal.NewFromFloat(resp.RefundAmount)

	returnTable := database.GetReturnTableName(shopID)
	orderTable := database.GetOrderTableName(shopID)
	orderItemTable := database.GetOrderItemTableName(shopID)
	shipmentRecordTable := database.GetOrderShipmentRecordTableName(shopID)

	var order models.Order
	if err := s.db.Table(orderTable).
		Select("id", "shop_id", "order_sn", "total_amount", "prepayment_amount", "prepayment_status").
		Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
		First(&order).Error; err != nil {
		fmt.Printf("[ReturnService] 店铺=%d 退货=%s 关联订单=%s 未找到: %v\n", shopID, returnSN, orderSN, err)
		s.markRefundStatus(returnTable, shopID, returnSN, models.ReturnRefundSkipped)
		return
	}

	if order.PrepaymentStatus != models.PrepaymentSufficient {
		fmt.Printf("[ReturnService] 店铺=%d 退货=%s 订单=%s 预付款状态=%d 无需返还\n", shopID, returnSN, orderSN, order.PrepaymentStatus)
		s.markRefundStatus(returnTable, shopID, returnSN, models.ReturnRefundSkipped)
		return
	}

	var shop models.Shop
	if err := s.db.Select("admin_id").Where("shop_id = ?", shopID).First(&shop).Error; err != nil || shop.AdminID == 0 {
		fmt.Printf("[ReturnService] 店铺=%d 查找店主失败\n", shopID)
		s.markRefundStatus(returnTable, shopID, returnSN, models.ReturnRefundFailed)
		return
	}

	// 是否发货前（运营未发货）
	var shipmentRecord models.OrderShipmentRecord
	beforeShip := !(s.db.Table(shipmentRecordTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
		First(&shipmentRecord).Error == nil && shipmentRecord.Status == models.ShipmentRecordStatusShipped)

	isFullRefund := refundAmount.GreaterThanOrEqual(order.TotalAmount)
	var unfreezeAmount decimal.Decimal

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if beforeShip && isFullRefund {
			unfreezeAmount = order.PrepaymentAmount
			if !unfreezeAmount.IsPositive() {
				unfreezeAmount = order.TotalAmount
			}
		} else if beforeShip && !isFullRefund {
			unfreezeAmount = s.computePartialRefundPrepaymentInTx(tx, shopID, orderSN, order, detail)
		} else {
			unfreezeAmount = refundAmount
			if unfreezeAmount.GreaterThan(order.TotalAmount) {
				unfreezeAmount = order.TotalAmount
			}
		}

		_, innerErr := s.accountService.UnfreezePrepaymentInTx(tx, ctx, shop.AdminID, unfreezeAmount, orderSN,
			fmt.Sprintf("退货退款返还-退货单%s 订单%s", returnSN, orderSN))
		if innerErr != nil {
			return fmt.Errorf("返还预付款失败: %w", innerErr)
		}

		if beforeShip {
			if isFullRefund {
				tx.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
					Updates(map[string]interface{}{"order_status": consts.OrderStatusCancelledBeforeShip, "prepayment_status": models.PrepaymentUnchecked})
				tx.Table(orderItemTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
					Update("order_status", consts.OrderStatusCancelledBeforeShip)
			} else {
				tx.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
					Update("prepayment_status", models.PrepaymentUnchecked)
				// 子单已在 computePartialRefundPrepayment 中更新
			}
		} else if isFullRefund {
			tx.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
				Update("prepayment_status", models.PrepaymentUnchecked)
		}

		now := time.Now()
		return tx.Table(returnTable).
			Where("shop_id = ? AND return_sn = ? AND refund_status = ?", shopID, returnSN, models.ReturnRefundProcessing).
			Updates(map[string]interface{}{"refund_status": models.ReturnRefundProcessed, "refund_processed_at": now}).Error
	})

	if err != nil {
		fmt.Printf("[ReturnService] 店铺=%d 退货=%s 退款处理失败: %v\n", shopID, returnSN, err)
		s.markRefundStatus(returnTable, shopID, returnSN, models.ReturnRefundFailed)
		return
	}
	fmt.Printf("[ReturnService] 店铺=%d 退货=%s 订单=%s 退款返还成功 金额=%s\n", shopID, returnSN, orderSN, unfreezeAmount.StringFixed(2))
}

// computePartialRefundPrepaymentInTx 在事务内计算部分退款应返还的预付款，并更新退款子单的 order_status、prepayment_amount
func (s *ReturnService) computePartialRefundPrepaymentInTx(tx *gorm.DB, shopID uint64, orderSN string, order models.Order, detail *shopee.ReturnDetailResponse) decimal.Decimal {
	resp := &detail.Response
	orderItemTable := database.GetOrderItemTableName(shopID)
	var orderItems []models.OrderItem
	if err := tx.Table(orderItemTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).Find(&orderItems).Error; err != nil {
		return decimal.Zero
	}

	refundAmount := decimal.NewFromFloat(resp.RefundAmount)
	orderPrepayment := order.PrepaymentAmount
	if !orderPrepayment.IsPositive() {
		orderPrepayment = order.TotalAmount
	}

	var totalUnfreeze decimal.Decimal
	itemMap := make(map[string]int) // "itemID:modelID" -> index in resp.Item
	for i, it := range resp.Item {
		itemMap[fmt.Sprintf("%d:%d", it.ItemID, it.ModelID)] = i
	}

	for i := range orderItems {
		oi := &orderItems[i]
		key := fmt.Sprintf("%d:%d", oi.ItemID, oi.ModelID)
		idx, ok := itemMap[key]
		if !ok {
			continue
		}
		ritem := resp.Item[idx]
		var itemPrepayment decimal.Decimal
		if ritem.EscrowAmount > 0 {
			itemPrepayment = decimal.NewFromFloat(ritem.EscrowAmount)
		} else {
			// 兜底：按退款金额比例分配
			itemRefund := decimal.NewFromFloat(float64(ritem.Amount) * ritem.ItemPrice)
			if order.TotalAmount.IsPositive() {
				itemPrepayment = orderPrepayment.Mul(itemRefund).Div(order.TotalAmount)
			}
		}
		totalUnfreeze = totalUnfreeze.Add(itemPrepayment)
		tx.Table(orderItemTable).Where("id = ?", oi.ID).Updates(map[string]interface{}{
			"order_status":      consts.OrderStatusCancelledBeforeShip,
			"prepayment_amount": itemPrepayment,
		})
	}

	if totalUnfreeze.IsZero() && len(resp.Item) > 0 {
		if order.TotalAmount.IsPositive() {
			totalUnfreeze = orderPrepayment.Mul(refundAmount).Div(order.TotalAmount)
		}
	}
	return totalUnfreeze
}

// markRefundStatus 更新退货记录的退款处理状态
func (s *ReturnService) markRefundStatus(returnTable string, shopID uint64, returnSN string, status int8) {
	now := time.Now()
	if err := s.db.Table(returnTable).
		Where("shop_id = ? AND return_sn = ? AND refund_status IN ?", shopID, returnSN,
			[]int8{models.ReturnRefundUnprocessed, models.ReturnRefundProcessing}).
		Updates(map[string]interface{}{
			"refund_status":       status,
			"refund_processed_at": now,
		}).Error; err != nil {
		fmt.Printf("[ReturnService] 店铺=%d 退货=%s 标记状态失败: %v\n", shopID, returnSN, err)
	}
}

// SyncReturns 同步指定店铺近期退货记录
//
// 调用时机：巡检时调用，拉取近15天内的退货记录并逐条同步详情
//
// 流程：
//  1. 调用 GetReturnList 分页获取退货列表
//  2. 过滤出本地不存在或状态有变更的记录
//  3. 调用 GetReturnDetail 获取详情
//  4. 调用 UpsertReturn 写入/更新
func (s *ReturnService) SyncReturns(ctx context.Context, shopID uint64, accessToken string, region string) error {
	client := shopee.NewClient(region)
	returnTable := database.GetReturnTableName(shopID)

	// 查询近15天的退货
	now := time.Now()
	createTimeTo := now.Unix()
	createTimeFrom := now.Add(-15 * 24 * time.Hour).Unix()

	cursor := ""
	pageSize := 50

	for {
		// 限流
		if err := shopee.WaitForRateLimit(ctx, shopID); err != nil {
			return fmt.Errorf("限流等待失败: %w", err)
		}

		var listResp *shopee.ReturnListResponse
		err := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
			var err error
			listResp, err = client.GetReturnList(accessToken, shopID, createTimeFrom, createTimeTo, pageSize, cursor)
			return err
		})
		if err != nil {
			return fmt.Errorf("获取退货列表失败: %w", err)
		}

		for _, item := range listResp.Response.ReturnList {
			// 检查本地是否已有且状态一致
			var existing models.Return
			localErr := s.db.Table(returnTable).
				Select("status").
				Where("shop_id = ? AND return_sn = ?", shopID, item.ReturnSN).
				First(&existing).Error

			if localErr == nil && existing.Status == item.Status {
				continue // 状态一致，跳过
			}

			// 需要同步详情
			if err := shopee.WaitForRateLimit(ctx, shopID); err != nil {
				if ctx.Err() != nil {
					return ctx.Err() // context 取消，提前退出
				}
				continue
			}

			var detailResp *shopee.ReturnDetailResponse
			detailErr := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
				var err error
				detailResp, err = client.GetReturnDetail(accessToken, shopID, item.ReturnSN)
				return err
			})
			if detailErr != nil {
				fmt.Printf("[ReturnSync] 店铺=%d 退货=%s 获取详情失败: %v\n", shopID, item.ReturnSN, detailErr)
				continue
			}

			if err := s.UpsertReturn(ctx, shopID, detailResp); err != nil {
				fmt.Printf("[ReturnSync] 店铺=%d 退货=%s 写入失败: %v\n", shopID, item.ReturnSN, err)
			}
		}

		if !listResp.Response.More {
			break
		}
		cursor = listResp.Response.NextCursor
	}

	return nil
}
