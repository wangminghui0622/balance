package shopower

import (
	"context"
	"fmt"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/shopee"
	"balance/backend/internal/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// FinanceService 财务收入服务
type FinanceService struct {
	db          *gorm.DB
	shopService *ShopService
	shardedDB   *database.ShardedDB
}

// NewFinanceService 创建财务收入服务
func NewFinanceService() *FinanceService {
	db := database.GetDB()
	return &FinanceService{
		db:          db,
		shopService: NewShopService(),
		shardedDB:   database.NewShardedDB(db),
	}
}

// SyncWalletTransactions 同步店铺钱包交易记录
func (s *FinanceService) SyncWalletTransactions(ctx context.Context, adminID int64, shopID int64) (int, error) {
	shop, err := s.shopService.GetShop(ctx, adminID, shopID)
	if err != nil {
		return 0, err
	}

	accessToken, err := s.getAccessToken(ctx, uint64(shopID))
	if err != nil {
		return 0, err
	}

	return s.fetchAndSaveTransactions(ctx, uint64(shopID), shop.Region, accessToken)
}

// SyncWalletTransactionsInternal 内部同步（无权限检查，供调度器使用）
func (s *FinanceService) SyncWalletTransactionsInternal(ctx context.Context, shopID uint64) (int, error) {
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return 0, utils.ErrShopNotFound
	}

	accessToken, err := s.getAccessToken(ctx, shopID)
	if err != nil {
		return 0, err
	}

	return s.fetchAndSaveTransactions(ctx, shopID, shop.Region, accessToken)
}

// fetchAndSaveTransactions 获取并保存交易记录
func (s *FinanceService) fetchAndSaveTransactions(ctx context.Context, shopID uint64, region, accessToken string) (int, error) {
	client := shopee.NewClient(region)
	totalCount := 0
	pageNo := 1
	pageSize := 100

	for {
		resp, err := client.GetWalletTransactionList(accessToken, shopID, pageNo, pageSize, "")
		if err != nil {
			return totalCount, fmt.Errorf("获取钱包交易记录失败: %w", err)
		}

		if len(resp.Response.TransactionList) == 0 {
			break
		}

		for _, tx := range resp.Response.TransactionList {
			// 跳过提现记录（只记录收入）
			if tx.TransactionType == models.TransactionTypeWithdrawalCreated ||
				tx.TransactionType == models.TransactionTypeWithdrawalCompleted {
				continue
			}

			// 检查是否已存在 - 使用分表
			financeTable := database.GetFinanceIncomeTableName(shopID)
			var existing models.FinanceIncome
			if err := s.db.Table(financeTable).Where("transaction_id = ?", tx.TransactionID).First(&existing).Error; err == nil {
				continue // 已存在，跳过
			}

			income := models.FinanceIncome{
				ShopID:             shopID,
				TransactionID:      tx.TransactionID,
				OrderSN:            tx.OrderSN,
				RefundSN:           tx.RefundSN,
				Status:             tx.Status,
				WalletType:         tx.WalletType,
				TransactionType:    tx.TransactionType,
				Amount:             decimal.NewFromFloat(tx.Amount),
				CurrentBalance:     decimal.NewFromFloat(tx.CurrentBalance),
				TransactionTime:    tx.CreateTime,
				TransactionFee:     decimal.NewFromFloat(tx.TransactionFee),
				Description:        tx.Description,
				BuyerName:          tx.BuyerName,
				Reason:             tx.Reason,
				WithdrawalID:       tx.WithdrawalID,
				WithdrawalType:     tx.WithdrawalType,
				TransactionTabType: tx.TransactionTabType,
				MoneyFlow:          tx.MoneyFlow,
				SettlementHandleStatus: models.SettlementStatusPending,
			}

			if err := s.db.Table(financeTable).Create(&income).Error; err == nil {
				totalCount++
			}
		}

		if !resp.Response.More {
			break
		}
		pageNo++

		// 避免请求过快
		time.Sleep(500 * time.Millisecond)
	}

	return totalCount, nil
}

// ListFinanceIncomes 获取财务收入列表 - 使用分表
func (s *FinanceService) ListFinanceIncomes(ctx context.Context, adminID int64, shopID int64, orderSN string, transactionType string, settlementStatus int, page, pageSize int) ([]models.FinanceIncome, int64, error) {
	if _, err := s.shopService.GetShop(ctx, adminID, shopID); err != nil {
		return nil, 0, err
	}

	var incomes []models.FinanceIncome
	var total int64

	financeTable := database.GetFinanceIncomeTableName(uint64(shopID))
	query := s.db.Table(financeTable).Where("shop_id = ?", shopID)

	if orderSN != "" {
		query = query.Where("order_sn = ?", orderSN)
	}
	if transactionType != "" {
		query = query.Where("transaction_type = ?", transactionType)
	}
	if settlementStatus >= 0 {
		query = query.Where("settlement_handle_status = ?", settlementStatus)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("transaction_time DESC").Offset(offset).Limit(pageSize).Find(&incomes).Error; err != nil {
		return nil, 0, err
	}

	return incomes, total, nil
}

// GetFinanceIncomeByOrderSN 根据订单号获取财务收入 - 使用分表
func (s *FinanceService) GetFinanceIncomeByOrderSN(ctx context.Context, shopID uint64, orderSN string) (*models.FinanceIncome, error) {
	financeTable := database.GetFinanceIncomeTableName(shopID)
	var income models.FinanceIncome
	if err := s.db.Table(financeTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&income).Error; err != nil {
		return nil, fmt.Errorf("财务记录不存在")
	}
	return &income, nil
}

// GetUnsettledIncomes 获取待结算的收入记录 - 使用分表
func (s *FinanceService) GetUnsettledIncomes(ctx context.Context, shopID uint64, limit int) ([]models.FinanceIncome, error) {
	financeTable := database.GetFinanceIncomeTableName(shopID)
	var incomes []models.FinanceIncome
	err := s.db.Table(financeTable).Where("shop_id = ? AND settlement_handle_status = ? AND transaction_type = ?",
		shopID, models.SettlementStatusPending, models.TransactionTypeEscrowVerifiedAdd).
		Order("transaction_time ASC").
		Limit(limit).
		Find(&incomes).Error
	return incomes, err
}

// MarkSettled 标记为已结算 - 需要shopID来定位分表
func (s *FinanceService) MarkSettled(ctx context.Context, shopID uint64, incomeID uint64) error {
	financeTable := database.GetFinanceIncomeTableName(shopID)
	return s.db.Table(financeTable).
		Where("id = ?", incomeID).
		Update("settlement_handle_status", models.SettlementStatusCompleted).Error
}

// GetShopIncomeStats 获取店铺收入统计 - 使用分表
func (s *FinanceService) GetShopIncomeStats(ctx context.Context, adminID int64, shopID int64) (map[string]interface{}, error) {
	if _, err := s.shopService.GetShop(ctx, adminID, shopID); err != nil {
		return nil, err
	}

	financeTable := database.GetFinanceIncomeTableName(uint64(shopID))
	var totalIncome decimal.Decimal
	var unsettledIncome decimal.Decimal
	var settledIncome decimal.Decimal

	// 总收入
	s.db.Table(financeTable).
		Where("shop_id = ? AND transaction_type = ?", shopID, models.TransactionTypeEscrowVerifiedAdd).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalIncome)

	// 待结算收入
	s.db.Table(financeTable).
		Where("shop_id = ? AND transaction_type = ? AND settlement_handle_status = ?",
			shopID, models.TransactionTypeEscrowVerifiedAdd, models.SettlementStatusPending).
		Select("COALESCE(SUM(amount), 0)").Scan(&unsettledIncome)

	// 已结算收入
	s.db.Table(financeTable).
		Where("shop_id = ? AND transaction_type = ? AND settlement_handle_status = ?",
			shopID, models.TransactionTypeEscrowVerifiedAdd, models.SettlementStatusCompleted).
		Select("COALESCE(SUM(amount), 0)").Scan(&settledIncome)

	return map[string]interface{}{
		"total_income":      totalIncome,
		"unsettled_income":  unsettledIncome,
		"settled_income":    settledIncome,
	}, nil
}

func (s *FinanceService) getAccessToken(ctx context.Context, shopID uint64) (string, error) {
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
