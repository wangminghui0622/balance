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
	testShopID       = 226516274                          // 替换为订单所属的 shop_id
	testAccessToken  = "64586c50624c42586854425355547079" // 替换为该店铺的 access_token
	testRefreshToken = "536c66464b70596f42754f6863625778" // 替换为该店铺的 refresh_token（用于 token 过期时自动刷新）
	testOrderSN      = "2602183HM8UDVC"                   // 替换为订单号
)

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

// generateTestSign 生成 Shopee API 签名
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
	accessToken := testAccessToken
	refreshToken := testRefreshToken
	if refreshToken != "" {
		fmt.Println(">>> 已配置 refresh_token，token 失效时将自动刷新")
	}

	apiPath := "/api/v2/payment/get_escrow_detail"
	params := url.Values{}
	params.Set("order_sn", testOrderSN)

	fmt.Printf("请求: %s%s\n", testHost, apiPath)

	client := &http.Client{Timeout: 30 * time.Second}
	body, err := doAPIWithRetry(client, http.MethodGet, apiPath, params, nil, &accessToken, &refreshToken)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	// 格式化输出
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(body, &prettyJSON); err != nil {
		fmt.Printf("原始响应: %s\n", string(body))
		fmt.Printf("解析响应失败: %v\n", err)
		return
	}

	formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
	fmt.Printf("响应:\n%s\n", string(formatted))

	// 提取 escrow_amount
	if response, ok := prettyJSON["response"].(map[string]interface{}); ok {
		if orderIncome, ok := response["order_income"].(map[string]interface{}); ok {
			if escrowAmount, ok := orderIncome["escrow_amount"]; ok {
				fmt.Printf("\n>>> escrow_amount = %v <<<\n", escrowAmount)
			}
		}
	}
}

/*
   {
     "error": "",
     "message": "",
     "request_id": "b6419ed44b054f20248dc9bfde735000",
     "response": {
       "buyer_payment_info": {
         "bulky_handling_fee": 0,
         "buyer_paid_extended_warranty": 0,
         "buyer_paid_installation_fee": 0,
         "buyer_payment_method": "Apple Pay",
         "buyer_service_fee": 0,
         "buyer_tax_amount": 0,
         "buyer_total_amount": 4.99,
         "credit_card_promotion": -0,
         "discount_pix": -0,
         "footwear_tax": 0,
         "icms_tax_amount": 0,
         "import_duty_and_excise_tax": 0,
         "import_processing_charge": 0,
         "import_tax_amount": 0,
         "initial_buyer_txn_fee": 0,
         "insurance_premium": 0,
         "iof_tax_amount": 0,
         "is_paid_by_credit_card": false,
         "lvg_sales_tax_adjustment": 0,
         "merchant_subtotal": 3,
         "seller_voucher": -0,
         "shipping_fee": 1.99,
         "shipping_fee_sst_amount": 0,
         "shopee_coins_redeemed": -0,
         "shopee_voucher": -0,
         "total_tax_and_fees_amount": 0,
         "trade_in_bonus": 0,
         "trade_in_discount": 0,
         "vat": 0
       },
       "buyer_user_name": "local_main.sg",
       "order_income": {
         "actual_installation_fee": 0,
         "actual_shipping_fee": 0,
         "ads_escrow_top_up_fee_or_technical_support_fee": 0,
         "buyer_paid_extended_warranty": 0,
         "buyer_paid_shipping_fee": 1.99,
         "buyer_payment_method": "Apple Pay",
         "buyer_total_amount": 4.99,
         "buyer_transaction_fee": 0,
         "campaign_fee": 0,
         "coins": 0,
         "commission_fee": 0.07,
         "cost_of_goods_sold": 3,
         "credit_card_promotion": 0,
         "credit_card_transaction_fee": 0.1,
         "cross_border_tax": 0,
         "delivery_seller_protection_fee_premium_amount": 0,
         "drc_adjustable_refund": 0,
         "escrow_amount": 4.82,
         "escrow_amount_after_adjustment": 4.82,
         "escrow_import_tax": 0,
         "escrow_tax": 0,
         "estimated_shipping_fee": 1.99,
         "fbs_fee": 0,
         "final_escrow_product_gst": 0,
         "final_escrow_shipping_gst": 0,
         "final_product_protection": 0,
         "final_product_vat_tax": 0,
         "final_return_to_seller_shipping_fee": 0,
         "final_shipping_fee": 0,
         "final_shipping_vat_tax": 0,
         "fsf_seller_protection_fee_claim_amount": 0,
         "installation_fee_paid_by_buyer": 0,
         "instalment_plan": "N/A",
         "items": [
           {
             "activity_id": 0,
             "activity_type": "",
             "ams_commission_fee": 0,
             "buyer_paid_extended_warranty": 0,
             "discount_from_coin": 0,
             "discount_from_voucher_seller": 0,
             "discount_from_voucher_shopee": 0,
             "discounted_price": 3,
             "installation_fee_paid_by_buyer": 0,
             "is_b2c_shop_item": false,
             "is_main_item": false,
             "item_id": 844117783,
             "item_name": "123434",
             "item_sku": "",
             "model_id": 10006252414,
             "model_name": "blue",
             "model_sku": "",
             "original_price": 3,
             "quantity_purchased": 3,
             "seller_discount": 0,
             "seller_order_processing_fee": 0,
             "selling_price": 3,
             "shopee_discount": 0
           }
         ],
         "order_ams_commission_fee": 0,
         "order_chargeable_weight": 0,
         "order_discounted_price": 3,
         "order_original_price": 3,
         "order_seller_discount": 0,
         "order_selling_price": 3,
         "original_cost_of_goods_sold": 3,
         "original_price": 3,
         "original_shopee_discount": 0,
         "overseas_return_service_fee": 0,
         "payment_promotion": 0,
         "pix_discount": 0,
         "prorated_coins_value_offset_return_items": 0,
         "prorated_payment_channel_promo_bank_offset_return_items": 0,
         "prorated_payment_channel_promo_shopee_offset_return_items": 0,
         "prorated_pix_discount_offset_return_items": 0,
         "prorated_seller_voucher_offset_return_items": 0,
         "prorated_shopee_voucher_offset_return_items": 0,
         "return_to_seller_shipping_fee_sst": 0,
         "reverse_shipping_fee": 0,
         "reverse_shipping_fee_sst": 0,
         "rsf_seller_protection_fee_claim_amount": 0,
         "sales_tax_on_lvg": 0,
         "seller_coin_cash_back": 0,
         "seller_discount": 0,
         "seller_lost_compensation": 0,
         "seller_order_processing_fee": 0,
         "seller_return_refund": 0,
         "seller_shipping_discount": 0,
         "seller_transaction_fee": 0.1,
         "seller_voucher_code": [],
         "service_fee": 0,
         "shipping_fee_discount_from_3pl": 0,
         "shipping_fee_sst": 0,
         "shipping_seller_protection_fee_amount": 0,
         "shopee_discount": 0,
         "shopee_shipping_rebate": 0,
         "tax_registration_code": "",
         "tenure_info_list": [
           {
             "instalment_plan": "N/A"
           }
         ],
         "th_import_duty": 0,
         "total_adjustment_amount": 0,
         "trade_in_bonus_by_seller": 0,
         "vat_on_imported_goods": 0,
         "voucher_from_seller": 0,
         "voucher_from_shopee": 0,
         "withholding_pit_tax": 0,
         "withholding_tax": 0,
         "withholding_vat_tax": 0
       },
       "order_sn": "2602170GS576N1",
       "return_order_sn_list": []
     }
   }
*/
