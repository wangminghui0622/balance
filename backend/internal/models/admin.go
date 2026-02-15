package models

import (
	"time"

	"gorm.io/gorm"
)

// Admin 管理员/用户模型
type Admin struct {
	ID        int64          `gorm:"primaryKey;column:id;comment:主键ID" json:"id"`
	UserNo    string         `gorm:"column:user_no;size:32;comment:用户编号" json:"user_no"`
	UserType  int8           `gorm:"column:user_type;default:1;comment:用户类型(1店主/5运营/9平台)" json:"user_type"`
	Avatar    string         `gorm:"column:avatar;size:100;default:'';comment:头像URL" json:"avatar"`
	UserName  string         `gorm:"column:user_name;size:32;not null;uniqueIndex;comment:用户名" json:"user_name"`
	RealName  string         `gorm:"column:real_name;size:64;comment:真实姓名" json:"real_name"`
	Salt      string         `gorm:"column:salt;size:16;comment:密码盐值" json:"-"`
	Hash      string         `gorm:"column:hash;size:64;comment:密码哈希" json:"-"`
	Email     string         `gorm:"column:email;size:128;comment:邮箱" json:"email"`
	Phone     string         `gorm:"column:phone;size:16;comment:手机号" json:"phone"`
	LineID    string         `gorm:"column:line_id;size:64;comment:Line ID" json:"line_id"`
	Wechat    string         `gorm:"column:wechat;size:64;comment:微信号" json:"wechat"`
	Status    int8           `gorm:"column:status;default:1;comment:状态(1正常/2禁用)" json:"status"`
	Language  string         `gorm:"column:language;size:10;default:'zh';comment:语言偏好" json:"language"`
	Remark    string         `gorm:"column:remark;size:500;comment:备注" json:"remark"`
	LoginIP   string         `gorm:"column:login_ip;size:128;default:'';comment:最后登录IP" json:"login_ip"`
	LoginDate *time.Time     `gorm:"column:login_date;comment:最后登录时间" json:"login_date"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"-"`
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
