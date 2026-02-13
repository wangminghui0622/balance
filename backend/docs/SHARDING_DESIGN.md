# 分表设计文档

## 1. 概述

Balance 系统采用水平分表策略，将高频访问的大表按照特定规则分散到多个物理表中，以提升查询性能和支持数据增长。

## 2. 分表规则

### 2.1 按 shop_id 分表

大部分业务表按 `shop_id % 10` 进行分表，确保同一店铺的数据在同一组分表中。

```
shop_id = 12345
shard_index = 12345 % 10 = 5
table_name = orders_5
```

### 2.2 按 admin_id 分表

账户流水表按 `admin_id % 10` 进行分表，确保同一用户的流水在同一表中。

```
admin_id = 67890
shard_index = 67890 % 10 = 0
table_name = account_transactions_0
```

## 3. 分表列表

### 3.1 订单相关（80个分表）

| 表名 | 分表数 | 分表规则 | 说明 |
|------|--------|----------|------|
| orders | 10 | shop_id % 10 | 订单主表 |
| order_items | 10 | shop_id % 10 | 订单商品 |
| order_addresses | 10 | shop_id % 10 | 收货地址 |
| order_escrows | 10 | shop_id % 10 | 结算明细 |
| order_escrow_items | 10 | shop_id % 10 | 结算商品 |
| order_settlements | 10 | shop_id % 10 | 结算记录 |
| order_shipment_records | 10 | shop_id % 10 | 发货记录 |
| shipments | 10 | shop_id % 10 | 发货单 |

### 3.2 财务相关（20个分表）

| 表名 | 分表数 | 分表规则 | 说明 |
|------|--------|----------|------|
| finance_incomes | 10 | shop_id % 10 | 财务收入 |
| operation_logs | 10 | shop_id % 10 | 操作日志 |

### 3.3 账户相关（10个分表）

| 表名 | 分表数 | 分表规则 | 说明 |
|------|--------|----------|------|
| account_transactions | 10 | admin_id % 10 | 账户流水 |

### 3.4 归档表（10个分表）

| 表名 | 分表数 | 分表规则 | 说明 |
|------|--------|----------|------|
| operation_logs_archive | 10 | shop_id % 10 | 日志归档 |

### 3.5 统计汇总表（不分表）

| 表名 | 说明 |
|------|------|
| order_daily_stats | 订单每日统计（按店铺） |
| finance_daily_stats | 财务每日统计（按店铺） |
| platform_daily_stats | 平台每日统计（汇总） |

## 4. 分表路由工具

### 4.1 工具函数

```go
// internal/database/sharding.go

// 获取分表索引
func GetShardIndex(shopID uint64) int {
    return int(shopID % ShardCount)
}

// 获取订单表名
func GetOrderTableName(shopID uint64) string {
    return fmt.Sprintf("orders_%d", GetShardIndex(shopID))
}

// 获取账户流水表名
func GetAccountTransactionTableName(adminID int64) string {
    return fmt.Sprintf("account_transactions_%d", int(adminID%int64(ShardCount)))
}
```

### 4.2 ShardedDB 封装

```go
// ShardedDB 分表数据库封装
type ShardedDB struct {
    db *gorm.DB
}

// OrderTable 获取订单表
func (s *ShardedDB) OrderTable(shopID uint64) *gorm.DB {
    return s.db.Table(GetOrderTableName(shopID))
}

// AccountTransactionTable 获取账户流水表
func (s *ShardedDB) AccountTransactionTable(adminID int64) *gorm.DB {
    return s.db.Table(GetAccountTransactionTableName(adminID))
}
```

## 5. 查询模式

### 5.1 单店铺查询

直接路由到对应分表：

```go
// 查询单个店铺的订单
orderTable := database.GetOrderTableName(shopID)
db.Table(orderTable).Where("shop_id = ? AND order_sn = ?", shopID, orderSN).First(&order)
```

### 5.2 多店铺查询

按分表索引分组，遍历查询后内存合并：

```go
// 查询多个店铺的订单
shardShops := make(map[int][]uint64)
for _, sid := range shopIDs {
    idx := database.GetShardIndex(sid)
    shardShops[idx] = append(shardShops[idx], sid)
}

var allOrders []models.Order
for idx, sids := range shardShops {
    orderTable := fmt.Sprintf("orders_%d", idx)
    var orders []models.Order
    db.Table(orderTable).Where("shop_id IN ?", sids).Find(&orders)
    allOrders = append(allOrders, orders...)
}
```

### 5.3 全量查询（平台级）

遍历所有分表，内存合并分页：

```go
// 平台级统计查询
var total int64
for i := 0; i < database.ShardCount; i++ {
    orderTable := fmt.Sprintf("orders_%d", i)
    var count int64
    db.Table(orderTable).Where("order_status = ?", status).Count(&count)
    total += count
}
```

### 5.4 使用汇总表（推荐）

平台级查询优先使用汇总表：

```go
// 使用汇总表查询
db.Model(&models.PlatformDailyStat{}).
    Where("stat_date >= ? AND stat_date <= ?", startDate, endDate).
    Select("SUM(total_orders) as total_orders").
    Scan(&result)
```

## 6. 事务处理

### 6.1 同一店铺事务

同一店铺的数据在同一组分表中，可以使用数据库事务：

```go
db.Transaction(func(tx *gorm.DB) error {
    orderTable := database.GetOrderTableName(shopID)
    itemTable := database.GetOrderItemTableName(shopID)
    
    // 创建订单
    if err := tx.Table(orderTable).Create(&order).Error; err != nil {
        return err
    }
    
    // 创建订单商品
    if err := tx.Table(itemTable).Create(&items).Error; err != nil {
        return err
    }
    
    return nil
})
```

### 6.2 跨店铺事务

跨店铺操作需要使用分布式事务或最终一致性：

```go
// 使用 Redis 分布式锁 + 补偿机制
lockKey := fmt.Sprintf("lock:settlement:%d", settlementID)
if acquired := redis.SetNX(lockKey, "1", 5*time.Minute); !acquired {
    return errors.New("获取锁失败")
}
defer redis.Del(lockKey)

// 执行操作...
```

## 7. 索引设计

### 7.1 分表索引

每个分表都包含以下索引：

```sql
-- 订单表索引
CREATE INDEX idx_shop_order ON orders_X (shop_id, order_sn);
CREATE INDEX idx_order_status ON orders_X (order_status);
CREATE INDEX idx_create_time ON orders_X (create_time);
```

### 7.2 查询优化

- 查询时必须带 `shop_id` 条件，确保路由到正确分表
- 避免跨分表 JOIN，改用应用层关联
- 大量数据查询使用游标分页

## 8. 数据迁移

### 8.1 新增分表

如需扩展分表数量（如从10个扩展到100个）：

1. 创建新分表
2. 修改分表路由规则
3. 迁移历史数据
4. 切换路由

### 8.2 迁移脚本示例

```sql
-- 创建新分表
CALL create_orders_shards_100();

-- 迁移数据
INSERT INTO orders_new_X 
SELECT * FROM orders_old_Y 
WHERE shop_id % 100 = X;
```

## 9. 监控指标

| 指标 | 说明 |
|------|------|
| shard_table_rows | 各分表数据量 |
| shard_query_latency | 分表查询延迟 |
| cross_shard_queries | 跨分表查询次数 |

## 10. 最佳实践

1. **查询必带 shop_id**：确保路由到正确分表
2. **避免跨分表 JOIN**：使用应用层关联
3. **使用汇总表**：平台级统计优先使用汇总表
4. **定期归档**：日志表定期归档到归档表
5. **监控数据量**：单表超过500万行时考虑扩展
