# 部署指南

## 1. 环境要求

| 组件 | 版本要求 | 说明 |
|------|---------|------|
| Go | 1.21+ | 编译环境 |
| MySQL | 8.0+ | 主数据库 |
| Redis | 7.0+ | 缓存/分布式锁 |
| Nginx | 1.20+ | 反向代理/负载均衡 |
| Docker | 20.10+ | 容器化部署（可选） |

## 2. 单机部署

### 2.1 编译

```bash
cd backend/admin
go build -o balance-admin .
```

### 2.2 配置文件

```yaml
# /etc/balance/config.yaml
app:
  port: 8080
  mode: release

mysql:
  host: 127.0.0.1
  port: 3306
  user: balance
  password: your_password
  database: balance
  max_open_conns: 100
  max_idle_conns: 10

redis:
  host: 127.0.0.1
  port: 6379
  password: your_redis_password
  db: 0
  pool_size: 100

jwt:
  secret: your_jwt_secret_key_at_least_32_chars
  expire: 24h

shopee:
  partner_id: your_partner_id
  partner_key: your_partner_key
```

### 2.3 Systemd 服务

```ini
# /etc/systemd/system/balance-admin.service
[Unit]
Description=Balance Admin Service
After=network.target mysql.service redis.service

[Service]
Type=simple
User=balance
Group=balance
WorkingDirectory=/opt/balance
ExecStart=/opt/balance/balance-admin -config /etc/balance/config.yaml
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

```bash
# 启动服务
systemctl daemon-reload
systemctl enable balance-admin
systemctl start balance-admin
```

## 3. 多机分布式部署

### 3.1 架构图

```
                    ┌─────────────────┐
                    │   Nginx/SLB     │
                    │   (负载均衡)     │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
    ┌────▼────┐         ┌────▼────┐         ┌────▼────┐
    │ Admin-1 │         │ Admin-2 │         │ Admin-3 │
    │ :8080   │         │ :8080   │         │ :8080   │
    └────┬────┘         └────┬────┘         └────┬────┘
         │                   │                   │
         └───────────────────┼───────────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
         ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
         │  MySQL  │    │  Redis  │    │Prometheus│
         │  主从    │    │ Sentinel│    │ Grafana  │
         └─────────┘    └─────────┘    └──────────┘
```

### 3.2 Nginx 配置

```nginx
# /etc/nginx/conf.d/balance.conf
upstream balance_admin {
    least_conn;
    server 10.0.0.1:8080 weight=1;
    server 10.0.0.2:8080 weight=1;
    server 10.0.0.3:8080 weight=1;
    keepalive 32;
}

server {
    listen 443 ssl http2;
    server_name api.balance.com;

    ssl_certificate /etc/nginx/ssl/balance.crt;
    ssl_certificate_key /etc/nginx/ssl/balance.key;
    ssl_protocols TLSv1.2 TLSv1.3;

    location / {
        proxy_pass http://balance_admin;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection "";
        proxy_connect_timeout 30s;
        proxy_read_timeout 60s;
    }

    # Prometheus 指标（内网访问）
    location /metrics {
        allow 10.0.0.0/8;
        deny all;
        proxy_pass http://balance_admin;
    }
}
```

### 3.3 MySQL 主从配置

```ini
# 主库 /etc/mysql/my.cnf
[mysqld]
server-id = 1
log-bin = mysql-bin
binlog_format = ROW
sync_binlog = 1
innodb_flush_log_at_trx_commit = 1

# 从库 /etc/mysql/my.cnf
[mysqld]
server-id = 2
relay-log = relay-bin
read_only = 1
```

### 3.4 Redis Sentinel 配置

```ini
# /etc/redis/sentinel.conf
sentinel monitor mymaster 10.0.0.10 6379 2
sentinel down-after-milliseconds mymaster 5000
sentinel failover-timeout mymaster 60000
sentinel parallel-syncs mymaster 1
```

## 4. Docker 部署

### 4.1 Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o balance-admin ./admin

FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/balance-admin .

EXPOSE 8080
CMD ["./balance-admin", "-config", "/etc/balance/config.yaml"]
```

### 4.2 Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  balance-admin:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config:/etc/balance:ro
    environment:
      - TZ=Asia/Shanghai
    depends_on:
      - mysql
      - redis
    restart: unless-stopped
    deploy:
      replicas: 3

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root_password
      MYSQL_DATABASE: balance
    volumes:
      - mysql_data:/var/lib/mysql
      - ./database.sql:/docker-entrypoint-initdb.d/init.sql
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
```

## 5. Kubernetes 部署

### 5.1 Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: balance-admin
spec:
  replicas: 3
  selector:
    matchLabels:
      app: balance-admin
  template:
    metadata:
      labels:
        app: balance-admin
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: balance-admin
        image: balance-admin:latest
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /etc/balance
      volumes:
      - name: config
        configMap:
          name: balance-config
```

### 5.2 Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: balance-admin
spec:
  selector:
    app: balance-admin
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

## 6. 健康检查

| 端点 | 说明 |
|------|------|
| `GET /health` | 健康检查 |
| `GET /metrics` | Prometheus 指标 |

## 7. 日志配置

建议使用 ELK 或 Loki 收集日志：

```bash
# 日志输出到文件
./balance-admin -config config.yaml 2>&1 | tee -a /var/log/balance/admin.log

# 使用 logrotate 轮转
# /etc/logrotate.d/balance
/var/log/balance/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
}
```

## 8. 性能调优

### 8.1 系统参数

```bash
# /etc/sysctl.conf
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 30
fs.file-max = 1000000

# /etc/security/limits.conf
* soft nofile 65535
* hard nofile 65535
```

### 8.2 Go 运行时

```bash
# 设置 GOMAXPROCS
export GOMAXPROCS=4

# 设置 GC 目标
export GOGC=100
```
