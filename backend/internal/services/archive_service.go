package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"balance/backend/internal/database"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	archiveSchedulerLock = "archive:scheduler:lock"
	archiveLockTTL       = 10 * time.Minute
	
	// 归档配置
	OperationLogRetentionDays = 90  // 操作日志保留天数
	ArchiveBatchSize          = 1000 // 每批归档数量
)

// ArchiveService 归档服务
type ArchiveService struct {
	db  *gorm.DB
	rdb *redis.Client
}

// NewArchiveService 创建归档服务
func NewArchiveService() *ArchiveService {
	return &ArchiveService{
		db:  database.GetDB(),
		rdb: database.GetRedis(),
	}
}

// ArchiveOperationLogs 归档操作日志（将过期数据移动到归档表后删除）
func (s *ArchiveService) ArchiveOperationLogs(ctx context.Context) (int64, error) {
	// 获取分布式锁
	ok, err := s.rdb.SetNX(ctx, archiveSchedulerLock, "1", archiveLockTTL).Result()
	if err != nil {
		return 0, fmt.Errorf("获取归档锁失败: %w", err)
	}
	if !ok {
		log.Println("[Archive] 其他节点正在归档，跳过")
		return 0, nil
	}
	defer s.rdb.Del(ctx, archiveSchedulerLock)

	log.Println("[Archive] 开始归档操作日志...")

	cutoffTime := time.Now().AddDate(0, 0, -OperationLogRetentionDays)
	var totalArchived int64

	// 遍历所有分表
	for i := 0; i < database.ShardCount; i++ {
		sourceTable := fmt.Sprintf("operation_logs_%d", i)
		archiveTable := fmt.Sprintf("operation_logs_archive_%d", i)

		archived, err := s.archiveTableData(ctx, sourceTable, archiveTable, cutoffTime)
		if err != nil {
			log.Printf("[Archive] 归档 %s 失败: %v", sourceTable, err)
			continue
		}
		totalArchived += archived
		if archived > 0 {
			log.Printf("[Archive] 归档 %s 完成，移动 %d 条记录", sourceTable, archived)
		}
	}

	log.Printf("[Archive] 操作日志归档完成，共归档 %d 条记录", totalArchived)
	return totalArchived, nil
}

// archiveTableData 归档单个表的数据
func (s *ArchiveService) archiveTableData(ctx context.Context, sourceTable, archiveTable string, cutoffTime time.Time) (int64, error) {
	var totalMoved int64

	for {
		// 分批处理，避免长事务
		var count int64
		err := s.db.Transaction(func(tx *gorm.DB) error {
			// 1. 将数据插入归档表
			insertSQL := fmt.Sprintf(`
				INSERT INTO %s 
				SELECT * FROM %s 
				WHERE created_at < ? 
				ORDER BY id 
				LIMIT %d
			`, archiveTable, sourceTable, ArchiveBatchSize)

			result := tx.Exec(insertSQL, cutoffTime)
			if result.Error != nil {
				return result.Error
			}
			count = result.RowsAffected

			if count == 0 {
				return nil
			}

			// 2. 删除已归档的数据
			deleteSQL := fmt.Sprintf(`
				DELETE FROM %s 
				WHERE created_at < ? 
				ORDER BY id 
				LIMIT %d
			`, sourceTable, ArchiveBatchSize)

			if err := tx.Exec(deleteSQL, cutoffTime).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return totalMoved, err
		}

		totalMoved += count

		// 没有更多数据需要归档
		if count < ArchiveBatchSize {
			break
		}

		// 短暂休息，避免对数据库造成压力
		time.Sleep(100 * time.Millisecond)
	}

	return totalMoved, nil
}

// CleanupOldArchives 清理过期的归档数据（可选，保留1年）
func (s *ArchiveService) CleanupOldArchives(ctx context.Context, retentionDays int) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	var totalDeleted int64

	for i := 0; i < database.ShardCount; i++ {
		archiveTable := fmt.Sprintf("operation_logs_archive_%d", i)

		result := s.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE created_at < ?", archiveTable), cutoffTime)
		if result.Error != nil {
			log.Printf("[Archive] 清理 %s 失败: %v", archiveTable, result.Error)
			continue
		}
		totalDeleted += result.RowsAffected
	}

	log.Printf("[Archive] 清理过期归档完成，共删除 %d 条记录", totalDeleted)
	return totalDeleted, nil
}

// GetArchiveStats 获取归档统计
func (s *ArchiveService) GetArchiveStats(ctx context.Context) map[string]interface{} {
	var activeCount, archiveCount int64

	// 统计活跃表数据量
	for i := 0; i < database.ShardCount; i++ {
		var count int64
		s.db.Table(fmt.Sprintf("operation_logs_%d", i)).Count(&count)
		activeCount += count
	}

	// 统计归档表数据量
	for i := 0; i < database.ShardCount; i++ {
		var count int64
		s.db.Table(fmt.Sprintf("operation_logs_archive_%d", i)).Count(&count)
		archiveCount += count
	}

	return map[string]interface{}{
		"active_logs":   activeCount,
		"archived_logs": archiveCount,
		"retention_days": OperationLogRetentionDays,
	}
}
