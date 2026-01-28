package controllers

import (
	"balance/admin/services"
	"balance/internal/config"
	"balance/internal/models"
	"balance/internal/utils"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ShopeeAuthController 处理虾皮授权回调和换取 access_token
type ShopeeAuthController struct {
	orderService *services.OrderService
	authService  *services.AuthService
	cfg          *config.Config
	db           *gorm.DB
	redisClient  *redis.Client
}

// NewShopeeAuthController 创建 Shopee 授权控制器
func NewShopeeAuthController(cfg *config.Config, db *gorm.DB, authService *services.AuthService, orderService *services.OrderService, redisClient *redis.Client) *ShopeeAuthController {
	return &ShopeeAuthController{cfg: cfg, db: db, authService: authService, orderService: orderService, redisClient: redisClient}
}

// getShopeeConfig 从 shopee_shops 表获取店铺配置（已废弃，保留用于兼容）
// 现在应该直接使用 ShopeeShopRepository
func (ctrl *ShopeeAuthController) getShopeeConfig(shopID int64) (*models.ShopeeShop, error) {
	shopRepo := models.NewShopeeShopRepository(ctrl.db)
	shop, err := shopRepo.GetByShopID(shopID)
	if err == nil && shop != nil {
		return shop, nil
	}
	// 如果数据库中没有，尝试从配置文件获取（向后兼容）
	if ctrl.cfg.ShopeePartnerID > 0 && ctrl.cfg.ShopeePartnerKey != "" {
		// 返回一个临时的 ShopeeShop 对象（不保存到数据库）
		return &models.ShopeeShop{
			ShopID:    shopID,
			PartnerID: ctrl.cfg.ShopeePartnerID,
			// PartnerKey 不在 ShopeeShop 表中，需要从 AuthConfig 获取
		}, nil
	}
	return nil, fmt.Errorf("未找到 shop_id=%d 的 Shopee 配置", shopID)
}
func (ctrl *ShopeeAuthController) GenerateAuthURL(c *gin.Context) {
	var r struct {
		PartnerID  int64  `json:"partnerID"`
		PartnerKey string `json:"partnerKey"`
		IsSandbox  bool   `json:"isSandbox"`
		Redirect   string `json:"redirect"`
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "success",
		})
		return
	}
	var (
		baseRedirectURL string
	)
	if r.Redirect != "" {
		baseRedirectURL = r.Redirect
	} else {
		scheme := "https"
		if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
			scheme = "http"
		}
		host := c.Request.Host
		baseRedirectURL = scheme + "://" + host
	}
	// 构建完整的回调 URL（添加路径）
	callbackURL := baseRedirectURL + "/api/v1/balance/admin/shopee/auth/callback"
	// 沙箱或正式环境的授权地址（与 Java 项目保持一致）
	baseAuthURL := "https://partner.shopeemobile.com/api/v2/shop/auth_partner"
	if r.IsSandbox {
		baseAuthURL = "https://openplatform.sandbox.test-stable.shopee.cn/api/v2/shop/auth_partner"
	}
	// 生成 timestamp 和 sign
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)
	path := "/api/v2/shop/auth_partner"

	signString := strconv.FormatInt(r.PartnerID, 10) + path + timestampStr

	partnerKeyBytes := []byte(r.PartnerKey)

	mac := hmac.New(sha256.New, partnerKeyBytes)
	mac.Write([]byte(signString))
	signature := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac.Sum(nil)))

	// grantSource 参数（可选，用于标识授权来源）
	grantSource := "1" // 默认值，可以根据需要调整

	authURL := fmt.Sprintf("%s?partner_id=%d&timestamp=%s&sign=%s&redirect=%s?grantSource=%s",
		baseAuthURL, r.PartnerID, timestampStr, signature,
		url.QueryEscape(callbackURL), grantSource)
	c.JSON(http.StatusOK, gin.H{
		"code":       200,
		"message":    "生成授权链接成功，在浏览器打开该URL进行授权",
		"auth_url":   authURL,
		"callback":   callbackURL,
		"is_sandbox": r.IsSandbox,
	})
}

// GenerateAuthURL 生成虾皮店铺授权链接，方便在浏览器里直接打开
// GET /api/v1/balance/admin/shopee/auth/url
//func (ctrl *ShopeeAuthController) GenerateAuthURL(c *gin.Context) {
//
//	// 从查询参数或配置中获取 shop_id
//	shopID := ctrl.cfg.ShopeeShopID
//	shopIDStr := c.Query("shop_id")
//	fmt.Println(shopID, "*****************************GenerateAuthURL**********************************", shopIDStr)
//	if shopIDStr != "" {
//		if id, err := strconv.ParseInt(shopIDStr, 10, 64); err == nil {
//			shopID = id
//		}
//	}
//	if shopID == 0 {
//		c.JSON(http.StatusBadRequest, gin.H{
//			"code":    -1,
//			"message": "请提供 shop_id 参数或在数据库中配置 shop_id",
//		})
//		return
//	}
//
//	// 从数据库获取 Shopee 配置
//	shopeeConfig, err := ctrl.getShopeeConfig(shopID)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{
//			"code":    -1,
//			"message": "未找到 Shopee 配置: " + err.Error(),
//		})
//		return
//	}
//
//	if shopeeConfig.PartnerID == 0 || shopeeConfig.PartnerKey == "" {
//		c.JSON(http.StatusBadRequest, gin.H{
//			"code":    -1,
//			"message": "Shopee Partner 配置不完整，请先在数据库中配置 partner_id / partner_key",
//		})
//		return
//	}
//
//	// 获取回调地址：优先使用数据库中的 redirect，否则使用当前请求的域名动态生成
//	var baseRedirectURL string
//	var redirectDomainOnly string // 只包含域名，不包含协议（如 kx9y.com）
//	if shopeeConfig.Redirect != "" {
//		// 使用数据库中的 redirect（可能包含协议，如 https://kx9y.com）
//		baseRedirectURL = shopeeConfig.Redirect
//
//		// 解析出纯域名部分（不包含协议）
//		parsedURL, err := url.Parse(baseRedirectURL)
//		if err == nil && parsedURL.Host != "" {
//			redirectDomainOnly = parsedURL.Host
//		} else {
//			// 如果没有协议，直接使用
//			redirectDomainOnly = baseRedirectURL
//		}
//
//		log.Printf("使用数据库中的 redirect: %s", baseRedirectURL) //数据库中  https://kx9y.com
//		log.Printf("提取的纯域名（不含协议）: %s", redirectDomainOnly)  //去掉http后得到kx9y.com
//	} else {
//		// 动态生成回调地址的域名部分
//		scheme := "https"
//		if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
//			scheme = "http"
//		}
//		host := c.Request.Host
//		baseRedirectURL = scheme + "://" + host
//		redirectDomainOnly = host
//		log.Printf("动态生成回调地址域名: %s", baseRedirectURL)
//	}
//
//	// 构建完整的回调 URL（添加路径）
//	callbackURL := baseRedirectURL + "/api/v1/balance/admin/shopee/auth/callback"
//
//	// 沙箱或正式环境的授权地址（与 Java 项目保持一致）
//	baseAuthURL := "https://partner.shopeemobile.com/api/v2/shop/auth_partner"
//	if shopeeConfig.IsSandbox {
//		baseAuthURL = "https://openplatform.sandbox.test-stable.shopee.cn/api/v2/shop/auth_partner"
//	}
//
//	// 生成 timestamp 和 sign
//	timestamp := time.Now().Unix()
//	timestampStr := strconv.FormatInt(timestamp, 10)
//
//	// Shopee 授权链接的签名规则（根据 Java 项目和官方文档）：
//	// 签名字符串 = partnerId + apiPath + timestamp（不包含 redirect！）
//	// 签名格式 = 64位十六进制小写字符串
//	path := "/api/v2/shop/auth_partner"
//
//	// 先构建 redirect URL（用于后续签名测试）
//	// 注意：这里先不构建完整的 redirect，等签名生成后再构建
//	redirectForSign := baseRedirectURL // 先使用域名，后续可能会尝试包含路径
//
//	// 方式1：签名字符串 = partnerId + path + timestamp（不包含 redirect）
//	// 这是 Shopee 官方文档的标准格式
//	signString1 := strconv.FormatInt(shopeeConfig.PartnerID, 10) + path + timestampStr
//
//	// 方式2：签名字符串 = partnerId + path + timestamp + redirect（包含 redirect）
//	// 某些实现可能会包含 redirect，但官方文档说不需要
//	signString2 := strconv.FormatInt(shopeeConfig.PartnerID, 10) + path + timestampStr + redirectForSign
//
//	// 方式3：签名字符串 = partnerId + path + timestamp + redirect（包含完整回调URL）
//	signString3 := strconv.FormatInt(shopeeConfig.PartnerID, 10) + path + timestampStr + callbackURL
//
//	// 方式2已测试报错 "Partner_id is invalid"，说明签名字符串格式可能不对
//	// 回退到方式1（不包含 redirect），这是官方标准格式
//
//	// 处理 partner_key
//	partnerKeyRaw := shopeeConfig.PartnerKey
//
//	// 尝试三种方式处理 partner_key：
//	// 方式A：保留 shpk 前缀，直接作为字符串（与 ExchangeShopeeToken 一致）
//	partnerKeyBytesA := []byte(partnerKeyRaw)
//
//	// 定义 signature 变量
//	var signature string
//
//	// 方式B：去掉 shpk 前缀，作为字符串
//	partnerKeyWithoutShpk := partnerKeyRaw
//	if len(partnerKeyRaw) > 4 && partnerKeyRaw[:4] == "shpk" {
//		partnerKeyWithoutShpk = partnerKeyRaw[4:]
//		log.Printf("已去掉 partner_key 的 shpk 前缀，处理后: %s", partnerKeyWithoutShpk)
//	}
//	partnerKeyBytesB := []byte(partnerKeyWithoutShpk)
//
//	// 方式C：去掉 shpk 前缀，十六进制解码
//	var hexDecoded []byte
//	if decoded, err := hex.DecodeString(partnerKeyWithoutShpk); err == nil && len(decoded) > 0 {
//		hexDecoded = decoded
//		log.Printf("partner_key 十六进制解码成功，长度=%d字节", len(hexDecoded))
//	} else {
//		log.Printf("⚠️  partner_key 不是有效的十六进制字符串，跳过方式C")
//		hexDecoded = nil
//	}
//	partnerKeyBytesC := hexDecoded
//
//	// 签名字符串格式1：partnerId + path + timestamp（不包含 redirect）
//	mac1A := hmac.New(sha256.New, partnerKeyBytesA)
//	mac1A.Write([]byte(signString1))
//	signature1A := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac1A.Sum(nil)))
//
//	mac1B := hmac.New(sha256.New, partnerKeyBytesB)
//	mac1B.Write([]byte(signString1))
//	signature1B := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac1B.Sum(nil)))
//
//	var signature1C string
//	if partnerKeyBytesC != nil {
//		mac1C := hmac.New(sha256.New, partnerKeyBytesC)
//		mac1C.Write([]byte(signString1))
//		signature1C = fmt.Sprintf("%064x", new(big.Int).SetBytes(mac1C.Sum(nil)))
//	}
//
//	// 签名字符串格式2：partnerId + path + timestamp + redirect（包含域名）
//	mac2A := hmac.New(sha256.New, partnerKeyBytesA)
//	mac2A.Write([]byte(signString2))
//	signature2A := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac2A.Sum(nil)))
//
//	mac2B := hmac.New(sha256.New, partnerKeyBytesB)
//	mac2B.Write([]byte(signString2))
//	signature2B := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac2B.Sum(nil)))
//
//	var signature2C string
//	if partnerKeyBytesC != nil {
//		mac2C := hmac.New(sha256.New, partnerKeyBytesC)
//		mac2C.Write([]byte(signString2))
//		signature2C = fmt.Sprintf("%064x", new(big.Int).SetBytes(mac2C.Sum(nil)))
//	}
//
//	// 签名字符串格式3：partnerId + path + timestamp + redirect（包含完整回调URL）
//	mac3A := hmac.New(sha256.New, partnerKeyBytesA)
//	mac3A.Write([]byte(signString3))
//	signature3A := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac3A.Sum(nil)))
//
//	mac3B := hmac.New(sha256.New, partnerKeyBytesB)
//	mac3B.Write([]byte(signString3))
//	signature3B := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac3B.Sum(nil)))
//
//	var signature3C string
//	if partnerKeyBytesC != nil {
//		mac3C := hmac.New(sha256.New, partnerKeyBytesC)
//		mac3C.Write([]byte(signString3))
//		signature3C = fmt.Sprintf("%064x", new(big.Int).SetBytes(mac3C.Sum(nil)))
//	}
//
//	// 尝试：partner_key 方式A（保留 shpk 前缀）+ 签名字符串格式1（不包含 redirect）
//	// 因为 ExchangeShopeeToken 中 partner_key 是直接作为字符串使用的，没有去掉 shpk 前缀
//	signature = signature1A
//
//	log.Printf("========== 签名方式完整对比 ==========")
//	log.Printf("partner_key 方式A（保留shpk前缀，UTF-8字符串）:")
//	log.Printf("  格式1（不包含redirect）: %s", signature1A)
//	log.Printf("  格式2（包含redirect域名）: %s", signature2A)
//	log.Printf("  格式3（包含完整回调URL）: %s", signature3A)
//	log.Printf("partner_key 方式B（去掉shpk前缀，UTF-8字符串）:")
//	log.Printf("  格式1（不包含redirect）: %s", signature1B)
//	log.Printf("  格式2（包含redirect域名）: %s", signature2B)
//	log.Printf("  格式3（包含完整回调URL）: %s", signature3B)
//	if signature1C != "" {
//		log.Printf("partner_key 方式C（去掉shpk前缀，十六进制解码）:")
//		log.Printf("  格式1（不包含redirect）: %s", signature1C)
//		log.Printf("  格式2（包含redirect域名）: %s", signature2C)
//		log.Printf("  格式3（包含完整回调URL）: %s", signature3C)
//	}
//	log.Printf("当前使用: partner_key方式A（保留shpk前缀）+ 格式1（不包含redirect）")
//	log.Printf("⚠️  如果还是报错，可以尝试其他组合")
//	log.Printf("==========================================")
//
//	// grantSource 参数（可选，用于标识授权来源）
//	grantSource := "1" // 默认值，可以根据需要调整
//
//	// 构建授权链接
//	// Java URL格式：?partner_id=%s&timestamp=%s&sign=%s&redirect=%s?grantSource=%s
//	// 根据 Java 代码分析：
//	// - redirect 参数的值是 signBaseParamConfig.getRedirect()，即配置中的 redirect（只有域名，如 https://www.balanpay.com）
//	// - grantSource 是作为 URL 的一部分，格式是 redirect=%s?grantSource=%s
//	// - 所以最终的 URL 格式是：?partner_id=xxx&timestamp=xxx&sign=xxx&redirect=https://www.balanpay.com?grantSource=1
//	//
//	// 构建授权链接
//	// Java URL格式：?partner_id=%s&timestamp=%s&sign=%s&redirect=%s?grantSource=%s
//	//
//	// 尝试多种方式，找出 Shopee 验证逻辑：
//	// 方式1：只使用域名（不含协议），如 kx9y.com
//	// 方式2：使用域名+协议，如 https://kx9y.com
//	// 方式3：使用完整回调 URL，如 https://kx9y.com/api/v1/balance/admin/shopee/auth/callback
//	//
//	// 根据 Java 项目，redirect 参数的值应该是配置中的 redirect（包含协议，如 https://www.balanpay.com）
//	// 但 Shopee 验证时可能只验证域名部分，所以先尝试方式2（与 Java 项目一致）
//
//	// 方式1：只使用域名（不含协议）
//	redirectOption1 := redirectDomainOnly
//	authURL1 := fmt.Sprintf("%s?partner_id=%d&timestamp=%s&sign=%s&redirect=%s?grantSource=%s",
//		baseAuthURL, shopeeConfig.PartnerID, timestampStr, signature,
//		url.QueryEscape(redirectOption1), grantSource)
//
//	// 方式2：使用域名+协议（与 Java 项目一致）
//	redirectOption2 := baseRedirectURL
//	authURL2 := fmt.Sprintf("%s?partner_id=%d&timestamp=%s&sign=%s&redirect=%s?grantSource=%s",
//		baseAuthURL, shopeeConfig.PartnerID, timestampStr, signature,
//		url.QueryEscape(redirectOption2), grantSource)
//
//	// 方式3：使用完整回调 URL
//	redirectOption3 := callbackURL
//	authURL3 := fmt.Sprintf("%s?partner_id=%d&timestamp=%s&sign=%s&redirect=%s?grantSource=%s",
//		baseAuthURL, shopeeConfig.PartnerID, timestampStr, signature,
//		url.QueryEscape(redirectOption3), grantSource)
//
//	// 默认使用方式3（完整回调URL），因为方式1（只域名）报错 "redirect url is invalid"
//	// 如果方式3不行，再尝试方式2（域名+协议）
//	authURL := authURL3
//
//	log.Printf("========== 三种 redirect 参数方式对比 ==========")
//	log.Printf("方式1（只域名，不含协议）: redirect=%s", redirectOption1)
//	log.Printf("方式1 URL: %s", authURL1)
//	log.Printf("⚠️  方式1已测试：报错 'redirect url is invalid'，说明需要包含协议")
//	log.Printf("方式2（域名+协议）: redirect=%s", redirectOption2)
//	log.Printf("方式2 URL: %s", authURL2)
//	log.Printf("⚠️  方式2已测试：报错 'The domain of redirect is not consistent'")
//	log.Printf("方式3（完整回调URL）: redirect=%s", redirectOption3)
//	log.Printf("方式3 URL: %s", authURL3)
//	log.Printf("⚠️  方式3已测试：也报错 'The domain of redirect is not consistent'")
//	log.Printf("================================================")
//	log.Printf("")
//	log.Printf("========== Shopee 控制台配置检查清单 ==========")
//	log.Printf("1. 确认 Shopee 控制台配置的 Redirect URL Domain 格式：")
//	log.Printf("   - 如果控制台只允许填域名（如 kx9y.com），则配置应该是: kx9y.com")
//	log.Printf("   - 如果控制台允许填完整 URL（如 https://kx9y.com），则配置应该是: https://kx9y.com")
//	log.Printf("2. 确认当前使用的域名: %s", redirectDomainOnly)
//	log.Printf("3. 确认当前使用的完整域名: %s", baseRedirectURL)
//	log.Printf("4. 确认 Shopee 环境: %s", map[bool]string{true: "沙箱环境 (Sandbox)", false: "正式环境 (Production)"}[shopeeConfig.IsSandbox])
//	log.Printf("5. 确认 Shopee 控制台配置的环境与代码中的环境一致（沙箱/正式）")
//	log.Printf("6. 确认 Shopee 控制台配置已保存并等待 5-10 分钟生效")
//	log.Printf("7. 如果控制台配置的是 www.kx9y.com，而代码使用的是 kx9y.com，也会报错")
//	log.Printf("8. 如果控制台配置的是 kx9y.com，而代码使用的是 www.kx9y.com，也会报错")
//	log.Printf("================================================")
//
//	log.Printf("========== Shopee 授权链接生成调试信息 ==========")
//	log.Printf("partner_id: %d", shopeeConfig.PartnerID)
//	log.Printf("⚠️  重要：请确认此 partner_id 与 Shopee 控制台中的配置一致")
//	log.Printf("⚠️  控制台显示的 Test Partner_id: 1203446")
//	log.Printf("⚠️  控制台显示的 Live Partner_id: 2014300")
//	log.Printf("⚠️  如果当前使用的是沙箱环境，partner_id 应该是 1203446")
//	log.Printf("⚠️  如果当前使用的是正式环境，partner_id 应该是 2014300")
//	log.Printf("timestamp: %s", timestampStr)
//	log.Printf("callback: %s", callbackURL)
//
//	// 解析 callback URL 的域名部分，用于验证
//	callbackURLParsed, err := url.Parse(callbackURL)
//	if err == nil {
//		log.Printf("callback 域名: %s", callbackURLParsed.Host)
//		log.Printf("callback 路径: %s", callbackURLParsed.Path)
//		log.Printf("⚠️  请确保 Shopee 控制台配置的 Redirect URL Domain 域名部分与上述域名一致")
//		log.Printf("⚠️  控制台应填写: %s (只填域名，不包含协议和路径)", callbackURLParsed.Host)
//		log.Printf("⚠️  或者填写: https://%s (包含协议，不包含路径)", callbackURLParsed.Host)
//		log.Printf("⚠️  当前使用的完整 redirect URL: %s", callbackURL)
//		log.Printf("⚠️  如果控制台配置的是其他域名，请修改数据库中的 redirect 配置")
//	} else {
//		log.Printf("⚠️  解析 callback URL 失败: %v", err)
//	}
//
//	log.Printf("签名字符串 (partner_id+path+timestamp): %s", signString1)
//	log.Printf("partner_key 原始值: %s", shopeeConfig.PartnerKey)
//	log.Printf("partner_key 方式A（保留shpk前缀）长度: %d 字节", len(partnerKeyBytesA))
//	log.Printf("partner_key 方式B（去掉shpk前缀）长度: %d 字节", len(partnerKeyBytesB))
//	if partnerKeyBytesC != nil {
//		log.Printf("partner_key 方式C（十六进制解码）长度: %d 字节", len(partnerKeyBytesC))
//	}
//	log.Printf("生成的签名: %s (长度: %d)", signature, len(signature))
//	log.Printf("⚠️  如果还是报 'Wrong sign'，请尝试其他组合的签名")
//	log.Printf("完整授权URL: %s", authURL)
//	log.Printf("================================================")
//
//	c.JSON(http.StatusOK, gin.H{
//		"code":       200,
//		"message":    "生成授权链接成功，在浏览器打开该URL进行授权",
//		"auth_url":   authURL,
//		"callback":   callbackURL,
//		"is_sandbox": shopeeConfig.IsSandbox,
//	})
//}

// parseAuthParams 解析授权回调参数
func (ctrl *ShopeeAuthController) parseAuthParams(c *gin.Context) (int64, string, error) {
	code := c.Query("code")
	shopIDStr := c.Query("shop_id")
	if code == "" {
		return 0, "", fmt.Errorf("缺少参数 code")
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
	if shopID == 0 {
		return 0, "", fmt.Errorf("请提供 shop_id 参数")
	}
	return shopID, code, nil
}

// exchangeShopeeTokens 与虾皮API交换访问令牌
func (ctrl *ShopeeAuthController) exchangeShopeeTokens(shopID int64, code string, partnerID int64, partnerKey string, isSandbox bool) (string, string, int64, error) {
	accessToken, refreshToken, expireIn, err := utils.ExchangeShopeeToken(
		partnerID,
		partnerKey,
		shopID,
		code,
		isSandbox,
	)
	if err != nil {
		log.Printf("向虾皮换取 access_token 失败: %v", err)
		return "", "", 0, err
	}
	return accessToken, refreshToken, expireIn, nil
}

// saveTokensToDatabase 保存令牌到 shopee_shops 表
func (ctrl *ShopeeAuthController) saveTokensToDatabase(shopID int64, partnerID int64, partnerKey string, isSandbox bool, redirect string, accessToken, refreshToken string, expireIn int64) error {
	expireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
	shopRepo := models.NewShopeeShopRepository(ctrl.db)

	// 检查店铺是否已存在
	existingShop, err := shopRepo.GetByShopID(shopID)
	shopIDStr := strconv.FormatInt(shopID, 10)
	now := time.Now()

	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("❌ 查询店铺失败: %v", err)
		return err
	}

	if existingShop == nil {
		// 店铺不存在，创建新记录
		newShop := &models.ShopeeShop{
			ShopID:        shopID,
			ShopIDStr:     shopIDStr,
			ShopName:      "店铺名称待同步", // 默认值，后续从 Shopee API 同步
			Region:        "MY",      // 默认值，后续从 Shopee API 同步
			PartnerID:     partnerID,
			AccessToken:   &accessToken,
			RefreshToken:  &refreshToken,
			TokenExpireAt: &expireAt,
			AuthStatus:    1, // 已授权
			AuthTime:      &now,
			Status:        1,     // 正常
			Currency:      "MYR", // 默认货币
			AutoSync:      true,
			SyncInterval:  3600,
			SyncItems:     true,
			SyncOrders:    true,
			SyncLogistics: true,
			SyncFinance:   true,
		}
		err = shopRepo.CreateOrUpdate(newShop)
		if err != nil {
			log.Printf("❌ 保存 token 到 shopee_shops 表失败: %v", err)
			return err
		}
		log.Printf("✅ Shopee 授权成功并已保存到 shopee_shops 表: shop_id=%d, access_token=%s, refresh_token=%s, expire_in=%d秒",
			shopID, accessToken, refreshToken, expireIn)
	} else {
		// 店铺已存在，更新 token 信息
		updates := map[string]interface{}{
			"partner_id":      partnerID,
			"access_token":    accessToken,
			"refresh_token":   refreshToken,
			"token_expire_at": expireAt,
			"auth_status":     1,
			"auth_time":       now,
		}
		err = ctrl.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", shopID).Updates(updates).Error
		if err != nil {
			log.Printf("❌ 更新 token 到 shopee_shops 表失败: %v", err)
			return err
		}
		log.Printf("✅ Shopee token 已更新到 shopee_shops 表: shop_id=%d, access_token=%s, refresh_token=%s, expire_in=%d秒",
			shopID, accessToken, refreshToken, expireIn)
	}

	log.Printf("   token 过期时间: %s", expireAt.Format(time.RFC3339))
	return nil
}

// generateVerificationCode 生成6位数字验证码
func (ctrl *ShopeeAuthController) generateVerificationCode() string {
	code := make([]byte, 3)
	rand.Read(code)
	// 生成 100000-999999 之间的6位数字
	return fmt.Sprintf("%06d", (int(code[0])<<16|int(code[1])<<8|int(code[2]))%900000+100000)
}

// buildRebindCallbackURL 构建换绑确认页面 URL
func (ctrl *ShopeeAuthController) buildRebindCallbackURL(c *gin.Context, shopID int64, boundAdminID int64, boundUserName, boundEmail string) string {
	scheme := "https"
	if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	host := c.Request.Host

	// 构建前端换绑确认页面 URL
	frontendPath := "/shopee/auth/rebind"
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") || strings.Contains(host, ":") {
		// 开发环境：如果 host 是后端端口（19090），替换为前端端口（3000）
		if strings.Contains(host, ":19090") {
			host = strings.Replace(host, ":19090", ":3000", 1)
		} else if host == "localhost" || host == "127.0.0.1" {
			host = "localhost:3000"
		}
		return fmt.Sprintf("http://%s%s?shop_id=%d&bound_admin_id=%d&bound_user_name=%s&bound_email=%s",
			host, frontendPath, shopID, boundAdminID, url.QueryEscape(boundUserName), url.QueryEscape(boundEmail))
	} else {
		// 生产环境，需要加上 /balance/admin 前缀
		return fmt.Sprintf("%s://%s/balance/admin%s?shop_id=%d&bound_admin_id=%d&bound_user_name=%s&bound_email=%s",
			scheme, host, frontendPath, shopID, boundAdminID, url.QueryEscape(boundUserName), url.QueryEscape(boundEmail))
	}
}

// buildFrontendCallbackURL 构建前端回调URL
func (ctrl *ShopeeAuthController) buildFrontendCallbackURL(c *gin.Context, success bool, shopID int64, errorMessage string) string {
	scheme := "https"
	if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	host := c.Request.Host

	// 构建前端回调页面 URL（重定向到前端页面显示成功信息）
	frontendPath := "/shopee/auth/callback"
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") || strings.Contains(host, ":") {
		// 开发环境：如果 host 是后端端口（19090），替换为前端端口（3000）
		if strings.Contains(host, ":19090") {
			host = strings.Replace(host, ":19090", ":3000", 1)
		} else if host == "localhost" || host == "127.0.0.1" {
			host = "localhost:3000"
		}
		if success {
			return fmt.Sprintf("http://%s%s?success=true&shop_id=%d",
				host, frontendPath, shopID)
		} else {
			return fmt.Sprintf("http://%s%s?error=%s",
				host, frontendPath, url.QueryEscape(errorMessage))
		}
	} else {
		// 生产环境，需要加上 /balance/admin 前缀
		if success {
			return fmt.Sprintf("%s://%s/balance/admin%s?success=true&shop_id=%d",
				scheme, host, frontendPath, shopID)
		} else {
			return fmt.Sprintf("%s://%s/balance/admin%s?error=%s",
				scheme, host, frontendPath, url.QueryEscape(errorMessage))
		}
	}
}

// RefreshToken 刷新访问令牌
func (ctrl *ShopeeAuthController) RefreshToken(c *gin.Context) {
	var req struct {
		ShopID int64 `json:"shop_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数绑定失败: " + err.Error(),
		})
		return
	}
	// 从数据库获取店铺配置
	shopRepo := models.NewShopeeShopRepository(ctrl.db)
	shop, err := shopRepo.GetByShopID(req.ShopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "获取店铺配置失败: " + err.Error(),
		})
		return
	}

	if shop.RefreshToken == nil || *shop.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "没有可用的 refresh_token",
		})
		return
	}

	// 从 AuthConfig 获取 PartnerKey 和 IsSandbox
	authCfg, err := ctrl.authService.GetByPartnerId()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "获取授权配置失败: " + err.Error(),
		})
		return
	}

	// 调用工具函数，使用 refresh_token 刷新 access_token
	accessToken, newRefreshToken, expireIn, err := utils.RefreshShopeeToken(
		shop.PartnerID,
		authCfg.PartnerKey,
		req.ShopID,
		*shop.RefreshToken,
		authCfg.IsSandbox,
	)
	if err != nil {
		log.Printf("刷新 access_token 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "刷新 access_token 失败: " + err.Error(),
		})
		return
	}

	// 更新数据库中的令牌信息到 shopee_shops 表
	expireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
	err = ctrl.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", req.ShopID).Updates(map[string]interface{}{
		"access_token":       accessToken,
		"refresh_token":      newRefreshToken,
		"token_expire_at":    expireAt,
		"last_token_refresh": time.Now(),
	}).Error
	if err != nil {
		log.Printf("❌ 保存刷新后的 token 到数据库失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "刷新成功，但保存 token 到数据库失败: " + err.Error(),
		})
		return
	}

	log.Printf("✅ Shopee access_token 刷新成功: shop_id=%d, access_token=%s, refresh_token=%s, expire_in=%d秒",
		req.ShopID, accessToken, newRefreshToken, expireIn)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "刷新成功",
		"data": gin.H{
			"shop_id":       req.ShopID,
			"access_token":  accessToken,
			"refresh_token": newRefreshToken,
			"expire_in":     expireIn,
			"expire_at":     expireAt.Format(time.RFC3339),
		},
	})
}

// AutoRefreshTokens 自动刷新即将过期的tokens
func (ctrl *ShopeeAuthController) AutoRefreshTokens() {
	log.Printf("启动自动刷新 Shopee tokens 任务...")

	// 获取所有有refresh_token的店铺记录
	var shops []*models.ShopeeShop
	err := ctrl.db.Where("refresh_token IS NOT NULL AND refresh_token != ''").Find(&shops).Error
	if err != nil {
		log.Printf("获取所有 Shopee 店铺失败: %v", err)
		return
	}

	// 获取全局配置
	authCfg, err := ctrl.authService.GetByPartnerId()
	if err != nil {
		log.Printf("获取授权配置失败: %v", err)
		return
	}

	for _, shop := range shops {
		// 检查是否需要刷新（提前1小时刷新）
		if shop.TokenExpireAt != nil && shop.RefreshToken != nil {
			refreshTime := shop.TokenExpireAt.Add(-1 * time.Hour)
			if time.Now().After(refreshTime) {
				log.Printf("发现即将过期的 token (shop_id=%d)，开始自动刷新...", shop.ShopID)

				// 调用刷新函数
				accessToken, newRefreshToken, expireIn, err := utils.RefreshShopeeToken(
					shop.PartnerID,
					authCfg.PartnerKey,
					shop.ShopID,
					*shop.RefreshToken,
					authCfg.IsSandbox,
				)
				if err != nil {
					log.Printf("自动刷新 token 失败 (shop_id=%d): %v", shop.ShopID, err)
					continue
				}

				// 更新数据库
				expireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
				err = ctrl.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", shop.ShopID).Updates(map[string]interface{}{
					"access_token":       accessToken,
					"refresh_token":      newRefreshToken,
					"token_expire_at":    expireAt,
					"last_token_refresh": time.Now(),
				}).Error
				if err != nil {
					log.Printf("保存刷新后的 token 到数据库失败 (shop_id=%d): %v", shop.ShopID, err)
					continue
				}

				log.Printf("✅ 自动刷新 Shopee token 成功 (shop_id=%d): 新 token 有效期至 %s",
					shop.ShopID, expireAt.Format(time.RFC3339))
			}
		}
	}
}

func (ctrl *ShopeeAuthController) AuthBind(c *gin.Context) {
	log.Println("======= 收到 AuthBind 请求 =======")
	var req struct {
		Token  string `json:"token" form:"token" binding:"required"`
		ShopID int64  `json:"shop_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数绑定失败: " + err.Error(),
		})
		return
	}
	log.Printf("收到绑定请求 - Token: %s, ShopID: %d", req.Token, req.ShopID)
	// 验证token的有效性
	userID, err := utils.ParseToken(req.Token, []byte(ctrl.cfg.JWTSecret))
	if err != nil {
		log.Printf("无效的token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "无效的token",
		})
		return
	}
	log.Printf("点击返回首页时，用户ID: %d, shop的id为: %d, token: %s", userID, req.ShopID, req.Token)

	// 店铺注册和绑定逻辑
	shopRepo := models.NewShopeeShopRepository(ctrl.db)
	adminShopRepo := models.NewAdminShopRepository(ctrl.db)

	// 检查 shopee_shops 表中是否已有该 shop_id 的记录
	existingShop, err := shopRepo.GetByShopID(req.ShopID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("查询店铺失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "查询店铺失败: " + err.Error(),
		})
		return
	}

	now := time.Now()
	shopIDStr := strconv.FormatInt(req.ShopID, 10)

	if existingShop == nil {
		// 店铺不存在，创建新记录
		newShop := &models.ShopeeShop{
			ShopID:        req.ShopID,
			ShopIDStr:     shopIDStr,
			ShopName:      "店铺名称待同步", // 默认值，后续可以从 Shopee API 同步
			Region:        "MY",      // 默认值，后续可以从 Shopee API 同步
			PartnerID:     0,         // 默认值
			AuthStatus:    1,         // 已授权
			AuthTime:      &now,
			Status:        1,     // 正常
			Currency:      "MYR", // 默认货币
			AutoSync:      true,
			SyncInterval:  3600,
			SyncItems:     true,
			SyncOrders:    true,
			SyncLogistics: true,
			SyncFinance:   true,
		}

		err = shopRepo.CreateOrUpdate(newShop)
		if err != nil {
			log.Printf("创建店铺记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    -1,
				"message": "创建店铺记录失败: " + err.Error(),
			})
			return
		}
		log.Printf("✅ 成功创建店铺记录: shop_id=%d", req.ShopID)
	} else {
		// 店铺已存在，更新授权状态和时间
		err = shopRepo.UpdateAuthStatus(req.ShopID, 1, &now)
		if err != nil {
			log.Printf("更新店铺授权状态失败: %v", err)
		}
	}

	// 检查该 admin 是否已有其他店铺
	adminShops, err := adminShopRepo.GetByAdminID(userID)
	if err != nil {
		log.Printf("查询 admin 的店铺列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "查询店铺列表失败: " + err.Error(),
		})
		return
	}

	// 如果这是该 admin 的第一个店铺，设置为主店铺
	isPrimary := len(adminShops) == 0

	// 创建或更新关联关系
	adminShop := &models.AdminShop{
		AdminID:   userID,
		ShopID:    req.ShopID,
		IsPrimary: isPrimary,
		Status:    1, // 正常
	}

	err = adminShopRepo.CreateOrUpdate(adminShop)
	if err != nil {
		log.Printf("创建关联关系失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "创建关联关系失败: " + err.Error(),
		})
		return
	}

	// 如果设置为主店铺，确保其他店铺不是主店铺
	if isPrimary {
		err = adminShopRepo.SetPrimary(userID, req.ShopID)
		if err != nil {
			log.Printf("设置主店铺失败: %v", err)
			// 不返回错误，因为这不是关键操作
		}
	}

	log.Printf("✅ 成功绑定 admin 和 shop: admin_id=%d, shop_id=%d, is_primary=%v", userID, req.ShopID, isPrimary)

	log.Println("======= AuthBind 请求处理完成 =======")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "店铺绑定成功!",
		"data": gin.H{
			"user_id": userID,
			"shop_id": req.ShopID,
		},
	})
	return
}

// AuthCallback 虾皮授权回调
// 用于接收 code，并调用 Shopee 接口换取 access_token
// 示例回调地址： https://kx9y.com/api/v1/balance/admin/shopee/auth/callback
func (ctrl *ShopeeAuthController) AuthCallback(c *gin.Context) {
	// 解析参数
	shopID, code, err := ctrl.parseAuthParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	authCfg, err := ctrl.authService.GetByPartnerId()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}
	accessToken, refreshToken, expireIn, err := ctrl.exchangeShopeeTokens(shopID, code, authCfg.PartnerID, authCfg.PartnerKey, authCfg.IsSandbox)
	if err != nil {
		// 构建前端错误回调页面 URL
		errorMsg := "向虾皮换取 access_token 失败: " + err.Error()
		frontendCallbackURL := ctrl.buildFrontendCallbackURL(c, false, shopID, errorMsg)
		c.Redirect(http.StatusFound, frontendCallbackURL)
		return
	}
	fmt.Println("***************accessToken, refreshToken, expireIn获取正常***********************")
	// 检查 admin_shop 是否已经绑定
	adminShopRepo := models.NewAdminShopRepository(ctrl.db)
	existingBindings, err := adminShopRepo.GetByShopID(shopID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("查询店铺绑定关系失败: %v", err)
		// 继续执行，不阻断流程
	}

	// 如果已绑定，需要验证邮箱换绑
	if len(existingBindings) > 0 {
		fmt.Println("***************已绑定过***********************")
		// 获取已绑定的 admin 信息
		adminRepo := models.NewAdminRepository(ctrl.db)
		boundAdmin, err := adminRepo.GetByID(existingBindings[0].AdminID)
		if err != nil {
			log.Printf("获取已绑定管理员信息失败: %v", err)
			// 继续执行，不阻断流程
		} else {
			// 将 token 信息临时存储到 Redis，等待用户点击发送验证码
			// Key: shop_bind_pending:{shop_id}
			// Value: {admin_id}:{access_token}:{refresh_token}:{expire_in}
			ctx := context.Background()
			pendingKey := fmt.Sprintf("shop_bind_pending:%d", shopID)
			pendingValue := fmt.Sprintf("%d:%s:%s:%d", boundAdmin.ID, accessToken, refreshToken, expireIn)
			err = ctrl.redisClient.Set(ctx, pendingKey, pendingValue, 10*time.Minute).Err()
			if err != nil {
				fmt.Println("***************33333333333333333***********************")
				log.Printf("保存待绑定信息到 Redis 失败: %v", err)
				// 如果 Redis 失败，继续执行，直接更新 token
			} else {
				fmt.Println("***************444444444444444444***********************")
				// 重定向到前端换绑确认页面（用户需要点击"发送验证码"按钮）
				frontendCallbackURL := ctrl.buildRebindCallbackURL(c, shopID, boundAdmin.ID, boundAdmin.UserName, boundAdmin.Email)
				log.Printf("重定向到前端换绑页面: %s", frontendCallbackURL)
				c.Redirect(http.StatusFound, frontendCallbackURL)
				return
			}
		}
	}
	fmt.Println("************************未绑定过***********************************")
	// 保存 token 到数据库
	err = ctrl.saveTokensToDatabase(shopID, authCfg.PartnerID, authCfg.PartnerKey, authCfg.IsSandbox,
		authCfg.Redirect, accessToken, refreshToken, expireIn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "授权成功，但保存 token 到数据库失败: " + err.Error(),
			"data": gin.H{
				"shop_id":       shopID,
				"access_token":  accessToken,
				"refresh_token": refreshToken,
				"expire_in":     expireIn,
				"expire_at":     time.Now().Add(time.Duration(expireIn) * time.Second).Format(time.RFC3339),
			},
		})
		return
	}
	go func() {
		// 调用服务拉取店铺详情
		result, err := ctrl.orderService.FetchShopDetailFromShopee(shopID)
		if err != nil {
			// 检查错误是否与token过期相关
			errMsg := err.Error()
			if strings.Contains(errMsg, "access_token") || strings.Contains(errMsg, "token") || strings.Contains(errMsg, "Wrong sign") {
				log.Printf("检测到token相关错误，尝试刷新token后重试: %v", err)

				// 尝试刷新token
				refreshErr := ctrl.orderService.RefreshTokenAndRetry()
				if refreshErr != nil {
					log.Printf("刷新token失败: %v", refreshErr)
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    -1,
						"message": "获取店铺详情失败: " + err.Error(),
					})
					return
				}

				// 再次尝试调用API
				log.Printf("刷新token成功，重新尝试拉取店铺详情...")
				result, err = ctrl.orderService.FetchShopDetailFromShopee(shopID)
				if err != nil {
					log.Printf("重试拉取店铺详情仍失败: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    -1,
						"message": "获取店铺详情失败: " + err.Error(),
					})
					return
				}
			} else {
				log.Printf("拉取店铺详情失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    -1,
					"message": "获取店铺详情失败: " + err.Error(),
				})
				return
			}
		}
		//shop_fulfillment_flag
		/*使用此字段标识当前店铺的履约类型，适用的值包括：
		- Pure - FBS Shop：单一模式，指仅拥有Shopee官方仓库存货的本地/跨境店铺，订单由Shopee从Shopee官方仓库履约；
		- Pure - 3PF Shop：单一模式，指仅拥有本地卖家仓库存货的跨境店铺，订单由卖家通过本地物流渠道从本地卖家仓库履约；
		- PFF - FBS Shop：
		混合模式，指同时拥有Shopee官方仓库存货和本地卖家仓库存货的本地店铺，订单既可由Shopee从Shopee官方仓库履约，也可由卖家通过本地物流渠道从本地卖家仓库履约；
		混合模式，指同时拥有Shopee官方仓库存货和跨境卖家仓库存货的跨境店铺，订单既可由Shopee从Shopee官方仓库履约，也可由卖家通过跨境物流渠道从跨境卖家仓库履约；
		- PFF - 3PF Shop：混合模式，指同时拥有本地卖家仓库存货和跨境卖家仓库存货的跨境店铺，订单可由卖家通过本地物流渠道从本地卖家仓库履约，也可由卖家通过跨境物流渠道从跨境卖家仓库履约；
		- LFF Hybrid Shop：混合模式，指拥有三种存货类型的跨境店铺：FBS存货（Shopee官方仓库存货）、3PF存货（跨境卖家在本地市场的自有存货）和CB SLS存货（跨境卖家在中国/香港/韩国的自有存货）；
		- Others：其他
		- Unknown：当获取shop_fulfillment_flag信息失败时返回
		*/
		type ProfileResponse struct {
			Description string `json:"description"`
			ShopLogo    string `json:"shop_logo"`
			ShopName    string `json:"shop_name"`
		}
		type Profile struct {
			RequestID string          `json:"request_id"`
			Response  ProfileResponse `json:"response"`
		}
		type AuthResponse struct {
			AuthTime             float64     `json:"auth_time"`
			Error                string      `json:"error,omitempty"`
			ExpireTime           float64     `json:"expire_time"`
			IsCb                 bool        `json:"is_cb"`
			IsDirectShop         bool        `json:"is_direct_shop"`
			IsMainShop           bool        `json:"is_main_shop"`
			IsMartShop           bool        `json:"is_mart_shop"`
			IsOneAwb             bool        `json:"is_one_awb"`
			IsOutletShop         bool        `json:"is_outlet_shop"`
			IsSip                bool        `json:"is_sip"`
			IsUpgradedCbsc       bool        `json:"is_upgraded_cbsc"`
			LinkedDirectShopList []string    `json:"linked_direct_shop_list"`
			LinkedMainShopID     int         `json:"linked_main_shop_id"`
			MerchantID           interface{} `json:"merchant_id"` // 根据实际类型调整
			Message              string      `json:"message,omitempty"`
			Profile              Profile     `json:"profile"`
			Region               string      `json:"region"`
			RequestID            string      `json:"request_id"`
			ShopFulfillmentFlag  string      `json:"shop_fulfillment_flag"`
			ShopName             string      `json:"shop_name"`
			Status               string      `json:"status"`
		}
		jsonData, _ := json.Marshal(result)
		var r AuthResponse
		json.Unmarshal(jsonData, &r)
		fmt.Println(
			r.IsOneAwb,   //是否航空运单(相对应的统一运单)
			r.IsMartShop, //是否为Mart Shop（商城店/超市店）
			r.IsMainShop, //是否为关联到跨境直购店的本地店铺
			r.ShopName,
			r.Status,
			r.Region,
			r.IsCb,
			r.AuthTime,
			r.ExpireTime,
			r.MerchantID,
			r.Profile.Response.ShopLogo,
			r.Profile.Response.Description,
			r.IsSip, //SIP主店铺或联盟店铺调用时，此字段将返回"true"
		)

		// 将店铺信息保存到 shopee_shops 表
		shopRepo := models.NewShopeeShopRepository(ctrl.db)
		shopIDStr := strconv.FormatInt(shopID, 10)

		// 解析 MerchantID
		var merchantID *int64
		if r.MerchantID != nil {
			switch v := r.MerchantID.(type) {
			case float64:
				id := int64(v)
				merchantID = &id
			case int:
				id := int64(v)
				merchantID = &id
			case int64:
				merchantID = &v
			}
		}

		// 解析时间
		var authTime *time.Time
		if r.AuthTime > 0 {
			t := time.Unix(int64(r.AuthTime), 0)
			authTime = &t
		}

		// 解析店铺状态
		var status int16 = 1 // 默认正常
		if r.Status == "NORMAL" {
			status = 1
		} else if r.Status == "BANNED" {
			status = 2
		} else if r.Status == "FROZEN" {
			status = 3
		} else if r.Status == "CLOSED" {
			status = 4
		}

		// 获取现有店铺信息（保留 token 等信息）
		existingShop, err := shopRepo.GetByShopID(shopID)
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Printf("❌ 查询店铺信息失败: %v", err)
			return
		}

		// 创建或更新店铺记录
		shop := &models.ShopeeShop{
			ShopID:     shopID,
			ShopIDStr:  shopIDStr,
			ShopName:   r.ShopName,
			Region:     r.Region,
			IsCbShop:   r.IsCb,
			Status:     status,
			AuthStatus: 1,
			AuthTime:   authTime,
			MerchantID: merchantID,
		}

		// 如果店铺已存在，合并现有数据
		if existingShop != nil {
			shop.PartnerID = existingShop.PartnerID
			shop.AccessToken = existingShop.AccessToken
			shop.RefreshToken = existingShop.RefreshToken
			shop.TokenExpireAt = existingShop.TokenExpireAt
			// 不再更新 owner_id，因为现在使用关联表
			shop.Currency = existingShop.Currency
			shop.AutoSync = existingShop.AutoSync
			shop.SyncInterval = existingShop.SyncInterval
			shop.SyncItems = existingShop.SyncItems
			shop.SyncOrders = existingShop.SyncOrders
			shop.SyncLogistics = existingShop.SyncLogistics
			shop.SyncFinance = existingShop.SyncFinance
		} else {
			// 新店铺，设置默认值
			shop.Currency = "MYR"
			shop.AutoSync = true
			shop.SyncInterval = 3600
			shop.SyncItems = true
			shop.SyncOrders = true
			shop.SyncLogistics = true
			shop.SyncFinance = true
		}

		err = shopRepo.CreateOrUpdate(shop)
		if err != nil {
			log.Printf("❌ 保存店铺信息到 shopee_shops 表失败: %v", err)
		} else {
			log.Printf("✅ 店铺信息已保存到 shopee_shops 表: shop_id=%d, shop_name=%s, region=%s", shopID, r.ShopName, r.Region)
		}
	}()
	// 构建前端成功回调页面 URL 并重定向
	frontendCallbackURL := ctrl.buildFrontendCallbackURL(c, true, shopID, "")
	c.Redirect(http.StatusFound, frontendCallbackURL)
}

// SendRebindCode 发送换绑验证码到邮箱
// POST /api/v1/balance/admin/shopee/auth/rebind/send-code
func (ctrl *ShopeeAuthController) SendRebindCode(c *gin.Context) {
	var req struct {
		ShopID int64 `json:"shop_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 从 Redis 获取待绑定信息
	ctx := context.Background()
	pendingKey := fmt.Sprintf("shop_bind_pending:%d", req.ShopID)
	pendingValue, err := ctrl.redisClient.Get(ctx, pendingKey).Result()
	if err == redis.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "待绑定信息已过期或不存在",
		})
		return
	} else if err != nil {
		log.Printf("从 Redis 获取待绑定信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "获取信息失败: " + err.Error(),
		})
		return
	}

	// 解析待绑定信息: {admin_id}:{access_token}:{refresh_token}:{expire_in}
	parts := strings.Split(pendingValue, ":")
	if len(parts) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "待绑定信息格式错误",
		})
		return
	}

	boundAdminID, _ := strconv.ParseInt(parts[0], 10, 64)
	accessToken := parts[1]
	refreshToken := parts[2]
	expireIn, _ := strconv.ParseInt(parts[3], 10, 64)

	// 获取已绑定的 admin 信息
	adminRepo := models.NewAdminRepository(ctrl.db)
	boundAdmin, err := adminRepo.GetByID(boundAdminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "获取管理员信息失败: " + err.Error(),
		})
		return
	}

	// 生成验证码（6位数字）
	verificationCode := ctrl.generateVerificationCode()

	// 将验证码存储到 Redis，有效期 10 分钟
	// Key: shop_bind_verify:{shop_id}
	// Value: {admin_id}:{verification_code}:{access_token}:{refresh_token}:{expire_in}
	verifyKey := fmt.Sprintf("shop_bind_verify:%d", req.ShopID)
	verifyValue := fmt.Sprintf("%d:%s:%s:%s:%d", boundAdminID, verificationCode, accessToken, refreshToken, expireIn)
	err = ctrl.redisClient.Set(ctx, verifyKey, verifyValue, 10*time.Minute).Err()
	if err != nil {
		log.Printf("保存验证码到 Redis 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "生成验证码失败: " + err.Error(),
		})
		return
	}

	// 发送验证码到邮箱（这里先记录日志，实际应该发送邮件）
	log.Printf("验证码已生成并发送: shop_id=%d, bound_admin_id=%d, bound_email=%s, verification_code=%s",
		req.ShopID, boundAdminID, boundAdmin.Email, verificationCode)

	// TODO: 实际发送邮件到 boundAdmin.Email
	// 这里应该调用邮件服务发送验证码

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "验证码已发送到邮箱",
		"data": gin.H{
			"shop_id": req.ShopID,
			"email":   boundAdmin.Email,
		},
	})
}

// VerifyRebindCode 验证换绑验证码
// POST /api/v1/balance/admin/shopee/auth/rebind/verify
func (ctrl *ShopeeAuthController) VerifyRebindCode(c *gin.Context) {
	var req struct {
		ShopID int64  `json:"shop_id" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 从 Redis 获取验证码信息
	ctx := context.Background()
	verifyKey := fmt.Sprintf("shop_bind_verify:%d", req.ShopID)
	verifyValue, err := ctrl.redisClient.Get(ctx, verifyKey).Result()
	if err == redis.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "验证码已过期或不存在，请重新发送",
		})
		return
	} else if err != nil {
		log.Printf("从 Redis 获取验证码失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "验证失败: " + err.Error(),
		})
		return
	}

	// 解析验证码信息: {admin_id}:{verification_code}:{access_token}:{refresh_token}:{expire_in}
	parts := strings.Split(verifyValue, ":")
	if len(parts) != 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "验证码格式错误",
		})
		return
	}

	boundAdminID, _ := strconv.ParseInt(parts[0], 10, 64)
	storedCode := parts[1]
	// accessToken, refreshToken, expireIn 在 ConfirmRebind 中使用

	// 验证验证码
	if req.Code != storedCode {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "验证码错误",
		})
		return
	}

	// 验证码正确，返回成功（前端可以继续调用换绑接口）
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "验证码验证成功",
		"data": gin.H{
			"shop_id":        req.ShopID,
			"bound_admin_id": boundAdminID,
		},
	})
}

// ConfirmRebind 确认换绑（需要先验证验证码）
// POST /api/v1/balance/admin/shopee/auth/rebind/confirm
func (ctrl *ShopeeAuthController) ConfirmRebind(c *gin.Context) {
	var req struct {
		ShopID     int64  `json:"shop_id" binding:"required"`
		Code       string `json:"code" binding:"required"`
		NewAdminID int64  `json:"new_admin_id" binding:"required"` // 新绑定的 admin ID（从 token 中获取）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 从 Redis 获取验证码信息
	ctx := context.Background()
	verifyKey := fmt.Sprintf("shop_bind_verify:%d", req.ShopID)
	verifyValue, err := ctrl.redisClient.Get(ctx, verifyKey).Result()
	if err == redis.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "验证码已过期或不存在",
		})
		return
	} else if err != nil {
		log.Printf("从 Redis 获取验证码失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "验证失败: " + err.Error(),
		})
		return
	}

	// 解析验证码信息
	parts := strings.Split(verifyValue, ":")
	if len(parts) != 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "验证码格式错误",
		})
		return
	}

	boundAdminID, _ := strconv.ParseInt(parts[0], 10, 64)
	storedCode := parts[1]
	accessToken := parts[2]
	refreshToken := parts[3]
	expireIn, _ := strconv.ParseInt(parts[4], 10, 64)

	// 验证验证码
	if req.Code != storedCode {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "验证码错误",
		})
		return
	}

	// 删除旧的绑定关系
	adminShopRepo := models.NewAdminShopRepository(ctrl.db)
	err = adminShopRepo.Delete(boundAdminID, req.ShopID)
	if err != nil {
		log.Printf("删除旧绑定关系失败: %v", err)
		// 继续执行，不阻断流程
	}

	// 创建新的绑定关系
	// 检查新 admin 是否已有其他店铺
	adminShops, err := adminShopRepo.GetByAdminID(req.NewAdminID)
	if err != nil {
		log.Printf("查询新 admin 的店铺列表失败: %v", err)
	}
	isPrimary := len(adminShops) == 0

	newAdminShop := &models.AdminShop{
		AdminID:   req.NewAdminID,
		ShopID:    req.ShopID,
		IsPrimary: isPrimary,
		Status:    1,
	}
	err = adminShopRepo.CreateOrUpdate(newAdminShop)
	if err != nil {
		log.Printf("创建新绑定关系失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "换绑失败: " + err.Error(),
		})
		return
	}

	// 如果设置为主店铺，确保其他店铺不是主店铺
	if isPrimary {
		err = adminShopRepo.SetPrimary(req.NewAdminID, req.ShopID)
		if err != nil {
			log.Printf("设置主店铺失败: %v", err)
		}
	}

	// 更新 token 到店铺表
	expireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
	err = ctrl.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", req.ShopID).Updates(map[string]interface{}{
		"access_token":    accessToken,
		"refresh_token":   refreshToken,
		"token_expire_at": expireAt,
		"auth_status":     1,
		"auth_time":       time.Now(),
	}).Error
	if err != nil {
		log.Printf("更新 token 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "换绑成功，但更新 token 失败: " + err.Error(),
		})
		return
	}

	// 删除 Redis 中的验证码和待绑定信息
	ctrl.redisClient.Del(ctx, verifyKey)
	pendingKey := fmt.Sprintf("shop_bind_pending:%d", req.ShopID)
	ctrl.redisClient.Del(ctx, pendingKey)

	log.Printf("✅ 换绑成功: shop_id=%d, 旧 admin_id=%d, 新 admin_id=%d", req.ShopID, boundAdminID, req.NewAdminID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "换绑成功",
		"data": gin.H{
			"shop_id":      req.ShopID,
			"old_admin_id": boundAdminID,
			"new_admin_id": req.NewAdminID,
		},
	})
}

// ShopList 获取当前用户的所有店铺列表
// POST /api/v1/balance/admin/shopee/shop/list
func (ctrl *ShopeeAuthController) ShopList(c *gin.Context) {
	fmt.Println("********************ShopList11111*****************************")
	// 从请求头获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "缺少Authorization头",
		})
		return
	}
	fmt.Println("********************ShopList2222222*****************************")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "Authorization格式错误",
		})
		return
	}
	fmt.Println("********************ShopList3333333*****************************")
	tokenStr := parts[1]
	userID, err := utils.ParseToken(tokenStr, []byte(ctrl.cfg.JWTSecret))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "无效或过期的token",
		})
		return
	}

	// 验证用户类型，只有店主类型（userType=1）才能访问此接口
	adminRepo := models.NewAdminRepository(ctrl.db)
	admin, err := adminRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "获取用户信息失败: " + err.Error(),
		})
		return
	}

	// 只有店主类型（userType=1）才能访问
	if admin.UserType != 1 {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    -1,
			"message": "此接口仅限店主类型用户访问",
		})
		return
	}

	fmt.Println("********************ShopList4444444*****************************")
	// 查询 admin_shop 表，获取该用户关联的所有店铺 shop_id
	adminShopRepo := models.NewAdminShopRepository(ctrl.db)
	adminShops, err := adminShopRepo.GetByAdminID(userID)
	if err != nil {
		log.Printf("查询用户店铺关联失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "查询店铺关联失败: " + err.Error(),
		})
		return
	}
	fmt.Println("********************ShopList55555555*****************************")
	// 如果没有关联的店铺，返回空列表
	if len(adminShops) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"data":    []interface{}{},
		})
		return
	}

	// 提取所有 shop_id
	shopIDs := make([]int64, 0, len(adminShops))
	for _, adminShop := range adminShops {
		shopIDs = append(shopIDs, adminShop.ShopID)
	}

	// 根据多个 shop_id 查询 shopee_shops 表，获取店铺详细信息
	var shops []*models.ShopeeShop
	err = ctrl.db.Where("shop_id IN ?", shopIDs).Find(&shops).Error
	if err != nil {
		log.Printf("查询店铺详细信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "查询店铺详细信息失败: " + err.Error(),
		})
		return
	}

	// 构建返回数据 - 返回完整的店铺信息（与数据库模型一致）
	shopList := make([]gin.H, 0, len(shops))
	for _, shop := range shops {
		// 查找对应的 admin_shop 关联信息
		var adminShop *models.AdminShop
		for _, as := range adminShops {
			if as.ShopID == shop.ShopID {
				adminShop = as
				break
			}
		}
		t := time.Unix(shop.ExpireTime, 0)
		const layout = "2006-01-02 15:04:05"
		expireTime := t.Format(layout)
		shopData := gin.H{
			"id":                  shop.ID,
			"shopId":              shop.ShopID,
			"shopIdStr":           shop.ShopIDStr,
			"shopName":            shop.ShopName,
			"shopSlug":            shop.ShopSlug,
			"region":              shop.Region,
			"partnerId":           shop.PartnerID,
			"expireTime":          expireTime,
			"authStatus":          shop.AuthStatus,
			"status":              shop.Status,
			"suspensionStatus":    shop.SuspensionStatus,
			"isCbShop":            shop.IsCbShop,
			"isCodShop":           shop.IsCodShop,
			"isPreferredPlusShop": shop.IsPreferredPlusShop,
			"isShopeeVerified":    shop.IsShopeeVerified,
			"ratingStar":          shop.RatingStar,
			"ratingBad":           shop.RatingBad,
			"ratingGood":          shop.RatingGood,
			"ratingNormal":        shop.RatingNormal,
			"itemCount":           shop.ItemCount,
			"followerCount":       shop.FollowerCount,
			"responseRate":        shop.ResponseRate,
			"responseTime":        shop.ResponseTime,
			"cancellationRate":    shop.CancellationRate,
			"totalSales":          shop.TotalSales,
			"totalOrders":         shop.TotalOrders,
			"totalViews":          shop.TotalViews,
			"dailySales":          shop.DailySales,
			"monthlySales":        shop.MonthlySales,
			"yearlySales":         shop.YearlySales,
			"currency":            shop.Currency,
			"balance":             shop.Balance,
			"pendingBalance":      shop.PendingBalance,
			"withdrawnBalance":    shop.WithdrawnBalance,
			"contactEmail":        shop.ContactEmail,
			"contactPhone":        shop.ContactPhone,
			"country":             shop.Country,
			"city":                shop.City,
			"address":             shop.Address,
			"zipcode":             shop.Zipcode,
			"autoSync":            shop.AutoSync,
			"syncInterval":        shop.SyncInterval,
			"syncItems":           shop.SyncItems,
			"syncOrders":          shop.SyncOrders,
			"syncLogistics":       shop.SyncLogistics,
			"syncFinance":         shop.SyncFinance,
			"isPrimary":           false,
			"authTime":            nil,
			"tokenExpireAt":       nil,
			"lastSyncAt":          nil,
			"nextSyncAt":          nil,
			"shopCreatedAt":       nil,
			"createdAt":           shop.CreatedAt.Format(time.RFC3339),
			"updatedAt":           shop.UpdatedAt.Format(time.RFC3339),
		}

		// 添加 admin_shop 关联信息
		if adminShop != nil {
			shopData["isPrimary"] = adminShop.IsPrimary
		}

		// 处理可选的时间字段
		if shop.AuthTime != nil {
			shopData["authTime"] = shop.AuthTime.Format(time.RFC3339)
		}
		if shop.TokenExpireAt != nil {
			shopData["tokenExpireAt"] = shop.TokenExpireAt.Format(time.RFC3339)
		}
		if shop.LastSyncAt != nil {
			shopData["lastSyncAt"] = shop.LastSyncAt.Format(time.RFC3339)
		}
		if shop.NextSyncAt != nil {
			shopData["nextSyncAt"] = shop.NextSyncAt.Format(time.RFC3339)
		}
		if shop.ShopCreatedAt != nil {
			shopData["shopCreatedAt"] = shop.ShopCreatedAt.Format(time.RFC3339)
		}

		shopList = append(shopList, shopData)
	}
	fmt.Println("**************shopList:**********************", shopList)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    shopList,
	})
}

// CancelRebind 取消换绑（只更新 token，不换绑）
// POST /api/v1/balance/admin/shopee/auth/rebind/cancel
func (ctrl *ShopeeAuthController) CancelRebind(c *gin.Context) {
	var req struct {
		ShopID int64 `json:"shop_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 从 Redis 获取待绑定信息（包含 token）
	ctx := context.Background()
	pendingKey := fmt.Sprintf("shop_bind_pending:%d", req.ShopID)
	pendingValue, err := ctrl.redisClient.Get(ctx, pendingKey).Result()
	if err == redis.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "待绑定信息已过期或不存在",
		})
		return
	} else if err != nil {
		log.Printf("从 Redis 获取待绑定信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "操作失败: " + err.Error(),
		})
		return
	}

	// 解析待绑定信息: {admin_id}:{access_token}:{refresh_token}:{expire_in}
	parts := strings.Split(pendingValue, ":")
	if len(parts) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "待绑定信息格式错误",
		})
		return
	}

	accessToken := parts[1]
	refreshToken := parts[2]
	expireIn, _ := strconv.ParseInt(parts[3], 10, 64)

	// 只更新 token，不换绑
	expireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
	err = ctrl.db.Model(&models.ShopeeShop{}).Where("shop_id = ?", req.ShopID).Updates(map[string]interface{}{
		"access_token":    accessToken,
		"refresh_token":   refreshToken,
		"token_expire_at": expireAt,
		"auth_status":     1,
		"auth_time":       time.Now(),
	}).Error
	if err != nil {
		log.Printf("更新 token 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "更新 token 失败: " + err.Error(),
		})
		return
	}

	// 删除 Redis 中的待绑定信息和验证码
	ctrl.redisClient.Del(ctx, pendingKey)
	verifyKey := fmt.Sprintf("shop_bind_verify:%d", req.ShopID)
	ctrl.redisClient.Del(ctx, verifyKey)

	log.Printf("✅ 取消换绑，已更新 token: shop_id=%d", req.ShopID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "已取消换绑，token 已更新",
		"data": gin.H{
			"shop_id": req.ShopID,
		},
	})
}

//func (ctrl *ShopeeAuthController) AuthCallback(c *gin.Context) {
//	code := c.Query("code")
//	shopIDStr := c.Query("shop_id")
//	if code == "" {
//		c.JSON(http.StatusBadRequest, gin.H{
//			"code":    -1,
//			"message": "缺少参数 code",
//		})
//		return
//	}
//	var shopID int64
//	if shopIDStr != "" {
//		if id, err := strconv.ParseInt(shopIDStr, 10, 64); err == nil {
//			shopID = id
//		}
//	}
//	if shopID == 0 {
//		shopID = ctrl.cfg.ShopeeShopID
//	}
//	if shopID == 0 {
//		c.JSON(http.StatusBadRequest, gin.H{
//			"code":    -1,
//			"message": "请提供 shop_id 参数",
//		})
//		return
//	}
//	// 从数据库获取 Shopee 配置
//	shopeeConfig, err := ctrl.getShopeeConfig(shopID)
//	if err != nil {
//		if err != gorm.ErrRecordNotFound {
//			return
//		}
//	}
//	// 调用工具函数，向虾皮请求 access_token
//	accessToken, refreshToken, expireIn, err := utils.ExchangeShopeeToken(
//		shopeeConfig.PartnerID,
//		shopeeConfig.PartnerKey,
//		shopID,
//		code,
//		shopeeConfig.IsSandbox,
//	)
//	if err != nil {
//		log.Printf("向虾皮换取 access_token 失败: %v", err)
//		// 构建前端错误回调页面 URL
//		scheme := "https"
//		if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
//			scheme = "http"
//		}
//		host := c.Request.Host
//		frontendPath := "/shopee/auth/callback"
//		if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") || strings.Contains(host, ":") {
//			// 开发环境
//			frontendCallbackURL := fmt.Sprintf("http://%s%s?error=%s",
//				host, frontendPath, url.QueryEscape("向虾皮换取 access_token 失败: "+err.Error()))
//			c.Redirect(http.StatusFound, frontendCallbackURL)
//		} else {
//			// 生产环境
//			frontendCallbackURL := fmt.Sprintf("%s://%s/balance/admin%s?error=%s",
//				scheme, host, frontendPath, url.QueryEscape("向虾皮换取 access_token 失败: "+err.Error()))
//			c.Redirect(http.StatusFound, frontendCallbackURL)
//		}
//		return
//	}
//
//	// 保存 token 到数据库（保留原有的 partner_key 和 redirect 配置）
//	expireAt := time.Now().Add(time.Duration(expireIn) * time.Second)
//	tokenRepo := models.NewShopeeTokenRepository(ctrl.db)
//	shopeeToken := &models.ShopeeToken{
//		ShopID:        shopID,
//		PartnerID:     shopeeConfig.PartnerID,
//		PartnerKey:    shopeeConfig.PartnerKey, // 保留原有的 partner_key
//		AccessToken:   accessToken,
//		RefreshToken:  refreshToken,
//		TokenExpireAt: &expireAt,
//		IsSandbox:     shopeeConfig.IsSandbox,
//		Redirect:      shopeeConfig.Redirect, // 保留原有的 redirect
//	}
//
//	err = tokenRepo.CreateOrUpdate(shopeeToken)
//	if err != nil {
//		log.Printf("❌ 保存 token 到数据库失败: %v", err)
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"code":    -1,
//			"message": "授权成功，但保存 token 到数据库失败: " + err.Error(),
//			"data": gin.H{
//				"shop_id":       shopID,
//				"access_token":  accessToken,
//				"refresh_token": refreshToken,
//				"expire_in":     expireIn,
//				"expire_at":     expireAt.Format(time.RFC3339),
//			},
//		})
//		return
//	}
//
//	log.Printf("✅ Shopee 授权成功并已保存到数据库: shop_id=%d, access_token=%s, refresh_token=%s, expire_in=%d秒",
//		shopID, accessToken, refreshToken, expireIn)
//	log.Printf("   token 过期时间: %s", expireAt.Format(time.RFC3339))
//
//	// 构建前端回调页面 URL（重定向到前端页面显示成功信息）
//	// 根据请求头判断是开发环境还是生产环境
//	scheme := "https"
//	if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
//		scheme = "http"
//	}
//	host := c.Request.Host
//
//	// 构建前端回调页面 URL（重定向到前端页面显示成功信息）
//	frontendPath := "/shopee/auth/callback"
//	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") || strings.Contains(host, ":") {
//		// 开发环境
//		frontendCallbackURL := fmt.Sprintf("http://%s%s?success=true&shop_id=%d",
//			host, frontendPath, shopID)
//		c.Redirect(http.StatusFound, frontendCallbackURL)
//	} else {
//		// 生产环境，需要加上 /balance/admin 前缀
//		frontendCallbackURL := fmt.Sprintf("%s://%s/balance/admin%s?success=true&shop_id=%d",
//			scheme, host, frontendPath, shopID)
//		c.Redirect(http.StatusFound, frontendCallbackURL)
//	}
//}
