/*
SEO Middleware - SEO中间件

在HTTP请求处理过程中自动生成和设置SEO meta标签

Author: Felix Wang
Email: felixwang.biz@gmail.com
Date: 2025-10-31
*/

package seo

import (
	"fmt"
	"strconv"
	"strings"

	"resource-share-site/internal/model"

	"github.com/gin-gonic/gin"
)

// SEOContext SEO上下文
type SEOContext struct {
	PageType    model.SEOConfigType
	TargetID    *uint
	Title       string
	Name        string
	Category    string
	Resource    *model.Resource
	CategoryObj *model.Category
	CustomData  map[string]interface{}
}

// Middleware SEO中间件
type Middleware struct {
	configService *ConfigService
}

// NewMiddleware 创建SEO中间件
func NewMiddleware(configService *ConfigService) *Middleware {
	return &Middleware{
		configService: configService,
	}
}

// SEOMiddleware SEO中间件处理函数
func (m *Middleware) SEOMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析SEO上下文
		seoCtx := m.parseSEOContext(c)

		// 生成Meta标签
		metaTags, err := m.configService.GenerateMetaTags(
			seoCtx.PageType,
			seoCtx.TargetID,
			m.buildContextData(seoCtx),
		)

		if err == nil && len(metaTags) > 0 {
			// 将Meta标签存储到上下文
			c.Set("seo_meta_tags", metaTags)
		}

		c.Next()
	}
}

// parseSEOContext 解析SEO上下文
func (m *Middleware) parseSEOContext(c *gin.Context) *SEOContext {
	path := c.Request.URL.Path
	seoCtx := &SEOContext{
		CustomData: make(map[string]interface{}),
	}

	// 根据URL路径确定页面类型
	switch {
	case path == "/" || path == "/home":
		seoCtx.PageType = model.SEOConfigTypeHome

	case strings.HasPrefix(path, "/resource/"):
		seoCtx.PageType = model.SEOConfigTypeResource
		if id := m.extractIDFromPath(path, 2); id != nil {
			seoCtx.TargetID = id
			// 可以在这里加载资源对象
		}

	case strings.HasPrefix(path, "/category/"):
		seoCtx.PageType = model.SEOConfigTypeCategory
		if id := m.extractIDFromPath(path, 2); id != nil {
			seoCtx.TargetID = id
		}

	case strings.HasPrefix(path, "/list"):
		seoCtx.PageType = model.SEOConfigTypeList

	default:
		seoCtx.PageType = model.SEOConfigTypeDetail
	}

	// 从查询参数获取额外信息
	if title := c.Query("title"); title != "" {
		seoCtx.Title = title
	}
	if name := c.Query("name"); name != "" {
		seoCtx.Name = name
	}
	if category := c.Query("category"); category != "" {
		seoCtx.Category = category
	}

	// 解析分页参数
	if page := c.Query("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 1 {
			seoCtx.CustomData["page"] = pageNum
		}
	}

	return seoCtx
}

// extractIDFromPath 从路径中提取ID
func (m *Middleware) extractIDFromPath(path string, index int) *uint {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > index {
		if id, err := strconv.Atoi(parts[index]); err == nil && id > 0 {
			uintID := uint(id)
			return &uintID
		}
	}
	return nil
}

// buildContextData 构建上下文数据
func (m *Middleware) buildContextData(seoCtx *SEOContext) map[string]interface{} {
	data := make(map[string]interface{})

	// 基本信息
	if seoCtx.Title != "" {
		data["title"] = seoCtx.Title
	}
	if seoCtx.Name != "" {
		data["name"] = seoCtx.Name
	}
	if seoCtx.Category != "" {
		data["category"] = seoCtx.Category
	}

	// 目标ID
	if seoCtx.TargetID != nil {
		data["target_id"] = *seoCtx.TargetID
	}

	// 分页信息
	if page, ok := seoCtx.CustomData["page"]; ok {
		data["page"] = page
		data["page_suffix"] = fmt.Sprintf(" - 第%d页", page)
	}

	// 站点信息
	data["site_name"] = "资源分享网站"
	data["site_url"] = "https://example.com"

	// 时间信息
	data["current_year"] = "2025"

	return data
}

// logSEOCEvent 记录SEO事件（简化版）
func (m *Middleware) logSEOCEvent(path, clientIP string) {
	// 这里可以记录访问日志
	// 实际项目中可以发送到分析系统
}

// GetMetaTagsFromContext 从上下文中获取Meta标签
func GetMetaTagsFromContext(c *gin.Context) map[string]string {
	if tags, exists := c.Get("seo_meta_tags"); exists {
		if metaTags, ok := tags.(map[string]string); ok {
			return metaTags
		}
	}
	return make(map[string]string)
}

// RenderMetaTags 渲染Meta标签为HTML
func RenderMetaTags(metaTags map[string]string) string {
	if len(metaTags) == 0 {
		return ""
	}

	var builder strings.Builder

	// 渲染title
	if title, ok := metaTags["title"]; ok {
		builder.WriteString(fmt.Sprintf("<title>%s</title>\n", title))
	}

	// 渲染基本Meta标签
	if description, ok := metaTags["description"]; ok {
		builder.WriteString(fmt.Sprintf(`<meta name="description" content="%s">`, description) + "\n")
	}

	if keywords, ok := metaTags["keywords"]; ok {
		builder.WriteString(fmt.Sprintf(`<meta name="keywords" content="%s">`, keywords) + "\n")
	}

	if author, ok := metaTags["author"]; ok {
		builder.WriteString(fmt.Sprintf(`<meta name="author" content="%s">`, author) + "\n")
	}

	if robots, ok := metaTags["robots"]; ok {
		builder.WriteString(fmt.Sprintf(`<meta name="robots" content="%s">`, robots) + "\n")
	}

	// 渲染Open Graph标签
	for key, value := range metaTags {
		if strings.HasPrefix(key, "og:") {
			builder.WriteString(fmt.Sprintf(`<meta property="%s" content="%s">`, key, value) + "\n")
		}
	}

	// 渲染Twitter Card标签
	for key, value := range metaTags {
		if strings.HasPrefix(key, "twitter:") {
			builder.WriteString(fmt.Sprintf(`<meta name="%s" content="%s">`, key, value) + "\n")
		}
	}

	// 渲染Canonical URL
	if canonical, ok := metaTags["canonical"]; ok {
		builder.WriteString(fmt.Sprintf(`<link rel="canonical" href="%s">`, canonical) + "\n")
	}

	return builder.String()
}

// GenerateJSONLD 生成JSON-LD结构化数据
func GenerateJSONLD(seoCtx *SEOContext) string {
	var jsonld strings.Builder

	switch seoCtx.PageType {
	case model.SEOConfigTypeHome:
		jsonld.WriteString(`{
			"@context": "https://schema.org",
			"@type": "WebSite",
			"name": "资源分享网站",
			"url": "https://example.com",
			"description": "优质资源一站式分享平台",
			"potentialAction": {
				"@type": "SearchAction",
				"target": "https://example.com/search?q={search_term_string}",
				"query-input": "required name=search_term_string"
			}
		}`)

	case model.SEOConfigTypeResource:
		if seoCtx.Resource != nil {
			jsonld.WriteString(fmt.Sprintf(`{
				"@context": "https://schema.org",
				"@type": "CreativeWork",
				"name": "%s",
				"description": "%s",
				"author": {
					"@type": "Person",
					"name": "资源分享网站"
				},
				"datePublished": "%s"
			}`, seoCtx.Resource.Title, seoCtx.Resource.Description, seoCtx.Resource.CreatedAt.Format("2006-01-02")))
		}

	case model.SEOConfigTypeCategory:
		if seoCtx.CategoryObj != nil {
			jsonld.WriteString(fmt.Sprintf(`{
				"@context": "https://schema.org",
				"@type": "CollectionPage",
				"name": "%s",
				"description": "%s",
				"url": "https://example.com/category/%d"
			}`, seoCtx.CategoryObj.Name, seoCtx.CategoryObj.Description, *seoCtx.TargetID))
		}
	}

	return jsonld.String()
}

// SEOResponse SEO响应结构
type SEOResponse struct {
	MetaTags    map[string]string `json:"meta_tags"`
	JSONLD      string            `json:"json_ld"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    string            `json:"keywords"`
	Canonical   string            `json:"canonical"`
}

// GenerateSEOResponse 生成完整的SEO响应
func (m *Middleware) GenerateSEOResponse(c *gin.Context) *SEOResponse {
	metaTags := GetMetaTagsFromContext(c)

	response := &SEOResponse{
		MetaTags: metaTags,
	}

	// 提取基本信息
	if title, ok := metaTags["title"]; ok {
		response.Title = title
	}
	if description, ok := metaTags["description"]; ok {
		response.Description = description
	}
	if keywords, ok := metaTags["keywords"]; ok {
		response.Keywords = keywords
	}
	if canonical, ok := metaTags["canonical"]; ok {
		response.Canonical = canonical
	}

	// 生成JSON-LD
	seoCtx := m.parseSEOContext(c)
	response.JSONLD = GenerateJSONLD(seoCtx)

	return response
}

// 设置页面SEO信息
func (m *Middleware) SetPageSEO(c *gin.Context, pageType model.SEOConfigType, targetID *uint, customData map[string]interface{}) {
	// 生成Meta标签
	metaTags, err := m.configService.GenerateMetaTags(pageType, targetID, customData)
	if err != nil {
		return
	}

	// 存储到上下文
	c.Set("seo_meta_tags", metaTags)
}

// 批量设置SEO
func (m *Middleware) BulkSetSEO(c *gin.Context, configs []struct {
	PageType model.SEOConfigType
	TargetID *uint
	Data     map[string]interface{}
}) {
	for _, config := range configs {
		metaTags, err := m.configService.GenerateMetaTags(
			config.PageType,
			config.TargetID,
			config.Data,
		)
		if err != nil {
			continue
		}

		// 存储配置特定的Meta标签
		key := string(config.PageType)
		if config.TargetID != nil {
			key += fmt.Sprintf("_%d", *config.TargetID)
		}
		c.Set("seo_meta_tags_"+key, metaTags)
	}
}
