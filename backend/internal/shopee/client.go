package shopee

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"balance/backend/internal/config"
)

// Client 虾皮API客户端
type Client struct {
	partnerID  int64
	partnerKey string
	host       string
	httpClient *http.Client
}

// NewClient 创建虾皮API客户端
func NewClient(region string) *Client {
	cfg := config.Get().Shopee
	return &Client{
		partnerID:  cfg.PartnerID,
		partnerKey: cfg.PartnerKey,
		host:       cfg.GetHost(region),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetHost 获取当前客户端的API Host（用于调试日志）
func (c *Client) GetHost() string {
	return c.host
}

func (c *Client) generateSign(path string, timestamp int64, accessToken string, shopID uint64) string {
	var baseStr string
	if accessToken != "" && shopID > 0 {
		baseStr = fmt.Sprintf("%d%s%d%s%d", c.partnerID, path, timestamp, accessToken, shopID)
	} else {
		baseStr = fmt.Sprintf("%d%s%d", c.partnerID, path, timestamp)
	}

	h := hmac.New(sha256.New, []byte(c.partnerKey))
	h.Write([]byte(baseStr))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) buildCommonParams(timestamp int64, sign string, accessToken string, shopID uint64) url.Values {
	params := url.Values{}
	params.Set("partner_id", strconv.FormatInt(c.partnerID, 10))
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", sign)
	if accessToken != "" {
		params.Set("access_token", accessToken)
	}
	if shopID > 0 {
		params.Set("shop_id", strconv.FormatUint(shopID, 10))
	}
	return params
}

func (c *Client) doRequest(method, path string, params url.Values, body interface{}, accessToken string, shopID uint64) ([]byte, error) {
	timestamp := time.Now().Unix()
	sign := c.generateSign(path, timestamp, accessToken, shopID)

	commonParams := c.buildCommonParams(timestamp, sign, accessToken, shopID)
	for k, v := range params {
		commonParams[k] = v
	}

	urlStr := fmt.Sprintf("%s%s?%s", c.host, path, commonParams.Encode())

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, urlStr, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("执行请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return respBody, nil
}

// Get 执行GET请求
func (c *Client) Get(path string, params url.Values, accessToken string, shopID uint64) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, params, nil, accessToken, shopID)
}

// Post 执行POST请求
func (c *Client) Post(path string, params url.Values, body interface{}, accessToken string, shopID uint64) ([]byte, error) {
	return c.doRequest(http.MethodPost, path, params, body, accessToken, shopID)
}

// GetAuthURL 获取授权URL
func (c *Client) GetAuthURL(redirectURL string, state string) string {
	timestamp := time.Now().Unix()
	path := "/api/v2/shop/auth_partner"
	sign := c.generateSign(path, timestamp, "", 0)

	params := url.Values{}
	params.Set("partner_id", strconv.FormatInt(c.partnerID, 10))
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", sign)
	params.Set("redirect", redirectURL)
	if state != "" {
		params.Set("state", state)
	}

	return fmt.Sprintf("%s%s?%s", c.host, path, params.Encode())
}

// BaseResponse 基础响应结构
type BaseResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// TokenResponse 获取Token响应
type TokenResponse struct {
	BaseResponse
	AccessToken     string  `json:"access_token"`
	RefreshToken    string  `json:"refresh_token"`
	ExpireIn        int64   `json:"expire_in"`
	RefreshExpireIn int64   `json:"refresh_token_expire_in"`
	PartnerID       int64   `json:"partner_id"`
	ShopIDList      []int64 `json:"shop_id_list"`
	MerchantIDList  []int64 `json:"merchant_id_list"`
}

// GetAccessToken 使用授权码获取AccessToken
func (c *Client) GetAccessToken(code string, shopID uint64) (*TokenResponse, error) {
	path := "/api/v2/auth/token/get"
	timestamp := time.Now().Unix()
	sign := c.generateSign(path, timestamp, "", 0)

	body := map[string]interface{}{
		"code":       code,
		"partner_id": c.partnerID,
		"shop_id":    shopID,
	}

	params := url.Values{}
	params.Set("partner_id", strconv.FormatInt(c.partnerID, 10))
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", sign)

	urlStr := fmt.Sprintf("%s%s?%s", c.host, path, params.Encode())
	jsonData, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TokenResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取Token失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// RefreshAccessToken 刷新AccessToken
func (c *Client) RefreshAccessToken(refreshToken string, shopID uint64) (*TokenResponse, error) {
	path := "/api/v2/auth/access_token/get"
	timestamp := time.Now().Unix()
	sign := c.generateSign(path, timestamp, "", 0)

	body := map[string]interface{}{
		"refresh_token": refreshToken,
		"partner_id":    c.partnerID,
		"shop_id":       shopID,
	}

	params := url.Values{}
	params.Set("partner_id", strconv.FormatInt(c.partnerID, 10))
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", sign)

	urlStr := fmt.Sprintf("%s%s?%s", c.host, path, params.Encode())
	jsonData, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TokenResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("刷新Token失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	BaseResponse
	Response struct {
		More       bool   `json:"more"`
		NextCursor string `json:"next_cursor"`
		OrderList  []struct {
			OrderSN string `json:"order_sn"`
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

	fmt.Printf("[SyncDebug][API] GetOrderList 请求: host=%s shop_id=%d time_from=%d time_to=%d cursor=%q\n",
		c.host, shopID, timeFrom, timeTo, cursor)

	respBody, err := c.Get(path, params, accessToken, shopID)
	if err != nil {
		fmt.Printf("[SyncDebug][API] GetOrderList 请求失败: %v\n", err)
		return nil, err
	}

	// [调试日志] 打印原始响应（截断到500字符）
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

// ShipOrderResponse 发货响应
type ShipOrderResponse struct {
	BaseResponse
}

// ShipOrder 订单发货
func (c *Client) ShipOrder(accessToken string, shopID uint64, orderSN string, trackingNumber string) (*ShipOrderResponse, error) {
	path := "/api/v2/logistics/ship_order"

	body := map[string]interface{}{
		"order_sn":        orderSN,
		"tracking_number": trackingNumber,
	}

	respBody, err := c.Post(path, nil, body, accessToken, shopID)
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

// EscrowDetailResponse 订单结算明细响应
type EscrowDetailResponse struct {
	BaseResponse
	Response struct {
		OrderSN     string `json:"order_sn"`
		BuyerUserID int64  `json:"buyer_user_id"`
		OrderIncome struct {
			EscrowAmount                    float64 `json:"escrow_amount"`
			BuyerTotalAmount                float64 `json:"buyer_total_amount"`
			OriginalPrice                   float64 `json:"original_price"`
			SellerDiscount                  float64 `json:"seller_discount"`
			ShopeeDiscount                  float64 `json:"shopee_discount"`
			VoucherFromSeller               float64 `json:"voucher_from_seller"`
			VoucherFromShopee               float64 `json:"voucher_from_shopee"`
			Coins                           float64 `json:"coins"`
			BuyerPaidShippingFee            float64 `json:"buyer_paid_shipping_fee"`
			BuyerTransactionFee             float64 `json:"buyer_transaction_fee"`
			CrossBorderTax                  float64 `json:"cross_border_tax"`
			PaymentPromotion                float64 `json:"payment_promotion"`
			CommissionFee                   float64 `json:"commission_fee"`
			ServiceFee                      float64 `json:"service_fee"`
			SellerTransactionFee            float64 `json:"seller_transaction_fee"`
			SellerLostCompensation          float64 `json:"seller_lost_compensation"`
			SellerCoinCashBack              float64 `json:"seller_coin_cash_back"`
			EscrowTax                       float64 `json:"escrow_tax"`
			FinalShippingFee                float64 `json:"final_shipping_fee"`
			ActualShippingFee               float64 `json:"actual_shipping_fee"`
			ShippingFeeDiscountFrom3PL      float64 `json:"shopee_shipping_rebate"`
			SellerShippingDiscount          float64 `json:"seller_shipping_discount"`
			EstimatedShippingFee            float64 `json:"estimated_shipping_fee"`
			SellerVoucherCode               []string `json:"seller_voucher_code"`
			DrcAdjustableRefund             float64 `json:"drc_adjustable_refund"`
			CostOfGoodsSold                 float64 `json:"cost_of_goods_sold"`
			OriginalCostOfGoodsSold         float64 `json:"original_cost_of_goods_sold"`
			OriginalShopeeDiscount          float64 `json:"original_shopee_discount"`
			SellerReturnRefund              float64 `json:"seller_return_refund"`
			ItemsCount                      int     `json:"items_count"`
			ReverseShippingFee              float64 `json:"reverse_shipping_fee"`
			FinalProductProtection          float64 `json:"final_product_protection"`
			CreditCardPromotion             float64 `json:"credit_card_promotion"`
			CreditCardTransactionFee        float64 `json:"credit_card_transaction_fee"`
		} `json:"order_income"`
		Items []struct {
			ItemID                    int64   `json:"item_id"`
			ItemName                  string  `json:"item_name"`
			ItemSKU                   string  `json:"item_sku"`
			ModelID                   int64   `json:"model_id"`
			ModelName                 string  `json:"model_name"`
			ModelSKU                  string  `json:"model_sku"`
			OriginalPrice             float64 `json:"original_price"`
			DiscountedPrice           float64 `json:"discounted_price"`
			SellerDiscount            float64 `json:"seller_discount"`
			ShopeeDiscount            float64 `json:"shopee_discount"`
			DiscountFromCoin          float64 `json:"discount_from_coin"`
			DiscountFromVoucher       float64 `json:"discount_from_voucher"`
			DiscountFromVoucherSeller float64 `json:"discount_from_voucher_seller"`
			DiscountFromVoucherShopee float64 `json:"discount_from_voucher_shopee"`
			ActivityType              string  `json:"activity_type"`
			ActivityID                int64   `json:"activity_id"`
			QuantityPurchased         int     `json:"quantity_purchased"`
		} `json:"items"`
	} `json:"response"`
}

// WalletTransactionResponse 钱包交易记录响应
type WalletTransactionResponse struct {
	BaseResponse
	Response struct {
		More            bool          `json:"more"`
		TransactionList []Transaction `json:"transaction_list"`
	} `json:"response"`
}

// Transaction 交易记录
type Transaction struct {
	TransactionID      int64   `json:"transaction_id"`
	Status             string  `json:"status"`
	WalletType         string  `json:"wallet_type"`
	TransactionType    string  `json:"transaction_type"`
	Amount             float64 `json:"amount"`
	CurrentBalance     float64 `json:"current_balance"`
	CreateTime         int64   `json:"create_time"`
	Reason             string  `json:"reason"`
	OrderSN            string  `json:"order_sn"`
	RefundSN           string  `json:"refund_sn"`
	WithdrawalType     string  `json:"withdrawal_type"`
	TransactionFee     float64 `json:"transaction_fee"`
	Description        string  `json:"description"`
	BuyerName          string  `json:"buyer_name"`
	WithdrawalID       int64   `json:"withdrawal_id"`
	TransactionTabType string  `json:"transaction_tab_type"`
	MoneyFlow          string  `json:"money_flow"`
}

// GetWalletTransactionList 获取钱包交易记录列表
func (c *Client) GetWalletTransactionList(accessToken string, shopID uint64, pageNo, pageSize int, walletType string) (*WalletTransactionResponse, error) {
	path := "/api/v2/payment/get_wallet_transaction_list"
	params := url.Values{}
	params.Set("page_no", strconv.Itoa(pageNo))
	params.Set("page_size", strconv.Itoa(pageSize))
	if walletType != "" {
		params.Set("wallet_type", walletType)
	}

	respBody, err := c.Get(path, params, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result WalletTransactionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取钱包交易记录失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetEscrowDetail 获取订单结算明细
func (c *Client) GetEscrowDetail(accessToken string, shopID uint64, orderSN string) (*EscrowDetailResponse, error) {
	path := "/api/v2/payment/get_escrow_detail"
	params := url.Values{}
	params.Set("order_sn", orderSN)

	respBody, err := c.Get(path, params, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result EscrowDetailResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取结算明细失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
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
