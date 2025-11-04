/*
User Status Management Test Examples

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package main

import (
	"fmt"
	"log"
	"time"

	"resource-share-site/internal/config"
	"resource-share-site/internal/database"
	"resource-share-site/internal/model"
	"resource-share-site/internal/service/auth"
	"resource-share-site/internal/service/user"

	"gorm.io/gorm"
)

// 用户状态管理测试程序
func main() {
	fmt.Println("=== 用户状态管理测试 ===\n")

	// 1. 初始化数据库
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
		panic(fmt.Sprintf("数据库初始化失败: %v", err))
	}
	fmt.Println("✅ 数据库初始化成功\n")

	// 2. 创建服务
	fmt.Println("2. 创建服务...")
	authService := auth.NewAuthService(db)
	userStatusService := user.NewUserStatusService(db)
	fmt.Println("✅ 服务创建成功\n")

	// 3. 创建测试用户
	fmt.Println("3. 创建测试用户...")
	testUserID := createTestUser(authService, db)
	fmt.Println()

	// 4. 获取用户状态
	fmt.Println("4. 获取用户状态...")
	getUserStatusExample(userStatusService, db, testUserID)
	fmt.Println()

	// 5. 封禁用户
	fmt.Println("5. 封禁用户...")
	banUserExample(userStatusService, db, testUserID, 1 /* admin ID */)
	fmt.Println()

	// 6. 检查用户是否可登录
	fmt.Println("6. 检查用户登录状态...")
	checkUserActiveExample(userStatusService, db, testUserID)
	fmt.Println()

	// 7. 获取被封禁用户列表
	fmt.Println("7. 获取被封禁用户列表...")
	getBannedUsersExample(userStatusService, db)
	fmt.Println()

	// 8. 解封用户
	fmt.Println("8. 解封用户...")
	unbanUserExample(userStatusService, db, testUserID, 1 /* admin ID */)
	fmt.Println()

	// 9. 激活/禁用用户
	fmt.Println("9. 测试激活/禁用用户...")
	activateDeactivateExample(userStatusService, db, testUserID, 1 /* admin ID */)
	fmt.Println()

	// 10. 批量操作示例
	fmt.Println("10. 批量操作示例...")
	batchOperationsExample(userStatusService, authService, db, 1 /* admin ID */)
	fmt.Println()

	fmt.Println("=== 所有测试完成 ===")
}

// 创建测试用户
func createTestUser(authService auth.AuthService, db *gorm.DB) uint {
	var user model.User
	result := db.Where("username = ?", "statustest").First(&user)
	if result.Error == nil {
		fmt.Printf("  ✅ 用户已存在: %s\n", user.Username)
		return user.ID
	}

	// 注册新用户
	req := &auth.RegisterRequest{
		Username:        "statustest",
		Email:           "statustest@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	response, err := authService.Register(&auth.GORMContext{DB: db}, req)
	if err != nil {
		log.Printf("  ❌ 注册失败: %v\n", err)
		return 0
	}

	fmt.Printf("  ✅ 用户创建成功: %s (%s)\n", response.Username, response.Email)
	return response.ID
}

// 获取用户状态示例
func getUserStatusExample(service user.UserStatusService, db *gorm.DB, userID uint) {
	status, err := service.GetUserStatus(&auth.GORMContext{DB: db}, userID)
	if err != nil {
		fmt.Printf("  ❌ 获取状态失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 用户状态信息:\n")
	fmt.Printf("     ID: %d\n", status.ID)
	fmt.Printf("     用户名: %s\n", status.Username)
	fmt.Printf("     邮箱: %s\n", status.Email)
	fmt.Printf("     状态: %s\n", status.Status)
	fmt.Printf("     是否被封禁: %v\n", status.IsBanned)
	fmt.Printf("     上传权限: %v\n", status.CanUpload)
	fmt.Printf("     积分余额: %d\n", status.PointsBalance)
	fmt.Printf("     创建时间: %v\n", status.CreatedAt)
}

// 封禁用户示例
func banUserExample(service user.UserStatusService, db *gorm.DB, userID, adminID uint) {
	reason := "测试封禁原因"
	duration := 30 * 24 * time.Hour // 30天

	err := service.BanUser(&auth.GORMContext{DB: db}, adminID, userID, reason, duration)
	if err != nil {
		fmt.Printf("  ❌ 封禁失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 用户封禁成功\n")
	fmt.Printf("     用户ID: %d\n", userID)
	fmt.Printf("     封禁原因: %s\n", reason)
	fmt.Printf("     封禁时长: %v\n", duration)
}

// 检查用户是否可登录
func checkUserActiveExample(service user.UserStatusService, db *gorm.DB, userID uint) {
	isActive, err := service.IsUserActive(&auth.GORMContext{DB: db}, userID)
	if err != nil {
		fmt.Printf("  ❌ 检查失败: %v\n", err)
		return
	}

	if isActive {
		fmt.Printf("  ✅ 用户可以正常登录\n")
	} else {
		fmt.Printf("  ❌ 用户被封禁，无法登录\n")
	}
}

// 获取被封禁用户列表示例
func getBannedUsersExample(service user.UserStatusService, db *gorm.DB) {
	users, total, err := service.GetBannedUsers(&auth.GORMContext{DB: db}, 1, 10)
	if err != nil {
		fmt.Printf("  ❌ 获取失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 被封禁用户列表:\n")
	fmt.Printf("     总数: %d\n", total)
	for _, user := range users {
		fmt.Printf("     - %s (%s): %s\n", user.Username, user.Email, user.Status)
	}
}

// 解封用户示例
func unbanUserExample(service user.UserStatusService, db *gorm.DB, userID, adminID uint) {
	reason := "测试解封原因"

	err := service.UnbanUser(&auth.GORMContext{DB: db}, adminID, userID, reason)
	if err != nil {
		fmt.Printf("  ❌ 解封失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 用户解封成功\n")
	fmt.Printf("     用户ID: %d\n", userID)
	fmt.Printf("     解封原因: %s\n", reason)
}

// 激活/禁用用户示例
func activateDeactivateExample(service user.UserStatusService, db *gorm.DB, userID, adminID uint) {
	// 先禁用用户
	deactivateReason := "测试禁用原因"
	err := service.DeactivateUser(&auth.GORMContext{DB: db}, adminID, userID, deactivateReason)
	if err != nil {
		fmt.Printf("  ❌ 禁用失败: %v\n", err)
		return
	}
	fmt.Printf("  ✅ 用户已禁用: %d\n", userID)

	// 再激活用户
	activateReason := "测试激活原因"
	err = service.ActivateUser(&auth.GORMContext{DB: db}, adminID, userID, activateReason)
	if err != nil {
		fmt.Printf("  ❌ 激活失败: %v\n", err)
		return
	}
	fmt.Printf("  ✅ 用户已激活: %d\n", userID)
}

// 批量操作示例
func batchOperationsExample(service user.UserStatusService, authService auth.AuthService, db *gorm.DB, adminID uint) {
	// 创建多个测试用户
	userIDs := createMultipleTestUsers(authService, db)
	if len(userIDs) == 0 {
		fmt.Println("  ℹ️  没有可用的测试用户进行批量操作")
		return
	}

	fmt.Printf("  📝 准备批量封禁 %d 个用户\n", len(userIDs))

	// 批量封禁
	reason := "批量封禁测试"
	duration := 7 * 24 * time.Hour // 7天
	err := service.BatchBanUsers(&auth.GORMContext{DB: db}, adminID, userIDs, reason, duration)
	if err != nil {
		fmt.Printf("  ❌ 批量封禁失败: %v\n", err)
		return
	}
	fmt.Printf("  ✅ 批量封禁成功: %d 个用户\n", len(userIDs))

	// 验证封禁结果
	fmt.Println("  📋 验证封禁结果:")
	for _, userID := range userIDs {
		isActive, _ := service.IsUserActive(&auth.GORMContext{DB: db}, userID)
		if !isActive {
			fmt.Printf("     用户 %d: 已封禁\n", userID)
		}
	}

	// 批量解封
	fmt.Println("  📝 批量解封用户...")
	unbanReason := "批量解封测试"
	err = service.BatchUnbanUsers(&auth.GORMContext{DB: db}, adminID, userIDs, unbanReason)
	if err != nil {
		fmt.Printf("  ❌ 批量解封失败: %v\n", err)
		return
	}
	fmt.Printf("  ✅ 批量解封成功: %d 个用户\n", len(userIDs))
}

// 创建多个测试用户
func createMultipleTestUsers(authService auth.AuthService, db *gorm.DB) []uint {
	var userIDs []uint

	for i := 1; i <= 3; i++ {
		username := fmt.Sprintf("batchtest%d", i)
		email := fmt.Sprintf("batchtest%d@example.com", i)

		// 检查用户是否已存在
		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error == nil {
			userIDs = append(userIDs, user.ID)
			continue
		}

		// 注册新用户
		req := &auth.RegisterRequest{
			Username:        username,
			Email:           email,
			Password:        "password123",
			ConfirmPassword: "password123",
		}

		response, err := authService.Register(&auth.GORMContext{DB: db}, req)
		if err != nil {
			fmt.Printf("  ❌ 创建用户失败 %s: %v\n", username, err)
			continue
		}

		userIDs = append(userIDs, response.ID)
		fmt.Printf("  ✅ 创建测试用户: %s\n", username)
	}

	return userIDs
}
