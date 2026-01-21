package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

// 测试 Shopee 授权链接签名生成
func main() {
	partnerID := int64(1203446)
	partnerKey := "724b6a656d626b696b756345464e6b614d524664716c61525a4e4e4f466c" // 去掉 shpk 前缀后
	path := "/api/v2/shop/auth_partner"
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)

	// 签名字符串
	signString := strconv.FormatInt(partnerID, 10) + path + timestampStr
	fmt.Printf("签名字符串: %s\n", signString)

	// 方式1：partner_key 作为十六进制字符串解码
	fmt.Println("\n=== 方式1：十六进制解码 ===")
	partnerKeyBytes1, err := hex.DecodeString(partnerKey)
	if err != nil {
		fmt.Printf("解码失败: %v\n", err)
	} else {
		fmt.Printf("解码后长度: %d 字节\n", len(partnerKeyBytes1))
		mac1 := hmac.New(sha256.New, partnerKeyBytes1)
		mac1.Write([]byte(signString))
		signatureBytes1 := mac1.Sum(nil)
		signature1 := fmt.Sprintf("%064x", new(big.Int).SetBytes(signatureBytes1))
		fmt.Printf("签名: %s\n", signature1)
	}

	// 方式2：partner_key 作为普通字符串（UTF-8）
	fmt.Println("\n=== 方式2：普通字符串（UTF-8）===")
	partnerKeyBytes2 := []byte(partnerKey)
	fmt.Printf("字符串长度: %d 字节\n", len(partnerKeyBytes2))
	mac2 := hmac.New(sha256.New, partnerKeyBytes2)
	mac2.Write([]byte(signString))
	signatureBytes2 := mac2.Sum(nil)
	signature2 := fmt.Sprintf("%064x", new(big.Int).SetBytes(signatureBytes2))
	fmt.Printf("签名: %s\n", signature2)

	// 方式3：尝试不同的签名字符串格式
	fmt.Println("\n=== 方式3：不同的签名字符串格式 ===")
	
	// 3.1: partner_id + timestamp + redirect
	redirect := "https://kx9y.com/api/v1/balance/admin/shopee/auth/callback"
	signString3_1 := strconv.FormatInt(partnerID, 10) + timestampStr + redirect
	fmt.Printf("签名字符串3.1 (partner_id+timestamp+redirect): %s\n", signString3_1)
	mac3_1 := hmac.New(sha256.New, partnerKeyBytes1)
	mac3_1.Write([]byte(signString3_1))
	signature3_1 := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac3_1.Sum(nil)))
	fmt.Printf("签名3.1: %s\n", signature3_1)

	// 3.2: partner_id + path + timestamp + redirect
	signString3_2 := strconv.FormatInt(partnerID, 10) + path + timestampStr + redirect
	fmt.Printf("签名字符串3.2 (partner_id+path+timestamp+redirect): %s\n", signString3_2)
	mac3_2 := hmac.New(sha256.New, partnerKeyBytes1)
	mac3_2.Write([]byte(signString3_2))
	signature3_2 := fmt.Sprintf("%064x", new(big.Int).SetBytes(mac3_2.Sum(nil)))
	fmt.Printf("签名3.2: %s\n", signature3_2)

	fmt.Println("\n=== 建议 ===")
	fmt.Println("请将以上所有签名结果与 Shopee 返回的错误信息对比")
	fmt.Println("或者查看 Shopee 开放平台文档确认正确的签名规则")
}
