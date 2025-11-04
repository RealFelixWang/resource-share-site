/*
Database Connection Test Program

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package main

import (
	"fmt"
	"log"

	"resource-share-site/internal/config"
	"resource-share-site/internal/database"
	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// 数据库连接测试程序
func main() {
	fmt.Println("=== 数据库连接测试 ===\n")

	// 1. 测试数据库配置
	fmt.Println("1. 初始化数据库配置...")
	dbConfig := &config.DatabaseConfig{
		Type:     "sqlite", // 可以改为 "mysql" 使用 MySQL
		Name:     "resource_share_site",
		Host:     "localhost",
		Port:     "3306",
		User:     "root",
		Password: "123456",
		Charset:  "utf8mb4",
	}
	fmt.Printf("数据库类型: %s\n", dbConfig.Type)
	fmt.Printf("数据库名称: %s\n\n", dbConfig.Name)

	// 2. 初始化数据库连接
	fmt.Println("2. 连接数据库...")
	db, err := config.InitDatabase(dbConfig)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	fmt.Println("✅ 数据库连接成功\n")

	// 3. 测试连接
	fmt.Println("3. 测试数据库连接...")
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("数据库Ping失败: %v", err)
	}
	fmt.Println("✅ 数据库Ping成功\n")

	// 4. 测试自动迁移
	fmt.Println("4. 测试数据库迁移...")
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	fmt.Println("✅ 数据库迁移成功\n")

	// 5. 创建默认数据
	fmt.Println("5. 创建默认数据...")
	if err := database.CreateDefaultData(db); err != nil {
		log.Fatalf("创建默认数据失败: %v", err)
	}
	fmt.Println("✅ 默认数据创建成功\n")

	// 6. 测试CRUD操作
	fmt.Println("6. 测试CRUD操作...")
	if err := testCRUD(db); err != nil {
		log.Fatalf("CRUD测试失败: %v", err)
	}
	fmt.Println("✅ CRUD测试通过\n")

	// 7. 查询数据验证
	fmt.Println("7. 查询数据验证...")
	if err := queryData(db); err != nil {
		log.Fatalf("数据验证失败: %v", err)
	}
	fmt.Println("✅ 数据验证通过\n")

	// 8. 显示数据库状态
	fmt.Println("8. 数据库状态...")
	if err := database.GetMigrationStatus(db); err != nil {
		log.Fatalf("获取数据库状态失败: %v", err)
	}

	fmt.Println("\n=== 所有测试通过! ===")
}

// 测试CRUD操作
func testCRUD(db *gorm.DB) error {
	// 创建测试用户
	testUser := model.User{
		Username:      "testuser",
		Email:         "test@example.com",
		PasswordHash:  "hashedpassword",
		Role:          "user",
		Status:        "active",
		CanUpload:     false,
		InviteCode:    "TEST123456789",
		PointsBalance: 100,
	}

	if err := db.Create(&testUser).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	fmt.Printf("  ✅ 创建用户: %s (ID: %d)\n", testUser.Username, testUser.ID)

	// 读取用户
	var retrievedUser model.User
	if err := db.First(&retrievedUser, testUser.ID).Error; err != nil {
		return fmt.Errorf("读取用户失败: %w", err)
	}
	fmt.Printf("  ✅ 读取用户: %s\n", retrievedUser.Username)

	// 更新用户
	if err := db.Model(&retrievedUser).Update("points_balance", 150).Error; err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}
	fmt.Printf("  ✅ 更新用户积分: %d\n", retrievedUser.PointsBalance)

	// 删除用户（软删除）
	if err := db.Delete(&retrievedUser).Error; err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}
	fmt.Printf("  ✅ 软删除用户: %s\n", retrievedUser.Username)

	return nil
}

// 查询数据验证
func queryData(db *gorm.DB) error {
	// 查询用户数量
	var userCount int64
	if err := db.Model(&model.User{}).Count(&userCount).Error; err != nil {
		return fmt.Errorf("查询用户数量失败: %w", err)
	}
	fmt.Printf("  📊 用户总数: %d\n", userCount)

	// 查询分类数量
	var categoryCount int64
	if err := db.Model(&model.Category{}).Count(&categoryCount).Error; err != nil {
		return fmt.Errorf("查询分类数量失败: %w", err)
	}
	fmt.Printf("  📊 分类总数: %d\n", categoryCount)

	// 查询积分规则
	var rules []model.PointsRule
	if err := db.Find(&rules).Error; err != nil {
		return fmt.Errorf("查询积分规则失败: %w", err)
	}
	fmt.Printf("  📊 积分规则数: %d\n", len(rules))
	for _, rule := range rules {
		fmt.Printf("    - %s: %d积分 (%v)\n", rule.RuleName, rule.Points, rule.IsEnabled)
	}

	// 查询权限
	var permissions []model.Permission
	if err := db.Find(&permissions).Error; err != nil {
		return fmt.Errorf("查询权限失败: %w", err)
	}
	fmt.Printf("  📊 权限配置数: %d\n", len(permissions))
	for _, perm := range permissions {
		fmt.Printf("    - %s: %s\n", perm.Name, perm.Description)
	}

	// 查询管理员用户
	var adminCount int64
	if err := db.Model(&model.User{}).Where("role = ?", "admin").Count(&adminCount).Error; err != nil {
		return fmt.Errorf("查询管理员数量失败: %w", err)
	}
	fmt.Printf("  📊 管理员数量: %d\n", adminCount)

	return nil
}
