package ratelimit

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
)

var (
	initOnce sync.Once
	initErr  error
)

// Init 初始化 Sentinel
func Init() error {
	initOnce.Do(func() {
		conf := config.NewDefaultConfig()
		conf.Sentinel.Log.Logger = &sentinelLogger{}
		initErr = sentinel.InitWithConfig(conf)
		if initErr != nil {
			log.Printf("[Sentinel] 初始化失败: %v", initErr)
			return
		}
		log.Println("[Sentinel] 限流器初始化成功")
	})
	return initErr
}

// sentinelLogger 自定义日志适配器
type sentinelLogger struct{}

func (l *sentinelLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (l *sentinelLogger) DebugEnabled() bool                             { return false }
func (l *sentinelLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *sentinelLogger) InfoEnabled() bool                              { return false }
func (l *sentinelLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (l *sentinelLogger) WarnEnabled() bool                              { return false }
func (l *sentinelLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Printf("[Sentinel] ERROR: %s, err=%v", msg, err)
}
func (l *sentinelLogger) ErrorEnabled() bool { return true }

// ResourceType 资源类型
type ResourceType string

const (
	ResourceTypeShopeeAPI ResourceType = "shopee_api"    // Shopee API 调用
	ResourceTypeFinance   ResourceType = "finance_sync"  // 财务同步
	ResourceTypeHTTP      ResourceType = "http_api"      // HTTP 接口
)

// LoadShopeeAPIRules 加载 Shopee API 限流规则
// 每个店铺独立限流，QPS 限制
func LoadShopeeAPIRules(shopID uint64, qps float64) error {
	resourceName := fmt.Sprintf("%s:%d", ResourceTypeShopeeAPI, shopID)
	
	rules := []*flow.Rule{
		{
			Resource:               resourceName,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Throttling, // 匀速排队
			Threshold:              qps,
			StatIntervalInMs:       1000, // 1秒统计窗口
			MaxQueueingTimeMs:      5000, // 最大排队等待5秒
		},
	}
	
	_, err := flow.LoadRules(rules)
	return err
}

// LoadFinanceSyncRules 加载财务同步限流规则
func LoadFinanceSyncRules(shopID uint64, qps float64, burst int32) error {
	resourceName := fmt.Sprintf("%s:%d", ResourceTypeFinance, shopID)
	
	rules := []*flow.Rule{
		{
			Resource:               resourceName,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject, // 直接拒绝
			Threshold:              qps,
			StatIntervalInMs:       1000,
		},
	}
	
	_, err := flow.LoadRules(rules)
	return err
}

// LoadHTTPRules 加载 HTTP 接口限流规则
func LoadHTTPRules(path string, qps float64) error {
	resourceName := fmt.Sprintf("%s:%s", ResourceTypeHTTP, path)
	
	rules := []*flow.Rule{
		{
			Resource:               resourceName,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              qps,
			StatIntervalInMs:       1000,
		},
	}
	
	_, err := flow.LoadRules(rules)
	return err
}

// Entry 获取限流入口
// 返回 entry 和 blockError，调用方需要在完成后调用 entry.Exit()
func Entry(resourceName string) (*base.SentinelEntry, *base.BlockError) {
	return sentinel.Entry(resourceName)
}

// Wait 等待直到获取到令牌或超时
// 这是一个阻塞方法，会等待直到获取到令牌
func Wait(ctx context.Context, resourceName string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			entry, blockErr := sentinel.Entry(resourceName)
			if blockErr == nil {
				entry.Exit()
				return nil
			}
			// 被限流，等待一小段时间后重试
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// WaitWithEntry 等待并返回 entry，调用方需要在完成后调用 entry.Exit()
func WaitWithEntry(ctx context.Context, resourceName string) (*base.SentinelEntry, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			entry, blockErr := sentinel.Entry(resourceName)
			if blockErr == nil {
				return entry, nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// ShopeeAPIResourceName 生成 Shopee API 资源名
func ShopeeAPIResourceName(shopID uint64) string {
	return fmt.Sprintf("%s:%d", ResourceTypeShopeeAPI, shopID)
}

// FinanceSyncResourceName 生成财务同步资源名
func FinanceSyncResourceName(shopID uint64) string {
	return fmt.Sprintf("%s:%d", ResourceTypeFinance, shopID)
}

// HTTPResourceName 生成 HTTP 接口资源名
func HTTPResourceName(path string) string {
	return fmt.Sprintf("%s:%s", ResourceTypeHTTP, path)
}
