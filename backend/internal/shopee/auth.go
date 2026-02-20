package shopee

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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
