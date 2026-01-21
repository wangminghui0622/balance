package controllers

import (
	"balance/admin/services"
	"balance/internal/models"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OrderController 订单控制器
type OrderController struct {
	orderService *services.OrderService
	db           *gorm.DB
}

// NewOrderController 创建订单控制器
func NewOrderController(orderService *services.OrderService, db *gorm.DB) *OrderController {
	return &OrderController{
		orderService: orderService,
		db:           db,
	}
}

// ShopeeCallback 接收虾皮的订单推送回调
func (ctrl *OrderController) ShopeeCallback(c *gin.Context) {
	// 记录请求信息（用于调试）
	log.Printf("收到虾皮回调请求: Method=%s, Path=%s, Content-Type=%s, RemoteAddr=%s",
		c.Request.Method, c.Request.URL.Path, c.GetHeader("Content-Type"), c.ClientIP())

	// 解析 JSON（不要先读取 RawData，否则 ShouldBindJSON 无法读取）
	var orderPush models.OrderStatusPush
	if err := c.ShouldBindJSON(&orderPush); err != nil {
		// 如果解析失败，尝试读取原始数据用于日志和入库
		body, _ := c.GetRawData()
		log.Printf("解析JSON失败: %v, 原始数据: %s", err, string(body))

		// 将非订单类型或格式错误的原始数据记录到 push_data 表
		if len(body) > 0 && ctrl.db != nil {
			record := &models.PushData{
				Data: string(body),
			}
			if saveErr := ctrl.db.Create(record).Error; saveErr != nil {
				log.Printf("保存非订单推送到 push_data 失败: %v", saveErr)
			}
		}

		// 如果是验证请求或格式不正确的请求，返回成功
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
		})
		return
	}

	// 记录订单信息
	log.Printf("订单推送信息: shop_id=%d, order_sn=%s, status=%s, code=%d, msg_id=%s",
		orderPush.ShopID, orderPush.Data.OrderSn, orderPush.Data.Status, orderPush.Code, orderPush.MsgID)

	// 参数校验（参考 balance 系统的校验逻辑）
	if orderPush.ShopID == 0 || orderPush.Data == nil {
		log.Printf("订单数据不完整: shop_id=%d, data=%+v", orderPush.ShopID, orderPush.Data)
		// 如果是验证请求，返回成功
		if orderPush.ShopID == 0 {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "订单数据不完整",
		})
		return
	}

	// 检查订单号（ordersn 或 order_sn）
	orderSn := orderPush.Data.OrderSn
	if orderSn == "" {
		orderSn = orderPush.Data.OrderSN
	}
	if orderSn == "" {
		log.Printf("订单号为空: shop_id=%d", orderPush.ShopID)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "订单号为空",
		})
		return
	}

	// 处理订单推送
	err := ctrl.orderService.ReceiveOrderFromShopee(&orderPush)
	if err != nil {
		log.Printf("处理订单推送失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "处理订单失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "订单接收成功",
		"data": gin.H{
			"order_sn": orderPush.Data.OrderSn,
			"shop_id":  orderPush.ShopID,
		},
	})
}

// FetchOrders 拉取虾皮订单列表
func (ctrl *OrderController) FetchOrders(c *gin.Context) {
	// 解析请求参数
	var req struct {
		TimeRangeField string `json:"time_range_field" form:"time_range_field"` // create_time/update_time
		TimeFrom       int64  `json:"time_from" form:"time_from"`               // 开始时间戳（秒）
		TimeTo         int64  `json:"time_to" form:"time_to"`                   // 结束时间戳（秒）
		PageSize       int    `json:"page_size" form:"page_size"`               // 每页数量，最大100
		Cursor         string `json:"cursor" form:"cursor"`                       // 分页游标
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 调用服务拉取订单
	result, err := ctrl.orderService.FetchOrdersFromShopee(
		req.TimeRangeField,
		req.TimeFrom,
		req.TimeTo,
		req.PageSize,
		req.Cursor,
	)
	if err != nil {
		log.Printf("拉取订单失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "拉取订单失败: " + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "拉取订单成功",
		"data":    result,
	})
}

// FetchOrderDetail 拉取虾皮订单详情
func (ctrl *OrderController) FetchOrderDetail(c *gin.Context) {
	// 解析请求参数
	var req struct {
		OrderSnList string `json:"order_sn_list" form:"order_sn_list" binding:"required"` // 订单号列表，逗号分隔
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 解析订单号列表
	orderSnList := strings.Split(req.OrderSnList, ",")
	for i := range orderSnList {
		orderSnList[i] = strings.TrimSpace(orderSnList[i])
	}

	// 调用服务拉取订单详情
	result, err := ctrl.orderService.FetchOrderDetailFromShopee(orderSnList)
	if err != nil {
		log.Printf("拉取订单详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "拉取订单详情失败: " + err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "拉取订单详情成功",
		"data":    result,
	})
}
