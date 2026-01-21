package controllers

import (
	"balance/internal/config"
	"balance/internal/utils"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ShopeeAuthController 处理虾皮授权回调和换取 access_token
type ShopeeAuthController struct {
	cfg *config.Config
}

// NewShopeeAuthController 创建 Shopee 授权控制器
func NewShopeeAuthController(cfg *config.Config) *ShopeeAuthController {
	return &ShopeeAuthController{cfg: cfg}
}

// GenerateAuthURL 生成虾皮店铺授权链接，方便在浏览器里直接打开
// GET /api/v1/balance/admin/shopee/auth/url
func (ctrl *ShopeeAuthController) GenerateAuthURL(c *gin.Context) {
	if ctrl.cfg.ShopeePartnerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "Shopee Partner 配置不完整，请先在 config.yaml 中配置 shopee.partner_id",
		})
		return
	}

	// 获取回调地址：优先使用配置文件中的 redirect，否则使用当前请求的域名动态生成
	// 注意：Java 项目中 redirect 配置只有域名（如 https://www.balanpay.com），路径在代码中动态添加
	var baseRedirectURL string
	var redirectDomainOnly string // 只包含域名，不包含协议（如 kx9y.com）
	if ctrl.cfg.ShopeeRedirect != "" {
		// 使用配置文件中的 redirect（可能包含协议，如 https://kx9y.com）
		baseRedirectURL = ctrl.cfg.ShopeeRedirect
		
		// 解析出纯域名部分（不包含协议）
		parsedURL, err := url.Parse(baseRedirectURL)
		if err == nil && parsedURL.Host != "" {
			redirectDomainOnly = parsedURL.Host
		} else {
			// 如果没有协议，直接使用
			redirectDomainOnly = baseRedirectURL
		}
		
		log.Printf("使用配置文件中的 redirect: %s", baseRedirectURL)
		log.Printf("提取的纯域名（不含协议）: %s", redirectDomainOnly)
	} else {
		// 动态生成回调地址的域名部分
		scheme := "https"
		if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
			scheme = "http"
		}
		host := c.Request.Host
		baseRedirectURL = scheme + "://" + host
		redirectDomainOnly = host
		log.Printf("动态生成回调地址域名: %s", baseRedirectURL)
	}
	
	// 构建完整的回调 URL（添加路径）
	callbackURL := baseRedirectURL + "/api/v1/balance/admin/shopee/auth/callback"

	// 沙箱或正式环境的授权地址（与 Java 项目保持一致）
	baseAuthURL := "https://partner.shopeemobile.com/api/v2/shop/auth_partner"
	if ctrl.cfg.ShopeeIsSandbox {
		baseAuthURL = "https://openplatform.sandbox.test-stable.shopee.cn/api/v2/shop/auth_partner"
	}

	// 生成 timestamp 和 sign
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)
	
	// Shopee 授权链接的签名规则（根据 Java 项目和官方文档）：
	// 签名字符串 = partnerId + apiPath + timestamp（不包含 redirect！）
	// 签名格式 = 64位十六进制小写字符串
	path := "/api/v2/shop/auth_partner"
	
	// 先构建 redirect URL（用于后续签名测试）
	// 注意：这里先不构建完整的 redirect，等签名生成后再构建
	redirectForSign := baseRedirectURL // 先使用域名，后续可能会尝试包含路径
	
	// 方式1：签名字符串 = partnerId + path + timestamp（不包含 redirect）
	// 这是 Shopee 官方文档的标准格式
	signString1 := strconv.FormatInt(ctrl.cfg.ShopeePartnerID, 10) + path + timestampStr
	
	// 方式2：签名字符串 = partnerId + path + timestamp + redirect（包含 redirect）
	// 某些实现可能会包含 redirect，但官方文档说不需要
	signString2 := strconv.FormatInt(ctrl.cfg.ShopeePartnerID, 10) + path + timestampStr + redirectForSign
	
	// 方式3：签名字符串 = partnerId + path + timestamp + redirect（包含完整回调URL）
	signString3 := strconv.FormatInt(ctrl.cfg.ShopeePartnerID, 10) + path + timestampStr + callbackURL
	
	// 方式2已测试报错 "Partner_id is invalid"，说明签名字符串格式可能不对
	// 回退到方式1（不包含 redirect），这是官方标准格式
	
	// 处理 partner_key
	partnerKeyRaw := ctrl.cfg.ShopeePartnerKey
	
	// 尝试三种方式处理 partner_key：
	// 方式A：保留 shpk 前缀，直接作为字符串（与 ExchangeShopeeToken 一致）
	partnerKeyBytesA := []byte(partnerKeyRaw)
	
	// 定义 signature 变量
	var signature string
	
	// 方式B：去掉 shpk 前缀，作为字符串
	partnerKeyWithoutShpk := partnerKeyRaw
	if len(partnerKeyRaw) > 4 && partnerKeyRaw[:4] == "shpk" {
		partnerKeyWithoutShpk = partnerKeyRaw[4:]
		log.Printf("已去掉 partner_key 的 shpk 前缀，处理后: %s", partnerKeyWithoutShpk)
	}
	partnerKeyBytesB := []byte(partnerKeyWithoutShpk)
	
	// 方式C：去掉 shpk 前缀，十六进制解码
	var hexDecoded []byte
	if decoded, err := hex.DecodeString(partnerKeyWithoutShpk); err == nil && len(decoded) > 0 {
		hexDecoded = decoded
		log.Printf("partner_key 十六进制解码成功，长度=%d字节", len(hexDecoded))
	} else {
		log.Printf("⚠️  partner_key 不是有效的十六进制字符串，跳过方式C")
		hexDecoded = nil
	}
	partnerKeyBytesC := hexDecoded
	
	// 签名字符串格式1：partnerId + path + timestamp（不包含 redirect）
	mac1A := hmac.New(sha256.New, partnerKeyBytesA)
	mac1A.Write([]byte(signString1))
	signature1A := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac1A.Sum(nil)))
	
	mac1B := hmac.New(sha256.New, partnerKeyBytesB)
	mac1B.Write([]byte(signString1))
	signature1B := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac1B.Sum(nil)))
	
	var signature1C string
	if partnerKeyBytesC != nil {
		mac1C := hmac.New(sha256.New, partnerKeyBytesC)
		mac1C.Write([]byte(signString1))
		signature1C = fmt.Sprintf("%064x", new(big.Int).SetBytes(mac1C.Sum(nil)))
	}
	
	// 签名字符串格式2：partnerId + path + timestamp + redirect（包含域名）
	mac2A := hmac.New(sha256.New, partnerKeyBytesA)
	mac2A.Write([]byte(signString2))
	signature2A := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac2A.Sum(nil)))
	
	mac2B := hmac.New(sha256.New, partnerKeyBytesB)
	mac2B.Write([]byte(signString2))
	signature2B := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac2B.Sum(nil)))
	
	var signature2C string
	if partnerKeyBytesC != nil {
		mac2C := hmac.New(sha256.New, partnerKeyBytesC)
		mac2C.Write([]byte(signString2))
		signature2C = fmt.Sprintf("%064x", new(big.Int).SetBytes(mac2C.Sum(nil)))
	}
	
	// 签名字符串格式3：partnerId + path + timestamp + redirect（包含完整回调URL）
	mac3A := hmac.New(sha256.New, partnerKeyBytesA)
	mac3A.Write([]byte(signString3))
	signature3A := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac3A.Sum(nil)))
	
	mac3B := hmac.New(sha256.New, partnerKeyBytesB)
	mac3B.Write([]byte(signString3))
	signature3B := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac3B.Sum(nil)))
	
	var signature3C string
	if partnerKeyBytesC != nil {
		mac3C := hmac.New(sha256.New, partnerKeyBytesC)
		mac3C.Write([]byte(signString3))
		signature3C = fmt.Sprintf("%064x", new(big.Int).SetBytes(mac3C.Sum(nil)))
	}
	
	// 尝试：partner_key 方式A（保留 shpk 前缀）+ 签名字符串格式1（不包含 redirect）
	// 因为 ExchangeShopeeToken 中 partner_key 是直接作为字符串使用的，没有去掉 shpk 前缀
	signature = signature1A
	
	log.Printf("========== 签名方式完整对比 ==========")
	log.Printf("partner_key 方式A（保留shpk前缀，UTF-8字符串）:")
	log.Printf("  格式1（不包含redirect）: %s", signature1A)
	log.Printf("  格式2（包含redirect域名）: %s", signature2A)
	log.Printf("  格式3（包含完整回调URL）: %s", signature3A)
	log.Printf("partner_key 方式B（去掉shpk前缀，UTF-8字符串）:")
	log.Printf("  格式1（不包含redirect）: %s", signature1B)
	log.Printf("  格式2（包含redirect域名）: %s", signature2B)
	log.Printf("  格式3（包含完整回调URL）: %s", signature3B)
	if signature1C != "" {
		log.Printf("partner_key 方式C（去掉shpk前缀，十六进制解码）:")
		log.Printf("  格式1（不包含redirect）: %s", signature1C)
		log.Printf("  格式2（包含redirect域名）: %s", signature2C)
		log.Printf("  格式3（包含完整回调URL）: %s", signature3C)
	}
	log.Printf("当前使用: partner_key方式A（保留shpk前缀）+ 格式1（不包含redirect）")
	log.Printf("⚠️  如果还是报错，可以尝试其他组合")
	log.Printf("==========================================")
	
	
	// grantSource 参数（可选，用于标识授权来源）
	grantSource := "1" // 默认值，可以根据需要调整
	
	// 构建授权链接
	// Java URL格式：?partner_id=%s&timestamp=%s&sign=%s&redirect=%s?grantSource=%s
	// 根据 Java 代码分析：
	// - redirect 参数的值是 signBaseParamConfig.getRedirect()，即配置中的 redirect（只有域名，如 https://www.balanpay.com）
	// - grantSource 是作为 URL 的一部分，格式是 redirect=%s?grantSource=%s
	// - 所以最终的 URL 格式是：?partner_id=xxx&timestamp=xxx&sign=xxx&redirect=https://www.balanpay.com?grantSource=1
	// 
	// 构建授权链接
	// Java URL格式：?partner_id=%s&timestamp=%s&sign=%s&redirect=%s?grantSource=%s
	// 
	// 尝试多种方式，找出 Shopee 验证逻辑：
	// 方式1：只使用域名（不含协议），如 kx9y.com
	// 方式2：使用域名+协议，如 https://kx9y.com
	// 方式3：使用完整回调 URL，如 https://kx9y.com/api/v1/balance/admin/shopee/auth/callback
	// 
	// 根据 Java 项目，redirect 参数的值应该是配置中的 redirect（包含协议，如 https://www.balanpay.com）
	// 但 Shopee 验证时可能只验证域名部分，所以先尝试方式2（与 Java 项目一致）
	
	// 方式1：只使用域名（不含协议）
	redirectOption1 := redirectDomainOnly
	authURL1 := fmt.Sprintf("%s?partner_id=%d&timestamp=%s&sign=%s&redirect=%s?grantSource=%s",
		baseAuthURL, ctrl.cfg.ShopeePartnerID, timestampStr, signature,
		url.QueryEscape(redirectOption1), grantSource)
	
	// 方式2：使用域名+协议（与 Java 项目一致）
	redirectOption2 := baseRedirectURL
	authURL2 := fmt.Sprintf("%s?partner_id=%d&timestamp=%s&sign=%s&redirect=%s?grantSource=%s",
		baseAuthURL, ctrl.cfg.ShopeePartnerID, timestampStr, signature,
		url.QueryEscape(redirectOption2), grantSource)
	
	// 方式3：使用完整回调 URL
	redirectOption3 := callbackURL
	authURL3 := fmt.Sprintf("%s?partner_id=%d&timestamp=%s&sign=%s&redirect=%s?grantSource=%s",
		baseAuthURL, ctrl.cfg.ShopeePartnerID, timestampStr, signature,
		url.QueryEscape(redirectOption3), grantSource)
	
	// 默认使用方式3（完整回调URL），因为方式1（只域名）报错 "redirect url is invalid"
	// 如果方式3不行，再尝试方式2（域名+协议）
	authURL := authURL3
	
	log.Printf("========== 三种 redirect 参数方式对比 ==========")
	log.Printf("方式1（只域名，不含协议）: redirect=%s", redirectOption1)
	log.Printf("方式1 URL: %s", authURL1)
	log.Printf("⚠️  方式1已测试：报错 'redirect url is invalid'，说明需要包含协议")
	log.Printf("方式2（域名+协议）: redirect=%s", redirectOption2)
	log.Printf("方式2 URL: %s", authURL2)
	log.Printf("⚠️  方式2已测试：报错 'The domain of redirect is not consistent'")
	log.Printf("方式3（完整回调URL）: redirect=%s", redirectOption3)
	log.Printf("方式3 URL: %s", authURL3)
	log.Printf("⚠️  方式3已测试：也报错 'The domain of redirect is not consistent'")
	log.Printf("================================================")
	log.Printf("")
	log.Printf("========== Shopee 控制台配置检查清单 ==========")
	log.Printf("1. 确认 Shopee 控制台配置的 Redirect URL Domain 格式：")
	log.Printf("   - 如果控制台只允许填域名（如 kx9y.com），则配置应该是: kx9y.com")
	log.Printf("   - 如果控制台允许填完整 URL（如 https://kx9y.com），则配置应该是: https://kx9y.com")
	log.Printf("2. 确认当前使用的域名: %s", redirectDomainOnly)
	log.Printf("3. 确认当前使用的完整域名: %s", baseRedirectURL)
	log.Printf("4. 确认 Shopee 环境: %s", map[bool]string{true: "沙箱环境 (Sandbox)", false: "正式环境 (Production)"}[ctrl.cfg.ShopeeIsSandbox])
	log.Printf("5. 确认 Shopee 控制台配置的环境与代码中的环境一致（沙箱/正式）")
	log.Printf("6. 确认 Shopee 控制台配置已保存并等待 5-10 分钟生效")
	log.Printf("7. 如果控制台配置的是 www.kx9y.com，而代码使用的是 kx9y.com，也会报错")
	log.Printf("8. 如果控制台配置的是 kx9y.com，而代码使用的是 www.kx9y.com，也会报错")
	log.Printf("================================================")
	
	log.Printf("========== Shopee 授权链接生成调试信息 ==========")
	log.Printf("partner_id: %d", ctrl.cfg.ShopeePartnerID)
	log.Printf("⚠️  重要：请确认此 partner_id 与 Shopee 控制台中的配置一致")
	log.Printf("⚠️  控制台显示的 Test Partner_id: 1203446")
	log.Printf("⚠️  控制台显示的 Live Partner_id: 2014300")
	log.Printf("⚠️  如果当前使用的是沙箱环境，partner_id 应该是 1203446")
	log.Printf("⚠️  如果当前使用的是正式环境，partner_id 应该是 2014300")
	log.Printf("timestamp: %s", timestampStr)
	log.Printf("callback: %s", callbackURL)
	
	// 解析 callback URL 的域名部分，用于验证
	callbackURLParsed, err := url.Parse(callbackURL)
	if err == nil {
		log.Printf("callback 域名: %s", callbackURLParsed.Host)
		log.Printf("callback 路径: %s", callbackURLParsed.Path)
		log.Printf("⚠️  请确保 Shopee 控制台配置的 Redirect URL Domain 域名部分与上述域名一致")
		log.Printf("⚠️  控制台应填写: %s (只填域名，不包含协议和路径)", callbackURLParsed.Host)
		log.Printf("⚠️  或者填写: https://%s (包含协议，不包含路径)", callbackURLParsed.Host)
		log.Printf("⚠️  当前使用的完整 redirect URL: %s", callbackURL)
		log.Printf("⚠️  如果控制台配置的是其他域名，请修改 config.yaml 中的 redirect 配置")
	} else {
		log.Printf("⚠️  解析 callback URL 失败: %v", err)
	}
	
	log.Printf("签名字符串 (partner_id+path+timestamp): %s", signString1)
	log.Printf("partner_key 原始值: %s", ctrl.cfg.ShopeePartnerKey)
	log.Printf("partner_key 方式A（保留shpk前缀）长度: %d 字节", len(partnerKeyBytesA))
	log.Printf("partner_key 方式B（去掉shpk前缀）长度: %d 字节", len(partnerKeyBytesB))
	if partnerKeyBytesC != nil {
		log.Printf("partner_key 方式C（十六进制解码）长度: %d 字节", len(partnerKeyBytesC))
	}
	log.Printf("生成的签名: %s (长度: %d)", signature, len(signature))
	log.Printf("⚠️  如果还是报 'Wrong sign'，请尝试其他组合的签名")
	log.Printf("完整授权URL: %s", authURL)
	log.Printf("================================================")

	c.JSON(http.StatusOK, gin.H{
		"code":       200,
		"message":    "生成授权链接成功，在浏览器打开该URL进行授权",
		"auth_url":   authURL,
		"callback":   callbackURL,
		"is_sandbox": ctrl.cfg.ShopeeIsSandbox,
	})
}

// AuthCallback 虾皮授权回调
// 用于接收 code，并调用 Shopee 接口换取 access_token
// 示例回调地址： https://kx9y.com/api/v1/balance/admin/shopee/auth/callback
func (ctrl *ShopeeAuthController) AuthCallback(c *gin.Context) {
	code := c.Query("code")
	shopIDStr := c.Query("shop_id")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "缺少参数 code",
		})
		return
	}

	var shopID int64
	if shopIDStr != "" {
		if id, err := strconv.ParseInt(shopIDStr, 10, 64); err == nil {
			shopID = id
		}
	}
	if shopID == 0 {
		shopID = ctrl.cfg.ShopeeShopID
	}

	if ctrl.cfg.ShopeePartnerID == 0 || ctrl.cfg.ShopeePartnerKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "Shopee Partner 配置不完整，请先在 config.yaml 或 环境变量中配置 partner_id / partner_key",
		})
		return
	}

	// 调用工具函数，向虾皮请求 access_token
	accessToken, refreshToken, expireIn, err := utils.ExchangeShopeeToken(
		ctrl.cfg.ShopeePartnerID,
		ctrl.cfg.ShopeePartnerKey,
		shopID,
		code,
		ctrl.cfg.ShopeeIsSandbox,
	)
	if err != nil {
		log.Printf("向虾皮换取 access_token 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "向虾皮换取 access_token 失败: " + err.Error(),
		})
		return
	}

	// 打日志，方便你在服务日志里直接复制
	log.Printf("Shopee 授权成功: shop_id=%d, access_token=%s, refresh_token=%s, expire_in=%d秒",
		shopID, accessToken, refreshToken, expireIn)

	// 直接返回给前端，方便你复制到 config.yaml
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Shopee 授权成功，请复制 access_token 到 backend/config.yaml",
		"data": gin.H{
			"shop_id":       shopID,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expire_in":     expireIn,
			"expire_at":     time.Now().Add(time.Duration(expireIn) * time.Second).Format(time.RFC3339),
		},
	})
}

