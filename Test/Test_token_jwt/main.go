package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT Claims（与 backend/internal/services/auth_service 保持一致）
type Claims struct {
	UserID   int64 `json:"user_id"`
	UserType int   `json:"user_type"`
	jwt.RegisteredClaims
}

func parseToken(token, jwtSecret string) {
	claims, err := verifyToken(token, jwtSecret)
	if err != nil {
		fmt.Printf("Token 验证失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Token 有效\n")
	fmt.Printf("  UserID:   %d\n", claims.UserID)
	fmt.Printf("  UserType: %d\n", claims.UserType)
	fmt.Printf("  ExpiresAt: %v\n", claims.ExpiresAt)
	fmt.Printf("  IssuedAt:  %v\n", claims.IssuedAt)
	return
}
func main() {
	userType := flag.Int("user_type", 5, "用户类型: 1=店主 5=运营 9=平台（运营发货接口需5）")
	flag.Parse()

	jwtSecret := "sheepx-jwt-secret-key-2026"
	expHours := 168
	token, err := generateToken(59906070668, *userType, jwtSecret, expHours)
	if err != nil {
		fmt.Printf("生成 Token 失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(token)
	parseToken(token, jwtSecret)
}

func generateToken(userID int64, userType int, secret string, expireHours int) (string, error) {
	claims := Claims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "sheepx",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func verifyToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("无效的 token")
}
