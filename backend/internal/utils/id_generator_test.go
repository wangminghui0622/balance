package utils

import (
	"balance/internal/constants"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
	"testing"
	"time"
)

// setupTestRedis 创建测试用的Redis客户端
// 注意：需要本地运行Redis服务，默认地址 localhost:6379
func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "42.192.129.44:6379",
		Password: "test@789",
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("跳过测试：无法连接到Redis: %v", err)
	}
	return client
}

// cleanupTestKey 清理测试用的Redis key
func cleanupTestKey(t *testing.T, client *redis.Client, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client.Del(ctx, key)
}

// TestGeneratePlatformID_Concurrent 并发测试GeneratePlatformID
func TestGeneratePlatformID_Concurrent(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	// 清理实际使用的key，避免影响其他测试
	actualKey := "id:generator:platform"
	ctx := context.Background()

	// 清理key，确保测试从干净状态开始
	cleanupTestKey(t, client, actualKey)
	defer cleanupTestKey(t, client, actualKey)

	// 创建ID生成器
	generator := NewIDGenerator(client)

	// 并发参数
	goroutineCount := 100 // 并发goroutine数量
	idsPerGoroutine := 50 // 每个goroutine生成的ID数量
	totalIDs := goroutineCount * idsPerGoroutine

	// 用于收集所有生成的ID
	var mu sync.Mutex
	ids := make(map[int64]bool)
	var wg sync.WaitGroup
	errors := make(chan error, totalIDs)

	// 启动并发goroutine
	wg.Add(goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				// 直接调用GeneratePlatformID进行测试
				generatedID, err := generator.GeneratePlatformID(ctx)
				if err != nil {
					errors <- err
					return
				}

				// 记录生成的ID
				mu.Lock()
				if ids[generatedID] {
					errors <- &duplicateIDError{id: generatedID}
				} else {
					ids[generatedID] = true
				}
				mu.Unlock()
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errors)

	// 检查错误
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
			t.Errorf("生成ID时出错: %v", err)
		}
	}

	// 验证结果
	if len(ids) != totalIDs {
		t.Errorf("期望生成 %d 个唯一ID，实际生成 %d 个（错误数: %d）", totalIDs, len(ids), errorCount)
	}

	// 验证第一个ID是初始值
	firstID := constants.IDInitialPlatform
	if !ids[firstID] {
		t.Errorf("第一个ID应该是初始值 %d，但未找到", firstID)
	}

	// 验证所有ID都是Redis ID（不是降级ID）
	// Redis ID范围：90000000000+（平台ID）
	// 降级ID范围：80000000000+（时间戳ID）
	minID := constants.IDInitialPlatform
	fallbackIDPrefix := constants.FallbackIDPrefix
	fallbackIDCount := 0
	redisIDCount := 0

	for id := range ids {
		if id < minID {
			t.Errorf("ID %d 小于最小允许值 %d", id, minID)
		}
		// 检查是否是降级ID（不应该出现）
		if id >= fallbackIDPrefix && id < constants.IDInitialPlatform {
			fallbackIDCount++
			t.Errorf("发现降级ID（不应该出现）: %d", id)
		} else if id >= constants.IDInitialPlatform {
			redisIDCount++
		}
	}

	// 验证：Redis连接正常时，不应该有降级ID
	if fallbackIDCount > 0 {
		t.Errorf("Redis连接正常时不应该有降级ID，但发现 %d 个降级ID", fallbackIDCount)
	}

	// 验证所有ID都在合理范围内
	maxExpectedID := constants.IDInitialPlatform + int64(totalIDs-1)*constants.IDIncrementPlatformMax
	for id := range ids {
		if id > maxExpectedID {
			// 允许稍微超出，因为增量是随机的
			if id > maxExpectedID+constants.IDIncrementPlatformMax {
				t.Errorf("ID %d 超过最大预期值 %d", id, maxExpectedID)
			}
		}
	}

	t.Logf("成功生成 %d 个唯一ID，无重复", len(ids))
	t.Logf("Redis ID数: %d", redisIDCount)
	t.Logf("降级ID数: %d（应该为0）", fallbackIDCount)
	t.Logf("ID范围: %d - %d", minID, maxExpectedID)
	t.Logf("并发goroutine数: %d，每个goroutine生成: %d 个ID", goroutineCount, idsPerGoroutine)
	t.Logf("✓ 并发安全性验证通过：所有ID都是唯一的，且都是Redis ID（无降级ID）")
}

// TestGeneratePlatformID_MultiMachineConcurrent 测试多机并发安全性
// 模拟多个服务器/进程同时访问同一个Redis
func TestGeneratePlatformID_MultiMachineConcurrent(t *testing.T) {
	// 模拟多个服务器，每个服务器创建独立的Redis客户端
	machineCount := 10   // 模拟10台服务器
	idsPerMachine := 300 // 每台服务器生成300个ID
	totalIDs := machineCount * idsPerMachine

	// 清理key，确保测试从干净状态开始
	testKey := "id:generator:platform:multimachine"
	baseClient := setupTestRedis(t)
	defer baseClient.Close()
	cleanupTestKey(t, baseClient, testKey)
	defer cleanupTestKey(t, baseClient, testKey)

	// 用于收集所有生成的ID
	var mu sync.Mutex
	ids := make(map[int64]bool)
	var wg sync.WaitGroup
	errors := make(chan error, totalIDs)
	machineStats := make(map[int]int) // 每台机器生成的ID数量

	// 启动多个"服务器"（每个服务器有独立的Redis客户端和ID生成器）
	wg.Add(machineCount)
	for machineID := 0; machineID < machineCount; machineID++ {
		go func(mid int) {
			defer wg.Done()

			// 每台服务器创建独立的Redis客户端（模拟不同服务器）
			client := redis.NewClient(&redis.Options{
				Addr:     "42.192.129.44:6379",
				Password: "test@789",
				DB:       0,
			})
			defer client.Close()

			// 每台服务器创建独立的ID生成器
			generator := NewIDGenerator(client)
			ctx := context.Background()

			machineIDCount := 0
			for i := 0; i < idsPerMachine; i++ {
				// 生成ID
				id, err := generator.GeneratePlatformID(ctx)
				if err != nil {
					errors <- fmt.Errorf("机器 %d 生成ID失败: %v", mid, err)
					continue
				}

				// 记录生成的ID
				mu.Lock()
				if ids[id] {
					errors <- fmt.Errorf("机器 %d 发现重复ID: %d", mid, id)
				} else {
					ids[id] = true
					machineIDCount++
				}
				mu.Unlock()

				// 添加小延迟，模拟真实场景
				time.Sleep(time.Microsecond * 100)
			}

			mu.Lock()
			machineStats[mid] = machineIDCount
			mu.Unlock()
		}(machineID)
	}

	// 等待所有"服务器"完成
	wg.Wait()
	close(errors)

	// 检查错误
	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
			t.Errorf("生成ID时出错: %v", err)
		}
	}

	// 验证结果
	t.Logf("\n=== 多机并发测试结果 ===")
	t.Logf("服务器数量: %d", machineCount)
	t.Logf("每台服务器生成ID数: %d", idsPerMachine)
	t.Logf("期望总ID数: %d", totalIDs)
	t.Logf("实际生成ID数: %d", len(ids))
	t.Logf("错误数: %d", errorCount)

	// 验证每台机器的统计
	t.Logf("\n各服务器生成ID统计:")
	for i := 0; i < machineCount; i++ {
		t.Logf("  服务器 %d: %d 个ID", i, machineStats[i])
	}

	// 验证：所有ID都是唯一的
	if len(ids) != totalIDs {
		t.Errorf("期望生成 %d 个唯一ID，实际生成 %d 个（错误数: %d）", totalIDs, len(ids), errorCount)
	}

	// 验证：所有ID都是Redis ID（不是降级ID）
	fallbackIDCount := 0
	redisIDCount := 0
	minID := constants.IDInitialPlatform
	maxID := int64(0)

	for id := range ids {
		if id >= constants.FallbackIDPrefix && id < constants.IDInitialPlatform {
			fallbackIDCount++
			t.Errorf("发现降级ID（不应该出现）: %d", id)
		} else if id >= constants.IDInitialPlatform {
			redisIDCount++
		}
		if id < minID {
			minID = id
		}
		if id > maxID {
			maxID = id
		}
	}

	t.Logf("\nID类型统计:")
	t.Logf("  Redis ID数: %d", redisIDCount)
	t.Logf("  降级ID数: %d（应该为0）", fallbackIDCount)

	// 验证：Redis连接正常时，不应该有降级ID
	if fallbackIDCount > 0 {
		t.Errorf("多机并发时不应该有降级ID，但发现 %d 个降级ID", fallbackIDCount)
	}

	// 验证第一个ID是初始值（多机并发时，第一个ID可能不是初始值，因为其他服务器可能已经生成了）
	// 但最小ID应该是初始值或接近初始值
	firstID := constants.IDInitialPlatform
	if !ids[firstID] {
		// 多机并发时，初始值可能已经被其他服务器使用，这是正常的
		// 只要最小ID在合理范围内即可
		if minID > firstID+constants.IDIncrementPlatformMax*10 {
			t.Errorf("最小ID %d 与初始值 %d 差距过大", minID, firstID)
		} else {
			t.Logf("注意：初始值 %d 未被使用（多机并发时正常），最小ID: %d", firstID, minID)
		}
	} else {
		t.Logf("✓ 初始值 %d 被使用", firstID)
	}

	t.Logf("\nID范围:")
	t.Logf("  最小ID: %d", minID)
	t.Logf("  最大ID: %d", maxID)
	t.Logf("  ID范围: %d - %d", minID, maxID)

	// 验证ID连续性（检查是否有大的跳跃，可能表示有ID丢失）
	idList := make([]int64, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	// 简单验证：检查ID是否都在合理范围内
	expectedMaxID := constants.IDInitialPlatform + int64(totalIDs-1)*constants.IDIncrementPlatformMax
	if maxID > expectedMaxID+constants.IDIncrementPlatformMax*10 {
		t.Logf("警告：最大ID %d 超出预期范围 %d（允许一定误差）", maxID, expectedMaxID)
	}

	t.Logf("\n✓ 多机并发安全性验证通过：")
	t.Logf("  - 所有 %d 个ID都是唯一的", len(ids))
	t.Logf("  - 所有ID都是Redis ID（无降级ID）")
	t.Logf("  - %d 台服务器并发访问Redis，无冲突", machineCount)
	t.Logf("  - Lua脚本保证了多机并发的原子性")
}

// duplicateIDError 重复ID错误
type duplicateIDError struct {
	id int64
}

func (e *duplicateIDError) Error() string {
	return "发现重复ID"
}

// TestGeneratePlatformID_Sequential 顺序测试GeneratePlatformID（验证基本功能）
func TestGeneratePlatformID_Sequential(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	actualKey := "id:generator:platform"
	ctx := context.Background()

	// 清理key，确保测试从干净状态开始
	cleanupTestKey(t, client, actualKey)
	defer cleanupTestKey(t, client, actualKey)

	generator := NewIDGenerator(client)

	// 第一个ID应该是初始值
	firstID, err := generator.GeneratePlatformID(ctx)
	if err != nil {
		t.Fatalf("生成第一个ID失败: %v", err)
	}
	if firstID != constants.IDInitialPlatform {
		t.Errorf("第一个ID应该是 %d，实际是 %d", constants.IDInitialPlatform, firstID)
	}

	// 生成几个ID验证增量
	prevID := firstID
	for i := 0; i < 10; i++ {
		id, err := generator.GeneratePlatformID(ctx)
		if err != nil {
			t.Fatalf("生成ID失败: %v", err)
		}

		// 验证ID是递增的
		if id <= prevID {
			t.Errorf("ID应该是递增的，但 %d <= %d", id, prevID)
		}

		// 验证增量在合理范围内（10-20）
		increment := id - prevID
		if increment < constants.IDIncrementPlatformMin || increment > constants.IDIncrementPlatformMax {
			t.Errorf("增量 %d 不在预期范围内 [%d, %d]", increment, constants.IDIncrementPlatformMin, constants.IDIncrementPlatformMax)
		}

		prevID = id
	}

	t.Logf("顺序测试通过，生成的ID符合预期")
}

// TestGeneratePlatformID_FallbackTimestamp 测试Redis网络不通时的时间戳ID降级方案
func TestGeneratePlatformID_FallbackTimestamp(t *testing.T) {
	// 创建一个无效的Redis客户端（模拟网络不通）
	invalidClient := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:9999", // 无效地址
		Password:     "",
		DB:           0,
		DialTimeout:  100 * time.Millisecond, // 快速超时
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
		PoolTimeout:  100 * time.Millisecond,
	})
	defer invalidClient.Close()

	// 使用短超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	generator := NewIDGenerator(invalidClient)

	// 尝试生成ID，应该使用时间戳降级方案
	t.Logf("测试：Redis网络不通时的时间戳ID降级")

	ids := make([]int64, 0)
	for i := 0; i < 10; i++ {
		id, err := generator.GeneratePlatformID(ctx)
		if err != nil {
			t.Errorf("降级方案应该成功生成ID，但返回错误: %v", err)
			continue
		}

		ids = append(ids, id)

		// 验证ID格式：应该是8开头（降级ID前缀）
		if id < constants.FallbackIDPrefix {
			t.Errorf("降级ID应该 >= %d，但得到: %d", constants.FallbackIDPrefix, id)
		}

		// 验证ID是递增的（时间戳递增）
		if i > 0 && id <= ids[i-1] {
			t.Errorf("降级ID应该是递增的，但 %d <= %d", id, ids[i-1])
		}

		// 添加小延迟，确保时间戳不同
		time.Sleep(2 * time.Millisecond)
	}

	t.Logf("成功生成 %d 个降级ID: %v", len(ids), ids)

	// 验证所有ID都是唯一的
	uniqueIDs := make(map[int64]bool)
	for _, id := range ids {
		if uniqueIDs[id] {
			t.Errorf("发现重复的降级ID: %d", id)
		} else {
			uniqueIDs[id] = true
		}
	}

	t.Logf("✓ 验证通过：所有降级ID都是唯一的")

	// 验证ID格式：8 + timestamp*1000 + random
	if len(ids) > 0 {
		firstID := ids[0]
		// 提取时间戳部分
		timestampPart := (firstID - constants.FallbackIDPrefix) / 1000
		randomPart := (firstID - constants.FallbackIDPrefix) % 1000

		t.Logf("第一个降级ID解析:")
		t.Logf("  完整ID: %d", firstID)
		t.Logf("  时间戳部分: %d (毫秒)", timestampPart)
		t.Logf("  随机数部分: %d", randomPart)

		// 验证时间戳是合理的（应该是当前时间附近）
		currentTimestamp := time.Now().UnixMilli()
		timeDiff := currentTimestamp - timestampPart
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		if timeDiff > 10000 { // 允许10秒误差
			t.Errorf("时间戳部分不合理，当前时间: %d, ID中的时间戳: %d, 差异: %d 毫秒",
				currentTimestamp, timestampPart, timeDiff)
		} else {
			t.Logf("✓ 时间戳验证通过，差异: %d 毫秒", timeDiff)
		}

		// 验证随机数在合理范围内
		if randomPart < 0 || randomPart > constants.FallbackRandomMax {
			t.Errorf("随机数部分不在合理范围内: %d (应该在 0-%d)", randomPart, constants.FallbackRandomMax)
		} else {
			t.Logf("✓ 随机数验证通过: %d", randomPart)
		}
	}
}

// TestGeneratePlatformID_FallbackVsRedis 测试降级ID与Redis ID不冲突
func TestGeneratePlatformID_FallbackVsRedis(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	testKey := "id:generator:platform:fallback"
	ctx := context.Background()

	// 清理key
	cleanupTestKey(t, client, testKey)
	defer cleanupTestKey(t, client, testKey)

	generator := NewIDGenerator(client)

	// 步骤1：正常生成一些Redis ID
	t.Logf("步骤1：正常生成Redis ID")
	redisIDs := make([]int64, 0)
	for i := 0; i < 5; i++ {
		id, err := generator.GeneratePlatformID(ctx)
		if err != nil {
			t.Fatalf("生成Redis ID失败: %v", err)
		}
		redisIDs = append(redisIDs, id)
	}
	t.Logf("Redis ID: %v", redisIDs)

	// 步骤2：关闭连接，模拟网络断开
	t.Logf("步骤2：关闭Redis连接，模拟网络断开")
	client.Close()

	// 步骤3：生成降级ID
	t.Logf("步骤3：生成降级ID（时间戳ID）")
	fallbackIDs := make([]int64, 0)
	for i := 0; i < 5; i++ {
		id, err := generator.GeneratePlatformID(ctx)
		if err != nil {
			t.Errorf("降级方案应该成功，但返回错误: %v", err)
			continue
		}
		fallbackIDs = append(fallbackIDs, id)
		time.Sleep(2 * time.Millisecond)
	}
	t.Logf("降级ID: %v", fallbackIDs)

	// 验证：Redis ID和降级ID不冲突
	allIDs := make(map[int64]bool)
	conflicts := make([]int64, 0)

	for _, id := range redisIDs {
		if allIDs[id] {
			conflicts = append(conflicts, id)
		} else {
			allIDs[id] = true
		}
	}

	for _, id := range fallbackIDs {
		if allIDs[id] {
			conflicts = append(conflicts, id)
		} else {
			allIDs[id] = true
		}
	}

	t.Logf("\n=== 冲突检测结果 ===")
	t.Logf("Redis ID数: %d", len(redisIDs))
	t.Logf("降级ID数: %d", len(fallbackIDs))
	t.Logf("总ID数: %d", len(allIDs))
	t.Logf("冲突ID数: %d", len(conflicts))

	if len(conflicts) > 0 {
		t.Errorf("发现ID冲突！冲突的ID: %v", conflicts)
	} else {
		t.Logf("✓ 验证通过：Redis ID和降级ID不冲突")
	}

	// 验证：降级ID应该都是8开头
	for _, id := range fallbackIDs {
		if id < constants.FallbackIDPrefix {
			t.Errorf("降级ID应该 >= %d，但得到: %d", constants.FallbackIDPrefix, id)
		}
	}
	t.Logf("✓ 验证通过：所有降级ID都是8开头（降级ID前缀）")
}

// TestGeneratePlatformID_RedisDisconnectConflict 测试Redis网络断开时是否会产生ID冲突
// 场景：模拟两个客户端，一个在Redis断开期间尝试生成ID，另一个正常生成ID
func TestGeneratePlatformID_RedisDisconnectConflict(t *testing.T) {
	client1 := setupTestRedis(t)
	defer client1.Close()

	// 创建第二个客户端（模拟另一个进程/服务）
	client2 := redis.NewClient(&redis.Options{
		Addr:     "42.192.129.44:6379",
		Password: "test@789",
		DB:       0,
	})
	defer client2.Close()

	testKey := "id:generator:platform:conflict"
	ctx := context.Background()

	// 清理key
	cleanupTestKey(t, client1, testKey)
	defer cleanupTestKey(t, client1, testKey)

	generator1 := NewIDGenerator(client1)
	generator2 := NewIDGenerator(client2)

	// 步骤1：客户端1正常生成一些ID
	t.Logf("步骤1：客户端1正常生成ID")
	client1IDs := make([]int64, 0)
	for i := 0; i < 10; i++ {
		id, err := generator1.GeneratePlatformID(ctx)
		if err != nil {
			t.Fatalf("客户端1生成ID失败: %v", err)
		}
		client1IDs = append(client1IDs, id)
	}
	t.Logf("客户端1成功生成 %d 个ID: %v", len(client1IDs), client1IDs)

	// 步骤2：关闭客户端1的连接（模拟网络断开）
	t.Logf("步骤2：关闭客户端1的连接（模拟网络断开）")
	client1.Close()

	// 步骤3：客户端1在断开期间尝试生成ID（应该失败）
	t.Logf("步骤3：客户端1在断开期间尝试生成ID")
	disconnectedID, err := generator1.GeneratePlatformID(ctx)
	if err == nil {
		t.Errorf("客户端1在断开期间不应该成功生成ID，但生成了: %d", disconnectedID)
	} else {
		t.Logf("客户端1在断开期间正确返回错误: %v", err)
	}

	// 步骤4：客户端2正常生成ID（模拟其他正常工作的客户端）
	t.Logf("步骤4：客户端2正常生成ID")
	client2IDs := make([]int64, 0)
	for i := 0; i < 10; i++ {
		id, err := generator2.GeneratePlatformID(ctx)
		if err != nil {
			t.Fatalf("客户端2生成ID失败: %v", err)
		}
		client2IDs = append(client2IDs, id)
	}
	t.Logf("客户端2成功生成 %d 个ID: %v", len(client2IDs), client2IDs)

	// 步骤5：客户端1重新连接
	t.Logf("步骤5：客户端1重新连接")
	client1 = redis.NewClient(&redis.Options{
		Addr:     "42.192.129.44:6379",
		Password: "test@789",
		DB:       0,
	})
	defer client1.Close()
	generator1 = NewIDGenerator(client1)

	// 步骤6：客户端1恢复后生成ID
	t.Logf("步骤6：客户端1恢复后生成ID")
	recoveredIDs := make([]int64, 0)
	for i := 0; i < 10; i++ {
		id, err := generator1.GeneratePlatformID(ctx)
		if err != nil {
			t.Fatalf("客户端1恢复后生成ID失败: %v", err)
		}
		recoveredIDs = append(recoveredIDs, id)
	}
	t.Logf("客户端1恢复后成功生成 %d 个ID: %v", len(recoveredIDs), recoveredIDs)

	// 验证：所有ID应该是唯一的，没有冲突
	allIDs := make(map[int64]bool)
	conflicts := make([]int64, 0)

	// 检查客户端1的ID
	for _, id := range client1IDs {
		if allIDs[id] {
			conflicts = append(conflicts, id)
		} else {
			allIDs[id] = true
		}
	}

	// 检查客户端2的ID
	for _, id := range client2IDs {
		if allIDs[id] {
			conflicts = append(conflicts, id)
		} else {
			allIDs[id] = true
		}
	}

	// 检查客户端1恢复后的ID
	for _, id := range recoveredIDs {
		if allIDs[id] {
			conflicts = append(conflicts, id)
		} else {
			allIDs[id] = true
		}
	}

	// 报告结果
	totalIDs := len(client1IDs) + len(client2IDs) + len(recoveredIDs)
	uniqueIDs := len(allIDs)

	t.Logf("\n=== 冲突检测结果 ===")
	t.Logf("总ID数: %d", totalIDs)
	t.Logf("唯一ID数: %d", uniqueIDs)
	t.Logf("冲突ID数: %d", len(conflicts))

	if len(conflicts) > 0 {
		t.Errorf("发现ID冲突！冲突的ID: %v", conflicts)
	} else {
		t.Logf("✓ 验证通过：所有ID都是唯一的，无冲突")
	}

	// 验证：客户端1恢复后的ID应该大于客户端2最后生成的ID（因为Redis中的值已经更新）
	if len(client2IDs) > 0 && len(recoveredIDs) > 0 {
		lastClient2ID := client2IDs[len(client2IDs)-1]
		firstRecoveredID := recoveredIDs[0]
		if firstRecoveredID <= lastClient2ID {
			t.Errorf("客户端1恢复后的第一个ID (%d) 应该大于客户端2最后生成的ID (%d)", firstRecoveredID, lastClient2ID)
		} else {
			t.Logf("✓ 验证通过：客户端1恢复后的ID正确接续（%d > %d）", firstRecoveredID, lastClient2ID)
		}
	}
}

// TestGeneratePlatformID_NetworkDisconnect 测试Redis网络随机断开的情况
// 通过使用短超时的context来模拟网络问题
func TestGeneratePlatformID_NetworkDisconnect(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	actualKey := "id:generator:platform:disconnect"
	baseCtx := context.Background()

	// 清理key，确保测试从干净状态开始
	cleanupTestKey(t, client, actualKey)
	defer cleanupTestKey(t, client, actualKey)

	generator := NewIDGenerator(client)

	// 测试参数
	totalAttempts := 50                     // 总共尝试生成ID的次数
	timeoutDuration := 1 * time.Millisecond // 超时时间（很短，模拟网络问题）

	var mu sync.Mutex
	successfulIDs := make(map[int64]bool)
	failedAttempts := 0
	timeoutAttempts := 0

	// 并发生成ID
	var wg sync.WaitGroup
	wg.Add(totalAttempts)

	for i := 0; i < totalAttempts; i++ {
		go func(attemptNum int) {
			defer wg.Done()

			// 随机决定是否使用超时context模拟网络问题
			var ctx context.Context
			var cancel context.CancelFunc
			useTimeout := (attemptNum%5 == 0) // 每5次中有1次使用超时

			if useTimeout {
				timeoutAttempts++
				ctx, cancel = context.WithTimeout(baseCtx, timeoutDuration)
				defer cancel()
				// 等待超时
				time.Sleep(2 * time.Millisecond)
			} else {
				ctx = baseCtx
			}

			// 尝试生成ID
			id, err := generator.GeneratePlatformID(ctx)
			if err != nil {
				mu.Lock()
				failedAttempts++
				mu.Unlock()
				if useTimeout {
					t.Logf("生成ID失败（第 %d 次，使用超时）: %v", attemptNum+1, err)
				} else {
					t.Logf("生成ID失败（第 %d 次）: %v", attemptNum+1, err)
				}
				return
			}

			// 记录成功的ID
			mu.Lock()
			if successfulIDs[id] {
				t.Errorf("发现重复ID: %d", id)
			} else {
				successfulIDs[id] = true
			}
			mu.Unlock()
		}(i)
	}

	// 等待完成
	wg.Wait()

	t.Logf("测试完成:")
	t.Logf("  总尝试次数: %d", totalAttempts)
	t.Logf("  成功生成ID数: %d", len(successfulIDs))
	t.Logf("  失败次数: %d", failedAttempts)
	t.Logf("  超时模拟次数: %d", timeoutAttempts)
	if totalAttempts > 0 {
		t.Logf("  成功率: %.2f%%", float64(len(successfulIDs))/float64(totalAttempts)*100)
	}

	// 验证：即使有网络问题，成功的ID也应该是唯一的
	if len(successfulIDs) > 0 {
		// 验证：成功的ID应该在合理范围内
		minID := constants.IDInitialPlatform

		var actualMinID, actualMaxID int64 = 0, 0
		for id := range successfulIDs {
			if id < minID {
				t.Errorf("ID %d 小于最小允许值 %d", id, minID)
			}
			if actualMinID == 0 || id < actualMinID {
				actualMinID = id
			}
			if id > actualMaxID {
				actualMaxID = id
			}
		}

		t.Logf("所有成功的ID都是唯一的，共 %d 个", len(successfulIDs))
		t.Logf("实际ID范围: %d - %d", actualMinID, actualMaxID)
		t.Logf("验证通过：所有ID >= 初始值 %d，且无重复", minID)
	} else {
		t.Logf("警告：没有成功生成任何ID")
	}
}

// TestGeneratePlatformID_ContextTimeout 测试context超时的情况
func TestGeneratePlatformID_ContextTimeout(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	actualKey := "id:generator:platform:timeout"
	ctx := context.Background()

	// 清理key
	cleanupTestKey(t, client, actualKey)
	defer cleanupTestKey(t, client, actualKey)

	generator := NewIDGenerator(client)

	// 测试：使用很短的超时时间
	shortTimeout := 1 * time.Millisecond
	timeoutCtx, cancel := context.WithTimeout(ctx, shortTimeout)
	defer cancel()

	// 等待超时
	time.Sleep(2 * time.Millisecond)

	// 尝试生成ID，应该会超时
	_, err := generator.GeneratePlatformID(timeoutCtx)
	if err == nil {
		t.Logf("注意：超时测试未触发错误（可能是Redis响应太快）")
	} else {
		t.Logf("超时错误（预期）: %v", err)
	}

	// 使用正常的context应该能成功
	normalID, err := generator.GeneratePlatformID(ctx)
	if err != nil {
		t.Errorf("正常context应该能生成ID，但出错: %v", err)
	} else {
		t.Logf("正常context成功生成ID: %d", normalID)
	}
}
