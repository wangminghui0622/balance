package services

import (
	"context"
	"fmt"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/services/shopower"
	"balance/backend/internal/shopee"
	"balance/backend/internal/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PrepaymentCheckService 预付款实时检查服务
// 当订单进入 READY_TO_SHIP 状态时，自动从店铺老板预付款扣除订单金额。
// 核心保证：同一订单无论从 Webhook 还是 Sync 进来，预付款只扣一次。
type PrepaymentCheckService struct {
	db             *gorm.DB
	accountService *AccountService
	idGenerator    *utils.IDGenerator
}

// NewPrepaymentCheckService 创建预付款检查服务
func NewPrepaymentCheckService() *PrepaymentCheckService {
	return &PrepaymentCheckService{
		db:             database.GetDB(),
		accountService: NewAccountService(),
		idGenerator:    utils.NewIDGenerator(database.GetRedis()),
	}
}

// NewOrderServiceWithPrepaymentCheck 创建带预付款检查和结算明细获取的订单服务
// 用于 Sync / Handler 等场景：
//   - READY_TO_SHIP 时自动调 get_escrow_detail 获取费用明细
//   - 然后用 escrow_amount 检查并冻结预付款
func NewOrderServiceWithPrepaymentCheck() *shopower.OrderService {
	svc := shopower.NewOrderService()
	checkSvc := NewPrepaymentCheckService()

	// 注入预付款检查回调
	svc.SetPrepaymentCheckFunc(func(ctx context.Context, shopID uint64, orderSN string, prepaymentAmount decimal.Decimal, orderTable string) {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if err := checkSvc.CheckAndDeductForOrder(bgCtx, shopID, orderSN, prepaymentAmount, orderTable); err != nil {
				fmt.Printf("[PrepaymentCheck] 店铺=%d 订单=%s 预付款检查失败: %v\n", shopID, orderSN, err)
			}
		}()
	})

	// 注入结算明细获取回调（READY_TO_SHIP 时调用，获取虾皮费用明细）
	svc.SetEscrowFetchFunc(func(ctx context.Context, shopID uint64, orderSN string) (*shopee.EscrowDetailResponse, error) {
		// 获取店铺信息和授权令牌
		db := database.GetDB()
		var shop models.Shop
		if err := db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
			return nil, fmt.Errorf("查询店铺失败: %w", err)
		}
		var auth models.ShopAuthorization
		if err := db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
			return nil, fmt.Errorf("查询授权失败: %w", err)
		}

		// 等待限流
		if err := shopee.WaitForRateLimit(ctx, shopID); err != nil {
			return nil, fmt.Errorf("限流等待被取消: %w", err)
		}

		// 调用 Shopee get_escrow_detail API
		client := shopee.NewClient(shop.Region)
		var escrowResp *shopee.EscrowDetailResponse
		err := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
			var err error
			escrowResp, err = client.GetEscrowDetail(auth.AccessToken, shopID, orderSN)
			return err
		})
		if err != nil {
			return nil, fmt.Errorf("获取结算明细失败: %w", err)
		}
		return escrowResp, nil
	})

	return svc
}

// CheckAndDeductForOrder 检查并扣除订单预付款（幂等，防重复扣款）
//
// 调用时机：订单状态变为 READY_TO_SHIP 时（无论来自 Webhook 还是 Sync）
//
// 幂等保证：
//
//	在同一个 DB 事务内，先用 SELECT ... FOR UPDATE 锁住订单行并检查
//	prepayment_status=0，然后再执行扣款和标记。整个「检查→扣款→标记」
//	是原子的，彻底杜绝 Sync+Webhook 并发导致的重复扣款。
//
// 流程：
//  1. 查找 shop -> shop_owner
//  2. 在一个事务中：行锁订单 → 检查 prepayment_status=0 → 扣款/标记
//  3. 余额不足时标记订单（prepayment_status=2），发通知
//
// 返回 error 仅在系统错误时返回，预付款不足属于业务正常流程不返回 error
func (s *PrepaymentCheckService) CheckAndDeductForOrder(ctx context.Context, shopID uint64, orderSN string, orderAmount decimal.Decimal, orderTable string) error {
	// ===== 第 1 步：快速预检查（无锁，避免无谓加锁开销）=====
	var precheck models.Order
	if err := s.db.Table(orderTable).
		Select("id", "prepayment_status").
		Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
		First(&precheck).Error; err != nil {
		return fmt.Errorf("查询订单失败: %w", err)
	}
	if precheck.PrepaymentStatus != models.PrepaymentUnchecked {
		return nil // 已处理过，跳过
	}

	// ===== 第 2 步：查找店铺老板 =====
	var shop models.Shop
	if err := s.db.Select("admin_id").Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return fmt.Errorf("查询店铺失败: %w", err)
	}
	if shop.AdminID == 0 {
		return nil
	}
	shopOwnerID := shop.AdminID

	// ===== 第 3 步：在事务中「行锁订单 → 检查 → 扣款 → 标记」原子完成 =====
	now := time.Now()
	var insufficientBalance decimal.Decimal
	var needNotify bool

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 行锁订单行：防止 Sync 和 Webhook 并发处理同一订单
		var lockedOrder models.Order
		if err := tx.Table(orderTable).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "prepayment_status", "total_amount", "prepayment_amount").
			Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
			First(&lockedOrder).Error; err != nil {
			return fmt.Errorf("锁定订单失败: %w", err)
		}

		// 优先使用订单记录中的 prepayment_amount（来自 escrow_detail），其次使用传入参数
		if lockedOrder.PrepaymentAmount.IsPositive() {
			orderAmount = lockedOrder.PrepaymentAmount
		}

		// 二次检查（拿到锁后再确认）
		if lockedOrder.PrepaymentStatus != models.PrepaymentUnchecked {
			return nil // 已被其他协程处理，跳过
		}

		// 获取预付款账户（在事务外获取余额快照即可，扣款在 FreezePrepayment 内部有自己的行锁）
		account, err := s.accountService.GetOrCreatePrepaymentAccount(ctx, shopOwnerID)
		if err != nil {
			return fmt.Errorf("获取预付款账户失败: %w", err)
		}

		if account.Balance.GreaterThanOrEqual(orderAmount) {
			// 余额充足：冻结预付款（使用 InTx 版本，复用外层事务连接，避免连接池死锁）
			_, err := s.accountService.FreezePrepaymentInTx(tx, ctx, shopOwnerID, orderAmount, orderSN,
				fmt.Sprintf("订单预付款冻结(自动)-订单%s", orderSN))
			if err != nil {
				return fmt.Errorf("冻结预付款失败: %w", err)
			}

			// 获取冻结后余额快照（简单读查询，直接在事务内执行）
			var acct models.PrepaymentAccount
			var balanceAfter decimal.Decimal
			if err := tx.Where("admin_id = ?", shopOwnerID).First(&acct).Error; err == nil {
				balanceAfter = acct.Balance
			}

			// 标记订单为「充足」
			return tx.Table(orderTable).
				Where("id = ?", lockedOrder.ID).
				Updates(map[string]interface{}{
					"prepayment_status":     models.PrepaymentSufficient,
					"prepayment_snapshot":   balanceAfter,
					"prepayment_checked_at": now,
				}).Error
		}

		// 余额不足：标记订单
		insufficientBalance = account.Balance
		needNotify = true
		return tx.Table(orderTable).
			Where("id = ?", lockedOrder.ID).
			Updates(map[string]interface{}{
				"prepayment_status":     models.PrepaymentInsufficient,
				"prepayment_snapshot":   account.Balance,
				"prepayment_checked_at": now,
			}).Error
	})

	if err != nil {
		return err
	}

	// 通知在事务外发送（避免事务持有时间过长）
	if needNotify {
		s.notifyInsufficientBalance(ctx, shopID, shopOwnerID, orderSN, orderAmount, insufficientBalance, &now)
	}

	return nil
}

// BackfillInsufficientOrders 充值后补扣历史「预付款不足」订单
//
// 调用时机：店主充值预付款成功后
//
// 流程：
//  1. 查找该店主名下所有店铺
//  2. 对每个店铺，在对应分表中查出 prepayment_status=2 且 order_status=READY_TO_SHIP 的订单
//  3. 按订单创建时间升序，逐笔尝试冻结预付款
//  4. 每笔订单在独立事务中：行锁订单行 -> CAS检查 -> 冻结预付款 -> 更新订单状态
//  5. 余额不足时停止（后续订单也不可能够）
//
// 并发安全：
//   - 订单行使用 SELECT ... FOR UPDATE 行锁，防止同一订单被重复补扣
//   - 预付款账户在 FreezePrepayment 内部也有 FOR UPDATE 行锁
//   - 两者组合保证补扣的原子性和幂等性
func (s *PrepaymentCheckService) BackfillInsufficientOrders(ctx context.Context, adminID int64) (successCount int, failCount int, err error) {
	// 第 1 步：查找该店主名下所有店铺
	var shops []models.Shop
	if err := s.db.Select("shop_id").Where("admin_id = ?", adminID).Find(&shops).Error; err != nil {
		return 0, 0, fmt.Errorf("查询店主店铺失败: %w", err)
	}
	if len(shops) == 0 {
		return 0, 0, nil
	}

	// 第 2 步：收集所有「预付款不足」的 READY_TO_SHIP 订单
	type insufficientOrder struct {
		ShopID     uint64
		OrderSN    string
		TableName  string
	}
	var pendingOrders []insufficientOrder

	for _, shop := range shops {
		tableName := database.GetOrderTableName(shop.ShopID)
		var orders []models.Order
		if err := s.db.Table(tableName).
			Select("shop_id", "order_sn").
			Where("shop_id = ? AND prepayment_status = ? AND order_status = ?",
				shop.ShopID, models.PrepaymentInsufficient, "READY_TO_SHIP").
			Order("create_time ASC").
			Find(&orders).Error; err != nil {
			fmt.Printf("[BackfillPrepayment] 查询店铺 %d 不足订单失败: %v\n", shop.ShopID, err)
			continue
		}
		for _, o := range orders {
			pendingOrders = append(pendingOrders, insufficientOrder{
				ShopID:    o.ShopID,
				OrderSN:   o.OrderSN,
				TableName: tableName,
			})
		}
	}

	if len(pendingOrders) == 0 {
		return 0, 0, nil
	}

	// 第 3 步：逐笔补扣
	for _, po := range pendingOrders {
		deductErr := s.db.Transaction(func(tx *gorm.DB) error {
			// 行锁订单行，防止并发补扣同一订单
			var lockedOrder models.Order
			if err := tx.Table(po.TableName).
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Select("id", "prepayment_status", "total_amount", "prepayment_amount").
				Where("shop_id = ? AND order_sn = ?", po.ShopID, po.OrderSN).
				First(&lockedOrder).Error; err != nil {
				return fmt.Errorf("锁定订单失败: %w", err)
			}

			// CAS 检查：只处理仍然是「不足」状态的订单
			if lockedOrder.PrepaymentStatus != models.PrepaymentInsufficient {
				return nil // 已被其他协程处理，跳过
			}

			// 优先使用 prepayment_amount（来自 escrow_detail），其次 total_amount
			orderAmount := lockedOrder.PrepaymentAmount
			if !orderAmount.IsPositive() {
				orderAmount = lockedOrder.TotalAmount
			}

			// 冻结预付款（使用 InTx 版本，复用外层事务连接，避免连接池死锁）
			_, freezeErr := s.accountService.FreezePrepaymentInTx(tx, ctx, adminID, orderAmount, po.OrderSN,
				fmt.Sprintf("订单预付款补扣(充值后)-订单%s", po.OrderSN))
			if freezeErr != nil {
				return freezeErr // 余额不足或其他错误
			}

			// 冻结成功，获取最新余额快照（在事务内读取）
			var acct models.PrepaymentAccount
			var balanceAfter decimal.Decimal
			if err := tx.Where("admin_id = ?", adminID).First(&acct).Error; err == nil {
				balanceAfter = acct.Balance
			}

			// 更新订单状态为「充足」
			if err := tx.Table(po.TableName).
				Where("shop_id = ? AND order_sn = ? AND prepayment_status = ?",
					po.ShopID, po.OrderSN, models.PrepaymentInsufficient).
				Updates(map[string]interface{}{
					"prepayment_status":     models.PrepaymentSufficient,
					"prepayment_snapshot":   balanceAfter,
					"prepayment_checked_at": time.Now(),
				}).Error; err != nil {
				return fmt.Errorf("更新订单预付款状态失败: %w", err)
			}

			return nil
		})

		if deductErr != nil {
			// 如果是余额不足，停止后续补扣（余额只会越来越少）
			fmt.Printf("[BackfillPrepayment] 订单 %s 补扣失败: %v，停止后续补扣\n", po.OrderSN, deductErr)
			failCount += len(pendingOrders) - successCount - failCount
			break
		}
		successCount++
	}

	return successCount, failCount, nil
}

// notifyInsufficientBalance 发送预付款不足通知（带冷却时间去重）
func (s *PrepaymentCheckService) notifyInsufficientBalance(ctx context.Context, shopID uint64, shopOwnerID int64, orderSN string, required, available decimal.Decimal, now *time.Time) {
	rdb := database.GetRedis()
	notifyKey := fmt.Sprintf(consts.KeyPrepaymentNotified, shopID)

	// 冷却时间内不重复通知
	ok, err := rdb.SetNX(ctx, notifyKey, 1, consts.PrepaymentNotifiedCooldown).Result()
	if err != nil || !ok {
		return
	}

	// 查询店铺名称
	var shop models.Shop
	shopName := fmt.Sprintf("店铺%d", shopID)
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err == nil {
		shopName = shop.ShopName
	}

	title := fmt.Sprintf("【预付款不足】%s", shopName)
	content := fmt.Sprintf(
		"您的店铺「%s」有新订单 %s 进入待发货状态，订单金额 %s，但预付款余额仅 %s，不足以覆盖此订单。请尽快充值预付款以确保正常发货。",
		shopName, orderSN, required.StringFixed(2), available.StringFixed(2),
	)

	notifyID, _ := s.idGenerator.GenerateNotificationID(ctx)
	notification := &models.Notification{
		ID:        uint64(notifyID),
		AdminID:   shopOwnerID,
		ShopID:    shopID,
		OrderSN:   orderSN,
		Type:      models.NotifyTypePrepaymentLow,
		Title:     title,
		Content:   content,
		CreatedAt: *now,
	}
	s.db.Create(notification)
}
