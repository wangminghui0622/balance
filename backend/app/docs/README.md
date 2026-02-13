# Balance App 移动端/小程序 API

## 项目概述

Balance App 是虾皮店铺订单发货平台的移动端/小程序后端服务，为店主和运营提供移动端访问能力。

## 技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.21+ |
| Web框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0+ |
| 缓存 | Redis 7.0+ |
| 认证 | JWT |

## 与 Admin 的区别

| 特性 | Admin | App |
|------|-------|-----|
| 用途 | Web后台管理 | 移动端/小程序 |
| 用户 | 店主/运营/平台 | 店主/运营 |
| 功能 | 完整功能 | 核心功能 |
| 接口 | RESTful | RESTful (精简) |

## 目录结构

```
backend/app/
├── main.go                 # 入口文件
├── config/                 # 配置文件
│   └── config.yaml
├── docs/                   # 文档
│   ├── README.md           # 项目说明
│   ├── api.md              # API文档
│   └── database_design.md  # 数据库设计
└── internal/
    ├── handlers/           # HTTP处理器
    │   └── auth_handler.go
    └── router/             # 路由配置
        └── router.go
```

## 快速开始

### 1. 配置文件

```yaml
# config/config.yaml
app:
  port: 8081
  mode: release

mysql:
  host: localhost
  port: 3306
  user: root
  password: your_password
  database: balance

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: your_jwt_secret
  expire: 168h  # 移动端token有效期更长
```

### 2. 启动服务

```bash
go run main.go -config config/config.yaml
```

## API 路由

| 模块 | 路由前缀 |
|------|----------|
| 认证 | `/api/v1/balance/app/auth` |
| 店铺 | `/api/v1/balance/app/shops` |
| 订单 | `/api/v1/balance/app/orders` |
| 发货 | `/api/v1/balance/app/shipments` |
| 账户 | `/api/v1/balance/app/account` |

## 核心功能

### 店主功能
- 登录/注册
- 查看店铺列表
- 查看订单列表
- 快速发货
- 查看账户余额
- 提现申请

### 运营功能
- 登录
- 查看分配店铺
- 待发货订单
- 快速发货
- 查看收益

## 推送通知

支持以下推送场景：
- 新订单通知
- 发货成功通知
- 提现审核通知
- 结算到账通知

## 相关文档

- [API文档](api.md)
- [数据库设计](database_design.md)
