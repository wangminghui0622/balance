package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"balance/backend/internal/config"
	"balance/backend/internal/consts"
	"balance/backend/internal/database"
	"balance/backend/internal/models"
	"balance/backend/internal/shopee"

	"gorm.io/gorm"
)

// ShopService 店铺服务
type ShopService struct {
	db *gorm.DB
}

// NewShopService 创建店铺服务
func NewShopService() *ShopService {
	return &ShopService{
		db: database.GetDB(),
	}
}

// GetAuthURL 获取授权链接（带用户ID）
func (s *ShopService) GetAuthURL(region string, adminID int64) string {
	client := shopee.NewClient(region)
	redirectURL := config.Get().Shopee.RedirectURL
	state := fmt.Sprintf("%d", adminID)
	return client.GetAuthURL(redirectURL, state)
}

// HandleAuthCallback 处理授权回调
func (s *ShopService) HandleAuthCallback(ctx context.Context, code string, shopID uint64, region string, adminID int64) error {
	client := shopee.NewClient(region)
	cfg := config.Get().Shopee

	tokenResp, err := client.GetAccessToken(code, shopID)
	if err != nil {
		return fmt.Errorf("获取Token失败: %w", err)
	}

	shopInfo, err := client.GetShopInfo(tokenResp.AccessToken, shopID)
	if err != nil {
		return fmt.Errorf("获取店铺信息失败: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(tokenResp.ExpireIn) * time.Second)

	var refreshExpiresAt time.Time
	if shopInfo.ExpireTime > 0 {
		refreshExpiresAt = time.Unix(shopInfo.ExpireTime, 0)
	} else if tokenResp.RefreshExpireIn > 0 {
		refreshExpiresAt = now.Add(time.Duration(tokenResp.RefreshExpireIn) * time.Second)
	} else {
		refreshExpiresAt = now.Add(30 * 24 * time.Hour)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var existingShop models.Shop
		err := tx.Where("shop_id = ?", shopID).First(&existingShop).Error

		if err == nil {
			updates := map[string]interface{}{
				"shop_name":          shopInfo.Response.ShopName,
				"shop_id_str":        fmt.Sprintf("%d", shopID),
				"region":             shopInfo.Response.Region,
				"partner_id":         cfg.PartnerID,
				"auth_status":        int8(1),
				"status":             consts.ShopStatusEnabled,
				"is_cb_shop":         shopInfo.Response.IsCB,
				"is_shopee_verified": len(shopInfo.Response.SIPAffiliate) > 0,
			}
			if existingShop.AdminID == 0 && adminID > 0 {
				updates["admin_id"] = adminID
			}
			if err := tx.Model(&existingShop).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新店铺信息失败: %w", err)
			}
		} else if err == gorm.ErrRecordNotFound {
			shop := models.Shop{
				ShopID:           shopID,
				ShopIDStr:        fmt.Sprintf("%d", shopID),
				AdminID:          adminID,
				ShopName:         shopInfo.Response.ShopName,
				Region:           shopInfo.Response.Region,
				PartnerID:        cfg.PartnerID,
				AuthStatus:       1,
				Status:           consts.ShopStatusEnabled,
				IsCbShop:         shopInfo.Response.IsCB,
				IsShopeeVerified: len(shopInfo.Response.SIPAffiliate) > 0,
				Currency:         getCurrencyByRegion(shopInfo.Response.Region),
				AutoSync:         true,
				SyncInterval:     3600,
				SyncItems:        true,
				SyncOrders:       true,
				SyncLogistics:    true,
				SyncFinance:      true,
			}
			if err := tx.Create(&shop).Error; err != nil {
				return fmt.Errorf("创建店铺失败: %w", err)
			}
		} else {
			return fmt.Errorf("查询店铺失败: %w", err)
		}

		auth := models.ShopAuthorization{
			ShopID:           shopID,
			AccessToken:      tokenResp.AccessToken,
			RefreshToken:     tokenResp.RefreshToken,
			TokenType:        "Bearer",
			ExpiresAt:        expiresAt,
			RefreshExpiresAt: refreshExpiresAt,
		}

		if err := tx.Where("shop_id = ?", shopID).Assign(auth).FirstOrCreate(&auth).Error; err != nil {
			return fmt.Errorf("保存授权信息失败: %w", err)
		}

		if err := s.cacheToken(ctx, shopID, tokenResp.AccessToken, tokenResp.RefreshToken, expiresAt); err != nil {
			fmt.Printf("缓存Token失败: %v\n", err)
		}

		return nil
	})
}

func getCurrencyByRegion(region string) string {
	currencyMap := map[string]string{
		"SG": "SGD", "MY": "MYR", "TH": "THB", "TW": "TWD",
		"VN": "VND", "PH": "PHP", "ID": "IDR", "BR": "BRL",
		"MX": "MXN", "CO": "COP", "CL": "CLP",
	}
	if currency, ok := currencyMap[region]; ok {
		return currency
	}
	return "USD"
}

func (s *ShopService) cacheToken(ctx context.Context, shopID uint64, accessToken, refreshToken string, expiresAt time.Time) error {
	rdb := database.GetRedis()
	key := fmt.Sprintf(consts.KeyShopToken, shopID)

	data := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_at":    expiresAt.Unix(),
	}

	jsonData, _ := json.Marshal(data)
	ttl := time.Until(expiresAt) - consts.TokenExpireBuffer
	if ttl <= 0 {
		ttl = time.Minute
	}

	return rdb.Set(ctx, key, jsonData, ttl).Err()
}

// GetAccessToken 获取店铺访问令牌
func (s *ShopService) GetAccessToken(ctx context.Context, shopID uint64) (string, error) {
	rdb := database.GetRedis()
	key := fmt.Sprintf(consts.KeyShopToken, shopID)

	data, err := rdb.Get(ctx, key).Result()
	if err == nil {
		var tokenData map[string]interface{}
		if json.Unmarshal([]byte(data), &tokenData) == nil {
			if token, ok := tokenData["access_token"].(string); ok {
				return token, nil
			}
		}
	}

	var auth models.ShopAuthorization
	if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
		return "", fmt.Errorf("店铺未授权: %w", err)
	}

	if auth.IsAccessTokenExpired() {
		if err := s.RefreshToken(ctx, shopID); err != nil {
			return "", err
		}
		if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
			return "", err
		}
	}

	if err := s.cacheToken(ctx, shopID, auth.AccessToken, auth.RefreshToken, auth.ExpiresAt); err != nil {
		fmt.Printf("缓存Token失败: %v\n", err)
	}

	return auth.AccessToken, nil
}

// RefreshToken 刷新Token
func (s *ShopService) RefreshToken(ctx context.Context, shopID uint64) error {
	var auth models.ShopAuthorization
	if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
		return fmt.Errorf("店铺未授权: %w", err)
	}

	if auth.IsRefreshTokenExpired() {
		return fmt.Errorf("刷新Token已过期，请重新授权")
	}

	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return fmt.Errorf("店铺不存在: %w", err)
	}

	client := shopee.NewClient(shop.Region)
	tokenResp, err := client.RefreshAccessToken(auth.RefreshToken, shopID)
	if err != nil {
		return fmt.Errorf("刷新Token失败: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(tokenResp.ExpireIn) * time.Second)
	refreshExpireIn := tokenResp.RefreshExpireIn
	if refreshExpireIn <= 0 {
		refreshExpireIn = 30 * 24 * 3600
	}
	refreshExpiresAt := now.Add(time.Duration(refreshExpireIn) * time.Second)

	auth.AccessToken = tokenResp.AccessToken
	auth.RefreshToken = tokenResp.RefreshToken
	auth.ExpiresAt = expiresAt
	auth.RefreshExpiresAt = refreshExpiresAt

	if err := s.db.Save(&auth).Error; err != nil {
		return fmt.Errorf("保存Token失败: %w", err)
	}

	if err := s.cacheToken(ctx, shopID, tokenResp.AccessToken, tokenResp.RefreshToken, expiresAt); err != nil {
		fmt.Printf("缓存Token失败: %v\n", err)
	}

	return nil
}

// ListShops 获取店铺列表
func (s *ShopService) ListShops(ctx context.Context, page, pageSize int, status *int8, adminID int64) ([]models.ShopWithAuth, int64, error) {
	var shops []models.Shop
	var total int64

	query := s.db.Model(&models.Shop{})

	if adminID > 0 {
		query = query.Where("admin_id = ?", adminID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&shops).Error; err != nil {
		return nil, 0, err
	}

	shopIDs := make([]uint64, len(shops))
	for i, shop := range shops {
		shopIDs[i] = shop.ShopID
	}

	var auths []models.ShopAuthorization
	if len(shopIDs) > 0 {
		s.db.Where("shop_id IN ?", shopIDs).Find(&auths)
	}

	authMap := make(map[uint64]*models.ShopAuthorization)
	for i := range auths {
		auth := &auths[i]
		authMap[auth.ShopID] = auth
	}

	result := make([]models.ShopWithAuth, len(shops))
	for i := range shops {
		shop := &shops[i]
		if shop.ShopIDStr == "" {
			shop.ShopIDStr = fmt.Sprintf("%d", shop.ShopID)
		}

		result[i] = models.ShopWithAuth{Shop: *shop}

		if auth, ok := authMap[shop.ShopID]; ok {
			result[i].AuthTime = &auth.CreatedAt
			result[i].ExpireTime = &auth.RefreshExpiresAt

			if auth.AccessToken != "" && !auth.IsAccessTokenExpired() {
				result[i].AuthStatus = 1
			} else if auth.RefreshToken != "" && !auth.IsRefreshTokenExpired() {
				result[i].AuthStatus = 1
			} else {
				result[i].AuthStatus = 2
			}
		} else {
			result[i].AuthStatus = 0
		}
	}

	return result, total, nil
}

// GetShop 获取店铺详情
func (s *ShopService) GetShop(ctx context.Context, shopID uint64) (*models.Shop, error) {
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return nil, err
	}
	return &shop, nil
}

// UpdateShopStatus 更新店铺状态
func (s *ShopService) UpdateShopStatus(ctx context.Context, shopID uint64, status int8) error {
	return s.db.Model(&models.Shop{}).Where("shop_id = ?", shopID).Update("status", status).Error
}

// DeleteShop 删除店铺
func (s *ShopService) DeleteShop(ctx context.Context, shopID uint64) error {
	return s.UpdateShopStatus(ctx, shopID, consts.ShopStatusDisabled)
}

// BindShopToAdmin 将店铺绑定到用户
func (s *ShopService) BindShopToAdmin(ctx context.Context, shopID uint64, adminID int64) error {
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return fmt.Errorf("店铺不存在: %w", err)
	}

	if shop.AdminID > 0 && shop.AdminID != adminID {
		return fmt.Errorf("该店铺已被其他用户绑定")
	}

	if shop.AdminID == adminID {
		return nil
	}

	return s.db.Model(&shop).Update("admin_id", adminID).Error
}

// GetAuthorizedShops 获取所有已授权的店铺
func (s *ShopService) GetAuthorizedShops() ([]models.Shop, error) {
	var shops []models.Shop
	err := s.db.Where("auth_status = ?", 1).Find(&shops).Error
	return shops, err
}

// UpdateLastSyncTime 更新店铺最后同步时间
func (s *ShopService) UpdateLastSyncTime(shopID uint64) error {
	now := time.Now()
	nextSync := now.Add(30 * time.Minute)
	return s.db.Model(&models.Shop{}).Where("shop_id = ?", shopID).Updates(map[string]interface{}{
		"last_sync_at": now,
		"next_sync_at": nextSync,
	}).Error
}
