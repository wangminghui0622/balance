# Balance 数据库设计文档

## 一、数据库概览

- **数据库**: MySQL 8.0+
- **字符集**: utf8mb4
- **排序规则**: utf8mb4_unicode_ci

---

## 二、表结构

### 2.1 用户与认证

#### admins (管理员/用户表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| email | varchar(100) | 邮箱 (唯一) |
| password | varchar(255) | 密码哈希 |
| nickname | varchar(50) | 昵称 |
| avatar | varchar(255) | 头像URL |
| user_type | tinyint | 用户类型: 1=店主, 5=运营, 9=平台 |
| status | tinyint | 状态: 1=正常, 2=禁用 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

---

### 2.2 店铺相关

#### shops (店铺表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | Shopee 店铺ID (唯一) |
| admin_id | bigint | 所属用户ID |
| shop_name | varchar(200) | 店铺名称 |
| region | varchar(10) | 地区代码 |
| status | tinyint | 状态: 1=正常, 2=禁用 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### shop_authorizations (店铺授权表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | 店铺ID |
| access_token | text | 访问令牌 |
| refresh_token | text | 刷新令牌 |
| expire_in | int | 过期时间(秒) |
| token_created_at | datetime | 令牌创建时间 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### shop_operator_relations (店铺-运营关系表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | 店铺ID |
| shop_owner_id | bigint | 店主ID |
| operator_id | bigint | 运营ID |
| status | tinyint | 状态: 1=生效, 2=失效 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

---

### 2.3 订单相关

#### orders (订单表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | 店铺ID |
| order_sn | varchar(64) | 订单号 (唯一) |
| order_status | varchar(30) | 订单状态 |
| total_amount | decimal(15,2) | 订单总金额 |
| currency | varchar(10) | 货币 |
| buyer_user_id | bigint | 买家ID |
| buyer_username | varchar(100) | 买家用户名 |
| pay_time | datetime | 支付时间 |
| ship_by_date | datetime | 最晚发货时间 |
| create_time | datetime | Shopee创建时间 |
| update_time | datetime | Shopee更新时间 |
| created_at | datetime | 本地创建时间 |
| updated_at | datetime | 本地更新时间 |

**索引**:
- `uk_order_sn` (order_sn) - 唯一索引
- `idx_shop_id` (shop_id)
- `idx_order_status` (order_status)
- `idx_create_time` (create_time)

#### order_items (订单商品表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| order_id | bigint unsigned | 订单ID |
| item_id | bigint unsigned | 商品ID |
| item_name | varchar(500) | 商品名称 |
| item_sku | varchar(100) | SKU |
| model_id | bigint unsigned | 规格ID |
| model_name | varchar(200) | 规格名称 |
| model_quantity_purchased | int | 购买数量 |
| model_original_price | decimal(15,2) | 原价 |
| model_discounted_price | decimal(15,2) | 折后价 |
| created_at | datetime | 创建时间 |

#### order_addresses (订单地址表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| order_id | bigint unsigned | 订单ID |
| name | varchar(200) | 收件人姓名 |
| phone | varchar(50) | 电话 |
| full_address | text | 完整地址 |
| city | varchar(100) | 城市 |
| state | varchar(100) | 省/州 |
| zipcode | varchar(20) | 邮编 |
| created_at | datetime | 创建时间 |

---

### 2.4 发货相关

#### shipments (发货记录表 - Shopee同步)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | 店铺ID |
| order_sn | varchar(64) | 订单号 |
| tracking_number | varchar(100) | 物流单号 |
| shipping_carrier | varchar(100) | 物流商 |
| ship_status | varchar(30) | 发货状态 |
| shipped_at | datetime | 发货时间 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### order_shipment_records (运营发货记录表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | 店铺ID |
| order_sn | varchar(64) | 订单号 (唯一) |
| order_id | bigint unsigned | 订单ID |
| shop_owner_id | bigint | 店主ID |
| operator_id | bigint | 运营ID |
| goods_cost | decimal(15,2) | 商品成本 |
| shipping_cost | decimal(15,2) | 运费成本 |
| total_cost | decimal(15,2) | 总成本 |
| currency | varchar(10) | 货币 |
| prepayment_amount | decimal(15,2) | 预付款金额 |
| deduct_tx_id | bigint unsigned | 扣款流水ID |
| shipping_carrier | varchar(100) | 物流商 |
| tracking_number | varchar(100) | 物流单号 |
| shipped_at | datetime | 发货时间 |
| status | tinyint | 状态: 0=待发货, 1=已发货, 2=已完成, 3=已取消, 4=发货失败 |
| settlement_id | bigint unsigned | 关联结算记录ID |
| remark | varchar(500) | 备注 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

**索引**:
- `uk_order_sn` (order_sn) - 唯一索引
- `idx_shop_id` (shop_id)
- `idx_shop_owner_id` (shop_owner_id)
- `idx_operator_id` (operator_id)
- `idx_status` (status)

---

### 2.5 结算相关

#### order_escrows (订单结算明细表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | 店铺ID |
| order_sn | varchar(64) | 订单号 |
| escrow_amount | decimal(15,2) | 最终结算金额 |
| commission_fee | decimal(15,2) | 平台佣金 |
| service_fee | decimal(15,2) | 服务费 |
| seller_discount | decimal(15,2) | 卖家折扣 |
| shopee_discount | decimal(15,2) | 平台折扣 |
| actual_shipping_fee | decimal(15,2) | 实际运费 |
| ... | ... | 30+ 结算相关字段 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### order_settlements (订单结算记录表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| settlement_no | varchar(64) | 结算单号 (唯一) |
| shop_id | bigint unsigned | 店铺ID |
| order_sn | varchar(64) | 订单号 (唯一) |
| order_id | bigint unsigned | 订单ID |
| shop_owner_id | bigint | 店主ID |
| operator_id | bigint | 运营ID |
| currency | varchar(10) | 货币 |
| escrow_amount | decimal(15,2) | Shopee结算金额 |
| goods_cost | decimal(15,2) | 商品成本 |
| shipping_cost | decimal(15,2) | 运费成本 |
| total_cost | decimal(15,2) | 总成本 |
| profit | decimal(15,2) | 利润 |
| platform_share_rate | decimal(5,2) | 平台分成比例% |
| operator_share_rate | decimal(5,2) | 运营分成比例% |
| shop_owner_share_rate | decimal(5,2) | 店主分成比例% |
| platform_share | decimal(15,2) | 平台分成 |
| operator_share | decimal(15,2) | 运营分成 |
| shop_owner_share | decimal(15,2) | 店主分成 |
| operator_income | decimal(15,2) | 运营实际收入 |
| status | tinyint | 状态: 0=待结算, 1=已结算, 2=已取消 |
| settled_at | datetime | 结算时间 |
| remark | varchar(500) | 备注 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

**索引**:
- `uk_settlement_no` (settlement_no) - 唯一索引
- `uk_order_sn` (order_sn) - 唯一索引
- `idx_shop_id` (shop_id)
- `idx_shop_owner_id` (shop_owner_id)
- `idx_operator_id` (operator_id)
- `idx_status` (status)

#### profit_share_configs (利润分成配置表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | 店铺ID |
| operator_id | bigint | 运营ID |
| platform_share_rate | decimal(5,2) | 平台分成比例% (默认5) |
| operator_share_rate | decimal(5,2) | 运营分成比例% (默认45) |
| shop_owner_share_rate | decimal(5,2) | 店主分成比例% (默认50) |
| status | tinyint | 状态: 1=生效, 2=失效 |
| effective_from | datetime | 生效时间 |
| effective_to | datetime | 失效时间 |
| remark | varchar(500) | 备注 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

**索引**:
- `uk_shop_operator_config` (shop_id, operator_id) - 唯一索引

---

### 2.6 财务相关

#### finance_incomes (财务收入表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | 店铺ID |
| transaction_id | bigint | Shopee交易ID (唯一) |
| order_sn | varchar(64) | 关联订单号 |
| transaction_type | varchar(50) | 交易类型 |
| amount | decimal(15,2) | 金额 |
| current_balance | decimal(15,2) | 当前余额 |
| create_time | datetime | Shopee创建时间 |
| settlement_handle_status | varchar(30) | 结算处理状态 |
| created_at | datetime | 本地创建时间 |
| updated_at | datetime | 本地更新时间 |

**交易类型**:
- `ESCROW_VERIFIED_ADD` - 订单结算收入
- `WITHDRAWAL_CREATED` - 提现创建
- `WITHDRAWAL_COMPLETED` - 提现完成
- `REFUND` - 退款

---

### 2.7 账户相关

#### prepayment_accounts (预付款账户表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| admin_id | bigint | 用户ID (唯一) |
| balance | decimal(15,2) | 可用余额 |
| pending_amount | decimal(15,2) | 待结算金额 |
| total_recharge | decimal(15,2) | 累计充值 |
| total_consume | decimal(15,2) | 累计消费 |
| currency | varchar(10) | 货币 (默认TWD) |
| status | tinyint | 状态: 1=正常, 2=冻结 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### deposit_accounts (保证金账户表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| admin_id | bigint | 用户ID (唯一) |
| balance | decimal(15,2) | 保证金余额 |
| required_amount | decimal(15,2) | 应缴金额 |
| currency | varchar(10) | 货币 |
| status | tinyint | 状态: 1=正常, 2=不足, 3=冻结 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### operator_accounts (运营账户表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| admin_id | bigint | 用户ID (唯一) |
| balance | decimal(15,2) | 可用余额 |
| pending_amount | decimal(15,2) | 待结算金额 |
| total_earnings | decimal(15,2) | 累计收益 |
| total_withdrawn | decimal(15,2) | 累计提现 |
| currency | varchar(10) | 货币 |
| status | tinyint | 状态: 1=正常, 2=冻结 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### shop_owner_commission_accounts (店主佣金账户表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| admin_id | bigint | 用户ID (唯一) |
| balance | decimal(15,2) | 可用余额 |
| pending_amount | decimal(15,2) | 待结算金额 |
| total_earnings | decimal(15,2) | 累计收益 |
| total_withdrawn | decimal(15,2) | 累计提现 |
| currency | varchar(10) | 货币 |
| status | tinyint | 状态: 1=正常, 2=冻结 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### platform_commission_accounts (平台佣金账户表 - 单例)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| balance | decimal(15,2) | 可用余额 |
| pending_amount | decimal(15,2) | 待结算金额 |
| total_earnings | decimal(15,2) | 累计收益 |
| total_withdrawn | decimal(15,2) | 累计提现 |
| currency | varchar(10) | 货币 |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### penalty_bonus_accounts (罚补账户表 - 单例)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| balance | decimal(15,2) | 余额 |
| total_penalty | decimal(15,2) | 累计罚款 |
| total_bonus | decimal(15,2) | 累计补贴 |
| currency | varchar(10) | 货币 |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### escrow_accounts (托管账户表 - 单例)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| balance | decimal(15,2) | 托管余额 |
| total_in | decimal(15,2) | 累计转入 |
| total_out | decimal(15,2) | 累计转出 |
| currency | varchar(10) | 货币 |
| status | tinyint | 状态 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### account_transactions (账户流水表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| transaction_no | varchar(64) | 流水号 (唯一) |
| account_type | varchar(20) | 账户类型 |
| admin_id | bigint | 用户ID |
| transaction_type | varchar(30) | 交易类型 |
| amount | decimal(15,2) | 金额 (正=入账, 负=出账) |
| balance_before | decimal(15,2) | 交易前余额 |
| balance_after | decimal(15,2) | 交易后余额 |
| related_order_sn | varchar(64) | 关联订单号 |
| related_id | bigint unsigned | 关联ID |
| remark | varchar(500) | 备注 |
| operator_id | bigint | 操作人ID |
| status | tinyint | 状态: 0=待审批, 1=已完成, 2=已拒绝 |
| created_at | datetime | 创建时间 |

**索引**:
- `uk_transaction_no` (transaction_no) - 唯一索引
- `idx_account_type` (account_type)
- `idx_admin_id` (admin_id)
- `idx_transaction_type` (transaction_type)
- `idx_related_order_sn` (related_order_sn)
- `idx_created_at` (created_at)

---

### 2.8 申请相关

#### withdraw_applications (提现申请表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| application_no | varchar(64) | 申请单号 (唯一) |
| admin_id | bigint | 申请人ID |
| account_type | varchar(30) | 账户类型 |
| amount | decimal(15,2) | 提现金额 |
| fee | decimal(15,2) | 手续费 |
| actual_amount | decimal(15,2) | 实际到账金额 |
| currency | varchar(10) | 货币 |
| collection_account_id | bigint unsigned | 收款账户ID |
| status | tinyint | 状态: 0=待审核, 1=已通过, 2=已拒绝, 3=已打款 |
| audit_remark | varchar(500) | 审核备注 |
| audit_by | bigint | 审核人ID |
| audit_at | datetime | 审核时间 |
| paid_at | datetime | 打款时间 |
| remark | varchar(500) | 申请备注 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

**索引**:
- `uk_application_no` (application_no) - 唯一索引
- `idx_admin_id` (admin_id)
- `idx_account_type` (account_type)
- `idx_status` (status)
- `idx_created_at` (created_at)

#### recharge_applications (充值申请表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| application_no | varchar(64) | 申请单号 (唯一) |
| admin_id | bigint | 申请人ID |
| account_type | varchar(30) | 账户类型 |
| amount | decimal(15,2) | 充值金额 |
| currency | varchar(10) | 货币 |
| payment_method | varchar(30) | 支付方式 |
| payment_proof | varchar(500) | 支付凭证URL |
| status | tinyint | 状态: 0=待审核, 1=已通过, 2=已拒绝 |
| audit_remark | varchar(500) | 审核备注 |
| audit_by | bigint | 审核人ID |
| audit_at | datetime | 审核时间 |
| remark | varchar(500) | 申请备注 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### collection_accounts (收款账户表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| admin_id | bigint | 用户ID |
| account_type | varchar(20) | 账户类型: wallet/bank |
| account_name | varchar(100) | 账户名称 |
| account_no | varchar(100) | 账号 |
| bank_name | varchar(100) | 银行名称 |
| bank_branch | varchar(200) | 银行支行 |
| payee | varchar(100) | 收款人 |
| is_default | tinyint(1) | 是否默认 |
| status | tinyint | 状态: 1=正常, 2=未激活 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

---

### 2.9 同步与日志

#### shop_sync_records (店铺同步记录表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | 店铺ID |
| sync_type | varchar(30) | 同步类型 |
| last_sync_time | datetime | 上次同步时间 |
| last_sync_cursor | varchar(100) | 上次同步游标 |
| status | tinyint | 状态 |
| error_message | text | 错误信息 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

#### operation_logs (操作日志表)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | 店铺ID |
| admin_id | bigint | 操作人ID |
| operation_type | varchar(50) | 操作类型 |
| target_type | varchar(50) | 目标类型 |
| target_id | varchar(100) | 目标ID |
| content | text | 操作内容 |
| ip | varchar(50) | IP地址 |
| created_at | datetime | 创建时间 |

---

## 三、ER 关系图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              核心实体关系                                    │
└─────────────────────────────────────────────────────────────────────────────┘

                              ┌──────────┐
                              │  admins  │
                              └────┬─────┘
                                   │
          ┌────────────────────────┼────────────────────────┐
          │                        │                        │
          ▼                        ▼                        ▼
    ┌──────────┐            ┌──────────┐            ┌──────────────┐
    │  shops   │            │ accounts │            │ applications │
    └────┬─────┘            └──────────┘            └──────────────┘
         │                        │
         │                        │
    ┌────┴────┐              ┌────┴────┐
    │         │              │         │
    ▼         ▼              ▼         ▼
┌────────┐ ┌────────┐  ┌──────────┐ ┌──────────┐
│ orders │ │shipment│  │prepayment│ │ operator │
│        │ │records │  │ account  │ │ account  │
└────┬───┘ └────────┘  └──────────┘ └──────────┘
     │
     │
┌────┴────┐
│         │
▼         ▼
┌────────┐ ┌────────────┐
│ items  │ │ settlements│
└────────┘ └────────────┘
```

---

## 四、索引策略

### 4.1 主要查询场景

| 场景 | 涉及表 | 索引 |
|------|--------|------|
| 按店铺查订单 | orders | idx_shop_id |
| 按状态查订单 | orders | idx_order_status |
| 按时间查订单 | orders | idx_create_time |
| 按用户查账户 | *_accounts | uk_admin_id |
| 按状态查申请 | *_applications | idx_status |
| 按订单查结算 | order_settlements | uk_order_sn |

### 4.2 复合索引

| 表 | 索引 | 字段 |
|------|------|------|
| orders | idx_shop_status | (shop_id, order_status) |
| account_transactions | idx_account_admin | (account_type, admin_id) |
| shop_operator_relations | uk_shop_operator | (shop_id, operator_id) |

---

## 五、数据迁移

### 5.1 初始化脚本

```sql
-- 初始化平台账户 (单例)
INSERT INTO platform_commission_accounts (balance, currency, status) 
VALUES (0.00, 'TWD', 1);

INSERT INTO penalty_bonus_accounts (balance, currency, status) 
VALUES (0.00, 'TWD', 1);

INSERT INTO escrow_accounts (balance, currency, status) 
VALUES (0.00, 'TWD', 1);
```

### 5.2 迁移注意事项

1. **账户表**: 用户首次访问时自动创建
2. **平台账户**: 系统启动时初始化单例
3. **分成配置**: 创建合作关系时设置默认值
