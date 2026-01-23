package models

import (
	"crypto/rand"
	"math/big"
	"time"
)

//PartnerID := int64(1203446)
//PartnerKey := "shpk724b6a656d626b696b756345464e6b614d524664716c61525a4e4e4f466c"
//IsSandbox := true
//redirect := "https://kx9y.com"

type AuthUrlReq struct {
	PartnerID  int64  `json:"partnerID"`
	PartnerKey string `json:"partnerKey"`
	IsSandbox  bool   `json:"isSandbox"`
	Redirect   string `json:"redirect"`
}

// OrderStatusPush 订单状态推送数据结构
type OrderStatusPush struct {
	MsgID     string     `json:"msg_id,omitempty"` // 消息ID
	Data      *OrderData `json:"data"`
	ShopID    int64      `json:"shop_id"`
	Code      int        `json:"code"`
	Timestamp int64      `json:"timestamp"`
}

// OrderData 订单数据
type OrderData struct {
	OrderSn           string        `json:"ordersn"`
	Status            string        `json:"status"` // 可能是 "UNPAID", "0", "1" 等
	CompletedScenario string        `json:"completed_scenario,omitempty"`
	UpdateTime        int64         `json:"update_time"`
	OrderSN           string        `json:"order_sn,omitempty"`
	ReturnSN          string        `json:"return_sn,omitempty"`
	ItemID            string        `json:"item_id,omitempty"`
	ItemName          string        `json:"item_name,omitempty"`
	ItemStatus        string        `json:"item_status,omitempty"`
	Items             []interface{} `json:"items,omitempty"` // 商品列表（可能是空数组）
}

// NewOrderStatusPush 创建订单推送数据
// code: 3=订单推送, 29=退款/退货推送, 5=店铺冻结, 16=违规商品
func NewOrderStatusPush(shopID int64, orderSn string, status string, code int) *OrderStatusPush {
	return &OrderStatusPush{
		Data: &OrderData{
			OrderSn:    orderSn,
			Status:     status,
			UpdateTime: time.Now().Unix(),
		},
		ShopID:    shopID,
		Code:      code,
		Timestamp: time.Now().Unix(),
	}
}

// GenerateOrderSn 生成订单号
func GenerateOrderSn() string {
	return "SP" + time.Now().Format("20060102150405") + generateRandomString(6)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range b {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			// 如果随机数生成失败，使用时间戳作为fallback
			b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		} else {
			b[i] = charset[n.Int64()]
		}
	}
	return string(b)
}
