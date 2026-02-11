-- ============================================================
-- 虾皮店铺订单发货平台 - MySQL 完整初始化脚本
-- 数据库: shopeex
-- 字符集: utf8mb4
-- ============================================================

CREATE DATABASE IF NOT EXISTS `shopeex` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `shopeex`;

-- ============================================================
-- 1. 用户表 (admin)
-- ID由Redis+Lua脚本生成，不使用AUTO_INCREMENT
-- 店主ID以1开头，运营ID以5开头，平台ID以9开头
-- ============================================================
CREATE TABLE IF NOT EXISTS `admin` (
  `id` BIGINT NOT NULL COMMENT '用户ID(由IDGenerator生成)',
  `user_no` VARCHAR(32) DEFAULT NULL COMMENT '用户编号(U+11位ID)',
  `user_type` TINYINT DEFAULT 1 COMMENT '用户类型: 1店主 5运营 9平台',
  `avatar` VARCHAR(100) DEFAULT '' COMMENT '头像',
  `user_name` VARCHAR(32) NOT NULL COMMENT '用户名',
  `salt` VARCHAR(16) DEFAULT NULL COMMENT '密码盐',
  `hash` VARCHAR(64) DEFAULT NULL COMMENT '密码哈希',
  `email` VARCHAR(128) DEFAULT NULL COMMENT '邮箱',
  `phone` VARCHAR(16) DEFAULT NULL COMMENT '手机号',
  `status` TINYINT DEFAULT 1 COMMENT '状态: 1正常 2禁用',
  `language` VARCHAR(10) DEFAULT 'zh' COMMENT '语言: zh/en',
  `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
  `login_ip` VARCHAR(128) DEFAULT '' COMMENT '最后登录IP',
  `login_date` DATETIME(3) DEFAULT NULL COMMENT '最后登录时间',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_name` (`user_name`),
  KEY `idx_user_no` (`user_no`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ============================================================
-- 2. 店铺表 (shops)
-- 注意: 授权时间从 shop_authorizations 表获取
-- ============================================================
CREATE TABLE IF NOT EXISTS `shops` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '虾皮店铺ID',
    `shop_id_str` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '店铺ID字符串',
    `admin_id` BIGINT NOT NULL DEFAULT 0 COMMENT '关联用户ID',
    `shop_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '店铺名称',
    `shop_slug` VARCHAR(256) DEFAULT NULL COMMENT '店铺短链接',
    `region` VARCHAR(16) NOT NULL COMMENT '地区(SG/MY/TH/TW/VN/PH/ID/BR/MX)',
    `partner_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合作伙伴ID',
    `auth_status` TINYINT NOT NULL DEFAULT 0 COMMENT '授权状态: 0未授权 1已授权 2已过期',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '店铺状态: 1正常 2禁用',
    `suspension_status` TINYINT NOT NULL DEFAULT 0 COMMENT '平台状态: 0正常 1警告 2限制 3暂停',
    `is_cb_shop` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '跨境店铺',
    `is_cod_shop` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '支持货到付款',
    `is_preferred_plus_shop` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '优选+店铺',
    `is_shopee_verified` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '虾皮认证',
    `rating_star` DECIMAL(3,2) NOT NULL DEFAULT 0.00 COMMENT '评分(0-5)',
    `rating_bad` INT NOT NULL DEFAULT 0 COMMENT '差评数',
    `rating_good` INT NOT NULL DEFAULT 0 COMMENT '好评数',
    `rating_normal` INT NOT NULL DEFAULT 0 COMMENT '中评数',
    `item_count` INT NOT NULL DEFAULT 0 COMMENT '商品数',
    `follower_count` INT NOT NULL DEFAULT 0 COMMENT '粉丝数',
    `response_rate` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '响应率%',
    `response_time` INT NOT NULL DEFAULT 0 COMMENT '响应时间(秒)',
    `cancellation_rate` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '取消率%',
    `total_sales` INT NOT NULL DEFAULT 0 COMMENT '总销量',
    `total_orders` INT NOT NULL DEFAULT 0 COMMENT '总订单数',
    `total_views` INT NOT NULL DEFAULT 0 COMMENT '总浏览量',
    `daily_sales` INT NOT NULL DEFAULT 0 COMMENT '日销量',
    `monthly_sales` INT NOT NULL DEFAULT 0 COMMENT '月销量',
    `yearly_sales` INT NOT NULL DEFAULT 0 COMMENT '年销量',
    `currency` VARCHAR(10) NOT NULL DEFAULT 'MYR' COMMENT '货币',
    `balance` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '余额',
    `pending_balance` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '待结算',
    `withdrawn_balance` DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '已提现',
    `contact_email` VARCHAR(200) DEFAULT NULL COMMENT '联系邮箱',
    `contact_phone` VARCHAR(50) DEFAULT NULL COMMENT '联系电话',
    `country` VARCHAR(100) DEFAULT NULL COMMENT '国家',
    `city` VARCHAR(100) DEFAULT NULL COMMENT '城市',
    `address` TEXT DEFAULT NULL COMMENT '地址',
    `zipcode` VARCHAR(20) DEFAULT NULL COMMENT '邮编',
    `auto_sync` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '自动同步',
    `sync_interval` INT NOT NULL DEFAULT 3600 COMMENT '同步间隔(秒)',
    `sync_items` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '同步商品',
    `sync_orders` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '同步订单',
    `sync_logistics` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '同步物流',
    `sync_finance` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '同步财务',
    `is_primary` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '主店铺',
    `last_sync_at` DATETIME DEFAULT NULL COMMENT '最后同步时间',
    `next_sync_at` DATETIME DEFAULT NULL COMMENT '下次同步时间',
    `shop_created_at` DATETIME DEFAULT NULL COMMENT '虾皮创建时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_id` (`shop_id`),
    KEY `idx_admin_id` (`admin_id`),
    KEY `idx_region` (`region`),
    KEY `idx_status` (`status`),
    KEY `idx_auth_status` (`auth_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺表';

-- ============================================================
-- 3. 店铺授权表 (shop_authorizations)
-- 存储 Shopee OAuth Token
-- access_token: ~4小时有效
-- refresh_token: 30-365天有效(商家选择)
-- ============================================================
CREATE TABLE IF NOT EXISTS `shop_authorizations` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '虾皮店铺ID',
    `access_token` VARCHAR(512) NOT NULL COMMENT '访问令牌',
    `refresh_token` VARCHAR(512) NOT NULL COMMENT '刷新令牌',
    `token_type` VARCHAR(50) NOT NULL DEFAULT 'Bearer' COMMENT '令牌类型',
    `expires_at` DATETIME NOT NULL COMMENT 'access_token过期时间',
    `refresh_expires_at` DATETIME NOT NULL COMMENT 'refresh_token过期时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '首次授权时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最近刷新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_id` (`shop_id`),
    KEY `idx_expires_at` (`expires_at`),
    KEY `idx_refresh_expires_at` (`refresh_expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺授权表';

-- ============================================================
-- 4. 订单表 (orders)
-- ============================================================
CREATE TABLE IF NOT EXISTS `orders` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    `order_sn` VARCHAR(64) NOT NULL COMMENT '订单号',
    `region` VARCHAR(10) NOT NULL COMMENT '地区',
    `order_status` VARCHAR(50) NOT NULL COMMENT '订单状态',
    `buyer_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '买家ID',
    `buyer_username` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '买家名',
    `total_amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '总金额',
    `currency` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '货币',
    `shipping_carrier` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '物流商',
    `tracking_number` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '物流单号',
    `ship_by_date` DATETIME DEFAULT NULL COMMENT '最晚发货时间',
    `pay_time` DATETIME DEFAULT NULL COMMENT '付款时间',
    `create_time` DATETIME DEFAULT NULL COMMENT '虾皮创建时间',
    `update_time` DATETIME DEFAULT NULL COMMENT '虾皮更新时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_order` (`shop_id`, `order_sn`),
    KEY `idx_order_sn` (`order_sn`),
    KEY `idx_order_status` (`order_status`),
    KEY `idx_ship_by_date` (`ship_by_date`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- ============================================================
-- 5. 订单商品表 (order_items)
-- ============================================================
CREATE TABLE IF NOT EXISTS `order_items` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    `order_sn` VARCHAR(64) NOT NULL COMMENT '订单号',
    `item_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    `item_name` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '商品名',
    `item_sku` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '商品SKU',
    `model_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '规格ID',
    `model_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '规格名',
    `model_sku` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '规格SKU',
    `quantity` INT NOT NULL DEFAULT 0 COMMENT '数量',
    `item_price` DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '单价',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_order_id` (`order_id`),
    KEY `idx_shop_order` (`shop_id`, `order_sn`),
    KEY `idx_item_id` (`item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单商品表';

-- ============================================================
-- 6. 收货地址表 (order_addresses)
-- ============================================================
CREATE TABLE IF NOT EXISTS `order_addresses` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    `order_sn` VARCHAR(64) NOT NULL COMMENT '订单号',
    `name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '收货人',
    `phone` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '电话',
    `town` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '乡镇',
    `district` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '区',
    `city` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '城市',
    `state` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '省/州',
    `region` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '国家',
    `zipcode` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '邮编',
    `full_address` TEXT COMMENT '完整地址',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_id` (`order_id`),
    KEY `idx_shop_order` (`shop_id`, `order_sn`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收货地址表';

-- ============================================================
-- 7. 发货记录表 (shipments)
-- ============================================================
CREATE TABLE IF NOT EXISTS `shipments` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    `order_sn` VARCHAR(64) NOT NULL COMMENT '订单号',
    `package_number` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '包裹号',
    `shipping_carrier` VARCHAR(100) NOT NULL COMMENT '物流商',
    `tracking_number` VARCHAR(100) NOT NULL COMMENT '物流单号',
    `ship_status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0待发货 1已发货 2失败',
    `ship_time` DATETIME DEFAULT NULL COMMENT '发货时间',
    `error_message` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '错误信息',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_order` (`shop_id`, `order_sn`),
    KEY `idx_tracking_number` (`tracking_number`),
    KEY `idx_ship_status` (`ship_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='发货记录表';

-- ============================================================
-- 8. 物流渠道表 (logistics_channels)
-- ============================================================
CREATE TABLE IF NOT EXISTS `logistics_channels` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `shop_id` BIGINT UNSIGNED NOT NULL COMMENT '店铺ID',
    `logistics_channel_id` BIGINT UNSIGNED NOT NULL COMMENT '物流渠道ID',
    `logistics_channel_name` VARCHAR(255) NOT NULL COMMENT '渠道名称',
    `cod_enabled` TINYINT NOT NULL DEFAULT 0 COMMENT '支持货到付款',
    `enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '启用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_shop_channel` (`shop_id`, `logistics_channel_id`),
    KEY `idx_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='物流渠道表';

-- ============================================================
-- 9. 操作日志表 (operation_logs)
-- ============================================================
CREATE TABLE IF NOT EXISTS `operation_logs` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `admin_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
    `shop_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '店铺ID',
    `order_sn` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '订单号',
    `operation_type` VARCHAR(50) NOT NULL COMMENT '操作类型',
    `operation_desc` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '操作描述',
    `request_data` TEXT COMMENT '请求数据',
    `response_data` TEXT COMMENT '响应数据',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1成功 0失败',
    `ip` VARCHAR(50) NOT NULL DEFAULT '' COMMENT 'IP',
    `user_agent` VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'UA',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_admin_id` (`admin_id`),
    KEY `idx_shop_id` (`shop_id`),
    KEY `idx_order_sn` (`order_sn`),
    KEY `idx_operation_type` (`operation_type`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';
