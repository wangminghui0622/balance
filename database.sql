-- Balance 系统完整数据库建表SQL
-- 生成时间: 2026-02-13
-- 数据库: MySQL 8.0+
-- 主键说明: 所有主键使用Redis生成的13位分布式ID，不使用自增
-- 分表说明: 订单相关表按 shop_id % 10 分成10个表

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================================
-- 第一部分：基础表（不分表）
-- ============================================================================

-- ----------------------------
-- 1. 管理员/用户表
-- ----------------------------
DROP TABLE IF EXISTS `admin`;
CREATE TABLE `admin` (
  `id` bigint NOT NULL,
  `user_no` varchar(32) DEFAULT NULL,
  `user_type` tinyint NOT NULL DEFAULT 1 COMMENT '1=店主 5=运营 9=平台',
  `avatar` varchar(100) NOT NULL DEFAULT '',
  `user_name` varchar(32) NOT NULL,
  `real_name` varchar(64) DEFAULT NULL,
  `salt` varchar(16) DEFAULT NULL,
  `hash` varchar(64) DEFAULT NULL,
  `email` varchar(128) DEFAULT NULL,
  `phone` varchar(16) DEFAULT NULL,
  `line_id` varchar(64) DEFAULT NULL,
  `wechat` varchar(64) DEFAULT NULL,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=正常 2=禁用',
  `language` varchar(10) NOT NULL DEFAULT 'zh',
  `remark` varchar(500) DEFAULT NULL,
  `login_ip` varchar(128) NOT NULL DEFAULT '',
  `login_date` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_name` (`user_name`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员/用户表';

-- ----------------------------
-- 2. 店铺表
-- ----------------------------
DROP TABLE IF EXISTS `shops`;
CREATE TABLE `shops` (
  `id` bigint unsigned NOT NULL,
  `shop_id` bigint unsigned NOT NULL,
  `shop_id_str` varchar(64) NOT NULL DEFAULT '',
  `admin_id` bigint NOT NULL DEFAULT 0 COMMENT '店铺老板ID',
  `shop_name` varchar(255) NOT NULL DEFAULT '',
  `shop_slug` varchar(256) DEFAULT NULL,
  `region` varchar(16) NOT NULL,
  `partner_id` bigint NOT NULL DEFAULT 0,
  `auth_status` tinyint NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1,
  `suspension_status` tinyint NOT NULL DEFAULT 0,
  `is_cb_shop` tinyint(1) NOT NULL DEFAULT 0,
  `is_cod_shop` tinyint(1) NOT NULL DEFAULT 0,
  `is_preferred_plus_shop` tinyint(1) NOT NULL DEFAULT 0,
  `is_shopee_verified` tinyint(1) NOT NULL DEFAULT 0,
  `rating_star` decimal(3,2) NOT NULL DEFAULT 0.00,
  `rating_bad` int NOT NULL DEFAULT 0,
  `rating_good` int NOT NULL DEFAULT 0,
  `rating_normal` int NOT NULL DEFAULT 0,
  `item_count` int NOT NULL DEFAULT 0,
  `follower_count` int NOT NULL DEFAULT 0,
  `response_rate` decimal(5,2) NOT NULL DEFAULT 0.00,
  `response_time` int NOT NULL DEFAULT 0,
  `cancellation_rate` decimal(5,2) NOT NULL DEFAULT 0.00,
  `total_sales` int NOT NULL DEFAULT 0,
  `total_orders` int NOT NULL DEFAULT 0,
  `total_views` int NOT NULL DEFAULT 0,
  `daily_sales` int NOT NULL DEFAULT 0,
  `monthly_sales` int NOT NULL DEFAULT 0,
  `yearly_sales` int NOT NULL DEFAULT 0,
  `currency` varchar(10) NOT NULL DEFAULT 'MYR',
  `balance` decimal(12,2) NOT NULL DEFAULT 0.00,
  `pending_balance` decimal(12,2) NOT NULL DEFAULT 0.00,
  `withdrawn_balance` decimal(12,2) NOT NULL DEFAULT 0.00,
  `contact_email` varchar(200) DEFAULT NULL,
  `contact_phone` varchar(50) DEFAULT NULL,
  `country` varchar(100) DEFAULT NULL,
  `city` varchar(100) DEFAULT NULL,
  `address` text,
  `zipcode` varchar(20) DEFAULT NULL,
  `auto_sync` tinyint(1) NOT NULL DEFAULT 1,
  `sync_interval` int NOT NULL DEFAULT 3600,
  `sync_items` tinyint(1) NOT NULL DEFAULT 1,
  `sync_orders` tinyint(1) NOT NULL DEFAULT 1,
  `sync_logistics` tinyint(1) NOT NULL DEFAULT 1,
  `sync_finance` tinyint(1) NOT NULL DEFAULT 1,
  `is_primary` tinyint(1) NOT NULL DEFAULT 0,
  `last_sync_at` datetime DEFAULT NULL,
  `next_sync_at` datetime DEFAULT NULL,
  `shop_created_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
  `id` bigint unsigned NOT NULL,
  `shop_id` bigint unsigned NOT NULL,
  `access_token` varchar(512) NOT NULL,
  `refresh_token` varchar(512) NOT NULL,
  `token_type` varchar(50) NOT NULL DEFAULT 'Bearer',
  `expires_at` datetime NOT NULL,
  `refresh_expires_at` datetime NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_id` (`shop_id`),
  KEY `idx_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺授权表';

-- ----------------------------
-- 4. 店铺-运营分配关系表
-- ----------------------------
DROP TABLE IF EXISTS `shop_operator_relations`;
CREATE TABLE `shop_operator_relations` (
  `id` bigint unsigned NOT NULL,
  `shop_id` bigint unsigned NOT NULL,
  `shop_owner_id` bigint NOT NULL COMMENT '店铺老板ID',
  `operator_id` bigint NOT NULL COMMENT '运营老板ID',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=正常 2=暂停 3=解除',
  `assigned_at` datetime NOT NULL,
  `remark` varchar(500) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_operator` (`shop_id`, `operator_id`),
  KEY `idx_shop_owner_id` (`shop_owner_id`),
  KEY `idx_operator_id` (`operator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺-运营分配关系表';

-- ----------------------------
-- 5. 店铺同步记录表
-- ----------------------------
DROP TABLE IF EXISTS `shop_sync_records`;
CREATE TABLE `shop_sync_records` (
  `id` bigint unsigned NOT NULL,
  `shop_id` bigint unsigned NOT NULL,
  `sync_type` varchar(50) NOT NULL COMMENT 'finance_income/order/escrow',
  `last_sync_time` bigint NOT NULL DEFAULT 0,
  `last_transaction_id` bigint NOT NULL DEFAULT 0,
  `last_sync_at` datetime DEFAULT NULL,
  `total_synced_count` bigint NOT NULL DEFAULT 0,
  `last_sync_count` int NOT NULL DEFAULT 0,
  `last_error` varchar(500) NOT NULL DEFAULT '',
  `consecutive_fail_count` int NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '0=禁用 1=启用 2=暂停',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_sync_type` (`shop_id`, `sync_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店铺同步记录表';

-- ----------------------------
-- 6. 利润分成配置表
-- ----------------------------
DROP TABLE IF EXISTS `profit_share_configs`;
CREATE TABLE `profit_share_configs` (
  `id` bigint unsigned NOT NULL,
  `shop_id` bigint unsigned NOT NULL,
  `operator_id` bigint NOT NULL,
  `platform_share_rate` decimal(5,2) NOT NULL DEFAULT 5.00 COMMENT '平台分成比例',
  `operator_share_rate` decimal(5,2) NOT NULL DEFAULT 45.00 COMMENT '运营分成比例',
  `shop_owner_share_rate` decimal(5,2) NOT NULL DEFAULT 50.00 COMMENT '店主分成比例',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=生效 2=失效',
  `effective_from` datetime NOT NULL,
  `effective_to` datetime DEFAULT NULL,
  `remark` varchar(500) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_operator_config` (`shop_id`, `operator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='利润分成配置表';

-- ----------------------------
-- 7. 物流渠道表
-- ----------------------------
DROP TABLE IF EXISTS `logistics_channels`;
CREATE TABLE `logistics_channels` (
  `id` bigint unsigned NOT NULL,
  `shop_id` bigint unsigned NOT NULL,
  `logistics_channel_id` bigint unsigned NOT NULL,
  `logistics_channel_name` varchar(255) NOT NULL,
  `cod_enabled` tinyint NOT NULL DEFAULT 0,
  `enabled` tinyint NOT NULL DEFAULT 1,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shop_channel` (`shop_id`, `logistics_channel_id`),
  KEY `idx_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='物流渠道表';

-- ----------------------------
-- 8. 财务收入表 (Shopee钱包交易)
-- ----------------------------
DROP TABLE IF EXISTS `finance_incomes`;
CREATE TABLE `finance_incomes` (
  `id` bigint unsigned NOT NULL,
  `shop_id` bigint unsigned NOT NULL,
  `transaction_id` bigint NOT NULL,
  `order_sn` varchar(64) NOT NULL,
  `refund_sn` varchar(64) NOT NULL DEFAULT '',
  `status` varchar(20) NOT NULL DEFAULT '',
  `wallet_type` varchar(20) NOT NULL DEFAULT '',
  `transaction_type` varchar(50) NOT NULL,
  `amount` decimal(15,2) NOT NULL DEFAULT 0.00,
  `current_balance` decimal(15,2) NOT NULL DEFAULT 0.00,
  `transaction_time` bigint NOT NULL,
  `transaction_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
  `description` varchar(500) NOT NULL DEFAULT '',
  `buyer_name` varchar(100) NOT NULL DEFAULT '',
  `reason` varchar(255) NOT NULL DEFAULT '',
  `withdrawal_id` bigint NOT NULL DEFAULT 0,
  `withdrawal_type` varchar(20) NOT NULL DEFAULT '',
  `transaction_tab_type` varchar(50) NOT NULL DEFAULT '',
  `money_flow` varchar(20) NOT NULL DEFAULT '',
  `settlement_handle_status` tinyint NOT NULL DEFAULT 0 COMMENT '0=待结算 1=已结算',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_transaction_id` (`transaction_id`),
  KEY `idx_shop_id` (`shop_id`),
  KEY `idx_order_sn` (`order_sn`),
  KEY `idx_transaction_type` (`transaction_type`),
  KEY `idx_transaction_time` (`transaction_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='财务收入表';

-- ----------------------------
-- 9. 预付款账户表
-- ----------------------------
DROP TABLE IF EXISTS `prepayment_accounts`;
CREATE TABLE `prepayment_accounts` (
  `id` bigint unsigned NOT NULL,
  `admin_id` bigint NOT NULL COMMENT '店铺老板ID',
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
  `frozen_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '冻结金额',
  `total_recharge` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计充值',
  `total_consume` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计消费',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=正常 2=冻结',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='预付款账户表';

-- ----------------------------
-- 10. 保证金账户表
-- ----------------------------
DROP TABLE IF EXISTS `deposit_accounts`;
CREATE TABLE `deposit_accounts` (
  `id` bigint unsigned NOT NULL,
  `admin_id` bigint NOT NULL COMMENT '店铺老板ID',
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '保证金余额',
  `required_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '应缴保证金',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=正常 2=不足 3=冻结',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='保证金账户表';

-- ----------------------------
-- 11. 运营老板账户表
-- ----------------------------
DROP TABLE IF EXISTS `operator_accounts`;
CREATE TABLE `operator_accounts` (
  `id` bigint unsigned NOT NULL,
  `admin_id` bigint NOT NULL COMMENT '运营老板ID',
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
  `frozen_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '冻结金额',
  `total_earnings` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计收益',
  `total_withdrawn` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计提现',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=正常 2=冻结',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='运营老板账户表';

-- ----------------------------
-- 12. 店主佣金账户表
-- ----------------------------
DROP TABLE IF EXISTS `shop_owner_commission_accounts`;
CREATE TABLE `shop_owner_commission_accounts` (
  `id` bigint unsigned NOT NULL,
  `admin_id` bigint NOT NULL COMMENT '店铺老板ID',
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
  `frozen_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '冻结金额',
  `total_earnings` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计收益',
  `total_withdrawn` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计提现',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=正常 2=冻结',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='店主佣金账户表';

-- ----------------------------
-- 13. 平台佣金账户表
-- ----------------------------
DROP TABLE IF EXISTS `platform_commission_accounts`;
CREATE TABLE `platform_commission_accounts` (
  `id` bigint unsigned NOT NULL,
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
  `frozen_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '冻结金额',
  `total_earnings` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计收益',
  `total_withdrawn` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计提现',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=正常 2=冻结',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台佣金账户表';

-- ----------------------------
-- 14. 罚补账户表
-- ----------------------------
DROP TABLE IF EXISTS `penalty_bonus_accounts`;
CREATE TABLE `penalty_bonus_accounts` (
  `id` bigint unsigned NOT NULL,
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '余额',
  `total_penalty` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计罚款',
  `total_bonus` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计补贴',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=正常 2=冻结',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='罚补账户表';

-- ----------------------------
-- 15. 托管账户表
-- ----------------------------
DROP TABLE IF EXISTS `escrow_accounts`;
CREATE TABLE `escrow_accounts` (
  `id` bigint unsigned NOT NULL,
  `balance` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '托管余额',
  `total_in` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计转入',
  `total_out` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '累计转出',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=正常 2=冻结',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='托管账户表';

-- ----------------------------
-- 16. 账户流水表
-- ----------------------------
DROP TABLE IF EXISTS `account_transactions`;
CREATE TABLE `account_transactions` (
  `id` bigint unsigned NOT NULL,
  `transaction_no` varchar(64) NOT NULL COMMENT '流水号',
  `account_type` varchar(20) NOT NULL COMMENT 'prepayment/deposit/operator等',
  `admin_id` bigint NOT NULL COMMENT '账户所属用户ID',
  `transaction_type` varchar(30) NOT NULL COMMENT '交易类型',
  `amount` decimal(15,2) NOT NULL COMMENT '金额',
  `balance_before` decimal(15,2) NOT NULL COMMENT '交易前余额',
  `balance_after` decimal(15,2) NOT NULL COMMENT '交易后余额',
  `related_order_sn` varchar(64) NOT NULL DEFAULT '' COMMENT '关联订单号',
  `related_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '关联ID',
  `remark` varchar(500) NOT NULL DEFAULT '',
  `operator_id` bigint NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '0=待审批 1=已完成 2=已拒绝',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_transaction_no` (`transaction_no`),
  KEY `idx_account_type` (`account_type`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_transaction_type` (`transaction_type`),
  KEY `idx_related_order_sn` (`related_order_sn`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='账户流水表';

-- ----------------------------
-- 17. 收款账户表
-- ----------------------------
DROP TABLE IF EXISTS `collection_accounts`;
CREATE TABLE `collection_accounts` (
  `id` bigint unsigned NOT NULL,
  `admin_id` bigint NOT NULL,
  `account_type` varchar(20) NOT NULL COMMENT 'wallet/bank',
  `account_name` varchar(100) NOT NULL,
  `account_no` varchar(100) NOT NULL,
  `bank_name` varchar(100) NOT NULL DEFAULT '',
  `bank_branch` varchar(200) NOT NULL DEFAULT '',
  `payee` varchar(100) NOT NULL,
  `is_default` tinyint(1) NOT NULL DEFAULT 0,
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=正常 2=未激活',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收款账户表';

-- ----------------------------
-- 18. 提现申请表
-- ----------------------------
DROP TABLE IF EXISTS `withdraw_applications`;
CREATE TABLE `withdraw_applications` (
  `id` bigint unsigned NOT NULL,
  `application_no` varchar(64) NOT NULL COMMENT '申请单号',
  `admin_id` bigint NOT NULL COMMENT '申请人ID',
  `account_type` varchar(30) NOT NULL COMMENT '账户类型',
  `amount` decimal(15,2) NOT NULL COMMENT '提现金额',
  `fee` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '手续费',
  `actual_amount` decimal(15,2) NOT NULL COMMENT '实际到账金额',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD',
  `collection_account_id` bigint unsigned NOT NULL COMMENT '收款账户ID',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=待审核 1=已通过 2=已拒绝 3=已打款',
  `audit_remark` varchar(500) NOT NULL DEFAULT '' COMMENT '审核备注',
  `audit_by` bigint NOT NULL DEFAULT 0 COMMENT '审核人ID',
  `audit_at` datetime DEFAULT NULL COMMENT '审核时间',
  `paid_at` datetime DEFAULT NULL COMMENT '打款时间',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '申请备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_application_no` (`application_no`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_account_type` (`account_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='提现申请表';

-- ----------------------------
-- 19. 充值申请表
-- ----------------------------
DROP TABLE IF EXISTS `recharge_applications`;
CREATE TABLE `recharge_applications` (
  `id` bigint unsigned NOT NULL,
  `application_no` varchar(64) NOT NULL COMMENT '申请单号',
  `admin_id` bigint NOT NULL COMMENT '申请人ID',
  `account_type` varchar(30) NOT NULL COMMENT '账户类型: prepayment/deposit',
  `amount` decimal(15,2) NOT NULL COMMENT '充值金额',
  `currency` varchar(10) NOT NULL DEFAULT 'TWD',
  `payment_method` varchar(30) NOT NULL COMMENT '支付方式: bank_transfer/cash',
  `payment_proof` varchar(500) NOT NULL DEFAULT '' COMMENT '支付凭证',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=待审核 1=已通过 2=已拒绝',
  `audit_remark` varchar(500) NOT NULL DEFAULT '' COMMENT '审核备注',
  `audit_by` bigint NOT NULL DEFAULT 0 COMMENT '审核人ID',
  `audit_at` datetime DEFAULT NULL COMMENT '审核时间',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '申请备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_application_no` (`application_no`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_account_type` (`account_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='充值申请表';

-- ----------------------------
-- 20. 操作日志表
-- ----------------------------
DROP TABLE IF EXISTS `operation_logs`;
CREATE TABLE `operation_logs` (
  `id` bigint unsigned NOT NULL,
  `shop_id` bigint unsigned NOT NULL DEFAULT 0,
  `order_sn` varchar(64) NOT NULL DEFAULT '',
  `operation_type` varchar(50) NOT NULL,
  `operation_desc` varchar(512) NOT NULL DEFAULT '',
  `request_data` text,
  `response_data` text,
  `status` tinyint NOT NULL DEFAULT 1,
  `ip` varchar(50) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_shop_id` (`shop_id`),
  KEY `idx_order_sn` (`order_sn`),
  KEY `idx_operation_type` (`operation_type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';

-- ============================================================================
-- 第二部分：分表（按 shop_id % 10 分成10个表）
-- 分表规则: table_name = base_name + "_" + (shop_id % 10)
-- ============================================================================

-- ----------------------------
-- 订单表分表 (orders_0 ~ orders_9)
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
          `id` bigint unsigned NOT NULL,
          `shop_id` bigint unsigned NOT NULL,
          `order_sn` varchar(64) NOT NULL,
          `region` varchar(10) NOT NULL,
          `order_status` varchar(50) NOT NULL,
          `status_locked` tinyint(1) NOT NULL DEFAULT 0,
          `status_remark` varchar(255) NOT NULL DEFAULT '''',
          `buyer_user_id` bigint unsigned NOT NULL DEFAULT 0,
          `buyer_username` varchar(255) NOT NULL DEFAULT '''',
          `total_amount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `currency` varchar(10) NOT NULL DEFAULT '''',
          `shipping_carrier` varchar(100) NOT NULL DEFAULT '''',
          `tracking_number` varchar(100) NOT NULL DEFAULT '''',
          `ship_by_date` datetime DEFAULT NULL,
          `pay_time` datetime DEFAULT NULL,
          `create_time` datetime DEFAULT NULL,
          `update_time` datetime DEFAULT NULL,
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
          PRIMARY KEY (`id`),
          UNIQUE KEY `uk_shop_order` (`shop_id`, `order_sn`),
          KEY `idx_order_sn` (`order_sn`),
          KEY `idx_order_status` (`order_status`),
          KEY `idx_ship_by_date` (`ship_by_date`),
          KEY `idx_create_time` (`create_time`)
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
-- 订单商品表分表 (order_items_0 ~ order_items_9)
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
          `id` bigint unsigned NOT NULL,
          `order_id` bigint unsigned NOT NULL,
          `shop_id` bigint unsigned NOT NULL,
          `order_sn` varchar(64) NOT NULL,
          `item_id` bigint unsigned NOT NULL,
          `item_name` varchar(512) NOT NULL DEFAULT '''',
          `item_sku` varchar(100) NOT NULL DEFAULT '''',
          `model_id` bigint unsigned NOT NULL DEFAULT 0,
          `model_name` varchar(255) NOT NULL DEFAULT '''',
          `model_sku` varchar(100) NOT NULL DEFAULT '''',
          `quantity` int NOT NULL DEFAULT 0,
          `item_price` decimal(15,2) NOT NULL DEFAULT 0.00,
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
-- 订单地址表分表 (order_addresses_0 ~ order_addresses_9)
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
          `id` bigint unsigned NOT NULL,
          `order_id` bigint unsigned NOT NULL,
          `shop_id` bigint unsigned NOT NULL,
          `order_sn` varchar(64) NOT NULL,
          `name` varchar(255) NOT NULL DEFAULT '''',
          `phone` varchar(50) NOT NULL DEFAULT '''',
          `town` varchar(255) NOT NULL DEFAULT '''',
          `district` varchar(255) NOT NULL DEFAULT '''',
          `city` varchar(255) NOT NULL DEFAULT '''',
          `state` varchar(255) NOT NULL DEFAULT '''',
          `region` varchar(10) NOT NULL DEFAULT '''',
          `zipcode` varchar(20) NOT NULL DEFAULT '''',
          `full_address` text,
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
-- 订单结算表分表 (order_escrows_0 ~ order_escrows_9)
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
          `id` bigint unsigned NOT NULL,
          `shop_id` bigint unsigned NOT NULL,
          `order_sn` varchar(64) NOT NULL,
          `order_id` bigint unsigned NOT NULL,
          `currency` varchar(10) NOT NULL DEFAULT '''',
          `escrow_amount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `buyer_total_amount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `original_price` decimal(15,2) NOT NULL DEFAULT 0.00,
          `seller_discount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `shopee_discount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `voucher_from_seller` decimal(15,2) NOT NULL DEFAULT 0.00,
          `voucher_from_shopee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `coins` decimal(15,2) NOT NULL DEFAULT 0.00,
          `buyer_paid_shipping_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `final_shipping_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `actual_shipping_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `estimated_shipping_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `shipping_fee_discount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `seller_shipping_discount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `reverse_shipping_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `commission_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `service_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `seller_transaction_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `buyer_transaction_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `credit_card_transaction_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `escrow_tax` decimal(15,2) NOT NULL DEFAULT 0.00,
          `cross_border_tax` decimal(15,2) NOT NULL DEFAULT 0.00,
          `payment_promotion` decimal(15,2) NOT NULL DEFAULT 0.00,
          `credit_card_promotion` decimal(15,2) NOT NULL DEFAULT 0.00,
          `seller_lost_compensation` decimal(15,2) NOT NULL DEFAULT 0.00,
          `seller_coin_cash_back` decimal(15,2) NOT NULL DEFAULT 0.00,
          `seller_return_refund` decimal(15,2) NOT NULL DEFAULT 0.00,
          `final_product_protection` decimal(15,2) NOT NULL DEFAULT 0.00,
          `cost_of_goods_sold` decimal(15,2) NOT NULL DEFAULT 0.00,
          `original_cost_of_goods_sold` decimal(15,2) NOT NULL DEFAULT 0.00,
          `drc_adjustable_refund` decimal(15,2) NOT NULL DEFAULT 0.00,
          `items_count` int NOT NULL DEFAULT 0,
          `sync_status` tinyint NOT NULL DEFAULT 0,
          `sync_time` datetime DEFAULT NULL,
          `sync_error` varchar(500) NOT NULL DEFAULT '''',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
-- 订单结算商品表分表 (order_escrow_items_0 ~ order_escrow_items_9)
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
          `id` bigint unsigned NOT NULL,
          `escrow_id` bigint unsigned NOT NULL,
          `shop_id` bigint unsigned NOT NULL,
          `order_sn` varchar(64) NOT NULL,
          `item_id` bigint unsigned NOT NULL,
          `item_name` varchar(512) NOT NULL DEFAULT '''',
          `item_sku` varchar(100) NOT NULL DEFAULT '''',
          `model_id` bigint unsigned NOT NULL DEFAULT 0,
          `model_name` varchar(255) NOT NULL DEFAULT '''',
          `model_sku` varchar(100) NOT NULL DEFAULT '''',
          `quantity_purchased` int NOT NULL DEFAULT 0,
          `original_price` decimal(15,2) NOT NULL DEFAULT 0.00,
          `discounted_price` decimal(15,2) NOT NULL DEFAULT 0.00,
          `seller_discount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `shopee_discount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `discount_from_coin` decimal(15,2) NOT NULL DEFAULT 0.00,
          `discount_from_voucher` decimal(15,2) NOT NULL DEFAULT 0.00,
          `discount_from_voucher_seller` decimal(15,2) NOT NULL DEFAULT 0.00,
          `discount_from_voucher_shopee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `activity_type` varchar(50) NOT NULL DEFAULT '''',
          `activity_id` bigint unsigned NOT NULL DEFAULT 0,
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
-- 订单结算记录表分表 (order_settlements_0 ~ order_settlements_9)
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
          `id` bigint unsigned NOT NULL,
          `settlement_no` varchar(64) NOT NULL,
          `shop_id` bigint unsigned NOT NULL,
          `order_sn` varchar(64) NOT NULL,
          `order_id` bigint unsigned NOT NULL,
          `shop_owner_id` bigint NOT NULL,
          `operator_id` bigint NOT NULL,
          `currency` varchar(10) NOT NULL DEFAULT ''TWD'',
          `escrow_amount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `goods_cost` decimal(15,2) NOT NULL DEFAULT 0.00,
          `shipping_cost` decimal(15,2) NOT NULL DEFAULT 0.00,
          `total_cost` decimal(15,2) NOT NULL DEFAULT 0.00,
          `profit` decimal(15,2) NOT NULL DEFAULT 0.00,
          `platform_share_rate` decimal(5,2) NOT NULL DEFAULT 0.00,
          `operator_share_rate` decimal(5,2) NOT NULL DEFAULT 0.00,
          `shop_owner_share_rate` decimal(5,2) NOT NULL DEFAULT 0.00,
          `platform_share` decimal(15,2) NOT NULL DEFAULT 0.00,
          `operator_share` decimal(15,2) NOT NULL DEFAULT 0.00,
          `shop_owner_share` decimal(15,2) NOT NULL DEFAULT 0.00,
          `operator_income` decimal(15,2) NOT NULL DEFAULT 0.00,
          `status` tinyint NOT NULL DEFAULT 0,
          `settled_at` datetime DEFAULT NULL,
          `remark` varchar(500) NOT NULL DEFAULT '''',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
-- 订单发货记录表分表 (order_shipment_records_0 ~ order_shipment_records_9)
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
          `id` bigint unsigned NOT NULL,
          `shop_id` bigint unsigned NOT NULL,
          `order_sn` varchar(64) NOT NULL,
          `order_id` bigint unsigned NOT NULL,
          `shop_owner_id` bigint NOT NULL,
          `operator_id` bigint NOT NULL,
          `goods_cost` decimal(15,2) NOT NULL DEFAULT 0.00,
          `shipping_cost` decimal(15,2) NOT NULL DEFAULT 0.00,
          `total_cost` decimal(15,2) NOT NULL DEFAULT 0.00,
          `currency` varchar(10) NOT NULL DEFAULT ''TWD'',
          `frozen_amount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `frozen_transaction_id` bigint unsigned NOT NULL DEFAULT 0,
          `shipping_carrier` varchar(100) NOT NULL DEFAULT '''',
          `tracking_number` varchar(100) NOT NULL DEFAULT '''',
          `shipped_at` datetime DEFAULT NULL,
          `status` tinyint NOT NULL DEFAULT 0,
          `settlement_id` bigint unsigned NOT NULL DEFAULT 0,
          `remark` varchar(500) NOT NULL DEFAULT '''',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
-- 发货记录表分表 (shipments_0 ~ shipments_9)
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
          `id` bigint unsigned NOT NULL,
          `shop_id` bigint unsigned NOT NULL,
          `order_sn` varchar(64) NOT NULL,
          `package_number` varchar(64) NOT NULL DEFAULT '''',
          `shipping_carrier` varchar(100) NOT NULL,
          `tracking_number` varchar(100) NOT NULL,
          `ship_status` tinyint NOT NULL DEFAULT 0,
          `ship_time` datetime DEFAULT NULL,
          `error_message` varchar(512) NOT NULL DEFAULT '''',
          `remark` varchar(512) NOT NULL DEFAULT '''',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
-- 财务收入表分表 (finance_incomes_0 ~ finance_incomes_9)
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
          `id` bigint unsigned NOT NULL,
          `shop_id` bigint unsigned NOT NULL,
          `transaction_id` bigint NOT NULL,
          `order_sn` varchar(64) NOT NULL,
          `refund_sn` varchar(64) NOT NULL DEFAULT '''',
          `status` varchar(20) NOT NULL DEFAULT '''',
          `wallet_type` varchar(20) NOT NULL DEFAULT '''',
          `transaction_type` varchar(50) NOT NULL,
          `amount` decimal(15,2) NOT NULL DEFAULT 0.00,
          `current_balance` decimal(15,2) NOT NULL DEFAULT 0.00,
          `transaction_time` bigint NOT NULL,
          `transaction_fee` decimal(15,2) NOT NULL DEFAULT 0.00,
          `description` varchar(500) NOT NULL DEFAULT '''',
          `buyer_name` varchar(100) NOT NULL DEFAULT '''',
          `reason` varchar(255) NOT NULL DEFAULT '''',
          `withdrawal_id` bigint NOT NULL DEFAULT 0,
          `withdrawal_type` varchar(20) NOT NULL DEFAULT '''',
          `transaction_tab_type` varchar(50) NOT NULL DEFAULT '''',
          `money_flow` varchar(20) NOT NULL DEFAULT '''',
          `settlement_handle_status` tinyint NOT NULL DEFAULT 0,
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
-- 账户流水表分表 (account_transactions_0 ~ account_transactions_9)
-- 注意：按 admin_id % 10 分表
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
          `id` bigint unsigned NOT NULL,
          `transaction_no` varchar(64) NOT NULL,
          `account_type` varchar(20) NOT NULL,
          `admin_id` bigint NOT NULL,
          `transaction_type` varchar(30) NOT NULL,
          `amount` decimal(15,2) NOT NULL,
          `balance_before` decimal(15,2) NOT NULL,
          `balance_after` decimal(15,2) NOT NULL,
          `related_order_sn` varchar(64) NOT NULL DEFAULT '''',
          `related_id` bigint unsigned NOT NULL DEFAULT 0,
          `remark` varchar(500) NOT NULL DEFAULT '''',
          `operator_id` bigint NOT NULL DEFAULT 0,
          `status` tinyint NOT NULL DEFAULT 1,
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
-- 操作日志表分表 (operation_logs_0 ~ operation_logs_9)
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
          `id` bigint unsigned NOT NULL,
          `shop_id` bigint unsigned NOT NULL DEFAULT 0,
          `order_sn` varchar(64) NOT NULL DEFAULT '''',
          `operation_type` varchar(50) NOT NULL,
          `operation_desc` varchar(512) NOT NULL DEFAULT '''',
          `request_data` text,
          `response_data` text,
          `status` tinyint NOT NULL DEFAULT 1,
          `ip` varchar(50) NOT NULL DEFAULT '''',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
-- 操作日志归档分表 (operation_logs_archive_0 ~ operation_logs_archive_9)
-- ----------------------------
DELIMITER //
CREATE PROCEDURE create_operation_logs_archive_shards()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 10 DO
        SET @create_sql = CONCAT('
        CREATE TABLE IF NOT EXISTS `operation_logs_archive_', i, '` (
          `id` bigint NOT NULL,
          `shop_id` bigint unsigned NOT NULL DEFAULT 0,
          `order_sn` varchar(64) NOT NULL DEFAULT '''',
          `operation_type` varchar(50) NOT NULL,
          `operation_desc` varchar(512) NOT NULL DEFAULT '''',
          `request_data` text,
          `response_data` text,
          `status` tinyint NOT NULL DEFAULT 1,
          `ip` varchar(50) NOT NULL DEFAULT '''',
          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
-- 统计汇总表（不分表）
-- ----------------------------

-- 订单每日统计（按店铺）
DROP TABLE IF EXISTS `order_daily_stats`;
CREATE TABLE `order_daily_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `stat_date` date NOT NULL COMMENT '统计日期',
  `shop_id` bigint unsigned NOT NULL COMMENT '店铺ID',
  `order_count` bigint NOT NULL DEFAULT 0 COMMENT '订单数量',
  `total_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '订单总金额',
  `shipped_count` bigint NOT NULL DEFAULT 0 COMMENT '已发货数量',
  `settled_count` bigint NOT NULL DEFAULT 0 COMMENT '已结算数量',
  `settled_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '结算金额',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_stat_shop` (`stat_date`, `shop_id`),
  KEY `idx_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单每日统计表';

-- 财务每日统计（按店铺）
DROP TABLE IF EXISTS `finance_daily_stats`;
CREATE TABLE `finance_daily_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `stat_date` date NOT NULL COMMENT '统计日期',
  `shop_id` bigint unsigned NOT NULL COMMENT '店铺ID',
  `income_count` bigint NOT NULL DEFAULT 0 COMMENT '收入笔数',
  `income_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '收入金额',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_stat_shop` (`stat_date`, `shop_id`),
  KEY `idx_shop_id` (`shop_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='财务每日统计表';

-- 平台每日统计（汇总）
DROP TABLE IF EXISTS `platform_daily_stats`;
CREATE TABLE `platform_daily_stats` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `stat_date` date NOT NULL COMMENT '统计日期',
  `total_orders` bigint NOT NULL DEFAULT 0 COMMENT '总订单数',
  `total_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '总订单金额',
  `settled_amount` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '总结算金额',
  `platform_share` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '平台分成',
  `total_income` decimal(15,2) NOT NULL DEFAULT 0.00 COMMENT '总收入',
  `active_shops` bigint NOT NULL DEFAULT 0 COMMENT '活跃店铺数',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_stat_date` (`stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台每日统计表';

-- ============================================================================
-- 第三部分：初始化数据
-- ============================================================================

INSERT INTO `platform_commission_accounts` (`id`, `balance`, `frozen_amount`, `total_earnings`, `total_withdrawn`, `currency`, `status`) 
VALUES (1, 0.00, 0.00, 0.00, 0.00, 'TWD', 1);

INSERT INTO `penalty_bonus_accounts` (`id`, `balance`, `total_penalty`, `total_bonus`, `currency`, `status`) 
VALUES (1, 0.00, 0.00, 0.00, 'TWD', 1);

INSERT INTO `escrow_accounts` (`id`, `balance`, `total_in`, `total_out`, `currency`, `status`) 
VALUES (1, 0.00, 0.00, 0.00, 'TWD', 1);

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================================
-- 使用说明
-- ============================================================================
-- 分表路由规则:
--   - 订单相关表: table_name = base_name + "_" + (shop_id % 10)
--   - 账户流水表: table_name = "account_transactions_" + (admin_id % 10)
-- 
-- 示例:
--   shop_id = 12345 -> 12345 % 10 = 5 -> orders_5, finance_incomes_5, ...
--   admin_id = 67890 -> 67890 % 10 = 0 -> account_transactions_0
--
-- 分表列表 (共12种 × 10个 = 120个分表):
--
--   按 shop_id % 10 分表（业务表）:
--   - orders_0 ~ orders_9
--   - order_items_0 ~ order_items_9
--   - order_addresses_0 ~ order_addresses_9
--   - order_escrows_0 ~ order_escrows_9
--   - order_escrow_items_0 ~ order_escrow_items_9
--   - order_settlements_0 ~ order_settlements_9
--   - order_shipment_records_0 ~ order_shipment_records_9
--   - shipments_0 ~ shipments_9
--   - finance_incomes_0 ~ finance_incomes_9
--   - operation_logs_0 ~ operation_logs_9
--
--   按 shop_id % 10 分表（归档表）:
--   - operation_logs_archive_0 ~ operation_logs_archive_9
--
--   按 admin_id % 10 分表:
--   - account_transactions_0 ~ account_transactions_9
--
-- 统计汇总表（不分表，用于平台级查询优化）:
--   - order_daily_stats      订单每日统计（按店铺）
--   - finance_daily_stats    财务每日统计（按店铺）
--   - platform_daily_stats   平台每日统计（汇总）
--
-- 维护任务:
--   - 每天凌晨2点: 归档90天前的操作日志到 operation_logs_archive_X
--   - 每天凌晨3点: 生成前一天的统计数据
--   - 每月1号凌晨4点: 清理365天前的归档数据
--
-- 注意事项:
--   1. 同一店铺的所有订单数据都在同一组分表中
--   2. 跨店铺查询需要遍历所有分表
--   3. 建议在应用层使用 ShardedDB 工具类进行分表路由
--   4. account_transactions 按 admin_id 分表，同一用户的流水在同一表中
--   5. 平台级统计查询优先使用汇总表，避免遍历分表
