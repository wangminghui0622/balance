package shopee

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	BaseResponse
	Response struct {
		More       bool   `json:"more"`
		NextCursor string `json:"next_cursor"`
		OrderList  []struct {
			OrderSN     string `json:"order_sn"`
			OrderStatus string `json:"order_status"`
		} `json:"order_list"`
	} `json:"response"`
}

// GetOrderList 获取订单列表
func (c *Client) GetOrderList(accessToken string, shopID uint64, timeRangeField string, timeFrom, timeTo int64, pageSize int, cursor string, orderStatus string) (*OrderListResponse, error) {
	path := "/api/v2/order/get_order_list"
	params := url.Values{}
	params.Set("time_range_field", timeRangeField)
	params.Set("time_from", strconv.FormatInt(timeFrom, 10))
	params.Set("time_to", strconv.FormatInt(timeTo, 10))
	params.Set("page_size", strconv.Itoa(pageSize))
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	if orderStatus != "" {
		params.Set("order_status", orderStatus)
	}
	params.Set("response_optional_fields", "order_status")

	fmt.Printf("[SyncDebug][API] GetOrderList 请求: host=%s shop_id=%d time_from=%d time_to=%d cursor=%q\n",
		c.host, shopID, timeFrom, timeTo, cursor)

	respBody, err := c.Get(path, params, accessToken, shopID)
	if err != nil {
		fmt.Printf("[SyncDebug][API] GetOrderList 请求失败: %v\n", err)
		return nil, err
	}

	respStr := string(respBody)
	if len(respStr) > 500 {
		respStr = respStr[:500] + "...(truncated)"
	}
	fmt.Printf("[SyncDebug][API] GetOrderList 原始响应: %s\n", respStr)

	var result OrderListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		fmt.Printf("[SyncDebug][API] GetOrderList API错误: error=%s message=%s request_id=%s\n",
			result.Error, result.Message, result.RequestID)
		return nil, fmt.Errorf("获取订单列表失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
