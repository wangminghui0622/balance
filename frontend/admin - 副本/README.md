# XShopee V2.1 管理后台

基于 Vue3 + Element Plus + TypeScript + Vite 构建的管理后台系统。

## 项目结构

```
frontend/admin/
├── src/                    # 主应用代码
│   ├── main.ts            # 入口文件
│   ├── App.vue            # 根组件
│   └── router/            # 路由配置
├── share/                 # 共享模块
│   ├── components/        # 共享组件
│   ├── utils/             # 工具函数
│   └── types/             # 类型定义
├── platform/              # 平台模块
│   ├── layouts/           # 布局组件
│   ├── views/             # 页面组件
│   └── components/        # 业务组件
├── operator/              # 运维人员模块
│   ├── layouts/
│   └── views/
└── shopower/              # 店主模块
    ├── layouts/
    └── views/
```

## 安装依赖

```bash
npm install
```

## 开发运行

```bash
npm run dev
```

访问 http://localhost:3000

## 构建生产版本

```bash
npm run build
```

## 模块说明

### Platform（平台模块）
- 平台用户数据统计
- 平台警示用户
- 今日订单（含图表）
- 店主充值预警
- 运营业绩预警
- 运营违规预警
- 店主异常预警

### Operator（运维人员模块）
- 店铺管理
- 订单管理
- 数据报表

### Shopowner（店主模块）
- 我的店铺
- 订单管理
- 财务管理

## 技术栈

- Vue 3.3+
- Element Plus 2.4+
- TypeScript 5.2+
- Vite 5.0+
- Vue Router 4.2+
- ECharts 5.4+
- Axios 1.6+
