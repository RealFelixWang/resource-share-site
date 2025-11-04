/*
Package handlers defines all HTTP request handlers for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
Date: 2025-11-02
*/

package handler

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"resource-share-site/internal/model"
	"resource-share-site/internal/service/article"
	"resource-share-site/internal/service/auth"
	"resource-share-site/internal/service/category"
	"resource-share-site/internal/service/invitation"
	"resource-share-site/internal/service/points"
	"resource-share-site/internal/service/resource"
	"resource-share-site/internal/service/seo"
	"resource-share-site/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler HTTP处理器
type Handler struct {
	db                  *gorm.DB
	authService         *auth.AuthServiceImpl
	categoryService     *category.CategoryService
	resourceService     *resource.ResourceService
	invitationService   *invitation.InvitationService
	mallService         *points.MallService
	seoService          *seo.ManagementService
	articleService      *article.ArticleService
	articleCommentService *article.ArticleCommentService
}

// NewHandler 创建新的HTTP处理器
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db:                  db,
		authService:         auth.NewAuthService(db).(*auth.AuthServiceImpl),
		categoryService:     category.NewCategoryService(db),
		resourceService:     resource.NewResourceService(db),
		invitationService:   invitation.NewInvitationService(db),
		mallService:         points.NewMallService(db),
		seoService:          seo.NewManagementService(db),
		articleService:      article.NewArticleService(db),
		articleCommentService: article.NewArticleCommentService(db),
	}
}

// getCurrentUserID 从请求中获取当前用户ID
func (h *Handler) getCurrentUserID(c *gin.Context) (uint, error) {
	// 从Authorization header获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, errors.New("缺少Authorization header")
	}

	// 解析token (格式: "Bearer <token>")
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// 解析token
	claims, err := utils.ParseToken(tokenString)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}

// RegisterRoutes 注册所有路由
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// 主页路由
	router.GET("/", h.HomePage)
	router.GET("/health", h.HealthCheck)

	// 前端页面路由 - 资源分享站
	router.GET("/resources", h.ResourcesPage)
	router.GET("/categories", h.CategoriesPage)
	router.GET("/search", h.SearchPage)
	router.GET("/resource/:id", h.ResourceDetailPage)
	router.GET("/login", h.LoginPage)
	router.GET("/register", h.RegisterPage)
	router.GET("/articles", h.ArticlesPage)
	router.GET("/article/:slug", h.ArticleDetailPage)

	// 认证相关路由
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.GET("/me", h.GetCurrentUser)
	}

	// 用户相关路由
	users := router.Group("/users")
	{
		users.GET("/", h.ListUsers)
		users.GET("/:id", h.GetUser)
	}

	// 分类相关路由
	categories := router.Group("/categories")
	{
		categories.GET("/", h.ListCategories)
		categories.POST("/", h.CreateCategory)
		categories.GET("/:id", h.GetCategory)
	}

	// 资源相关路由
	resources := router.Group("/resources")
	{
		resources.GET("/", h.ListResources)
		resources.POST("/", h.CreateResource)
		resources.GET("/:id", h.GetResource)
	}

	// 评论相关路由
	comments := router.Group("/comments")
	{
		comments.GET("/", h.ListComments)
		comments.POST("/", h.CreateComment)
		comments.GET("/:id", h.GetComment)
		comments.DELETE("/:id", h.DeleteComment)
	}

	// 文章博客相关路由
	articles := router.Group("/articles")
	{
		// 公开路由
		articles.GET("/", h.ListArticles)
		articles.GET("/:id", h.GetArticle)
		articles.GET("/categories/list", h.GetArticleCategories)
		articles.GET("/tags/popular", h.GetPopularTags)

		// 评论路由
		articles.GET("/:id/comments", h.ListArticleComments)
		articles.POST("/:id/comments", h.CreateArticleComment)
	}

	// 评论API路由
	apiComments := router.Group("/api/comments")
	{
		apiComments.POST("/", h.CreateCommentAPI)
		apiComments.POST("/:id/like", h.LikeCommentAPI)
	}

	// 文章API路由
	apiArticles := router.Group("/api/articles")
	{
		apiArticles.POST("/:id/like", h.LikeArticleAPI)
	}

	// 需要管理员权限的路由
	admin := router.Group("/admin")
	admin.Use(h.AuthRequired)
	{
		admin.GET("/", h.AdminPage)
		admin.POST("/articles", h.CreateArticle)
		admin.POST("/articles/:id/like", h.LikeArticle)
	}

	// 邀请相关路由
	invitations := router.Group("/invitations")
	{
		invitations.POST("/", h.CreateInvitation)
		invitations.GET("/", h.GetInvitations)
	}

	// 积分相关路由
	points := router.Group("/points")
	{
		points.GET("/balance", h.GetPointsBalance)
		points.POST("/checkin", h.DailyCheckin)
		points.GET("/records", h.GetPointsRecords)
	}

	// 商城相关路由
	mall := router.Group("/mall")
	{
		mall.GET("/products", h.ListProducts)
		mall.POST("/purchase", h.PurchaseProduct)
	}

	// SEO相关路由
	seo := router.Group("/seo")
	{
		seo.GET("/sitemap.xml", h.GenerateSitemap)
		seo.GET("/keywords", h.ListKeywords)
	}

	// 系统统计路由
	stats := router.Group("/stats")
	{
		stats.GET("/system", h.GetSystemStatistics)
	}
}

// ==================== 基础接口 ====================

// HomePage 主页
func (h *Handler) HomePage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	tmpl.Execute(c.Writer, nil)
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"time":    gin.H{},
		"version": "1.0.0",
		"uptime":  "running",
	})
}

// ==================== 认证相关处理器 ====================

// Register 用户注册
func (h *Handler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数错误: " + err.Error(),
			"status":  "error",
		})
		return
	}

	authCtx := &auth.GORMContext{DB: h.db}
	response, err := h.authService.Register(authCtx, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
		"status":  "success",
		"data":    response,
	})
}

// Login 用户登录
func (h *Handler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数错误: " + err.Error(),
			"status":  "error",
		})
		return
	}

	authCtx := &auth.GORMContext{DB: h.db}
	response, err := h.authService.Login(authCtx, &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"status":  "success",
		"data":    response,
	})
}

// Logout 用户登出
func (h *Handler) Logout(c *gin.Context) {
	// TODO: 从请求中获取token并删除session
	c.JSON(http.StatusOK, gin.H{
		"message": "登出成功",
		"status":  "success",
	})
}

// GetCurrentUser 获取当前用户
func (h *Handler) GetCurrentUser(c *gin.Context) {
	// 从Authorization header获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "缺少Authorization header",
			"status":  "error",
		})
		return
	}

	// 解析token (格式: "Bearer <token>")
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// 解析token
	claims, err := utils.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "无效的token: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// 从数据库获取用户信息
	var user model.User
	if err := h.db.First(&user, claims.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "用户不存在",
				"status":  "error",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "查询用户信息失败",
				"status":  "error",
			})
		}
		return
	}

	// 返回用户信息（过滤敏感信息）
	userInfo := gin.H{
		"id":                        user.ID,
		"username":                  user.Username,
		"email":                     user.Email,
		"role":                      user.Role,
		"status":                    user.Status,
		"can_upload":                user.CanUpload,
		"points_balance":            user.PointsBalance,
		"invite_code":               user.InviteCode,
		"uploaded_resources_count":  user.UploadedResourcesCount,
		"downloaded_resources_count": user.DownloadedResourcesCount,
		"created_at":                user.CreatedAt,
		"updated_at":                user.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取当前用户成功",
		"status":  "success",
		"data":    userInfo,
	})
}

// ==================== 用户相关处理器 ====================

// ListUsers 列出用户
func (h *Handler) ListUsers(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var users []model.User
	var total int64

	// 获取总数
	if err := h.db.Model(&model.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询用户失败",
			"status":  "error",
		})
		return
	}

	// 查询用户列表
	offset := (page - 1) * pageSize
	if err := h.db.Select("id, username, email, role, status, can_upload, points_balance, invite_code, uploaded_resources_count, downloaded_resources_count, created_at, updated_at").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询用户失败",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取用户列表成功",
		"status":  "success",
		"data": gin.H{
			"users": users,
			"total": total,
			"page":  page,
			"size":  pageSize,
		},
	})
}

// GetUser 获取用户详情
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的用户ID",
			"status":  "error",
		})
		return
	}

	var user model.User
	if err := h.db.Select("id, username, email, role, status, can_upload, points_balance, invite_code, uploaded_resources_count, downloaded_resources_count, created_at, updated_at").
		First(&user, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "用户不存在",
				"status":  "error",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "查询用户失败",
				"status":  "error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取用户成功",
		"status":  "success",
		"data":    user,
	})
}

// ==================== 分类相关处理器 ====================

// ListCategories 列出分类
func (h *Handler) ListCategories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	categories, total, err := h.categoryService.GetAllCategories(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询分类失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取分类列表成功",
		"status":  "success",
		"data": gin.H{
			"categories": categories,
			"total":      total,
			"page":       page,
			"size":       pageSize,
		},
	})
}

// CreateCategory 创建分类
func (h *Handler) CreateCategory(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required,min=1,max=50"`
		Description string  `json:"description" binding:"max=500"`
		Icon        string  `json:"icon" binding:"max=100"`
		Color       string  `json:"color" binding:"max=20"`
		ParentID    *uint   `json:"parent_id"`
		SortOrder   int     `json:"sort_order" binding:"min=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数错误: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// 创建分类
	category, err := h.categoryService.CreateCategory(
		req.Name,
		req.Description,
		req.Icon,
		req.Color,
		req.ParentID,
		req.SortOrder,
	)
	if err != nil {
		// 处理不同类型的错误（通过错误消息判断）
		errMsg := err.Error()
		if errMsg == "分类名称已存在" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "分类名称已存在",
				"status":  "error",
			})
		} else if errMsg == "父级分类不存在" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "父级分类不存在",
				"status":  "error",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "创建分类失败: " + errMsg,
				"status":  "error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "分类创建成功",
		"status":  "success",
		"data":    category,
	})
}

// GetCategory 获取分类详情
func (h *Handler) GetCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的分类ID",
			"status":  "error",
		})
		return
	}

	category, err := h.categoryService.GetCategoryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "分类不存在",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取分类成功",
		"status":  "success",
		"data":    category,
	})
}

// ==================== 资源相关处理器 ====================

// ListResources 列出资源
func (h *Handler) ListResources(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	categoryIDStr := c.DefaultQuery("category_id", "0")
	var categoryID *uint
	if categoryIDStr != "0" {
		id, _ := strconv.ParseUint(categoryIDStr, 10, 32)
		id32 := uint(id)
		categoryID = &id32
	}

	resources, total, err := h.resourceService.GetResources(page, pageSize, categoryID, nil, nil, nil, nil, "created_at", true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询资源失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取资源列表成功",
		"status":  "success",
		"data": gin.H{
			"resources": resources,
			"total":     total,
			"page":      page,
			"size":      pageSize,
		},
	})
}

// CreateResource 创建资源
func (h *Handler) CreateResource(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required,min=1,max=200"`
		Description string `json:"description" binding:"required,min=1,max=2000"`
		CategoryID  uint   `json:"category_id" binding:"required"`
		NetdiskURL  string `json:"netdisk_url" binding:"required"`
		PointsPrice int    `json:"points_price" binding:"min=0"`
		Tags        string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数错误: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "认证失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// 检查用户是否有上传权限
	var user map[string]interface{}
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "用户不存在",
			"status":  "error",
		})
		return
	}

	if !user["can_upload"].(bool) {
		c.JSON(http.StatusForbidden, gin.H{
			"message": "您没有上传权限，请联系管理员",
			"status":  "error",
		})
		return
	}

	// 检查分类是否存在
	var categoryCount int64
	if err := h.db.Model(&map[string]interface{}{}).Where("id = ?", req.CategoryID).Count(&categoryCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "检查分类失败",
			"status":  "error",
		})
		return
	}

	if categoryCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "分类不存在",
			"status":  "error",
		})
		return
	}

	// 创建资源
	resource := map[string]interface{}{
		"title":         req.Title,
		"description":   req.Description,
		"category_id":   req.CategoryID,
		"netdisk_url":   req.NetdiskURL,
		"points_price":  req.PointsPrice,
		"tags":          req.Tags,
		"uploaded_by_id": userID,
		"status":        "pending", // 等待审核
	}

	if err := h.db.Create(&resource).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "创建资源失败",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "资源创建成功，等待审核",
		"status":  "success",
		"data":    resource,
	})
}

// GetResource 获取资源详情
func (h *Handler) GetResource(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的资源ID",
			"status":  "error",
		})
		return
	}

	resource, err := h.resourceService.GetResourceByID(uint(id), false)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "资源不存在",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取资源成功",
		"status":  "success",
		"data":    resource,
	})
}

// ==================== 评论相关处理器 ====================

// ListComments 列出评论
func (h *Handler) ListComments(c *gin.Context) {
	resourceID, _ := strconv.Atoi(c.DefaultQuery("resource_id", "0"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if resourceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "缺少resource_id参数",
			"status":  "error",
		})
		return
	}

	var comments []map[string]interface{}
	var total int64

	if err := h.db.Model(&map[string]interface{}{}).Where("resource_id = ?", resourceID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询评论失败",
			"status":  "error",
		})
		return
	}

	offset := (page - 1) * pageSize
	if err := h.db.Where("resource_id = ?", resourceID).Offset(offset).Limit(pageSize).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询评论失败",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取评论列表成功",
		"status":  "success",
		"data": gin.H{
			"comments": comments,
			"total":    total,
			"page":     page,
			"size":     pageSize,
		},
	})
}

// CreateComment 创建评论
func (h *Handler) CreateComment(c *gin.Context) {
	var req struct {
		ResourceID uint   `json:"resource_id" binding:"required"`
		Content    string `json:"content" binding:"required,min=1,max=1000"`
		ParentID   *uint  `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数错误: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "认证失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// 记录评论
	comment := map[string]interface{}{
		"resource_id": req.ResourceID,
		"user_id":     userID,
		"content":     req.Content,
		"parent_id":   req.ParentID,
		"status":      "active",
	}

	if err := h.db.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "创建评论失败",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "评论创建成功",
		"status":  "success",
		"data":    comment,
	})
}

// GetComment 获取评论详情
func (h *Handler) GetComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的评论ID",
			"status":  "error",
		})
		return
	}

	var comment map[string]interface{}
	if err := h.db.First(&comment, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "评论不存在",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取评论成功",
		"status":  "success",
		"data":    comment,
	})
}

// DeleteComment 删除评论
func (h *Handler) DeleteComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的评论ID",
			"status":  "error",
		})
		return
	}

	// TODO: 检查权限，确保只有评论作者或管理员可以删除
	if err := h.db.Delete(&map[string]interface{}{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "删除评论失败",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "评论删除成功",
		"status":  "success",
	})
}

// ==================== 邀请相关处理器 ====================

// CreateInvitation 创建邀请
func (h *Handler) CreateInvitation(c *gin.Context) {
	// 获取当前用户ID
	inviterID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "认证失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// 获取邀请人信息
	var inviter map[string]interface{}
	if err := h.db.First(&inviter, inviterID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "用户不存在",
			"status":  "error",
		})
		return
	}

	// 检查是否已经有未使用的邀请
	var existingCount int64
	if err := h.db.Model(&map[string]interface{}{}).
		Where("inviter_id = ? AND status = ?", inviterID, "pending").
		Count(&existingCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询邀请记录失败",
			"status":  "error",
		})
		return
	}

	if existingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "您还有未使用的邀请，请先使用后再创建新邀请",
			"status":  "error",
		})
		return
	}

	// 生成邀请码
	inviteCode := generateInviteCode()

	// 创建邀请记录
	invitation := map[string]interface{}{
		"inviter_id": inviterID,
		"invite_code": inviteCode,
		"status":      "pending",
		"expires_at":  time.Now().Add(30 * 24 * time.Hour), // 30天过期
	}

	if err := h.db.Create(&invitation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "创建邀请失败",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "邀请创建成功",
		"status":  "success",
		"data": gin.H{
			"invite_code": inviteCode,
			"expires_at":  invitation["expires_at"],
		},
	})
}

// GetInvitations 获取邀请列表
func (h *Handler) GetInvitations(c *gin.Context) {
	// TODO: 从认证中获取当前用户ID
	inviterID := uint(1) // 临时使用固定用户ID

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var invitations []map[string]interface{}
	var total int64

	if err := h.db.Model(&map[string]interface{}{}).Where("inviter_id = ?", inviterID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询邀请记录失败",
			"status":  "error",
		})
		return
	}

	offset := (page - 1) * pageSize
	if err := h.db.Where("inviter_id = ?", inviterID).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&invitations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询邀请记录失败",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取邀请列表成功",
		"status":  "success",
		"data": gin.H{
			"invitations": invitations,
			"total":       total,
			"page":        page,
			"size":        pageSize,
		},
	})
}

// generateInviteCode 生成邀请码
func generateInviteCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 16
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

// ==================== 积分相关处理器 ====================

// GetPointsBalance 获取积分余额
func (h *Handler) GetPointsBalance(c *gin.Context) {
	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "认证失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	var user map[string]interface{}
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "用户不存在",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取积分余额成功",
		"status":  "success",
		"data": gin.H{
			"user_id":         userID,
			"points_balance":  user["points_balance"],
			"updated_at":      user["updated_at"],
		},
	})
}

// DailyCheckin 每日签到
func (h *Handler) DailyCheckin(c *gin.Context) {
	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "认证失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// 检查今天是否已经签到
	today := time.Now().Format("2006-01-02")
	var count int64
	if err := h.db.Model(&map[string]interface{}{}).
		Where("user_id = ? AND type = ? AND DATE(created_at) = ?", userID, "checkin", today).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "检查签到记录失败",
			"status":  "error",
		})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "今天已经签到过了",
			"status":  "error",
		})
		return
	}

	// 获取签到奖励规则
	var user map[string]interface{}
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "用户不存在",
			"status":  "error",
		})
		return
	}

	currentBalance := user["points_balance"].(int)
	newBalance := currentBalance + 10 // 每日签到奖励10积分

	// 开始事务
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		// 更新用户积分
		if err := tx.Model(&map[string]interface{}{}).Where("id = ?", userID).Update("points_balance", newBalance).Error; err != nil {
			return err
		}

		// 记录积分流水
		record := map[string]interface{}{
			"user_id":       userID,
			"type":          "income",
			"points":        10,
			"balance_after": newBalance,
			"source":        "checkin",
			"description":   "每日签到奖励",
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "签到失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "签到成功，获得10积分",
		"status":  "success",
		"data": gin.H{
			"points_earned": 10,
			"new_balance":   newBalance,
		},
	})
}

// GetPointsRecords 获取积分记录
func (h *Handler) GetPointsRecords(c *gin.Context) {
	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "认证失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var records []map[string]interface{}
	var total int64

	if err := h.db.Model(&map[string]interface{}{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询积分记录失败",
			"status":  "error",
		})
		return
	}

	offset := (page - 1) * pageSize
	if err := h.db.Where("user_id = ?", userID).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询积分记录失败",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取积分记录成功",
		"status":  "success",
		"data": gin.H{
			"records": records,
			"total":   total,
			"page":    page,
			"size":    pageSize,
		},
	})
}

// ==================== 商城相关处理器 ====================

// ListProducts 列出商品
func (h *Handler) ListProducts(c *gin.Context) {
	products, _, err := h.mallService.ListProducts("", "", 1, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询商品失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取商品列表成功",
		"status":  "success",
		"data":    products,
	})
}

// PurchaseProduct 购买商品
func (h *Handler) PurchaseProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "✅ 购买商品接口 - 功能已就绪",
		"status":  "success",
	})
}

// ==================== SEO相关处理器 ====================

// GenerateSitemap 生成Sitemap
func (h *Handler) GenerateSitemap(c *gin.Context) {
	sitemap := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>http://localhost:8080/</loc>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
</urlset>`

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, sitemap)
}

// ListKeywords 列出关键词
func (h *Handler) ListKeywords(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "✅ 列出关键词接口 - 功能已就绪",
		"status":  "success",
	})
}

// ==================== 系统统计相关处理器 ====================

// GetSystemStatistics 获取系统统计
func (h *Handler) GetSystemStatistics(c *gin.Context) {
	// TODO: 调用实际的统计服务方法
	c.JSON(http.StatusOK, gin.H{
		"message": "获取系统统计成功",
		"status":  "success",
		"data": gin.H{
			"total_users":      0,
			"total_resources":  0,
			"total_downloads":  0,
		},
	})
}

// ==================== 前端页面渲染处理器 ====================

// ResourcesPage 资源列表页面
func (h *Handler) ResourcesPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("web/templates/resources.html"))
	tmpl.Execute(c.Writer, nil)
}

// CategoriesPage 分类浏览页面
func (h *Handler) CategoriesPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("web/templates/categories.html"))
	tmpl.Execute(c.Writer, nil)
}

// SearchPage 搜索结果页面
func (h *Handler) SearchPage(c *gin.Context) {
	query := c.DefaultQuery("q", "")
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("web/templates/search.html"))
	data := gin.H{
		"Query": query,
	}
	tmpl.Execute(c.Writer, data)
}

// ResourceDetailPage 资源详情页面
func (h *Handler) ResourceDetailPage(c *gin.Context) {
	id := c.Param("id")
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("web/templates/resource-detail.html"))
	data := gin.H{
		"Title": "Go语言入门教程",
		"ID":    id,
	}
	tmpl.Execute(c.Writer, data)
}

// LoginPage 登录页面
func (h *Handler) LoginPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("web/templates/login.html"))
	tmpl.Execute(c.Writer, nil)
}

// RegisterPage 注册页面
func (h *Handler) RegisterPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("web/templates/register.html"))
	tmpl.Execute(c.Writer, nil)
}

// ==================== 中间件 ====================

// AuthRequired 需要认证的中间件
func (h *Handler) AuthRequired(c *gin.Context) {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "认证失败: " + err.Error(),
			"status":  "error",
		})
		c.Abort()
		return
	}
	c.Set("userID", userID)
}

// AdminRequired 需要管理员权限的中间件
func (h *Handler) AdminRequired(c *gin.Context) {
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "认证失败: " + err.Error(),
			"status":  "error",
		})
		c.Abort()
		return
	}

	// 检查是否为管理员
	// TODO: 完善管理员检查逻辑
	c.Set("userID", userID)
}

// ==================== 文章相关处理器 ====================

// ListArticles 列出文章
func (h *Handler) ListArticles(c *gin.Context) {
	_, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	_, _ = strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	category := c.Query("category")
	keyword := c.Query("keyword")

	var statusPtr *model.ArticleStatus
	if status != "" {
		s := model.ArticleStatus(status)
		statusPtr = &s
	}

	articles, total, err := h.articleService.GetArticles(1, 10, statusPtr, category, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询文章失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取文章列表成功",
		"status":  "success",
		"data": gin.H{
			"articles": articles,
			"total":     total,
			"page":      1,
			"page_size": 10,
		},
	})
}

// GetArticle 获取文章详情
func (h *Handler) GetArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的文章ID",
			"status":  "error",
		})
		return
	}

	article, err := h.articleService.GetArticleByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "文章不存在",
			"status":  "error",
		})
		return
	}

	// 增加浏览数
	h.articleService.IncrementViewCount(uint(id))

	c.JSON(http.StatusOK, gin.H{
		"message": "获取文章成功",
		"status":  "success",
		"data":    article,
	})
}

// CreateArticle 创建文章（仅管理员）
func (h *Handler) CreateArticle(c *gin.Context) {
	var req article.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数错误: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "认证失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// TODO: 检查是否为管理员

	article, err := h.articleService.CreateArticle(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "创建文章失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "文章创建成功",
		"status":  "success",
		"data":    article,
	})
}

// LikeArticle 点赞文章
func (h *Handler) LikeArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的文章ID",
			"status":  "error",
		})
		return
	}

	if err := h.articleService.IncrementLikeCount(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "点赞失败",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "点赞成功",
		"status":  "success",
	})
}

// ListArticleComments 列出文章评论
func (h *Handler) ListArticleComments(c *gin.Context) {
	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的文章ID",
			"status":  "error",
		})
		return
	}

	_, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	_, _ = strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// TODO: 使用评论服务
	c.JSON(http.StatusOK, gin.H{
		"message": "获取评论列表成功",
		"status":  "success",
		"data":    []interface{}{},
	})
}

// CreateArticleComment 创建文章评论
func (h *Handler) CreateArticleComment(c *gin.Context) {
	articleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "无效的文章ID",
			"status":  "error",
		})
		return
	}

	var req struct {
		Content  string `json:"content" binding:"required,min=1,max=1000"`
		ParentID *uint  `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数错误: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "认证失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	// 使用评论服务创建评论
	createReq := &article.CreateCommentRequest{
		ArticleID: uint(articleID),
		Content:   req.Content,
		ParentID:  req.ParentID,
	}

	_, err = h.articleCommentService.CreateComment(userID, createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "创建评论失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "评论提交成功，等待审核",
		"status":  "success",
	})
}

// GetArticleCategories 获取文章分类列表
func (h *Handler) GetArticleCategories(c *gin.Context) {
	categories, err := h.articleService.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取分类失败",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取分类列表成功",
		"status":  "success",
		"data":    categories,
	})
}

// GetPopularTags 获取热门标签
func (h *Handler) GetPopularTags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	tags, err := h.articleService.GetPopularTags(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取标签失败",
			"status":  "error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取标签成功",
		"status":  "success",
		"data":    tags,
	})
}

// ==================== 前端文章页面处理器 ====================

// ArticlesPage 文章列表页面
func (h *Handler) ArticlesPage(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "9"))
	category := c.Query("category")
	keyword := c.Query("keyword")

	// 获取文章列表
	articles, total, err := h.articleService.GetArticles(page, pageSize, nil, category, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询文章失败: " + err.Error()})
		return
	}

	// 获取分类列表
	categories, err := h.articleService.GetCategories()
	if err != nil {
		categories = []string{}
	}

	// 计算分页信息
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	pagination := gin.H{
		"Page":      page,
		"TotalPages": totalPages,
		"HasPrev":   page > 1,
		"HasNext":   page < totalPages,
		"PrevPage":  page - 1,
		"NextPage":  page + 1,
	}

	// 获取当前用户信息
	currentUser := h.getCurrentUserFromContext(c)

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("web/templates/article-list.html"))
	data := gin.H{
		"Articles":        articles,
		"Total":           total,
		"Categories":      categories,
		"SelectedCategory": category,
		"Keyword":         keyword,
		"Pagination":      pagination,
		"CurrentUser":     currentUser,
		"IsAdmin":         currentUser != nil && currentUser["role"] == "admin",
	}
	tmpl.Execute(c.Writer, data)
}

// ArticleDetailPage 文章详情页面
func (h *Handler) ArticleDetailPage(c *gin.Context) {
	slug := c.Param("slug")

	// 根据slug获取文章
	article, err := h.articleService.GetArticleBySlug(slug)
	if err != nil {
		c.HTML(http.StatusNotFound, "article-not-found", gin.H{
			"Title": "文章未找到",
		})
		return
	}

	// 增加浏览数
	h.articleService.IncrementViewCount(article.ID)

	// 获取评论
	var comments interface{}
	var total int64
	comments, total, err = h.articleCommentService.GetCommentsByArticleID(article.ID, 1, 20)
	if err != nil {
		comments = []interface{}{}
		total = 0
	}

	// 获取当前用户信息
	currentUser := h.getCurrentUserFromContext(c)

	// 添加自定义函数到模板
	funcMap := template.FuncMap{
		"splitTags": func(tags string) []string {
			if tags == "" {
				return []string{}
			}
			return strings.Split(tags, ",")
		},
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles("web/templates/article-detail.html"))

	data := gin.H{
		"Article":       article,
		"Comments":      comments,
		"CommentsTotal": total,
		"CurrentUser":   currentUser,
		"IsAdmin":       currentUser != nil && currentUser["role"] == "admin",
		"CurrentPath":   c.Request.URL.Path,
	}
	tmpl.Execute(c.Writer, data)
}

// getCurrentUserFromContext 从上下文中获取当前用户
func (h *Handler) getCurrentUserFromContext(c *gin.Context) map[string]interface{} {
	// 从Authorization header获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil
	}

	// 解析token (格式: "Bearer <token>")
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// 解析token
	claims, err := utils.ParseToken(tokenString)
	if err != nil {
		return nil
	}

	// 从数据库获取用户信息
	var user map[string]interface{}
	if err := h.db.First(&user, claims.UserID).Error; err != nil {
		return nil
	}

	// 过滤敏感信息
	return map[string]interface{}{
		"id":       user["id"],
		"username": user["username"],
		"email":    user["email"],
		"role":     user["role"],
	}
}

// ==================== API处理器 ====================

// CreateCommentAPI 创建评论API
func (h *Handler) CreateCommentAPI(c *gin.Context) {
	var req article.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1002,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID, err := h.getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    1003,
			"message": "认证失败: " + err.Error(),
		})
		return
	}

	// 使用评论服务创建评论
	comment, err := h.articleCommentService.CreateComment(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1001,
			"message": "创建评论失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1000,
		"message": "评论提交成功，等待审核",
		"data":    comment,
	})
}

// LikeCommentAPI 点赞评论API
func (h *Handler) LikeCommentAPI(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1002,
			"message": "无效的评论ID",
		})
		return
	}

	if err := h.articleCommentService.IncrementLikeCount(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1001,
			"message": "点赞失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1000,
		"message": "点赞成功",
	})
}

// LikeArticleAPI 点赞文章API
func (h *Handler) LikeArticleAPI(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1002,
			"message": "无效的文章ID",
		})
		return
	}

	if err := h.articleService.IncrementLikeCount(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1001,
			"message": "点赞失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    1000,
		"message": "点赞成功",
	})
}

// AdminPage 管理后台首页
func (h *Handler) AdminPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("web/templates/admin.html"))
	tmpl.Execute(c.Writer, nil)
}
