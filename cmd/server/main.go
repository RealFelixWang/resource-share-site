/*
Resource Share Site - 主服务器入口

整合所有模块的完整Web服务器：
- 用户认证系统
- 邀请系统
- 分类系统
- 资源系统
- 积分系统
- SEO优化系统

Author: Felix Wang
Email: felixwang.biz@gmail.com
Date: 2025-10-31
*/

package main

import (
	"log"
	"net/http"
	"os"

	"resource-share-site/internal/config"
	"resource-share-site/internal/handler"
	"resource-share-site/internal/middleware"
	"resource-share-site/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	// 1. 初始化数据库（使用SQLite便于测试）
	cfgDB := &config.DatabaseConfig{
		Type:    "sqlite",
		Name:    "resource_share.db",
		Charset: "utf8mb4",
	}
	db, err := config.InitDatabase(cfgDB)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 3. 自动迁移数据表
	if err := migrateDatabase(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 4. 初始化Gin
	gin.SetMode(gin.ReleaseMode)

	// 设置Gin配置
	router := gin.Default()

	// 5. 添加中间件
	middleware.RegisterMiddlewares(router)

	// 6. 创建HTTP处理器
	h := handler.NewHandler(db)

	// 7. 注册所有路由
	h.RegisterRoutes(router)

	// 8. 设置路由
	setupRoutes(router)

	// 9. 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 10. 启动服务器
	log.Printf("🚀 资源分享网站服务器启动中...")
	log.Printf("📍 访问地址: http://localhost:%s", port)
	log.Printf("📖 API文档: http://localhost:%s/", port)
	log.Printf("🏥 健康检查: http://localhost:%s/health", port)
	log.Printf("🗺️  Sitemap: http://localhost:%s/seo/sitemap.xml", port)
	log.Printf("=====================================")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// migrateDatabase 执行数据库迁移
func migrateDatabase(db *gorm.DB) error {
	// 使用GORM自动迁移
	return db.AutoMigrate(
		// 用户相关
		&model.User{},

		// 邀请相关
		&model.Invitation{},

		// 分类相关
		&model.Category{},

		// 资源相关
		&model.Resource{},
		&model.Comment{},

		// 文章博客相关
		&model.Article{},
		&model.ArticleComment{},

		// 积分相关
		&model.PointsRule{},
		&model.PointRecord{},

		// 商城相关
		&model.Product{},
		&model.MallOrder{},

		// SEO相关
		&model.SEOConfig{},
		&model.MetaTag{},
		&model.SitemapUrl{},
		&model.SEOKeyword{},
		&model.SEORank{},
		&model.SEOReport{},
		&model.SEOEvent{},

		// 其他
		&model.Ad{},
		&model.VisitLog{},
		&model.IPBlacklist{},
	)
}

// setupRoutes 设置静态路由和错误处理
func setupRoutes(router *gin.Engine) {
	// 静态文件服务
	router.Static("/static", "./web/static")
	router.StaticFS("/uploads", http.Dir("./uploads"))

	// 全局错误处理
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error":   "404 - 页面未找到",
			"message": "请检查您的请求路径是否正确",
		})
	})

	router.NoMethod(func(c *gin.Context) {
		c.JSON(405, gin.H{
			"error":   "405 - 方法不允许",
			"message": "请检查您的HTTP方法是否正确",
		})
	})

	// 恢复Panic
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("Panic recovered: %v", recovered)
		c.JSON(500, gin.H{
			"error":   "500 - 服务器内部错误",
			"message": "请稍后再试或联系管理员",
		})
	}))
}
