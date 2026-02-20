package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	configPath := flag.String("config", "backend/admin/config/config.yaml", "配置文件路径")
	region := flag.String("region", "SG", "地区代码 (SG/MY/TW/TH/VN/PH 等)")
	code := flag.String("code", "", "授权回调中的 code，用于换取 access_token")
	shopIDStr := flag.String("shop_id", "", "店铺 ID (授权回调中的 shop_id)")
	refreshToken := flag.String("refresh", "", "refresh_token，用于刷新获取新的 access_token")
	flag.Parse()

	// 加载配置
	cfg, err := Load(*configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		fmt.Println("提示: 请在项目根目录下运行，或使用 -config 指定配置文件路径")
		os.Exit(1)
	}

	client := NewClient(*region)
	fmt.Printf("[Config] Region=%s Host=%s PartnerID=%d\n\n", *region, client.GetHost(), cfg.Shopee.PartnerID)

	// 模式1: 使用 refresh_token 刷新
	if *refreshToken != "" && *shopIDStr != "" {
		shopID, _ := strconv.ParseUint(*shopIDStr, 10, 64)
		tokenResp, err := client.RefreshAccessToken(*refreshToken, shopID)
		if err != nil {
			fmt.Printf("刷新 Token 失败: %v\n", err)
			os.Exit(1)
		}
		printTokenResult(tokenResp)
		return
	}

	// 模式2: 使用 code 换取 token
	if *code != "" && *shopIDStr != "" {
		shopID, _ := strconv.ParseUint(*shopIDStr, 10, 64)
		tokenResp, err := client.GetAccessToken(*code, shopID)
		if err != nil {
			fmt.Printf("获取 Token 失败: %v\n", err)
			os.Exit(1)
		}
		printTokenResult(tokenResp)
		return
	}

	// 模式3: 仅生成授权 URL
	authURL := client.GetAuthURL(cfg.Shopee.RedirectURL, "test_state")
	fmt.Println("=== Shopee 授权 URL ===")
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("=== 使用说明 ===")
	fmt.Println("1. 在浏览器中打开上述 URL 完成授权")
	fmt.Println("2. 授权成功后，Shopee 会跳转到 redirect_url，URL 中会包含 code 和 shop_id 参数")
	fmt.Println("3. 从回调 URL 中复制 code 和 shop_id，然后执行:")
	fmt.Printf("   go run main.go -code=<你的code> -shop_id=<你的shop_id> -config=%s\n", *configPath)
	fmt.Println()
	fmt.Println("或使用 refresh_token 刷新 (若已有 refresh_token):")
	fmt.Printf("   go run main.go -refresh=<refresh_token> -shop_id=<shop_id> -config=%s\n", *configPath)
	fmt.Println()
	fmt.Print("也可在此处粘贴 code 和 shop_id (格式: code shop_id): ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		var inputCode, inputShopID string
		n, _ := fmt.Sscanf(scanner.Text(), "%s %s", &inputCode, &inputShopID)
		if n >= 2 && inputCode != "" && inputShopID != "" {
			shopID, _ := strconv.ParseUint(inputShopID, 10, 64)
			tokenResp, err := client.GetAccessToken(inputCode, shopID)
			if err != nil {
				fmt.Printf("获取 Token 失败: %v\n", err)
				os.Exit(1)
			}
			printTokenResult(tokenResp)
			return
		}
	}
}

func printTokenResult(r *TokenResponse) {
	fmt.Println()
	fmt.Println("=== Token 结果 ===")
	fmt.Printf("AccessToken:     %s\n", r.AccessToken)
	fmt.Printf("RefreshToken:    %s\n", r.RefreshToken)
	fmt.Printf("ExpireIn:        %d 秒\n", r.ExpireIn)
	fmt.Printf("RefreshExpireIn: %d 秒\n", r.RefreshExpireIn)
	fmt.Printf("PartnerID:       %d\n", r.PartnerID)
	fmt.Printf("ShopIDList:      %v\n", r.ShopIDList)
}
