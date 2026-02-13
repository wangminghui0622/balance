package sync

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/shopee"

	"github.com/shopspring/decimal"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

// FinanceSyncService 财务收入增量同步服务
type FinanceSyncService struct {
	db           *gorm.DB
	shardedDB    *database.ShardedDB
	limiters     sync.Map // map[uint64]*rate.Limiter
	workerCount  int
	taskChan     chan models.SyncTask
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewFinanceSyncService 创建财务同步服务
func NewFinanceSyncService(workerCount int) *FinanceSyncService {
	ctx, cancel := context.WithCancel(context.Background())
	db := database.GetDB()
	return &FinanceSyncService{
		db:          db,
		shardedDB:   database.NewShardedDB(db),
		workerCount: workerCount,
		taskChan:    make(chan models.SyncTask, 1000),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动同步服务
func (s *FinanceSyncService) Start() {
	log.Printf("启动财务同步服务，Worker数量: %d", s.workerCount)
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
}

// Stop 停止同步服务
func (s *FinanceSyncService) Stop() {
	log.Println("停止财务同步服务...")
	s.cancel()
	close(s.taskChan)
	s.wg.Wait()
	log.Println("财务同步服务已停止")
}

// ScheduleAllShops 调度所有店铺的同步任务
func (s *FinanceSyncService) ScheduleAllShops() {
	var shops []models.Shop
	if err := s.db.Where("status = ?", 1).Find(&shops).Error; err != nil {
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
			record = models.ShopSyncRecord{
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

		// 添加到任务队列
		select {
		case s.taskChan <- models.SyncTask{
			ShopID:   shop.ShopID,
			SyncType: models.SyncTypeFinanceIncome,
		}:
		default:
			log.Printf("任务队列已满，跳过店铺 %d", shop.ShopID)
		}
	}
}

// worker 工作协程
func (s *FinanceSyncService) worker(id int) {
	defer s.wg.Done()
	log.Printf("Worker %d 启动", id)

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("Worker %d 退出", id)
			return
		case task, ok := <-s.taskChan:
			if !ok {
				return
			}
			s.processTask(task)
		}
	}
}

// processTask 处理同步任务
func (s *FinanceSyncService) processTask(task models.SyncTask) {
	limiter := s.getShopLimiter(task.ShopID)
	
	// 等待限流
	if err := limiter.Wait(s.ctx); err != nil {
		return
	}

	count, err := s.syncShopFinance(s.ctx, task.ShopID)
	
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
		// 限流等待
		limiter := s.getShopLimiter(shopID)
		if err := limiter.Wait(ctx); err != nil {
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

			// 检查是否已存在 - 使用分表
			financeTable := database.GetFinanceIncomeTableName(shopID)
			var existing models.FinanceIncome
			if err := s.db.Table(financeTable).Where("transaction_id = ?", tx.TransactionID).First(&existing).Error; err == nil {
				continue
			}

			// 保存记录
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

			if err := s.db.Table(financeTable).Create(&income).Error; err == nil {
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

// getShopLimiter 获取店铺限流器
func (s *FinanceSyncService) getShopLimiter(shopID uint64) *rate.Limiter {
	if limiter, ok := s.limiters.Load(shopID); ok {
		return limiter.(*rate.Limiter)
	}
	// 每秒5次请求，突发10次
	limiter := rate.NewLimiter(5, 10)
	s.limiters.Store(shopID, limiter)
	return limiter
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
		"queue_size":    len(s.taskChan),
	}
}
