package models

import (
	"time"
)

// Shop 店铺模型
type Shop struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID              uint64     `gorm:"uniqueIndex;not null" json:"shopId"`
	ShopIDStr           string     `gorm:"size:64;not null;default:''" json:"shopIdStr"`
	AdminID             int64      `gorm:"not null;default:0;index" json:"adminId"`
	ShopName            string     `gorm:"size:255;not null;default:''" json:"shopName"`
	ShopSlug            *string    `gorm:"size:256" json:"shopSlug"`
	Region              string     `gorm:"size:16;not null;index" json:"region"`
	PartnerID           int64      `gorm:"not null;default:0" json:"partnerId"`
	AuthStatus          int8       `gorm:"not null;default:0;index" json:"authStatus"`
	Status              int8       `gorm:"not null;default:1;index" json:"status"`
	SuspensionStatus    int8       `gorm:"not null;default:0" json:"suspensionStatus"`
	IsCbShop            bool       `gorm:"default:false" json:"isCbShop"`
	IsCodShop           bool       `gorm:"default:false" json:"isCodShop"`
	IsPreferredPlusShop bool       `gorm:"default:false" json:"isPreferredPlusShop"`
	IsShopeeVerified    bool       `gorm:"default:false" json:"isShopeeVerified"`
	RatingStar          float64    `gorm:"type:decimal(3,2);default:0" json:"ratingStar"`
	RatingBad           int        `gorm:"default:0" json:"ratingBad"`
	RatingGood          int        `gorm:"default:0" json:"ratingGood"`
	RatingNormal        int        `gorm:"default:0" json:"ratingNormal"`
	ItemCount           int        `gorm:"default:0" json:"itemCount"`
	FollowerCount       int        `gorm:"default:0" json:"followerCount"`
	ResponseRate        float64    `gorm:"type:decimal(5,2);default:0" json:"responseRate"`
	ResponseTime        int        `gorm:"default:0" json:"responseTime"`
	CancellationRate    float64    `gorm:"type:decimal(5,2);default:0" json:"cancellationRate"`
	TotalSales          int        `gorm:"default:0" json:"totalSales"`
	TotalOrders         int        `gorm:"default:0" json:"totalOrders"`
	TotalViews          int        `gorm:"default:0" json:"totalViews"`
	DailySales          int        `gorm:"default:0" json:"dailySales"`
	MonthlySales        int        `gorm:"default:0" json:"monthlySales"`
	YearlySales         int        `gorm:"default:0" json:"yearlySales"`
	Currency            string     `gorm:"size:10;default:'MYR'" json:"currency"`
	Balance             float64    `gorm:"type:decimal(12,2);default:0" json:"balance"`
	PendingBalance      float64    `gorm:"type:decimal(12,2);default:0" json:"pendingBalance"`
	WithdrawnBalance    float64    `gorm:"type:decimal(12,2);default:0" json:"withdrawnBalance"`
	ContactEmail        *string    `gorm:"size:200" json:"contactEmail"`
	ContactPhone        *string    `gorm:"size:50" json:"contactPhone"`
	Country             *string    `gorm:"size:100" json:"country"`
	City                *string    `gorm:"size:100" json:"city"`
	Address             *string    `gorm:"type:text" json:"address"`
	Zipcode             *string    `gorm:"size:20" json:"zipcode"`
	AutoSync            bool       `gorm:"default:true" json:"autoSync"`
	SyncInterval        int        `gorm:"default:3600" json:"syncInterval"`
	SyncItems           bool       `gorm:"default:true" json:"syncItems"`
	SyncOrders          bool       `gorm:"default:true" json:"syncOrders"`
	SyncLogistics       bool       `gorm:"default:true" json:"syncLogistics"`
	SyncFinance         bool       `gorm:"default:true" json:"syncFinance"`
	IsPrimary           bool       `gorm:"default:false" json:"isPrimary"`
	LastSyncAt          *time.Time `gorm:"type:datetime" json:"lastSyncAt"`
	NextSyncAt          *time.Time `gorm:"type:datetime" json:"nextSyncAt"`
	ShopCreatedAt       *time.Time `gorm:"type:datetime" json:"shopCreatedAt"`
	CreatedAt           time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (Shop) TableName() string {
	return "shops"
}

// ShopWithAuth 店铺信息（包含授权时间）
type ShopWithAuth struct {
	Shop
	AuthTime   *time.Time `json:"authTime"`
	ExpireTime *time.Time `json:"expireTime"`
}

// ShopAuthorization 店铺授权模型
type ShopAuthorization struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShopID           uint64    `gorm:"uniqueIndex;not null" json:"shop_id"`
	AccessToken      string    `gorm:"size:512;not null" json:"access_token"`
	RefreshToken     string    `gorm:"size:512;not null" json:"refresh_token"`
	TokenType        string    `gorm:"size:50;not null;default:'Bearer'" json:"token_type"`
	ExpiresAt        time.Time `gorm:"not null;index" json:"expires_at"`
	RefreshExpiresAt time.Time `gorm:"not null" json:"refresh_expires_at"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (ShopAuthorization) TableName() string {
	return "shop_authorizations"
}

// IsAccessTokenExpired 检查访问令牌是否过期
func (s *ShopAuthorization) IsAccessTokenExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsRefreshTokenExpired 检查刷新令牌是否过期
func (s *ShopAuthorization) IsRefreshTokenExpired() bool {
	return time.Now().After(s.RefreshExpiresAt)
}
