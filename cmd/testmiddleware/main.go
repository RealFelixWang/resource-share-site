/*
Middleware Usage Examples

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package main

import (
	"fmt"
	"net/http"
	"time"

	"resource-share-site/internal/config"
	"resource-share-site/internal/database"
	"resource-share-site/internal/middleware"
	"resource-share-site/internal/service/auth"
	"resource-share-site/internal/service/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 中间件使用示例程序
func main() {
	fmt.Println("=== 中间件使用示例 ===\n")

	// 1. 初始化服务
	fmt.Println("1. 初始化服务...")
	authService, userStatusService := initServices()
	fmt.Println("✅ 服务初始化成功\n")

	// 2. 创建Gin引擎
	fmt.Println("2. 配置Gin引擎...")
	r := setupGinEngine(authService, userStatusService)
	fmt.Println("✅ Gin引擎配置成功\n")

	// 3. 启动服务器
	fmt.Println("3. 启动服务器...")
	fmt.Println("   API文档: http://localhost:8080/docs")
	fmt.Println("   健康检查: http://localhost:8080/health")
	fmt.Println("")
	fmt.Println("=== 测试端点 ===")
	fmt.Println("")
	fmt.Println("1. 公开端点:")
	fmt.Println("   GET  http://localhost:8080/public")
	fmt.Println("   GET  http://localhost:8080/health")
	fmt.Println("")
	fmt.Println("2. 需要登录的端点:")
	fmt.Println("   GET  http://localhost:8080/profile")
	fmt.Println("   POST http://localhost:8080/logout")
	fmt.Println("   Header: Authorization: Bearer <token>")
	fmt.Println("")
	fmt.Println("3. 管理员端点:")
	fmt.Println("   GET  http://localhost:8080/admin/users")
	fmt.Println("   Header: Authorization: Bearer <admin_token>")
	fmt.Println("")
	fmt.Println("4. 上传权限端点:")
	fmt.Println("   POST http://localhost:8080/resources")
	fmt.Println("   Header: Authorization: Bearer <uploader_token>")
	fmt.Println("")
	fmt.Println("按 Ctrl+C 退出服务器\n")

	if err := r.Run(":8080"); err != nil {
		panic(fmt.Sprintf("服务器启动失败: %v", err))
	}
}

// 初始化服务
func initServices() (auth.AuthService, user.UserStatusService) {
	// 初始化数据库
	dbConfig := &config.DatabaseConfig{
		Type:     "sqlite",
		Name:     "resource_share_site",
		Host:     "localhost",
		Port:     "3306",
		User:     "root",
		Password: "123456",
		Charset:  "utf8mb4",
	}

	db, err := database.InitDatabaseWithConfig(dbConfig)
	if err != nil {
		panic(fmt.Sprintf("数据库初始化失败: %v", err))
	}

	// 创建服务
	authService := auth.NewAuthService(db)
	userStatusService := user.NewUserStatusService(db)

	return authService, userStatusService
}

// 配置Gin引擎
func setupGinEngine(authService auth.AuthService, userStatusService user.UserStatusService) *gin.Engine {
	// 创建Gin引擎
	r := gin.Default()

	// 添加全局中间件
	r.Use(middleware.LoggingMiddleware())   // 日志
	r.Use(middleware.RecoverMiddleware())   // 恢复
	r.Use(middleware.CORSMiddleware())      // CORS
	r.Use(middleware.RequestIDMiddleware()) // 请求ID

	// 公开路由
	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "这是公开端点",
			"status":  "success",
		})
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "resource-share-site",
		})
	})

	// 认证路由组
	authGroup := r.Group("/auth")
	{
		// 注册
		authGroup.POST("/register", func(c *gin.Context) {
			var req auth.RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			// 模拟注册
			response, err := authService.Register(&auth.GORMContext{DB: db}, &req)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "注册成功",
				"data":    response,
			})
		})

		// 登录
		authGroup.POST("/login", func(c *gin.Context) {
			var req auth.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			// 模拟登录
			response, err := authService.Login(&auth.GORMContext{DB: db}, &req)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "登录成功",
				"data":    response,
			})
		})
	}

	// 用户路由组（需要认证）
	userGroup := r.Group("/user")
	{
		userGroup.Use(middleware.AuthMiddleware()) // JWT认证中间件

		// 获取用户资料
		userGroup.GET("/profile", func(c *gin.Context) {
			// 从中间件获取用户信息
			authUser, err := middleware.AuthUserFromContext(c)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "用户未认证",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "获取成功",
				"data":    authUser,
			})
		})

		// 更新用户资料
		userGroup.PUT("/profile", func(c *gin.Context) {
			var req auth.UpdateProfileRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			authUser, err := middleware.AuthUserFromContext(c)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "用户未认证",
				})
				return
			}

			// 模拟更新
			c.JSON(http.StatusOK, gin.H{
				"message": "更新成功",
				"data": gin.H{
					"id":       authUser.ID,
					"username": req.Username,
					"email":    req.Email,
				},
			})
		})

		// 登出（删除会话）
		userGroup.POST("/logout", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "登出成功",
			})
		})
	}

	// 管理员路由组（需要管理员权限）
	adminGroup := r.Group("/admin")
	{
		adminGroup.Use(middleware.AdminMiddleware()) // 管理员权限中间件

		// 获取用户列表
		adminGroup.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "获取用户列表成功",
				"data":    []gin.H{},
			})
		})

		// 封禁用户
		adminGroup.POST("/ban/:user_id", func(c *gin.Context) {
			userID := c.Param("user_id")
			reason := c.Query("reason")

			c.JSON(http.StatusOK, gin.H{
				"message": "用户封禁成功",
				"data": gin.H{
					"user_id": userID,
					"reason":  reason,
				},
			})
		})
	}

	// 资源路由组（需要上传权限）
	resourceGroup := r.Group("/resources")
	{
		resourceGroup.Use(middleware.RequireUploadMiddleware()) // 上传权限中间件

		// 创建资源
		resourceGroup.POST("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "资源创建成功",
			})
		})

		// 获取资源列表
		resourceGroup.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "获取资源列表成功",
				"data":    []gin.H{},
			})
		})
	}

	// 速率限制示例
	limitedGroup := r.Group("/limited")
	{
		limitedGroup.Use(middleware.RateLimitMiddleware(5, time.Minute)) // 每分钟最多5次请求

		limitedGroup.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "请求成功",
			})
		})
	}

	return r
}

// 添加全局变量（临时方案）
var db *gorm.DB

// 注意：在实际项目中，db应该通过依赖注入传递
func setupGinEngineWithDB(authService auth.AuthService, userStatusService user.UserStatusService, database *gorm.DB) *gin.Engine {
	db = database // 设置全局变量

	// 其余代码与setupGinEngine相同
	return setupGinEngine(authService, userStatusService)
}
