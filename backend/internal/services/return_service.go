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
		s.processRefund(ctx, shopID, resp.ReturnSN, resp.OrderSN, refundAmount)
	}

	return nil
}

// processRefund 处理退货退款 — 解冻预付款
//
// 前置条件：refund_status 已在 UpsertReturn 事务内被设为 PROCESSING（防止并发重入）
//
// 逻辑：
//  1. 查找关联订单的预付款状态：只有 prepayment_status=1（已冻结）才需要解冻
//  2. 部分退款：解冻 refund_amount；全部退款：解冻 total_amount
//  3. 解冻预付款 + 托管退回 + 更新订单 + 标记退货记录 全部在同一事务中完成
//  4. 失败时将 refund_status 标记为 FAILED（避免无限重试）
func (s *ReturnService) processRefund(ctx context.Context, shopID uint64, returnSN, orderSN string, refundAmount decimal.Decimal) {
	returnTable := database.GetReturnTableName(shopID)
	orderTable := database.GetOrderTableName(shopID)

	// 查找关联订单
	var order models.Order
	if err := s.db.Table(orderTable).
		Select("id", "shop_id", "order_sn", "total_amount", "prepayment_status").
		Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
		First(&order).Error; err != nil {
		fmt.Printf("[ReturnService] 店铺=%d 退货=%s 关联订单=%s 未找到: %v\n", shopID, returnSN, orderSN, err)
		s.markRefundStatus(returnTable, shopID, returnSN, models.ReturnRefundSkipped)
		return
	}

	// 只有预付款已冻结（prepayment_status=1）的订单才需要解冻
	if order.PrepaymentStatus != models.PrepaymentSufficient {
		fmt.Printf("[ReturnService] 店铺=%d 退货=%s 订单=%s 预付款状态=%d 无需解冻\n",
			shopID, returnSN, orderSN, order.PrepaymentStatus)
		s.markRefundStatus(returnTable, shopID, returnSN, models.ReturnRefundSkipped)
		return
	}

	// 查找店铺老板
	var shop models.Shop
	if err := s.db.Select("admin_id").Where("shop_id = ?", shopID).First(&shop).Error; err != nil || shop.AdminID == 0 {
		fmt.Printf("[ReturnService] 店铺=%d 查找店主失败\n", shopID)
		s.markRefundStatus(returnTable, shopID, returnSN, models.ReturnRefundFailed)
		return
	}

	// 确定解冻金额：取退款金额和订单总额的较小值
	unfreezeAmount := refundAmount
	if unfreezeAmount.GreaterThan(order.TotalAmount) {
		unfreezeAmount = order.TotalAmount
	}

	// 在同一事务中完成：解冻预付款 + 托管退回 + 订单状态更新 + 退货标记
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 解冻预付款（使用 InTx 版本，复用同一事务连接）
		_, err := s.accountService.UnfreezePrepaymentInTx(tx, ctx, shop.AdminID, unfreezeAmount, orderSN,
			fmt.Sprintf("退货退款解冻-退货单%s 订单%s", returnSN, orderSN))
		if err != nil {
			return fmt.Errorf("解冻预付款失败: %w", err)
		}

		// 从托管账户退回（尽力而为，失败不影响退款主流程）
		_ = s.accountService.TransferFromEscrowInTx(tx, ctx, shop.AdminID, unfreezeAmount, orderSN,
			fmt.Sprintf("退货退款退回-退货单%s", returnSN))

		// 如果是全额退款，更新订单预付款状态为未检查
		if refundAmount.GreaterThanOrEqual(order.TotalAmount) {
			if err := tx.Table(orderTable).
				Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
				Updates(map[string]interface{}{
					"prepayment_status": models.PrepaymentUnchecked,
				}).Error; err != nil {
				return fmt.Errorf("更新订单预付款状态失败: %w", err)
			}
		}

		// 标记退货记录为已处理
		now := time.Now()
		if err := tx.Table(returnTable).
			Where("shop_id = ? AND return_sn = ? AND refund_status = ?", shopID, returnSN, models.ReturnRefundProcessing).
			Updates(map[string]interface{}{
				"refund_status":       models.ReturnRefundProcessed,
				"refund_processed_at": now,
			}).Error; err != nil {
			return fmt.Errorf("标记退货已处理失败: %w", err)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("[ReturnService] 店铺=%d 退货=%s 退款处理失败: %v\n", shopID, returnSN, err)
		s.markRefundStatus(returnTable, shopID, returnSN, models.ReturnRefundFailed)
		return
	}

	fmt.Printf("[ReturnService] 店铺=%d 退货=%s 订单=%s 退款解冻成功 金额=%s\n",
		shopID, returnSN, orderSN, unfreezeAmount.StringFixed(2))
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
