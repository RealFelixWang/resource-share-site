/*
Package article provides article and blog services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package article

import (
	"errors"
	"strings"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// ArticleService 文章服务
type ArticleService struct {
	db *gorm.DB
}

// NewArticleService 创建文章服务实例
func NewArticleService(db *gorm.DB) *ArticleService {
	return &ArticleService{
		db: db,
	}
}

// CreateArticleRequest 创建文章请求
type CreateArticleRequest struct {
	Title          string `json:"title" binding:"required,min=1,max=200"`
	Content        string `json:"content" binding:"required"`
	Excerpt        string `json:"excerpt"`
	FeaturedImage  string `json:"featured_image"`
	Tags           string `json:"tags"`
	Category       string `json:"category" binding:"required"`
	Status         model.ArticleStatus `json:"status"`
	MetaTitle      string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords   string `json:"meta_keywords"`
}

// UpdateArticleRequest 更新文章请求
type UpdateArticleRequest struct {
	Title          string `json:"title" binding:"required,min=1,max=200"`
	Content        string `json:"content" binding:"required"`
	Excerpt        string `json:"excerpt"`
	FeaturedImage  string `json:"featured_image"`
	Tags           string `json:"tags"`
	Category       string `json:"category" binding:"required"`
	Status         model.ArticleStatus `json:"status"`
	MetaTitle      string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
	MetaKeywords   string `json:"meta_keywords"`
}

// ArticleListItem 文章列表项
type ArticleListItem struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Excerpt     string    `json:"excerpt"`
	FeaturedImage string  `json:"featured_image"`
	Tags        string    `json:"tags"`
	Category    string    `json:"category"`
	Status      model.ArticleStatus `json:"status"`
	PublishedAt *time.Time `json:"published_at"`
	ViewCount   uint      `json:"view_count"`
	LikeCount   uint      `json:"like_count"`
	CommentCount uint     `json:"comment_count"`
	CreatedAt   time.Time `json:"created_at"`

	// 作者信息
	AuthorID   uint   `json:"author_id"`
	AuthorName string `json:"author_name"`
}

// CreateArticle 创建文章
func (s *ArticleService) CreateArticle(authorID uint, req *CreateArticleRequest) (*model.Article, error) {
	// 生成slug
	slug := s.generateSlug(req.Title)

	// 检查slug是否已存在
	if err := s.checkSlugExists(slug); err != nil {
		return nil, err
	}

	// 创建文章
	article := &model.Article{
		Title:          req.Title,
		Slug:           slug,
		Content:        req.Content,
		Excerpt:        req.Excerpt,
		FeaturedImage:  req.FeaturedImage,
		Tags:           req.Tags,
		Category:       req.Category,
		Status:         req.Status,
		AuthorID:       authorID,
		MetaTitle:      req.MetaTitle,
		MetaDescription: req.MetaDescription,
		MetaKeywords:   req.MetaKeywords,
	}

	// 如果状态是已发布，设置发布时间
	if req.Status == model.ArticleStatusPublished {
		now := time.Now()
		article.PublishedAt = &now
	}

	if err := s.db.Create(article).Error; err != nil {
		return nil, err
	}

	return article, nil
}

// GetArticleByID 根据ID获取文章
func (s *ArticleService) GetArticleByID(id uint) (*model.Article, error) {
	var article model.Article
	if err := s.db.Preload("Author").First(&article, id).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

// GetArticleBySlug 根据slug获取文章
func (s *ArticleService) GetArticleBySlug(slug string) (*model.Article, error) {
	var article model.Article
	if err := s.db.Preload("Author").Where("slug = ?", slug).First(&article).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

// UpdateArticle 更新文章
func (s *ArticleService) UpdateArticle(id uint, req *UpdateArticleRequest) (*model.Article, error) {
	article, err := s.GetArticleByID(id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	article.Title = req.Title
	article.Content = req.Content
	article.Excerpt = req.Excerpt
	article.FeaturedImage = req.FeaturedImage
	article.Tags = req.Tags
	article.Category = req.Category
	article.Status = req.Status
	article.MetaTitle = req.MetaTitle
	article.MetaDescription = req.MetaDescription
	article.MetaKeywords = req.MetaKeywords

	// 如果状态从未发布变为已发布，设置发布时间
	if req.Status == model.ArticleStatusPublished && article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
	}

	if err := s.db.Save(article).Error; err != nil {
		return nil, err
	}

	return article, nil
}

// DeleteArticle 删除文章（软删除）
func (s *ArticleService) DeleteArticle(id uint) error {
	if err := s.db.Delete(&model.Article{}, id).Error; err != nil {
		return err
	}
	return nil
}

// PublishArticle 发布文章
func (s *ArticleService) PublishArticle(id uint) (*model.Article, error) {
	article, err := s.GetArticleByID(id)
	if err != nil {
		return nil, err
	}

	article.Status = model.ArticleStatusPublished
	if article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
	}

	if err := s.db.Save(article).Error; err != nil {
		return nil, err
	}

	return article, nil
}

// UnpublishArticle 取消发布文章
func (s *ArticleService) UnpublishArticle(id uint) (*model.Article, error) {
	article, err := s.GetArticleByID(id)
	if err != nil {
		return nil, err
	}

	article.Status = model.ArticleStatusDraft

	if err := s.db.Save(article).Error; err != nil {
		return nil, err
	}

	return article, nil
}

// IncrementViewCount 增加浏览数
func (s *ArticleService) IncrementViewCount(id uint) error {
	return s.db.Model(&model.Article{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// IncrementLikeCount 增加点赞数
func (s *ArticleService) IncrementLikeCount(id uint) error {
	return s.db.Model(&model.Article{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// DecrementLikeCount 减少点赞数
func (s *ArticleService) DecrementLikeCount(id uint) error {
	return s.db.Model(&model.Article{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
}

// GetArticles 获取文章列表
func (s *ArticleService) GetArticles(page, pageSize int, status *model.ArticleStatus, category, keyword string) ([]ArticleListItem, int64, error) {
	var articles []ArticleListItem
	var total int64

	// 构建查询
	query := s.db.Table("articles").
		Select(`
			articles.*,
			users.username as author_name
		`).
		Joins("LEFT JOIN users ON articles.author_id = users.id")

	// 状态过滤
	if status != nil && *status != "" {
		query = query.Where("articles.status = ?", *status)
	} else {
		// 默认只显示已发布的文章
		query = query.Where("articles.status = ?", model.ArticleStatusPublished)
	}

	// 分类过滤
	if category != "" {
		query = query.Where("articles.category = ?", category)
	}

	// 关键词搜索
	if keyword != "" {
		query = query.Where("articles.title LIKE ? OR articles.content LIKE ? OR articles.tags LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("articles.created_at DESC").
		Scan(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// GetArticlesByAuthor 根据作者获取文章列表
func (s *ArticleService) GetArticlesByAuthor(authorID uint, page, pageSize int) ([]ArticleListItem, int64, error) {
	var articles []ArticleListItem
	var total int64

	query := s.db.Table("articles").
		Select(`
			articles.*,
			users.username as author_name
		`).
		Joins("LEFT JOIN users ON articles.author_id = users.id").
		Where("articles.author_id = ?", authorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("articles.created_at DESC").
		Scan(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// GetCategories 获取所有文章分类
func (s *ArticleService) GetCategories() ([]string, error) {
	var categories []string
	if err := s.db.Model(&model.Article{}).
		Where("status = ?", model.ArticleStatusPublished).
		Distinct("category").
		Pluck("category", &categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// GetPopularTags 获取热门标签
func (s *ArticleService) GetPopularTags(limit int) ([]string, error) {
	var tags []string

	// 获取所有已发布文章的标签
	var articles []model.Article
	if err := s.db.Where("status = ? AND tags IS NOT NULL", model.ArticleStatusPublished).
		Find(&articles).Error; err != nil {
		return nil, err
	}

	// 统计标签频率
	tagCount := make(map[string]int)
	for _, article := range articles {
		if article.Tags != "" {
			// 解析标签（逗号分隔）
			tagList := strings.Split(article.Tags, ",")
			for _, tag := range tagList {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tagCount[tag]++
				}
			}
		}
	}

	// 排序并返回热门标签
	type tagFreq struct {
		Tag  string
		Count int
	}
	var tagFreqs []tagFreq
	for tag, count := range tagCount {
		tagFreqs = append(tagFreqs, tagFreq{Tag: tag, Count: count})
	}

	// 排序
	for i := 0; i < len(tagFreqs); i++ {
		for j := i + 1; j < len(tagFreqs); j++ {
			if tagFreqs[i].Count < tagFreqs[j].Count {
				tagFreqs[i], tagFreqs[j] = tagFreqs[j], tagFreqs[i]
			}
		}
	}

	// 返回指定数量的标签
	for i := 0; i < len(tagFreqs) && i < limit; i++ {
		tags = append(tags, tagFreqs[i].Tag)
	}

	return tags, nil
}

// generateSlug 生成URL友好的slug
func (s *ArticleService) generateSlug(title string) string {
	// 将中文标题转换为拼音（简单实现：使用标题的拼音首字母或英文部分）
	slug := strings.ToLower(title)

	// 替换特殊字符
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// 移除特殊字符
	allowed := "abcdefghijklmnopqrstuvwxyz0123456789-"
	var result strings.Builder
	for _, char := range slug {
		if strings.ContainsRune(allowed, char) {
			result.WriteRune(char)
		} else {
			result.WriteRune('-')
		}
	}

	return result.String()
}

// checkSlugExists 检查slug是否已存在
func (s *ArticleService) checkSlugExists(slug string) error {
	var count int64
	if err := s.db.Model(&model.Article{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.New("slug已存在")
	}

	return nil
}
