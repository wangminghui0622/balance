package models

import (
	"time"

	"gorm.io/gorm"
)

// Admin 管理员/用户模型
type Admin struct {
	ID        int64          `gorm:"primaryKey;column:id" json:"id"`
	UserNo    string         `gorm:"column:user_no;size:32" json:"user_no"`
	UserType  int8           `gorm:"column:user_type;default:1" json:"user_type"` // 1=店铺 5=运营 9=平台
	Avatar    string         `gorm:"column:avatar;size:100;default:''" json:"avatar"`
	UserName  string         `gorm:"column:user_name;size:32;not null;uniqueIndex" json:"user_name"`
	RealName  string         `gorm:"column:real_name;size:64" json:"real_name"`
	Salt      string         `gorm:"column:salt;size:16" json:"-"`
	Hash      string         `gorm:"column:hash;size:64" json:"-"`
	Email     string         `gorm:"column:email;size:128" json:"email"`
	Phone     string         `gorm:"column:phone;size:16" json:"phone"`
	LineID    string         `gorm:"column:line_id;size:64" json:"line_id"`
	Wechat    string         `gorm:"column:wechat;size:64" json:"wechat"`
	Status    int8           `gorm:"column:status;default:1" json:"status"` // 1=正常 2=禁用
	Language  string         `gorm:"column:language;size:10;default:'zh'" json:"language"`
	Remark    string         `gorm:"column:remark;size:500" json:"remark"`
	LoginIP   string         `gorm:"column:login_ip;size:128;default:''" json:"login_ip"`
	LoginDate *time.Time     `gorm:"column:login_date" json:"login_date"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (Admin) TableName() string {
	return "admin"
}

// 用户类型常量
const (
	UserTypeShopOwner = 1 // 店主
	UserTypeOperator  = 5 // 运营
	UserTypePlatform  = 9 // 平台
)

// 用户状态常量
const (
	UserStatusNormal   = 1 // 正常
	UserStatusDisabled = 2 // 禁用
)

// IsDisabled 检查用户是否被禁用
func (a *Admin) IsDisabled() bool {
	return a.Status == UserStatusDisabled
}
