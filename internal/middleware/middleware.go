/*
Package middleware provides HTTP middleware for authentication and authorization.

Author: Felix Wang
Email: felixwang.biz@gmail.com
Date: 2025-10-31
*/

package middleware

import (
	"github.com/gin-gonic/gin"
)

// RegisterMiddlewares 注册所有中间件
func RegisterMiddlewares(router *gin.Engine) {
	// 1. CORS中间件
	router.Use(CORS())

	// 2. 请求日志中间件
	router.Use(RequestLogger())

	// 3. 恢复中间件
	router.Use(Recovery())

	// 4. 限流中间件（可选）
	// router.Use(RateLimit(100, 60)) // 每分钟100次请求

	// 5. 缓存中间件（可选）
	// router.Use(CacheMiddleware())
}

// CORS CORS中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return gin.Logger()
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 记录恢复的Panic
		// log.Printf("Panic recovered: %v", recovered)

		// 返回错误响应
		c.JSON(500, gin.H{
			"error":   "服务器内部错误",
			"message": "请稍后再试或联系管理员",
		})

		c.Abort()
	})
}

// RateLimit 限流中间件（简单实现）
func RateLimit(requests int, perSeconds int) gin.HandlerFunc {
	// 这里可以实现更复杂的限流逻辑
	// 目前返回空实现避免编译错误
	return func(c *gin.Context) {
		c.Next()
	}
}

// CacheMiddleware 缓存中间件（简单实现）
func CacheMiddleware() gin.HandlerFunc {
	// 这里可以实现缓存逻辑
	// 目前返回空实现避免编译错误
	return func(c *gin.Context) {
		c.Next()
	}
}
