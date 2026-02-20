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

// ===== 在这里填写你的参数 =====
const (
	testPartnerID  = 1203446
	testPartnerKey = "shpk724b6a656d626b696b756345464e6b614d524664716c61525a4e4e4f466c"
	testHost       = "https://openplatform.sandbox.test-stable.shopee.cn"

	// TODO: 替换为实际值（从数据库 shop_authorizations 表查询）
	testShopID       = 226516274                                                                                                  // 替换为店铺的 shop_id
	testAccessToken  = "eyJhbGciOiJIUzI1NiJ9.CPa5SRABGLK6gWwgASiPjuHMBjCnjOiNDDgBQAE.7fyYH0XTIsUmZKi8pPoS-g9VVylROKrUSOaEtJX1DYY" // 替换为该店铺的 access_token
	testRefreshToken = "eyJhbGciOiJIUzI1NiJ9.CPa5SRABGLK6gWwgAiiPjuHMBjClvfKBDjgBQAE.vKHS9plKPQM_qfVHimqfzcVjpXutdO6yZj_8iiNxtXI" // 替换为该店铺的 refresh_token（用于 token 过期时自动刷新）
	testOrderSN      = "2602183HM8UDVC"                                                                                           // 替换为要取消的订单号（需为 READY_TO_SHIP 等可取消状态）
)

// generateTestSign 生成 Shopee API 签名（业务 API，需 access_token + shop_id）
func generateTestSign(partnerID int64, path string, timestamp int64, accessToken string, shopID uint64, partnerKey string) string {
	baseStr := fmt.Sprintf("%d%s%d%s%d", partnerID, path, timestamp, accessToken, shopID)
	h := hmac.New(sha256.New, []byte(partnerKey))
	h.Write([]byte(baseStr))
	return hex.EncodeToString(h.Sum(nil))
}

// generateAuthSign 生成认证 API 签名（auth 接口无需 access_token）
func generateAuthSign(partnerID int64, path string, timestamp int64, partnerKey string) string {
	baseStr := fmt.Sprintf("%d%s%d", partnerID, path, timestamp)
	h := hmac.New(sha256.New, []byte(partnerKey))
	h.Write([]byte(baseStr))
	return hex.EncodeToString(h.Sum(nil))
}

// refreshAccessToken 调用 Shopee 刷新 token 接口
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

// doAPI 发起 Shopee API 请求
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

// 是否为 token 无效类错误（Shopee 有时返回 invalid_acceess_token 拼写错误）
func isTokenInvalid(errCode string) bool {
	return errCode == "invalid_access_token" || errCode == "invalid_acceess_token"
}

func isTransientError(errCode, message string) bool {
	if errCode == "error_data" {
		return true
	}
	return strings.Contains(message, "try later") || strings.Contains(message, "Inner error")
}

// doAPIWithRetry 发起请求，若返回 token 无效则自动刷新后重试；若遇临时错误则延迟重试
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

// orderItem 订单商品（用于 cancel_order 的 item_list）
type orderItem struct {
	ItemID  int64 `json:"item_id"`
	ModelID int64 `json:"model_id"`
}

// getOrderDetailResp 订单详情响应
type getOrderDetailResp struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		OrderList []struct {
			OrderSN  string `json:"order_sn"`
			ItemList []struct {
				ItemID  int64 `json:"item_id"`
				ModelID int64 `json:"model_id"`
			} `json:"item_list"`
		} `json:"order_list"`
	} `json:"response"`
}

func main() {
	if testShopID == 0 || testAccessToken == "" {
		fmt.Println("请先填写 testShopID 和 testAccessToken（从数据库 shop_authorizations 表查询）")
		return
	}
	if testOrderSN == "" {
		fmt.Println("请先填写 testOrderSN（要取消的订单号，需为可取消状态如 READY_TO_SHIP）")
		return
	}
	accessToken := testAccessToken
	refreshToken := testRefreshToken
	if refreshToken != "" {
		fmt.Println(">>> 已配置 refresh_token，token 失效时将自动刷新")
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// Step 1: 调用 get_order_detail 获取订单商品列表（cancel_order 必须传 item_list）
	fmt.Println("\n>>> Step 1: 获取订单详情（含 item_list）")
	getDetailPath := "/api/v2/order/get_order_detail"
	getParams := url.Values{}
	getParams.Set("order_sn_list", testOrderSN)
	getParams.Set("response_optional_fields", strings.Join([]string{"order_sn", "item_list"}, ","))

	getRespBody, err := doAPIWithRetry(client, http.MethodGet, getDetailPath, getParams, nil, &accessToken, &refreshToken)
	if err != nil {
		fmt.Printf("获取订单详情失败: %v\n", err)
		return
	}

	var detailResp getOrderDetailResp
	if err := json.Unmarshal(getRespBody, &detailResp); err != nil {
		fmt.Printf("解析订单详情失败: %v\n", err)
		fmt.Printf("原始响应: %s\n", string(getRespBody))
		return
	}
	if detailResp.Error != "" {
		fmt.Printf("获取订单详情 API 错误: %s - %s\n", detailResp.Error, detailResp.Message)
		return
	}

	var itemList []orderItem
	for _, o := range detailResp.Response.OrderList {
		if o.OrderSN == testOrderSN {
			for _, it := range o.ItemList {
				itemList = append(itemList, orderItem{ItemID: it.ItemID, ModelID: it.ModelID})
			}
			break
		}
	}
	if len(itemList) == 0 {
		fmt.Println("未找到该订单或订单无商品，无法取消")
		return
	}
	fmt.Printf("订单包含 %d 个商品，将取消整单\n", len(itemList))

	// Step 2: 调用 cancel_order，必须传入 item_list
	fmt.Println("\n>>> Step 2: 调用 cancel_order")
	cancelPath := "/api/v2/order/cancel_order"
	cancelBody := map[string]interface{}{
		"order_sn":      testOrderSN,
		"cancel_reason": "OUT_OF_STOCK",
		"item_list":     itemList,
	}
	bodyJSON, _ := json.Marshal(cancelBody)
	fmt.Printf("请求体: %s\n", string(bodyJSON))

	respBody, err := doAPIWithRetry(client, http.MethodPost, cancelPath, nil, cancelBody, &accessToken, &refreshToken)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(respBody, &prettyJSON); err != nil {
		fmt.Printf("原始响应: %s\n", string(respBody))
		fmt.Printf("解析响应失败: %v\n", err)
		return
	}

	formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
	fmt.Printf("响应:\n%s\n", string(formatted))

	if errMsg, ok := prettyJSON["error"].(string); ok && errMsg != "" {
		fmt.Printf("\n>>> 错误: %s <<<\n", errMsg)
	} else {
		fmt.Println("\n>>> 取消订单请求已发送 <<<")
	}
}
