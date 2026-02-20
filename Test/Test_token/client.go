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
	"strings"
	"time"
)

// Client 虾皮API客户端
type Client struct {
	partnerID  int64
	partnerKey string
	host       string
	httpClient *http.Client
}

// NewClient 创建虾皮API客户端
func NewClient(region string) *Client {
	cfg := Get().Shopee
	return &Client{
		partnerID:  cfg.PartnerID,
		partnerKey: cfg.PartnerKey,
		host:       cfg.GetHost(region),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetHost 获取当前客户端的API Host（用于调试日志）
func (c *Client) GetHost() string {
	return c.host
}

func (c *Client) generateSign(path string, timestamp int64, accessToken string, shopID uint64) string {
	var baseStr string
	if accessToken != "" && shopID > 0 {
		baseStr = fmt.Sprintf("%d%s%d%s%d", c.partnerID, path, timestamp, accessToken, shopID)
	} else {
		baseStr = fmt.Sprintf("%d%s%d", c.partnerID, path, timestamp)
	}

	h := hmac.New(sha256.New, []byte(c.partnerKey))
	h.Write([]byte(baseStr))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) buildCommonParams(timestamp int64, sign string, accessToken string, shopID uint64) url.Values {
	params := url.Values{}
	params.Set("partner_id", strconv.FormatInt(c.partnerID, 10))
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", sign)
	if accessToken != "" {
		params.Set("access_token", accessToken)
	}
	if shopID > 0 {
		params.Set("shop_id", strconv.FormatUint(shopID, 10))
	}
	return params
}

func (c *Client) doRequest(method, path string, params url.Values, body interface{}, accessToken string, shopID uint64) ([]byte, error) {
	timestamp := time.Now().Unix()
	sign := c.generateSign(path, timestamp, accessToken, shopID)

	commonParams := c.buildCommonParams(timestamp, sign, accessToken, shopID)
	for k, v := range params {
		commonParams[k] = v
	}

	urlStr := fmt.Sprintf("%s%s?%s", c.host, path, commonParams.Encode())

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, urlStr, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("执行请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return respBody, nil
}

// Get 执行GET请求
func (c *Client) Get(path string, params url.Values, accessToken string, shopID uint64) ([]byte, error) {
	return c.doRequest(http.MethodGet, path, params, nil, accessToken, shopID)
}

// Post 执行POST请求
func (c *Client) Post(path string, params url.Values, body interface{}, accessToken string, shopID uint64) ([]byte, error) {
	return c.doRequest(http.MethodPost, path, params, body, accessToken, shopID)
}

// BaseResponse 基础响应结构
type BaseResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}
