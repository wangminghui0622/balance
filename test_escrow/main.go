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
	"strconv"
	"testing"
	"time"
)

// ===== 在这里填写你的参数 =====
const (
	testPartnerID  = 1203446
	testPartnerKey = "shpk724b6a656d626b696b756345464e6b614d524664716c61525a4e4e4f466c"
	testHost       = "https://openplatform.sandbox.test-stable.shopee.cn"

	// TODO: 替换为实际值（从数据库 shop_authorizations 表查询）
	testShopID      = 0  // 替换为订单所属的 shop_id
	testAccessToken = "" // 替换为该店铺的 access_token
	testOrderSN     = "2602170GS576N1"
)

// generateTestSign 生成 Shopee API 签名
func generateTestSign(partnerID int64, path string, timestamp int64, accessToken string, shopID uint64, partnerKey string) string {
	baseStr := fmt.Sprintf("%d%s%d%s%d", partnerID, path, timestamp, accessToken, shopID)
	h := hmac.New(sha256.New, []byte(partnerKey))
	h.Write([]byte(baseStr))
	return hex.EncodeToString(h.Sum(nil))
}

// TestGetEscrowDetail 测试获取订单结算明细
// 运行方式: cd backend && go test ./internal/shopee/ -run TestGetEscrowDetail -v
func TestGetEscrowDetail(t *testing.T) {
	if testShopID == 0 || testAccessToken == "" {
		t.Skip("请先填写 testShopID 和 testAccessToken（从数据库查询）")
	}

	apiPath := "/api/v2/payment/get_escrow_detail"
	timestamp := time.Now().Unix()
	sign := generateTestSign(testPartnerID, apiPath, timestamp, testAccessToken, testShopID, testPartnerKey)

	// 构造请求参数
	params := url.Values{}
	params.Set("partner_id", strconv.FormatInt(testPartnerID, 10))
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", sign)
	params.Set("access_token", testAccessToken)
	params.Set("shop_id", strconv.FormatUint(testShopID, 10))
	params.Set("order_sn", testOrderSN)

	reqURL := fmt.Sprintf("%s%s?%s", testHost, apiPath, params.Encode())
	t.Logf("请求 URL: %s", reqURL)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("读取响应失败: %v", err)
	}

	// 格式化输出
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(body, &prettyJSON); err != nil {
		t.Logf("原始响应: %s", string(body))
		t.Fatalf("解析响应失败: %v", err)
	}

	formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
	t.Logf("响应:\n%s", string(formatted))

	// 提取 escrow_amount
	if response, ok := prettyJSON["response"].(map[string]interface{}); ok {
		if orderIncome, ok := response["order_income"].(map[string]interface{}); ok {
			if escrowAmount, ok := orderIncome["escrow_amount"]; ok {
				t.Logf("\n>>> escrow_amount = %v <<<", escrowAmount)
			}
		}
	}
}
