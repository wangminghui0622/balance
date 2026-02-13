# 备份策略

## 概述

本文档定义了 Balance 系统的数据备份策略，确保数据安全和业务连续性。

---

## 1. MySQL 备份策略

### 1.1 全量备份

| 项目 | 配置 |
|------|------|
| **频率** | 每天凌晨 1:00 |
| **保留期** | 30天 |
| **存储位置** | 本地 + 云存储（S3/OSS） |
| **工具** | mysqldump / xtrabackup |

```bash
#!/bin/bash
# 全量备份脚本: /opt/scripts/mysql_full_backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/data/backup/mysql"
DB_NAME="balance"
DB_USER="backup_user"
DB_PASS="your_password"

# 创建备份目录
mkdir -p ${BACKUP_DIR}

# 执行备份
mysqldump -u${DB_USER} -p${DB_PASS} \
  --single-transaction \
  --routines \
  --triggers \
  --databases ${DB_NAME} \
  | gzip > ${BACKUP_DIR}/${DB_NAME}_full_${DATE}.sql.gz

# 上传到云存储
aws s3 cp ${BACKUP_DIR}/${DB_NAME}_full_${DATE}.sql.gz \
  s3://your-bucket/mysql-backup/

# 清理30天前的备份
find ${BACKUP_DIR} -name "*.sql.gz" -mtime +30 -delete

echo "MySQL全量备份完成: ${DB_NAME}_full_${DATE}.sql.gz"
```

### 1.2 增量备份（binlog）

| 项目 | 配置 |
|------|------|
| **频率** | 实时 |
| **保留期** | 7天 |
| **配置** | my.cnf |

```ini
# /etc/mysql/my.cnf
[mysqld]
log-bin = mysql-bin
binlog_format = ROW
expire_logs_days = 7
max_binlog_size = 100M
sync_binlog = 1
```

### 1.3 分表备份注意事项

由于系统使用了120个分表，备份时需要确保：

1. **事务一致性**：使用 `--single-transaction` 保证备份一致性
2. **分表完整性**：确保所有分表都被备份
3. **存储过程**：使用 `--routines` 备份存储过程

---

## 2. Redis 备份策略

### 2.1 RDB 快照

| 项目 | 配置 |
|------|------|
| **频率** | 每小时 |
| **保留期** | 7天 |
| **配置** | redis.conf |

```ini
# /etc/redis/redis.conf
save 3600 1        # 1小时内有1个key变化则保存
save 300 100       # 5分钟内有100个key变化则保存
save 60 10000      # 1分钟内有10000个key变化则保存

dbfilename dump.rdb
dir /data/redis/
```

### 2.2 AOF 持久化

```ini
# /etc/redis/redis.conf
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
```

### 2.3 Redis 备份脚本

```bash
#!/bin/bash
# Redis备份脚本: /opt/scripts/redis_backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/data/backup/redis"
REDIS_DIR="/data/redis"

mkdir -p ${BACKUP_DIR}

# 触发RDB保存
redis-cli BGSAVE
sleep 10

# 复制RDB文件
cp ${REDIS_DIR}/dump.rdb ${BACKUP_DIR}/dump_${DATE}.rdb
gzip ${BACKUP_DIR}/dump_${DATE}.rdb

# 上传到云存储
aws s3 cp ${BACKUP_DIR}/dump_${DATE}.rdb.gz \
  s3://your-bucket/redis-backup/

# 清理7天前的备份
find ${BACKUP_DIR} -name "*.rdb.gz" -mtime +7 -delete

echo "Redis备份完成: dump_${DATE}.rdb.gz"
```

---

## 3. 定时任务配置

```bash
# /etc/crontab

# MySQL全量备份 - 每天凌晨1点
0 1 * * * root /opt/scripts/mysql_full_backup.sh >> /var/log/backup/mysql.log 2>&1

# Redis备份 - 每小时
0 * * * * root /opt/scripts/redis_backup.sh >> /var/log/backup/redis.log 2>&1

# 备份日志清理 - 每周日
0 0 * * 0 root find /var/log/backup -name "*.log" -mtime +30 -delete
```

---

## 4. 恢复流程

### 4.1 MySQL 恢复

```bash
# 1. 停止应用服务
systemctl stop balance-admin

# 2. 恢复全量备份
gunzip -c /data/backup/mysql/balance_full_20260213.sql.gz | mysql -uroot -p

# 3. 应用增量binlog（如需要）
mysqlbinlog mysql-bin.000001 mysql-bin.000002 | mysql -uroot -p

# 4. 验证数据
mysql -uroot -p -e "SELECT COUNT(*) FROM balance.shops;"

# 5. 重启应用
systemctl start balance-admin
```

### 4.2 Redis 恢复

```bash
# 1. 停止Redis
systemctl stop redis

# 2. 恢复RDB文件
cp /data/backup/redis/dump_20260213.rdb /data/redis/dump.rdb

# 3. 启动Redis
systemctl start redis

# 4. 验证数据
redis-cli DBSIZE
```

---

## 5. 监控告警

### 5.1 备份监控指标

| 指标 | 告警阈值 |
|------|---------|
| 备份文件大小 | 小于前一天的50% |
| 备份执行时间 | 超过2小时 |
| 备份失败次数 | 连续2次失败 |
| 存储空间使用率 | 超过80% |

### 5.2 告警配置（Prometheus）

```yaml
# prometheus/rules/backup.yml
groups:
  - name: backup_alerts
    rules:
      - alert: BackupFailed
        expr: backup_last_success_timestamp < (time() - 86400)
        for: 1h
        labels:
          severity: critical
        annotations:
          summary: "备份失败超过24小时"
          
      - alert: BackupStorageLow
        expr: backup_storage_usage_percent > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "备份存储空间不足"
```

---

## 6. 灾难恢复

### 6.1 RPO/RTO 目标

| 指标 | 目标 |
|------|------|
| **RPO (恢复点目标)** | < 1小时 |
| **RTO (恢复时间目标)** | < 4小时 |

### 6.2 灾难恢复步骤

1. **评估损失**：确定数据丢失范围
2. **准备环境**：启动备用服务器
3. **恢复数据**：按顺序恢复 MySQL → Redis
4. **验证数据**：检查关键业务数据完整性
5. **切换流量**：DNS切换或负载均衡调整
6. **监控观察**：持续监控系统状态

---

## 7. 备份验证

### 7.1 定期恢复测试

| 项目 | 频率 |
|------|------|
| 全量恢复测试 | 每月1次 |
| 增量恢复测试 | 每季度1次 |
| 灾难恢复演练 | 每年1次 |

### 7.2 验证脚本

```bash
#!/bin/bash
# 备份验证脚本: /opt/scripts/verify_backup.sh

# 在测试环境恢复最新备份
LATEST_BACKUP=$(ls -t /data/backup/mysql/*.sql.gz | head -1)

# 恢复到测试数据库
gunzip -c ${LATEST_BACKUP} | mysql -utest -ptest -h test-db

# 验证关键表数据
mysql -utest -ptest -h test-db -e "
  SELECT 'shops' as tbl, COUNT(*) as cnt FROM balance.shops
  UNION ALL
  SELECT 'orders_0', COUNT(*) FROM balance.orders_0
  UNION ALL
  SELECT 'admin', COUNT(*) FROM balance.admin;
"

echo "备份验证完成"
```

---

## 8. 安全要求

1. **加密存储**：备份文件使用 AES-256 加密
2. **访问控制**：备份目录权限 700，仅 root 可访问
3. **传输加密**：上传云存储使用 HTTPS/TLS
4. **审计日志**：记录所有备份/恢复操作
