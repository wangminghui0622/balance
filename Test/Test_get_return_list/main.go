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
	"os"
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
	testShopID       = 226516274                          // 替换为店铺的 shop_id
	testAccessToken  = "64586c50624c42586854425355547079" // 替换为该店铺的 access_token
	testRefreshToken = "536c66464b70596f42754f6863625778" // 替换为该店铺的 refresh_token（用于 token 过期时自动刷新）
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
	const maxRetries = 5
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
		if isTransientError(errResp.Error, errResp.Message) {
			if attempt < maxRetries-1 {
				wait := 5 + attempt*3
				fmt.Printf(">>> 临时错误 %s，%d 秒后重试 (%d/%d)\n", errResp.Message, wait, attempt+1, maxRetries)
				time.Sleep(time.Duration(wait) * time.Second)
				continue
			}
			return nil, fmt.Errorf("重试 %d 次后仍失败: %s - %s（可能是 Shopee 沙箱不稳定，可稍后再试）", maxRetries, errResp.Error, errResp.Message)
		}
		if errResp.Error != "" {
			return nil, fmt.Errorf("API 错误: %s - %s", errResp.Error, errResp.Message)
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

	apiPath := "/api/v2/returns/get_return_list"
	now := time.Now()
	getParams := url.Values{}
	getParams.Set("create_time_from", strconv.FormatInt(now.AddDate(0, 0, -30).Unix(), 10))
	getParams.Set("create_time_to", strconv.FormatInt(now.Unix(), 10))
	getParams.Set("page_size", "20")

	fmt.Printf("请求: %s%s\n", testHost, apiPath)

	client := &http.Client{Timeout: 30 * time.Second}
	body, err := doAPIWithRetry(client, http.MethodGet, apiPath, getParams, nil, &accessToken, &refreshToken)
	if err != nil {
		fmt.Printf(">>> 请求失败: %v\n", err)
		os.Exit(1)
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
		if returnList, ok := response["return_list"].([]interface{}); ok {
			fmt.Printf("\n>>> return_list 数量 = %d <<<\n", len(returnList))
		}
		if more, ok := response["more"].(bool); ok {
			fmt.Printf(">>> more = %v <<<\n", more)
		}
	}
}
