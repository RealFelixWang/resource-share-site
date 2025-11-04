/*
Test program for category system.

This program tests all category services:
- CategoryService: Category hierarchy management
- PermissionService: Category permission control
- StatisticsService: Category statistics

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
	"resource-share-site/internal/service/user"

	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== 分类系统测试开始 ===\n")

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
	categoryService := category.NewCategoryService(db)
	permissionService := category.NewPermissionService(db)
	statisticsService := category.NewStatisticsService(db)
	authService := auth.NewAuthService(db)
	userStatusService := user.NewUserStatusService(db)
	fmt.Println("✅ 服务初始化成功\n")

	// 1. 测试分类层级管理
	fmt.Println("3. 测试分类层级管理...")
	testCategoryHierarchy(db, categoryService)
	time.Sleep(100 * time.Millisecond)

	// 2. 测试分类权限控制
	fmt.Println("\n4. 测试分类权限控制...")
	testCategoryPermissions(db, categoryService, permissionService, authService, userStatusService)
	time.Sleep(100 * time.Millisecond)

	// 3. 测试分类统计功能
	fmt.Println("\n5. 测试分类统计功能...")
	testCategoryStatistics(db, categoryService, statisticsService)
	time.Sleep(100 * time.Millisecond)

	// 4. 测试分类树结构
	fmt.Println("\n6. 测试分类树结构...")
	testCategoryTree(db, categoryService)
	time.Sleep(100 * time.Millisecond)

	// 5. 测试分类排行榜
	fmt.Println("\n7. 测试分类排行榜...")
	testCategoryRanking(db, categoryService, statisticsService)

	// 清理数据（可选）
	if os.Getenv("CLEANUP") == "true" {
		fmt.Println("\n\n清理测试数据...")
		cleanupTestData(db)
	}

	fmt.Println("\n=== 分类系统测试完成 ===")
}

// testCategoryHierarchy 测试分类层级管理
func testCategoryHierarchy(db *gorm.DB, categoryService *category.CategoryService) {
	// 创建顶级分类
	root1, err := categoryService.CreateCategory("编程语言", "各种编程语言相关资源", "code", "#3b82f6", nil, 1)
	if err != nil {
		log.Printf("创建顶级分类1失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 创建顶级分类成功: ID=%d, 名称=%s\n", root1.ID, root1.Name)

	root2, err := categoryService.CreateCategory("前端技术", "前端开发相关资源", "frontend", "#10b981", nil, 2)
	if err != nil {
		log.Printf("创建顶级分类2失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 创建顶级分类成功: ID=%d, 名称=%s\n", root2.ID, root2.Name)

	// 创建子分类
	child1, err := categoryService.CreateCategory("Go语言", "Go语言学习资源", "go", "#00add8", &root1.ID, 1)
	if err != nil {
		log.Printf("创建子分类1失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 创建子分类成功: ID=%d, 名称=%s\n", child1.ID, child1.Name)

	child2, err := categoryService.CreateCategory("JavaScript", "JavaScript学习资源", "js", "#f7df1e", &root1.ID, 2)
	if err != nil {
		log.Printf("创建子分类2失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 创建子分类成功: ID=%d, 名称=%s\n", child2.ID, child2.Name)

	child3, err := categoryService.CreateCategory("React", "React相关资源", "react", "#61dafb", &root2.ID, 1)
	if err != nil {
		log.Printf("创建子分类3失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 创建子分类成功: ID=%d, 名称=%s\n", child3.ID, child3.Name)

	// 测试更新分类
	updated, err := categoryService.UpdateCategory(child1.ID, "Go语言进阶", "Go语言高级学习资源", "go-advanced", "#00add8", 1)
	if err != nil {
		log.Printf("更新分类失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 更新分类成功: 新名称=%s\n", updated.Name)

	// 测试获取分类
	fetched, err := categoryService.GetCategoryByID(child2.ID)
	if err != nil {
		log.Printf("获取分类失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取分类成功: ID=%d, 名称=%s\n", fetched.ID, fetched.Name)

	// 测试获取所有分类
	allCategories, total, err := categoryService.GetAllCategories(1, 10)
	if err != nil {
		log.Printf("获取所有分类失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取所有分类成功: 总数=%d\n", total)

	// 测试获取顶级分类
	rootCategories, err := categoryService.GetRootCategories()
	if err != nil {
		log.Printf("获取顶级分类失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取顶级分类成功: 数量=%d\n", len(rootCategories))

	// 测试获取子分类
	children, err := categoryService.GetChildCategories(root1.ID)
	if err != nil {
		log.Printf("获取子分类失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取子分类成功: 数量=%d\n", len(children))
}

// testCategoryPermissions 测试分类权限控制
func testCategoryPermissions(db *gorm.DB, categoryService *category.CategoryService, permissionService *category.PermissionService, authService auth.AuthService, userStatusService user.UserStatusService) {
	// 创建测试用户
	user1, err := authService.Register("user1", "user1@example.com", "password123", "", "user")
	if err != nil {
		log.Printf("注册用户1失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 注册用户1成功: ID=%d\n", user1.ID)

	adminUser, err := authService.Register("admin", "admin@example.com", "password123", "", "admin")
	if err != nil {
		log.Printf("注册管理员失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 注册管理员成功: ID=%d\n", adminUser.ID)

	// 激活用户
	if err := userStatusService.ActivateUser(user1.ID, true); err != nil {
		log.Printf("激活用户1失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 激活用户1成功\n")

	// 获取分类
	var category model.Category
	if err := db.Raw("SELECT * FROM categories WHERE name = ?", "Go语言进阶").Scan(&category).Error; err != nil {
		log.Printf("获取分类失败: %v", err)
		return
	}

	// 测试授予权限
	permission, err := permissionService.GrantPermission(user1.ID, category.ID, category.PermissionView, adminUser.ID)
	if err != nil {
		log.Printf("授予权限失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 授予查看权限成功: 权限ID=%d\n", permission.ID)

	// 测试检查权限
	hasPermission, err := permissionService.HasPermission(user1.ID, category.ID, category.PermissionView)
	if err != nil {
		log.Printf("检查权限失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 检查权限成功: 有权限=%v\n", hasPermission)

	// 测试撤销权限
	if err := permissionService.RevokePermission(user1.ID, category.ID, category.PermissionView); err != nil {
		log.Printf("撤销权限失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 撤销权限成功\n")

	// 再次检查权限
	hasPermission, err = permissionService.HasPermission(user1.ID, category.ID, category.PermissionView)
	if err != nil {
		log.Printf("检查权限失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 检查权限成功: 有权限=%v\n", hasPermission)

	// 测试获取用户权限
	permission2, err := permissionService.GrantPermission(user1.ID, category.ID, category.PermissionEdit, adminUser.ID)
	if err != nil {
		log.Printf("授予编辑权限失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 授予编辑权限成功: 权限ID=%d\n", permission2.ID)

	userPermissions, err := permissionService.GetUserPermissions(user1.ID, nil)
	if err != nil {
		log.Printf("获取用户权限失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取用户权限成功: 权限数量=%d\n", len(userPermissions))

	// 测试获取分类权限
	categoryPermissions, err := permissionService.GetCategoryPermissions(category.ID)
	if err != nil {
		log.Printf("获取分类权限失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取分类权限成功: 权限数量=%d\n", len(categoryPermissions))

	// 测试批量授予权限
	// 注意：这里需要先创建更多用户才能测试
	// count, err := permissionService.BatchGrantPermission([]uint{user1.ID}, category.ID, category.PermissionDelete, adminUser.ID)
	// if err != nil {
	// 	log.Printf("批量授予权限失败: %v", err)
	// 	return
	// }
	// fmt.Printf("  ✓ 批量授予权限成功: 授予数量=%d\n", count)
}

// testCategoryStatistics 测试分类统计功能
func testCategoryStatistics(db *gorm.DB, categoryService *category.CategoryService, statisticsService *category.StatisticsService) {
	// 获取一个分类
	var category model.Category
	if err := db.Raw("SELECT * FROM categories WHERE name = ?", "Go语言进阶").Scan(&category).Error; err != nil {
		log.Printf("获取分类失败: %v", err)
		return
	}

	// 测试获取分类统计信息
	stats, err := statisticsService.GetCategoryStats(category.ID)
	if err != nil {
		log.Printf("获取分类统计失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取分类统计成功: 资源数=%d, 子分类数=%d\n", stats.ResourcesCount, stats.ChildrenCount)

	// 测试获取所有分类统计
	allStats, total, err := statisticsService.GetAllCategoriesStats(1, 10)
	if err != nil {
		log.Printf("获取所有分类统计失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取所有分类统计成功: 总数=%d\n", total)

	// 测试更新分类资源计数
	if err := statisticsService.UpdateCategoryResourceCount(category.ID); err != nil {
		log.Printf("更新分类资源计数失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 更新分类资源计数成功\n")

	// 测试批量更新所有分类计数
	updatedCount, err := statisticsService.BatchUpdateAllCategoryCounts()
	if err != nil {
		log.Printf("批量更新分类计数失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 批量更新分类计数成功: 更新数量=%d\n", updatedCount)

	// 测试获取分类趋势
	trends, err := statisticsService.GetCategoryTrends(category.ID, 7)
	if err != nil {
		log.Printf("获取分类趋势失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取分类趋势成功: 天数=%d\n", len(trends["resources"]))
}

// testCategoryTree 测试分类树结构
func testCategoryTree(db *gorm.DB, categoryService *category.CategoryService) {
	// 测试获取分类树
	tree, err := categoryService.GetCategoryTree(nil, 5)
	if err != nil {
		log.Printf("获取分类树失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取分类树成功: 顶级分类数量=%d\n", len(tree))

	// 测试获取分类路径
	var category model.Category
	if err := db.Raw("SELECT * FROM categories WHERE name = ?", "Go语言进阶").Scan(&category).Error; err != nil {
		log.Printf("获取分类失败: %v", err)
		return
	}

	path, err := categoryService.GetCategoryPath(category.ID)
	if err != nil {
		log.Printf("获取分类路径失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取分类路径成功: 路径长度=%d\n", len(path))
	for i, cat := range path {
		fmt.Printf("    第%d级: %s\n", i+1, cat.Name)
	}

	// 测试移动分类
	var parentCategory model.Category
	if err := db.Raw("SELECT * FROM categories WHERE name = ?", "前端技术").Scan(&parentCategory).Error; err != nil {
		log.Printf("获取父级分类失败: %v", err)
		return
	}

	if err := categoryService.MoveCategory(category.ID, &parentCategory.ID); err != nil {
		log.Printf("移动分类失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 移动分类成功\n")

	// 移动回原位置
	var rootCategory model.Category
	if err := db.Raw("SELECT * FROM categories WHERE name = ?", "编程语言").Scan(&rootCategory).Error; err != nil {
		log.Printf("获取根分类失败: %v", err)
		return
	}

	if err := categoryService.MoveCategory(category.ID, &rootCategory.ID); err != nil {
		log.Printf("移动分类失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 移动分类回原位置成功\n")

	// 测试更新排序
	if err := categoryService.UpdateSortOrder(category.ID, 10); err != nil {
		log.Printf("更新排序失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 更新排序成功\n")
}

// testCategoryRanking 测试分类排行榜
func testCategoryRanking(db *gorm.DB, categoryService *category.CategoryService, statisticsService *category.StatisticsService) {
	// 测试资源排行榜
	resourcesRanking, err := statisticsService.GetCategoryRanking("resources", "all", 10, 0)
	if err != nil {
		log.Printf("获取资源排行榜失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取资源排行榜成功: 数量=%d\n", len(resourcesRanking))

	// 测试浏览量排行榜
	viewsRanking, err := statisticsService.GetCategoryRanking("views", "all", 10, 0)
	if err != nil {
		log.Printf("获取浏览量排行榜失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取浏览量排行榜成功: 数量=%d\n", len(viewsRanking))

	// 测试增长率排行榜
	growthRanking, err := statisticsService.GetCategoryRanking("growth", "all", 10, 0)
	if err != nil {
		log.Printf("获取增长率排行榜失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取增长率排行榜成功: 数量=%d\n", len(growthRanking))

	// 测试热门度排行榜
	popularityRanking, err := statisticsService.GetCategoryRanking("popularity", "all", 10, 0)
	if err != nil {
		log.Printf("获取热门度排行榜失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取热门度排行榜成功: 数量=%d\n", len(popularityRanking))

	// 测试指定时间范围的排行榜
	monthlyRanking, err := statisticsService.GetCategoryRanking("resources", "month", 10, 0)
	if err != nil {
		log.Printf("获取月度排行榜失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取月度排行榜成功: 数量=%d\n", len(monthlyRanking))
}

// cleanupTestData 清理测试数据
func cleanupTestData(db *gorm.DB) {
	// 删除权限
	if err := db.Exec("DELETE FROM category_permissions").Error; err != nil {
		log.Printf("清理权限数据失败: %v", err)
	}

	// 删除分类
	if err := db.Exec("DELETE FROM categories WHERE name IN ('编程语言', '前端技术', 'Go语言进阶', 'JavaScript', 'React')").Error; err != nil {
		log.Printf("清理分类数据失败: %v", err)
	}

	// 删除用户
	if err := db.Exec("DELETE FROM users WHERE username IN ('user1', 'admin')").Error; err != nil {
		log.Printf("清理用户数据失败: %v", err)
	}

	fmt.Println("  ✓ 清理测试数据完成")
}
