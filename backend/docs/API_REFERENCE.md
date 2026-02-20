# Balance API 接口文档

## 基础信息

- **Base URL**: `/api/v1/balance/admin`
- **认证方式**: JWT Token (Header: `Authorization: Bearer <token>`)
- **响应格式**: JSON

### 通用响应结构

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 分页响应结构

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "list": []
  }
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 一、认证接口

### 1.1 用户注册

```
POST /auth/register
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱 |
| password | string | 是 | 密码 |
| nickname | string | 否 | 昵称 |
| user_type | int | 是 | 用户类型: 1=店主, 5=运营 |

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "nickname": "用户昵称",
    "user_type": 1
  }
}
```

### 1.2 用户登录

```
POST /auth/login
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱 |
| password | string | 是 | 密码 |

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "user_type": 1
    }
  }
}
```

### 1.3 获取当前用户信息

```
GET /auth/me
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "nickname": "用户昵称",
    "user_type": 1,
    "status": 1
  }
}
```

---

## 二、店主接口 (/shopower)

### 2.1 店铺管理

#### 获取授权URL

```
GET /shopower/shops/auth-url
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "auth_url": "https://partner.shopeemobile.com/api/v2/shop/auth_partner?..."
  }
}
```

#### 店铺列表

```
GET /shopower/shops
```

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认20 |

#### 店铺详情

```
GET /shopower/shops/:shop_id
```

### 2.2 订单管理

#### 同步订单

```
POST /shopower/orders/sync
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| start_time | int64 | 否 | 开始时间戳 |
| end_time | int64 | 否 | 结束时间戳 |

#### 订单列表

```
GET /shopower/orders
```

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 否 | 店铺ID |
| order_status | string | 否 | 订单状态 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

#### 待发货订单

```
GET /shopower/orders/ready-to-ship
```

#### 订单详情

```
GET /shopower/orders/:shop_id/:order_sn
```

### 2.3 发货管理

#### 发货

```
POST /shopower/shipments/ship
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| order_sn | string | 是 | 订单号 |
| pickup_info | object | 否 | 取件信息 |

#### 批量发货

```
POST /shopower/shipments/batch-ship
```

#### 获取发货参数

```
GET /shopower/shipments/shipping-parameter
```

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| order_sn | string | 是 | 订单号 |

### 2.4 账户管理

#### 获取预付款账户

```
GET /shopower/account/prepayment
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "admin_id": 1,
    "balance": "10000.00",
    "pending_amount": "500.00",
    "total_recharge": "15000.00",
    "total_consume": "5000.00",
    "currency": "TWD",
    "status": 1
  }
}
```

#### 获取保证金账户

```
GET /shopower/account/deposit
```

#### 获取佣金账户

```
GET /shopower/account/commission
```

#### 获取所有账户汇总

```
GET /shopower/account/summary
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "prepayment": { ... },
    "deposit": { ... },
    "commission": { ... }
  }
}
```

#### 获取账户流水

```
GET /shopower/account/prepayment/transactions
GET /shopower/account/deposit/transactions
GET /shopower/account/commission/transactions
```

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

### 2.5 提现管理

#### 申请提现

```
POST /shopower/withdraw/apply
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| account_type | string | 是 | 账户类型: shop_owner_commission/deposit |
| amount | float | 是 | 提现金额 |
| collection_account_id | uint64 | 是 | 收款账户ID |
| remark | string | 否 | 备注 |

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "application_no": "WD1234567890",
    "amount": "1000.00",
    "status": 0
  }
}
```

#### 提现申请列表

```
GET /shopower/withdraw/list
```

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | int | 否 | 状态: -1=全部, 0=待审核, 1=已通过, 2=已拒绝, 3=已打款 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

### 2.6 充值管理

#### 申请充值

```
POST /shopower/recharge
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| account_type | string | 是 | 账户类型: prepayment/deposit |
| amount | float | 是 | 充值金额 |
| payment_method | string | 是 | 支付方式: bank_transfer/cash |
| payment_proof | string | 否 | 支付凭证URL |
| remark | string | 否 | 备注 |

#### 充值申请列表

```
GET /shopower/recharge/list
```

### 2.7 结算管理

#### 结算记录列表

```
GET /shopower/settlements
```

#### 结算统计

```
GET /shopower/settlements/stats
```

---

## 三、运营接口 (/operator)

### 3.1 店铺管理

#### 店铺列表

```
GET /operator/shops
```

#### 店铺详情

```
GET /operator/shops/:shop_id
```

### 3.2 订单管理

#### 订单列表

```
GET /operator/orders
```

#### 订单详情

```
GET /operator/orders/:shop_id/:order_sn
```

#### 待发货订单

```
GET /operator/orders/pending
```

### 3.3 发货管理

#### 运营发货

```
POST /operator/shipments/ship
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| order_sn | string | 是 | 订单号 |
| goods_cost | float | 是 | 商品成本 |
| shipping_cost | float | 是 | 运费成本 |
| pickup_info | object | 否 | 取件信息 |

#### 发货记录列表

```
GET /operator/shipments
```

### 3.4 账户管理

#### 获取运营账户

```
GET /operator/account
```

#### 账户流水

```
GET /operator/account/transactions
```

### 3.5 提现管理

#### 申请提现

```
POST /operator/withdraw/apply
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| amount | float | 是 | 提现金额 |
| collection_account_id | uint64 | 是 | 收款账户ID |
| remark | string | 否 | 备注 |

#### 提现申请列表

```
GET /operator/withdraw/list
```

### 3.6 结算管理

#### 结算记录列表

```
GET /operator/settlements
```

#### 结算统计

```
GET /operator/settlements/stats
```

---

## 四、平台接口 (/platform)

### 4.1 用户管理

#### 用户列表

```
GET /platform/users
```

#### 用户详情

```
GET /platform/users/:id
```

#### 更新用户状态

```
PUT /platform/users/:id/status
```

### 4.2 店铺管理

#### 店铺列表

```
GET /platform/shops
```

#### 店铺详情

```
GET /platform/shops/:shop_id
```

### 4.3 账户管理

#### 预付款账户列表

```
GET /platform/accounts/prepayment
```

#### 保证金账户列表

```
GET /platform/accounts/deposit
```

#### 运营账户列表

```
GET /platform/accounts/operator
```

#### 账户统计

```
GET /platform/accounts/stats
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "prepayment": {
      "total_balance": "100000.00",
      "account_count": 10
    },
    "deposit": {
      "total_balance": "150000.00",
      "account_count": 10
    },
    "operator": {
      "total_balance": "50000.00",
      "account_count": 5
    },
    "shop_owner_commission": {
      "total_balance": "30000.00",
      "account_count": 10
    },
    "platform_commission": {
      "balance": "5000.00",
      "total_earnings": "8000.00"
    }
  }
}
```

#### 平台佣金账户

```
GET /platform/account/commission
```

#### 平台佣金流水

```
GET /platform/account/commission/transactions
```

#### 直接充值预付款

```
POST /platform/accounts/prepayment/recharge
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| admin_id | int64 | 是 | 用户ID |
| amount | float | 是 | 充值金额 |
| remark | string | 否 | 备注 |

#### 直接缴纳保证金

```
POST /platform/accounts/deposit/pay
```

### 4.4 提现审核

#### 提现申请列表

```
GET /platform/withdraw/list
```

**查询参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | int | 否 | 状态筛选 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

#### 审批通过

```
POST /platform/withdraw/approve
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| application_id | uint64 | 是 | 申请ID |
| audit_remark | string | 否 | 审核备注 |

#### 审批拒绝

```
POST /platform/withdraw/reject
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| application_id | uint64 | 是 | 申请ID |
| audit_remark | string | 是 | 拒绝原因 |

#### 确认打款

```
POST /platform/withdraw/confirm_paid
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| application_id | uint64 | 是 | 申请ID |

### 4.5 充值审核

#### 充值申请列表

```
GET /platform/recharge/list
```

#### 审批通过

```
POST /platform/recharge/approve
```

#### 审批拒绝

```
POST /platform/recharge/reject
```

### 4.6 结算管理

#### 结算记录列表

```
GET /platform/settlements
```

#### 结算统计

```
GET /platform/settlements/stats
```

#### 待结算订单

```
GET /platform/settlements/pending
```

#### 执行结算

```
POST /platform/settlements/process
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_sn | string | 是 | 订单号 |

### 4.7 合作管理

#### 合作关系列表

```
GET /platform/cooperations
```

#### 创建合作关系

```
POST /platform/cooperations
```

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | uint64 | 是 | 店铺ID |
| operator_id | int64 | 是 | 运营ID |
| platform_share_rate | float | 否 | 平台分成比例 |
| operator_share_rate | float | 否 | 运营分成比例 |
| shop_owner_share_rate | float | 否 | 店主分成比例 |

#### 更新合作关系

```
PUT /platform/cooperations/:id
```

#### 取消合作关系

```
DELETE /platform/cooperations/:id
```

---

## 五、Webhook 接口

### 5.1 Shopee Webhook

```
POST /webhook/shopee
```

**支持的事件类型**:

| Code | 事件 | 说明 |
|------|------|------|
| 3 | ORDER_STATUS_UPDATE | 订单状态更新 |
| 4 | TRACKING_NO_UPDATE | 物流单号更新 |
| 5 | BUYER_CANCEL_ORDER | 买家取消订单 |

---

## 六、数据字典

### 6.1 订单状态

| 状态 | 说明 |
|------|------|
| UNPAID | 未付款 |
| READY_TO_SHIP | 待发货 |
| PROCESSED | 处理中 |
| SHIPPED | 已发货 |
| COMPLETED | 已完成 |
| CANCELLED | 已取消 |
| IN_CANCEL | 取消中 |
| TO_RETURN | 待退货 |

### 6.2 账户类型

| 类型 | 常量 | 说明 |
|------|------|------|
| 预付款 | prepayment | 店主预付款账户 |
| 保证金 | deposit | 店主保证金账户 |
| 运营账户 | operator | 运营回款账户 |
| 店主佣金 | shop_owner_commission | 店主佣金账户 |
| 平台佣金 | platform_commission | 平台佣金账户 |

### 6.3 申请状态

| 状态 | 值 | 说明 |
|------|-----|------|
| 待审核 | 0 | 等待平台审核 |
| 已通过 | 1 | 审核通过 |
| 已拒绝 | 2 | 审核拒绝 |
| 已打款 | 3 | 提现已打款 |

### 6.4 用户类型

| 类型 | 值 | 说明 |
|------|-----|------|
| 店主 | 1 | 管理自己的店铺 |
| 运营 | 5 | 代运营发货 |
| 平台 | 9 | 平台管理员 |
