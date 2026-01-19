package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken 生成JWT Token
func GenerateToken(userID int64, secret []byte, expiration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(expiration).Unix(),
		"iat":    time.Now().Unix(),
		"iss":    "balance-admin",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken 解析JWT Token
func ParseToken(tokenStr string, secret []byte) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, jwt.ErrSignatureInvalid
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if userId, ok := claims["userId"].(float64); ok {
			return int64(userId), nil
		}
	}

	return 0, jwt.ErrSignatureInvalid
}
