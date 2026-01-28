package services

import (
	"balance/internal/models"
	shareUtils "balance/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"strconv"
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
func NewOrderService(db *gorm.DB) *OrderService {
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
	var shops []*models.ShopeeShop
	err := s.db.Where("access_token IS NOT NULL AND access_token != ''").Find(&shops).Error
	if err != nil || len(shops) == 0 {
		return nil, errors.New("没有找到有效的店铺配置")
	}

	// Use the first available shop configuration
	shop := shops[0]

	// 获取全局配置（PartnerKey 和 IsSandbox）
	authService := NewAuthService(s.db)
	authCfg, err := authService.GetByPartnerId()
	if err != nil {
		return nil, fmt.Errorf("获取授权配置失败: %v", err)
	}

	log.Printf("开始拉取虾皮订单: shop_id=%d, timeRangeField=%s, timeFrom=%d, timeTo=%d, pageSize=%d, cursor=%s",
		shop.ShopID, timeRangeField, timeFrom, timeTo, pageSize, cursor)

	// Create a temporary Shopee API client for this specific shop
	tokenExpireAt := time.Now().Add(4 * time.Hour) // Default expiration if not set
	if shop.TokenExpireAt != nil {
		tokenExpireAt = *shop.TokenExpireAt
	}

	var accessToken, refreshToken string
	if shop.AccessToken != nil {
		accessToken = *shop.AccessToken
	}
	if shop.RefreshToken != nil {
		refreshToken = *shop.RefreshToken
	}

	shopeeClient := shareUtils.NewShopeeAPIClientWithRefresh(
		shop.PartnerID,
		authCfg.PartnerKey,
		shop.ShopID,
		accessToken,
		refreshToken,
		tokenExpireAt,
		authCfg.IsSandbox,
		// Token 刷新回调：当 token 自动刷新时，保存到数据库
		func(accessToken, refreshToken string, expireIn int64) {
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			err := s.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", shop.ShopID).Updates(map[string]interface{}{
				"access_token":       accessToken,
				"refresh_token":      refreshToken,
				"token_expire_at":    tokenExpireAt,
				"last_token_refresh": time.Now(),
			}).Error
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
			if shop.RefreshToken == nil || *shop.RefreshToken == "" {
				return nil, errors.New("没有可用的 refresh_token")
			}
			accessToken, newRefreshToken, expireIn, refreshErr := shareUtils.RefreshShopeeToken(
				shop.PartnerID,
				authCfg.PartnerKey,
				shop.ShopID,
				*shop.RefreshToken,
				authCfg.IsSandbox,
			)
			if refreshErr != nil {
				log.Printf("刷新token失败: %v", refreshErr)
				return nil, err // Return original error
			}

			// Update token in database
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			updateErr := s.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", shop.ShopID).Updates(map[string]interface{}{
				"access_token":       accessToken,
				"refresh_token":      newRefreshToken,
				"token_expire_at":    tokenExpireAt,
				"last_token_refresh": time.Now(),
			}).Error
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
	var shops []*models.ShopeeShop
	err := s.db.Where("access_token IS NOT NULL AND access_token != ''").Find(&shops).Error
	if err != nil || len(shops) == 0 {
		return nil, errors.New("没有找到有效的店铺配置")
	}

	// Use the first available shop configuration
	shop := shops[0]

	// 获取全局配置（PartnerKey 和 IsSandbox）
	authService := NewAuthService(s.db)
	authCfg, err := authService.GetByPartnerId()
	if err != nil {
		return nil, fmt.Errorf("获取授权配置失败: %v", err)
	}

	if len(orderSnList) == 0 {
		return nil, errors.New("订单号列表为空")
	}

	log.Printf("开始拉取虾皮订单详情: shop_id=%d, orderSnList=%v", shop.ShopID, orderSnList)

	// Create a temporary Shopee API client for this specific shop
	tokenExpireAt := time.Now().Add(4 * time.Hour) // Default expiration if not set
	if shop.TokenExpireAt != nil {
		tokenExpireAt = *shop.TokenExpireAt
	}

	var accessToken, refreshToken string
	if shop.AccessToken != nil {
		accessToken = *shop.AccessToken
	}
	if shop.RefreshToken != nil {
		refreshToken = *shop.RefreshToken
	}

	shopeeClient := shareUtils.NewShopeeAPIClientWithRefresh(
		shop.PartnerID,
		authCfg.PartnerKey,
		shop.ShopID,
		accessToken,
		refreshToken,
		tokenExpireAt,
		authCfg.IsSandbox,
		// Token 刷新回调：当 token 自动刷新时，保存到数据库
		func(accessToken, refreshToken string, expireIn int64) {
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			err := s.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", shop.ShopID).Updates(map[string]interface{}{
				"access_token":       accessToken,
				"refresh_token":      refreshToken,
				"token_expire_at":    tokenExpireAt,
				"last_token_refresh": time.Now(),
			}).Error
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
			if shop.RefreshToken == nil || *shop.RefreshToken == "" {
				return nil, errors.New("没有可用的 refresh_token")
			}
			accessToken, newRefreshToken, expireIn, refreshErr := shareUtils.RefreshShopeeToken(
				shop.PartnerID,
				authCfg.PartnerKey,
				shop.ShopID,
				*shop.RefreshToken,
				authCfg.IsSandbox,
			)
			if refreshErr != nil {
				log.Printf("刷新token失败: %v", refreshErr)
				return nil, err // Return original error
			}

			// Update token in database
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			updateErr := s.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", shop.ShopID).Updates(map[string]interface{}{
				"access_token":       accessToken,
				"refresh_token":      newRefreshToken,
				"token_expire_at":    tokenExpireAt,
				"last_token_refresh": time.Now(),
			}).Error
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
	var shops []*models.ShopeeShop
	err := s.db.Find(&shops).Error
	if err != nil {
		return nil, errors.New("获取店铺配置列表失败: " + err.Error())
	}

	if len(shops) == 0 {
		return nil, errors.New("没有找到任何店铺配置")
	}

	// For now, return info about all configured shops
	result := make(map[string]interface{})
	result["shops"] = shops
	result["total"] = len(shops)

	return result, nil
}

// FetchShopDetailFromShopee 从虾皮拉取店铺详情
func (s *OrderService) FetchShopDetailFromShopee(shopID int64) (map[string]interface{}, error) {
	// Get the specific shop configuration from database
	shopRepo := models.NewShopeeShopRepository(s.db)
	shop, err := shopRepo.GetByShopID(shopID)
	if err != nil {
		return nil, fmt.Errorf("获取店铺配置失败 (shop_id=%d): %v", shopID, err)
	}

	// 获取全局配置（PartnerKey 和 IsSandbox）
	authService := NewAuthService(s.db)
	authCfg, err := authService.GetByPartnerId()
	if err != nil {
		return nil, fmt.Errorf("获取授权配置失败: %v", err)
	}

	log.Printf("开始拉取虾皮店铺详情: shop_id=%d", shopID)

	// Create a temporary Shopee API client for this specific shop
	tokenExpireAt := time.Now().Add(4 * time.Hour) // Default expiration if not set
	if shop.TokenExpireAt != nil {
		tokenExpireAt = *shop.TokenExpireAt
	}

	var accessToken, refreshToken string
	if shop.AccessToken != nil {
		accessToken = *shop.AccessToken
	}
	if shop.RefreshToken != nil {
		refreshToken = *shop.RefreshToken
	}

	shopeeClient := shareUtils.NewShopeeAPIClientWithRefresh(
		shop.PartnerID,
		authCfg.PartnerKey,
		shop.ShopID,
		accessToken,
		refreshToken,
		tokenExpireAt,
		authCfg.IsSandbox,
		// Token 刷新回调：当 token 自动刷新时，保存到数据库
		func(accessToken, refreshToken string, expireIn int64) {
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			err := s.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", shop.ShopID).Updates(map[string]interface{}{
				"access_token":       accessToken,
				"refresh_token":      refreshToken,
				"token_expire_at":    tokenExpireAt,
				"last_token_refresh": time.Now(),
			}).Error
			if err != nil {
				log.Printf("❌ 保存刷新后的 token 到数据库失败: %v", err)
			} else {
				log.Printf("✅ Shopee access_token 已自动刷新并保存到数据库")
			}
		},
	)

	// Call the client's GetShopInfo method
	basicInfo, err := shopeeClient.GetShopInfo()
	if err != nil {
		// Check if error is related to token expiration
		errMsg := err.Error()
		if strings.Contains(errMsg, "access_token") || strings.Contains(errMsg, "token") || strings.Contains(errMsg, "Wrong sign") {
			log.Printf("检测到token相关错误，尝试刷新token后重试: %v", err)

			// Attempt to refresh token
			if shop.RefreshToken == nil || *shop.RefreshToken == "" {
				return nil, errors.New("没有可用的 refresh_token")
			}
			accessToken, newRefreshToken, expireIn, refreshErr := shareUtils.RefreshShopeeToken(
				shop.PartnerID,
				authCfg.PartnerKey,
				shop.ShopID,
				*shop.RefreshToken,
				authCfg.IsSandbox,
			)
			if refreshErr != nil {
				log.Printf("刷新token失败: %v", refreshErr)
				return nil, err // Return original error
			}

			// Update token in database
			tokenExpireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
			updateErr := s.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", shopID).Updates(map[string]interface{}{
				"access_token":       accessToken,
				"refresh_token":      newRefreshToken,
				"token_expire_at":    tokenExpireAt,
				"last_token_refresh": time.Now(),
			}).Error
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
			basicInfo, err = shopeeClient.GetShopInfo()
			if err != nil {
				log.Printf("重试拉取店铺详情仍失败: %v", err)
				return nil, err
			}
		} else {
			log.Printf("拉取虾皮店铺详情失败: %v", err)
			return nil, err
		}
	}

	// Get additional detailed shop profile information
	profileInfo, err := shopeeClient.GetShopProfileInfo()
	if err != nil {
		log.Printf("获取店铺详细信息失败 (可选): %v", err)
		// Don't fail the whole operation if profile info fails
		profileInfo = nil
	}

	// Get shop performance information
	performanceInfo, err := shopeeClient.GetShopPerformance()
	if err != nil {
		log.Printf("获取店铺表现评分失败 (可选): %v", err)
		// Don't fail the whole operation if performance info fails
		performanceInfo = nil
	}

	// 将店铺信息同步到数据库
	err = s.SyncShopInfoToDatabase(shopID, basicInfo, profileInfo, performanceInfo)
	if err != nil {
		log.Printf("同步店铺信息到数据库失败: %v", err)
		// 不返回错误，因为API调用已经成功
	}

	// Combine all information into a comprehensive result
	result := make(map[string]interface{})

	// Add basic shop info
	for k, v := range basicInfo {
		result[k] = v
	}

	// Add detailed profile info if available
	if profileInfo != nil {
		result["profile"] = profileInfo
	}

	// Add performance info if available
	if performanceInfo != nil {
		result["performance"] = performanceInfo
	}

	return result, nil
}

// SyncShopInfoToDatabase 将店铺信息同步到数据库
func (s *OrderService) SyncShopInfoToDatabase(shopID int64, basicInfo, profileInfo, performanceInfo map[string]interface{}) error {
	shopRepo := models.NewShopeeShopRepository(s.db)
	shopIDStr := strconv.FormatInt(shopID, 10)

	// 解析基本店铺信息
	type AuthResponse struct {
		AuthTime             float64     `json:"auth_time"`
		Error                string      `json:"error,omitempty"`
		ExpireTime           float64     `json:"expire_time"`
		IsCb                 bool        `json:"is_cb"`
		IsDirectShop         bool        `json:"is_direct_shop"`
		IsMainShop           bool        `json:"is_main_shop"`
		IsMartShop           bool        `json:"is_mart_shop"`
		IsOneAwb             bool        `json:"is_one_awb"`
		IsOutletShop         bool        `json:"is_outlet_shop"`
		IsSip                bool        `json:"is_sip"`
		IsUpgradedCbsc       bool        `json:"is_upgraded_cbsc"`
		LinkedDirectShopList []string    `json:"linked_direct_shop_list"`
		LinkedMainShopID     int         `json:"linked_main_shop_id"`
		MerchantID           interface{} `json:"merchant_id"`
		Message              string      `json:"message,omitempty"`
		Region               string      `json:"region"`
		RequestID            string      `json:"request_id"`
		ShopFulfillmentFlag  string      `json:"shop_fulfillment_flag"`
		ShopName             string      `json:"shop_name"`
		Status               string      `json:"status"`
	}

	// 解析 basicInfo
	jsonData, _ := json.Marshal(basicInfo)
	var r AuthResponse
	if err := json.Unmarshal(jsonData, &r); err != nil {
		return fmt.Errorf("解析店铺基本信息失败: %v", err)
	}

	// 解析 MerchantID
	var merchantID *int64
	if r.MerchantID != nil {
		switch v := r.MerchantID.(type) {
		case float64:
			id := int64(v)
			merchantID = &id
		case int:
			id := int64(v)
			merchantID = &id
		case int64:
			merchantID = &v
		}
	}

	// 解析时间
	var authTime *time.Time
	if r.AuthTime > 0 {
		t := time.Unix(int64(r.AuthTime), 0)
		authTime = &t
	}

	// 解析店铺状态
	var status int16 = 1 // 默认正常
	if r.Status == "NORMAL" {
		status = 1
	} else if r.Status == "BANNED" {
		status = 2
	} else if r.Status == "FROZEN" {
		status = 3
	} else if r.Status == "CLOSED" {
		status = 4
	}

	// 获取现有店铺信息（保留 token 等信息）
	existingShop, err := shopRepo.GetByShopID(shopID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("查询店铺信息失败: %v", err)
	}

	// 创建或更新店铺记录
	shop := &models.ShopeeShop{
		ShopID:     shopID,
		ShopIDStr:  shopIDStr,
		ShopName:   r.ShopName,
		Region:     r.Region,
		IsCbShop:   r.IsCb,
		Status:     status,
		AuthStatus: 1,
		ExpireTime: int64(r.ExpireTime),
		AuthTime:   authTime,
		MerchantID: merchantID,
	}

	// 如果店铺已存在，合并现有数据
	if existingShop != nil {
		shop.PartnerID = existingShop.PartnerID
		shop.AccessToken = existingShop.AccessToken
		shop.RefreshToken = existingShop.RefreshToken
		shop.TokenExpireAt = existingShop.TokenExpireAt
		shop.Currency = existingShop.Currency
		shop.AutoSync = existingShop.AutoSync
		shop.SyncInterval = existingShop.SyncInterval
		shop.SyncItems = existingShop.SyncItems
		shop.SyncOrders = existingShop.SyncOrders
		shop.SyncLogistics = existingShop.SyncLogistics
		shop.SyncFinance = existingShop.SyncFinance
	} else {
		// 新店铺，设置默认值
		shop.Currency = "MYR"
		shop.AutoSync = true
		shop.SyncInterval = 3600
		shop.SyncItems = true
		shop.SyncOrders = true
		shop.SyncLogistics = true
		shop.SyncFinance = true
	}

	// 更新最后同步时间
	now := time.Now()
	shop.LastSyncAt = &now

	err = shopRepo.CreateOrUpdate(shop)
	if err != nil {
		return fmt.Errorf("保存店铺信息到数据库失败: %v", err)
	}

	log.Printf("✅ 店铺信息已同步到数据库: shop_id=%d, shop_name=%s, region=%s", shopID, r.ShopName, r.Region)
	return nil
}

// GetShopList 从数据库获取店铺列表（支持过滤）
func (s *OrderService) GetShopList(filters map[string]interface{}) ([]*models.ShopeeShop, int64, error) {
	var shops []*models.ShopeeShop
	var total int64

	query := s.db.Model(&models.ShopeeShop{})

	// 应用过滤条件
	if authStatus, ok := filters["auth_status"]; ok {
		query = query.Where("auth_status = ?", authStatus)
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if region, ok := filters["region"]; ok {
		query = query.Where("region = ?", region)
	}
	if shopName, ok := filters["shop_name"]; ok {
		query = query.Where("shop_name LIKE ?", "%"+shopName.(string)+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	if page, ok := filters["page"]; ok {
		pageSize := 20
		if ps, ok := filters["page_size"]; ok {
			pageSize = ps.(int)
		}
		offset := (page.(int) - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// 排序
	orderBy := "created_at DESC"
	if ob, ok := filters["order_by"]; ok {
		orderBy = ob.(string)
	}
	query = query.Order(orderBy)

	// 查询
	err := query.Find(&shops).Error
	if err != nil {
		return nil, 0, err
	}

	return shops, total, nil
}

// UpdateShopStatus 更新店铺状态
func (s *OrderService) UpdateShopStatus(shopID int64, status int16) error {
	return s.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", shopID).Update("status", status).Error
}

// UpdateShopAuthStatus 更新店铺授权状态
func (s *OrderService) UpdateShopAuthStatus(shopID int64, authStatus int16) error {
	shopRepo := models.NewShopeeShopRepository(s.db)
	now := time.Now()
	return shopRepo.UpdateAuthStatus(shopID, authStatus, &now)
}

// SyncAllShops 批量同步所有店铺信息
func (s *OrderService) SyncAllShops() error {
	// 获取所有已授权的店铺
	var shops []*models.ShopeeShop
	err := s.db.Where("access_token IS NOT NULL AND access_token != '' AND auth_status = ?", 1).Find(&shops).Error
	if err != nil {
		return fmt.Errorf("获取店铺列表失败: %v", err)
	}

	if len(shops) == 0 {
		log.Printf("没有需要同步的店铺")
		return nil
	}

	log.Printf("开始批量同步店铺信息，共 %d 个店铺", len(shops))

	successCount := 0
	failCount := 0

	for _, shop := range shops {
		_, err := s.FetchShopDetailFromShopee(shop.ShopID)
		if err != nil {
			log.Printf("同步店铺失败 (shop_id=%d): %v", shop.ShopID, err)
			failCount++
		} else {
			successCount++
		}
	}

	log.Printf("批量同步完成: 成功 %d 个，失败 %d 个", successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("部分店铺同步失败: 成功 %d 个，失败 %d 个", successCount, failCount)
	}

	return nil
}

// ProcessShopFrozen 处理店铺冻结推送
func (s *OrderService) ProcessShopFrozen(orderPush *models.OrderStatusPush) error {
	log.Printf("处理店铺冻结: shop_id=%d", orderPush.ShopID)

	// 更新店铺状态为冻结
	err := s.UpdateShopStatus(orderPush.ShopID, 3) // 3 = 冻结
	if err != nil {
		log.Printf("更新店铺冻结状态失败: %v", err)
		return err
	}

	// 更新授权状态为已撤销
	err = s.UpdateShopAuthStatus(orderPush.ShopID, 3) // 3 = 已撤销
	if err != nil {
		log.Printf("更新店铺授权状态失败: %v", err)
		return err
	}

	log.Printf("✅ 店铺冻结状态已更新: shop_id=%d", orderPush.ShopID)
	return nil
}
