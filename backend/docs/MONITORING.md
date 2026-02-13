# 监控告警文档

## 1. 监控架构

```
┌─────────────────────────────────────────────────────────────┐
│                      应用服务                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  /metrics (Prometheus 指标端点)                      │    │
│  │  /health  (健康检查端点)                             │    │
│  └─────────────────────────────────────────────────────┘    │
└────────────────────────────┬────────────────────────────────┘
                             │
                        ┌────▼────┐
                        │Prometheus│
                        │ (采集)   │
                        └────┬────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
    ┌────▼────┐         ┌────▼────┐         ┌────▼────┐
    │ Grafana │         │Alertmgr │         │  告警   │
    │ (展示)  │         │ (告警)  │         │  通知   │
    └─────────┘         └─────────┘         └─────────┘
```

## 2. 指标分类

### 2.1 HTTP 请求指标

| 指标名 | 类型 | 说明 |
|--------|------|------|
| http_requests_total | Counter | HTTP请求总数 |
| http_request_duration_seconds | Histogram | HTTP请求延迟 |
| http_requests_in_flight | Gauge | 当前处理中的请求数 |

### 2.2 业务指标

| 指标名 | 类型 | 说明 |
|--------|------|------|
| orders_synced_total | Counter | 同步订单总数 |
| orders_shipped_total | Counter | 发货订单总数 |
| finance_income_synced_total | Counter | 同步财务收入总数 |
| settlements_processed_total | Counter | 处理结算总数 |
| account_transactions_total | Counter | 账户交易总数 |
| withdraw_applications_total | Counter | 提现申请总数 |

### 2.3 系统指标

| 指标名 | 类型 | 说明 |
|--------|------|------|
| sync_tasks_queued | Gauge | 同步任务队列长度 |
| sync_tasks_processing | Gauge | 正在处理的同步任务数 |
| distributed_lock_acquired_total | Counter | 分布式锁获取成功数 |
| distributed_lock_failed_total | Counter | 分布式锁获取失败数 |
| archive_records_total | Counter | 归档记录总数 |
| shard_table_rows | Gauge | 各分表数据量 |

### 2.4 外部依赖指标

| 指标名 | 类型 | 说明 |
|--------|------|------|
| shopee_api_calls_total | Counter | Shopee API调用总数 |
| shopee_api_latency_seconds | Histogram | Shopee API延迟 |
| db_connections_active | Gauge | 数据库活跃连接数 |
| db_connections_idle | Gauge | 数据库空闲连接数 |
| redis_commands_total | Counter | Redis命令执行总数 |

## 3. 告警规则

### 3.1 服务可用性

| 告警名 | 条件 | 严重级别 |
|--------|------|----------|
| ServiceDown | 服务宕机 > 1分钟 | Critical |
| HighErrorRate | 5xx错误率 > 5% | Warning |
| HighLatency | P95延迟 > 2秒 | Warning |

### 3.2 数据库告警

| 告警名 | 条件 | 严重级别 |
|--------|------|----------|
| MySQLDown | MySQL宕机 > 1分钟 | Critical |
| MySQLConnectionsHigh | 连接数 > 100 | Warning |
| MySQLSlowQueries | 慢查询 > 1/秒 | Warning |
| DBConnectionPoolExhausted | 连接池使用率 > 90% | Warning |

### 3.3 Redis告警

| 告警名 | 条件 | 严重级别 |
|--------|------|----------|
| RedisDown | Redis宕机 > 1分钟 | Critical |
| RedisMemoryHigh | 内存使用率 > 80% | Warning |
| RedisConnectionsHigh | 连接数 > 500 | Warning |

### 3.4 业务告警

| 告警名 | 条件 | 严重级别 |
|--------|------|----------|
| SyncQueueBacklog | 队列积压 > 100 | Warning |
| ShopeeAPIErrorRate | API错误率 > 10% | Warning |
| ShopeeAPILatencyHigh | P95延迟 > 5秒 | Warning |
| DistributedLockFailure | 锁获取失败率 > 10% | Warning |
| ShardTableTooLarge | 分表数据量 > 500万 | Warning |

### 3.5 系统资源告警

| 告警名 | 条件 | 严重级别 |
|--------|------|----------|
| HighCPUUsage | CPU > 80% | Warning |
| HighMemoryUsage | 内存 > 85% | Warning |
| DiskSpaceLow | 磁盘 > 85% | Warning |

## 4. Grafana 仪表盘

### 4.1 概览仪表盘

- 服务状态（UP/DOWN）
- QPS 趋势图
- 错误率趋势图
- P95/P99 延迟趋势图

### 4.2 业务仪表盘

- 订单同步数量趋势
- 发货成功率
- 结算处理数量
- 提现申请数量

### 4.3 数据库仪表盘

- MySQL 连接数
- 查询延迟
- 慢查询数量
- 各分表数据量

### 4.4 Redis 仪表盘

- 内存使用量
- 命令执行数
- 连接数
- 队列长度

## 5. 部署配置

### 5.1 启动监控栈

```bash
cd deploy
docker-compose -f docker-compose.monitoring.yml up -d
```

### 5.2 访问地址

| 服务 | 地址 | 默认账号 |
|------|------|----------|
| Prometheus | http://localhost:9090 | - |
| Grafana | http://localhost:3000 | admin/admin123 |
| Alertmanager | http://localhost:9093 | - |

### 5.3 配置文件位置

| 文件 | 说明 |
|------|------|
| deploy/prometheus/prometheus.yml | Prometheus配置 |
| deploy/prometheus/rules/*.yml | 告警规则 |
| deploy/alertmanager/alertmanager.yml | 告警管理配置 |
| deploy/grafana/provisioning/ | Grafana数据源配置 |

## 6. 告警通知

### 6.1 通知渠道

| 渠道 | 配置 |
|------|------|
| 邮件 | SMTP配置 |
| 钉钉 | Webhook |
| 企业微信 | Webhook |
| Slack | Webhook |

### 6.2 告警分级

| 级别 | 通知方式 | 重复间隔 |
|------|----------|----------|
| Critical | 邮件+即时通讯 | 1小时 |
| Warning | 邮件 | 4小时 |

## 7. 运维操作

### 7.1 查看当前告警

```bash
# Prometheus
curl http://localhost:9090/api/v1/alerts

# Alertmanager
curl http://localhost:9093/api/v2/alerts
```

### 7.2 静默告警

```bash
# 创建静默
curl -X POST http://localhost:9093/api/v2/silences \
  -H "Content-Type: application/json" \
  -d '{
    "matchers": [{"name": "alertname", "value": "HighLatency"}],
    "startsAt": "2026-02-13T00:00:00Z",
    "endsAt": "2026-02-14T00:00:00Z",
    "createdBy": "admin",
    "comment": "维护期间静默"
  }'
```

### 7.3 重新加载配置

```bash
# Prometheus
curl -X POST http://localhost:9090/-/reload

# Alertmanager
curl -X POST http://localhost:9093/-/reload
```
