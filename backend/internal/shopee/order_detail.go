package shopee

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// OrderDetailResponse 订单详情响应
type OrderDetailResponse struct {
	BaseResponse
	Response struct {
		OrderList []OrderDetail `json:"order_list"`
	} `json:"response"`
}

// OrderDetail 订单详情
type OrderDetail struct {
	OrderSN          string  `json:"order_sn"`
	Region           string  `json:"region"`
	Currency         string  `json:"currency"`
	COD              bool    `json:"cod"`
	TotalAmount      float64 `json:"total_amount"`
	OrderStatus      string  `json:"order_status"`
	ShippingCarrier  string  `json:"shipping_carrier"`
	PaymentMethod    string  `json:"payment_method"`
	CreateTime       int64   `json:"create_time"`
	UpdateTime       int64   `json:"update_time"`
	PayTime          int64   `json:"pay_time"`
	ShipByDate       int64   `json:"ship_by_date"`
	BuyerUserID      int64   `json:"buyer_user_id"`
	BuyerUsername    string  `json:"buyer_username"`
	TrackingNo       string  `json:"tracking_no"`
	RecipientAddress struct {
		Name        string `json:"name"`
		Phone       string `json:"phone"`
		Town        string `json:"town"`
		District    string `json:"district"`
		City        string `json:"city"`
		State       string `json:"state"`
		Region      string `json:"region"`
		Zipcode     string `json:"zipcode"`
		FullAddress string `json:"full_address"`
	} `json:"recipient_address"`
	ItemList []struct {
		ItemID             int64   `json:"item_id"`
		ItemName           string  `json:"item_name"`
		ItemSKU            string  `json:"item_sku"`
		ModelID            int64   `json:"model_id"`
		ModelName          string  `json:"model_name"`
		ModelSKU           string  `json:"model_sku"`
		ModelQuantity      int     `json:"model_quantity_purchased"`
		ModelOriginalPrice float64 `json:"model_original_price"`
	} `json:"item_list"`
	PackageList []struct {
		PackageNumber   string `json:"package_number"`
		LogisticsStatus string `json:"logistics_status"`
		ShippingCarrier string `json:"shipping_carrier"`
		ItemList        []struct {
			ItemID   int64 `json:"item_id"`
			ModelID  int64 `json:"model_id"`
			Quantity int   `json:"quantity"`
		} `json:"item_list"`
	} `json:"package_list"`
}

// GetOrderDetail 获取订单详情
func (c *Client) GetOrderDetail(accessToken string, shopID uint64, orderSNs []string) (*OrderDetailResponse, error) {
	path := "/api/v2/order/get_order_detail"
	params := url.Values{}

	sort.Strings(orderSNs)
	params.Set("order_sn_list", strings.Join(orderSNs, ","))

	responseFields := []string{
		"order_sn", "region", "currency", "cod", "total_amount", "order_status",
		"shipping_carrier", "payment_method", "create_time", "update_time", "pay_time",
		"ship_by_date", "buyer_user_id", "buyer_username", "tracking_no",
		"recipient_address", "item_list", "package_list",
	}
	params.Set("response_optional_fields", strings.Join(responseFields, ","))

	respBody, err := c.Get(path, params, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result OrderDetailResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取订单详情失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
