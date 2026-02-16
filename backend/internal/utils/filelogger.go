package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewFileLogger 创建基于 zap 的文件日志记录器
// 特性：异步缓冲写入、日志轮转、30天自动清理
// logDir: 日志目录路径
// filename: 日志文件名（如 "distributed_sync.log"）
// 返回 *zap.SugaredLogger 和一个关闭函数（刷新缓冲区）
func NewFileLogger(logDir, filename string) (*zap.SugaredLogger, func(), error) {
	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	logPath := filepath.Join(logDir, filename)

	// lumberjack 日志轮转配置
	lumberLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100,  // 单个日志文件最大 100MB
		MaxBackups: 10,   // 最多保留 10 个备份文件
		MaxAge:     30,   // 日志文件保留 30 天
		Compress:   true, // 压缩旧日志文件
		LocalTime:  true, // 使用本地时间命名备份文件
	}

	// 自定义时间格式
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		MessageKey:    "msg",
		CallerKey:     "",
		EncodeTime:    zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05.000000"),
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeCaller:  zapcore.ShortCallerEncoder,
		LineEnding:    zapcore.DefaultLineEnding,
		ConsoleSeparator: " ",
	}

	// 使用 BufferedWriteSyncer 实现异步缓冲写入
	// 默认缓冲 256KB，每秒自动刷新一次
	bufferedFileWriter := &zapcore.BufferedWriteSyncer{
		WS:   zapcore.AddSync(lumberLogger),
		Size: 256 * 1024, // 256KB 缓冲区
	}

	// 同时写入文件和标准输出
	multiWriter := zapcore.NewMultiWriteSyncer(
		bufferedFileWriter,
		zapcore.AddSync(os.Stdout),
	)

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		multiWriter,
		zapcore.InfoLevel,
	)

	logger := zap.New(core)
	sugar := logger.Sugar()

	// 关闭函数：刷新缓冲区并关闭 lumberjack
	closeFunc := func() {
		_ = logger.Sync()
		_ = bufferedFileWriter.Stop()
		_ = lumberLogger.Close()
	}

	sugar.Infof("日志文件已创建: %s（轮转: 100MB/文件，保留30天，gzip压缩）", logPath)
	return sugar, closeFunc, nil
}

// DefaultSugaredLogger 返回默认的 SugaredLogger（输出到标准输出）
// 当文件日志创建失败时使用
func DefaultSugaredLogger() *zap.SugaredLogger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		MessageKey:    "msg",
		CallerKey:     "",
		EncodeTime:    zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05.000000"),
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		LineEnding:    zapcore.DefaultLineEnding,
		ConsoleSeparator: " ",
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	return zap.New(core).Sugar()
}
