/*
Package middleware provides HTTP middleware for authentication and authorization.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"resource-share-site/internal/service/auth"
	"resource-share-site/internal/service/user"
	"resource-share-site/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthUser 认证用户信息
type AuthUser struct {
	ID            uint   `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Role          string `json:"role"`   // user, admin
	Status        string `json:"status"` // active, banned, inactive
	CanUpload     bool   `json:"can_upload"`
	PointsBalance int    `json:"points_balance"`
	InviteCode    string `json:"invite_code"`
}

// ContextKey 上下文键类型
type ContextKey string

const (
	// Context键名
	AuthUserKey ContextKey = "auth_user"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header中获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少认证令牌",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			return
		}

		// 验证Token格式 (Bearer token)
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证格式",
				"code":  "INVALID_TOKEN_FORMAT",
			})
			c.Abort()
			return
		}

		// 解析Token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token无效或已过期",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// TODO: 可以添加Token黑名单检查

		// 将用户信息存储到上下文
		c.Set(AuthUserKey, &AuthUser{
			ID:       claims.UserID,
			Username: claims.Username,
		})

		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件（不强制要求登录）
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header中获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有Token，继续执行但不设置用户信息
			c.Next()
			return
		}

		// 验证Token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		// 尝试解析Token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			// Token无效，但不阻断请求
			c.Next()
			return
		}

		// 将用户信息存储到上下文
		c.Set(AuthUserKey, &AuthUser{
			ID:       claims.UserID,
			Username: claims.Username,
		})

		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行JWT认证
		AuthMiddleware()(c)

		// 检查是否被中断
		if c.IsAborted() {
			return
		}

		// 获取用户信息
		user, exists := c.Get(string(AuthUserKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户未认证",
				"code":  "USER_NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		authUser := user.(*AuthUser)

		// 检查是否为管理员
		if authUser.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "需要管理员权限",
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireUploadMiddleware 需要上传权限中间件
func RequireUploadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行JWT认证
		AuthMiddleware()(c)

		// 检查是否被中断
		if c.IsAborted() {
			return
		}

		// 获取用户信息
		user, exists := c.Get(string(AuthUserKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户未认证",
				"code":  "USER_NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		authUser := user.(*AuthUser)

		// 检查是否有上传权限
		if !authUser.CanUpload {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "没有上传权限，请联系管理员",
				"code":  "NO_UPLOAD_PERMISSION",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ActiveUserMiddleware 活跃用户验证中间件
func ActiveUserMiddleware(userStatusService user.UserStatusService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行JWT认证
		AuthMiddleware()(c)

		// 检查是否被中断
		if c.IsAborted() {
			return
		}

		// 获取用户信息
		user, exists := c.Get(string(AuthUserKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户未认证",
				"code":  "USER_NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		authUser := user.(*AuthUser)

		// 检查用户状态
		isActive, err := userStatusService.IsUserActive(&auth.GORMContext{}, authUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "服务器错误",
				"code":  "SERVER_ERROR",
			})
			c.Abort()
			return
		}

		if !isActive {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "账户已被禁用或未激活",
				"code":  "ACCOUNT_DISABLED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware 简单速率限制中间件
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	// 简单的内存存储，实际应用中应使用Redis
	var requests = make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 获取当前时间
		now := time.Now()

		// 清理过旧请求
		var validTimes []time.Time
		for _, reqTime := range requests[clientIP] {
			if now.Sub(reqTime) < window {
				validTimes = append(validTimes, reqTime)
			}
		}
		requests[clientIP] = validTimes

		// 检查是否超过限制
		if len(requests[clientIP]) >= limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":  "请求过于频繁，请稍后再试",
				"code":   "RATE_LIMIT_EXCEEDED",
				"limit":  limit,
				"window": window.String(),
			})
			c.Abort()
			return
		}

		// 记录当前请求
		requests[clientIP] = append(requests[clientIP], now)

		c.Next()
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Max-Age", "86400")

		if method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RecoverMiddleware 恢复中间件（Panic处理）
func RecoverMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "服务器内部错误",
				"message": err,
				"code":    "INTERNAL_SERVER_ERROR",
			})
		} else if recovered != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "服务器内部错误",
				"message": "系统发生未知错误",
				"code":    "INTERNAL_SERVER_ERROR",
			})
		}
		c.Abort()
	})
}

// RequestIDMiddleware 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成或获取请求ID
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 设置响应头
		c.Header("X-Request-ID", requestID)

		// 存储到上下文
		c.Set("request_id", requestID)

		c.Next()
	}
}

// LoggingMiddleware 请求日志中间件
func LoggingMiddleware() gin.HandlerFunc {
	// 使用默认的日志格式
	return gin.Logger()
}

// AuthUserFromContext 从上下文中获取认证用户
func AuthUserFromContext(c *gin.Context) (*AuthUser, error) {
	user, exists := c.Get(string(AuthUserKey))
	if !exists {
		return nil, errors.New("用户未认证")
	}

	authUser, ok := user.(*AuthUser)
	if !ok {
		return nil, errors.New("用户信息类型错误")
	}

	return authUser, nil
}

// RequireUserIDMiddleware 强制要求用户ID参数
func RequireUserIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "缺少用户ID参数",
				"code":  "MISSING_USER_ID",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OwnerOrAdminMiddleware 资源所有者或管理员权限中间件
func OwnerOrAdminMiddleware(getResourceOwnerID func(*gin.Context) (uint, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行JWT认证
		AuthMiddleware()(c)

		// 检查是否被中断
		if c.IsAborted() {
			return
		}

		// 获取当前用户
		authUser, err := AuthUserFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户未认证",
				"code":  "USER_NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		// 获取资源所有者ID
		ownerID, err := getResourceOwnerID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无法获取资源所有者信息",
				"code":  "CANNOT_GET_OWNER",
			})
			c.Abort()
			return
		}

		// 检查是否为资源所有者或管理员
		if authUser.ID != ownerID && authUser.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "没有权限访问此资源",
				"code":  "NO_PERMISSION",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + generateRandomString(8)
}

// generateRandomString 生成随机字符串
func generateRandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
