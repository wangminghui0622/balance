package services

import (
	"balance/internal/models"
	shareUtils "balance/internal/utils"
	"errors"
	"log"
	"strings"
	"time"
)

// OrderService 订单服务
// 本系统本身就是商家服务器，直接处理订单
type OrderService struct {
	shopeeClient *shareUtils.ShopeeAPIClient
}

// NewOrderService 创建订单服务
func NewOrderService(merchantURL string) *OrderService {
	// merchantURL 参数保留以兼容现有代码，但不再使用
	return &OrderService{}
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
	if s.shopeeClient == nil {
		return nil, errors.New("虾皮API客户端未配置")
	}

	// 参数校验
	if timeRangeField == "" {
		timeRangeField = "create_time"
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if timeFrom <= 0 {
		timeFrom = time.Now().AddDate(0, 0, -7).Unix() // 默认拉取最近7天
	}
	if timeTo <= 0 {
		timeTo = time.Now().Unix()
	}

	log.Printf("开始拉取虾皮订单: timeRangeField=%s, timeFrom=%d, timeTo=%d, pageSize=%d, cursor=%s",
		timeRangeField, timeFrom, timeTo, pageSize, cursor)

	// 调用虾皮API
	result, err := s.shopeeClient.GetOrderList(timeRangeField, timeFrom, timeTo, pageSize, cursor)
	if err != nil {
		log.Printf("拉取虾皮订单失败: %v", err)
		return nil, err
	}

	return result, nil
}

// FetchOrderDetailFromShopee 从虾皮拉取订单详情
func (s *OrderService) FetchOrderDetailFromShopee(orderSnList []string) (map[string]interface{}, error) {
	if s.shopeeClient == nil {
		return nil, errors.New("虾皮API客户端未配置")
	}

	if len(orderSnList) == 0 {
		return nil, errors.New("订单号列表为空")
	}

	log.Printf("开始拉取虾皮订单详情: orderSnList=%v", orderSnList)

	// 调用虾皮API
	result, err := s.shopeeClient.GetOrderDetail(orderSnList)
	if err != nil {
		log.Printf("拉取虾皮订单详情失败: %v", err)
		return nil, err
	}

	return result, nil
}
