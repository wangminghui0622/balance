# 虾皮店铺订单发货平台 API 文档

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json`

## 统一响应格式

```json
{
    "code": 0,
    "message": "success",
    "data": {}
}
```

- `code`: 0 表示成功，其他表示错误
- `message`: 响应消息
- `data`: 响应数据

## 分页响应格式

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 100,
        "page": 1,
        "page_size": 10,
        "list": []
    }
}
```

---

## 店铺管理

### 1. 获取授权链接

获取虾皮OAuth授权链接，用于店铺授权。

**请求**
```
GET /shops/auth-url?region=SG
```

**参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| region | string | 否 | 地区代码，默认SG |

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "url": "https://partner.shopeemobile.com/api/v2/shop/auth_partner?..."
    }
}
```

### 2. 授权回调

虾皮授权完成后的回调接口。

**请求**
```
GET /auth/callback?code=xxx&shop_id=xxx&region=SG
```

**参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 授权码 |
| shop_id | int | 是 | 虾皮店铺ID |
| region | string | 否 | 地区代码，默认SG |

### 3. 获取店铺列表

**请求**
```
GET /shops?page=1&page_size=10&status=1
```

**参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页数量，默认10 |
| status | int | 否 | 状态筛选：0-禁用 1-正常 |

### 4. 获取店铺详情

**请求**
```
GET /shops/:shop_id
```

### 5. 更新店铺状态

**请求**
```
PUT /shops/:shop_id/status
```

**请求体**
```json
{
    "status": 1
}
```

### 6. 删除店铺

**请求**
```
DELETE /shops/:shop_id
```

### 7. 刷新Token

**请求**
```
POST /shops/:shop_id/refresh-token
```

---

## 订单管理

### 1. 同步订单

从虾皮同步订单到本地数据库。

**请求**
```
POST /orders/sync
```

**请求体**
```json
{
    "shop_id": 123456,
    "time_from": 1704067200,
    "time_to": 1704672000,
    "order_status": "READY_TO_SHIP"
}
```

**参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | int | 是 | 店铺ID |
| time_from | int | 否 | 开始时间戳，默认7天前 |
| time_to | int | 否 | 结束时间戳，默认当前 |
| order_status | string | 否 | 订单状态筛选 |

### 2. 获取订单列表

**请求**
```
GET /orders?shop_id=123456&status=READY_TO_SHIP&page=1&page_size=10
```

**参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | int | 否 | 店铺ID |
| status | string | 否 | 订单状态 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

### 3. 获取待发货订单

**请求**
```
GET /orders/ready-to-ship?shop_id=123456&page=1&page_size=10
```

### 4. 获取订单详情

**请求**
```
GET /orders/:shop_id/:order_sn
```

### 5. 刷新单个订单

从虾皮API重新获取订单最新信息。

**请求**
```
POST /orders/:shop_id/:order_sn/refresh
```

---

## 发货管理

### 1. 订单发货

**请求**
```
POST /shipments/ship
```

**请求体**
```json
{
    "shop_id": 123456,
    "order_sn": "2401010001",
    "tracking_number": "SF1234567890",
    "shipping_carrier": "J&T Express"
}
```

### 2. 批量发货

**请求**
```
POST /shipments/batch-ship
```

**请求体**
```json
{
    "orders": [
        {
            "shop_id": 123456,
            "order_sn": "2401010001",
            "tracking_number": "SF1234567890"
        },
        {
            "shop_id": 123456,
            "order_sn": "2401010002",
            "tracking_number": "SF1234567891"
        }
    ]
}
```

**响应**
```json
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "shop_id": 123456,
            "order_sn": "2401010001",
            "success": true
        },
        {
            "shop_id": 123456,
            "order_sn": "2401010002",
            "success": false,
            "error": "订单状态不允许发货"
        }
    ]
}
```

### 3. 获取发货参数

获取订单的发货参数（物流方式、取件时间等）。

**请求**
```
GET /shipments/shipping-parameter?shop_id=123456&order_sn=2401010001
```

### 4. 获取运单号

**请求**
```
GET /shipments/tracking-number?shop_id=123456&order_sn=2401010001
```

### 5. 获取发货记录列表

**请求**
```
GET /shipments?shop_id=123456&status=1&page=1&page_size=10
```

**参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| shop_id | int | 否 | 店铺ID |
| status | int | 否 | 发货状态：0-待发货 1-已发货 2-失败 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

### 6. 获取发货记录详情

**请求**
```
GET /shipments/:shop_id/:order_sn
```

### 7. 同步物流渠道

**请求**
```
POST /shipments/sync-logistics/:shop_id
```

### 8. 获取物流渠道列表

**请求**
```
GET /shipments/logistics/:shop_id
```

---

## 订单状态说明

| 状态 | 说明 |
|------|------|
| UNPAID | 未付款 |
| READY_TO_SHIP | 待发货 |
| PROCESSED | 已处理 |
| SHIPPED | 已发货 |
| COMPLETED | 已完成 |
| IN_CANCEL | 取消中 |
| CANCELLED | 已取消 |
| INVOICE_PENDING | 待开票 |

---

## 地区代码

| 代码 | 地区 |
|------|------|
| SG | 新加坡 |
| MY | 马来西亚 |
| TH | 泰国 |
| TW | 台湾 |
| VN | 越南 |
| PH | 菲律宾 |
| BR | 巴西 |
| MX | 墨西哥 |
| CO | 哥伦比亚 |
| CL | 智利 |
