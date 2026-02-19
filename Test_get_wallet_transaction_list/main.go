package main

import (
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

//a. 虾皮有没有结算给店铺就是根据这个接口，ESCROW_VERIFIED_ADD
//b. finance_incomes表会出现
//	 transaction_id	  order_sn	    transaction_type	  amount
//	 12345678	      240101xxxx	ESCROW_VERIFIED_ADD	  100
//	 12345999	      240101xxxx	ESCROW_ADJUSTMENT	  -5

// ===== 在这里填写你的参数 =====
const (
	testPartnerID  = 1203446
	testPartnerKey = "shpk724b6a656d626b696b756345464e6b614d524664716c61525a4e4e4f466c"
	testHost       = "https://openplatform.sandbox.test-stable.shopee.cn"

	// TODO: 替换为实际值（从数据库 shop_authorizations 表查询）
	testShopID      = 226516274 // 替换为店铺的 shop_id
	testAccessToken = ""        // 替换为该店铺的 access_token
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

	apiPath := "/api/v2/payment/get_wallet_transaction_list"
	timestamp := time.Now().Unix()
	sign := generateTestSign(testPartnerID, apiPath, timestamp, testAccessToken, testShopID, testPartnerKey)

	params := url.Values{}
	params.Set("partner_id", strconv.FormatInt(testPartnerID, 10))
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", sign)
	params.Set("access_token", testAccessToken)
	params.Set("shop_id", strconv.FormatUint(testShopID, 10))
	params.Set("page_no", "1")
	params.Set("page_size", "20")

	reqURL := fmt.Sprintf("%s%s?%s", testHost, apiPath, params.Encode())
	fmt.Printf("请求 URL: %s\n", reqURL)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
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
		if list, ok := response["transaction_list"].([]interface{}); ok {
			fmt.Printf("\n>>> transaction_list 数量 = %d <<<\n", len(list))
		}
		if more, ok := response["more"].(bool); ok {
			fmt.Printf(">>> more = %v <<<\n", more)
		}
	}
}
