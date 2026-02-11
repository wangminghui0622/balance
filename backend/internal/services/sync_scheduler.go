package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// SyncScheduler 同步调度器
type SyncScheduler struct {
	db           *gorm.DB
	orderService *OrderService
	shopService  *ShopService
	interval     time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
	running      bool
	mu           sync.Mutex
}

// NewSyncScheduler 创建同步调度器
func NewSyncScheduler(db *gorm.DB, orderService *OrderService, shopService *ShopService) *SyncScheduler {
	return &SyncScheduler{
		db:           db,
		orderService: orderService,
		shopService:  shopService,
		interval:     30 * time.Minute,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动定时同步
func (s *SyncScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run()
	log.Printf("[SyncScheduler] 定时同步已启动，间隔: %v", s.interval)
}

// Stop 停止定时同步
func (s *SyncScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
	log.Println("[SyncScheduler] 定时同步已停止")
}

func (s *SyncScheduler) run() {
	defer s.wg.Done()

	s.syncAllShops()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.syncAllShops()
		case <-s.stopChan:
			return
		}
	}
}

func (s *SyncScheduler) syncAllShops() {
	log.Println("[SyncScheduler] 开始同步所有店铺订单...")

	shops, err := s.shopService.GetAuthorizedShops()
	if err != nil {
		log.Printf("[SyncScheduler] 获取已授权店铺失败: %v", err)
		return
	}

	if len(shops) == 0 {
		log.Println("[SyncScheduler] 没有已授权的店铺")
		return
	}

	log.Printf("[SyncScheduler] 找到 %d 个已授权店铺", len(shops))

	ctx := context.Background()
	successCount := 0
	failCount := 0

	for _, shop := range shops {
		if !shop.SyncOrders {
			log.Printf("[SyncScheduler] 店铺 %d (%s) 未启用订单同步，跳过", shop.ShopID, shop.ShopName)
			continue
		}

		timeTo := time.Now()
		timeFrom := timeTo.Add(-7 * 24 * time.Hour)

		if err := s.orderService.SyncOrders(ctx, shop.ShopID, timeFrom, timeTo, ""); err != nil {
			log.Printf("[SyncScheduler] 店铺 %d (%s) 同步失败: %v", shop.ShopID, shop.ShopName, err)
			failCount++
		} else {
			log.Printf("[SyncScheduler] 店铺 %d (%s) 同步成功", shop.ShopID, shop.ShopName)
			successCount++
			s.shopService.UpdateLastSyncTime(shop.ShopID)
		}

		time.Sleep(2 * time.Second)
	}

	log.Printf("[SyncScheduler] 同步完成: 成功 %d, 失败 %d", successCount, failCount)
}

// SyncShopOrders 立即同步指定店铺的订单
func (s *SyncScheduler) SyncShopOrders(shopID uint64) error {
	log.Printf("[SyncScheduler] 开始同步店铺 %d 的订单...", shopID)

	ctx := context.Background()
	timeTo := time.Now()
	timeFrom := timeTo.Add(-30 * 24 * time.Hour)

	if err := s.orderService.SyncOrders(ctx, shopID, timeFrom, timeTo, ""); err != nil {
		return fmt.Errorf("同步店铺 %d 订单失败: %w", shopID, err)
	}

	s.shopService.UpdateLastSyncTime(shopID)

	log.Printf("[SyncScheduler] 店铺 %d 订单同步完成", shopID)
	return nil
}
