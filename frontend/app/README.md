# 天平系统 H5 应用

基于 UniApp 开发的虾皮订单管理 H5 应用，支持店铺端、运营端、平台端三种角色。

## 项目结构

```
app/
├── src/                    # 主包页面
│   ├── pages/
│   │   ├── index/         # 首页（路由分发）
│   │   ├── login/         # 登录页
│   │   └── register/      # 注册页
│   ├── App.vue
│   └── main.ts
├── shopower/              # 店铺端分包
│   └── pages/
│       ├── home/          # 店铺首页
│       ├── stores/        # 店铺管理
│       └── orders/        # 订单管理
├── operator/              # 运营端分包
│   └── pages/
│       └── home/          # 运营首页
├── platform/              # 平台端分包
│   └── pages/
│       └── home/          # 平台首页
├── share/                 # 共享模块
│   ├── api/               # API 接口
│   ├── config/            # 配置（环境切换）
│   ├── constants/         # 常量定义
│   └── utils/             # 工具函数
├── pages.json             # 页面配置
├── manifest.json          # 应用配置
├── package.json
└── vite.config.ts
```

## 开发环境

```bash
# 安装依赖
pnpm install

# 启动 H5 开发服务器
pnpm dev:h5

# 构建 H5 生产版本
pnpm build:h5
```

## 环境配置

网络请求自动区分开发/线上环境：

- **开发环境**：使用 Vite proxy 代理到 `http://localhost:19090`
- **生产环境**：直接请求 `https://kx9y.com`

配置文件：`share/config/api.ts`

## 用户类型

| 类型 | ID前缀 | 说明 |
|------|--------|------|
| 店主 | 1 | 店铺拥有者 |
| 运营 | 5 | 代运营人员 |
| 平台 | 9 | 平台管理员 |

## API 接口

所有 App 端接口前缀：`/api/v1/balance/app`

### 认证接口
- `POST /auth/register` - 注册
- `POST /auth/login` - 登录
- `GET /auth/me` - 获取当前用户信息

### 店铺管理
- `GET /shops` - 店铺列表
- `GET /shops/auth-url` - 获取授权URL
- `POST /shops/bind` - 绑定店铺
- `GET /shops/:shop_id` - 店铺详情
- `PUT /shops/:shop_id/status` - 更新店铺状态
- `DELETE /shops/:shop_id` - 删除店铺

### 订单管理
- `GET /orders` - 订单列表
- `POST /orders/sync` - 同步订单
- `GET /orders/ready-to-ship` - 待发货订单
- `GET /orders/:shop_id/:order_sn` - 订单详情

### 发货管理
- `POST /shipments/ship` - 发货
- `POST /shipments/batch-ship` - 批量发货
- `GET /shipments` - 发货记录列表

## 部署

1. 构建生产版本：
```bash
pnpm build:h5
```

2. 将 `dist/build/h5` 目录部署到服务器 `/usr/local/nginx/balance/app/`

3. Nginx 配置已在 `nginx.conf.good` 中配置完成
