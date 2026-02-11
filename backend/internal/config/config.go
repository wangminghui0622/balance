package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构
type Config struct {
	App    AppConfig    `yaml:"app"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
	Shopee ShopeeConfig `yaml:"shopee"`
	Log    LogConfig    `yaml:"log"`
	JWT    JWTConfig    `yaml:"jwt"`
	Email  EmailConfig  `yaml:"email"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	FromName string `yaml:"from_name"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name string `yaml:"name"`
	Mode string `yaml:"mode"`
	Port int    `yaml:"port"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Charset         string `yaml:"charset"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// DSN 生成MySQL连接字符串
func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// Addr 生成Redis地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ShopeeConfig 虾皮API配置
type ShopeeConfig struct {
	PartnerID    int64             `yaml:"partner_id"`
	PartnerKey   string            `yaml:"partner_key"`
	IsProduction bool              `yaml:"is_production"` // true-正式环境 false-沙箱环境
	Hosts        map[string]string `yaml:"hosts"`         // 正式环境Host
	SandboxHosts map[string]string `yaml:"sandbox_hosts"` // 沙箱环境Host
	RedirectURL  string            `yaml:"redirect_url"`
}

// GetHost 获取指定地区的API Host
func (c *ShopeeConfig) GetHost(region string) string {
	var hosts map[string]string
	if c.IsProduction {
		hosts = c.Hosts
	} else {
		hosts = c.SandboxHosts
	}

	if host, ok := hosts[region]; ok {
		return host
	}
	// 默认返回SG
	if host, ok := hosts["SG"]; ok {
		return host
	}
	return ""
}

// IsSandbox 是否为沙箱环境
func (c *ShopeeConfig) IsSandbox() bool {
	return !c.IsProduction
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

var globalConfig *Config

// Load 加载配置文件
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}
