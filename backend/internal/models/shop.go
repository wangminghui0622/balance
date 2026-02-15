package models

import (
	"time"
)

// 店铺状态常量
const (
	ShopStatusDisabled = 0 // 禁用
	ShopStatusEnabled  = 1 // 启用
)

// Shop 店铺模型
type Shop struct {
	ID                  uint64     `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID              uint64     `gorm:"uniqueIndex;not null;comment:Shopee店铺ID" json:"shopId"`
	ShopIDStr           string     `gorm:"size:64;not null;default:'';comment:店铺ID字符串" json:"shopIdStr"`
	AdminID             int64      `gorm:"not null;default:0;index;comment:所属管理员ID" json:"adminId"`
	ShopName            string     `gorm:"size:255;not null;default:'';comment:店铺名称" json:"shopName"`
	ShopSlug            *string    `gorm:"size:256;comment:店铺别名" json:"shopSlug"`
	Region              string     `gorm:"size:16;not null;index;comment:地区代码(MY/SG/TW等)" json:"region"`
	PartnerID           int64      `gorm:"not null;default:0;comment:合作伙伴ID" json:"partnerId"`
	AuthStatus          int8       `gorm:"not null;default:0;index;comment:授权状态(0未授权/1已授权)" json:"authStatus"`
	Status              int8       `gorm:"not null;default:1;index;comment:店铺状态(0禁用/1启用)" json:"status"`
	SuspensionStatus    int8       `gorm:"not null;default:0;comment:暂停状态" json:"suspensionStatus"`
	IsCbShop            bool       `gorm:"default:false;comment:是否跨境店铺" json:"isCbShop"`
	IsCodShop           bool       `gorm:"default:false;comment:是否支持货到付款" json:"isCodShop"`
	IsPreferredPlusShop bool       `gorm:"default:false;comment:是否优选Plus店铺" json:"isPreferredPlusShop"`
	IsShopeeVerified    bool       `gorm:"default:false;comment:是否Shopee认证" json:"isShopeeVerified"`
	RatingStar          float64    `gorm:"type:decimal(3,2);default:0;comment:店铺评分" json:"ratingStar"`
	RatingBad           int        `gorm:"default:0;comment:差评数" json:"ratingBad"`
	RatingGood          int        `gorm:"default:0;comment:好评数" json:"ratingGood"`
	RatingNormal        int        `gorm:"default:0;comment:中评数" json:"ratingNormal"`
	ItemCount           int        `gorm:"default:0;comment:商品数量" json:"itemCount"`
	FollowerCount       int        `gorm:"default:0;comment:粉丝数" json:"followerCount"`
	ResponseRate        float64    `gorm:"type:decimal(5,2);default:0;comment:回复率" json:"responseRate"`
	ResponseTime        int        `gorm:"default:0;comment:平均回复时间(秒)" json:"responseTime"`
	CancellationRate    float64    `gorm:"type:decimal(5,2);default:0;comment:取消率" json:"cancellationRate"`
	TotalSales          int        `gorm:"default:0;comment:总销售额" json:"totalSales"`
	TotalOrders         int        `gorm:"default:0;comment:总订单数" json:"totalOrders"`
	TotalViews          int        `gorm:"default:0;comment:总浏览量" json:"totalViews"`
	DailySales          int        `gorm:"default:0;comment:日销售额" json:"dailySales"`
	MonthlySales        int        `gorm:"default:0;comment:月销售额" json:"monthlySales"`
	YearlySales         int        `gorm:"default:0;comment:年销售额" json:"yearlySales"`
	Currency            string     `gorm:"size:10;default:'MYR';comment:货币代码" json:"currency"`
	Balance             float64    `gorm:"type:decimal(12,2);default:0;comment:可用余额" json:"balance"`
	PendingBalance      float64    `gorm:"type:decimal(12,2);default:0;comment:待结算余额" json:"pendingBalance"`
	WithdrawnBalance    float64    `gorm:"type:decimal(12,2);default:0;comment:已提现金额" json:"withdrawnBalance"`
	ContactEmail        *string    `gorm:"size:200;comment:联系邮箱" json:"contactEmail"`
	ContactPhone        *string    `gorm:"size:50;comment:联系电话" json:"contactPhone"`
	Country             *string    `gorm:"size:100;comment:国家" json:"country"`
	City                *string    `gorm:"size:100;comment:城市" json:"city"`
	Address             *string    `gorm:"type:text;comment:详细地址" json:"address"`
	Zipcode             *string    `gorm:"size:20;comment:邮编" json:"zipcode"`
	AutoSync            bool       `gorm:"default:true;comment:是否自动同步" json:"autoSync"`
	SyncInterval        int        `gorm:"default:3600;comment:同步间隔(秒)" json:"syncInterval"`
	SyncItems           bool       `gorm:"default:true;comment:是否同步商品" json:"syncItems"`
	SyncOrders          bool       `gorm:"default:true;comment:是否同步订单" json:"syncOrders"`
	SyncLogistics       bool       `gorm:"default:true;comment:是否同步物流" json:"syncLogistics"`
	SyncFinance         bool       `gorm:"default:true;comment:是否同步财务" json:"syncFinance"`
	IsPrimary           bool       `gorm:"default:false;comment:是否主店铺" json:"isPrimary"`
	LastSyncAt          *time.Time `gorm:"type:datetime;comment:上次同步时间" json:"lastSyncAt"`
	NextSyncAt          *time.Time `gorm:"type:datetime;comment:下次同步时间" json:"nextSyncAt"`
	ShopCreatedAt       *time.Time `gorm:"type:datetime;comment:店铺创建时间" json:"shopCreatedAt"`
	CreatedAt           time.Time  `gorm:"autoCreateTime;comment:记录创建时间" json:"createdAt"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime;comment:记录更新时间" json:"updatedAt"`
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
	ID               uint64    `gorm:"primaryKey;comment:主键ID" json:"id"`
	ShopID           uint64    `gorm:"uniqueIndex;not null;comment:店铺ID" json:"shop_id"`
	AccessToken      string    `gorm:"size:512;not null;comment:访问令牌" json:"access_token"`
	RefreshToken     string    `gorm:"size:512;not null;comment:刷新令牌" json:"refresh_token"`
	TokenType        string    `gorm:"size:50;not null;default:'Bearer';comment:令牌类型" json:"token_type"`
	ExpiresAt        time.Time `gorm:"not null;index;comment:访问令牌过期时间" json:"expires_at"`
	RefreshExpiresAt time.Time `gorm:"not null;comment:刷新令牌过期时间" json:"refresh_expires_at"`
	CreatedAt        time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
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
