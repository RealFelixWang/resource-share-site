/*
SEO Configuration Service - SEO配置服务

提供SEO配置管理功能，包括：
- SEO配置管理
- Meta标签生成
- Sitemap生成
- SEO优化建议

Author: Felix Wang
Email: felixwang.biz@gmail.com
Date: 2025-10-31
*/

package seo

import (
	"fmt"
	"strings"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// ConfigService SEO配置服务
type ConfigService struct {
	db *gorm.DB
}

// NewConfigService 创建新的SEO配置服务
func NewConfigService(db *gorm.DB) *ConfigService {
	return &ConfigService{
		db: db,
	}
}

// CreateSEOConfig 创建SEO配置
func (s *ConfigService) CreateSEOConfig(config *model.SEOConfig) error {
	if config.ConfigType == "" {
		return fmt.Errorf("配置类型不能为空")
	}

	// 设置默认值
	if config.Robots == "" {
		config.Robots = "index, follow"
	}
	if config.ChangeFreq == "" {
		config.ChangeFreq = "weekly"
	}
	if config.Priority == 0 {
		config.Priority = 0.5
	}

	if err := s.db.Create(config).Error; err != nil {
		return fmt.Errorf("创建SEO配置失败: %w", err)
	}

	return nil
}

// UpdateSEOConfig 更新SEO配置
func (s *ConfigService) UpdateSEOConfig(configID uint, updates map[string]interface{}) error {
	if err := s.db.Model(&model.SEOConfig{}).
		Where("id = ?", configID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("更新SEO配置失败: %w", err)
	}

	return nil
}

// DeleteSEOConfig 删除SEO配置
func (s *ConfigService) DeleteSEOConfig(configID uint) error {
	if err := s.db.Delete(&model.SEOConfig{}, configID).Error; err != nil {
		return fmt.Errorf("删除SEO配置失败: %w", err)
	}

	return nil
}

// GetSEOConfig 获取SEO配置
func (s *ConfigService) GetSEOConfig(configType model.SEOConfigType, targetID *uint) (*model.SEOConfig, error) {
	var config model.SEOConfig

	query := s.db.Where("config_type = ? AND is_active = ?", configType, true)

	if targetID != nil {
		query = query.Where("target_id = ?", *targetID)
	} else {
		query = query.Where("target_id IS NULL")
	}

	if err := query.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 没有找到配置
		}
		return nil, fmt.Errorf("查询SEO配置失败: %w", err)
	}

	return &config, nil
}

// ListSEOConfigs 获取SEO配置列表
func (s *ConfigService) ListSEOConfigs(configType model.SEOConfigType, page, pageSize int) ([]model.SEOConfig, int64, error) {
	var configs []model.SEOConfig
	var total int64

	query := s.db.Model(&model.SEOConfig{})

	// 按类型筛选
	if configType != "" {
		query = query.Where("config_type = ?", configType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取配置总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("获取配置列表失败: %w", err)
	}

	return configs, total, nil
}

// GenerateMetaTags 生成Meta标签
func (s *ConfigService) GenerateMetaTags(configType model.SEOConfigType, targetID *uint, context map[string]interface{}) (map[string]string, error) {
	metaTags := make(map[string]string)

	// 获取SEO配置
	config, err := s.GetSEOConfig(configType, targetID)
	if err != nil {
		return nil, fmt.Errorf("获取SEO配置失败: %w", err)
	}

	// 如果没有配置，返回默认值
	if config == nil {
		return s.getDefaultMetaTags(configType, context)
	}

	// 解析模板
	title := s.parseTemplate(config.MetaTitle, context)
	description := s.parseTemplate(config.MetaDescription, context)
	keywords := s.parseTemplate(config.MetaKeywords, context)
	author := s.parseTemplate(config.MetaAuthor, context)

	// 设置Meta标签
	if title != "" {
		metaTags["title"] = title
		metaTags["og:title"] = title
		metaTags["twitter:title"] = title
	}

	if description != "" {
		metaTags["description"] = description
		metaTags["og:description"] = description
		metaTags["twitter:description"] = description
	}

	if keywords != "" {
		metaTags["keywords"] = keywords
	}

	if author != "" {
		metaTags["author"] = author
	}

	// Open Graph标签
	if config.OGTitle != "" {
		metaTags["og:title"] = s.parseTemplate(config.OGTitle, context)
	}
	if config.OGDescription != "" {
		metaTags["og:description"] = s.parseTemplate(config.OGDescription, context)
	}
	if config.OGImage != "" {
		metaTags["og:image"] = config.OGImage
	}
	if config.OGType != "" {
		metaTags["og:type"] = config.OGType
	} else {
		metaTags["og:type"] = "website"
	}
	if config.OGUrl != "" {
		metaTags["og:url"] = s.parseTemplate(config.OGUrl, context)
	}

	// Twitter Card标签
	if config.TwitterCard != "" {
		metaTags["twitter:card"] = config.TwitterCard
	} else {
		metaTags["twitter:card"] = "summary_large_image"
	}
	if config.TwitterTitle != "" {
		metaTags["twitter:title"] = s.parseTemplate(config.TwitterTitle, context)
	}
	if config.TwitterDescription != "" {
		metaTags["twitter:description"] = s.parseTemplate(config.TwitterDescription, context)
	}
	if config.TwitterImage != "" {
		metaTags["twitter:image"] = config.TwitterImage
	}

	// 其他重要标签
	if config.CanonicalURL != "" {
		metaTags["canonical"] = s.parseTemplate(config.CanonicalURL, context)
	}
	if config.Robots != "" {
		metaTags["robots"] = config.Robots
	}

	return metaTags, nil
}

// parseTemplate 解析模板字符串
func (s *ConfigService) parseTemplate(template string, context map[string]interface{}) string {
	if template == "" {
		return ""
	}

	// 简单的模板解析，支持 {{key}} 格式
	result := template

	// 这里可以实现更复杂的模板解析逻辑
	// 目前先做简单的字符串替换

	for key, value := range context {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}

// getDefaultMetaTags 获取默认Meta标签
func (s *ConfigService) getDefaultMetaTags(configType model.SEOConfigType, context map[string]interface{}) (map[string]string, error) {
	metaTags := make(map[string]string)

	// 根据类型设置默认标签
	switch configType {
	case model.SEOConfigTypeHome:
		metaTags["title"] = "资源分享网站 - 优质资源一站式分享平台"
		metaTags["description"] = "提供优质的技术资源、学习资料、工具软件等，助力您的学习和工作"
		metaTags["keywords"] = "资源分享,技术资料,学习资源,工具软件"
		metaTags["author"] = "资源分享网站"
		metaTags["robots"] = "index, follow"

	case model.SEOConfigTypeResource:
		if title, ok := context["title"].(string); ok {
			metaTags["title"] = fmt.Sprintf("%s - 资源分享网站", title)
		} else {
			metaTags["title"] = "资源详情 - 资源分享网站"
		}
		metaTags["description"] = "下载优质资源，提升您的技能和效率"
		metaTags["robots"] = "index, follow"

	case model.SEOConfigTypeCategory:
		if name, ok := context["name"].(string); ok {
			metaTags["title"] = fmt.Sprintf("%s - 分类浏览 - 资源分享网站", name)
			metaTags["description"] = fmt.Sprintf("浏览 %s 分类下的优质资源", name)
			metaTags["keywords"] = name + ",资源分类"
		}
		metaTags["robots"] = "index, follow"

	default:
		metaTags["title"] = "资源分享网站"
		metaTags["description"] = "优质资源一站式分享平台"
		metaTags["robots"] = "index, follow"
	}

	// 设置通用标签
	metaTags["og:type"] = "website"
	metaTags["twitter:card"] = "summary_large_image"

	return metaTags, nil
}

// GenerateSitemap 生成Sitemap
func (s *ConfigService) GenerateSitemap(baseURL string) (string, error) {
	var urls []model.SitemapUrl

	if err := s.db.Where("is_active = ?", true).
		Order("priority DESC, updated_at DESC").
		Find(&urls).Error; err != nil {
		return "", fmt.Errorf("查询Sitemap URL失败: %w", err)
	}

	var sitemap strings.Builder
	sitemap.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sitemap.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	for _, url := range urls {
		loc := url.Loc
		if !strings.HasPrefix(loc, "http") {
			loc = baseURL + loc
		}

		sitemap.WriteString(`  <url>` + "\n")
		sitemap.WriteString(fmt.Sprintf(`    <loc>%s</loc>`+"\n", loc))

		if url.LastMod != nil {
			sitemap.WriteString(fmt.Sprintf(`    <lastmod>%s</lastmod>`+"\n", url.LastMod.Format("2006-01-02")))
		}

		if url.ChangeFreq != "" {
			sitemap.WriteString(fmt.Sprintf(`    <changefreq>%s</changefreq>`+"\n", url.ChangeFreq))
		}

		if url.Priority > 0 {
			sitemap.WriteString(fmt.Sprintf(`    <priority>%.1f</priority>`+"\n", url.Priority))
		}

		sitemap.WriteString(`  </url>` + "\n")
	}

	sitemap.WriteString(`</urlset>`)

	return sitemap.String(), nil
}

// AddSitemapUrl 添加Sitemap URL
func (s *ConfigService) AddSitemapUrl(url *model.SitemapUrl) error {
	// 检查URL是否已存在
	var count int64
	s.db.Model(&model.SitemapUrl{}).
		Where("loc = ? AND page_type = ?", url.Loc, url.PageType).
		Count(&count)

	if count > 0 {
		return fmt.Errorf("该URL已存在于Sitemap中")
	}

	if err := s.db.Create(url).Error; err != nil {
		return fmt.Errorf("添加Sitemap URL失败: %w", err)
	}

	return nil
}

// UpdateSitemapUrl 更新Sitemap URL
func (s *ConfigService) UpdateSitemapUrl(urlID uint, updates map[string]interface{}) error {
	if err := s.db.Model(&model.SitemapUrl{}).
		Where("id = ?", urlID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("更新Sitemap URL失败: %w", err)
	}

	return nil
}

// DeleteSitemapUrl 删除Sitemap URL
func (s *ConfigService) DeleteSitemapUrl(urlID uint) error {
	if err := s.db.Delete(&model.SitemapUrl{}, urlID).Error; err != nil {
		return fmt.Errorf("删除Sitemap URL失败: %w", err)
	}

	return nil
}

// GetSitemapUrls 获取Sitemap URL列表
func (s *ConfigService) GetSitemapUrls(pageType model.SEOConfigType, page, pageSize int) ([]model.SitemapUrl, int64, error) {
	var urls []model.SitemapUrl
	var total int64

	query := s.db.Model(&model.SitemapUrl{})

	// 按类型筛选
	if pageType != "" {
		query = query.Where("page_type = ?", pageType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取URL总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("priority DESC, created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&urls).Error; err != nil {
		return nil, 0, fmt.Errorf("获取URL列表失败: %w", err)
	}

	return urls, total, nil
}

// AutoGenerateSitemap 自动生成Sitemap
func (s *ConfigService) AutoGenerateSitemap(baseURL string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 清空现有Sitemap URL
		if err := tx.Exec("DELETE FROM sitemap_urls WHERE page_type IN ('resource', 'category')").Error; err != nil {
			return fmt.Errorf("清空Sitemap URL失败: %w", err)
		}

		// 2. 添加资源页面
		var resources []model.Resource
		if err := tx.Where("status = ?", "approved").
			Select("id, title, updated_at").
			Find(&resources).Error; err != nil {
			return fmt.Errorf("查询资源失败: %w", err)
		}

		for _, resource := range resources {
			url := model.SitemapUrl{
				Loc:        fmt.Sprintf("/resource/%d", resource.ID),
				LastMod:    &resource.UpdatedAt,
				ChangeFreq: "weekly",
				Priority:   0.8,
				PageType:   model.SEOConfigTypeResource,
				TargetID:   &resource.ID,
				IsActive:   true,
			}
			if err := tx.Create(&url).Error; err != nil {
				return fmt.Errorf("添加资源URL失败: %w", err)
			}
		}

		// 3. 添加分类页面
		var categories []model.Category
		if err := tx.Select("id, name, updated_at").
			Find(&categories).Error; err != nil {
			return fmt.Errorf("查询分类失败: %w", err)
		}

		for _, category := range categories {
			url := model.SitemapUrl{
				Loc:        fmt.Sprintf("/category/%d", category.ID),
				LastMod:    &category.UpdatedAt,
				ChangeFreq: "daily",
				Priority:   0.9,
				PageType:   model.SEOConfigTypeCategory,
				TargetID:   &category.ID,
				IsActive:   true,
			}
			if err := tx.Create(&url).Error; err != nil {
				return fmt.Errorf("添加分类URL失败: %w", err)
			}
		}

		// 4. 添加首页
		now := time.Now()
		homeURL := model.SitemapUrl{
			Loc:        "/",
			LastMod:    &now,
			ChangeFreq: "daily",
			Priority:   1.0,
			PageType:   model.SEOConfigTypeHome,
			IsActive:   true,
		}
		if err := tx.Create(&homeURL).Error; err != nil {
			return fmt.Errorf("添加首页URL失败: %w", err)
		}

		return nil
	})
}

// OptimizeSEO SEO优化建议
func (s *ConfigService) OptimizeSEO(pageType model.SEOConfigType, targetID *uint) ([]string, error) {
	var suggestions []string

	// 获取当前配置
	config, err := s.GetSEOConfig(pageType, targetID)
	if err != nil {
		return nil, fmt.Errorf("获取SEO配置失败: %w", err)
	}

	// 检查基本Meta标签
	if config == nil || config.MetaTitle == "" {
		suggestions = append(suggestions, "建议设置页面标题（Meta Title）")
	}

	if config == nil || config.MetaDescription == "" {
		suggestions = append(suggestions, "建议设置页面描述（Meta Description）")
	}

	if config == nil || config.MetaKeywords == "" {
		suggestions = append(suggestions, "建议设置关键词（Meta Keywords）")
	}

	// 检查Open Graph标签
	if config == nil || config.OGTitle == "" {
		suggestions = append(suggestions, "建议设置Open Graph标题")
	}

	if config == nil || config.OGDescription == "" {
		suggestions = append(suggestions, "建议设置Open Graph描述")
	}

	if config == nil || config.OGImage == "" {
		suggestions = append(suggestions, "建议设置Open Graph图片")
	}

	// 检查Twitter Card
	if config == nil || config.TwitterCard == "" {
		suggestions = append(suggestions, "建议设置Twitter Card")
	}

	// 检查结构化数据
	if config == nil || config.StructuredData == "" {
		suggestions = append(suggestions, "建议添加结构化数据（JSON-LD）")
	}

	// 检查规范URL
	if config == nil || config.CanonicalURL == "" {
		suggestions = append(suggestions, "建议设置规范URL（Canonical URL）")
	}

	// 检查图片优化
	if pageType == model.SEOConfigTypeResource || pageType == model.SEOConfigTypeDetail {
		suggestions = append(suggestions, "建议优化图片，使用alt标签和适当的大小")
	}

	return suggestions, nil
}
