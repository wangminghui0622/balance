package shopee

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ReturnListResponse 退货列表响应
type ReturnListResponse struct {
	BaseResponse
	Response struct {
		More       bool   `json:"more"`
		NextCursor string `json:"next_cursor"`
		ReturnList []struct {
			ReturnSN   string `json:"return_sn"`
			OrderSN    string `json:"order_sn"`
			Status     string `json:"status"`
			UpdateTime int64  `json:"update_time"`
		} `json:"return_list"`
	} `json:"response"`
}

// GetReturnList 获取退货列表
func (c *Client) GetReturnList(accessToken string, shopID uint64, createTimeFrom, createTimeTo int64, pageSize int, cursor string) (*ReturnListResponse, error) {
	path := "/api/v2/returns/get_return_list"
	params := url.Values{}
	params.Set("create_time_from", strconv.FormatInt(createTimeFrom, 10))
	params.Set("create_time_to", strconv.FormatInt(createTimeTo, 10))
	params.Set("page_size", strconv.Itoa(pageSize))
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	respBody, err := c.Get(path, params, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result ReturnListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取退货列表失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// ReturnDetailResponse 退货详情响应
type ReturnDetailResponse struct {
	BaseResponse
	Response struct {
		Image            []string `json:"image"`
		Reason           string   `json:"reason"`
		TextReason       string   `json:"text_reason"`
		ReturnSN         string   `json:"return_sn"`
		RefundAmount     float64  `json:"refund_amount"`
		Currency         string   `json:"currency"`
		CreateTime       int64    `json:"create_time"`
		UpdateTime       int64    `json:"update_time"`
		Status           string   `json:"status"`
		DueDate          int64    `json:"due_date"`
		TrackingNumber   string   `json:"tracking_number"`
		NeedsLogistics   bool     `json:"needs_logistics"`
		AmountBeforeDisc float64  `json:"amount_before_discount"`
		OrderSN          string   `json:"order_sn"`
		LogisticsStatus  string   `json:"logistics_status"`
		User             struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Portrait string `json:"portrait"`
		} `json:"user"`
		Item []struct {
			ItemID       int64    `json:"item_id"`
			Name         string   `json:"name"`
			Images       []string `json:"images"`
			Amount       int      `json:"amount"`
			ItemPrice    float64  `json:"item_price"`
			EscrowAmount float64  `json:"escrow_amount"`
			IsAddOnDeal  bool     `json:"is_add_on_deal"`
			IsMainItem   bool     `json:"is_main_item"`
			AddOnDealID  int64    `json:"add_on_deal_id"`
			ItemSKU      string   `json:"item_sku"`
			VariationSKU string   `json:"variation_sku"`
			ModelID      int64    `json:"model_id"`
		} `json:"item"`
		Negotiation struct {
			NegotiationStatus   string  `json:"negotiation_status"`
			LatestSolution      string  `json:"latest_solution"`
			LatestOfferAmount   float64 `json:"latest_offer_amount"`
			LatestOfferCreator  string  `json:"latest_offer_creator"`
			CounterLimit        int     `json:"counter_limit"`
			OfferDueDate        int64   `json:"offer_due_date"`
		} `json:"negotiation"`
		DisputeReason       []string `json:"dispute_reason"`
		DisputeTextReason   []string `json:"dispute_text_reason"`
		ReturnShipDueDate   int64    `json:"return_ship_due_date"`
		ReturnSellerDueDate int64    `json:"return_seller_due_date"`
	} `json:"response"`
}

// GetReturnDetail 获取退货详情
func (c *Client) GetReturnDetail(accessToken string, shopID uint64, returnSN string) (*ReturnDetailResponse, error) {
	path := "/api/v2/returns/get_return_detail"
	params := url.Values{}
	params.Set("return_sn", returnSN)

	respBody, err := c.Get(path, params, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result ReturnDetailResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取退货详情失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
