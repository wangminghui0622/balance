package shopower

import (
	"context"
	"fmt"
	"time"

	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/shopee"
	"balance/backend/internal/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// EscrowService 结算明细服务
type EscrowService struct {
	db          *gorm.DB
	shopService *ShopService
	shardedDB   *database.ShardedDB
}

// NewEscrowService 创建结算明细服务
func NewEscrowService() *EscrowService {
	db := database.GetDB()
	return &EscrowService{
		db:          db,
		shopService: NewShopService(),
		shardedDB:   database.NewShardedDB(db),
	}
}

// SyncOrderEscrow 同步单个订单的结算明细
func (s *EscrowService) SyncOrderEscrow(ctx context.Context, adminID int64, shopID int64, orderSN string) (*models.OrderEscrow, error) {
	shop, err := s.shopService.GetShop(ctx, adminID, shopID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.getAccessToken(ctx, uint64(shopID))
	if err != nil {
		return nil, err
	}

	return s.fetchAndSaveEscrow(ctx, uint64(shopID), shop.Region, shop.Currency, accessToken, orderSN)
}

// SyncOrderEscrowInternal 内部同步（无权限检查，供调度器使用）
func (s *EscrowService) SyncOrderEscrowInternal(ctx context.Context, shopID uint64, orderSN string) (*models.OrderEscrow, error) {
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return nil, utils.ErrShopNotFound
	}

	accessToken, err := s.getAccessToken(ctx, shopID)
	if err != nil {
		return nil, err
	}

	return s.fetchAndSaveEscrow(ctx, shopID, shop.Region, shop.Currency, accessToken, orderSN)
}

// fetchAndSaveEscrow 获取并保存结算明细
func (s *EscrowService) fetchAndSaveEscrow(ctx context.Context, shopID uint64, region, currency, accessToken, orderSN string) (*models.OrderEscrow, error) {
	client := shopee.NewClient(region)
	limiter := shopee.GetRateLimiter(shopID)

	if err := limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("限流等待被取消: %w", err)
	}

	var escrowResp *shopee.EscrowDetailResponse
	err := shopee.RetryWithBackoff(ctx, consts.ShopeeAPIRetryTimes, func() error {
		var err error
		escrowResp, err = client.GetEscrowDetail(accessToken, shopID, orderSN)
		return err
	})
	if err != nil {
		// 记录同步失败
		s.markSyncFailed(shopID, orderSN, err.Error())
		return nil, fmt.Errorf("获取结算明细失败: %w", err)
	}

	return s.saveEscrow(ctx, shopID, currency, escrowResp)
}

// saveEscrow 保存结算明细
func (s *EscrowService) saveEscrow(ctx context.Context, shopID uint64, currency string, resp *shopee.EscrowDetailResponse) (*models.OrderEscrow, error) {
	income := resp.Response.OrderIncome

	// 查找订单ID - 使用分表
	orderTable := database.GetOrderTableName(shopID)
	var order models.Order
	var orderID uint64
	if err := s.db.Table(orderTable).Select("id").Where("shop_id = ? AND order_sn = ?", shopID, resp.Response.OrderSN).First(&order).Error; err == nil {
		orderID = order.ID
	}

	now := time.Now()
	escrow := models.OrderEscrow{
		ShopID:                   shopID,
		OrderSN:                  resp.Response.OrderSN,
		OrderID:                  orderID,
		Currency:                 currency,
		EscrowAmount:             decimal.NewFromFloat(income.EscrowAmount),
		BuyerTotalAmount:         decimal.NewFromFloat(income.BuyerTotalAmount),
		OriginalPrice:            decimal.NewFromFloat(income.OriginalPrice),
		SellerDiscount:           decimal.NewFromFloat(income.SellerDiscount),
		ShopeeDiscount:           decimal.NewFromFloat(income.ShopeeDiscount),
		VoucherFromSeller:        decimal.NewFromFloat(income.VoucherFromSeller),
		VoucherFromShopee:        decimal.NewFromFloat(income.VoucherFromShopee),
		Coins:                    decimal.NewFromFloat(income.Coins),
		BuyerPaidShippingFee:     decimal.NewFromFloat(income.BuyerPaidShippingFee),
		FinalShippingFee:         decimal.NewFromFloat(income.FinalShippingFee),
		ActualShippingFee:        decimal.NewFromFloat(income.ActualShippingFee),
		EstimatedShippingFee:     decimal.NewFromFloat(income.EstimatedShippingFee),
		ShippingFeeDiscount:      decimal.NewFromFloat(income.ShippingFeeDiscountFrom3PL),
		SellerShippingDiscount:   decimal.NewFromFloat(income.SellerShippingDiscount),
		ReverseShippingFee:       decimal.NewFromFloat(income.ReverseShippingFee),
		CommissionFee:            decimal.NewFromFloat(income.CommissionFee),
		ServiceFee:               decimal.NewFromFloat(income.ServiceFee),
		SellerTransactionFee:     decimal.NewFromFloat(income.SellerTransactionFee),
		BuyerTransactionFee:      decimal.NewFromFloat(income.BuyerTransactionFee),
		CreditCardTransactionFee: decimal.NewFromFloat(income.CreditCardTransactionFee),
		EscrowTax:                decimal.NewFromFloat(income.EscrowTax),
		CrossBorderTax:           decimal.NewFromFloat(income.CrossBorderTax),
		PaymentPromotion:         decimal.NewFromFloat(income.PaymentPromotion),
		CreditCardPromotion:      decimal.NewFromFloat(income.CreditCardPromotion),
		SellerLostCompensation:   decimal.NewFromFloat(income.SellerLostCompensation),
		SellerCoinCashBack:       decimal.NewFromFloat(income.SellerCoinCashBack),
		SellerReturnRefund:       decimal.NewFromFloat(income.SellerReturnRefund),
		FinalProductProtection:   decimal.NewFromFloat(income.FinalProductProtection),
		CostOfGoodsSold:          decimal.NewFromFloat(income.CostOfGoodsSold),
		OriginalCostOfGoodsSold:  decimal.NewFromFloat(income.OriginalCostOfGoodsSold),
		DrcAdjustableRefund:      decimal.NewFromFloat(income.DrcAdjustableRefund),
		ItemsCount:               income.ItemsCount,
		SyncStatus:               models.EscrowSyncStatusSuccess,
		SyncTime:                 &now,
	}

	// 使用事务保存 - 使用分表
	escrowTable := database.GetOrderEscrowTableName(shopID)
	escrowItemTable := database.GetOrderEscrowItemTableName(shopID)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 保存或更新结算明细
		if err := tx.Table(escrowTable).Where("shop_id = ? AND order_sn = ?", shopID, resp.Response.OrderSN).
			Assign(escrow).FirstOrCreate(&escrow).Error; err != nil {
			return err
		}

		// 删除旧的商品明细
		if err := tx.Table(escrowItemTable).Where("escrow_id = ?", escrow.ID).Delete(&models.OrderEscrowItem{}).Error; err != nil {
			return err
		}

		// 保存商品明细
		for _, item := range resp.Response.Items {
			escrowItem := models.OrderEscrowItem{
				EscrowID:                  escrow.ID,
				ShopID:                    shopID,
				OrderSN:                   resp.Response.OrderSN,
				ItemID:                    uint64(item.ItemID),
				ItemName:                  item.ItemName,
				ItemSKU:                   item.ItemSKU,
				ModelID:                   uint64(item.ModelID),
				ModelName:                 item.ModelName,
				ModelSKU:                  item.ModelSKU,
				QuantityPurchased:         item.QuantityPurchased,
				OriginalPrice:             decimal.NewFromFloat(item.OriginalPrice),
				DiscountedPrice:           decimal.NewFromFloat(item.DiscountedPrice),
				SellerDiscount:            decimal.NewFromFloat(item.SellerDiscount),
				ShopeeDiscount:            decimal.NewFromFloat(item.ShopeeDiscount),
				DiscountFromCoin:          decimal.NewFromFloat(item.DiscountFromCoin),
				DiscountFromVoucher:       decimal.NewFromFloat(item.DiscountFromVoucher),
				DiscountFromVoucherSeller: decimal.NewFromFloat(item.DiscountFromVoucherSeller),
				DiscountFromVoucherShopee: decimal.NewFromFloat(item.DiscountFromVoucherShopee),
				ActivityType:              item.ActivityType,
				ActivityID:                uint64(item.ActivityID),
			}
			if err := tx.Table(escrowItemTable).Create(&escrowItem).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("保存结算明细失败: %w", err)
	}

	return &escrow, nil
}

// markSyncFailed 标记同步失败 - 使用分表
func (s *EscrowService) markSyncFailed(shopID uint64, orderSN, errMsg string) {
	now := time.Now()
	escrowTable := database.GetOrderEscrowTableName(shopID)
	s.db.Table(escrowTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).
		Assign(models.OrderEscrow{
			ShopID:     shopID,
			OrderSN:    orderSN,
			SyncStatus: models.EscrowSyncStatusFailed,
			SyncTime:   &now,
			SyncError:  errMsg,
		}).FirstOrCreate(&models.OrderEscrow{})
}

// GetOrderEscrow 获取订单结算明细 - 使用分表
func (s *EscrowService) GetOrderEscrow(ctx context.Context, adminID int64, shopID int64, orderSN string) (*models.OrderEscrow, error) {
	if _, err := s.shopService.GetShop(ctx, adminID, shopID); err != nil {
		return nil, err
	}

	escrowTable := database.GetOrderEscrowTableName(uint64(shopID))
	var escrow models.OrderEscrow
	if err := s.db.Table(escrowTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&escrow).Error; err != nil {
		return nil, fmt.Errorf("结算明细不存在")
	}
	return &escrow, nil
}

// GetOrderEscrowWithItems 获取订单结算明细（含商品明细）- 使用分表
func (s *EscrowService) GetOrderEscrowWithItems(ctx context.Context, adminID int64, shopID int64, orderSN string) (*models.OrderEscrow, []models.OrderEscrowItem, error) {
	escrow, err := s.GetOrderEscrow(ctx, adminID, shopID, orderSN)
	if err != nil {
		return nil, nil, err
	}

	escrowItemTable := database.GetOrderEscrowItemTableName(uint64(shopID))
	var items []models.OrderEscrowItem
	s.db.Table(escrowItemTable).Where("escrow_id = ?", escrow.ID).Find(&items)

	return escrow, items, nil
}

// ListPendingEscrows 获取待同步的结算明细订单 - 使用分表
func (s *EscrowService) ListPendingEscrows(ctx context.Context, shopID uint64, limit int) ([]string, error) {
	var orderSNs []string

	orderTable := database.GetOrderTableName(shopID)
	escrowTable := database.GetOrderEscrowTableName(shopID)

	// 查找已完成但未同步结算明细的订单
	subQuery := s.db.Table(escrowTable).Select("order_sn").Where("shop_id = ?", shopID)

	err := s.db.Table(orderTable).
		Select("order_sn").
		Where("shop_id = ? AND order_status = ?", shopID, consts.OrderStatusCompleted).
		Where("order_sn NOT IN (?)", subQuery).
		Limit(limit).
		Pluck("order_sn", &orderSNs).Error

	return orderSNs, err
}

// BatchSyncEscrows 批量同步结算明细
func (s *EscrowService) BatchSyncEscrows(ctx context.Context, shopID uint64, orderSNs []string) (int, int, error) {
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return 0, 0, utils.ErrShopNotFound
	}

	accessToken, err := s.getAccessToken(ctx, shopID)
	if err != nil {
		return 0, 0, err
	}

	successCount := 0
	failCount := 0

	for _, orderSN := range orderSNs {
		_, err := s.fetchAndSaveEscrow(ctx, shopID, shop.Region, shop.Currency, accessToken, orderSN)
		if err != nil {
			failCount++
		} else {
			successCount++
		}
		// 避免请求过快
		time.Sleep(500 * time.Millisecond)
	}

	return successCount, failCount, nil
}

func (s *EscrowService) getAccessToken(ctx context.Context, shopID uint64) (string, error) {
	var auth models.ShopAuthorization
	if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
		return "", utils.ErrShopUnauthorized
	}
	if auth.IsAccessTokenExpired() {
		if err := s.shopService.RefreshToken(ctx, shopID); err != nil {
			return "", err
		}
		s.db.Where("shop_id = ?", shopID).First(&auth)
	}
	return auth.AccessToken, nil
}
