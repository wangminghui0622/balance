package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// NewFileLogger 创建文件日志记录器，同时输出到文件和标准输出
// logDir: 日志目录路径
// filename: 日志文件名（如 "distributed_sync.log"）
// prefix: 日志前缀（如 "[DistributedSync] "）
// 返回 *log.Logger 和一个关闭函数
func NewFileLogger(logDir, filename, prefix string) (*log.Logger, func(), error) {
	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	logPath := filepath.Join(logDir, filename)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 同时写入文件和标准输出
	multiWriter := io.MultiWriter(os.Stdout, file)
	logger := log.New(multiWriter, prefix, log.LstdFlags|log.Lmicroseconds)

	closeFunc := func() {
		file.Close()
	}

	log.Printf("日志文件已创建: %s", logPath)
	return logger, closeFunc, nil
}
