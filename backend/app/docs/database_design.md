# 虾皮店铺订单发货平台 - 数据库设计

## MySQL 数据库设计

### 1. 店铺表 (shops)
存储授权店铺的基本信息

```sql
CREATE TABLE `shops` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '虾皮店铺ID',
    `shop_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '店铺名称',
    `region` VARCHAR(10) NOT NULL COMMENT '地区代码(SG/MY/TH/TW/VN/PH/BR/MX/CO/CL)',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-正常 0-禁用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_id` (`shop_id`),
    KEY `idx_region` (`region`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺表';
```

### 2. 店铺授权表 (shop_authorizations)
存储店铺的OAuth授权信息

```sql
CREATE TABLE `shop_authorizations` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '虾皮店铺ID',
    `access_token` VARCHAR(512) NOT NULL COMMENT '访问令牌',
    `refresh_token` VARCHAR(512) NOT NULL COMMENT '刷新令牌',
    `token_type` VARCHAR(50) NOT NULL DEFAULT 'Bearer' COMMENT '令牌类型',
    `expires_at` DATETIME NOT NULL COMMENT '访问令牌过期时间',
    `refresh_expires_at` DATETIME NOT NULL COMMENT '刷新令牌过期时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_id` (`shop_id`),
    KEY `idx_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺授权表';
```

### 3. 订单表 (orders)
存储从虾皮同步过来的订单信息

```sql
CREATE TABLE `orders` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '虾皮店铺ID',
    `order_sn` VARCHAR(64) NOT NULL COMMENT '虾皮订单号',
    `region` VARCHAR(10) NOT NULL COMMENT '地区代码',
    `order_status` VARCHAR(50) NOT NULL COMMENT '订单状态',
    `buyer_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '买家用户ID',
    `buyer_username` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '买家用户名',
    `total_amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '订单总金额',
    `currency` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '货币代码',
    `shipping_carrier` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '物流商',
    `tracking_number` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '物流单号',
    `ship_by_date` DATETIME DEFAULT NULL COMMENT '最晚发货时间',
    `pay_time` DATETIME DEFAULT NULL COMMENT '付款时间',
    `create_time` DATETIME DEFAULT NULL COMMENT '订单创建时间(虾皮)',
    `update_time` DATETIME DEFAULT NULL COMMENT '订单更新时间(虾皮)',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '本地创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '本地更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_order` (`shop_id`, `order_sn`),
    KEY `idx_order_sn` (`order_sn`),
    KEY `idx_order_status` (`order_status`),
    KEY `idx_ship_by_date` (`ship_by_date`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';
```

### 4. 订单商品表 (order_items)
存储订单中的商品明细

```sql
CREATE TABLE `order_items` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `order_id` BIGINT UNSIGNED NOT NULL COMMENT '本地订单ID',
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '虾皮店铺ID',
    `order_sn` VARCHAR(64) NOT NULL COMMENT '虾皮订单号',
    `item_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    `item_name` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '商品名称',
    `item_sku` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '商品SKU',
    `model_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '规格ID',
    `model_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '规格名称',
    `model_sku` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '规格SKU',
    `quantity` INT NOT NULL DEFAULT 0 COMMENT '数量',
    `item_price` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '商品单价',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_shop_order` (`shop_id`, `order_sn`),
    KEY `idx_item_id` (`item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单商品表';
```

### 5. 收货地址表 (order_addresses)
存储订单的收货地址信息

```sql
CREATE TABLE `order_addresses` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `order_id` BIGINT UNSIGNED NOT NULL COMMENT '本地订单ID',
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '虾皮店铺ID',
    `order_sn` VARCHAR(64) NOT NULL COMMENT '虾皮订单号',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '收货人姓名',
    `phone` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '电话号码',
    `town` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '乡镇',
    `district` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '区',
    `city` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '城市',
    `state` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '省/州',
    `region` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '国家/地区代码',
    `zipcode` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '邮编',
    `full_address` TEXT COMMENT '完整地址',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_id` (`order_id`),
    KEY `idx_shop_order` (`shop_id`, `order_sn`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收货地址表';
```

### 6. 发货记录表 (shipments)
存储发货操作记录

```sql
CREATE TABLE `shipments` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '虾皮店铺ID',
    `order_sn` VARCHAR(64) NOT NULL COMMENT '虾皮订单号',
    `package_number` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '包裹号',
    `shipping_carrier` VARCHAR(100) NOT NULL COMMENT '物流商',
    `tracking_number` VARCHAR(100) NOT NULL COMMENT '物流单号',
    `ship_status` TINYINT NOT NULL DEFAULT 0 COMMENT '发货状态: 0-待发货 1-已发货 2-发货失败',
    `ship_time` DATETIME DEFAULT NULL COMMENT '发货时间',
    `error_message` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '错误信息',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_order` (`shop_id`, `order_sn`),
    KEY `idx_tracking_number` (`tracking_number`),
    KEY `idx_ship_status` (`ship_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='发货记录表';
```

### 7. 物流商配置表 (logistics_channels)
存储物流商配置信息

```sql
CREATE TABLE `logistics_channels` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '虾皮店铺ID',
    `logistics_channel_id` BIGINT UNSIGNED NOT NULL COMMENT '虾皮物流渠道ID',
    `logistics_channel_name` VARCHAR(255) NOT NULL COMMENT '物流渠道名称',
    `cod_enabled` TINYINT NOT NULL DEFAULT 0 COMMENT '是否支持货到付款: 0-否 1-是',
    `enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用: 0-否 1-是',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_channel` (`shop_id`, `logistics_channel_id`),
    KEY `idx_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='物流渠道配置表';
```

### 8. 操作日志表 (operation_logs)
存储系统操作日志

```sql
CREATE TABLE `operation_logs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `shop_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '店铺ID',
    `order_sn` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '订单号',
    `operation_type` VARCHAR(50) NOT NULL COMMENT '操作类型',
    `operation_desc` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '操作描述',
    `request_data` TEXT COMMENT '请求数据',
    `response_data` TEXT COMMENT '响应数据',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1-成功 0-失败',
    `ip` VARCHAR(50) NOT NULL DEFAULT '' COMMENT 'IP地址',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_shop_id` (`shop_id`),
    KEY `idx_order_sn` (`order_sn`),
    KEY `idx_operation_type` (`operation_type`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';
```

---

## Redis 缓存设计

### 1. 店铺Token缓存
用于缓存店铺的access_token，避免频繁查库

**Key格式:** `shopee:token:{shop_id}`
**Value类型:** Hash
**过期时间:** 根据token过期时间动态设置，提前5分钟过期
**字段:**
```
- access_token: 访问令牌
- refresh_token: 刷新令牌
- expires_at: 过期时间戳
```

### 2. 店铺信息缓存
缓存店铺基本信息

**Key格式:** `shopee:shop:{shop_id}`
**Value类型:** Hash
**过期时间:** 1小时
**字段:**
```
- shop_name: 店铺名称
- region: 地区代码
- status: 状态
```

### 3. 订单状态缓存
缓存订单状态，用于快速判断订单是否可发货

**Key格式:** `shopee:order:status:{shop_id}:{order_sn}`
**Value类型:** String
**过期时间:** 30分钟
**Value:** 订单状态值

### 4. 订单同步锁
防止同一店铺同时多次同步订单

**Key格式:** `shopee:lock:sync:{shop_id}`
**Value类型:** String
**过期时间:** 5分钟
**Value:** 锁定时间戳

### 5. 发货操作锁
防止同一订单重复发货

**Key格式:** `shopee:lock:ship:{shop_id}:{order_sn}`
**Value类型:** String
**过期时间:** 1分钟
**Value:** 锁定时间戳

### 6. API限流计数器
用于控制API请求频率，避免超出虾皮API限制

**Key格式:** `shopee:ratelimit:{shop_id}:{api_name}`
**Value类型:** String (计数器)
**过期时间:** 1分钟
**Value:** 请求次数

### 7. 待发货订单队列
存储待处理的发货订单

**Key格式:** `shopee:queue:ship`
**Value类型:** List
**Value格式:** JSON字符串 `{"shop_id":xxx,"order_sn":"xxx","tracking_number":"xxx"}`

### 8. 店铺物流渠道缓存
缓存店铺的物流渠道列表

**Key格式:** `shopee:logistics:{shop_id}`
**Value类型:** String (JSON数组)
**过期时间:** 1小时
**Value:** 物流渠道列表JSON

---

## 数据库索引优化建议

1. **订单表** - 按照查询频率，主要索引在 `order_status` 和 `ship_by_date` 上
2. **发货记录表** - 主要索引在 `ship_status` 上，用于查询待发货订单
3. **操作日志表** - 建议定期归档，保留最近3个月数据

## 分表策略建议

当订单量达到千万级别时，建议：
1. `orders` 表按 `shop_id` 进行分表
2. `order_items` 表随 `orders` 表同步分表
3. `operation_logs` 表按时间（月份）进行分表
