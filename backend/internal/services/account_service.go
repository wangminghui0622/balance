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

// AccountService 账户服务
type AccountService struct {
	db        *gorm.DB
	shardedDB *database.ShardedDB
}

// NewAccountService 创建账户服务
func NewAccountService() *AccountService {
	db := database.GetDB()
	return &AccountService{
		db:        db,
		shardedDB: database.NewShardedDB(db),
	}
}

// GenerateTransactionNo 生成流水号
func (s *AccountService) GenerateTransactionNo(accountType string) string {
	return fmt.Sprintf("%s%d%d", accountType[:3], time.Now().UnixNano(), time.Now().UnixMicro()%1000)
}

// createTransaction 创建账户流水（使用分表）
func (s *AccountService) createTransaction(tx *gorm.DB, at *models.AccountTransaction) error {
	txTable := database.GetAccountTransactionTableName(at.AdminID)
	return tx.Table(txTable).Create(at).Error
}

// ==================== 预付款账户 ====================

// GetOrCreatePrepaymentAccount 获取或创建预付款账户
func (s *AccountService) GetOrCreatePrepaymentAccount(ctx context.Context, adminID int64) (*models.PrepaymentAccount, error) {
	var account models.PrepaymentAccount
	err := s.db.Where("admin_id = ?", adminID).First(&account).Error
	if err == nil {
		return &account, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 创建新账户
	account = models.PrepaymentAccount{
		AdminID:  adminID,
		Balance:  decimal.Zero,
		Currency: "TWD",
		Status:   models.AccountStatusNormal,
	}
	if err := s.db.Create(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// GetPrepaymentBalance 获取预付款余额
func (s *AccountService) GetPrepaymentBalance(ctx context.Context, adminID int64) (decimal.Decimal, decimal.Decimal, error) {
	account, err := s.GetOrCreatePrepaymentAccount(ctx, adminID)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	return account.Balance, account.FrozenAmount, nil
}

// RechargePrepayment 预付款充值
func (s *AccountService) RechargePrepayment(ctx context.Context, adminID int64, amount decimal.Decimal, remark string, operatorID int64) (*models.AccountTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("充值金额必须大于0")
	}

	var tx *models.AccountTransaction
	err := s.db.Transaction(func(db *gorm.DB) error {
		var account models.PrepaymentAccount
		// 使用 FOR UPDATE 行锁防止并发更新
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				account = models.PrepaymentAccount{
					AdminID:  adminID,
					Currency: "TWD",
					Status:   models.AccountStatusNormal,
				}
				if err := db.Create(&account).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		balanceBefore := account.Balance
		account.Balance = account.Balance.Add(amount)
		account.TotalRecharge = account.TotalRecharge.Add(amount)

		if err := db.Save(&account).Error; err != nil {
			return err
		}

		// 记录流水
		tx = &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(models.AccountTypePrepayment),
			AccountType:     models.AccountTypePrepayment,
			AdminID:         adminID,
			TransactionType: models.TxTypeRecharge,
			Amount:          amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    account.Balance,
			Remark:          remark,
			OperatorID:      operatorID,
		}
		return s.createTransaction(db, tx)
	})

	return tx, err
}

// FreezePrepayment 冻结预付款 (发货时调用)
func (s *AccountService) FreezePrepayment(ctx context.Context, adminID int64, amount decimal.Decimal, orderSN string, remark string) (*models.AccountTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("冻结金额必须大于0")
	}

	var tx *models.AccountTransaction
	err := s.db.Transaction(func(db *gorm.DB) error {
		var account models.PrepaymentAccount
		// 使用 FOR UPDATE 行锁防止并发更新
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			return fmt.Errorf("预付款账户不存在")
		}

		if account.Status != models.AccountStatusNormal {
			return fmt.Errorf("预付款账户已冻结")
		}

		if account.Balance.LessThan(amount) {
			return fmt.Errorf("预付款余额不足，当前余额: %s, 需要: %s", account.Balance.String(), amount.String())
		}

		balanceBefore := account.Balance
		account.Balance = account.Balance.Sub(amount)
		account.FrozenAmount = account.FrozenAmount.Add(amount)

		if err := db.Save(&account).Error; err != nil {
			return err
		}

		// 记录流水
		tx = &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(models.AccountTypePrepayment),
			AccountType:     models.AccountTypePrepayment,
			AdminID:         adminID,
			TransactionType: models.TxTypeFreeze,
			Amount:          amount.Neg(),
			BalanceBefore:   balanceBefore,
			BalanceAfter:    account.Balance,
			RelatedOrderSN:  orderSN,
			Remark:          remark,
		}
		return s.createTransaction(db, tx)
	})

	return tx, err
}

// UnfreezePrepayment 解冻预付款 (订单取消时调用)
func (s *AccountService) UnfreezePrepayment(ctx context.Context, adminID int64, amount decimal.Decimal, orderSN string, remark string) (*models.AccountTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("解冻金额必须大于0")
	}

	var tx *models.AccountTransaction
	err := s.db.Transaction(func(db *gorm.DB) error {
		var account models.PrepaymentAccount
		// 使用 FOR UPDATE 行锁防止并发更新
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			return fmt.Errorf("预付款账户不存在")
		}

		if account.FrozenAmount.LessThan(amount) {
			return fmt.Errorf("冻结金额不足")
		}

		balanceBefore := account.Balance
		account.Balance = account.Balance.Add(amount)
		account.FrozenAmount = account.FrozenAmount.Sub(amount)

		if err := db.Save(&account).Error; err != nil {
			return err
		}

		// 记录流水
		tx = &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(models.AccountTypePrepayment),
			AccountType:     models.AccountTypePrepayment,
			AdminID:         adminID,
			TransactionType: models.TxTypeUnfreeze,
			Amount:          amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    account.Balance,
			RelatedOrderSN:  orderSN,
			Remark:          remark,
		}
		return s.createTransaction(db, tx)
	})

	return tx, err
}

// SettlePrepayment 结算预付款 (订单完成时，从冻结金额扣除)
func (s *AccountService) SettlePrepayment(ctx context.Context, adminID int64, amount decimal.Decimal, orderSN string, remark string) (*models.AccountTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("结算金额必须大于0")
	}

	var at *models.AccountTransaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var account models.PrepaymentAccount
		// 使用 FOR UPDATE 行锁防止并发更新
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			return fmt.Errorf("预付款账户不存在")
		}

		if account.FrozenAmount.LessThan(amount) {
			return fmt.Errorf("冻结金额不足")
		}

		balanceBefore := account.Balance
		account.FrozenAmount = account.FrozenAmount.Sub(amount)
		account.TotalConsume = account.TotalConsume.Add(amount)

		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		// 记录流水
		at = &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(models.AccountTypePrepayment),
			AccountType:     models.AccountTypePrepayment,
			AdminID:         adminID,
			TransactionType: models.TxTypeOrderPay,
			Amount:          amount.Neg(),
			BalanceBefore:   balanceBefore,
			BalanceAfter:    account.Balance,
			RelatedOrderSN:  orderSN,
			Remark:          remark,
		}
		return s.createTransaction(tx, at)
	})

	return at, err
}

// ==================== 运营账户 ====================

// GetOrCreateOperatorAccount 获取或创建运营账户
func (s *AccountService) GetOrCreateOperatorAccount(ctx context.Context, adminID int64) (*models.OperatorAccount, error) {
	var account models.OperatorAccount
	err := s.db.Where("admin_id = ?", adminID).First(&account).Error
	if err == nil {
		return &account, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 创建新账户
	account = models.OperatorAccount{
		AdminID:  adminID,
		Balance:  decimal.Zero,
		Currency: "TWD",
		Status:   models.AccountStatusNormal,
	}
	if err := s.db.Create(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// AddOperatorIncome 增加运营收入 (结算时调用)
func (s *AccountService) AddOperatorIncome(ctx context.Context, adminID int64, amount decimal.Decimal, orderSN string, remark string) (*models.AccountTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("收入金额必须大于0")
	}

	var at *models.AccountTransaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var account models.OperatorAccount
		// 使用 FOR UPDATE 行锁防止并发更新
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				account = models.OperatorAccount{
					AdminID:  adminID,
					Currency: "TWD",
					Status:   models.AccountStatusNormal,
				}
				if err := tx.Create(&account).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		balanceBefore := account.Balance
		account.Balance = account.Balance.Add(amount)
		account.TotalEarnings = account.TotalEarnings.Add(amount)

		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		// 记录流水
		at = &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(models.AccountTypeOperator),
			AccountType:     models.AccountTypeOperator,
			AdminID:         adminID,
			TransactionType: models.TxTypeCostSettle,
			Amount:          amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    account.Balance,
			RelatedOrderSN:  orderSN,
			Remark:          remark,
		}
		return s.createTransaction(tx, at)
	})

	return at, err
}

// DeductOperatorBalance 扣除运营账户余额 (调账扣款时调用)
func (s *AccountService) DeductOperatorBalance(ctx context.Context, adminID int64, amount decimal.Decimal, orderSN string, remark string) (*models.AccountTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("扣款金额必须大于0")
	}

	var tx *models.AccountTransaction
	err := s.db.Transaction(func(db *gorm.DB) error {
		var account models.OperatorAccount
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			return fmt.Errorf("运营账户不存在")
		}

		if account.Balance.LessThan(amount) {
			return fmt.Errorf("运营账户余额不足")
		}

		balanceBefore := account.Balance
		account.Balance = account.Balance.Sub(amount)

		if err := db.Save(&account).Error; err != nil {
			return err
		}

		tx = &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(models.AccountTypeOperator),
			AccountType:     models.AccountTypeOperator,
			AdminID:         adminID,
			TransactionType: models.TxTypeAdjustment,
			Amount:          amount.Neg(),
			BalanceBefore:   balanceBefore,
			BalanceAfter:    account.Balance,
			RelatedOrderSN:  orderSN,
			Remark:          remark,
		}
		return s.createTransaction(db, tx)
	})

	return tx, err
}

// ==================== 保证金账户 ====================

// GetOrCreateDepositAccount 获取或创建保证金账户
func (s *AccountService) GetOrCreateDepositAccount(ctx context.Context, adminID int64) (*models.DepositAccount, error) {
	var account models.DepositAccount
	err := s.db.Where("admin_id = ?", adminID).First(&account).Error
	if err == nil {
		return &account, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 创建新账户
	account = models.DepositAccount{
		AdminID:  adminID,
		Balance:  decimal.Zero,
		Currency: "TWD",
		Status:   models.AccountStatusNormal,
	}
	if err := s.db.Create(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// PayDeposit 缴纳保证金
func (s *AccountService) PayDeposit(ctx context.Context, adminID int64, amount decimal.Decimal, remark string, operatorID int64) (*models.AccountTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("保证金金额必须大于0")
	}

	var tx *models.AccountTransaction
	err := s.db.Transaction(func(db *gorm.DB) error {
		var account models.DepositAccount
		// 使用 FOR UPDATE 行锁防止并发更新
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				account = models.DepositAccount{
					AdminID:  adminID,
					Currency: "TWD",
					Status:   models.AccountStatusNormal,
				}
				if err := db.Create(&account).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		balanceBefore := account.Balance
		account.Balance = account.Balance.Add(amount)

		// 检查是否达到应缴金额
		if account.Balance.GreaterThanOrEqual(account.RequiredAmount) {
			account.Status = models.AccountStatusNormal
		}

		if err := db.Save(&account).Error; err != nil {
			return err
		}

		// 记录流水
		tx = &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(models.AccountTypeDeposit),
			AccountType:     models.AccountTypeDeposit,
			AdminID:         adminID,
			TransactionType: models.TxTypeDepositPay,
			Amount:          amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    account.Balance,
			Remark:          remark,
			OperatorID:      operatorID,
		}
		return s.createTransaction(db, tx)
	})

	return tx, err
}

// ==================== 查询 ====================

// GetAccountTransactions 获取账户流水 - 使用分表
func (s *AccountService) GetAccountTransactions(ctx context.Context, accountType string, adminID int64, page, pageSize int) ([]models.AccountTransaction, int64, error) {
	var transactions []models.AccountTransaction
	var total int64

	txTable := database.GetAccountTransactionTableName(adminID)
	query := s.db.Table(txTable).Where("account_type = ? AND admin_id = ?", accountType, adminID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&transactions).Error
	return transactions, total, err
}

// ==================== 店主佣金账户 ====================

// GetOrCreateShopOwnerCommissionAccount 获取或创建店主佣金账户
func (s *AccountService) GetOrCreateShopOwnerCommissionAccount(ctx context.Context, adminID int64) (*models.ShopOwnerCommissionAccount, error) {
	var account models.ShopOwnerCommissionAccount
	err := s.db.Where("admin_id = ?", adminID).First(&account).Error
	if err == nil {
		return &account, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	account = models.ShopOwnerCommissionAccount{
		AdminID:  adminID,
		Balance:  decimal.Zero,
		Currency: "TWD",
		Status:   models.AccountStatusNormal,
	}
	if err := s.db.Create(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// AddShopOwnerCommission 增加店主佣金 (结算时调用)
func (s *AccountService) AddShopOwnerCommission(ctx context.Context, adminID int64, amount decimal.Decimal, orderSN string, remark string) (*models.AccountTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("佣金金额必须大于0")
	}

	var at *models.AccountTransaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var account models.ShopOwnerCommissionAccount
		// 使用 FOR UPDATE 行锁防止并发更新
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				account = models.ShopOwnerCommissionAccount{
					AdminID:  adminID,
					Currency: "TWD",
					Status:   models.AccountStatusNormal,
				}
				if err := tx.Create(&account).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		balanceBefore := account.Balance
		account.Balance = account.Balance.Add(amount)
		account.TotalEarnings = account.TotalEarnings.Add(amount)

		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		at = &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(models.AccountTypeShopOwnerCommission),
			AccountType:     models.AccountTypeShopOwnerCommission,
			AdminID:         adminID,
			TransactionType: models.TxTypeProfitShare,
			Amount:          amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    account.Balance,
			RelatedOrderSN:  orderSN,
			Remark:          remark,
		}
		return s.createTransaction(tx, at)
	})

	return at, err
}

// DeductShopOwnerCommission 扣除店主佣金 (调账扣款时调用)
func (s *AccountService) DeductShopOwnerCommission(ctx context.Context, adminID int64, amount decimal.Decimal, orderSN string, remark string) (*models.AccountTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("扣款金额必须大于0")
	}

	var tx *models.AccountTransaction
	err := s.db.Transaction(func(db *gorm.DB) error {
		var account models.ShopOwnerCommissionAccount
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			return fmt.Errorf("店主佣金账户不存在")
		}

		if account.Balance.LessThan(amount) {
			return fmt.Errorf("店主佣金账户余额不足")
		}

		balanceBefore := account.Balance
		account.Balance = account.Balance.Sub(amount)

		if err := db.Save(&account).Error; err != nil {
			return err
		}

		tx = &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(models.AccountTypeShopOwnerCommission),
			AccountType:     models.AccountTypeShopOwnerCommission,
			AdminID:         adminID,
			TransactionType: models.TxTypeAdjustment,
			Amount:          amount.Neg(),
			BalanceBefore:   balanceBefore,
			BalanceAfter:    account.Balance,
			RelatedOrderSN:  orderSN,
			Remark:          remark,
		}
		return s.createTransaction(db, tx)
	})

	return tx, err
}

// ==================== 平台佣金账户 ====================

// GetOrCreatePlatformCommissionAccount 获取或创建平台佣金账户 (单例)
func (s *AccountService) GetOrCreatePlatformCommissionAccount(ctx context.Context) (*models.PlatformCommissionAccount, error) {
	var account models.PlatformCommissionAccount
	err := s.db.First(&account).Error
	if err == nil {
		return &account, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	account = models.PlatformCommissionAccount{
		Balance:  decimal.Zero,
		Currency: "TWD",
		Status:   models.AccountStatusNormal,
	}
	if err := s.db.Create(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// AddPlatformCommission 增加平台佣金 (结算时调用)
func (s *AccountService) AddPlatformCommission(ctx context.Context, amount decimal.Decimal, orderSN string, remark string) (*models.AccountTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("佣金金额必须大于0")
	}

	var at *models.AccountTransaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var account models.PlatformCommissionAccount
		// 使用 FOR UPDATE 行锁防止并发更新
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&account).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				account = models.PlatformCommissionAccount{
					Currency: "TWD",
					Status:   models.AccountStatusNormal,
				}
				if err := tx.Create(&account).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		balanceBefore := account.Balance
		account.Balance = account.Balance.Add(amount)
		account.TotalEarnings = account.TotalEarnings.Add(amount)

		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		at = &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(models.AccountTypePlatformCommission),
			AccountType:     models.AccountTypePlatformCommission,
			AdminID:         0, // 平台账户无 AdminID
			TransactionType: models.TxTypePlatformFee,
			Amount:          amount,
			BalanceBefore:   balanceBefore,
			BalanceAfter:    account.Balance,
			RelatedOrderSN:  orderSN,
			Remark:          remark,
		}
		return s.createTransaction(tx, at)
	})

	return at, err
}

// DeductPlatformCommission 扣除平台佣金 (调账扣款时调用)
func (s *AccountService) DeductPlatformCommission(ctx context.Context, amount decimal.Decimal, orderSN string, remark string) (*models.AccountTransaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("扣款金额必须大于0")
	}

	var tx *models.AccountTransaction
	err := s.db.Transaction(func(db *gorm.DB) error {
		var account models.PlatformCommissionAccount
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&account).Error; err != nil {
			return fmt.Errorf("平台佣金账户不存在")
		}

		if account.Balance.LessThan(amount) {
			return fmt.Errorf("平台佣金账户余额不足")
		}

		balanceBefore := account.Balance
		account.Balance = account.Balance.Sub(amount)

		if err := db.Save(&account).Error; err != nil {
			return err
		}

		tx = &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(models.AccountTypePlatformCommission),
			AccountType:     models.AccountTypePlatformCommission,
			AdminID:         0, // 平台账户无 AdminID
			TransactionType: models.TxTypeAdjustment,
			Amount:          amount.Neg(),
			BalanceBefore:   balanceBefore,
			BalanceAfter:    account.Balance,
			RelatedOrderSN:  orderSN,
			Remark:          remark,
		}
		return s.createTransaction(db, tx)
	})

	return tx, err
}

// ==================== 托管账户 ====================

// GetOrCreateEscrowAccount 获取或创建托管账户
func (s *AccountService) GetOrCreateEscrowAccount(ctx context.Context, adminID int64) (*models.EscrowAccount, error) {
	var account models.EscrowAccount
	err := s.db.Where("admin_id = ?", adminID).First(&account).Error
	if err == nil {
		return &account, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	account = models.EscrowAccount{
		AdminID:  adminID,
		Balance:  decimal.Zero,
		Currency: "TWD",
		Status:   models.AccountStatusNormal,
	}
	if err := s.db.Create(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// TransferToEscrow 转入托管账户 (发货冻结时调用)
func (s *AccountService) TransferToEscrow(ctx context.Context, adminID int64, amount decimal.Decimal, orderSN string, remark string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("托管金额必须大于0")
	}

	return s.db.Transaction(func(db *gorm.DB) error {
		var account models.EscrowAccount
		// 使用 FOR UPDATE 行锁防止并发更新
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				account = models.EscrowAccount{
					AdminID:  adminID,
					Currency: "TWD",
					Status:   models.AccountStatusNormal,
				}
				if err := db.Create(&account).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		account.Balance = account.Balance.Add(amount)
		account.TotalIn = account.TotalIn.Add(amount)
		return db.Save(&account).Error
	})
}

// TransferFromEscrow 从托管账户转出 (结算时调用)
func (s *AccountService) TransferFromEscrow(ctx context.Context, adminID int64, amount decimal.Decimal, orderSN string, remark string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("转出金额必须大于0")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var account models.EscrowAccount
		// 使用 FOR UPDATE 行锁防止并发更新
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
			return fmt.Errorf("托管账户不存在")
		}

		if account.Balance.LessThan(amount) {
			return fmt.Errorf("托管账户余额不足")
		}

		account.Balance = account.Balance.Sub(amount)
		account.TotalOut = account.TotalOut.Add(amount)
		return tx.Save(&account).Error
	})
}

// ==================== 提现功能 ====================

// GenerateApplicationNo 生成申请单号
func (s *AccountService) GenerateApplicationNo(prefix string) string {
	return fmt.Sprintf("%s%d%d", prefix, time.Now().UnixNano(), time.Now().UnixMicro()%1000)
}

// ApplyWithdraw 申请提现
func (s *AccountService) ApplyWithdraw(ctx context.Context, adminID int64, accountType string, amount decimal.Decimal, collectionAccountID uint64, remark string) (*models.WithdrawApplication, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("提现金额必须大于0")
	}

	// 检查余额
	var availableBalance decimal.Decimal
	switch accountType {
	case models.AccountTypeOperator:
		account, err := s.GetOrCreateOperatorAccount(ctx, adminID)
		if err != nil {
			return nil, err
		}
		availableBalance = account.Balance
	case models.AccountTypeShopOwnerCommission:
		account, err := s.GetOrCreateShopOwnerCommissionAccount(ctx, adminID)
		if err != nil {
			return nil, err
		}
		availableBalance = account.Balance
	case models.AccountTypeDeposit:
		account, err := s.GetOrCreateDepositAccount(ctx, adminID)
		if err != nil {
			return nil, err
		}
		// 保证金提现：只能提取超出15000的部分
		minDeposit := decimal.NewFromInt(15000)
		if account.Balance.LessThanOrEqual(minDeposit) {
			return nil, fmt.Errorf("保证金余额不足，最低需保留 %s TWD", minDeposit.String())
		}
		availableBalance = account.Balance.Sub(minDeposit)
	default:
		return nil, fmt.Errorf("不支持的账户类型: %s", accountType)
	}

	if availableBalance.LessThan(amount) {
		return nil, fmt.Errorf("可提现余额不足，当前可提现: %s", availableBalance.String())
	}

	// 检查收款账户
	var collectionAccount models.CollectionAccount
	if err := s.db.Where("id = ? AND admin_id = ?", collectionAccountID, adminID).First(&collectionAccount).Error; err != nil {
		return nil, fmt.Errorf("收款账户不存在")
	}

	// 创建提现申请
	application := &models.WithdrawApplication{
		ApplicationNo:       s.GenerateApplicationNo("WD"),
		AdminID:             adminID,
		AccountType:         accountType,
		Amount:              amount,
		Fee:                 decimal.Zero, // 暂不收取手续费
		ActualAmount:        amount,
		Currency:            "TWD",
		CollectionAccountID: collectionAccountID,
		Status:              models.ApplicationStatusPending,
		Remark:              remark,
	}

	if err := s.db.Create(application).Error; err != nil {
		return nil, err
	}

	// 冻结提现金额
	err := s.freezeForWithdraw(ctx, adminID, accountType, amount)
	if err != nil {
		s.db.Delete(application)
		return nil, err
	}

	return application, nil
}

// freezeForWithdraw 冻结提现金额
func (s *AccountService) freezeForWithdraw(ctx context.Context, adminID int64, accountType string, amount decimal.Decimal) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		switch accountType {
		case models.AccountTypeOperator:
			var account models.OperatorAccount
			// 使用 FOR UPDATE 行锁防止并发更新
			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
				return err
			}
			if account.Balance.LessThan(amount) {
				return fmt.Errorf("余额不足")
			}
			account.Balance = account.Balance.Sub(amount)
			account.FrozenAmount = account.FrozenAmount.Add(amount)
			return db.Save(&account).Error

		case models.AccountTypeShopOwnerCommission:
			var account models.ShopOwnerCommissionAccount
			// 使用 FOR UPDATE 行锁防止并发更新
			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
				return err
			}
			if account.Balance.LessThan(amount) {
				return fmt.Errorf("余额不足")
			}
			account.Balance = account.Balance.Sub(amount)
			account.FrozenAmount = account.FrozenAmount.Add(amount)
			return db.Save(&account).Error

		case models.AccountTypeDeposit:
			var account models.DepositAccount
			// 使用 FOR UPDATE 行锁防止并发更新
			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
				return err
			}
			minDeposit := decimal.NewFromInt(15000)
			if account.Balance.Sub(amount).LessThan(minDeposit) {
				return fmt.Errorf("保证金余额不足")
			}
			account.Balance = account.Balance.Sub(amount)
			return db.Save(&account).Error
		}
		return nil
	})
}

// ApproveWithdraw 审批通过提现
func (s *AccountService) ApproveWithdraw(ctx context.Context, applicationID uint64, auditBy int64, auditRemark string) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		var application models.WithdrawApplication
		// 使用 FOR UPDATE 行锁防止并发审批
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", applicationID, models.ApplicationStatusPending).First(&application).Error; err != nil {
			return fmt.Errorf("提现申请不存在或已处理")
		}

		now := time.Now()
		application.Status = models.ApplicationStatusApproved
		application.AuditBy = auditBy
		application.AuditAt = &now
		application.AuditRemark = auditRemark

		return db.Save(&application).Error
	})
}

// RejectWithdraw 拒绝提现
func (s *AccountService) RejectWithdraw(ctx context.Context, applicationID uint64, auditBy int64, auditRemark string) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		var application models.WithdrawApplication
		// 使用 FOR UPDATE 行锁防止并发审批
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", applicationID, models.ApplicationStatusPending).First(&application).Error; err != nil {
			return fmt.Errorf("提现申请不存在或已处理")
		}

		// 解冻金额
		if err := s.unfreezeForWithdraw(ctx, application.AdminID, application.AccountType, application.Amount); err != nil {
			return err
		}

		now := time.Now()
		application.Status = models.ApplicationStatusRejected
		application.AuditBy = auditBy
		application.AuditAt = &now
		application.AuditRemark = auditRemark

		return db.Save(&application).Error
	})
}

// unfreezeForWithdraw 解冻提现金额
func (s *AccountService) unfreezeForWithdraw(ctx context.Context, adminID int64, accountType string, amount decimal.Decimal) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		switch accountType {
		case models.AccountTypeOperator:
			var account models.OperatorAccount
			// 使用 FOR UPDATE 行锁防止并发更新
			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
				return err
			}
			account.Balance = account.Balance.Add(amount)
			account.FrozenAmount = account.FrozenAmount.Sub(amount)
			return db.Save(&account).Error

		case models.AccountTypeShopOwnerCommission:
			var account models.ShopOwnerCommissionAccount
			// 使用 FOR UPDATE 行锁防止并发更新
			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
				return err
			}
			account.Balance = account.Balance.Add(amount)
			account.FrozenAmount = account.FrozenAmount.Sub(amount)
			return db.Save(&account).Error

		case models.AccountTypeDeposit:
			var account models.DepositAccount
			// 使用 FOR UPDATE 行锁防止并发更新
			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
				return err
			}
			account.Balance = account.Balance.Add(amount)
			return db.Save(&account).Error
		}
		return nil
	})
}

// ConfirmWithdrawPaid 确认提现已打款
func (s *AccountService) ConfirmWithdrawPaid(ctx context.Context, applicationID uint64, operatorID int64) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		var application models.WithdrawApplication
		// 使用 FOR UPDATE 行锁防止并发操作
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", applicationID, models.ApplicationStatusApproved).First(&application).Error; err != nil {
			return fmt.Errorf("提现申请不存在或状态不正确")
		}

		// 从冻结金额中扣除并记录流水
		if err := s.completeWithdraw(ctx, application.AdminID, application.AccountType, application.Amount, application.ApplicationNo); err != nil {
			return err
		}

		now := time.Now()
		application.Status = models.ApplicationStatusPaid
		application.PaidAt = &now

		return db.Save(&application).Error
	})
}

// completeWithdraw 完成提现
func (s *AccountService) completeWithdraw(ctx context.Context, adminID int64, accountType string, amount decimal.Decimal, applicationNo string) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		var balanceBefore decimal.Decimal

		switch accountType {
		case models.AccountTypeOperator:
			var account models.OperatorAccount
			// 使用 FOR UPDATE 行锁防止并发更新
			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
				return err
			}
			balanceBefore = account.Balance
			account.FrozenAmount = account.FrozenAmount.Sub(amount)
			account.TotalWithdrawn = account.TotalWithdrawn.Add(amount)
			if err := db.Save(&account).Error; err != nil {
				return err
			}

		case models.AccountTypeShopOwnerCommission:
			var account models.ShopOwnerCommissionAccount
			// 使用 FOR UPDATE 行锁防止并发更新
			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
				return err
			}
			balanceBefore = account.Balance
			account.FrozenAmount = account.FrozenAmount.Sub(amount)
			account.TotalWithdrawn = account.TotalWithdrawn.Add(amount)
			if err := db.Save(&account).Error; err != nil {
				return err
			}

		case models.AccountTypeDeposit:
			var account models.DepositAccount
			// 使用 FOR UPDATE 行锁防止并发更新
			if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("admin_id = ?", adminID).First(&account).Error; err != nil {
				return err
			}
			balanceBefore = account.Balance
			// 保证金已在申请时扣除，无需再扣
		}

		// 记录流水
		tx := &models.AccountTransaction{
			TransactionNo:   s.GenerateTransactionNo(accountType),
			AccountType:     accountType,
			AdminID:         adminID,
			TransactionType: models.TxTypeWithdraw,
			Amount:          amount.Neg(),
			BalanceBefore:   balanceBefore,
			BalanceAfter:    balanceBefore, // 提现不影响可用余额（已冻结）
			Remark:          fmt.Sprintf("提现申请: %s", applicationNo),
		}
		return s.createTransaction(db, tx)
	})
}

// GetWithdrawApplications 获取提现申请列表
func (s *AccountService) GetWithdrawApplications(ctx context.Context, adminID int64, status int8, page, pageSize int) ([]models.WithdrawApplication, int64, error) {
	var applications []models.WithdrawApplication
	var total int64

	query := s.db.Model(&models.WithdrawApplication{})
	if adminID > 0 {
		query = query.Where("admin_id = ?", adminID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&applications).Error
	return applications, total, err
}

// ==================== 充值申请功能 ====================

// ApplyRecharge 申请充值 (线下充值)
func (s *AccountService) ApplyRecharge(ctx context.Context, adminID int64, accountType string, amount decimal.Decimal, paymentMethod string, paymentProof string, remark string) (*models.RechargeApplication, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("充值金额必须大于0")
	}

	// 验证账户类型
	if accountType != models.AccountTypePrepayment && accountType != models.AccountTypeDeposit {
		return nil, fmt.Errorf("只能充值预付款账户或保证金账户")
	}

	application := &models.RechargeApplication{
		ApplicationNo: s.GenerateApplicationNo("RC"),
		AdminID:       adminID,
		AccountType:   accountType,
		Amount:        amount,
		Currency:      "TWD",
		PaymentMethod: paymentMethod,
		PaymentProof:  paymentProof,
		Status:        models.ApplicationStatusPending,
		Remark:        remark,
	}

	if err := s.db.Create(application).Error; err != nil {
		return nil, err
	}

	return application, nil
}

// ApproveRecharge 审批通过充值
func (s *AccountService) ApproveRecharge(ctx context.Context, applicationID uint64, auditBy int64, auditRemark string) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		var application models.RechargeApplication
		// 使用 FOR UPDATE 行锁防止并发审批
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", applicationID, models.ApplicationStatusPending).First(&application).Error; err != nil {
			return fmt.Errorf("充值申请不存在或已处理")
		}

		// 执行充值
		var err error
		switch application.AccountType {
		case models.AccountTypePrepayment:
			_, err = s.RechargePrepayment(ctx, application.AdminID, application.Amount, fmt.Sprintf("线下充值审核通过: %s", application.ApplicationNo), auditBy)
		case models.AccountTypeDeposit:
			_, err = s.PayDeposit(ctx, application.AdminID, application.Amount, fmt.Sprintf("线下保证金缴纳审核通过: %s", application.ApplicationNo), auditBy)
		}
		if err != nil {
			return fmt.Errorf("充值失败: %w", err)
		}

		now := time.Now()
		application.Status = models.ApplicationStatusApproved
		application.AuditBy = auditBy
		application.AuditAt = &now
		application.AuditRemark = auditRemark

		return db.Save(&application).Error
	})
}

// RejectRecharge 拒绝充值
func (s *AccountService) RejectRecharge(ctx context.Context, applicationID uint64, auditBy int64, auditRemark string) error {
	return s.db.Transaction(func(db *gorm.DB) error {
		var application models.RechargeApplication
		// 使用 FOR UPDATE 行锁防止并发审批
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", applicationID, models.ApplicationStatusPending).First(&application).Error; err != nil {
			return fmt.Errorf("充值申请不存在或已处理")
		}

		now := time.Now()
		application.Status = models.ApplicationStatusRejected
		application.AuditBy = auditBy
		application.AuditAt = &now
		application.AuditRemark = auditRemark

		return db.Save(&application).Error
	})
}

// GetRechargeApplications 获取充值申请列表
func (s *AccountService) GetRechargeApplications(ctx context.Context, adminID int64, status int8, page, pageSize int) ([]models.RechargeApplication, int64, error) {
	var applications []models.RechargeApplication
	var total int64

	query := s.db.Model(&models.RechargeApplication{})
	if adminID > 0 {
		query = query.Where("admin_id = ?", adminID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&applications).Error
	return applications, total, err
}
