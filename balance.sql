-- ============================================================================
-- Balance 系统完整数据库建表SQL（全新安装版）
-- ============================================================================
-- 生成时间: 2026-02-16
-- 数据库: MySQL 8.0+
-- 字符集: utf8mb4_unicode_ci
-- 主键说明: 所有主键使用Redis生成的13位分布式ID，不使用自增（特殊标注的除外）
-- 分表说明: 订单相关按 shop_id % 10 | 账户流水按 admin_id % 10
-- 使用方式: 删除老库后执行 mysql -u user -p database < balance.sql
-- ============================================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================================
-- 第一部分：基础表（不分表）共 21 张
-- ============================================================================

-- ----------------------------
-- 1. 管理员/用户表
-- ----------------------------
DROP TABLE IF EXISTS `admin`;
CREATE TABLE `admin` (
  `id` bigint NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `user_no` varchar(32) DEFAULT NULL COMMENT '用户编号',
  `user_type` tinyint NOT NULL DEFAULT 1 COMMENT '用户类型: 1=店主 5=运营 9=平台管理员',
  `avatar` varchar(100) NOT NULL DEFAULT '' COMMENT '头像URL',
  `user_name` varchar(32) NOT NULL COMMENT '用户名(登录账号)',
  `real_name` varchar(64) DEFAULT NULL COMMENT '真实姓名',
  `salt` varchar(16) DEFAULT NULL COMMENT '密码盐值',
  `hash` varchar(64) DEFAULT NULL COMMENT '密码哈希值',
  `email` varchar(128) DEFAULT NULL COMMENT '邮箱地址',
  `phone` varchar(16) DEFAULT NULL COMMENT '手机号码',
  `line_id` varchar(64) DEFAULT NULL COMMENT 'LINE ID',
  `wechat` varchar(64) DEFAULT NULL COMMENT '微信号',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 2=禁用',
  `language` varchar(10) NOT NULL DEFAULT 'zh' COMMENT '语言偏好',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `login_ip` varchar(128) NOT NULL DEFAULT '' COMMENT '最后登录IP',
  `login_date` datetime DEFAULT NULL COMMENT '最后登录时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间(软删除)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_name` (`user_name`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员/用户表';

-- ----------------------------
-- 2. 店铺表
-- ----------------------------
DROP TABLE IF EXISTS `shops`;
CREATE TABLE `shops` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `shop_id` bigint unsigned NOT NULL COMMENT 'Shopee店铺ID',
  `shop_id_str` varchar(64) NOT NULL DEFAULT '' COMMENT 'Shopee店铺ID字符串',
  `admin_id` bigint NOT NULL DEFAULT 0 COMMENT '店铺老板ID(关联admin表)',
  `shop_name` varchar(255) NOT NULL DEFAULT '' COMMENT '店铺名称',
  `shop_slug` varchar(256) DEFAULT NULL COMMENT '店铺URL别名',
  `region` varchar(16) NOT NULL COMMENT '地区代码(SG/MY/TW等)',
  `partner_id` bigint NOT NULL DEFAULT 0 COMMENT 'Shopee合作伙伴ID',
  `auth_status` tinyint NOT NULL DEFAULT 0 COMMENT '授权状态: 0=未授权 1=已授权 2=已过期',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1=启用 2=禁用',
  `suspension_status` tinyint NOT NULL DEFAULT 0 COMMENT '暂停状态: 0=正常 1=暂停',
  `is_cb_shop` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否跨境店铺',
  `is_cod_shop` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否支持货到付款',
  `is_preferred_plus_shop` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否优选Plus店铺',
  `is_shopee_verified` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否Shopee认证',
  `rating_star` decimal(3,2) NOT NULL DEFAULT 0.00 COMMENT '店铺评分(星级)',
  `rating_bad` int NOT NULL DEFAULT 0 COMMENT '差评数',
  `rating_good` int NOT NULL DEFAULT 0 COMMENT '好评数',
  `rating_normal` int NOT NULL DEFAULT 0 COMMENT '中评数',
  `item_count` int NOT NULL DEFAULT 0 COMMENT '商品数量',
  `follower_count` int NOT NULL DEFAULT 0 COMMENT '粉丝数',
  `response_rate` decimal(5,2) NOT NULL DEFAULT 0.00 COMMENT '回复率(%)',
  `response_time` int NOT NULL DEFAULT 0 COMMENT '平均回复时间(秒)',
  `cancellation_rate` decimal(5,2) NOT NULL DEFAULT 0.00 COMMENT '取消率(%)',
  `total_sales` int NOT NULL DEFAULT 0 COMMENT '总销量',
  `total_orders` int NOT NULL DEFAULT 0 COMMENT '总订单数',
  `total_views` int NOT NULL DEFAULT 0 COMMENT '总浏览量',
  `daily_sales` int NOT NULL DEFAULT 0 COMMENT '日销量',
  `monthly_sales` int NOT NULL DEFAULT 0 COMMENT '月销量',
  `yearly_sales` int NOT NULL DEFAULT 0 COMMENT '年销量',
  `currency` varchar(10) NOT NULL DEFAULT 'MYR' COMMENT '货币代码',
  `balance` decimal(12,2) NOT NULL DEFAULT 0.00 COMMENT 'Shopee钱包余额',
  `pending_balance` decimal(12,2) NOT NULL DEFAULT 0.00 COMMENT '待结算余额',
  `withdrawn_balance` decimal(12,2) NOT NULL DEFAULT 0.00 COMMENT '已提现金额',
  `contact_email` varchar(200) DEFAULT NULL COMMENT '联系邮箱',
  `contact_phone` varchar(50) DEFAULT NULL COMMENT '联系电话',
  `country` varchar(100) DEFAULT NULL COMMENT '国家',
  `city` varchar(100) DEFAULT NULL COMMENT '城市',
  `address` text COMMENT '详细地址',
  `zipcode` varchar(20) DEFAULT NULL COMMENT '邮编',
  `auto_sync` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否自动同步',
  `sync_interval` int NOT NULL DEFAULT 3600 COMMENT '同步间隔(秒)',
  `sync_items` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否同步商品',
  `sync_orders` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否同步订单',
  `sync_logistics` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否同步物流',
  `sync_finance` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否同步财务',
  `is_primary` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否主店铺',
  `last_sync_at` datetime DEFAULT NULL COMMENT '上次同步/巡检时间',
  `next_sync_at` datetime DEFAULT NULL COMMENT '下次同步/巡检时间',
  `shop_created_at` datetime DEFAULT NULL COMMENT 'Shopee店铺创建时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_id` (`shop_id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_region` (`region`),
  KEY `idx_auth_status` (`auth_status`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺表';

-- ----------------------------
-- 3. 店铺授权表
-- ----------------------------
DROP TABLE IF EXISTS `shop_authorizations`;
CREATE TABLE `shop_authorizations` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `shop_id` bigint unsigned NOT NULL COMMENT 'Shopee店铺ID',
  `access_token` varchar(512) NOT NULL COMMENT '访问令牌',
  `refresh_token` varchar(512) NOT NULL COMMENT '刷新令牌',
  `token_type` varchar(50) NOT NULL DEFAULT 'Bearer' COMMENT '令牌类型',
  `expires_at` datetime NOT NULL COMMENT '访问令牌过期时间',
  `refresh_expires_at` datetime NOT NULL COMMENT '刷新令牌过期时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_id` (`shop_id`),
  KEY `idx_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺授权表';

-- ----------------------------
-- 4. 店铺-运营分配关系表
-- ----------------------------
DROP TABLE IF EXISTS `shop_operator_relations`;
CREATE TABLE `shop_operator_relations` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `shop_id` bigint unsigned NOT NULL COMMENT 'Shopee店铺ID',
  `shop_owner_id` bigint NOT NULL COMMENT '店铺老板ID(关联admin表)',
  `operator_id` bigint NOT NULL COMMENT '运营老板ID(关联admin表)',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 2=暂停 3=解除',
  `assigned_at` datetime NOT NULL COMMENT '分配时间',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_operator` (`shop_id`, `operator_id`),
  KEY `idx_shop_owner_id` (`shop_owner_id`),
  KEY `idx_operator_id` (`operator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺-运营分配关系表';

-- ----------------------------
-- 5. 店铺同步记录表（按类型拆分为独立表，便于多机部署时并发安全）
-- ----------------------------
DROP TABLE IF EXISTS `shop_sync_finance_income_records`;
CREATE TABLE `shop_sync_finance_income_records` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `shop_id` bigint unsigned NOT NULL COMMENT 'Shopee店铺ID',
  `last_sync_time` bigint NOT NULL DEFAULT 0 COMMENT '上次同步时间戳',
  `last_transaction_id` bigint NOT NULL DEFAULT 0 COMMENT '上次同步的交易ID',
  `last_sync_at` datetime DEFAULT NULL COMMENT '上次同步时间',
  `total_synced_count` bigint NOT NULL DEFAULT 0 COMMENT '累计同步数量',
  `last_sync_count` int NOT NULL DEFAULT 0 COMMENT '上次同步数量',
  `last_error` varchar(500) NOT NULL DEFAULT '' COMMENT '上次错误信息',
  `consecutive_fail_count` int NOT NULL DEFAULT 0 COMMENT '连续失败次数',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 0=禁用 1=启用 2=暂停',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺财务收入同步记录表';

DROP TABLE IF EXISTS `shop_sync_order_records`;
CREATE TABLE `shop_sync_order_records` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `shop_id` bigint unsigned NOT NULL COMMENT 'Shopee店铺ID',
  `last_sync_time` bigint NOT NULL DEFAULT 0 COMMENT '上次同步时间戳',
  `last_sync_at` datetime DEFAULT NULL COMMENT '上次同步时间',
  `total_synced_count` bigint NOT NULL DEFAULT 0 COMMENT '累计同步数量',
  `last_sync_count` int NOT NULL DEFAULT 0 COMMENT '上次同步数量',
  `last_error` varchar(500) NOT NULL DEFAULT '' COMMENT '上次错误信息',
  `consecutive_fail_count` int NOT NULL DEFAULT 0 COMMENT '连续失败次数',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 0=禁用 1=启用 2=暂停',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺订单同步记录表';

DROP TABLE IF EXISTS `shop_sync_escrow_records`;
CREATE TABLE `shop_sync_escrow_records` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `shop_id` bigint unsigned NOT NULL COMMENT 'Shopee店铺ID',
  `last_sync_time` bigint NOT NULL DEFAULT 0 COMMENT '上次同步时间戳',
  `last_sync_at` datetime DEFAULT NULL COMMENT '上次同步时间',
  `total_synced_count` bigint NOT NULL DEFAULT 0 COMMENT '累计同步数量',
  `last_sync_count` int NOT NULL DEFAULT 0 COMMENT '上次同步数量',
  `last_error` varchar(500) NOT NULL DEFAULT '' COMMENT '上次错误信息',
  `consecutive_fail_count` int NOT NULL DEFAULT 0 COMMENT '连续失败次数',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 0=禁用 1=启用 2=暂停',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺结算明细同步记录表';

-- ----------------------------
-- 6. 利润分成配置表
-- ----------------------------
DROP TABLE IF EXISTS `profit_share_configs`;
CREATE TABLE `profit_share_configs` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `shop_id` bigint unsigned NOT NULL COMMENT 'Shopee店铺ID',
  `operator_id` bigint NOT NULL COMMENT '运营老板ID',
  `platform_share_rate` decimal(5,2) NOT NULL DEFAULT 5.00 COMMENT '平台分成比例(%)',
  `operator_share_rate` decimal(5,2) NOT NULL DEFAULT 45.00 COMMENT '运营分成比例(%)',
  `shop_owner_share_rate` decimal(5,2) NOT NULL DEFAULT 50.00 COMMENT '店主分成比例(%)',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1=生效 2=失效',
  `effective_from` datetime NOT NULL COMMENT '生效开始时间',
  `effective_to` datetime DEFAULT NULL COMMENT '生效结束时间',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_operator_config` (`shop_id`, `operator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='利润分成配置表';

-- ----------------------------
-- 7. 物流渠道表
-- ----------------------------
DROP TABLE IF EXISTS `logistics_channels`;
CREATE TABLE `logistics_channels` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `shop_id` bigint unsigned NOT NULL COMMENT 'Shopee店铺ID',
  `logistics_channel_id` bigint unsigned NOT NULL COMMENT '物流渠道ID',
  `logistics_channel_name` varchar(255) NOT NULL COMMENT '物流渠道名称',
  `cod_enabled` tinyint NOT NULL DEFAULT 0 COMMENT '是否支持货到付款',
  `enabled` tinyint NOT NULL DEFAULT 1 COMMENT '是否启用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_channel` (`shop_id`, `logistics_channel_id`),
  KEY `idx_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='物流渠道表';

-- ----------------------------
-- 8. 预付款账户表（店主发货前预付成本）
-- ----------------------------
DROP TABLE IF EXISTS `prepayment_accounts`;
CREATE TABLE `prepayment_accounts` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `admin_id` bigint NOT NULL COMMENT '店铺老板ID(关联admin表)',
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
  `pending_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '待结算金额(订单入系统时已扣除，等待分账)',
  `total_recharge` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计充值金额',
  `total_consume` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计消费金额',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD' COMMENT '货币代码',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 2=暂停',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='预付款账户表';

-- ----------------------------
-- 9. 保证金账户表（店主缴纳的保证金）
-- ----------------------------
DROP TABLE IF EXISTS `deposit_accounts`;
CREATE TABLE `deposit_accounts` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `admin_id` bigint NOT NULL COMMENT '店铺老板ID(关联admin表)',
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '保证金余额',
  `required_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '应缴保证金金额',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD' COMMENT '货币代码',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 2=不足 3=暂停',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='保证金账户表';

-- ----------------------------
-- 10. 运营老板账户表（运营收到的成本+分成）
-- ----------------------------
DROP TABLE IF EXISTS `operator_accounts`;
CREATE TABLE `operator_accounts` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `admin_id` bigint NOT NULL COMMENT '运营老板ID(关联admin表)',
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
  `pending_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '待结算金额(提现申请时暂扣，打款后扣除)',
  `total_earnings` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计收益金额',
  `total_withdrawn` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计提现金额',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD' COMMENT '货币代码',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 2=暂停',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='运营老板账户表';

-- ----------------------------
-- 11. 店主佣金账户表（店主利润分成）
-- ----------------------------
DROP TABLE IF EXISTS `shop_owner_commission_accounts`;
CREATE TABLE `shop_owner_commission_accounts` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `admin_id` bigint NOT NULL COMMENT '店铺老板ID(关联admin表)',
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
  `pending_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '待结算金额(提现申请时暂扣，打款后扣除)',
  `total_earnings` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计收益金额',
  `total_withdrawn` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计提现金额',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD' COMMENT '货币代码',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 2=暂停',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店主佣金账户表';

-- ----------------------------
-- 12. 平台佣金账户表（平台利润分成，单例记录）
-- ----------------------------
DROP TABLE IF EXISTS `platform_commission_accounts`;
CREATE TABLE `platform_commission_accounts` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(单例,固定为1)',
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
  `pending_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '待结算金额(提现申请时暂扣)',
  `total_earnings` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计收益金额',
  `total_withdrawn` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计提现金额',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD' COMMENT '货币代码',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 2=暂停',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台佣金账户表';

-- ----------------------------
-- 13. 罚补账户表（运营罚款和补贴）
-- ----------------------------
DROP TABLE IF EXISTS `penalty_bonus_accounts`;
CREATE TABLE `penalty_bonus_accounts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID(自增)',
  `admin_id` bigint NOT NULL COMMENT '用户ID(关联admin表)',
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '余额(正=待付罚款,负=待发补贴)',
  `total_penalty` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计罚款金额',
  `total_bonus` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计补贴金额',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD' COMMENT '货币代码',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 2=暂停',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='罚补账户表';



-- ----------------------------
-- 14. 收款账户表（用户绑定的银行卡/电子钱包）
-- ----------------------------
DROP TABLE IF EXISTS `collection_accounts`;
CREATE TABLE `collection_accounts` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `admin_id` bigint NOT NULL COMMENT '用户ID(关联admin表)',
  `account_type` varchar(20) NOT NULL COMMENT '账户类型: wallet=电子钱包 bank=银行账户',
  `account_name` varchar(100) NOT NULL COMMENT '账户名称',
  `account_no` varchar(100) NOT NULL COMMENT '账户号码',
  `bank_name` varchar(100) NOT NULL DEFAULT '' COMMENT '银行名称',
  `bank_branch` varchar(200) NOT NULL DEFAULT '' COMMENT '银行支行',
  `payee` varchar(100) NOT NULL COMMENT '收款人姓名',
  `is_default` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否默认账户',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1=正常 2=未激活',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收款账户表';

-- ----------------------------
-- 15. 提现申请表
-- ----------------------------
DROP TABLE IF EXISTS `withdraw_applications`;
CREATE TABLE `withdraw_applications` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `application_no` varchar(64) NOT NULL COMMENT '申请单号(唯一)',
  `admin_id` bigint NOT NULL COMMENT '申请人ID(关联admin表)',
  `account_type` varchar(30) NOT NULL COMMENT '账户类型: operator/shop_owner_commission等',
  `amount` decimal(15,2) NOT NULL COMMENT '提现金额',
  `fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '手续费',
  `actual_amount` decimal(15,2) NOT NULL COMMENT '实际到账金额',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD' COMMENT '货币代码',
  `collection_account_id` bigint unsigned NOT NULL COMMENT '收款账户ID(关联collection_accounts)',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态: 0=待审核 1=已通过 2=已拒绝 3=已打款',
  `audit_remark` varchar(500) NOT NULL DEFAULT '' COMMENT '审核备注',
  `audit_by` bigint NOT NULL DEFAULT 0 COMMENT '审核人ID',
  `audit_at` datetime DEFAULT NULL COMMENT '审核时间',
  `paid_at` datetime DEFAULT NULL COMMENT '打款时间',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '申请备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_application_no` (`application_no`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_account_type` (`account_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='提现申请表';

-- ----------------------------
-- 16. 充值记录表（预付款/保证金充值，需审核后入账）
-- ----------------------------
DROP TABLE IF EXISTS `recharge_record`;
CREATE TABLE `recharge_record` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `application_no` varchar(64) NOT NULL COMMENT '申请单号(唯一)',
  `admin_id` bigint NOT NULL COMMENT '申请人ID(关联admin表)',
  `account_type` varchar(30) NOT NULL COMMENT '账户类型: prepayment=预付款 deposit=保证金',
  `amount` decimal(15,2) NOT NULL COMMENT '充值金额',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD' COMMENT '货币代码',
  `payment_method` varchar(30) NOT NULL COMMENT '支付方式: bank_transfer=银行转账 cash=现金',
  `payment_proof` varchar(500) NOT NULL DEFAULT '' COMMENT '支付凭证图片URL',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '状态: 0=待审核 1=已通过 2=已拒绝',
  `audit_remark` varchar(500) NOT NULL DEFAULT '' COMMENT '审核备注',
  `audit_by` bigint NOT NULL DEFAULT 0 COMMENT '审核人ID',
  `audit_at` datetime DEFAULT NULL COMMENT '审核时间',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '申请备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_application_no` (`application_no`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_account_type` (`account_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值记录表';

-- ----------------------------
-- 17. 站内消息通知表
-- ----------------------------
DROP TABLE IF EXISTS `notifications`;
CREATE TABLE `notifications` (
  `id` bigint unsigned NOT NULL COMMENT '主键ID(Redis分布式ID)',
  `admin_id` bigint NOT NULL COMMENT '接收人ID(店铺老板,关联admin表)',
  `shop_id` bigint unsigned NOT NULL COMMENT '关联店铺ID',
  `order_sn` varchar(64) NOT NULL DEFAULT '' COMMENT '关联订单号',
  `type` varchar(30) NOT NULL COMMENT '消息类型: prepayment_low=预付款不足',
  `title` varchar(200) NOT NULL COMMENT '消息标题',
  `content` text NOT NULL COMMENT '消息内容',
  `is_read` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已读: 0=未读 1=已读',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_shop_id` (`shop_id`),
  KEY `idx_order_sn` (`order_sn`),
  KEY `idx_type` (`type`),
  KEY `idx_is_read` (`is_read`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='站内消息通知表';

-- ----------------------------
-- 18. 订单每日统计表（按店铺汇总）
-- ----------------------------
DROP TABLE IF EXISTS `order_daily_stats`;
CREATE TABLE `order_daily_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID(自增)',
  `stat_date` date NOT NULL COMMENT '统计日期',
  `shop_id` bigint unsigned NOT NULL COMMENT '店铺ID',
  `order_count` bigint NOT NULL DEFAULT 0 COMMENT '订单数量',
  `total_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '订单总金额',
  `shipped_count` bigint NOT NULL DEFAULT 0 COMMENT '已发货数量',
  `settled_count` bigint NOT NULL DEFAULT 0 COMMENT '已结算数量',
  `settled_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '结算金额',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_stat_shop` (`stat_date`, `shop_id`),
  KEY `idx_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单每日统计表';

-- ----------------------------
-- 19. 财务每日统计表（按店铺汇总）
-- ----------------------------
DROP TABLE IF EXISTS `finance_daily_stats`;
CREATE TABLE `finance_daily_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID(自增)',
  `stat_date` date NOT NULL COMMENT '统计日期',
  `shop_id` bigint unsigned NOT NULL COMMENT '店铺ID',
  `income_count` bigint NOT NULL DEFAULT 0 COMMENT '收入笔数',
  `income_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '收入金额',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_stat_shop` (`stat_date`, `shop_id`),
  KEY `idx_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='财务每日统计表';

-- ----------------------------
-- 20. 平台每日统计表（全平台汇总）
-- ----------------------------
DROP TABLE IF EXISTS `platform_daily_stats`;
CREATE TABLE `platform_daily_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID(自增)',
  `stat_date` date NOT NULL COMMENT '统计日期',
  `total_orders` bigint NOT NULL DEFAULT 0 COMMENT '总订单数',
  `total_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '总订单金额',
  `settled_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '总结算金额',
  `platform_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '平台分成',
  `total_income` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '总收入',
  `active_shops` bigint NOT NULL DEFAULT 0 COMMENT '活跃店铺数',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_stat_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台每日统计表';


-- ============================================================================
-- 第二部分：分表（使用存储过程批量创建）
-- ============================================================================
-- 分表路由规则:
--   按 shop_id % 10 分表的表: orders, order_items, order_addresses, order_escrows,
--     order_escrow_items, order_settlements, order_shipment_records, shipments,
--     finance_incomes, operation_logs, operation_logs_archive, returns (共12种×10=120张)
--   按 admin_id % 10 分表的表: account_transactions (1种×10=10张)


-- ----------------------------
-- 1. 订单表分表 (orders_0 ~ orders_9)
-- 分表键: shop_id % 10
-- ----------------------------
DROP PROCEDURE IF EXISTS create_orders_shards;
DELIMITER //
CREATE PROCEDURE create_orders_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `orders_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `orders_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `shop_id` bigint unsigned NOT NULL COMMENT ''Shopee店铺ID'',
          `order_sn` varchar(64) NOT NULL COMMENT ''订单编号'',
          `region` varchar(10) NOT NULL COMMENT ''地区代码'',
          `order_status` varchar(50) NOT NULL COMMENT ''订单状态(UNPAID/READY_TO_SHIP/SHIPPED/COMPLETED等)'',
          `status_locked` tinyint(1) NOT NULL DEFAULT 0 COMMENT ''状态是否被锁定(锁定后不允许回退)'',
          `status_remark` varchar(255) NOT NULL DEFAULT '''' COMMENT ''状态备注'',
          `buyer_user_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''买家用户ID'',
          `buyer_username` varchar(255) NOT NULL DEFAULT '''' COMMENT ''买家用户名'',
          `total_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''订单总额'',
          `currency` varchar(10) NOT NULL DEFAULT '''' COMMENT ''货币代码'',
          `shipping_carrier` varchar(100) NOT NULL DEFAULT '''' COMMENT ''物流承运商'',
          `tracking_number` varchar(100) NOT NULL DEFAULT '''' COMMENT ''物流单号'',
          `ship_by_date` datetime DEFAULT NULL COMMENT ''最晚发货时间'',
          `pay_time` datetime DEFAULT NULL COMMENT ''支付时间'',
          `create_time` datetime DEFAULT NULL COMMENT ''Shopee订单创建时间'',
          `update_time` datetime DEFAULT NULL COMMENT ''Shopee订单更新时间'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''记录创建时间'',
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''记录更新时间'',
          `escrow_amount_snapshot` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''READY_TO_SHIP时获取的预估结算金额(get_escrow_detail)'',
          `buyer_paid_shipping_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''买家支付运费'',
          `original_cost_of_goods_sold` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''商品成本(COGS)'',
          `commission_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''平台佣金'',
          `seller_transaction_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''卖家交易手续费'',
          `credit_card_transaction_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''信用卡交易费'',
          `service_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''服务费'',
          `escrow_fee_x` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''预留费用X'',
          `escrow_fee_y` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''预留费用Y'',
          `escrow_fee_z` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''预留费用Z'',
          `prepayment_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''实际预付款扣除金额(=escrow_amount_snapshot)'',
          `prepayment_status` tinyint NOT NULL DEFAULT 0 COMMENT ''预付款状态: 0=未检查 1=充足 2=不足'',
          `prepayment_snapshot` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''检查时预付款总余额快照'',
          `prepayment_checked_at` datetime DEFAULT NULL COMMENT ''预付款检查时间'',
          PRIMARY KEY (`id`),
          UNIQUE KEY `uk_shop_order` (`shop_id`, `order_sn`),
          KEY `idx_order_sn` (`order_sn`),
          KEY `idx_order_status` (`order_status`),
          KEY `idx_ship_by_date` (`ship_by_date`),
          KEY `idx_create_time` (`create_time`),
          KEY `idx_prepayment_status` (`prepayment_status`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''订单表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_orders_shards();
DROP PROCEDURE IF EXISTS create_orders_shards;

-- ----------------------------
-- 2. 订单商品表分表 (order_items_0 ~ order_items_9)
-- 分表键: shop_id % 10
-- ----------------------------
DROP PROCEDURE IF EXISTS create_order_items_shards;
DELIMITER //
CREATE PROCEDURE create_order_items_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `order_items_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `order_items_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `order_id` bigint unsigned NOT NULL COMMENT ''订单ID'',
          `shop_id` bigint unsigned NOT NULL COMMENT ''Shopee店铺ID'',
          `order_sn` varchar(64) NOT NULL COMMENT ''订单编号'',
          `item_id` bigint unsigned NOT NULL COMMENT ''商品ID'',
          `item_name` varchar(512) NOT NULL DEFAULT '''' COMMENT ''商品名称'',
          `item_sku` varchar(100) NOT NULL DEFAULT '''' COMMENT ''商品SKU'',
          `model_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''规格ID'',
          `model_name` varchar(255) NOT NULL DEFAULT '''' COMMENT ''规格名称'',
          `model_sku` varchar(100) NOT NULL DEFAULT '''' COMMENT ''规格SKU'',
          `quantity` int NOT NULL DEFAULT 0 COMMENT ''数量'',
          `item_price` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''单价'',
          `order_status` varchar(50) NOT NULL DEFAULT '''' COMMENT ''子单状态: 空=正常 CANCELLED_BEFORE_SHIP=发货前取消'',
          `prepayment_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''该子单预付款金额。扣款时全为0；部分退款时在回调/拉取中填充用于计算返还'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''更新时间'',
          PRIMARY KEY (`id`),
          KEY `idx_order_id` (`order_id`),
          KEY `idx_shop_order` (`shop_id`, `order_sn`),
          KEY `idx_item_id` (`item_id`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''订单商品表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_order_items_shards();
DROP PROCEDURE IF EXISTS create_order_items_shards;

-- ----------------------------
-- 3. 订单地址表分表 (order_addresses_0 ~ order_addresses_9)
-- 分表键: shop_id % 10
-- ----------------------------
DROP PROCEDURE IF EXISTS create_order_addresses_shards;
DELIMITER //
CREATE PROCEDURE create_order_addresses_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `order_addresses_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `order_addresses_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `order_id` bigint unsigned NOT NULL COMMENT ''订单ID'',
          `shop_id` bigint unsigned NOT NULL COMMENT ''Shopee店铺ID'',
          `order_sn` varchar(64) NOT NULL COMMENT ''订单编号'',
          `name` varchar(255) NOT NULL DEFAULT '''' COMMENT ''收件人姓名'',
          `phone` varchar(50) NOT NULL DEFAULT '''' COMMENT ''收件人电话'',
          `town` varchar(255) NOT NULL DEFAULT '''' COMMENT ''乡镇'',
          `district` varchar(255) NOT NULL DEFAULT '''' COMMENT ''区县'',
          `city` varchar(255) NOT NULL DEFAULT '''' COMMENT ''城市'',
          `state` varchar(255) NOT NULL DEFAULT '''' COMMENT ''省/州'',
          `region` varchar(10) NOT NULL DEFAULT '''' COMMENT ''地区代码'',
          `zipcode` varchar(20) NOT NULL DEFAULT '''' COMMENT ''邮编'',
          `full_address` text COMMENT ''完整地址'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''更新时间'',
          PRIMARY KEY (`id`),
          UNIQUE KEY `uk_order_id` (`order_id`),
          KEY `idx_shop_order` (`shop_id`, `order_sn`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''订单地址表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_order_addresses_shards();
DROP PROCEDURE IF EXISTS create_order_addresses_shards;

-- ----------------------------
-- 4. 订单结算表分表 (order_escrows_0 ~ order_escrows_9)
-- 分表键: shop_id % 10
-- ----------------------------
DROP PROCEDURE IF EXISTS create_order_escrows_shards;
DELIMITER //
CREATE PROCEDURE create_order_escrows_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `order_escrows_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `order_escrows_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `shop_id` bigint unsigned NOT NULL COMMENT ''Shopee店铺ID'',
          `order_sn` varchar(64) NOT NULL COMMENT ''订单编号'',
          `order_id` bigint unsigned NOT NULL COMMENT ''订单ID'',
          `currency` varchar(10) NOT NULL DEFAULT '''' COMMENT ''货币代码'',
          `escrow_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''最终结算金额'',
          `buyer_total_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''买家支付总额'',
          `original_price` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''商品原价'',
          `seller_discount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''卖家折扣'',
          `shopee_discount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''平台折扣'',
          `voucher_from_seller` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''卖家优惠券'',
          `voucher_from_shopee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''平台优惠券'',
          `coins` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''虾皮币抵扣'',
          `buyer_paid_shipping_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''买家支付运费'',
          `final_shipping_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''最终运费'',
          `actual_shipping_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''实际运费'',
          `estimated_shipping_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''预估运费'',
          `shipping_fee_discount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''运费折扣'',
          `seller_shipping_discount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''卖家运费折扣'',
          `reverse_shipping_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''退货运费'',
          `commission_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''平台佣金'',
          `service_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''服务费'',
          `seller_transaction_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''卖家交易手续费'',
          `buyer_transaction_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''买家交易手续费'',
          `credit_card_transaction_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''信用卡手续费'',
          `escrow_tax` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''托管税费'',
          `cross_border_tax` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''跨境税费'',
          `payment_promotion` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''支付促销'',
          `credit_card_promotion` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''信用卡促销'',
          `seller_lost_compensation` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''卖家丢失补偿'',
          `seller_coin_cash_back` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''卖家虾皮币返现'',
          `seller_return_refund` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''卖家退货退款'',
          `final_product_protection` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''商品保护费'',
          `cost_of_goods_sold` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''商品成本'',
          `original_cost_of_goods_sold` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''原始商品成本'',
          `drc_adjustable_refund` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''DRC可调整退款'',
          `items_count` int NOT NULL DEFAULT 0 COMMENT ''商品数量'',
          `sync_status` tinyint NOT NULL DEFAULT 0 COMMENT ''同步状态: 0=未同步 1=已同步 2=失败'',
          `sync_time` datetime DEFAULT NULL COMMENT ''同步时间'',
          `sync_error` varchar(500) NOT NULL DEFAULT '''' COMMENT ''同步错误信息'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''更新时间'',
          PRIMARY KEY (`id`),
          UNIQUE KEY `uk_shop_order_escrow` (`shop_id`, `order_sn`),
          KEY `idx_order_id` (`order_id`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''订单结算表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_order_escrows_shards();
DROP PROCEDURE IF EXISTS create_order_escrows_shards;

-- ----------------------------
-- 5. 订单结算商品表分表 (order_escrow_items_0 ~ order_escrow_items_9)
-- 分表键: shop_id % 10
-- ----------------------------
DROP PROCEDURE IF EXISTS create_order_escrow_items_shards;
DELIMITER //
CREATE PROCEDURE create_order_escrow_items_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `order_escrow_items_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `order_escrow_items_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `escrow_id` bigint unsigned NOT NULL COMMENT ''结算记录ID'',
          `shop_id` bigint unsigned NOT NULL COMMENT ''Shopee店铺ID'',
          `order_sn` varchar(64) NOT NULL COMMENT ''订单编号'',
          `item_id` bigint unsigned NOT NULL COMMENT ''商品ID'',
          `item_name` varchar(512) NOT NULL DEFAULT '''' COMMENT ''商品名称'',
          `item_sku` varchar(100) NOT NULL DEFAULT '''' COMMENT ''商品SKU'',
          `model_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''规格ID'',
          `model_name` varchar(255) NOT NULL DEFAULT '''' COMMENT ''规格名称'',
          `model_sku` varchar(100) NOT NULL DEFAULT '''' COMMENT ''规格SKU'',
          `quantity_purchased` int NOT NULL DEFAULT 0 COMMENT ''购买数量'',
          `original_price` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''原价'',
          `discounted_price` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''折后价'',
          `seller_discount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''卖家折扣'',
          `shopee_discount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''平台折扣'',
          `discount_from_coin` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''虾皮币折扣'',
          `discount_from_voucher` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''优惠券折扣'',
          `discount_from_voucher_seller` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''卖家优惠券折扣'',
          `discount_from_voucher_shopee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''平台优惠券折扣'',
          `activity_type` varchar(50) NOT NULL DEFAULT '''' COMMENT ''活动类型'',
          `activity_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''活动ID'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''更新时间'',
          PRIMARY KEY (`id`),
          KEY `idx_escrow_id` (`escrow_id`),
          KEY `idx_shop_id` (`shop_id`),
          KEY `idx_order_sn` (`order_sn`),
          KEY `idx_item_id` (`item_id`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''订单结算商品表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_order_escrow_items_shards();
DROP PROCEDURE IF EXISTS create_order_escrow_items_shards;

-- ----------------------------
-- 6. 订单结算记录表分表 (order_settlements_0 ~ order_settlements_9)
-- 分表键: shop_id % 10
-- ----------------------------
DROP PROCEDURE IF EXISTS create_order_settlements_shards;
DELIMITER //
CREATE PROCEDURE create_order_settlements_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `order_settlements_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `order_settlements_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `settlement_no` varchar(64) NOT NULL COMMENT ''结算单号(唯一)'',
          `shop_id` bigint unsigned NOT NULL COMMENT ''Shopee店铺ID'',
          `order_sn` varchar(64) NOT NULL COMMENT ''订单编号'',
          `order_id` bigint unsigned NOT NULL COMMENT ''订单ID'',
          `shop_owner_id` bigint NOT NULL COMMENT ''店铺老板ID'',
          `operator_id` bigint NOT NULL COMMENT ''运营老板ID'',
          `currency` varchar(10) NOT NULL DEFAULT ''TWD'' COMMENT ''货币代码'',
          `escrow_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''Shopee结算金额'',
          `goods_cost` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''商品成本'',
          `shipping_cost` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''运费成本'',
          `total_cost` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''总成本'',
          `profit` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''利润'',
          `platform_share_rate` decimal(5,2) NOT NULL DEFAULT 0.00 COMMENT ''平台分成比例(%)'',
          `operator_share_rate` decimal(5,2) NOT NULL DEFAULT 0.00 COMMENT ''运营分成比例(%)'',
          `shop_owner_share_rate` decimal(5,2) NOT NULL DEFAULT 0.00 COMMENT ''店主分成比例(%)'',
          `platform_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''平台分成金额'',
          `operator_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''运营分成金额'',
          `shop_owner_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''店主分成金额'',
          `operator_income` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''运营实际收入'',
          `status` tinyint NOT NULL DEFAULT 0 COMMENT ''状态: 0=待结算 1=已结算 2=已取消'',
          `settled_at` datetime DEFAULT NULL COMMENT ''结算时间'',
          `remark` varchar(500) NOT NULL DEFAULT '''' COMMENT ''备注'',
          `adjustment_count` tinyint NOT NULL DEFAULT 0 COMMENT ''已发生调账次数(0~3)'',
          `adj1_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第1次调账金额(正补款/负扣款)'',
          `adj1_platform_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第1次调账-平台分成'',
          `adj1_operator_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第1次调账-运营分成'',
          `adj1_shop_owner_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第1次调账-店主分成'',
          `adj1_at` datetime DEFAULT NULL COMMENT ''第1次调账时间'',
          `adj1_remark` varchar(200) NOT NULL DEFAULT '''' COMMENT ''第1次调账备注'',
          `adj2_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第2次调账金额'',
          `adj2_platform_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第2次调账-平台分成'',
          `adj2_operator_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第2次调账-运营分成'',
          `adj2_shop_owner_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第2次调账-店主分成'',
          `adj2_at` datetime DEFAULT NULL COMMENT ''第2次调账时间'',
          `adj2_remark` varchar(200) NOT NULL DEFAULT '''' COMMENT ''第2次调账备注'',
          `adj3_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第3次调账金额'',
          `adj3_platform_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第3次调账-平台分成'',
          `adj3_operator_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第3次调账-运营分成'',
          `adj3_shop_owner_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''第3次调账-店主分成'',
          `adj3_at` datetime DEFAULT NULL COMMENT ''第3次调账时间'',
          `adj3_remark` varchar(200) NOT NULL DEFAULT '''' COMMENT ''第3次调账备注'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''更新时间'',
          PRIMARY KEY (`id`),
          UNIQUE KEY `uk_settlement_no` (`settlement_no`),
          UNIQUE KEY `uk_order_sn` (`order_sn`),
          KEY `idx_shop_id` (`shop_id`),
          KEY `idx_order_id` (`order_id`),
          KEY `idx_shop_owner_id` (`shop_owner_id`),
          KEY `idx_operator_id` (`operator_id`),
          KEY `idx_status` (`status`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''订单结算记录表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_order_settlements_shards();
DROP PROCEDURE IF EXISTS create_order_settlements_shards;

-- ----------------------------
-- 7. 订单发货记录表分表 (order_shipment_records_0 ~ order_shipment_records_9)
-- 分表键: shop_id % 10
-- ----------------------------
DROP PROCEDURE IF EXISTS create_order_shipment_records_shards;
DELIMITER //
CREATE PROCEDURE create_order_shipment_records_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `order_shipment_records_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `order_shipment_records_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `shop_id` bigint unsigned NOT NULL COMMENT ''Shopee店铺ID'',
          `order_sn` varchar(64) NOT NULL COMMENT ''订单编号'',
          `order_id` bigint unsigned NOT NULL COMMENT ''订单ID'',
          `shop_owner_id` bigint NOT NULL COMMENT ''店铺老板ID'',
          `operator_id` bigint NOT NULL COMMENT ''运营老板ID'',
          `goods_cost` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''商品成本'',
          `shipping_cost` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''运费成本'',
          `total_cost` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''总成本'',
          `currency` varchar(10) NOT NULL DEFAULT ''TWD'' COMMENT ''货币代码'',
          `prepayment_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''预付款金额(订单入系统时已扣除)'',
          `deduct_tx_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''扣款流水ID'',
          `shipping_carrier` varchar(100) NOT NULL DEFAULT '''' COMMENT ''物流承运商'',
          `tracking_number` varchar(100) NOT NULL DEFAULT '''' COMMENT ''物流单号'',
          `shipped_at` datetime DEFAULT NULL COMMENT ''发货时间'',
          `status` tinyint NOT NULL DEFAULT 0 COMMENT ''状态: 0=待发货 1=已发货 2=已完成 3=已取消 4=发货失败'',
          `settlement_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''关联结算记录ID'',
          `remark` varchar(500) NOT NULL DEFAULT '''' COMMENT ''备注'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''更新时间'',
          PRIMARY KEY (`id`),
          UNIQUE KEY `uk_order_sn` (`order_sn`),
          KEY `idx_shop_id` (`shop_id`),
          KEY `idx_order_id` (`order_id`),
          KEY `idx_shop_owner_id` (`shop_owner_id`),
          KEY `idx_operator_id` (`operator_id`),
          KEY `idx_status` (`status`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''订单发货记录表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_order_shipment_records_shards();
DROP PROCEDURE IF EXISTS create_order_shipment_records_shards;

-- ----------------------------
-- 8. 发货记录表分表 (shipments_0 ~ shipments_9)
-- 分表键: shop_id % 10
-- ----------------------------
DROP PROCEDURE IF EXISTS create_shipments_shards;
DELIMITER //
CREATE PROCEDURE create_shipments_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `shipments_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `shipments_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `shop_id` bigint unsigned NOT NULL COMMENT ''Shopee店铺ID'',
          `order_sn` varchar(64) NOT NULL COMMENT ''订单编号'',
          `package_number` varchar(64) NOT NULL DEFAULT '''' COMMENT ''包裹号'',
          `shipping_carrier` varchar(100) NOT NULL COMMENT ''物流承运商'',
          `tracking_number` varchar(100) NOT NULL COMMENT ''物流单号'',
          `ship_status` tinyint NOT NULL DEFAULT 0 COMMENT ''发货状态: 0=待发货 1=已发货 2=失败'',
          `ship_time` datetime DEFAULT NULL COMMENT ''发货时间'',
          `error_message` varchar(512) NOT NULL DEFAULT '''' COMMENT ''错误信息'',
          `remark` varchar(512) NOT NULL DEFAULT '''' COMMENT ''备注'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''更新时间'',
          PRIMARY KEY (`id`),
          UNIQUE KEY `uk_shop_order` (`shop_id`, `order_sn`),
          KEY `idx_tracking_number` (`tracking_number`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''发货记录表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_shipments_shards();
DROP PROCEDURE IF EXISTS create_shipments_shards;

-- ----------------------------
-- 9. 财务收入表分表 (finance_incomes_0 ~ finance_incomes_9)
-- 分表键: shop_id % 10
-- ----------------------------
DROP PROCEDURE IF EXISTS create_finance_incomes_shards;
DELIMITER //
CREATE PROCEDURE create_finance_incomes_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `finance_incomes_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `finance_incomes_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `shop_id` bigint unsigned NOT NULL COMMENT ''Shopee店铺ID'',
          `transaction_id` bigint NOT NULL COMMENT ''Shopee交易ID'',
          `order_sn` varchar(64) NOT NULL COMMENT ''关联订单号'',
          `refund_sn` varchar(64) NOT NULL DEFAULT '''' COMMENT ''退款单号'',
          `status` varchar(20) NOT NULL DEFAULT '''' COMMENT ''交易状态'',
          `wallet_type` varchar(20) NOT NULL DEFAULT '''' COMMENT ''钱包类型'',
          `transaction_type` varchar(50) NOT NULL COMMENT ''交易类型'',
          `amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''交易金额'',
          `current_balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''交易后余额'',
          `transaction_time` bigint NOT NULL COMMENT ''交易时间戳'',
          `transaction_fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''交易手续费'',
          `description` varchar(500) NOT NULL DEFAULT '''' COMMENT ''交易描述'',
          `buyer_name` varchar(100) NOT NULL DEFAULT '''' COMMENT ''买家名称'',
          `reason` varchar(255) NOT NULL DEFAULT '''' COMMENT ''交易原因'',
          `withdrawal_id` bigint NOT NULL DEFAULT 0 COMMENT ''提现ID'',
          `withdrawal_type` varchar(20) NOT NULL DEFAULT '''' COMMENT ''提现类型'',
          `transaction_tab_type` varchar(50) NOT NULL DEFAULT '''' COMMENT ''交易标签类型'',
          `money_flow` varchar(20) NOT NULL DEFAULT '''' COMMENT ''资金流向'',
          `settlement_handle_status` tinyint NOT NULL DEFAULT 0 COMMENT ''结算处理状态: 0=待结算 1=已结算'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''更新时间'',
          PRIMARY KEY (`id`),
          UNIQUE KEY `uk_transaction_id` (`transaction_id`),
          KEY `idx_shop_id` (`shop_id`),
          KEY `idx_order_sn` (`order_sn`),
          KEY `idx_transaction_type` (`transaction_type`),
          KEY `idx_transaction_time` (`transaction_time`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''财务收入表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_finance_incomes_shards();
DROP PROCEDURE IF EXISTS create_finance_incomes_shards;

-- ----------------------------
-- 10. 账户流水表分表 (account_transactions_0 ~ account_transactions_9)
-- 分表键: admin_id % 10（注意：与其他分表不同，按用户ID分表）
-- ----------------------------
DROP PROCEDURE IF EXISTS create_account_transactions_shards;
DELIMITER //
CREATE PROCEDURE create_account_transactions_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `account_transactions_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `account_transactions_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `transaction_no` varchar(64) NOT NULL COMMENT ''流水号(唯一)'',
          `account_type` varchar(20) NOT NULL COMMENT ''账户类型: prepayment/deposit/operator等'',
          `admin_id` bigint NOT NULL COMMENT ''账户所属用户ID'',
          `transaction_type` varchar(30) NOT NULL COMMENT ''交易类型: recharge/consume/freeze/order_refund等'',
          `amount` decimal(15,2) NOT NULL COMMENT ''金额(正=入账,负=出账)'',
          `balance_before` decimal(15,2) NOT NULL COMMENT ''交易前余额'',
          `balance_after` decimal(15,2) NOT NULL COMMENT ''交易后余额'',
          `related_order_sn` varchar(64) NOT NULL DEFAULT '''' COMMENT ''关联订单号'',
          `related_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''关联业务ID'',
          `remark` varchar(500) NOT NULL DEFAULT '''' COMMENT ''备注'',
          `operator_id` bigint NOT NULL DEFAULT 0 COMMENT ''操作人ID'',
          `status` tinyint NOT NULL DEFAULT 1 COMMENT ''状态: 0=待审批 1=已完成 2=已拒绝'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
          PRIMARY KEY (`id`),
          UNIQUE KEY `uk_transaction_no` (`transaction_no`),
          KEY `idx_account_type` (`account_type`),
          KEY `idx_admin_id` (`admin_id`),
          KEY `idx_transaction_type` (`transaction_type`),
          KEY `idx_related_order_sn` (`related_order_sn`),
          KEY `idx_status` (`status`),
          KEY `idx_created_at` (`created_at`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''账户流水表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_account_transactions_shards();
DROP PROCEDURE IF EXISTS create_account_transactions_shards;

-- ----------------------------
-- 11. 操作日志表分表 (operation_logs_0 ~ operation_logs_9)
-- 分表键: shop_id % 10
-- ----------------------------
DROP PROCEDURE IF EXISTS create_operation_logs_shards;
DELIMITER //
CREATE PROCEDURE create_operation_logs_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `operation_logs_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `operation_logs_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `shop_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''店铺ID'',
          `order_sn` varchar(64) NOT NULL DEFAULT '''' COMMENT ''订单号'',
          `operation_type` varchar(50) NOT NULL COMMENT ''操作类型'',
          `operation_desc` varchar(512) NOT NULL DEFAULT '''' COMMENT ''操作描述'',
          `request_data` text COMMENT ''请求数据(JSON)'',
          `response_data` text COMMENT ''响应数据(JSON)'',
          `status` tinyint NOT NULL DEFAULT 1 COMMENT ''状态: 1=成功 0=失败'',
          `ip` varchar(50) NOT NULL DEFAULT '''' COMMENT ''操作IP地址'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
          PRIMARY KEY (`id`),
          KEY `idx_shop_id` (`shop_id`),
          KEY `idx_order_sn` (`order_sn`),
          KEY `idx_operation_type` (`operation_type`),
          KEY `idx_created_at` (`created_at`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''操作日志表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_operation_logs_shards();
DROP PROCEDURE IF EXISTS create_operation_logs_shards;

-- ----------------------------
-- 12. 操作日志归档表分表 (operation_logs_archive_0 ~ operation_logs_archive_9)
-- 分表键: shop_id % 10
-- 用途: 90天前的操作日志自动归档到此表，365天后自动清理
-- ----------------------------
DROP PROCEDURE IF EXISTS create_operation_logs_archive_shards;
DELIMITER //
CREATE PROCEDURE create_operation_logs_archive_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `operation_logs_archive_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;

        SET @create_sql = CONCAT('CREATE TABLE `operation_logs_archive_', i, '` (
          `id` bigint NOT NULL COMMENT ''主键ID'',
          `shop_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''店铺ID'',
          `order_sn` varchar(64) NOT NULL DEFAULT '''' COMMENT ''订单号'',
          `operation_type` varchar(50) NOT NULL COMMENT ''操作类型'',
          `operation_desc` varchar(512) NOT NULL DEFAULT '''' COMMENT ''操作描述'',
          `request_data` text COMMENT ''请求数据(JSON)'',
          `response_data` text COMMENT ''响应数据(JSON)'',
          `status` tinyint NOT NULL DEFAULT 1 COMMENT ''状态: 1=成功 0=失败'',
          `ip` varchar(50) NOT NULL DEFAULT '''' COMMENT ''操作IP地址'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''创建时间'',
          PRIMARY KEY (`id`),
          KEY `idx_shop_id` (`shop_id`),
          KEY `idx_created_at` (`created_at`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''操作日志归档表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_operation_logs_archive_shards();
DROP PROCEDURE IF EXISTS create_operation_logs_archive_shards;


-- ----------------------------
-- 13. 退货退款表分表 (returns_0 ~ returns_9)
-- 分表键: shop_id % 10
-- 说明: 记录从 Shopee 同步的退货退款信息，退款确认后自动返还预付款
-- ----------------------------
DROP PROCEDURE IF EXISTS create_returns_shards;
DELIMITER //
CREATE PROCEDURE create_returns_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @drop_sql = CONCAT('DROP TABLE IF EXISTS `returns_', i, '`');
        PREPARE drop_stmt FROM @drop_sql;
        EXECUTE drop_stmt;
        DEALLOCATE PREPARE drop_stmt;
        
        SET @create_sql = CONCAT('CREATE TABLE `returns_', i, '` (
          `id` bigint unsigned NOT NULL COMMENT ''主键ID(Redis分布式ID)'',
          `shop_id` bigint unsigned NOT NULL COMMENT ''Shopee店铺ID'',
          `return_sn` varchar(64) NOT NULL COMMENT ''退货单号'',
          `order_sn` varchar(64) NOT NULL COMMENT ''关联订单号'',
          `reason` varchar(100) NOT NULL DEFAULT '''' COMMENT ''退货原因(Shopee枚举值)'',
          `text_reason` varchar(500) NOT NULL DEFAULT '''' COMMENT ''买家退货说明'',
          `refund_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''退款金额'',
          `amount_before_disc` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT ''折扣前金额'',
          `currency` varchar(10) NOT NULL DEFAULT '''' COMMENT ''币种'',
          `status` varchar(50) NOT NULL COMMENT ''退货状态(REQUESTED/ACCEPTED/CANCELLED/REFUND_PAID等)'',
          `needs_logistics` tinyint(1) NOT NULL DEFAULT 0 COMMENT ''是否需要退回商品'',
          `tracking_number` varchar(100) NOT NULL DEFAULT '''' COMMENT ''退货物流单号'',
          `logistics_status` varchar(50) NOT NULL DEFAULT '''' COMMENT ''退货物流状态'',
          `buyer_username` varchar(255) NOT NULL DEFAULT '''' COMMENT ''买家用户名'',
          `shopee_create_time` datetime DEFAULT NULL COMMENT ''Shopee退货创建时间'',
          `shopee_update_time` datetime DEFAULT NULL COMMENT ''Shopee退货更新时间'',
          `due_date` datetime DEFAULT NULL COMMENT ''卖家处理截止时间'',
          `refund_status` tinyint NOT NULL DEFAULT 0 COMMENT ''退款处理状态(0未处理/1已返还预付款/2跳过/3处理中/4处理失败)'',
          `refund_processed_at` datetime DEFAULT NULL COMMENT ''退款处理时间'',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT ''记录创建时间'',
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT ''记录更新时间'',
          PRIMARY KEY (`id`),
          UNIQUE KEY `uk_shop_return` (`shop_id`, `return_sn`),
          KEY `idx_order_sn` (`order_sn`),
          KEY `idx_status` (`status`),
          KEY `idx_refund_status` (`refund_status`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT=''退货退款表分表', i, '''');
        PREPARE create_stmt FROM @create_sql;
        EXECUTE create_stmt;
        DEALLOCATE PREPARE create_stmt;
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;
CALL create_returns_shards();
DROP PROCEDURE IF EXISTS create_returns_shards;


-- ============================================================================
-- 第三部分：初始化数据
-- ============================================================================

-- 平台佣金账户（单例记录，固定ID=1）
INSERT INTO `platform_commission_accounts` (`id`, `balance`, `pending_amount`, `total_earnings`, `total_withdrawn`, `currency`, `status`) 
VALUES (1, 0.00, 0.00, 0.00, 0.00, 'TWD', 1);

SET FOREIGN_KEY_CHECKS = 1;


-- ============================================================================
-- 附录：表清单与分表规则说明
-- ============================================================================
--
-- 一、基础表（不分表）共 20 张:
--   序号  表名                             说明
--   1     admin                            管理员/用户表
--   2     shops                            店铺表
--   3     shop_authorizations              店铺授权表
--   4     shop_operator_relations          店铺-运营分配关系表
--   5a    shop_sync_finance_income_records 店铺财务收入同步记录表
--   5b    shop_sync_order_records          店铺订单同步记录表
--   5c    shop_sync_escrow_records         店铺结算明细同步记录表
--   6     profit_share_configs             利润分成配置表
--   7     logistics_channels               物流渠道表
--   8     prepayment_accounts              预付款账户表
--   9     deposit_accounts                 保证金账户表
--   10    operator_accounts                运营老板账户表
--   11    shop_owner_commission_accounts   店主佣金账户表
--   12    platform_commission_accounts     平台佣金账户表(单例)
--   13    penalty_bonus_accounts           罚补账户表
--   14    collection_accounts              收款账户表
--   15    withdraw_applications            提现申请表
--   16    recharge_record                  充值记录表
--   17    notifications                    站内消息通知表
--   18    order_daily_stats                订单每日统计表
--   19    finance_daily_stats              财务每日统计表
--   20    platform_daily_stats             平台每日统计表
--
-- 二、分表（共 13 种基础表 × 10 个分片 = 130 张）:
--
--   按 shop_id % 10 分表（12种 × 10 = 120张）:
--   1     orders_0 ~ orders_9                             订单表
--   2     order_items_0 ~ order_items_9                   订单商品表
--   3     order_addresses_0 ~ order_addresses_9           订单地址表
--   4     order_escrows_0 ~ order_escrows_9               订单结算表
--   5     order_escrow_items_0 ~ order_escrow_items_9     订单结算商品表
--   6     order_settlements_0 ~ order_settlements_9       订单结算记录表
--   7     order_shipment_records_0 ~ order_shipment_records_9  订单发货记录表
--   8     shipments_0 ~ shipments_9                       发货记录表
--   9     finance_incomes_0 ~ finance_incomes_9           财务收入表
--   10    operation_logs_0 ~ operation_logs_9             操作日志表
--   11    operation_logs_archive_0 ~ operation_logs_archive_9  操作日志归档表
--   12    returns_0 ~ returns_9                           退货退款表
--
--   按 admin_id % 10 分表（1种 × 10 = 10张）:
--   13    account_transactions_0 ~ account_transactions_9  账户流水表
--
-- 三、总计物理表数量: 20 + 130 = 150 张
--
-- 四、分表路由示例:
--   shop_id  = 12345 → 12345 % 10 = 5 → orders_5, order_items_5, ...
--   admin_id = 67890 → 67890 % 10 = 0 → account_transactions_0
--
-- 五、定时维护任务:
--   每天凌晨 2:00  归档90天前的操作日志到 operation_logs_archive_X
--   每天凌晨 3:00  生成前一天的统计数据（order_daily_stats / finance_daily_stats / platform_daily_stats）
--   每月1号  4:00  清理365天前的归档数据
--
-- 六、注意事项:
--   1. 同一店铺的所有订单数据都在同一组分表中，保证关联查询高效
--   2. 跨店铺查询需要遍历所有分表（建议优先使用汇总统计表）
--   3. 应用层使用 database.ShardedDB 工具类进行分表路由
--   4. account_transactions 按 admin_id 分表，同一用户的流水在同一表中
--   5. 平台级统计查询优先使用汇总表，避免遍历分表
--   6. 充值（预付款/保证金）无需审核，直接入账；仅提现需要审核流程
