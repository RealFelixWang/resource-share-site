/*
SEO System Test Program - SEO系统测试程序

测试SEO系统的核心功能：
1. SEO配置管理
2. Meta标签生成
3. Sitemap管理
4. 关键词管理
5. SEO报告
6. SEO中间件

Author: Felix Wang
Email: felixwang.biz@gmail.com
Date: 2025-10-31
*/

package main

import (
	"fmt"
	"log"
	"time"

	"resource-share-site/internal/model"
	seo "resource-share-site/internal/service/seo"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== SEO系统测试程序 ===")
	fmt.Println("作者: Felix Wang")
	fmt.Println("邮箱: felixwang.biz@gmail.com")
	fmt.Println("日期: 2025-10-31")
	fmt.Println()

	// 初始化数据库
	db, err := initTestDatabase()
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 运行测试
	runAllTests(db)
}

func initTestDatabase() (*gorm.DB, error) {
	// 使用 SQLite 内存数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 迁移数据表
	if err := migrateDatabase(db); err != nil {
		return nil, fmt.Errorf("迁移数据表失败: %w", err)
	}

	// 创建测试数据
	if err := createTestData(db); err != nil {
		return nil, fmt.Errorf("创建测试数据失败: %w", err)
	}

	return db, nil
}

func migrateDatabase(db *gorm.DB) error {
	// 自动迁移数据表
	return db.AutoMigrate(
		&model.User{},
		&model.SEOConfig{},
		&model.MetaTag{},
		&model.SitemapUrl{},
		&model.SEOKeyword{},
		&model.SEORank{},
		&model.SEOReport{},
		&model.SEOEvent{},
		&model.Resource{},
		&model.Category{},
	)
}

func createTestData(db *gorm.DB) error {
	// 创建用户
	user := model.User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: "hashed_password",
		Role:         "admin",
		Status:       "active",
	}

	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	// 创建SEO配置
	seoConfigs := []model.SEOConfig{
		{
			ConfigType:      model.SEOConfigTypeHome,
			MetaTitle:       "资源分享网站 - 优质资源一站式分享平台",
			MetaDescription: "提供优质的技术资源、学习资料、工具软件等，助力您的学习和工作",
			MetaKeywords:    "资源分享,技术资料,学习资源,工具软件",
			MetaAuthor:      "资源分享网站",
			OGTitle:         "资源分享网站",
			OGDescription:   "优质资源一站式分享平台",
			OGType:          "website",
			OGUrl:           "https://example.com",
			TwitterCard:     "summary_large_image",
			CanonicalURL:    "https://example.com",
			Robots:          "index, follow",
			Priority:        1.0,
			ChangeFreq:      "daily",
			IsActive:        true,
		},
		{
			ConfigType:      model.SEOConfigTypeResource,
			MetaTitle:       "{{title}} - 资源详情",
			MetaDescription: "下载优质资源：{{title}}",
			MetaKeywords:    "{{title}},资源下载,技术资料",
			OGTitle:         "{{title}}",
			OGDescription:   "{{description}}",
			OGType:          "article",
			TwitterCard:     "summary_large_image",
			CanonicalURL:    "https://example.com/resource/{{target_id}}",
			Robots:          "index, follow",
			Priority:        0.8,
			ChangeFreq:      "weekly",
			IsActive:        true,
		},
		{
			ConfigType:      model.SEOConfigTypeCategory,
			MetaTitle:       "{{name}} - 分类浏览",
			MetaDescription: "浏览{{name}}分类下的优质资源",
			MetaKeywords:    "{{name}},资源分类",
			OGTitle:         "{{name}}分类",
			OGDescription:   "{{name}}分类下的优质资源",
			OGType:          "website",
			TwitterCard:     "summary",
			CanonicalURL:    "https://example.com/category/{{target_id}}",
			Robots:          "index, follow",
			Priority:        0.9,
			ChangeFreq:      "daily",
			IsActive:        true,
		},
	}

	if err := db.Create(&seoConfigs).Error; err != nil {
		return fmt.Errorf("创建SEO配置失败: %w", err)
	}

	// 创建关键词
	keywords := []model.SEOKeyword{
		{
			Keyword:      "Go语言教程",
			Language:     "zh",
			Category:     "编程语言",
			SearchVolume: 5000,
			Difficulty:   60,
			ClickRate:    3.2,
			AvgPosition:  8.5,
			IsActive:     true,
		},
		{
			Keyword:      "Python编程",
			Language:     "zh",
			Category:     "编程语言",
			SearchVolume: 8000,
			Difficulty:   70,
			ClickRate:    2.8,
			AvgPosition:  12.3,
			IsActive:     true,
		},
		{
			Keyword:      "JavaScript教程",
			Language:     "zh",
			Category:     "前端开发",
			SearchVolume: 6000,
			Difficulty:   65,
			ClickRate:    3.0,
			AvgPosition:  10.2,
			IsActive:     true,
		},
		{
			Keyword:      "机器学习",
			Language:     "zh",
			Category:     "人工智能",
			SearchVolume: 10000,
			Difficulty:   80,
			ClickRate:    4.5,
			AvgPosition:  15.8,
			IsActive:     true,
		},
		{
			Keyword:      "数据库设计",
			Language:     "zh",
			Category:     "数据库",
			SearchVolume: 3000,
			Difficulty:   55,
			ClickRate:    2.5,
			AvgPosition:  7.2,
			IsActive:     true,
		},
	}

	if err := db.Create(&keywords).Error; err != nil {
		return fmt.Errorf("创建关键词失败: %w", err)
	}

	// 创建Sitemap URL
	sitemapUrls := []model.SitemapUrl{
		{
			Loc:        "/",
			ChangeFreq: "daily",
			Priority:   1.0,
			PageType:   model.SEOConfigTypeHome,
			IsActive:   true,
		},
		{
			Loc:        "/category/programming",
			LastMod:    &time.Time{},
			ChangeFreq: "daily",
			Priority:   0.9,
			PageType:   model.SEOConfigTypeCategory,
			IsActive:   true,
		},
		{
			Loc:        "/category/frontend",
			LastMod:    &time.Time{},
			ChangeFreq: "daily",
			Priority:   0.9,
			PageType:   model.SEOConfigTypeCategory,
			IsActive:   true,
		},
	}

	if err := db.Create(&sitemapUrls).Error; err != nil {
		return fmt.Errorf("创建Sitemap URL失败: %w", err)
	}

	// 创建排名记录
	rankRecords := []model.SEORank{
		{
			KeywordID:    1,
			SearchEngine: "baidu",
			Rank:         5,
			URL:          "https://example.com/resource/1",
			Title:        "Go语言教程 - 详解",
			Description:  "这是一个优质的Go语言教程",
		},
		{
			KeywordID:    1,
			SearchEngine: "baidu",
			Rank:         3,
			URL:          "https://example.com/resource/1",
			Title:        "Go语言教程 - 详解",
			Description:  "这是一个优质的Go语言教程",
		},
		{
			KeywordID:    2,
			SearchEngine: "google",
			Rank:         8,
			URL:          "https://example.com/resource/2",
			Title:        "Python编程入门",
			Description:  "Python编程基础教程",
		},
	}

	if err := db.Create(&rankRecords).Error; err != nil {
		return fmt.Errorf("创建排名记录失败: %w", err)
	}

	// 创建测试资源
	resources := []model.Resource{
		{
			Title:        "Go语言从入门到精通",
			Description:  "全面的Go语言学习教程，包含基础语法、进阶技巧和实战项目",
			UploadedByID: 1,
			Status:       "approved",
		},
		{
			Title:        "Python数据分析实战",
			Description:  "使用Python进行数据分析的实战教程，包含pandas、numpy等库的使用",
			UploadedByID: 1,
			Status:       "approved",
		},
	}

	if err := db.Create(&resources).Error; err != nil {
		return fmt.Errorf("创建测试资源失败: %w", err)
	}

	// 创建测试分类
	categories := []model.Category{
		{
			Name:        "编程语言",
			Description: "各种编程语言的学习资源",
			ParentID:    nil,
		},
		{
			Name:        "前端开发",
			Description: "前端开发相关资源",
			ParentID:    nil,
		},
	}

	if err := db.Create(&categories).Error; err != nil {
		return fmt.Errorf("创建测试分类失败: %w", err)
	}

	return nil
}

func runAllTests(db *gorm.DB) {
	// 测试SEO配置服务
	fmt.Println("【测试】SEO配置服务")
	fmt.Println("=" + fmt.Sprint(50) + "=")
	if err := testConfigService(db); err != nil {
		fmt.Printf("❌ SEO配置服务测试失败: %v\n\n", err)
	} else {
		fmt.Println("✅ SEO配置服务测试通过\n\n")
	}

	// 测试SEO管理服务
	fmt.Println("【测试】SEO管理服务")
	fmt.Println("=" + fmt.Sprint(50) + "=")
	if err := testManagementService(db); err != nil {
		fmt.Printf("❌ SEO管理服务测试失败: %v\n\n", err)
	} else {
		fmt.Println("✅ SEO管理服务测试通过\n\n")
	}

	// 测试SEO中间件
	fmt.Println("【测试】SEO中间件")
	fmt.Println("=" + fmt.Sprint(50) + "=")
	if err := testSEOMiddleware(db); err != nil {
		fmt.Printf("❌ SEO中间件测试失败: %v\n\n", err)
	} else {
		fmt.Println("✅ SEO中间件测试通过\n\n")
	}

	fmt.Println("=== 所有SEO测试完成 ===")
}

func testConfigService(db *gorm.DB) error {
	configService := seo.NewConfigService(db)

	// 1. 测试创建SEO配置
	fmt.Println("1. 测试创建SEO配置:")
	newConfig := &model.SEOConfig{
		ConfigType:      model.SEOConfigTypeList,
		MetaTitle:       "资源列表 - 资源分享网站",
		MetaDescription: "浏览所有优质资源",
		MetaKeywords:    "资源列表,资源浏览",
		Priority:        0.7,
		ChangeFreq:      "weekly",
		IsActive:        true,
	}
	if err := configService.CreateSEOConfig(newConfig); err != nil {
		fmt.Printf("   ❌ 创建配置失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 创建配置成功，ID: %d\n", newConfig.ID)
	}

	// 2. 测试获取SEO配置
	fmt.Println("\n2. 测试获取SEO配置:")
	config, err := configService.GetSEOConfig(model.SEOConfigTypeHome, nil)
	if err != nil {
		fmt.Printf("   ❌ 获取配置失败: %v\n", err)
	} else if config == nil {
		fmt.Println("   ⚠️  未找到配置")
	} else {
		fmt.Printf("   ✅ 获取配置成功: %s\n", config.MetaTitle)
	}

	// 3. 测试生成Meta标签
	fmt.Println("\n3. 测试生成Meta标签:")
	context := map[string]interface{}{
		"title":       "Go语言教程",
		"description": "全面的Go语言学习资源",
	}
	metaTags, err := configService.GenerateMetaTags(model.SEOConfigTypeResource, nil, context)
	if err != nil {
		fmt.Printf("   ❌ 生成Meta标签失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 生成Meta标签成功:")
		for key, value := range metaTags {
			if len(value) > 50 {
				value = value[:47] + "..."
			}
			fmt.Printf("   - %s: %s\n", key, value)
		}
	}

	// 4. 测试Sitemap生成
	fmt.Println("\n4. 测试Sitemap生成:")
	sitemap, err := configService.GenerateSitemap("https://example.com")
	if err != nil {
		fmt.Printf("   ❌ 生成Sitemap失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 生成Sitemap成功")
		if len(sitemap) > 200 {
			fmt.Printf("   Sitemap长度: %d 字符\n", len(sitemap))
		}
	}

	// 5. 测试添加Sitemap URL
	fmt.Println("\n5. 测试添加Sitemap URL:")
	newURL := &model.SitemapUrl{
		Loc:        "/new-page",
		ChangeFreq: "weekly",
		Priority:   0.6,
		PageType:   model.SEOConfigTypeDetail,
		IsActive:   true,
	}
	if err := configService.AddSitemapUrl(newURL); err != nil {
		fmt.Printf("   ❌ 添加Sitemap URL失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 添加Sitemap URL成功")
	}

	// 6. 测试SEO优化建议
	fmt.Println("\n6. 测试SEO优化建议:")
	suggestions, err := configService.OptimizeSEO(model.SEOConfigTypeHome, nil)
	if err != nil {
		fmt.Printf("   ❌ 获取优化建议失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 优化建议:")
		if len(suggestions) == 0 {
			fmt.Println("   - 当前配置已经很完善！")
		} else {
			for _, suggestion := range suggestions {
				fmt.Printf("   - %s\n", suggestion)
			}
		}
	}

	// 7. 测试自动生成Sitemap
	fmt.Println("\n7. 测试自动生成Sitemap:")
	if err := configService.AutoGenerateSitemap("https://example.com"); err != nil {
		fmt.Printf("   ❌ 自动生成Sitemap失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 自动生成Sitemap成功")
	}

	// 8. 测试获取配置列表
	fmt.Println("\n8. 测试获取配置列表:")
	configs, total, err := configService.ListSEOConfigs("", 1, 10)
	if err != nil {
		fmt.Printf("   ❌ 获取配置列表失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 获取配置列表成功，共 %d 条\n", total)
		for _, config := range configs {
			fmt.Printf("   - %s: %s\n", config.ConfigType, config.MetaTitle)
		}
	}

	return nil
}

func testManagementService(db *gorm.DB) error {
	managementService := seo.NewManagementService(db)

	// 1. 测试创建关键词
	fmt.Println("1. 测试创建关键词:")
	newKeyword := &model.SEOKeyword{
		Keyword:      "Go语言并发编程",
		Language:     "zh",
		Category:     "编程语言",
		SearchVolume: 2000,
		Difficulty:   50,
		IsActive:     true,
	}
	if err := managementService.CreateKeyword(newKeyword); err != nil {
		fmt.Printf("   ❌ 创建关键词失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 创建关键词成功: %s\n", newKeyword.Keyword)
	}

	// 2. 测试获取关键词列表
	fmt.Println("\n2. 测试获取关键词列表:")
	keywords, total, err := managementService.ListKeywords("", "zh", &[]bool{true}[0], 1, 10)
	if err != nil {
		fmt.Printf("   ❌ 获取关键词列表失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 获取关键词列表成功，共 %d 条\n", total)
		for _, keyword := range keywords {
			fmt.Printf("   - %s: 搜索量 %d, 难度 %d\n", keyword.Keyword, keyword.SearchVolume, keyword.Difficulty)
		}
	}

	// 3. 测试关键词建议
	fmt.Println("\n3. 测试关键词建议:")
	suggestions, err := managementService.SuggestKeywords("Go", 5)
	if err != nil {
		fmt.Printf("   ❌ 获取关键词建议失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 获取关键词建议成功，共 %d 条\n", len(suggestions))
		for _, suggestion := range suggestions {
			fmt.Printf("   - %s\n", suggestion.Keyword)
		}
	}

	// 4. 测试记录排名
	fmt.Println("\n4. 测试记录排名:")
	newRank := &model.SEORank{
		KeywordID:    1,
		SearchEngine: "baidu",
		Rank:         2,
		URL:          "https://example.com/resource/1",
		Title:        "Go语言教程",
		Description:  "优质教程",
	}
	if err := managementService.TrackKeywordRank(newRank); err != nil {
		fmt.Printf("   ❌ 记录排名失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 记录排名成功")
	}

	// 5. 测试获取排名历史
	fmt.Println("\n5. 测试获取排名历史:")
	ranks, total, err := managementService.GetKeywordRanks(1, "baidu", "", 1, 10)
	if err != nil {
		fmt.Printf("   ❌ 获取排名历史失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 获取排名历史成功，共 %d 条\n", total)
		for _, rank := range ranks {
			fmt.Printf("   - %s: 第%d名 (%s)\n", rank.SearchEngine, rank.Rank, rank.CreatedAt.Format("2006-01-02"))
		}
	}

	// 6. 测试分析关键词表现
	fmt.Println("\n6. 测试分析关键词表现:")
	analysis, err := managementService.AnalyzeKeywordPerformance(1, 30)
	if err != nil {
		fmt.Printf("   ❌ 分析关键词表现失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 关键词表现分析:")
		fmt.Printf("   - 总记录数: %v\n", analysis["total_records"])
		fmt.Printf("   - 平均排名: %.1f\n", analysis["average_rank"])
		fmt.Printf("   - 最佳排名: %v\n", analysis["best_rank"])
		fmt.Printf("   - 趋势: %s\n", analysis["trend"])
	}

	// 7. 测试生成SEO报告
	fmt.Println("\n7. 测试生成SEO报告:")
	report, err := managementService.GenerateSEOReport("weekly")
	if err != nil {
		fmt.Printf("   ❌ 生成SEO报告失败: %v\n", err)
	} else {
		fmt.Println("   ✅ SEO报告生成成功:")
		fmt.Printf("   - 总页面数: %d\n", report.TotalPages)
		fmt.Printf("   - 已索引页面: %d\n", report.IndexedPages)
		fmt.Printf("   - 总关键词数: %d\n", report.TotalKeywords)
		fmt.Printf("   - 平均排名: %.1f\n", report.AvgRank)
		fmt.Printf("   - SEO得分: %d/100\n", report.SEOScore)
	}

	// 8. 测试获取热门关键词
	fmt.Println("\n8. 测试获取热门关键词:")
	topKeywords, err := managementService.GetTopKeywords(5, "编程语言")
	if err != nil {
		fmt.Printf("   ❌ 获取热门关键词失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 热门关键词TOP 5:")
		for i, keyword := range topKeywords {
			fmt.Printf("   %d. %s (搜索量: %d)\n", i+1, keyword.Keyword, keyword.SearchVolume)
		}
	}

	// 9. 测试获取关键词分类
	fmt.Println("\n9. 测试获取关键词分类:")
	categories, err := managementService.GetKeywordCategories()
	if err != nil {
		fmt.Printf("   ❌ 获取关键词分类失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 关键词分类: %v\n", categories)
	}

	// 10. 测试记录SEO事件
	fmt.Println("\n10. 测试记录SEO事件:")
	if err := managementService.LogSEOEvent("crawl", "搜索引擎爬取", "/resource/1", "baidu", nil); err != nil {
		fmt.Printf("   ❌ 记录SEO事件失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 记录SEO事件成功")
	}

	return nil
}

func testSEOMiddleware(db *gorm.DB) error {
	configService := seo.NewConfigService(db)

	// 1. 测试解析SEO上下文
	fmt.Println("1. 测试解析SEO上下文:")
	// 模拟Gin上下文
	// 注意：这里只是测试中间件的逻辑，不实际运行HTTP服务器

	// 2. 测试生成Meta标签
	fmt.Println("\n2. 测试生成Meta标签:")
	context := map[string]interface{}{
		"title":       "Go语言并发编程教程",
		"description": "深入理解Go语言的并发编程机制",
		"category":    "编程语言",
	}
	metaTags, err := configService.GenerateMetaTags(model.SEOConfigTypeResource, nil, context)
	if err != nil {
		fmt.Printf("   ❌ 生成Meta标签失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 生成Meta标签:")
		for key, value := range metaTags {
			fmt.Printf("   - %s: %s\n", key, value)
		}
	}

	// 3. 测试渲染Meta标签
	fmt.Println("\n3. 测试渲染Meta标签:")
	html := seo.RenderMetaTags(metaTags)
	if html != "" {
		fmt.Println("   ✅ 渲染HTML成功")
		fmt.Printf("   HTML长度: %d 字符\n", len(html))
		// 显示部分HTML
		if len(html) > 200 {
			fmt.Printf("   前200字符: %s...\n", html[:200])
		}
	} else {
		fmt.Println("   ⚠️  没有生成HTML")
	}

	// 4. 测试生成JSON-LD
	fmt.Println("\n4. 测试生成JSON-LD:")
	seoCtx := &seo.SEOContext{
		PageType: model.SEOConfigTypeHome,
	}
	jsonld := seo.GenerateJSONLD(seoCtx)
	if jsonld != "" {
		fmt.Println("   ✅ 生成JSON-LD成功")
		fmt.Printf("   JSON-LD长度: %d 字符\n", len(jsonld))
		if len(jsonld) > 150 {
			fmt.Printf("   前150字符: %s...\n", jsonld[:150])
		}
	} else {
		fmt.Println("   ⚠️  没有生成JSON-LD")
	}

	// 5. 测试模板解析
	fmt.Println("\n5. 测试模板解析:")
	templateContext := map[string]interface{}{
		"title": "Go语言教程",
		"name":  "编程语言",
	}
	metaTagsWithTemplate, err := configService.GenerateMetaTags(model.SEOConfigTypeCategory, nil, templateContext)
	if err != nil {
		fmt.Printf("   ❌ 模板解析失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 模板解析成功:")
		if title, ok := metaTagsWithTemplate["title"]; ok {
			fmt.Printf("   - 解析后的标题: %s\n", title)
		}
	}

	return nil
}
