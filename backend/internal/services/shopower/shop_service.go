package shopower

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

// ShopService 店铺服务（店主专用）
type ShopService struct {
	db *gorm.DB
}

// NewShopService 创建店铺服务
func NewShopService() *ShopService {
	return &ShopService{
		db: database.GetDB(),
	}
}

// GetAuthURL 获取授权链接
func (s *ShopService) GetAuthURL(ctx context.Context, adminID int64) (string, error) {
	cfg := config.Get().Shopee
	client := shopee.NewClient("SG") // 默认使用SG区域
	state := fmt.Sprintf("%d", adminID)
	return client.GetAuthURL(cfg.RedirectURL, state), nil
}

// HandleAuthCallback 处理授权回调
func (s *ShopService) HandleAuthCallback(ctx context.Context, code string, shopID int64, adminID int64) error {
	fmt.Printf("[DEBUG] HandleAuthCallback: shopID=%d, adminID=%d\n", shopID, adminID)
	cfg := config.Get().Shopee
	client := shopee.NewClient("SG") // 默认使用SG区域

	tokenResp, err := client.GetAccessToken(code, uint64(shopID))
	if err != nil {
		return fmt.Errorf("获取Token失败: %w", err)
	}

	shopInfo, err := client.GetShopInfo(tokenResp.AccessToken, uint64(shopID))
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
			// 如果店铺未绑定用户且传入了有效的 adminID，则绑定
			if existingShop.AdminID == 0 && adminID > 0 {
				updates["admin_id"] = adminID
				fmt.Printf("[DEBUG] 绑定店铺 %d 到用户 %d\n", shopID, adminID)
			}
			if err := tx.Model(&existingShop).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新店铺信息失败: %w", err)
			}
		} else if err == gorm.ErrRecordNotFound {
			shop := models.Shop{
				ShopID:           uint64(shopID),
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
			}
			if err := tx.Create(&shop).Error; err != nil {
				return fmt.Errorf("创建店铺失败: %w", err)
			}
		} else {
			return fmt.Errorf("查询店铺失败: %w", err)
		}

		var existingAuth models.ShopAuthorization
		if err := tx.Where("shop_id = ?", shopID).First(&existingAuth).Error; err == nil {
			// 更新已有授权信息，只更新需要的字段
			if err := tx.Model(&existingAuth).Updates(map[string]interface{}{
				"access_token":       tokenResp.AccessToken,
				"refresh_token":      tokenResp.RefreshToken,
				"expires_at":         expiresAt,
				"refresh_expires_at": refreshExpiresAt,
			}).Error; err != nil {
				return fmt.Errorf("更新授权信息失败: %w", err)
			}
		} else {
			// 创建新授权信息
			auth := models.ShopAuthorization{
				ShopID:           uint64(shopID),
				AccessToken:      tokenResp.AccessToken,
				RefreshToken:     tokenResp.RefreshToken,
				ExpiresAt:        expiresAt,
				RefreshExpiresAt: refreshExpiresAt,
			}
			if err := tx.Create(&auth).Error; err != nil {
				return fmt.Errorf("创建授权信息失败: %w", err)
			}
		}

		s.cacheToken(ctx, uint64(shopID), tokenResp.AccessToken, tokenResp.RefreshToken, expiresAt)
		return nil
	})
}

func (s *ShopService) cacheToken(ctx context.Context, shopID uint64, accessToken, refreshToken string, expiresAt time.Time) {
	rdb := database.GetRedis()
	key := fmt.Sprintf(consts.KeyShopToken, shopID)
	data, _ := json.Marshal(map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_at":    expiresAt.Unix(),
	})
	rdb.Set(ctx, key, data, time.Until(expiresAt))
}

func getCurrencyByRegion(region string) string {
	currencies := map[string]string{
		"TW": "TWD", "SG": "SGD", "MY": "MYR", "TH": "THB",
		"ID": "IDR", "VN": "VND", "PH": "PHP", "BR": "BRL",
	}
	if c, ok := currencies[region]; ok {
		return c
	}
	return "USD"
}

// GetAccessToken 获取店铺访问令牌（带缓存）
func (s *ShopService) GetAccessToken(ctx context.Context, shopID uint64) (string, error) {
	rdb := database.GetRedis()
	key := fmt.Sprintf(consts.KeyShopToken, shopID)

	// 尝试从缓存获取
	data, err := rdb.Get(ctx, key).Result()
	if err == nil {
		var tokenData map[string]interface{}
		if json.Unmarshal([]byte(data), &tokenData) == nil {
			if token, ok := tokenData["access_token"].(string); ok {
				return token, nil
			}
		}
	}

	// 从数据库获取
	var auth models.ShopAuthorization
	if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
		return "", fmt.Errorf("店铺未授权: %w", err)
	}

	// 检查是否过期，需要刷新
	if auth.IsAccessTokenExpired() {
		if err := s.RefreshToken(ctx, shopID); err != nil {
			return "", err
		}
		if err := s.db.Where("shop_id = ?", shopID).First(&auth).Error; err != nil {
			return "", err
		}
	}

	// 缓存Token
	s.cacheToken(ctx, shopID, auth.AccessToken, auth.RefreshToken, auth.ExpiresAt)

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

	s.cacheToken(ctx, shopID, tokenResp.AccessToken, tokenResp.RefreshToken, expiresAt)
	return nil
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

// ListShops 获取店铺列表（只能查看自己的店铺）
func (s *ShopService) ListShops(ctx context.Context, adminID int64, page, pageSize int) ([]models.ShopWithAuth, int64, error) {
	var shops []models.Shop
	var total int64

	query := s.db.Model(&models.Shop{}).Where("admin_id = ?", adminID)

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
		authMap[auths[i].ShopID] = &auths[i]
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
		}
	}

	return result, total, nil
}

// GetShop 获取店铺详情（验证归属权）
// 当 adminID 为 0 时，跳过权限检查（用于调度器等内部场景）
func (s *ShopService) GetShop(ctx context.Context, adminID int64, shopID int64) (*models.Shop, error) {
	var shop models.Shop
	query := s.db.Where("shop_id = ?", shopID)
	if adminID > 0 {
		query = query.Where("admin_id = ?", adminID)
	}
	if err := query.First(&shop).Error; err != nil {
		return nil, fmt.Errorf("店铺不存在或无权限访问")
	}
	return &shop, nil
}

// BindShop 绑定店铺
func (s *ShopService) BindShop(ctx context.Context, adminID int64, shopID int64) error {
	var shop models.Shop
	if err := s.db.Where("shop_id = ?", shopID).First(&shop).Error; err != nil {
		return fmt.Errorf("店铺不存在")
	}
	if shop.AdminID > 0 && shop.AdminID != adminID {
		return fmt.Errorf("该店铺已被其他用户绑定")
	}
	return s.db.Model(&shop).Update("admin_id", adminID).Error
}

// UpdateShopStatus 更新店铺状态
func (s *ShopService) UpdateShopStatus(ctx context.Context, adminID int64, shopID int64, status int) error {
	result := s.db.Model(&models.Shop{}).Where("shop_id = ? AND admin_id = ?", shopID, adminID).Update("status", status)
	if result.RowsAffected == 0 {
		return fmt.Errorf("店铺不存在或无权限操作")
	}
	return result.Error
}

// DeleteShop 删除店铺
func (s *ShopService) DeleteShop(ctx context.Context, adminID int64, shopID int64) error {
	return s.UpdateShopStatus(ctx, adminID, shopID, consts.ShopStatusDisabled)
}

// RefreshShopToken 刷新店铺Token
func (s *ShopService) RefreshShopToken(ctx context.Context, adminID int64, shopID int64) error {
	var shop models.Shop
	if err := s.db.Where("shop_id = ? AND admin_id = ?", shopID, adminID).First(&shop).Error; err != nil {
		return fmt.Errorf("店铺不存在或无权限操作")
	}
	return s.RefreshToken(ctx, uint64(shopID))
}
