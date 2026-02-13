# Balance Admin 后台管理系统

## 项目概述

Balance Admin 是虾皮店铺订单发货平台的后台管理系统，提供店铺管理、订单同步、发货管理、财务结算等功能。

## 技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.21+ |
| Web框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0+ |
| 缓存 | Redis 7.0+ |
| 认证 | JWT |
| 监控 | Prometheus + Grafana |

## 目录结构

```
backend/admin/
├── main.go                 # 入口文件
├── config/                 # 配置文件
│   └── config.yaml
├── docs/                   # 文档
│   ├── README.md           # 项目说明
│   ├── api.md              # API文档
│   ├── database_design.md  # 数据库设计
│   ├── backup_strategy.md  # 备份策略
│   ├── deployment.md       # 部署指南
│   └── architecture.md     # 架构设计
└── internal/
    ├── handlers/           # HTTP处理器
    │   ├── auth_handler.go
    │   ├── operator/       # 运营端处理器
    │   ├── platform/       # 平台端处理器
    │   └── shopower/       # 店主端处理器
    └── router/             # 路由配置
        └── router.go
```

## 快速开始

### 1. 环境准备

```bash
# 安装依赖
go mod download

# 配置数据库
mysql -u root -p < ../../database.sql
```

### 2. 配置文件

```yaml
# config/config.yaml
app:
  port: 8080
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
  expire: 24h
```

### 3. 启动服务

```bash
go run main.go -config config/config.yaml
```

## 用户角色

| 角色 | user_type | 说明 |
|------|-----------|------|
| 店主 | 1 | 店铺所有者，管理自己的店铺和订单 |
| 运营 | 5 | 代运营人员，帮助店主发货 |
| 平台 | 9 | 平台管理员，管理所有用户和店铺 |

## API 路由前缀

| 角色 | 路由前缀 |
|------|----------|
| 公共 | `/api/v1/balance/admin/auth` |
| 店主 | `/api/v1/balance/admin/shopower` |
| 运营 | `/api/v1/balance/admin/operator` |
| 平台 | `/api/v1/balance/admin/platform` |

## 核心功能

### 店主功能
- 店铺授权与管理
- 订单同步与查看
- 发货管理
- 财务收入查看
- 账户管理（预付款、保证金、佣金）
- 提现申请

### 运营功能
- 查看分配的店铺
- 代发货操作
- 结算查看
- 账户管理
- 提现申请

### 平台功能
- 用户管理
- 店铺管理
- 合作关系管理（店铺-运营分配）
- 结算审核
- 财务审核（提现/充值）
- 佣金管理
- 罚补管理

## 相关文档

- [API文档](api.md)
- [数据库设计](database_design.md)
- [备份策略](backup_strategy.md)
- [部署指南](deployment.md)
- [架构设计](architecture.md)
