/*
Auth Service Usage Examples

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package main

import (
	"fmt"

	"resource-share-site/internal/config"
	"resource-share-site/internal/database"
	"resource-share-site/internal/model"
	"resource-share-site/internal/service/auth"

	"gorm.io/gorm"
)

// 认证服务使用示例
func main() {
	fmt.Println("=== 认证服务使用示例 ===\n")

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

	// 2. 创建认证服务
	fmt.Println("2. 创建认证服务...")
	authService := auth.NewAuthService(db)
	fmt.Println("✅ 认证服务创建成功\n")

	// 3. 注册示例
	fmt.Println("3. 用户注册示例...")
	registerExample(authService, db)
	fmt.Println()

	// 4. 登录示例（支持用户名）
	fmt.Println("4. 使用用户名登录示例...")
	loginWithUsernameExample(authService, db)
	fmt.Println()

	// 5. 登录示例（支持邮箱）
	fmt.Println("5. 使用邮箱登录示例...")
	loginWithEmailExample(authService, db)
	fmt.Println()

	// 6. 修改密码示例
	fmt.Println("6. 修改密码示例...")
	changePasswordExample(authService, db)
	fmt.Println()

	// 7. 更新资料示例
	fmt.Println("7. 更新用户资料示例...")
	updateProfileExample(authService, db)
	fmt.Println()

	fmt.Println("=== 所有示例演示完成 ===")
}

// 注册示例
func registerExample(authService auth.AuthService, db *gorm.DB) {
	// 检查用户是否已存在
	var user model.User
	result := db.Where("username = ? OR email = ?", "testuser", "test@example.com").First(&user)
	if result.Error == nil {
		fmt.Printf("  ✅ 用户已存在: %s (%s)\n", user.Username, user.Email)
		return
	}

	// 注册新用户
	req := &auth.RegisterRequest{
		Username:        "testuser",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
		InviteCode:      "", // 可选
	}

	response, err := authService.Register(&auth.GORMContext{DB: db}, req)
	if err != nil {
		fmt.Printf("  ❌ 注册失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 注册成功\n")
	fmt.Printf("     用户ID: %d\n", response.ID)
	fmt.Printf("     用户名: %s\n", response.Username)
	fmt.Printf("     邮箱: %s\n", response.Email)
	fmt.Printf("     邀请码: %s\n", response.InviteCode)
	fmt.Printf("     积分余额: %d\n", response.PointsBalance)
}

// 使用用户名登录示例
func loginWithUsernameExample(authService auth.AuthService, db *gorm.DB) {
	req := &auth.LoginRequest{
		Identifier: "testuser", // 使用用户名
		Password:   "password123",
		Remember:   false,
	}

	response, err := authService.Login(&auth.GORMContext{DB: db}, req)
	if err != nil {
		fmt.Printf("  ❌ 登录失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 登录成功\n")
	fmt.Printf("     Token: %s\n", response.Token)
	fmt.Printf("     过期时间: %v\n", response.ExpiresAt)
	fmt.Printf("     用户名: %s\n", response.User.Username)
	fmt.Printf("     邮箱: %s\n", response.User.Email)
	fmt.Printf("     角色: %s\n", response.User.Role)
	fmt.Printf("     积分余额: %d\n", response.User.PointsBalance)
}

// 使用邮箱登录示例
func loginWithEmailExample(authService auth.AuthService, db *gorm.DB) {
	req := &auth.LoginRequest{
		Identifier: "test@example.com", // 使用邮箱
		Password:   "password123",
		Remember:   true, // 记住登录状态
	}

	response, err := authService.Login(&auth.GORMContext{DB: db}, req)
	if err != nil {
		fmt.Printf("  ❌ 登录失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 登录成功\n")
	fmt.Printf("     Token: %s\n", response.Token)
	fmt.Printf("     过期时间: %v\n", response.ExpiresAt)
	fmt.Printf("     用户名: %s\n", response.User.Username)
	fmt.Printf("     邮箱: %s\n", response.User.Email)
	fmt.Printf("     记住登录: %v\n", req.Remember)
}

// 修改密码示例
func changePasswordExample(authService auth.AuthService, db *gorm.DB) {
	// 获取用户
	var user model.User
	if err := db.Where("username = ?", "testuser").First(&user).Error; err != nil {
		fmt.Printf("  ❌ 用户不存在: %v\n", err)
		return
	}

	req := &auth.ChangePasswordRequest{
		OldPassword:     "password123",
		NewPassword:     "newpassword456",
		ConfirmPassword: "newpassword456",
	}

	if err := authService.ChangePassword(&auth.GORMContext{DB: db}, user.ID, req); err != nil {
		fmt.Printf("  ❌ 修改密码失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 密码修改成功\n")

	// 验证新密码
	loginReq := &auth.LoginRequest{
		Identifier: "testuser",
		Password:   "newpassword456",
	}

	if _, err := authService.Login(&auth.GORMContext{DB: db}, loginReq); err != nil {
		fmt.Printf("  ❌ 新密码验证失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 新密码验证成功\n")
	}
}

// 更新资料示例
func updateProfileExample(authService auth.AuthService, db *gorm.DB) {
	// 获取用户
	var user model.User
	if err := db.Where("username = ?", "testuser").First(&user).Error; err != nil {
		fmt.Printf("  ❌ 用户不存在: %v\n", err)
		return
	}

	req := &auth.UpdateProfileRequest{
		Username: "updateduser",         // 更新用户名
		Email:    "updated@example.com", // 更新邮箱
	}

	if err := authService.UpdateProfile(&auth.GORMContext{DB: db}, user.ID, req); err != nil {
		fmt.Printf("  ❌ 更新资料失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 资料更新成功\n")

	// 验证更新
	loginReq := &auth.LoginRequest{
		Identifier: "updateduser", // 使用新用户名
		Password:   "newpassword456",
	}

	if response, err := authService.Login(&auth.GORMContext{DB: db}, loginReq); err != nil {
		fmt.Printf("  ❌ 新用户名登录失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 新用户名登录成功\n")
		fmt.Printf("     新用户名: %s\n", response.User.Username)
		fmt.Printf("     新邮箱: %s\n", response.User.Email)
	}
}
