package sync

import (
	"balance/backend/internal/utils"
	"context"
	"fmt"
	"log"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/ratelimit"
	"balance/backend/internal/shopee"

	"github.com/panjf2000/ants/v2"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FinanceSyncService 财务收入增量同步服务
type FinanceSyncService struct {
	db          *gorm.DB
	shardedDB   *database.ShardedDB
	workerCount int
	pool        *ants.Pool
	ctx         context.Context
	cancel      context.CancelFunc
	idGenerator *utils.IDGenerator
}

// NewFinanceSyncService 创建财务同步服务
func NewFinanceSyncService(workerCount int) *FinanceSyncService {
	ctx, cancel := context.WithCancel(context.Background())
	db := database.GetDB()

	// 创建 ants 协程池
	pool, err := ants.NewPool(workerCount,
		ants.WithPreAlloc(true),
		ants.WithNonblocking(false), // 队列满时阻塞等待
	)
	if err != nil {
		log.Fatalf("创建财务同步协程池失败: %v", err)
	}

	return &FinanceSyncService{
		db:          db,
		shardedDB:   database.NewShardedDB(db),
		workerCount: workerCount,
		pool:        pool,
		ctx:         ctx,
		cancel:      cancel,
		idGenerator: utils.NewIDGenerator(database.GetRedis()),
	}
}

// Start 启动同步服务
// 使用 ants 协程池后，不再需要预先启动 worker goroutine
// 任务通过 ScheduleAllShops() 提交到协程池时自动分配 goroutine 执行
func (s *FinanceSyncService) Start() {
	log.Printf("启动财务同步服务，协程池大小: %d，当前运行: %d，空闲: %d",
		s.workerCount, s.pool.Running(), s.pool.Free())
}

// Stop 停止同步服务
func (s *FinanceSyncService) Stop() {
	log.Println("停止财务同步服务...")
	s.cancel()
	s.pool.Release()
	log.Println("财务同步服务已停止")
}

// ScheduleAllShops 调度所有店铺的同步任务
func (s *FinanceSyncService) ScheduleAllShops() {
	var shops []models.Shop
	if err := s.db.Where("status = ?", models.ShopStatusEnabled).Find(&shops).Error; err != nil {
		log.Printf("获取店铺列表失败: %v", err)
		return
	}
	log.Printf("开始调度 %d 个店铺的财务同步任务", len(shops))
	for _, shop := range shops {
		// 检查同步记录状态
		var record models.ShopSyncRecord
		err := s.db.Where("shop_id = ? AND sync_type = ?", shop.ShopID, models.SyncTypeFinanceIncome).First(&record).Error

		if err == gorm.ErrRecordNotFound {
			// 创建同步记录
			ShopSyncRecordid, _ := s.idGenerator.GenerateShopSyncRecordID(s.ctx)
			record = models.ShopSyncRecord{
				ID:       uint64(ShopSyncRecordid),
				ShopID:   shop.ShopID,
				SyncType: models.SyncTypeFinanceIncome,
				Status:   models.SyncStatusEnabled,
			}
			s.db.Create(&record)
		} else if record.Status != models.SyncStatusEnabled {
			continue // 跳过禁用或暂停的店铺
		}

		// 连续失败超过10次，暂停同步
		if record.ConsecutiveFailCount >= 10 {
			s.db.Model(&record).Update("status", models.SyncStatusPaused)
			continue
		}

		// 提交任务到协程池
		task := models.SyncTask{
			ShopID:   shop.ShopID,
			SyncType: models.SyncTypeFinanceIncome,
		}
		err = s.pool.Submit(func() {
			s.processTask(task)
		})
		if err != nil {
			log.Printf("提交任务到协程池失败: %v", err)
		}
	}
}

// 任务超时时间
const taskTimeout = 5 * time.Minute

// processTask 处理同步任务（在协程池中执行）
func (s *FinanceSyncService) processTask(task models.SyncTask) {
	// 检查服务是否已停止
	select {
	case <-s.ctx.Done():
		return
	default:
	}

	// 创建任务级超时 context
	taskCtx, cancel := context.WithTimeout(s.ctx, taskTimeout)
	defer cancel()

	// 等待限流（使用 Sentinel）
	if err := s.waitForRateLimit(taskCtx, task.ShopID); err != nil {
		if taskCtx.Err() == context.DeadlineExceeded {
			log.Printf("店铺 %d 同步任务超时（限流等待阶段）", task.ShopID)
		}
		return
	}

	count, err := s.syncShopFinance(taskCtx, task.ShopID)

	// 更新同步记录
	now := time.Now()
	updates := map[string]interface{}{
		"last_sync_at":    &now,
		"last_sync_count": count,
	}

	if err != nil {
		updates["last_error"] = err.Error()
		updates["consecutive_fail_count"] = gorm.Expr("consecutive_fail_count + 1")
		log.Printf("店铺 %d 同步失败: %v", task.ShopID, err)
	} else {
		updates["last_error"] = ""
		updates["consecutive_fail_count"] = 0
		updates["total_synced_count"] = gorm.Expr("total_synced_count + ?", count)
		if count > 0 {
			log.Printf("店铺 %d 同步成功，新增 %d 条记录", task.ShopID, count)
		}
	}

	s.db.Model(&models.ShopSyncRecord{}).
		Where("shop_id = ? AND sync_type = ?", task.ShopID, models.SyncTypeFinanceIncome).
		Updates(updates)
}

// syncShopFinance 同步单个店铺的财务收入（增量）
func (s *FinanceSyncService) syncShopFinance(ctx context.Context, shopID uint64) (int, error) {
	// 获取店铺信息
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return 0, fmt.Errorf("店铺不存在")
	}

	// 获取授权信息
	var auth models.ShopAuthorization
	if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
		return 0, fmt.Errorf("店铺未授权")
	}

	// 检查Token是否过期
	if auth.IsAccessTokenExpired() {
		// 尝试刷新Token
		if err := s.refreshToken(ctx, shopID, &auth, shop.Region); err != nil {
			return 0, fmt.Errorf("Token已过期: %w", err)
		}
	}

	// 获取同步记录
	var record models.ShopSyncRecord
	s.db.Where("shop_id = ? AND sync_type = ?", shopID, models.SyncTypeFinanceIncome).First(&record)

	// 拉取交易记录
	client := shopee.NewClient(shop.Region)
	totalCount := 0
	pageNo := 1
	pageSize := 100
	maxTransactionTime := record.LastSyncTime

	for {
		// 限流等待（使用 Sentinel）
		if err := s.waitForRateLimit(ctx, shopID); err != nil {
			break
		}

		resp, err := client.GetWalletTransactionList(auth.AccessToken, shopID, pageNo, pageSize, "")
		if err != nil {
			return totalCount, fmt.Errorf("获取交易记录失败: %w", err)
		}

		if len(resp.Response.TransactionList) == 0 {
			break
		}

		newCount := 0
		for _, tx := range resp.Response.TransactionList {
			// 增量判断：跳过已同步的记录
			if tx.CreateTime <= record.LastSyncTime {
				continue
			}

			// 跳过提现记录
			if tx.TransactionType == models.TransactionTypeWithdrawalCreated ||
				tx.TransactionType == models.TransactionTypeWithdrawalCompleted {
				continue
			}

			// 使用分表
			financeTable := database.GetFinanceIncomeTableName(shopID)

			// 构建记录
			income := models.FinanceIncome{
				ShopID:                 shopID,
				TransactionID:          tx.TransactionID,
				OrderSN:                tx.OrderSN,
				RefundSN:               tx.RefundSN,
				Status:                 tx.Status,
				WalletType:             tx.WalletType,
				TransactionType:        tx.TransactionType,
				Amount:                 decimal.NewFromFloat(tx.Amount),
				CurrentBalance:         decimal.NewFromFloat(tx.CurrentBalance),
				TransactionTime:        tx.CreateTime,
				TransactionFee:         decimal.NewFromFloat(tx.TransactionFee),
				Description:            tx.Description,
				BuyerName:              tx.BuyerName,
				Reason:                 tx.Reason,
				WithdrawalID:           tx.WithdrawalID,
				WithdrawalType:         tx.WithdrawalType,
				TransactionTabType:     tx.TransactionTabType,
				MoneyFlow:              tx.MoneyFlow,
				SettlementHandleStatus: models.SettlementStatusPending,
			}

			// 使用 ON DUPLICATE KEY 避免竞态条件
			// 如果 transaction_id 已存在则忽略（DoNothing）
			result := s.db.Table(financeTable).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "transaction_id"}},
				DoNothing: true,
			}).Create(&income)

			if result.Error == nil && result.RowsAffected > 0 {
				newCount++
				totalCount++
				if tx.CreateTime > maxTransactionTime {
					maxTransactionTime = tx.CreateTime
				}
			}
		}

		// 如果本页没有新数据，说明已经同步完成
		if newCount == 0 && !resp.Response.More {
			break
		}

		if !resp.Response.More {
			break
		}

		pageNo++

		// 避免请求过快
		time.Sleep(200 * time.Millisecond)
	}

	// 更新最后同步时间
	if maxTransactionTime > record.LastSyncTime {
		s.db.Model(&record).Update("last_sync_time", maxTransactionTime)
	}

	return totalCount, nil
}

// refreshToken 刷新Token
func (s *FinanceSyncService) refreshToken(ctx context.Context, shopID uint64, auth *models.ShopAuthorization, region string) error {
	if auth.IsRefreshTokenExpired() {
		return fmt.Errorf("RefreshToken已过期，请重新授权")
	}

	client := shopee.NewClient(region)
	tokenResp, err := client.RefreshAccessToken(auth.RefreshToken, shopID)
	if err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"access_token":  tokenResp.AccessToken,
		"refresh_token": tokenResp.RefreshToken,
		"expires_at":    now.Add(time.Duration(tokenResp.ExpireIn) * time.Second),
		"updated_at":    now,
	}

	if err := s.db.Model(auth).Updates(updates).Error; err != nil {
		return err
	}

	auth.AccessToken = tokenResp.AccessToken
	return nil
}

// waitForRateLimit 等待限流通过（使用 Sentinel）
func (s *FinanceSyncService) waitForRateLimit(ctx context.Context, shopID uint64) error {
	// 加载财务同步限流规则（每秒5次请求）
	ratelimit.LoadFinanceSyncRules(shopID, 5, 10)
	resourceName := ratelimit.FinanceSyncResourceName(shopID)
	return ratelimit.Wait(ctx, resourceName)
}

// GetSyncStats 获取同步统计
func (s *FinanceSyncService) GetSyncStats() map[string]interface{} {
	var totalShops int64
	var enabledShops int64
	var pausedShops int64
	var totalSynced int64

	s.db.Model(&models.ShopSyncRecord{}).Where("sync_type = ?", models.SyncTypeFinanceIncome).Count(&totalShops)
	s.db.Model(&models.ShopSyncRecord{}).Where("sync_type = ? AND status = ?", models.SyncTypeFinanceIncome, models.SyncStatusEnabled).Count(&enabledShops)
	s.db.Model(&models.ShopSyncRecord{}).Where("sync_type = ? AND status = ?", models.SyncTypeFinanceIncome, models.SyncStatusPaused).Count(&pausedShops)
	s.db.Model(&models.ShopSyncRecord{}).Where("sync_type = ?", models.SyncTypeFinanceIncome).Select("COALESCE(SUM(total_synced_count), 0)").Scan(&totalSynced)

	return map[string]interface{}{
		"total_shops":   totalShops,
		"enabled_shops": enabledShops,
		"paused_shops":  pausedShops,
		"total_synced":  totalSynced,
		"pool_running":  s.pool.Running(),
		"pool_free":     s.pool.Free(),
	}
}
