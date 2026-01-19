package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultJWTExpiration JWT默认过期时间：15天
const DefaultJWTExpiration = 15 * 24 * time.Hour

// DefaultJWTExpirationStr JWT默认过期时间字符串：15d
const DefaultJWTExpirationStr = "15d"

// Config 应用运行时配置（给代码用）
type Config struct {
	DBDSN         string
	RedisAddr     string
	RedisPassword string
	JWTSecret     string
	JWTExpiration time.Duration
	Port          string
}

// fileConfig 对应 config.yaml 结构
type fileConfig struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		Charset  string `yaml:"charset"`
	} `yaml:"database"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Password string `yaml:"password"`
	} `yaml:"redis"`
	JWT struct {
		Secret     string `yaml:"secret"`
		Expiration string `yaml:"expiration"`
	} `yaml:"jwt"`
	Services struct {
		Admin struct {
			Port string `yaml:"port"`
		} `yaml:"admin"`
		App struct {
			Port string `yaml:"port"`
		} `yaml:"app"`
	} `yaml:"services"`
}

var (
	cfgOnce      sync.Once
	loadedConfig *fileConfig
)

// loadFileConfig 读取 backend/config.yaml，如果不存在则返回空配置
func loadFileConfig() *fileConfig {
	cfgOnce.Do(func() {
		loadedConfig = &fileConfig{}

		// 尝试多个可能的配置文件路径
		configPaths := []string{
			"config.yaml",                    // 当前目录
			"backend/config.yaml",            // 从项目根目录
			"../config.yaml",                 // 上一级目录
		}

		var data []byte
		var err error
		for _, path := range configPaths {
			data, err = os.ReadFile(path)
			if err == nil {
				break
			}
		}

		if err != nil {
			// 没有配置文件则使用空配置，后续走默认值
			return
		}
		_ = yaml.Unmarshal(data, loadedConfig)
	})
	return loadedConfig
}

// LoadAdminConfig 加载 admin 服务配置：
// 优先：环境变量 -> 配置文件 config.yaml -> 代码默认值
func LoadAdminConfig() *Config {
	fc := loadFileConfig()

	dbHost := envOr("DB_HOST", fc.Database.Host, "localhost")
	dbPort := envOr("DB_PORT", fc.Database.Port, "3306")
	dbUser := envOr("DB_USER", fc.Database.User, "root")
	dbPassword := envOr("DB_PASSWORD", fc.Database.Password, "test789")
	dbName := envOr("DB_NAME", fc.Database.DBName, "balance")
	dbCharset := envOr("DB_CHARSET", fc.Database.Charset, "utf8mb4")

	redisHost := envOr("REDIS_HOST", fc.Redis.Host, "localhost")
	redisPort := envOr("REDIS_PORT", fc.Redis.Port, "6379")
	redisPassword := envOr("REDIS_PASSWORD", fc.Redis.Password, "test@789")

	// DSN：环境变量优先，其次 config.yaml，最后默认
	defaultDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbCharset)
	dsn := getEnv("DB_DSN", defaultDSN)

	// JWT
	defaultJWT := fc.JWT.Secret
	if defaultJWT == "" {
		defaultJWT = "balance@!*#6250198"
	}
	jwtSecret := getEnv("JWT_SECRET", defaultJWT)

	// JWT过期时间
	jwtExpirationStr := getEnv("JWT_EXPIRATION", fc.JWT.Expiration)
	if jwtExpirationStr == "" {
		jwtExpirationStr = DefaultJWTExpirationStr
	}
	jwtExpiration := parseDuration(jwtExpirationStr)

	// 端口
	defaultAdminPort := fc.Services.Admin.Port
	if defaultAdminPort == "" {
		defaultAdminPort = "19090"
	}
	port := getEnv("ADMIN_PORT", defaultAdminPort)

	return &Config{
		DBDSN:         dsn,
		RedisAddr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		RedisPassword: redisPassword,
		JWTSecret:     jwtSecret,
		JWTExpiration: jwtExpiration,
		Port:          port,
	}
}

// LoadAppConfig 加载 app 服务配置：
// 优先：环境变量 -> 配置文件 config.yaml -> 代码默认值
func LoadAppConfig() *Config {
	fc := loadFileConfig()

	dbHost := envOr("DB_HOST", fc.Database.Host, "localhost")
	dbPort := envOr("DB_PORT", fc.Database.Port, "3306")
	dbUser := envOr("DB_USER", fc.Database.User, "root")
	dbPassword := envOr("DB_PASSWORD", fc.Database.Password, "password")
	dbName := envOr("DB_NAME", fc.Database.DBName, "xshopee")
	dbCharset := envOr("DB_CHARSET", fc.Database.Charset, "utf8mb4")

	redisHost := envOr("REDIS_HOST", fc.Redis.Host, "localhost")
	redisPort := envOr("REDIS_PORT", fc.Redis.Port, "6379")
	redisPassword := envOr("REDIS_PASSWORD", fc.Redis.Password, "")

	defaultDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbCharset)
	dsn := getEnv("DB_DSN", defaultDSN)

	defaultJWT := fc.JWT.Secret
	if defaultJWT == "" {
		defaultJWT = "xshopee-secret-key-change-in-production"
	}
	jwtSecret := getEnv("JWT_SECRET", defaultJWT)

	// JWT过期时间
	jwtExpirationStr := getEnv("JWT_EXPIRATION", fc.JWT.Expiration)
	if jwtExpirationStr == "" {
		jwtExpirationStr = DefaultJWTExpirationStr
	}
	jwtExpiration := parseDuration(jwtExpirationStr)

	defaultAppPort := fc.Services.App.Port
	if defaultAppPort == "" {
		defaultAppPort = "19091"
	}
	port := getEnv("APP_PORT", defaultAppPort)

	return &Config{
		DBDSN:         dsn,
		RedisAddr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		RedisPassword: redisPassword,
		JWTSecret:     jwtSecret,
		JWTExpiration: jwtExpiration,
		Port:          port,
	}
}

// parseDuration 解析时间字符串，支持格式：15d, 360h, 24h, 30m, 60s
// 例如：15d = 15天, 360h = 360小时, 24h = 24小时
func parseDuration(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return DefaultJWTExpiration
	}
	// 移除末尾可能的空格
	s = strings.ToLower(s)
	// 解析数字部分
	var numStr string
	var unit string
	// 找到第一个非数字字符
	for i, r := range s {
		if r < '0' || r > '9' {
			numStr = s[:i]
			unit = s[i:]
			break
		}
	}
	if numStr == "" {
		return DefaultJWTExpiration
	}
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return DefaultJWTExpiration
	}
	// 根据单位转换
	switch unit {
	case "d", "day", "days":
		return time.Duration(num) * 24 * time.Hour
	case "h", "hour", "hours":
		return time.Duration(num) * time.Hour
	case "m", "min", "minute", "minutes":
		return time.Duration(num) * time.Minute
	case "s", "sec", "second", "seconds":
		return time.Duration(num) * time.Second
	default:
		return DefaultJWTExpiration
	}
}

// envOr：优先环境变量，其次配置文件值，最后默认值
func envOr(envKey, fileVal, defVal string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	if fileVal != "" {
		return fileVal
	}
	return defVal
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
