package models

import (
	"time"

	"gorm.io/gorm"
)

// Admin 用户信息表
type Admin struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement;comment:用户ID" json:"id"`
	UserNo    string         `gorm:"column:user_no;type:varchar(32);index:idx_user_no;comment:用户编号" json:"userNo"`
	UserType  int8           `gorm:"column:user_type;type:tinyint;default:1;comment:用户类型（1.店铺 5.运营 9.平台）" json:"userType"`
	Avatar    string         `gorm:"column:avatar;type:varchar(100);default:'';comment:头像地址" json:"avatar"`
	UserName  string         `gorm:"column:user_name;type:varchar(32);not null;index:idx_user_name;comment:用户名" json:"userName"`
	Salt      string         `gorm:"column:salt;type:varchar(16);comment:密码盐" json:"salt"`
	Hash      string         `gorm:"column:hash;type:varchar(64);comment:密码hash" json:"hash"`
	Email     string         `gorm:"column:email;type:varchar(128);comment:邮箱" json:"email"`
	Phone     string         `gorm:"column:phone;type:varchar(16);comment:手机号" json:"phone"`
	Status    int8           `gorm:"column:status;type:tinyint;default:1;comment:状态 1正常 2禁用" json:"status"`
	Language  string         `gorm:"column:language;type:varchar(10);default:'zh';comment:界面语言 zh/en" json:"language"`
	Remark    string         `gorm:"column:remark;type:varchar(500);comment:备注" json:"remark"`
	LoginIP   string         `gorm:"column:login_ip;type:varchar(128);default:'';comment:最后登录IP" json:"loginIp"`
	LoginDate *time.Time     `gorm:"column:login_date;comment:最后登录时间" json:"loginDate"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:删除时间（软删除）" json:"deletedAt"`
}

// TableName 指定表名
func (Admin) TableName() string {
	return "admin"
}

// AdminRepository Admin数据访问层
type AdminRepository struct {
	db *gorm.DB
}

// NewAdminRepository 创建AdminRepository实例
func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// Create 创建管理员
func (r *AdminRepository) Create(admin *Admin) error {
	return r.db.Create(admin).Error
}

// GetByID 根据ID查询管理员
func (r *AdminRepository) GetByID(id int64) (*Admin, error) {
	var admin Admin
	err := r.db.Where("id = ?", id).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetByUserName 根据用户名查询管理员
func (r *AdminRepository) GetByUserName(userName string) (*Admin, error) {
	var admin Admin
	err := r.db.Where("user_name = ?", userName).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetByUserNo 根据用户编号查询管理员
func (r *AdminRepository) GetByUserNo(userNo string) (*Admin, error) {
	var admin Admin
	err := r.db.Where("user_no = ?", userNo).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// Update 更新管理员信息
func (r *AdminRepository) Update(admin *Admin) error {
	return r.db.Save(admin).Error
}

// UpdateByID 根据ID更新指定字段
func (r *AdminRepository) UpdateByID(id int64, updates map[string]interface{}) error {
	return r.db.Model(&Admin{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 软删除管理员
func (r *AdminRepository) Delete(id int64) error {
	return r.db.Delete(&Admin{}, id).Error
}

// HardDelete 硬删除管理员
func (r *AdminRepository) HardDelete(id int64) error {
	return r.db.Unscoped().Delete(&Admin{}, id).Error
}

// List 查询管理员列表
func (r *AdminRepository) List(page, pageSize int, filters map[string]interface{}) ([]*Admin, int64, error) {
	var admins []*Admin
	var total int64

	query := r.db.Model(&Admin{})

	// 应用过滤条件
	if userType, ok := filters["user_type"]; ok {
		query = query.Where("user_type = ?", userType)
	}
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if userName, ok := filters["user_name"]; ok {
		query = query.Where("user_name LIKE ?", "%"+userName.(string)+"%")
	}
	if email, ok := filters["email"]; ok {
		query = query.Where("email LIKE ?", "%"+email.(string)+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&admins).Error
	if err != nil {
		return nil, 0, err
	}

	return admins, total, nil
}

// UpdateLoginInfo 更新登录信息
func (r *AdminRepository) UpdateLoginInfo(id int64, loginIP string) error {
	now := time.Now()
	return r.db.Model(&Admin{}).Where("id = ?", id).Updates(map[string]interface{}{
		"login_ip":   loginIP,
		"login_date": now,
	}).Error
}

// ExistsByUserName 检查用户名是否存在
func (r *AdminRepository) ExistsByUserName(userName string) (bool, error) {
	var count int64
	err := r.db.Model(&Admin{}).Where("user_name = ?", userName).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *AdminRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&Admin{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByPhone 检查手机号是否存在
func (r *AdminRepository) ExistsByPhone(phone string) (bool, error) {
	var count int64
	err := r.db.Model(&Admin{}).Where("phone = ?", phone).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
