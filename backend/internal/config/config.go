package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构
type Config struct {
	App     AppConfig     `yaml:"app"`
	MySQL   MySQLConfig   `yaml:"mysql"`
	Redis   RedisConfig   `yaml:"redis"`
	Shopee  ShopeeConfig  `yaml:"shopee"`
	Log     LogConfig     `yaml:"log"`
	JWT     JWTConfig     `yaml:"jwt"`
	Email   EmailConfig   `yaml:"email"`
	Payment PaymentConfig `yaml:"payment"`
}

// PaymentConfig 第三方支付配置（预留，待对接时填入真实密钥）
type PaymentConfig struct {
	PayPal  PayPalConfig  `yaml:"paypal"`
	Alipay  AlipayConfig  `yaml:"alipay"`
	LinePay LinePayConfig `yaml:"linepay"`
	Visa    VisaConfig    `yaml:"visa"`
	Wechat  WechatConfig  `yaml:"wechat"`
}

// PayPalConfig PayPal 配置
type PayPalConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Sandbox      bool   `yaml:"sandbox"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	WebhookID    string `yaml:"webhook_id"`
	ReturnURL    string `yaml:"return_url"`
	CancelURL    string `yaml:"cancel_url"`
}

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Sandbox    bool   `yaml:"sandbox"`
	AppID      string `yaml:"app_id"`
	PrivateKey string `yaml:"private_key"`
	PublicKey  string `yaml:"public_key"`
	NotifyURL  string `yaml:"notify_url"`
	ReturnURL  string `yaml:"return_url"`
}

// LinePayConfig LINE Pay 配置
type LinePayConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Sandbox       bool   `yaml:"sandbox"`
	ChannelID     string `yaml:"channel_id"`
	ChannelSecret string `yaml:"channel_secret"`
	ReturnURL     string `yaml:"return_url"`
	CancelURL     string `yaml:"cancel_url"`
}

// VisaConfig VISA/信用卡支付配置（通过 Stripe/TapPay 等通道）
type VisaConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Sandbox   bool   `yaml:"sandbox"`
	Provider  string `yaml:"provider"` // stripe / tappay / ecpay
	APIKey    string `yaml:"api_key"`
	SecretKey string `yaml:"secret_key"`
	WebhookSecret string `yaml:"webhook_secret"`
}

// WechatConfig 微信支付配置
type WechatConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Sandbox    bool   `yaml:"sandbox"`
	AppID      string `yaml:"app_id"`
	MchID      string `yaml:"mch_id"`
	APIKeyV3   string `yaml:"api_key_v3"`
	SerialNo   string `yaml:"serial_no"`
	PrivateKey string `yaml:"private_key"`
	NotifyURL  string `yaml:"notify_url"`
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
