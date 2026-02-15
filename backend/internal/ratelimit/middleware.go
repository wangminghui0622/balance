package ratelimit

import (
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/gin-gonic/gin"
)

// HTTPRateLimitMiddleware HTTP 接口限流中间件
// qps: 每秒允许的请求数
func HTTPRateLimitMiddleware(qps float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceName := HTTPResourceName(c.FullPath())
		
		// 确保规则已加载
		LoadHTTPRules(c.FullPath(), qps)
		
		entry, blockErr := sentinel.Entry(resourceName)
		if blockErr != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		defer entry.Exit()
		
		c.Next()
	}
}

// IPRateLimitMiddleware 基于 IP 的限流中间件
func IPRateLimitMiddleware(qps float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		resourceName := HTTPResourceName("ip:" + clientIP)
		
		// 动态加载规则
		LoadHTTPRules("ip:"+clientIP, qps)
		
		entry, blockErr := sentinel.Entry(resourceName)
		if blockErr != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		defer entry.Exit()
		
		c.Next()
	}
}

// UserRateLimitMiddleware 基于用户 ID 的限流中间件
func UserRateLimitMiddleware(qps float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}
		
		resourceName := HTTPResourceName("user:" + toString(userID))
		
		// 动态加载规则
		LoadHTTPRules("user:"+toString(userID), qps)
		
		entry, blockErr := sentinel.Entry(resourceName)
		if blockErr != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		defer entry.Exit()
		
		c.Next()
	}
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int64:
		return string(rune(val))
	case int:
		return string(rune(val))
	default:
		return ""
	}
}
