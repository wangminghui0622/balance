package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// ShopeeShop 店铺表
type ShopeeShop struct {
	ID                  uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	ShopID              int64          `gorm:"column:shop_id;type:bigint;not null;uniqueIndex:uk_shop_id;comment:虾皮店铺ID" json:"shopId"`
	ShopIDStr           string         `gorm:"column:shop_id_str;type:varchar(64);not null;uniqueIndex:uk_shop_id_str;comment:虾皮店铺ID(字符串，用于API关联)" json:"shopIdStr"`
	ShopName            string         `gorm:"column:shop_name;type:varchar(64);not null;index:idx_shop_name;comment:店铺名称" json:"shopName"`
	ShopSlug            *string        `gorm:"column:shop_slug;type:varchar(256);comment:店铺短链接名" json:"shopSlug"`
	Region              string         `gorm:"column:region;type:varchar(16);not null;index:idx_region;comment:地区编码: MY/ID/TH/VN/PH/SG/TW/BR" json:"region"`
	PartnerID           int64          `gorm:"column:partner_id;type:bigint;not null;comment:合作伙伴ID" json:"partnerId"`
	AccessToken         *string        `gorm:"column:access_token;type:varchar(500);comment:访问令牌" json:"accessToken"`
	RefreshToken        *string        `gorm:"column:refresh_token;type:varchar(500);comment:刷新令牌" json:"refreshToken"`
	TokenExpireAt       *time.Time     `gorm:"column:token_expire_at;type:datetime;index:idx_token_expire_at;comment:令牌过期时间" json:"tokenExpireAt"`
	AuthStatus          int16          `gorm:"column:auth_status;type:smallint;default:0;index:idx_auth_status;comment:授权状态: 0未授权 1已授权 2已过期 3已撤销" json:"authStatus"`
	AuthTime            *time.Time     `gorm:"column:auth_time;type:datetime;comment:授权时间" json:"authTime"`
	ExpireTime          int64          `gorm:"column:expire_time;type:bigint;comment:授权到期时间" json:"expireTime"`
	LastTokenRefresh    *time.Time     `gorm:"column:last_token_refresh;type:datetime;comment:最后令牌刷新时间" json:"lastTokenRefresh"`
	Status              int16          `gorm:"column:status;type:smallint;default:1;index:idx_status;comment:店铺状态: 1正常 2禁用 3冻结 4关闭" json:"status"`
	SuspensionStatus    int16          `gorm:"column:suspension_status;type:smallint;default:0;comment:虾皮平台状态: 0正常 1警告 2限制 3暂停" json:"suspensionStatus"`
	IsCbShop            bool           `gorm:"column:is_cb_shop;type:tinyint(1);default:0;comment:是否跨境店铺" json:"isCbShop"`
	IsCodShop           bool           `gorm:"column:is_cod_shop;type:tinyint(1);default:0;comment:是否支持货到付款" json:"isCodShop"`
	IsPreferredPlusShop bool           `gorm:"column:is_preferred_plus_shop;type:tinyint(1);default:0;comment:是否优选+店铺" json:"isPreferredPlusShop"`
	IsShopeeVerified    bool           `gorm:"column:is_shopee_verified;type:tinyint(1);default:0;comment:是否虾皮认证店铺" json:"isShopeeVerified"`
	RatingStar          float64        `gorm:"column:rating_star;type:float;default:0;comment:店铺评分(0-5)" json:"ratingStar"`
	RatingBad           int            `gorm:"column:rating_bad;type:int;default:0;comment:差评数" json:"ratingBad"`
	RatingGood          int            `gorm:"column:rating_good;type:int;default:0;comment:好评数" json:"ratingGood"`
	RatingNormal        int            `gorm:"column:rating_normal;type:int;default:0;comment:中评数" json:"ratingNormal"`
	ItemCount           int            `gorm:"column:item_count;type:int;default:0;comment:商品总数" json:"itemCount"`
	FollowerCount       int            `gorm:"column:follower_count;type:int;default:0;comment:粉丝数" json:"followerCount"`
	ResponseRate        float64        `gorm:"column:response_rate;type:float;default:0;comment:响应率(%)" json:"responseRate"`
	ResponseTime        int            `gorm:"column:response_time;type:int;default:0;comment:平均响应时间(秒)" json:"responseTime"`
	CancellationRate    float64        `gorm:"column:cancellation_rate;type:float;default:0;comment:取消率(%)" json:"cancellationRate"`
	ShopCreatedAt       *time.Time     `gorm:"column:shop_created_at;type:datetime;comment:店铺在虾皮的创建时间" json:"shopCreatedAt"`
	LastSyncAt          *time.Time     `gorm:"column:last_sync_at;type:datetime;index:idx_last_sync_at;comment:最后同步时间" json:"lastSyncAt"`
	NextSyncAt          *time.Time     `gorm:"column:next_sync_at;type:datetime;comment:下次同步时间" json:"nextSyncAt"`
	TotalSales          int            `gorm:"column:total_sales;type:int;default:0;comment:总销量" json:"totalSales"`
	TotalOrders         int            `gorm:"column:total_orders;type:int;default:0;comment:总订单数" json:"totalOrders"`
	TotalViews          int            `gorm:"column:total_views;type:int;default:0;comment:总浏览量" json:"totalViews"`
	DailySales          int            `gorm:"column:daily_sales;type:int;default:0;comment:日销量" json:"dailySales"`
	MonthlySales        int            `gorm:"column:monthly_sales;type:int;default:0;comment:月销量" json:"monthlySales"`
	YearlySales         int            `gorm:"column:yearly_sales;type:int;default:0;comment:年销量" json:"yearlySales"`
	Currency            string         `gorm:"column:currency;type:varchar(10);default:'MYR';comment:货币: MYR/IDR/THB/VND/PHP/SGD/TWD/BRL" json:"currency"`
	Balance             float64        `gorm:"column:balance;type:decimal(12,2);default:0.00;comment:账户余额" json:"balance"`
	PendingBalance      float64        `gorm:"column:pending_balance;type:decimal(12,2);default:0.00;comment:待结算金额" json:"pendingBalance"`
	WithdrawnBalance    float64        `gorm:"column:withdrawn_balance;type:decimal(12,2);default:0.00;comment:已提现金额" json:"withdrawnBalance"`
	ContactEmail        *string        `gorm:"column:contact_email;type:varchar(200);comment:联系邮箱" json:"contactEmail"`
	ContactPhone        *string        `gorm:"column:contact_phone;type:varchar(50);comment:联系电话" json:"contactPhone"`
	Country             *string        `gorm:"column:country;type:varchar(100);comment:国家" json:"country"`
	City                *string        `gorm:"column:city;type:varchar(100);comment:城市" json:"city"`
	Address             *string        `gorm:"column:address;type:text;comment:详细地址" json:"address"`
	Zipcode             *string        `gorm:"column:zipcode;type:varchar(20);comment:邮编" json:"zipcode"`
	AutoSync            bool           `gorm:"column:auto_sync;type:tinyint(1);default:1;comment:是否自动同步" json:"autoSync"`
	SyncInterval        int            `gorm:"column:sync_interval;type:int;default:3600;comment:同步间隔(秒)" json:"syncInterval"`
	SyncItems           bool           `gorm:"column:sync_items;type:tinyint(1);default:1;comment:是否同步商品" json:"syncItems"`
	SyncOrders          bool           `gorm:"column:sync_orders;type:tinyint(1);default:1;comment:是否同步订单" json:"syncOrders"`
	SyncLogistics       bool           `gorm:"column:sync_logistics;type:tinyint(1);default:1;comment:是否同步物流" json:"syncLogistics"`
	SyncFinance         bool           `gorm:"column:sync_finance;type:tinyint(1);default:1;comment:是否同步财务" json:"syncFinance"`
	DefaultLogisticID   *int64         `gorm:"column:default_logistic_id;type:bigint;comment:默认物流方式ID" json:"defaultLogisticId"`
	LogisticsJSON       *JSONField     `gorm:"column:logistics_json;type:json;comment:物流方式配置" json:"logisticsJson"`
	ShippingFeeConfig   *JSONField     `gorm:"column:shipping_fee_config;type:json;comment:运费配置" json:"shippingFeeConfig"`
	OwnerID             *int64         `gorm:"column:owner_id;type:bigint;index:idx_owner_id;comment:所属用户ID" json:"ownerId"`
	MerchantID          *int64         `gorm:"column:merchant_id;type:bigint;comment:商户ID" json:"merchantId"`
	Version             int            `gorm:"column:version;type:int;default:0;index:idx_version;comment:版本号(乐观锁)" json:"version"`
	CreatedAt           time.Time      `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;index:idx_created_at;comment:创建时间" json:"createdAt"`
	UpdatedAt           time.Time      `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
	DeletedAt           gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:删除时间" json:"deletedAt"`
}

// TableName 指定表名
func (ShopeeShop) TableName() string {
	return "shopee_shops"
}

// JSONField 用于处理 JSON 字段
type JSONField map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONField) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// ShopeeShopRepository ShopeeShop 数据访问层
type ShopeeShopRepository struct {
	db *gorm.DB
}

// NewShopeeShopRepository 创建 ShopeeShopRepository 实例
func NewShopeeShopRepository(db *gorm.DB) *ShopeeShopRepository {
	return &ShopeeShopRepository{db: db}
}

// CreateOrUpdate 创建或更新店铺（根据 shop_id）
func (r *ShopeeShopRepository) CreateOrUpdate(shop *ShopeeShop) error {
	var existing ShopeeShop
	err := r.db.Where("shop_id = ?", shop.ShopID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		return r.db.Create(shop).Error
	} else if err != nil {
		return err
	}

	// 存在，更新记录
	return r.db.Model(&existing).Updates(shop).Error
}

// GetByShopID 根据 shop_id 查询店铺
func (r *ShopeeShopRepository) GetByShopID(shopID int64) (*ShopeeShop, error) {
	var shop ShopeeShop
	err := r.db.Where("shop_id = ?", shopID).First(&shop).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// GetByOwnerID 根据 owner_id 查询店铺列表（使用关联表）
func (r *ShopeeShopRepository) GetByOwnerID(ownerID int64) ([]*ShopeeShop, error) {
	var shops []*ShopeeShop
	// 使用关联表查询
	err := r.db.Table("shopee_shops").
		Joins("INNER JOIN admin_shop ON shopee_shops.shop_id = admin_shop.shop_id").
		Where("admin_shop.admin_id = ? AND admin_shop.deleted_at IS NULL", ownerID).
		Find(&shops).Error
	if err != nil {
		return nil, err
	}
	return shops, nil
}

// UpdateOwnerID 更新店铺的 owner_id
func (r *ShopeeShopRepository) UpdateOwnerID(shopID int64, ownerID int64) error {
	return r.db.Model(&ShopeeShop{}).Where("shop_id = ?", shopID).Update("owner_id", ownerID).Error
}

// UpdateAuthStatus 更新授权状态
func (r *ShopeeShopRepository) UpdateAuthStatus(shopID int64, authStatus int16, authTime *time.Time) error {
	updates := map[string]interface{}{
		"auth_status": authStatus,
	}
	if authTime != nil {
		updates["auth_time"] = authTime
	}
	return r.db.Model(&ShopeeShop{}).Where("shop_id = ?", shopID).Updates(updates).Error
}
