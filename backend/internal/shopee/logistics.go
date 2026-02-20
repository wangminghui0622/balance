package shopee

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// ShipOrderResponse 发货响应
type ShipOrderResponse struct {
	BaseResponse
}

// ShipOrderParams 发货参数（pickup / dropoff / non_integrated 三选一）
type ShipOrderParams struct {
	// Pickup 上门取件
	Pickup *struct {
		AddressID    int64  `json:"address_id"`
		PickupTimeID string `json:"pickup_time_id"`
	}
	// Dropoff 自送站点
	Dropoff *struct {
		BranchID string `json:"branch_id"`
		Slug     string `json:"slug"`
	}
	// NonIntegrated 非集成物流（自有/第三方，需填 tracking_no）
	NonIntegrated *struct {
		TrackingNo string `json:"tracking_no"`
	}
}

// ShipOrder 订单发货
// Shopee API 要求且仅允许选择一种发货方式：pickup、dropoff 或 non_integrated
// 若 params 为空，会先调用 GetShippingParameter 并尝试自动选择第一个可用选项（pickup 或 dropoff）
func (c *Client) ShipOrder(accessToken string, shopID uint64, orderSN string, trackingNumber string) (*ShipOrderResponse, error) {
	return c.ShipOrderWithParams(accessToken, shopID, orderSN, nil, trackingNumber)
}

// ShipOrderWithParams 使用完整参数发货
func (c *Client) ShipOrderWithParams(accessToken string, shopID uint64, orderSN string, params *ShipOrderParams, trackingNumber string) (*ShipOrderResponse, error) {
	return c.ShipOrderWithParamsContext(context.Background(), accessToken, shopID, orderSN, params, trackingNumber)
}

// ShipOrderWithParamsContext 带 context 的发货（支持超时与取消）
func (c *Client) ShipOrderWithParamsContext(ctx context.Context, accessToken string, shopID uint64, orderSN string, params *ShipOrderParams, trackingNumber string) (*ShipOrderResponse, error) {
	path := "/api/v2/logistics/ship_order"
	body := c.buildShipOrderBody(orderSN, params, trackingNumber)
	if body == nil {
		// 无有效参数，尝试从 GetShippingParameter 自动获取
		paramResp, err := c.GetShippingParameter(accessToken, shopID, orderSN)
		if err != nil {
			return nil, fmt.Errorf("获取发货参数失败: %w", err)
		}
		body = c.buildShipOrderBodyFromParam(orderSN, paramResp, trackingNumber)
		if body == nil {
			return nil, fmt.Errorf("无法确定发货方式，请先调用 GetShippingParameter 获取地址/站点/物流单号后重试")
		}
	}

	// 发货请求默认 30 秒超时，避免长时间阻塞
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	respBody, err := c.PostWithContext(ctx, path, nil, body, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result ShipOrderResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("发货失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// buildShipOrderBody 根据 params 构建请求体（pickup/dropoff/non_integrated 三选一）
func (c *Client) buildShipOrderBody(orderSN string, params *ShipOrderParams, trackingNumber string) map[string]interface{} {
	if params != nil {
		if params.Pickup != nil && params.Pickup.AddressID > 0 && params.Pickup.PickupTimeID != "" {
			return map[string]interface{}{
				"order_sn": orderSN,
				"pickup": map[string]interface{}{
					"address_id":     params.Pickup.AddressID,
					"pickup_time_id": params.Pickup.PickupTimeID,
				},
			}
		}
		if params.Dropoff != nil && params.Dropoff.BranchID != "" && params.Dropoff.Slug != "" {
			return map[string]interface{}{
				"order_sn": orderSN,
				"dropoff": map[string]interface{}{
					"branch_id": params.Dropoff.BranchID,
					"slug":      params.Dropoff.Slug,
				},
			}
		}
		if params.NonIntegrated != nil {
			return map[string]interface{}{
				"order_sn": orderSN,
				"non_integrated": map[string]interface{}{
					"tracking_no": params.NonIntegrated.TrackingNo,
				},
			}
		}
	}
	if trackingNumber != "" {
		return map[string]interface{}{
			"order_sn": orderSN,
			"non_integrated": map[string]interface{}{
				"tracking_no": trackingNumber,
			},
		}
	}
	return nil
}

// buildShipOrderBodyFromParam 从 GetShippingParameter 响应中自动选择第一个可用选项
func (c *Client) buildShipOrderBodyFromParam(orderSN string, param *GetShippingParameterResponse, trackingNumber string) map[string]interface{} {
	// 若有 tracking_no，优先使用 non_integrated
	if trackingNumber != "" {
		return map[string]interface{}{
			"order_sn": orderSN,
			"non_integrated": map[string]interface{}{
				"tracking_no": trackingNumber,
			},
		}
	}
	// pickup：仅当 info_needed 要求 pickup 时，取第一个地址和第一个时间槽
	if len(param.Response.InfoNeeded.Pickup) > 0 && len(param.Response.Pickup.AddressList) > 0 {
		addr := param.Response.Pickup.AddressList[0]
		if len(addr.TimeSlotList) > 0 {
			return map[string]interface{}{
				"order_sn": orderSN,
				"pickup": map[string]interface{}{
					"address_id":     addr.AddressID,
					"pickup_time_id": addr.TimeSlotList[0].PickupTimeID,
				},
			}
		}
	}
	// dropoff：仅当 info_needed 要求 dropoff 时，取第一个 branch 和第一个 slug
	if len(param.Response.InfoNeeded.Dropoff) > 0 && len(param.Response.Dropoff.BranchList) > 0 && len(param.Response.Dropoff.SlugList) > 0 {
		branch := param.Response.Dropoff.BranchList[0].Branch
		slug := param.Response.Dropoff.SlugList[0].Slug
		return map[string]interface{}{
			"order_sn": orderSN,
			"dropoff": map[string]interface{}{
				"branch_id": branch,
				"slug":      slug,
			},
		}
	}
	// non_integrated 需用户提供 tracking_no
	return nil
}

// GetShippingParameterResponse 获取发货参数响应
type GetShippingParameterResponse struct {
	BaseResponse
	Response struct {
		InfoNeeded struct {
			Dropoff       []string `json:"dropoff"`
			Pickup        []string `json:"pickup"`
			NonIntegrated []string `json:"non_integrated"`
		} `json:"info_needed"`
		Dropoff struct {
			BranchList []struct {
				Branch   string `json:"branch"`
				Address  string `json:"address"`
				City     string `json:"city"`
				District string `json:"district"`
				State    string `json:"state"`
				Zipcode  string `json:"zipcode"`
			} `json:"branch_list"`
			SlugList []struct {
				Slug     string `json:"slug"`
				SlugName string `json:"slug_name"`
			} `json:"slug_list"`
		} `json:"dropoff"`
		Pickup struct {
			AddressList []struct {
				AddressID    int64    `json:"address_id"`
				Region       string   `json:"region"`
				State        string   `json:"state"`
				City         string   `json:"city"`
				District     string   `json:"district"`
				Town         string   `json:"town"`
				Address      string   `json:"address"`
				Zipcode      string   `json:"zipcode"`
				AddressFlag  []string `json:"address_flag"`
				TimeSlotList []struct {
					Date         string `json:"date"`
					TimeText     string `json:"time_text"`
					PickupTimeID string `json:"pickup_time_id"`
				} `json:"time_slot_list"`
			} `json:"address_list"`
		} `json:"pickup"`
	} `json:"response"`
}

// GetShippingParameter 获取发货参数
func (c *Client) GetShippingParameter(accessToken string, shopID uint64, orderSN string) (*GetShippingParameterResponse, error) {
	path := "/api/v2/logistics/get_shipping_parameter"
	params := url.Values{}
	params.Set("order_sn", orderSN)

	respBody, err := c.Get(path, params, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result GetShippingParameterResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取发货参数失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// LogisticsChannelListResponse 物流渠道列表响应
type LogisticsChannelListResponse struct {
	BaseResponse
	Response struct {
		LogisticsChannelList []struct {
			LogisticsChannelID   int64  `json:"logistics_channel_id"`
			LogisticsChannelName string `json:"logistics_channel_name"`
			CODEnabled           bool   `json:"cod_enabled"`
			Enabled              bool   `json:"enabled"`
		} `json:"logistics_channel_list"`
	} `json:"response"`
}

// GetLogisticsChannelList 获取物流渠道列表
func (c *Client) GetLogisticsChannelList(accessToken string, shopID uint64) (*LogisticsChannelListResponse, error) {
	path := "/api/v2/logistics/get_channel_list"

	respBody, err := c.Get(path, nil, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result LogisticsChannelListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取物流渠道列表失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetTrackingNumberResponse 获取运单号响应
type GetTrackingNumberResponse struct {
	BaseResponse
	Response struct {
		TrackingNumber          string `json:"tracking_number"`
		PlpNumber               string `json:"plp_number"`
		FirstMileTrackingNumber string `json:"first_mile_tracking_number"`
		LastMileTrackingNumber  string `json:"last_mile_tracking_number"`
	} `json:"response"`
}

// GetTrackingNumber 获取运单号
func (c *Client) GetTrackingNumber(accessToken string, shopID uint64, orderSN string) (*GetTrackingNumberResponse, error) {
	path := "/api/v2/logistics/get_tracking_number"
	params := url.Values{}
	params.Set("order_sn", orderSN)

	respBody, err := c.Get(path, params, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result GetTrackingNumberResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取运单号失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
