package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// get_order_detail 用于获取订单详情，含订单状态、商品列表、收货地址、物流等
// 参考文档: https://open.shopee.com/documents/v2/v2.order.get_order_detail
// 参数: order_sn_list(逗号分隔，最多50个), response_optional_fields(可选返回字段)

// ===== 在这里填写你的参数 =====
const (
	testPartnerID  = 1203446
	testPartnerKey = "shpk724b6a656d626b696b756345464e6b614d524664716c61525a4e4e4f466c"
	testHost       = "https://openplatform.sandbox.test-stable.shopee.cn"

	// TODO: 替换为实际值（从数据库 shop_authorizations 表查询）
	testShopID       = 226516274                          // 替换为店铺的 shop_id
	testAccessToken  = "6357667356615958434e674c49794b6c" // 替换为该店铺的 access_token
	testRefreshToken = "624e5445684f6b7667787161414a5463" // 替换为该店铺的 refresh_token（用于 token 过期时自动刷新）
	testOrderSN      = "2602183HM8UDVC"                   // 替换为订单号（可从 get_order_list 获取）
)

// 260201J845FR4T
func generateAuthSign(partnerID int64, path string, timestamp int64, partnerKey string) string {
	baseStr := fmt.Sprintf("%d%s%d", partnerID, path, timestamp)
	h := hmac.New(sha256.New, []byte(partnerKey))
	h.Write([]byte(baseStr))
	return hex.EncodeToString(h.Sum(nil))
}

func refreshAccessToken(client *http.Client, refreshToken string) (accessToken, newRefreshToken string, err error) {
	path := "/api/v2/auth/access_token/get"
	timestamp := time.Now().Unix()
	sign := generateAuthSign(testPartnerID, path, timestamp, testPartnerKey)
	params := url.Values{}
	params.Set("partner_id", strconv.FormatInt(testPartnerID, 10))
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", sign)
	reqURL := fmt.Sprintf("%s%s?%s", testHost, path, params.Encode())
	body := map[string]interface{}{
		"refresh_token": refreshToken,
		"partner_id":    testPartnerID,
		"shop_id":       testShopID,
	}
	jsonData, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(jsonData))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	var result struct {
		Error           string `json:"error"`
		Message         string `json:"message"`
		AccessToken     string `json:"access_token"`
		RefreshToken    string `json:"refresh_token"`
		ExpireIn        int64  `json:"expire_in"`
		RefreshExpireIn int64  `json:"refresh_token_expire_in"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("解析刷新响应失败: %w", err)
	}
	if result.Error != "" {
		return "", "", fmt.Errorf("刷新 Token 失败: %s - %s", result.Error, result.Message)
	}
	return result.AccessToken, result.RefreshToken, nil
}

func generateTestSign(partnerID int64, path string, timestamp int64, accessToken string, shopID uint64, partnerKey string) string {
	baseStr := fmt.Sprintf("%d%s%d%s%d", partnerID, path, timestamp, accessToken, shopID)
	h := hmac.New(sha256.New, []byte(partnerKey))
	h.Write([]byte(baseStr))
	return hex.EncodeToString(h.Sum(nil))
}

func doAPI(client *http.Client, method, apiPath string, extraParams url.Values, body interface{}, accessToken string) ([]byte, error) {
	timestamp := time.Now().Unix()
	sign := generateTestSign(testPartnerID, apiPath, timestamp, accessToken, testShopID, testPartnerKey)
	params := url.Values{}
	params.Set("partner_id", strconv.FormatInt(testPartnerID, 10))
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", sign)
	params.Set("access_token", accessToken)
	params.Set("shop_id", strconv.FormatUint(testShopID, 10))
	for k, v := range extraParams {
		params[k] = v
	}
	reqURL := fmt.Sprintf("%s%s?%s", testHost, apiPath, params.Encode())
	var reqBody io.Reader
	if body != nil {
		jsonData, _ := json.Marshal(body)
		reqBody = bytes.NewReader(jsonData)
	}
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func isTokenInvalid(errCode string) bool {
	return errCode == "invalid_access_token" || errCode == "invalid_acceess_token"
}

func isTransientError(errCode, message string) bool {
	if errCode == "error_data" {
		return true
	}
	return strings.Contains(message, "try later") || strings.Contains(message, "Inner error")
}

func doAPIWithRetry(client *http.Client, method, apiPath string, extraParams url.Values, body interface{}, accessToken, refreshToken *string) ([]byte, error) {
	const maxRetries = 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		respBody, err := doAPI(client, method, apiPath, extraParams, body, *accessToken)
		if err != nil {
			return nil, err
		}
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(respBody, &errResp)
		if isTokenInvalid(errResp.Error) && *refreshToken != "" {
			fmt.Println(">>> access_token 已失效，正在刷新...")
			newAccess, newRefresh, err := refreshAccessToken(client, *refreshToken)
			if err != nil {
				return nil, fmt.Errorf("刷新 Token 失败: %w", err)
			}
			*accessToken = newAccess
			*refreshToken = newRefresh
			fmt.Printf(">>> Token 已刷新（access_token 前8位: %s...）\n", newAccess)
			return doAPI(client, method, apiPath, extraParams, body, *accessToken)
		}
		if isTransientError(errResp.Error, errResp.Message) && attempt < maxRetries-1 {
			wait := 2 + attempt
			fmt.Printf(">>> 临时错误 %s，%d 秒后重试 (%d/%d)\n", errResp.Message, wait, attempt+1, maxRetries)
			time.Sleep(time.Duration(wait) * time.Second)
			continue
		}
		return respBody, nil
	}
	return nil, fmt.Errorf("重试 %d 次后仍失败", maxRetries)
}

func main() {
	if testShopID == 0 || testAccessToken == "" {
		fmt.Println("请先填写 testShopID 和 testAccessToken（从数据库 shop_authorizations 表查询）")
		return
	}
	if testOrderSN == "" {
		fmt.Println("请先填写 testOrderSN（可从 get_order_list 获取）")
		return
	}
	accessToken := testAccessToken
	refreshToken := testRefreshToken
	if refreshToken != "" {
		fmt.Println(">>> 已配置 refresh_token，token 失效时将自动刷新")
	}

	apiPath := "/api/v2/order/get_order_detail"
	params := url.Values{}
	params.Set("order_sn_list", testOrderSN)
	params.Set("response_optional_fields", strings.Join([]string{
		"order_sn", "region", "currency", "total_amount", "order_status",
		"shipping_carrier", "payment_method", "create_time", "update_time",
		"recipient_address", "item_list", "package_list",
	}, ","))

	fmt.Printf("请求: %s%s（order_sn=%s）\n", testHost, apiPath, testOrderSN)

	client := &http.Client{Timeout: 30 * time.Second}
	body, err := doAPIWithRetry(client, http.MethodGet, apiPath, params, nil, &accessToken, &refreshToken)
	if err != nil {
		fmt.Printf(">>> 请求失败: %v\n", err)
		return
	}

	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(body, &prettyJSON); err != nil {
		fmt.Printf("原始响应: %s\n", string(body))
		fmt.Printf("解析响应失败: %v\n", err)
		return
	}

	formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
	fmt.Printf("响应:\n%s\n", string(formatted))

	if response, ok := prettyJSON["response"].(map[string]interface{}); ok {
		if orderList, ok := response["order_list"].([]interface{}); ok && len(orderList) > 0 {
			if order, ok := orderList[0].(map[string]interface{}); ok {
				fmt.Println("\n>>> 关键字段 <<<")
				fmt.Printf("  order_sn = %v\n", order["order_sn"])
				fmt.Printf("  order_status = %v\n", order["order_status"])
				fmt.Printf("  total_amount = %v\n", order["total_amount"])
				fmt.Printf("  currency = %v\n", order["currency"])
				if items, ok := order["item_list"].([]interface{}); ok {
					fmt.Printf("  item_list 数量 = %d\n", len(items))
					for i, it := range items {
						if i >= 3 {
							fmt.Printf("  ... 共 %d 个商品\n", len(items))
							break
						}
						if m, ok := it.(map[string]interface{}); ok {
							fmt.Printf("  [%d] item_id=%v model_id=%v %v x%v\n", i+1, m["item_id"], m["model_id"], m["model_name"], m["model_quantity_purchased"])
						}
					}
				}
			}
		}
	}
}
