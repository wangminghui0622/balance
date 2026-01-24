package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ShopeeAPIClient 虾皮API客户端
type ShopeeAPIClient struct {
	PartnerID      int64
	PartnerKey     string
	ShopID         int64
	AccessToken    string
	RefreshToken   string
	TokenExpireAt  time.Time // access_token 过期时间
	BaseURL        string    // 沙箱: https://openplatform.sandbox.test-stable.shopee.cn, 正式: https://partner.shopeemobile.com
	OnTokenRefresh func(accessToken, refreshToken string, expireIn int64) // token 刷新回调函数，用于更新配置
}

// NewShopeeAPIClient 创建虾皮API客户端
func NewShopeeAPIClient(partnerID int64, partnerKey string, shopID int64, accessToken string, isSandbox bool) *ShopeeAPIClient {
	return NewShopeeAPIClientWithRefresh(partnerID, partnerKey, shopID, accessToken, "", time.Time{}, isSandbox, nil)
}

// NewShopeeAPIClientWithRefresh 创建虾皮API客户端（支持自动刷新）
func NewShopeeAPIClientWithRefresh(partnerID int64, partnerKey string, shopID int64, accessToken, refreshToken string, tokenExpireAt time.Time, isSandbox bool, onTokenRefresh func(accessToken, refreshToken string, expireIn int64)) *ShopeeAPIClient {
	baseURL := "https://partner.shopeemobile.com"
	if isSandbox {
		baseURL = "https://openplatform.sandbox.test-stable.shopee.cn"
	}
	return &ShopeeAPIClient{
		PartnerID:     partnerID,
		PartnerKey:    partnerKey,
		ShopID:        shopID,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		TokenExpireAt: tokenExpireAt,
		BaseURL:       baseURL,
		OnTokenRefresh: onTokenRefresh,
	}
}

// generateSignature 生成API签名
// 根据Java项目参考实现，虾皮API签名规则:
// base_string = partner_id + path + timestamp + access_token + shop_id (当有access_token时)
// base_string = partner_id + path + timestamp (当没有access_token时)
// 然后使用HMAC-SHA256计算签名
func (c *ShopeeAPIClient) generateSignature(path string, params map[string]string) string {
	// 从params中获取timestamp，如果不存在则使用当前时间
	timestamp, exists := params["timestamp"]
	if !exists {
		timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	}
	
	// 构建基础签名字符串
	var baseString string
	if c.AccessToken != "" {
		// 有access_token的API: partner_id + path + timestamp + access_token + shop_id
		baseString = fmt.Sprintf("%d%s%s%s%d",
			c.PartnerID,
			path,
			timestamp,
			c.AccessToken,
			c.ShopID,
		)
	} else {
		// 没有access_token的API: partner_id + path + timestamp
		baseString = fmt.Sprintf("%d%s%s",
			c.PartnerID,
			path,
			timestamp,
		)
	}
	
	// HMAC-SHA256签名
	mac := hmac.New(sha256.New, []byte(c.PartnerKey))
	mac.Write([]byte(baseString))
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
	// 确保 token 有效
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	path := "/api/v2/order/get_order_list"
	timestamp := time.Now().Unix()
	
	// 构建参数 (without sign and timestamp yet)
	params := map[string]string{
		"partner_id":      strconv.FormatInt(c.PartnerID, 10),
		"shop_id":         strconv.FormatInt(c.ShopID, 10),
		"access_token":    c.AccessToken,
		"time_range_field": timeRangeField,
		"time_from":       strconv.FormatInt(timeFrom, 10),
		"time_to":         strconv.FormatInt(timeTo, 10),
		"page_size":       strconv.Itoa(pageSize),
	}
	
	if cursor != "" {
		params["cursor"] = cursor
	}
	
	// Add timestamp to params for signature generation
	params["timestamp"] = strconv.FormatInt(timestamp, 10)
	
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
	// 注意：partner_key 应该保留 shpk 前缀（与 GenerateAuthURL 一致）
	signString := fmt.Sprintf("%d%s%d", partnerID, path, timestamp)
	mac := hmac.New(sha256.New, []byte(partnerKey))
	mac.Write([]byte(signString))
	// 使用 %064x 确保签名是64位小写十六进制（与 GenerateAuthURL 一致）
	signature := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac.Sum(nil)))

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

	log.Printf("========== 调用虾皮换取 access_token ==========")
	log.Printf("URL: %s", requestURL)
	log.Printf("Body: %s", string(bodyBytes))
	log.Printf("签名字符串: %s", signString)
	log.Printf("partner_key 长度: %d 字节", len(partnerKey))
	log.Printf("生成的签名: %s (长度: %d)", signature, len(signature))
	log.Printf("=============================================")

	// 创建带超时的 HTTP Client（30秒超时）
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 重试机制：最多重试3次
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			log.Printf("第 %d 次重试...", attempt)
			// 重试前等待1秒
			time.Sleep(1 * time.Second)
			// 更新 timestamp 和 signature
			timestamp = time.Now().Unix()
			signString = fmt.Sprintf("%d%s%d", partnerID, path, timestamp)
			mac = hmac.New(sha256.New, []byte(partnerKey))
			mac.Write([]byte(signString))
			signature = fmt.Sprintf("%064x", new(big.Int).SetBytes(mac.Sum(nil)))
			query.Set("timestamp", strconv.FormatInt(timestamp, 10))
			query.Set("sign", signature)
			requestURL = fmt.Sprintf("%s%s?%s", baseURL, path, query.Encode())
		}

		req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(string(bodyBytes)))
		if err != nil {
			return "", "", 0, fmt.Errorf("创建请求失败: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			if attempt < maxRetries {
				log.Printf("请求失败（第 %d 次尝试）: %v，将重试...", attempt, err)
				continue
			}
			return "", "", 0, fmt.Errorf("请求失败（已重试 %d 次）: %v", maxRetries, err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			if attempt < maxRetries {
				log.Printf("读取响应失败（第 %d 次尝试）: %v，将重试...", attempt, err)
				continue
			}
			return "", "", 0, fmt.Errorf("读取响应失败（已重试 %d 次）: %v", maxRetries, err)
		}

		log.Printf("虾皮换取access_token响应（HTTP %d）: %s", resp.StatusCode, string(respBody))

		var result struct {
			Error            string `json:"error"`
			Message          string `json:"message"`
			AccessToken      string `json:"access_token"`
			RefreshToken     string `json:"refresh_token"`
			ExpireIn         int64  `json:"expire_in"`
			RefreshExpireIn  int64  `json:"refresh_token_expire_in"`
		}

		if err := json.Unmarshal(respBody, &result); err != nil {
			return "", "", 0, fmt.Errorf("解析JSON失败: %v", err)
		}

		if result.Error != "" && result.Error != "0" {
			return "", "", 0, fmt.Errorf("虾皮返回错误: %s, message=%s", result.Error, result.Message)
		}

		if result.AccessToken == "" {
			return "", "", 0, fmt.Errorf("虾皮返回的 access_token 为空")
		}

		log.Printf("✅ 成功获取 access_token（长度: %d）", len(result.AccessToken))
		return result.AccessToken, result.RefreshToken, result.ExpireIn, nil
	}

	return "", "", 0, fmt.Errorf("请求失败：已达到最大重试次数 %d", maxRetries)
}

// RefreshShopeeToken 使用 refresh_token 刷新 access_token
// partnerID, partnerKey: 从开放平台获取
// shopID: 店铺ID
// refreshToken: 刷新令牌
// isSandbox: 是否沙箱环境
func RefreshShopeeToken(partnerID int64, partnerKey string, shopID int64, refreshToken string, isSandbox bool) (accessToken, newRefreshToken string, expireIn int64, err error) {
	baseURL := "https://partner.shopeemobile.com"
	if isSandbox {
		baseURL = "https://openplatform.sandbox.test-stable.shopee.cn"
	}

	path := "/api/v2/auth/token/refresh"
	timestamp := time.Now().Unix()

	// 签名: sign = HMAC-SHA256(partner_key, partner_id + path + timestamp)
	// 注意：partner_key 应该保留 shpk 前缀（与 GenerateAuthURL 一致）
	signString := fmt.Sprintf("%d%s%d", partnerID, path, timestamp)
	mac := hmac.New(sha256.New, []byte(partnerKey))
	mac.Write([]byte(signString))
	// 使用 %064x 确保签名是64位小写十六进制（与 GenerateAuthURL 一致）
	signature := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac.Sum(nil)))

	// 构建URL
	query := url.Values{}
	query.Set("partner_id", strconv.FormatInt(partnerID, 10))
	query.Set("timestamp", strconv.FormatInt(timestamp, 10))
	query.Set("sign", signature)

	requestURL := fmt.Sprintf("%s%s?%s", baseURL, path, query.Encode())

	// 请求体
	bodyMap := map[string]interface{}{
		"refresh_token": refreshToken,
		"shop_id":       shopID,
		"partner_id":    partnerID,
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	log.Printf("========== 调用虾皮刷新 access_token ==========")
	log.Printf("URL: %s", requestURL)
	log.Printf("Body: %s", string(bodyBytes))
	log.Printf("签名字符串: %s", signString)
	log.Printf("生成的签名: %s (长度: %d)", signature, len(signature))
	log.Printf("=============================================")

	// 创建带超时的 HTTP Client（30秒超时）
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 重试机制：最多重试3次
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			log.Printf("第 %d 次重试...", attempt)
			// 重试前等待1秒
			time.Sleep(1 * time.Second)
			// 更新 timestamp 和 signature
			timestamp = time.Now().Unix()
			signString = fmt.Sprintf("%d%s%d", partnerID, path, timestamp)
			mac = hmac.New(sha256.New, []byte(partnerKey))
			mac.Write([]byte(signString))
			signature = fmt.Sprintf("%064x", new(big.Int).SetBytes(mac.Sum(nil)))
			query.Set("timestamp", strconv.FormatInt(timestamp, 10))
			query.Set("sign", signature)
			requestURL = fmt.Sprintf("%s%s?%s", baseURL, path, query.Encode())
		}

		req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(string(bodyBytes)))
		if err != nil {
			return "", "", 0, fmt.Errorf("创建请求失败: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			if attempt < maxRetries {
				log.Printf("请求失败（第 %d 次尝试）: %v，将重试...", attempt, err)
				continue
			}
			return "", "", 0, fmt.Errorf("请求失败（已重试 %d 次）: %v", maxRetries, err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			if attempt < maxRetries {
				log.Printf("读取响应失败（第 %d 次尝试）: %v，将重试...", attempt, err)
				continue
			}
			return "", "", 0, fmt.Errorf("读取响应失败（已重试 %d 次）: %v", maxRetries, err)
		}

		log.Printf("虾皮刷新access_token响应（HTTP %d）: %s", resp.StatusCode, string(respBody))

		var result struct {
			Error            string `json:"error"`
			Message          string `json:"message"`
			AccessToken      string `json:"access_token"`
			RefreshToken     string `json:"refresh_token"`
			ExpireIn         int64  `json:"expire_in"`
			RefreshExpireIn  int64  `json:"refresh_token_expire_in"`
		}

		if err := json.Unmarshal(respBody, &result); err != nil {
			return "", "", 0, fmt.Errorf("解析JSON失败: %v", err)
		}

		if result.Error != "" && result.Error != "0" {
			return "", "", 0, fmt.Errorf("虾皮返回错误: %s, message=%s", result.Error, result.Message)
		}

		if result.AccessToken == "" {
			return "", "", 0, fmt.Errorf("虾皮返回的 access_token 为空")
		}

		log.Printf("✅ 成功刷新 access_token（长度: %d）", len(result.AccessToken))
		return result.AccessToken, result.RefreshToken, result.ExpireIn, nil
	}

	return "", "", 0, fmt.Errorf("请求失败：已达到最大重试次数 %d", maxRetries)
}

// ensureValidToken 确保 access_token 有效，如果过期则自动刷新
func (c *ShopeeAPIClient) ensureValidToken() error {
	// 如果没有 refresh_token，无法自动刷新
	if c.RefreshToken == "" {
		// 如果没有设置过期时间，假设 token 有效
		if c.TokenExpireAt.IsZero() {
			return nil
		}
		// 如果已过期但没有 refresh_token，返回错误
		if time.Now().After(c.TokenExpireAt) {
			return fmt.Errorf("access_token 已过期且没有 refresh_token，请重新授权")
		}
		return nil
	}

	// 检查 token 是否即将过期（提前5分钟刷新）
	refreshTime := c.TokenExpireAt.Add(-5 * time.Minute)
	if c.TokenExpireAt.IsZero() || time.Now().After(refreshTime) {
		log.Printf("access_token 即将过期或已过期，开始自动刷新...")
		
		// 判断是沙箱还是正式环境
		isSandbox := strings.Contains(c.BaseURL, "sandbox")
		
		// 刷新 token
		newAccessToken, newRefreshToken, expireIn, err := RefreshShopeeToken(
			c.PartnerID,
			c.PartnerKey,
			c.ShopID,
			c.RefreshToken,
			isSandbox,
		)
		if err != nil {
			return fmt.Errorf("自动刷新 access_token 失败: %v", err)
		}

		// 更新 token
		c.AccessToken = newAccessToken
		if newRefreshToken != "" {
			c.RefreshToken = newRefreshToken
		}
		c.TokenExpireAt = time.Now().Add(time.Duration(expireIn) * time.Second)

		log.Printf("✅ access_token 已自动刷新，新 token 有效期至: %s", c.TokenExpireAt.Format(time.RFC3339))

		// 调用回调函数，通知外部更新配置
		if c.OnTokenRefresh != nil {
			c.OnTokenRefresh(newAccessToken, c.RefreshToken, expireIn)
		}
	}

	return nil
}

// GetOrderDetail 获取订单详情
func (c *ShopeeAPIClient) GetOrderDetail(orderSnList []string) (map[string]interface{}, error) {
	// 确保 token 有效
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	path := "/api/v2/order/get_order_detail"
	timestamp := time.Now().Unix()
	
	// 构建参数 (without sign and timestamp yet)
	params := map[string]string{
		"partner_id":   strconv.FormatInt(c.PartnerID, 10),
		"shop_id":      strconv.FormatInt(c.ShopID, 10),
		"access_token": c.AccessToken,
		"order_sn_list": strings.Join(orderSnList, ","),
	}
	
	// Add timestamp to params for signature generation
	params["timestamp"] = strconv.FormatInt(timestamp, 10)
	
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

// GetShopInfo 获取店铺信息
func (c *ShopeeAPIClient) GetShopInfo() (map[string]interface{}, error) {
	// 确保 token 有效
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	path := "/api/v2/shop/get_shop_info"
	timestamp := time.Now().Unix()
	
	// 构建参数 (without sign and timestamp yet)
	params := map[string]string{
		"partner_id":   strconv.FormatInt(c.PartnerID, 10),
		"shop_id":      strconv.FormatInt(c.ShopID, 10),
		"access_token": c.AccessToken,
	}
	
	// Add timestamp to params for signature generation
	params["timestamp"] = strconv.FormatInt(timestamp, 10)
	
	// 生成签名
	signature := c.generateSignature(path, params)
	params["sign"] = signature
	
	// 构建URL
	queryValues := url.Values{}
	for k, v := range params {
		queryValues.Set(k, v)
	}
	requestURL := fmt.Sprintf("%s%s?%s", c.BaseURL, path, queryValues.Encode())
	
	log.Printf("调用虾皮API获取店铺信息: %s", requestURL)
	
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
	
	log.Printf("虾皮API店铺信息响应: %+v", result)
	
	return result, nil
}
