package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"time"

	"balance/backend/internal/config"
	"balance/backend/internal/database"
	"balance/backend/internal/utils"

	"gopkg.in/gomail.v2"
)

// EmailService 邮件服务
type EmailService struct{}

// NewEmailService 创建邮件服务
func NewEmailService() *EmailService {
	return &EmailService{}
}

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// SendCodeResponse 发送验证码响应
type SendCodeResponse struct {
	Message string `json:"message"`
}

// Redis key前缀和过期时间
const (
	EmailCodeKeyPrefix = "email:code:"
	EmailCodeExpire    = 5 * time.Minute // 验证码5分钟过期
	EmailCodeCooldown  = 60 * time.Second // 发送冷却时间60秒
	EmailCooldownKey   = "email:cooldown:"
)

// GenerateCode 生成6位随机验证码（不以0开头）
func GenerateCode() string {
	rand.Seed(time.Now().UnixNano())
	// 第一位1-9，后5位0-9
	first := rand.Intn(9) + 1 // 1-9
	rest := rand.Intn(100000) // 0-99999
	return fmt.Sprintf("%d%05d", first, rest)
}

// SendVerificationCode 发送邮箱验证码
func (s *EmailService) SendVerificationCode(ctx context.Context, req *SendCodeRequest) error {
	rdb := database.GetRedis()

	// TODO: 临时写死验证码为123456，方便测试注册
	code := "123456"

	// 存储到Redis
	codeKey := EmailCodeKeyPrefix + req.Email
	if err := rdb.Set(ctx, codeKey, code, EmailCodeExpire).Err(); err != nil {
		return fmt.Errorf("存储验证码失败: %w", err)
	}

	fmt.Printf("[DEBUG] 邮箱验证码: %s -> %s\n", req.Email, code)

	// // 检查发送冷却
	// cooldownKey := EmailCooldownKey + req.Email
	// exists, err := rdb.Exists(ctx, cooldownKey).Result()
	// if err != nil {
	// 	return fmt.Errorf("检查发送频率失败: %w", err)
	// }
	// if exists > 0 {
	// 	return fmt.Errorf("发送太频繁，请稍后再试")
	// }

	// // 生成验证码
	// code := GenerateCode()

	// // 存储到Redis
	// codeKey := EmailCodeKeyPrefix + req.Email
	// if err := rdb.Set(ctx, codeKey, code, EmailCodeExpire).Err(); err != nil {
	// 	return fmt.Errorf("存储验证码失败: %w", err)
	// }

	// // 设置发送冷却
	// if err := rdb.Set(ctx, cooldownKey, "1", EmailCodeCooldown).Err(); err != nil {
	// 	// 冷却设置失败不影响主流程
	// 	fmt.Printf("设置发送冷却失败: %v\n", err)
	// }

	// // 发送邮件
	// if err := s.sendEmail(req.Email, code); err != nil {
	// 	// 发送失败，删除验证码
	// 	rdb.Del(ctx, codeKey)
	// 	return fmt.Errorf("发送邮件失败: %w", err)
	// }

	return nil
}

// VerifyCode 验证邮箱验证码
func (s *EmailService) VerifyCode(ctx context.Context, email, code string) error {
	rdb := database.GetRedis()

	codeKey := EmailCodeKeyPrefix + email
	storedCode, err := rdb.Get(ctx, codeKey).Result()
	if err != nil {
		return utils.ErrEmailCodeExpired
	}

	if storedCode != code {
		return utils.ErrEmailCodeInvalid
	}

	// 验证成功后删除验证码
	rdb.Del(ctx, codeKey)

	return nil
}

// sendEmail 发送邮件
func (s *EmailService) sendEmail(to, code string) error {
	cfg := config.Get()
	emailCfg := cfg.Email

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(emailCfg.From, emailCfg.FromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", "【天平系统】邮箱验证码")

	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #ff6600;">天平系统</h2>
			<p>您好！</p>
			<p>您的邮箱验证码是：</p>
			<div style="background: #f5f5f5; padding: 20px; text-align: center; margin: 20px 0;">
				<span style="font-size: 32px; font-weight: bold; color: #ff6600; letter-spacing: 5px;">%s</span>
			</div>
			<p>验证码有效期为5分钟，请尽快使用。</p>
			<p>如果这不是您的操作，请忽略此邮件。</p>
			<hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
			<p style="color: #999; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
		</div>
	`, code)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(emailCfg.Host, emailCfg.Port, emailCfg.Username, emailCfg.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
