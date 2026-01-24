package services

import (
	"balance/internal/models"
	shareUtils "balance/internal/utils"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

// OrderService 订单服务
// 本系统本身就是商家服务器，直接处理订单
type OrderService struct {
	shopeeClient *shareUtils.ShopeeAPIClient
	db           *gorm.DB
}

// NewOrderService 创建订单服务
func NewOrderService(merchantURL string, db *gorm.DB) *OrderService {
	// merchantURL 参数保留以兼容现有代码，但不再使用
	return &OrderService{db: db}
}

// SetShopeeClient 设置虾皮API客户端
func (s *OrderService) SetShopeeClient(client *shareUtils.ShopeeAPIClient) {
	s.shopeeClient = client
}

// ReceiveOrderFromShopee 接收来自虾皮的订单推送
// xsheep 本身就是商家服务器，直接处理订单
func (s *OrderService) ReceiveOrderFromShopee(orderPush *models.OrderStatusPush) error {
	// 获取订单号
	orderSn := orderPush.Data.OrderSn
	if orderSn == "" {
		orderSn = orderPush.Data.OrderSN
	}

	log.Printf("收到虾皮订单推送: shop_id=%d, order_sn=%s, status=%s, code=%d, msg_id=%s",
		orderPush.ShopID, orderSn, orderPush.Data.Status, orderPush.Code, orderPush.MsgID)

	// 这里只处理“订单相关”的推送，其它店铺/商品类推送只记录日志后忽略
	switch orderPush.Code {
	case 3:
		// 订单状态推送（待支付、待付款、已付款、已发货、已完成、已取消等）
		return s.ProcessOrder(orderPush)
	case 29:
		// 退款/退货推送（仍然属于订单维度）
		log.Printf("收到退款/退货推送: order_sn=%s", orderSn)
		return s.ProcessReturnRefund(orderPush)
	default:
		// 其它 code（如店铺冻结、违规商品等）视为“非订单类”，这里只做记录，不做业务处理
		log.Printf("收到非订单类推送或未知类型推送, 忽略处理: code=%d, shop_id=%d, order_sn=%s",
			orderPush.Code, orderPush.ShopID, orderSn)
		return nil
	}
}

// ProcessOrder 处理订单推送（只关注订单相关状态）
func (s *OrderService) ProcessOrder(orderPush *models.OrderStatusPush) error {
	orderSn := orderPush.Data.OrderSn
	if orderSn == "" {
		orderSn = orderPush.Data.OrderSN
	}

	status := strings.ToUpper(orderPush.Data.Status)

	log.Printf("处理订单状态变更: order_sn=%s, raw_status=%s, normalized_status=%s, shop_id=%d",
		orderSn, orderPush.Data.Status, status, orderPush.ShopID)

	// 按照 Shopee 订单生命周期把所有订单相关状态分支写全
	switch status {
	case "UNPAID":
		// 未付款：买家下单但还未付款
		log.Printf("[订单状态] 未付款 UNPAID: order_sn=%s", orderSn)
		// TODO: 在这里处理“未付款”逻辑（如创建预订单、占库存等）

	case "READY_TO_SHIP":
		// 待出货：已付款，需要卖家备货发货
		log.Printf("[订单状态] 待出货 READY_TO_SHIP: order_sn=%s", orderSn)
		// TODO: 标记订单为待出货、推送到发货流程

	case "PROCESSED":
		// 已处理：卖家已确认并处理订单
		log.Printf("[订单状态] 已处理 PROCESSED: order_sn=%s", orderSn)
		// TODO: 结合你们业务定义“已处理”的含义做相应操作

	case "RETRY_SHIP":
		// 重试发货：之前发货失败，重新发货中
		log.Printf("[订单状态] 重试发货 RETRY_SHIP: order_sn=%s", orderSn)
		// TODO: 记录发货异常并重新安排物流

	case "SHIPPED":
		// 已发货：包裹已交给物流
		log.Printf("[订单状态] 已发货 SHIPPED: order_sn=%s", orderSn)
		// TODO: 更新订单为已发货，记录物流信息

	case "TO_CONFIRM_RECEIVE":
		// 待确认收货：等待买家确认收货
		log.Printf("[订单状态] 待确认收货 TO_CONFIRM_RECEIVE: order_sn=%s", orderSn)
		// TODO: 可以在这里记录“待收货”，用于风控或运营

	case "IN_CANCEL":
		// 待取消：买家/卖家发起取消，待平台最终确认
		log.Printf("[订单状态] 待取消 IN_CANCEL: order_sn=%s", orderSn)
		// TODO: 标记订单处于取消流程中

	case "CANCELLED":
		// 已取消：订单取消（超时未付款、卖家未发货、协商取消等）
		log.Printf("[订单状态] 已取消 CANCELLED: order_sn=%s", orderSn)
		// TODO: 回滚库存、释放资源、记录取消原因（如有额外字段）

	case "TO_RETURN":
		// 待退货：买家申请退货，尚未完成
		log.Printf("[订单状态] 待退货 TO_RETURN: order_sn=%s, return_sn=%s",
			orderSn, orderPush.Data.ReturnSN)
		// TODO: 标记订单进入退货流程，可结合退款/退货推送 code=29 一起处理

	case "COMPLETED":
		// 已完成：订单交易完全结束
		log.Printf("[订单状态] 已完成 COMPLETED: order_sn=%s", orderSn)
		// TODO: 标记订单完成、统计结算、发放积分等

	default:
		// 未知或暂未覆盖的状态，先记录日志，避免丢数据
		log.Printf("[订单状态] 未识别状态: status=%s, order_sn=%s, shop_id=%d",
			orderPush.Data.Status, orderSn, orderPush.ShopID)
	}

	log.Printf("订单状态处理完成: order_sn=%s", orderSn)
	return nil
}

// ProcessReturnRefund 处理退款/退货推送
func (s *OrderService) ProcessReturnRefund(orderPush *models.OrderStatusPush) error {
	orderSn := orderPush.Data.OrderSn
	if orderSn == "" {
		orderSn = orderPush.Data.OrderSN
	}

	log.Printf("处理退款/退货: order_sn=%s, return_sn=%s",
		orderSn, orderPush.Data.ReturnSN)

	// TODO: 在这里实现退款/退货处理逻辑

	return nil
}

// ProcessShopFrozen 处理店铺冻结推送
func (s *OrderService) ProcessShopFrozen(orderPush *models.OrderStatusPush) error {
	log.Printf("处理店铺冻结: shop_id=%d", orderPush.ShopID)

	// TODO: 在这里实现店铺冻结处理逻辑

	return nil
}

// ProcessViolationItem 处理违规商品推送
func (s *OrderService) ProcessViolationItem(orderPush *models.OrderStatusPush) error {
	log.Printf("处理违规商品: shop_id=%d, item_id=%s",
		orderPush.ShopID, orderPush.Data.ItemID)

	// TODO: 在这里实现违规商品处理逻辑

	return nil
}

// FetchOrdersFromShopee 从虾皮拉取订单
// timeRangeField: 时间字段类型 (create_time/update_time)，默认 create_time
// timeFrom: 开始时间戳（Unix时间戳，秒）
// timeTo: 结束时间戳（Unix时间戳，秒）
// pageSize: 每页数量，最大100，默认20
// cursor: 分页游标，首次请求为空
func (s *OrderService) FetchOrdersFromShopee(timeRangeField string, timeFrom, timeTo int64, pageSize int, cursor string) (map[string]interface{}, error) {
	// Get the shop configuration from database - use default or first available shop
	tokenRepo := models.NewShopeeTokenRepository(s.db)
	tokens, err := tokenRepo.GetAll()
	if err != nil || len(tokens) == 0 {
		return nil, errors.New("没有找到有效的店铺配置")
	}
	
	// Use the first available shop configuration
	shopeeToken := tokens[0]
	
	log.Printf("开始拉取虾皮订单: shop_id=%d, timeRangeField=%s, timeFrom=%d, timeTo=%d, pageSize=%d, cursor=%s",
		shopeeToken.ShopID, timeRangeField, timeFrom, timeTo, pageSize, cursor)

	// Create a temporary Shopee API client for this specific shop
	tokenExpireAt := time.Now().Add(4 * time.Hour) // Default expiration if not set
	if shopeeToken.TokenExpireAt != nil {
		tokenExpireAt = *shopeeToken.TokenExpireAt
	}
	
	shopeeClient := shareUtils.NewShopeeAPIClientWithRefresh(
		shopeeToken.PartnerID,
		shopeeToken.PartnerKey,
		shopeeToken.ShopID,
		shopeeToken.AccessToken,
		shopeeToken.RefreshToken,
		tokenExpireAt,
		shopeeToken.IsSandbox,
		// Token 刷新回调：当 token 自动刷新时，保存到数据库
		func(accessToken, refreshToken string, expireIn int64) {
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			err := tokenRepo.UpdateTokens(shopeeToken.ShopID, accessToken, refreshToken, &tokenExpireAt, nil)
			if err != nil {
				log.Printf("❌ 保存刷新后的 token 到数据库失败: %v", err)
			} else {
				log.Printf("✅ Shopee access_token 已自动刷新并保存到数据库")
			}
		},
	)

	// 调用虾皮API
	result, err := shopeeClient.GetOrderList(timeRangeField, timeFrom, timeTo, pageSize, cursor)
	if err != nil {
		// 检查错误是否与token过期相关
		errMsg := err.Error()
		if strings.Contains(errMsg, "access_token") || strings.Contains(errMsg, "token") || strings.Contains(errMsg, "Wrong sign") {
			log.Printf("检测到token相关错误，尝试刷新token后重试: %v", err)

			// Attempt to refresh token
			accessToken, newRefreshToken, expireIn, refreshErr := shareUtils.RefreshShopeeToken(
				shopeeToken.PartnerID,
				shopeeToken.PartnerKey,
				shopeeToken.ShopID,
				shopeeToken.RefreshToken,
				shopeeToken.IsSandbox,
			)
			if refreshErr != nil {
				log.Printf("刷新token失败: %v", refreshErr)
				return nil, err // Return original error
			}

			// Update token in database
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			updateErr := tokenRepo.UpdateTokens(shopeeToken.ShopID, accessToken, newRefreshToken, &tokenExpireAt, nil)
			if updateErr != nil {
				log.Printf("更新数据库中的token失败: %v", updateErr)
				// Continue anyway with the new tokens
			}

			// Update the client with new tokens
			shopeeClient.AccessToken = accessToken
			if newRefreshToken != "" {
				shopeeClient.RefreshToken = newRefreshToken
			}
			shopeeClient.TokenExpireAt = tokenExpireAt

			// Retry the API call
			log.Printf("刷新token成功，重新尝试拉取订单...")
			result, err = shopeeClient.GetOrderList(timeRangeField, timeFrom, timeTo, pageSize, cursor)
			if err != nil {
				log.Printf("重试拉取订单仍失败: %v", err)
				return nil, err
			}
		} else {
			log.Printf("拉取虾皮订单失败: %v", err)
			return nil, err
		}
	}

	return result, nil
}

// RefreshTokenAndRetry 公开的刷新token并重试方法，供控制器调用
func (s *OrderService) RefreshTokenAndRetry() error {
	// For backward compatibility, but this method is now deprecated in favor of per-request client creation
	if s.shopeeClient == nil || s.shopeeClient.RefreshToken == "" {
		return errors.New("无法刷新token：API客户端或refresh_token未配置")
	}

	log.Printf("开始刷新token: shop_id=%d", s.shopeeClient.ShopID)

	// 判断是沙箱还是正式环境
	isSandbox := strings.Contains(s.shopeeClient.BaseURL, "sandbox")
	accessToken, newRefreshToken, expireIn, err := shareUtils.RefreshShopeeToken(
		s.shopeeClient.PartnerID,
		s.shopeeClient.PartnerKey,
		s.shopeeClient.ShopID,
		s.shopeeClient.RefreshToken,
		isSandbox,
	)
	if err != nil {
		return err
	}

	// 更新API客户端中的token
	s.shopeeClient.AccessToken = accessToken
	if newRefreshToken != "" {
		s.shopeeClient.RefreshToken = newRefreshToken
	}
	s.shopeeClient.TokenExpireAt = time.Now().Add(time.Duration(expireIn) * time.Second)

	// 调用回调函数，通知外部更新配置（如果有设置）
	if s.shopeeClient.OnTokenRefresh != nil {
		s.shopeeClient.OnTokenRefresh(accessToken, s.shopeeClient.RefreshToken, expireIn)
	}

	log.Printf("✅ token刷新成功: shop_id=%d, access_token长度=%d, expire_in=%d秒",
		s.shopeeClient.ShopID, len(accessToken), expireIn)

	return nil
}

// GetShopeeClient 获取Shopee客户端，用于获取配置信息
func (s *OrderService) GetShopeeClient() *shareUtils.ShopeeAPIClient {
	return s.shopeeClient
}

// FetchOrderDetailFromShopee 从虾皮拉取订单详情
func (s *OrderService) FetchOrderDetailFromShopee(orderSnList []string) (map[string]interface{}, error) {
	// Get the shop configuration from database - use default or first available shop
	tokenRepo := models.NewShopeeTokenRepository(s.db)
	tokens, err := tokenRepo.GetAll()
	if err != nil || len(tokens) == 0 {
		return nil, errors.New("没有找到有效的店铺配置")
	}
	
	// Use the first available shop configuration
	shopeeToken := tokens[0]

	if len(orderSnList) == 0 {
		return nil, errors.New("订单号列表为空")
	}

	log.Printf("开始拉取虾皮订单详情: shop_id=%d, orderSnList=%v", shopeeToken.ShopID, orderSnList)

	// Create a temporary Shopee API client for this specific shop
	tokenExpireAt := time.Now().Add(4 * time.Hour) // Default expiration if not set
	if shopeeToken.TokenExpireAt != nil {
		tokenExpireAt = *shopeeToken.TokenExpireAt
	}
	
	shopeeClient := shareUtils.NewShopeeAPIClientWithRefresh(
		shopeeToken.PartnerID,
		shopeeToken.PartnerKey,
		shopeeToken.ShopID,
		shopeeToken.AccessToken,
		shopeeToken.RefreshToken,
		tokenExpireAt,
		shopeeToken.IsSandbox,
		// Token 刷新回调：当 token 自动刷新时，保存到数据库
		func(accessToken, refreshToken string, expireIn int64) {
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			err := tokenRepo.UpdateTokens(shopeeToken.ShopID, accessToken, refreshToken, &tokenExpireAt, nil)
			if err != nil {
				log.Printf("❌ 保存刷新后的 token 到数据库失败: %v", err)
			} else {
				log.Printf("✅ Shopee access_token 已自动刷新并保存到数据库")
			}
		},
	)

	// 调用虾皮API
	result, err := shopeeClient.GetOrderDetail(orderSnList)
	if err != nil {
		// 检查错误是否与token过期相关
		errMsg := err.Error()
		if strings.Contains(errMsg, "access_token") || strings.Contains(errMsg, "token") || strings.Contains(errMsg, "Wrong sign") {
			log.Printf("检测到token相关错误，尝试刷新token后重试: %v", err)

			// Attempt to refresh token
			accessToken, newRefreshToken, expireIn, refreshErr := shareUtils.RefreshShopeeToken(
				shopeeToken.PartnerID,
				shopeeToken.PartnerKey,
				shopeeToken.ShopID,
				shopeeToken.RefreshToken,
				shopeeToken.IsSandbox,
			)
			if refreshErr != nil {
				log.Printf("刷新token失败: %v", refreshErr)
				return nil, err // Return original error
			}

			// Update token in database
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			updateErr := tokenRepo.UpdateTokens(shopeeToken.ShopID, accessToken, newRefreshToken, &tokenExpireAt, nil)
			if updateErr != nil {
				log.Printf("更新数据库中的token失败: %v", updateErr)
				// Continue anyway with the new tokens
			}

			// Update the client with new tokens
			shopeeClient.AccessToken = accessToken
			if newRefreshToken != "" {
				shopeeClient.RefreshToken = newRefreshToken
			}
			shopeeClient.TokenExpireAt = tokenExpireAt

			// Retry the API call
			log.Printf("刷新token成功，重新尝试拉取订单详情...")
			result, err = shopeeClient.GetOrderDetail(orderSnList)
			if err != nil {
				log.Printf("重试拉取订单详情仍失败: %v", err)
				return nil, err
			}
		} else {
			log.Printf("拉取虾皮订单详情失败: %v", err)
			return nil, err
		}
	}

	return result, nil
}

// FetchShopListFromShopee 从虾皮拉取店铺列表
func (s *OrderService) FetchShopListFromShopee() (map[string]interface{}, error) {
	// 获取所有已配置的店铺
	tokenRepo := models.NewShopeeTokenRepository(s.db) // Use the service's database instance
	tokens, err := tokenRepo.GetAll()
	if err != nil {
		return nil, errors.New("获取店铺配置列表失败: " + err.Error())
	}

	if len(tokens) == 0 {
		return nil, errors.New("没有找到任何店铺配置")
	}

	// For now, return info about all configured shops
	result := make(map[string]interface{})
	result["shops"] = tokens
	result["total"] = len(tokens)

	return result, nil
}

// FetchShopDetailFromShopee 从虾皮拉取店铺详情
func (s *OrderService) FetchShopDetailFromShopee(shopID int64) (map[string]interface{}, error) {
	// Get the specific shop configuration from database
	tokenRepo := models.NewShopeeTokenRepository(s.db) // Use the service's database instance
	shopeeToken, err := tokenRepo.GetByShopID(shopID)
	if err != nil {
		return nil, fmt.Errorf("获取店铺配置失败 (shop_id=%d): %v", shopID, err)
	}

	log.Printf("开始拉取虾皮店铺详情: shop_id=%d", shopID)

	// Create a temporary Shopee API client for this specific shop
	tokenExpireAt := time.Now().Add(4 * time.Hour) // Default expiration if not set
	if shopeeToken.TokenExpireAt != nil {
		tokenExpireAt = *shopeeToken.TokenExpireAt
	}

	shopeeClient := shareUtils.NewShopeeAPIClientWithRefresh(
		shopeeToken.PartnerID,
		shopeeToken.PartnerKey,
		shopeeToken.ShopID,
		shopeeToken.AccessToken,
		shopeeToken.RefreshToken,
		tokenExpireAt,
		shopeeToken.IsSandbox,
		// Token 刷新回调：当 token 自动刷新时，保存到数据库
		func(accessToken, refreshToken string, expireIn int64) {
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			err := tokenRepo.UpdateTokens(shopeeToken.ShopID, accessToken, refreshToken, &tokenExpireAt, nil)
			if err != nil {
				log.Printf("❌ 保存刷新后的 token 到数据库失败: %v", err)
			} else {
				log.Printf("✅ Shopee access_token 已自动刷新并保存到数据库")
			}
		},
	)

	// Call the client's GetShopInfo method
	result, err := shopeeClient.GetShopInfo()
	if err != nil {
		// Check if error is related to token expiration
		errMsg := err.Error()
		if strings.Contains(errMsg, "access_token") || strings.Contains(errMsg, "token") || strings.Contains(errMsg, "Wrong sign") {
			log.Printf("检测到token相关错误，尝试刷新token后重试: %v", err)

			// Attempt to refresh token
			accessToken, newRefreshToken, expireIn, refreshErr := shareUtils.RefreshShopeeToken(
				shopeeToken.PartnerID,
				shopeeToken.PartnerKey,
				shopeeToken.ShopID,
				shopeeToken.RefreshToken,
				shopeeToken.IsSandbox,
			)
			if refreshErr != nil {
				log.Printf("刷新token失败: %v", refreshErr)
				return nil, err // Return original error
			}

			// Update token in database
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			updateErr := tokenRepo.UpdateTokens(shopID, accessToken, newRefreshToken, &tokenExpireAt, nil)
			if updateErr != nil {
				log.Printf("更新数据库中的token失败: %v", updateErr)
				// Continue anyway with the new tokens
			}

			// Update the client with new tokens
			shopeeClient.AccessToken = accessToken
			if newRefreshToken != "" {
				shopeeClient.RefreshToken = newRefreshToken
			}
			shopeeClient.TokenExpireAt = tokenExpireAt

			// Retry the API call
			log.Printf("刷新token成功，重新尝试拉取店铺详情...")
			result, err = shopeeClient.GetShopInfo()
			if err != nil {
				log.Printf("重试拉取店铺详情仍失败: %v", err)
				return nil, err
			}
		} else {
			log.Printf("拉取虾皮店铺详情失败: %v", err)
			return nil, err
		}
	}

	return result, nil
}
