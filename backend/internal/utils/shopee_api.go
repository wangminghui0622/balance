package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ShopeeAPIClient 虾皮API客户端
type ShopeeAPIClient struct {
	PartnerID   int64
	PartnerKey  string
	ShopID      int64
	AccessToken string
	BaseURL     string // 沙箱: https://openplatform.sandbox.test-stable.shopee.cn, 正式: https://partner.shopeemobile.com
}

// NewShopeeAPIClient 创建虾皮API客户端
func NewShopeeAPIClient(partnerID int64, partnerKey string, shopID int64, accessToken string, isSandbox bool) *ShopeeAPIClient {
	baseURL := "https://partner.shopeemobile.com"
	if isSandbox {
		baseURL = "https://openplatform.sandbox.test-stable.shopee.cn"
	}
	return &ShopeeAPIClient{
		PartnerID:   partnerID,
		PartnerKey:  partnerKey,
		ShopID:      shopID,
		AccessToken: accessToken,
		BaseURL:     baseURL,
	}
}

// generateSignature 生成API签名
// 虾皮API签名规则: partner_id + path + timestamp + access_token + shop_id + 其他参数（按key排序）
func (c *ShopeeAPIClient) generateSignature(path string, params map[string]string) string {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	
	// 排除sign参数，按key排序
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	
	// 构建签名字符串: partner_id + path + timestamp + access_token + shop_id + 其他参数（按key排序）
	signString := fmt.Sprintf("%d%s%s%s%d",
		c.PartnerID,
		path,
		timestamp,
		c.AccessToken,
		c.ShopID,
	)
	
	// 添加其他参数（按key排序）
	for _, k := range keys {
		signString += k + "=" + params[k]
	}
	
	// HMAC-SHA256签名
	mac := hmac.New(sha256.New, []byte(c.PartnerKey))
	mac.Write([]byte(signString))
	signature := hex.EncodeToString(mac.Sum(nil))
	
	return signature
}

// GetOrderList 获取订单列表
// timeRangeField: 时间字段类型 (create_time/update_time)
// timeFrom: 开始时间戳
// timeTo: 结束时间戳
// pageSize: 每页数量 (最大100)
// cursor: 分页游标
func (c *ShopeeAPIClient) GetOrderList(timeRangeField string, timeFrom, timeTo int64, pageSize int, cursor string) (map[string]interface{}, error) {
	path := "/api/v2/order/get_order_list"
	timestamp := time.Now().Unix()
	
	// 构建参数
	params := map[string]string{
		"partner_id":      strconv.FormatInt(c.PartnerID, 10),
		"shop_id":         strconv.FormatInt(c.ShopID, 10),
		"access_token":    c.AccessToken,
		"timestamp":       strconv.FormatInt(timestamp, 10),
		"time_range_field": timeRangeField,
		"time_from":       strconv.FormatInt(timeFrom, 10),
		"time_to":         strconv.FormatInt(timeTo, 10),
		"page_size":       strconv.Itoa(pageSize),
	}
	
	if cursor != "" {
		params["cursor"] = cursor
	}
	
	// 生成签名
	signature := c.generateSignature(path, params)
	params["sign"] = signature
	
	// 构建URL
	queryValues := url.Values{}
	for k, v := range params {
		queryValues.Set(k, v)
	}
	requestURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, queryValues.Encode())
	
	log.Printf("调用虾皮API: %s", requestURL)
	
	// 发送GET请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	
	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v, 响应: %s", err, string(body))
	}
	
	log.Printf("虾皮API响应: %+v", result)
	
	return result, nil
}

// ExchangeShopeeToken 使用授权 code 向虾皮换取 access_token / refresh_token
// partnerID, partnerKey: 从开放平台获取
// shopID: 店铺ID（回调参数或配置）
// code: 回调URL中的授权码
// isSandbox: 是否沙箱环境
func ExchangeShopeeToken(partnerID int64, partnerKey string, shopID int64, code string, isSandbox bool) (accessToken, refreshToken string, expireIn int64, err error) {
	baseURL := "https://partner.shopeemobile.com"
	if isSandbox {
		baseURL = "https://openplatform.sandbox.test-stable.shopee.cn"
	}

	path := "/api/v2/auth/token/get"
	timestamp := time.Now().Unix()

	// 签名: sign = HMAC-SHA256(partner_key, partner_id + path + timestamp)
	signString := fmt.Sprintf("%d%s%d", partnerID, path, timestamp)
	mac := hmac.New(sha256.New, []byte(partnerKey))
	mac.Write([]byte(signString))
	signature := hex.EncodeToString(mac.Sum(nil))

	// 构建URL
	query := url.Values{}
	query.Set("partner_id", strconv.FormatInt(partnerID, 10))
	query.Set("timestamp", strconv.FormatInt(timestamp, 10))
	query.Set("sign", signature)

	requestURL := fmt.Sprintf("%s%s?%s", baseURL, path, query.Encode())

	// 请求体
	bodyMap := map[string]interface{}{
		"code":       code,
		"shop_id":    shopID,
		"partner_id": partnerID,
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	log.Printf("调用虾皮换取access_token: %s, body=%s", requestURL, string(bodyBytes))

	req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", "", 0, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", 0, fmt.Errorf("读取响应失败: %v", err)
	}

	log.Printf("虾皮换取access_token响应: %s", string(respBody))

	var result struct {
		Error          string `json:"error"`
		Message        string `json:"message"`
		AccessToken    string `json:"access_token"`
		RefreshToken   string `json:"refresh_token"`
		ExpireIn       int64  `json:"expire_in"`
		RefreshExpireIn int64 `json:"refresh_token_expire_in"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", 0, fmt.Errorf("解析JSON失败: %v", err)
	}

	if result.Error != "" && result.Error != "0" {
		return "", "", 0, fmt.Errorf("虾皮返回错误: %s, message=%s", result.Error, result.Message)
	}

	return result.AccessToken, result.RefreshToken, result.ExpireIn, nil
}

// GetOrderDetail 获取订单详情
func (c *ShopeeAPIClient) GetOrderDetail(orderSnList []string) (map[string]interface{}, error) {
	path := "/api/v2/order/get_order_detail"
	timestamp := time.Now().Unix()
	
	// 构建参数
	params := map[string]string{
		"partner_id":   strconv.FormatInt(c.PartnerID, 10),
		"shop_id":      strconv.FormatInt(c.ShopID, 10),
		"access_token": c.AccessToken,
		"timestamp":    strconv.FormatInt(timestamp, 10),
		"order_sn_list": strings.Join(orderSnList, ","),
	}
	
	// 生成签名
	signature := c.generateSignature(path, params)
	params["sign"] = signature
	
	// 构建URL
	queryValues := url.Values{}
	for k, v := range params {
		queryValues.Set(k, v)
	}
	requestURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, queryValues.Encode())
	
	log.Printf("调用虾皮API获取订单详情: %s", requestURL)
	
	// 发送GET请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	
	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v, 响应: %s", err, string(body))
	}
	
	log.Printf("虾皮API订单详情响应: %+v", result)
	
	return result, nil
}
