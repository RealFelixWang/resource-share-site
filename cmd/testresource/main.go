/*
Test program for resource system.

This program tests all resource services:
- ResourceService: Resource upload and download
- ReviewService: Resource review mechanism
- CategoryManagementService: Resource category management
- StatisticsService: Resource statistics

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"resource-share-site/internal/config"
	"resource-share-site/internal/database"
	"resource-share-site/internal/model"
	"resource-share-site/internal/service/auth"
	"resource-share-site/internal/service/category"
	"resource-share-site/internal/service/resource"
	"resource-share-site/internal/service/user"

	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== 资源系统测试开始 ===\n")

	// 初始化数据库
	fmt.Println("1. 初始化数据库...")
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
		log.Fatalf("数据库初始化失败: %v", err)
	}
	fmt.Println("✅ 数据库初始化成功\n")

	// 运行迁移
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化服务
	fmt.Println("2. 初始化服务...")
	resourceService := resource.NewResourceService(db)
	reviewService := resource.NewReviewService(db)
	categoryManagementService := resource.NewCategoryManagementService(db)
	statisticsService := resource.NewStatisticsService(db)

	authService := auth.NewAuthService(db)
	userStatusService := user.NewUserStatusService(db)
	categoryService := category.NewCategoryService(db)

	fmt.Println("✅ 服务初始化成功\n")

	// 1. 测试资源上传功能
	fmt.Println("3. 测试资源上传功能...")
	testResourceUpload(db, resourceService, categoryService, authService, userStatusService)
	time.Sleep(100 * time.Millisecond)

	// 2. 测试资源审核功能
	fmt.Println("\n4. 测试资源审核功能...")
	testResourceReview(db, reviewService, resourceService, authService, userStatusService)
	time.Sleep(100 * time.Millisecond)

	// 3. 测试资源分类管理功能
	fmt.Println("\n5. 测试资源分类管理功能...")
	testResourceCategoryManagement(db, categoryManagementService, resourceService, authService, userStatusService)
	time.Sleep(100 * time.Millisecond)

	// 4. 测试资源统计功能
	fmt.Println("\n6. 测试资源统计功能...")
	testResourceStatistics(db, statisticsService, resourceService)
	time.Sleep(100 * time.Millisecond)

	// 5. 测试资源下载功能
	fmt.Println("\n7. 测试资源下载功能...")
	testResourceDownload(db, resourceService, authService, userStatusService)
	time.Sleep(100 * time.Millisecond)

	// 6. 测试资源搜索功能
	fmt.Println("\n8. 测试资源搜索功能...")
	testResourceSearch(db, resourceService)
	time.Sleep(100 * time.Millisecond)

	// 清理数据（可选）
	if os.Getenv("CLEANUP") == "true" {
		fmt.Println("\n\n清理测试数据...")
		cleanupTestData(db)
	}

	fmt.Println("\n=== 资源系统测试完成 ===")
}

// testResourceUpload 测试资源上传
func testResourceUpload(db *gorm.DB, resourceService *resource.ResourceService, categoryService *category.CategoryService, authService auth.AuthService, userStatusService user.UserStatusService) {
	// 创建测试用户
	user1, err := authService.Register("uploader1", "uploader1@example.com", "password123", "", "user")
	if err != nil {
		log.Printf("注册用户1失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 注册用户1成功: ID=%d\n", user1.ID)

	user2, err := authService.Register("uploader2", "uploader2@example.com", "password123", "", "user")
	if err != nil {
		log.Printf("注册用户2失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 注册用户2成功: ID=%d\n", user2.ID)

	// 创建管理员
	admin, err := authService.Register("admin", "admin@example.com", "password123", "", "admin")
	if err != nil {
		log.Printf("注册管理员失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 注册管理员成功: ID=%d\n", admin.ID)

	// 激活用户
	if err := userStatusService.ActivateUser(user1.ID, true); err != nil {
		log.Printf("激活用户1失败: %v", err)
		return
	}
	if err := userStatusService.ActivateUser(user2.ID, true); err != nil {
		log.Printf("激活用户2失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 激活用户成功\n")

	// 创建分类
	cat1, err := categoryService.CreateCategory("编程教程", "编程学习资源", "code", "#3b82f6", nil, 1)
	if err != nil {
		log.Printf("创建分类1失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 创建分类1成功: ID=%d, 名称=%s\n", cat1.ID, cat1.Name)

	cat2, err := categoryService.CreateCategory("前端开发", "前端开发相关资源", "frontend", "#10b981", nil, 2)
	if err != nil {
		log.Printf("创建分类2失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 创建分类2成功: ID=%d, 名称=%s\n", cat2.ID, cat2.Name)

	// 上传资源
	resource1, err := resourceService.CreateResource(
		"Go语言入门教程",
		"详细的Go语言学习教程，从基础到进阶",
		cat1.ID,
		"https://example.com/go-tutorial",
		50,
		`["Go", "教程", "编程"]`,
		user1.ID,
		model.ResourceSourceUser,
	)
	if err != nil {
		log.Printf("上传资源1失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 上传资源1成功: ID=%d, 标题=%s\n", resource1.ID, resource1.Title)

	resource2, err := resourceService.CreateResource(
		"React实战项目",
		"基于React的实战项目源码",
		cat2.ID,
		"https://example.com/react-project",
		100,
		`["React", "前端", "项目"]`,
		user1.ID,
		model.ResourceSourceUser,
	)
	if err != nil {
		log.Printf("上传资源2失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 上传资源2成功: ID=%d, 标题=%s\n", resource2.ID, resource2.Title)

	resource3, err := resourceService.CreateResource(
		"Vue.js权威指南",
		"Vue.js官方推荐学习指南",
		cat2.ID,
		"https://example.com/vue-guide",
		0, // 免费资源
		`["Vue", "前端", "指南"]`,
		user2.ID,
		model.ResourceSourceUser,
	)
	if err != nil {
		log.Printf("上传资源3失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 上传资源3成功: ID=%d, 标题=%s\n", resource3.ID, resource3.Title)

	// 更新资源
	updated, err := resourceService.UpdateResource(
		resource1.ID,
		"Go语言入门教程（更新版）",
		"更新的Go语言学习教程，包含最新特性",
		cat1.ID,
		"https://example.com/go-tutorial-updated",
		60,
		`["Go", "教程", "编程", "更新"]`,
	)
	if err != nil {
		log.Printf("更新资源1失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 更新资源1成功: 新标题=%s\n", updated.Title)

	// 获取资源列表
	resources, total, err := resourceService.GetResources(1, 10, nil, nil, nil, nil, nil, "created_at", true)
	if err != nil {
		log.Printf("获取资源列表失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取资源列表成功: 总数=%d, 第一页=%d条\n", total, len(resources))

	// 获取用户资源
	userResources, total, err := resourceService.GetUserResources(user1.ID, 1, 10, nil)
	if err != nil {
		log.Printf("获取用户资源失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取用户资源成功: 用户ID=%d, 资源数=%d\n", user1.ID, len(userResources))

	// 获取分类资源
	catResources, total, err := resourceService.GetResourcesByCategory(cat2.ID, 1, 10, nil, "created_at", true)
	if err != nil {
		log.Printf("获取分类资源失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取分类资源成功: 分类ID=%d, 资源数=%d\n", cat2.ID, len(catResources))
}

// testResourceReview 测试资源审核
func testResourceReview(db *gorm.DB, reviewService *resource.ReviewService, resourceService *resource.ResourceService, authService auth.AuthService, userStatusService user.UserStatusService) {
	// 获取待审核资源
	pendingResources, total, err := reviewService.GetPendingResources(1, 10, nil, nil, nil)
	if err != nil {
		log.Printf("获取待审核资源失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取待审核资源成功: 总数=%d\n", total)

	if len(pendingResources) > 0 {
		// 审核第一个资源（通过）
		resourceID := pendingResources[0].ID
		admin, _ := getUserByEmail(db, "admin@example.com")

		reviewLog, err := reviewService.ReviewResource(resourceID, admin.ID, resource.ReviewActionApprove, "资源质量良好，通过审核")
		if err != nil {
			log.Printf("审核资源1失败: %v", err)
			return
		}
		fmt.Printf("  ✓ 审核资源1通过: 资源ID=%d, 审核动作=%s\n", resourceID, reviewLog.Action)

		// 审核第二个资源（拒绝）
		if len(pendingResources) > 1 {
			resourceID2 := pendingResources[1].ID
			reviewLog2, err := reviewService.ReviewResource(resourceID2, admin.ID, resource.ReviewActionReject, "资源描述不完整，拒绝审核")
			if err != nil {
				log.Printf("审核资源2失败: %v", err)
				return
			}
			fmt.Printf("  ✓ 审核资源2拒绝: 资源ID=%d, 审核动作=%s\n", resourceID2, reviewLog2.Action)
		}

		// 获取已审核资源
		reviewed, total, err := reviewService.GetReviewedResources(1, 10, model.ResourceStatusApproved, nil, nil, nil, nil)
		if err != nil {
			log.Printf("获取已审核资源失败: %v", err)
			return
		}
		fmt.Printf("  ✓ 获取已审核资源成功: 已通过数=%d\n", len(reviewed))

		// 获取审核日志
		logs, total, err := reviewService.GetReviewLogs(nil, nil, 1, 10, nil, nil)
		if err != nil {
			log.Printf("获取审核日志失败: %v", err)
			return
		}
		fmt.Printf("  ✓ 获取审核日志成功: 总数=%d\n", total)

		// 获取审核统计
		stats, err := reviewService.GetReviewStatistics(nil, nil)
		if err != nil {
			log.Printf("获取审核统计失败: %v", err)
			return
		}
		fmt.Printf("  ✓ 获取审核统计成功: 总数=%d, 待审核=%d, 已通过=%d\n",
			stats["total"], stats["pending"], stats["approved"])
	}
}

// testResourceCategoryManagement 测试资源分类管理
func testResourceCategoryManagement(db *gorm.DB, categoryManagementService *resource.CategoryManagementService, resourceService *resource.ResourceService, authService auth.AuthService, userStatusService user.UserStatusService) {
	// 获取一个资源
	var resource model.Resource
	if err := db.Raw("SELECT * FROM resources WHERE title LIKE ?", "Go语言%").Scan(&resource).Error; err != nil {
		log.Printf("获取资源失败: %v", err)
		return
	}

	// 获取另一个分类
	var newCategory model.Category
	if err := db.Raw("SELECT * FROM categories WHERE name = ?", "前端开发").Scan(&newCategory).Error; err != nil {
		log.Printf("获取分类失败: %v", err)
		return
	}

	// 移动资源到新分类
	admin, _ := getUserByEmail(db, "admin@example.com")

	changeLog, err := categoryManagementService.MoveResourceToCategory(
		resource.ID,
		newCategory.ID,
		admin.ID,
		"资源内容调整，移动到更合适的分类",
	)
	if err != nil {
		log.Printf("移动资源分类失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 移动资源分类成功: 资源ID=%d, 新分类=%s\n", resource.ID, newCategory.Name)

	// 获取分类资源统计
	stats, err := categoryManagementService.GetCategoryResourceStats(newCategory.ID, false)
	if err != nil {
		log.Printf("获取分类资源统计失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取分类资源统计成功: 分类=%s, 总资源=%d\n", newCategory.Name, stats["total"])

	// 获取分类资源排行榜
	ranking, err := categoryManagementService.GetCategoryResourceRanking(newCategory.ID, "downloads", 5, false)
	if err != nil {
		log.Printf("获取分类资源排行榜失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取分类资源排行榜成功: 排行榜数=%d\n", len(ranking))

	// 获取分类变更日志
	changeLogs, total, err := categoryManagementService.GetCategoryChangeLogs(&newCategory.ID, nil, nil, 1, 10, nil, nil)
	if err != nil {
		log.Printf("获取分类变更日志失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取分类变更日志成功: 总数=%d\n", total)

	// 更新分类资源计数
	if err := categoryManagementService.UpdateCategoryResourceCount(newCategory.ID); err != nil {
		log.Printf("更新分类资源计数失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 更新分类资源计数成功\n")

	// 批量更新所有分类计数
	updatedCount, err := categoryManagementService.BatchUpdateAllCategoryResourceCounts()
	if err != nil {
		log.Printf("批量更新分类计数失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 批量更新分类计数成功: 更新数量=%d\n", updatedCount)
}

// testResourceStatistics 测试资源统计
func testResourceStatistics(db *gorm.DB, statisticsService *resource.StatisticsService, resourceService *resource.ResourceService) {
	// 获取总体统计
	overallStats, err := statisticsService.GetOverallStatistics(nil, nil)
	if err != nil {
		log.Printf("获取总体统计失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取总体统计成功: 总资源=%d, 总下载=%d, 总浏览=%d\n",
		overallStats.TotalResources, overallStats.TotalDownloads, overallStats.TotalViews)

	// 获取用户统计
	user1, _ := getUserByEmail(db, "uploader1@example.com")
	uploaderStats, err := statisticsService.GetUploaderStatistics(user1.ID, nil, nil)
	if err != nil {
		log.Printf("获取上传者统计失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取上传者统计成功: 用户=%s, 总资源=%d, 通过率=%.2f%%\n",
		uploaderStats.Username, uploaderStats.TotalResources, uploaderStats.ApprovalRate)

	// 获取热门资源排行
	popularResources, err := statisticsService.GetPopularResources("downloads", "all", nil, 5, 0)
	if err != nil {
		log.Printf("获取热门资源排行失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取热门资源排行成功: 排行数=%d\n", len(popularResources))
	for i, res := range popularResources {
		fmt.Printf("    排名%d: %s (下载数=%d)\n", i+1, res.Title, res.DownloadsCount)
	}

	// 获取资源趋势
	trends, err := statisticsService.GetResourceTrends(7, nil, "uploads")
	if err != nil {
		log.Printf("获取资源趋势失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取资源趋势成功: 天数=%d\n", len(trends))

	// 获取上传者排行榜
	uploadersRanking, err := statisticsService.GetUploadersRanking("resources", "all", 5, 0)
	if err != nil {
		log.Printf("获取上传者排行榜失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取上传者排行榜成功: 排行数=%d\n", len(uploadersRanking))
}

// testResourceDownload 测试资源下载
func testResourceDownload(db *gorm.DB, resourceService *resource.ResourceService, authService auth.AuthService, userStatusService user.UserStatusService) {
	// 获取一个已审核的资源
	var approvedResource model.Resource
	if err := db.Raw("SELECT * FROM resources WHERE status = ? LIMIT 1", model.ResourceStatusApproved).Scan(&approvedResource).Error; err != nil {
		log.Printf("获取已审核资源失败: %v", err)
		return
	}

	if approvedResource.ID == 0 {
		fmt.Printf("  ! 没有已审核的资源，跳过下载测试\n")
		return
	}

	// 创建下载用户
	downloader, err := authService.Register("downloader", "downloader@example.com", "password123", "", "user")
	if err != nil {
		log.Printf("注册下载用户失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 注册下载用户成功: ID=%d\n", downloader.ID)

	// 激活用户
	if err := userStatusService.ActivateUser(downloader.ID, true); err != nil {
		log.Printf("激活下载用户失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 激活下载用户成功\n")

	// 模拟浏览资源
	if err := resourceService.ViewResource(approvedResource.ID); err != nil {
		log.Printf("浏览资源失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 浏览资源成功: 资源ID=%d\n", approvedResource.ID)

	// 下载资源
	downloadURL, err := resourceService.DownloadResource(approvedResource.ID, downloader.ID)
	if err != nil {
		log.Printf("下载资源失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 下载资源成功: 资源ID=%d, 下载链接=%s\n", approvedResource.ID, downloadURL)

	// 更新用户积分（给下载用户一些积分用于测试）
	var downloaderUser model.User
	if err := db.First(&downloaderUser, downloader.ID).Error; err == nil {
		if err := db.Model(&downloaderUser).
			Where("id = ?", downloader.ID).
			Update("points_balance", 200).Error; err != nil {
			log.Printf("更新下载用户积分失败: %v", err)
		}
		fmt.Printf("  ✓ 更新下载用户积分成功\n")
	}

	// 再次尝试下载付费资源
	var paidResource model.Resource
	if err := db.Raw("SELECT * FROM resources WHERE points_price > 0 AND status = ? LIMIT 1", model.ResourceStatusApproved).Scan(&paidResource).Error; err != nil {
		log.Printf("获取付费资源失败: %v", err)
		return
	}

	if paidResource.ID == 0 {
		fmt.Printf("  ! 没有付费资源，跳过付费下载测试\n")
		return
	}

	downloadURL2, err := resourceService.DownloadResource(paidResource.ID, downloader.ID)
	if err != nil {
		if err == resource.ErrInsufficientPoints {
			fmt.Printf("  ✓ 积分不足测试成功: 需要积分=%d\n", paidResource.PointsPrice)
		} else {
			log.Printf("下载付费资源失败: %v", err)
		}
	} else {
		fmt.Printf("  ✓ 下载付费资源成功: 资源ID=%d\n", paidResource.ID)
	}
	_ = downloadURL2
}

// testResourceSearch 测试资源搜索
func testResourceSearch(db *gorm.DB, resourceService *resource.ResourceService) {
	// 搜索资源
	resources, total, err := resourceService.SearchResources("Go", 1, 10, nil, nil)
	if err != nil {
		log.Printf("搜索资源失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 搜索资源成功: 关键词='Go', 结果数=%d\n", len(resources))

	// 获取免费资源
	resources2, total, err := resourceService.GetFreeResources(1, 10, nil)
	if err != nil {
		log.Printf("获取免费资源失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取免费资源成功: 结果数=%d\n", len(resources2))

	// 获取热门资源
	resources3, total, err := resourceService.GetPopularResources(1, 10, nil, nil)
	if err != nil {
		log.Printf("获取热门资源失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取热门资源成功: 结果数=%d\n", len(resources3))
}

// getUserByEmail 根据邮箱获取用户
func getUserByEmail(db *gorm.DB, email string) (*model.User, error) {
	var user model.User
	if err := db.Raw("SELECT * FROM users WHERE email = ?", email).Scan(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// cleanupTestData 清理测试数据
func cleanupTestData(db *gorm.DB) {
	// 删除分类变更日志
	if err := db.Exec("DELETE FROM category_change_logs").Error; err != nil {
		log.Printf("清理分类变更日志失败: %v", err)
	}

	// 删除审核日志
	if err := db.Exec("DELETE FROM review_logs").Error; err != nil {
		log.Printf("清理审核日志失败: %v", err)
	}

	// 删除资源
	if err := db.Exec("DELETE FROM resources WHERE title LIKE 'Go语言%' OR title LIKE 'React%' OR title LIKE 'Vue%'").Error; err != nil {
		log.Printf("清理资源失败: %v", err)
	}

	// 删除分类
	if err := db.Exec("DELETE FROM categories WHERE name IN ('编程教程', '前端开发')").Error; err != nil {
		log.Printf("清理分类失败: %v", err)
	}

	// 删除用户
	if err := db.Exec("DELETE FROM users WHERE email LIKE '%@example.com'").Error; err != nil {
		log.Printf("清理用户失败: %v", err)
	}

	fmt.Println("  ✓ 清理测试数据完成")
}
