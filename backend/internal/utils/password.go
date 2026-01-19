package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateSalt 生成随机盐值
func GenerateSalt() (string, error) {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// HashPassword 使用SHA256加密密码
func HashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

// VerifyPassword 验证密码
func VerifyPassword(password, salt, hash string) bool {
	computedHash := HashPassword(password, salt)
	return computedHash == hash
}

// GenerateUserNo 生成用户编号（根据ID生成）
func GenerateUserNo(id int64) string {
	return fmt.Sprintf("U%011d", id)
}
