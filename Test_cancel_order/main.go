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
	"time"
)

// ===== 在这里填写你的参数 =====
const (
	testPartnerID  = 1203446
	testPartnerKey = "shpk724b6a656d626b696b756345464e6b614d524664716c61525a4e4e4f466c"
	testHost       = "https://openplatform.sandbox.test-stable.shopee.cn"

	// TODO: 替换为实际值（从数据库 shop_authorizations 表查询）
	testShopID      = 226516274 // 替换为店铺的 shop_id
	testAccessToken = ""        // 替换为该店铺的 access_token
	testOrderSN     = ""        // 替换为要取消的订单号（需为 READY_TO_SHIP 等可取消状态）
)

// generateTestSign 生成 Shopee API 签名
func generateTestSign(partnerID int64, path string, timestamp int64, accessToken string, shopID uint64, partnerKey string) string {
	baseStr := fmt.Sprintf("%d%s%d%s%d", partnerID, path, timestamp, accessToken, shopID)
	h := hmac.New(sha256.New, []byte(partnerKey))
	h.Write([]byte(baseStr))
	return hex.EncodeToString(h.Sum(nil))
}

func main() {
	if testShopID == 0 || testAccessToken == "" {
		fmt.Println("请先填写 testShopID 和 testAccessToken（从数据库查询）")
		return
	}
	if testOrderSN == "" {
		fmt.Println("请先填写 testOrderSN（要取消的订单号，需为可取消状态如 READY_TO_SHIP）")
		return
	}

	apiPath := "/api/v2/order/cancel_order"
	timestamp := time.Now().Unix()
	sign := generateTestSign(testPartnerID, apiPath, timestamp, testAccessToken, testShopID, testPartnerKey)

	params := url.Values{}
	params.Set("partner_id", strconv.FormatInt(testPartnerID, 10))
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", sign)
	params.Set("access_token", testAccessToken)
	params.Set("shop_id", strconv.FormatUint(testShopID, 10))

	reqURL := fmt.Sprintf("%s%s?%s", testHost, apiPath, params.Encode())
	fmt.Printf("请求 URL: %s\n", reqURL)

	// POST 请求体
	body := map[string]interface{}{
		"order_sn":     testOrderSN,
		"cancel_reason": "OUT_OF_STOCK", // 可选: OUT_OF_STOCK, CUSTOMER_REQUEST, etc.
	}
	bodyJSON, _ := json.Marshal(body)
	fmt.Printf("请求体: %s\n", string(bodyJSON))

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(bodyJSON))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
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
