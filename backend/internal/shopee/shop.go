package shopee

import (
	"encoding/json"
	"fmt"
)

// ShopInfoResponse 店铺信息响应
type ShopInfoResponse struct {
	BaseResponse
	AuthTime   int64 `json:"auth_time"`
	ExpireTime int64 `json:"expire_time"`
	Response   struct {
		ShopName     string  `json:"shop_name"`
		Region       string  `json:"region"`
		Status       string  `json:"status"`
		ShopCBSC     string  `json:"shop_cbsc"`
		SIPAffiliate []int64 `json:"sip_a_shops"`
		IsCB         bool    `json:"is_cb"`
		IsCNSC       bool    `json:"is_cnsc"`
	} `json:"response"`
}

// GetShopInfo 获取店铺信息
func (c *Client) GetShopInfo(accessToken string, shopID uint64) (*ShopInfoResponse, error) {
	path := "/api/v2/shop/get_shop_info"

	respBody, err := c.Get(path, nil, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result ShopInfoResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取店铺信息失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
