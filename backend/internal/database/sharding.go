package database

import (
	"fmt"

	"gorm.io/gorm"
)

const (
	// ShardCount 分表数量
	ShardCount = 10
)

// GetShardIndex 根据shop_id计算分表索引
func GetShardIndex(shopID uint64) int {
	return int(shopID % ShardCount)
}

// GetOrderTableName 获取订单表名
func GetOrderTableName(shopID uint64) string {
	return fmt.Sprintf("orders_%d", GetShardIndex(shopID))
}

// GetOrderItemTableName 获取订单商品表名
func GetOrderItemTableName(shopID uint64) string {
	return fmt.Sprintf("order_items_%d", GetShardIndex(shopID))
}

// GetOrderAddressTableName 获取订单地址表名
func GetOrderAddressTableName(shopID uint64) string {
	return fmt.Sprintf("order_addresses_%d", GetShardIndex(shopID))
}

// GetOrderEscrowTableName 获取订单结算表名
func GetOrderEscrowTableName(shopID uint64) string {
	return fmt.Sprintf("order_escrows_%d", GetShardIndex(shopID))
}

// GetOrderEscrowItemTableName 获取订单结算商品表名
func GetOrderEscrowItemTableName(shopID uint64) string {
	return fmt.Sprintf("order_escrow_items_%d", GetShardIndex(shopID))
}

// GetOrderSettlementTableName 获取订单结算记录表名
func GetOrderSettlementTableName(shopID uint64) string {
	return fmt.Sprintf("order_settlements_%d", GetShardIndex(shopID))
}

// GetOrderShipmentRecordTableName 获取订单发货记录表名
func GetOrderShipmentRecordTableName(shopID uint64) string {
	return fmt.Sprintf("order_shipment_records_%d", GetShardIndex(shopID))
}

// GetShipmentTableName 获取发货记录表名
func GetShipmentTableName(shopID uint64) string {
	return fmt.Sprintf("shipments_%d", GetShardIndex(shopID))
}

// ShardedDB 分表数据库操作封装
type ShardedDB struct {
	db *gorm.DB
}

// NewShardedDB 创建分表数据库操作实例
func NewShardedDB(db *gorm.DB) *ShardedDB {
	return &ShardedDB{db: db}
}

// OrderTable 获取指定shop_id的订单表DB
func (s *ShardedDB) OrderTable(shopID uint64) *gorm.DB {
	return s.db.Table(GetOrderTableName(shopID))
}

// OrderItemTable 获取指定shop_id的订单商品表DB
func (s *ShardedDB) OrderItemTable(shopID uint64) *gorm.DB {
	return s.db.Table(GetOrderItemTableName(shopID))
}

// OrderAddressTable 获取指定shop_id的订单地址表DB
func (s *ShardedDB) OrderAddressTable(shopID uint64) *gorm.DB {
	return s.db.Table(GetOrderAddressTableName(shopID))
}

// OrderEscrowTable 获取指定shop_id的订单结算表DB
func (s *ShardedDB) OrderEscrowTable(shopID uint64) *gorm.DB {
	return s.db.Table(GetOrderEscrowTableName(shopID))
}

// OrderEscrowItemTable 获取指定shop_id的订单结算商品表DB
func (s *ShardedDB) OrderEscrowItemTable(shopID uint64) *gorm.DB {
	return s.db.Table(GetOrderEscrowItemTableName(shopID))
}

// OrderSettlementTable 获取指定shop_id的订单结算记录表DB
func (s *ShardedDB) OrderSettlementTable(shopID uint64) *gorm.DB {
	return s.db.Table(GetOrderSettlementTableName(shopID))
}

// OrderShipmentRecordTable 获取指定shop_id的订单发货记录表DB
func (s *ShardedDB) OrderShipmentRecordTable(shopID uint64) *gorm.DB {
	return s.db.Table(GetOrderShipmentRecordTableName(shopID))
}

// ShipmentTable 获取指定shop_id的发货记录表DB
func (s *ShardedDB) ShipmentTable(shopID uint64) *gorm.DB {
	return s.db.Table(GetShipmentTableName(shopID))
}

// ==================== 财务收入分表（按shop_id % 10）====================

// GetFinanceIncomeTableName 获取财务收入表名
func GetFinanceIncomeTableName(shopID uint64) string {
	return fmt.Sprintf("finance_incomes_%d", GetShardIndex(shopID))
}

// FinanceIncomeTable 获取指定shop_id的财务收入表DB
func (s *ShardedDB) FinanceIncomeTable(shopID uint64) *gorm.DB {
	return s.db.Table(GetFinanceIncomeTableName(shopID))
}

// ==================== 账户流水分表（按admin_id % 10）====================

// GetAccountTransactionTableName 获取账户流水表名
func GetAccountTransactionTableName(adminID int64) string {
	return fmt.Sprintf("account_transactions_%d", int(adminID%int64(ShardCount)))
}

// AccountTransactionTable 获取指定admin_id的账户流水表DB
func (s *ShardedDB) AccountTransactionTable(adminID int64) *gorm.DB {
	return s.db.Table(GetAccountTransactionTableName(adminID))
}

// ==================== 操作日志分表（按shop_id % 10）====================

// GetOperationLogTableName 获取操作日志表名
func GetOperationLogTableName(shopID uint64) string {
	return fmt.Sprintf("operation_logs_%d", GetShardIndex(shopID))
}

// OperationLogTable 获取指定shop_id的操作日志表DB
func (s *ShardedDB) OperationLogTable(shopID uint64) *gorm.DB {
	return s.db.Table(GetOperationLogTableName(shopID))
}

// AllOrderTables 遍历所有订单分表执行操作
func (s *ShardedDB) AllOrderTables(fn func(tableName string, db *gorm.DB) error) error {
	for i := 0; i < ShardCount; i++ {
		tableName := fmt.Sprintf("orders_%d", i)
		if err := fn(tableName, s.db.Table(tableName)); err != nil {
			return err
		}
	}
	return nil
}

// AllShardTables 遍历所有分表（指定前缀）执行操作
func (s *ShardedDB) AllShardTables(prefix string, fn func(tableName string, db *gorm.DB) error) error {
	for i := 0; i < ShardCount; i++ {
		tableName := fmt.Sprintf("%s_%d", prefix, i)
		if err := fn(tableName, s.db.Table(tableName)); err != nil {
			return err
		}
	}
	return nil
}
