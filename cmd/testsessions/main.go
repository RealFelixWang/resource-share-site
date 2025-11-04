/*
Session Management Test Examples

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
	"resource-share-site/internal/service/session"

	"gorm.io/gorm"
)

// Session管理测试程序
func main() {
	fmt.Println("=== Session管理测试 ===\n")

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
	sessionService := session.NewSessionService(db, nil) // 不使用Redis
	fmt.Println("✅ 服务创建成功\n")

	// 3. 创建测试用户
	fmt.Println("3. 创建测试用户...")
	testUserID := createTestUser(authService, db)
	fmt.Println()

	// 4. 创建Session
	fmt.Println("4. 创建Session...")
	sessionID := createSessionExample(sessionService, db, testUserID)
	fmt.Println()

	// 5. 获取Session
	fmt.Println("5. 获取Session...")
	getSessionExample(sessionService, db, sessionID)
	fmt.Println()

	// 6. 更新Session
	fmt.Println("6. 更新Session...")
	updateSessionExample(sessionService, db, sessionID)
	fmt.Println()

	// 7. 刷新Session
	fmt.Println("7. 刷新Session...")
	refreshSessionExample(sessionService, db, sessionID)
	fmt.Println()

	// 8. 获取用户Session列表
	fmt.Println("8. 获取用户Session列表...")
	getUserSessionsExample(sessionService, db, testUserID)
	fmt.Println()

	// 9. 验证Session有效性
	fmt.Println("9. 验证Session有效性...")
	checkSessionValidExample(sessionService, db, sessionID)
	fmt.Println()

	// 10. 清理过期Session
	fmt.Println("10. 清理过期Session...")
	cleanExpiredSessionsExample(sessionService, db)
	fmt.Println()

	// 11. 删除Session
	fmt.Println("11. 删除Session...")
	deleteSessionExample(sessionService, db, sessionID)
	fmt.Println()

	// 12. 删除用户所有Session
	fmt.Println("12. 删除用户所有Session...")
	deleteUserSessionsExample(sessionService, db, testUserID)
	fmt.Println()

	fmt.Println("=== 所有测试完成 ===")
}

// 创建测试用户
func createTestUser(authService auth.AuthService, db *gorm.DB) uint {
	var user model.User
	result := db.Where("username = ?", "sessiontest").First(&user)
	if result.Error == nil {
		fmt.Printf("  ✅ 用户已存在: %s\n", user.Username)
		return user.ID
	}

	// 注册新用户
	req := &auth.RegisterRequest{
		Username:        "sessiontest",
		Email:           "sessiontest@example.com",
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

// 创建Session示例
func createSessionExample(service session.SessionService, db *gorm.DB, userID uint) string {
	// 准备会话数据
	data := map[string]interface{}{
		"username":   "sessiontest",
		"role":       "user",
		"login_time": time.Now().Format("2006-01-02 15:04:05"),
		"client":     "web",
	}

	// 创建会话
	sessionInfo, err := service.CreateSession(&auth.GORMContext{DB: db}, userID, data, 24*time.Hour)
	if err != nil {
		fmt.Printf("  ❌ 创建Session失败: %v\n", err)
		return ""
	}

	fmt.Printf("  ✅ Session创建成功\n")
	fmt.Printf("     SessionID: %s\n", sessionInfo.SessionID)
	fmt.Printf("     UserID: %d\n", sessionInfo.UserID)
	fmt.Printf("     过期时间: %v\n", sessionInfo.ExpiresAt)
	fmt.Printf("     会话数据: %v\n", sessionInfo.Data)

	return sessionInfo.SessionID
}

// 获取Session示例
func getSessionExample(service session.SessionService, db *gorm.DB, sessionID string) {
	if sessionID == "" {
		fmt.Println("  ❌ SessionID为空")
		return
	}

	sessionInfo, err := service.GetSession(&auth.GORMContext{DB: db}, sessionID)
	if err != nil {
		fmt.Printf("  ❌ 获取Session失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Session信息:\n")
	fmt.Printf("     SessionID: %s\n", sessionInfo.SessionID)
	fmt.Printf("     UserID: %d\n", sessionInfo.UserID)
	fmt.Printf("     过期时间: %v\n", sessionInfo.ExpiresAt)
	fmt.Printf("     IP: %s\n", sessionInfo.IP)
	fmt.Printf("     创建时间: %v\n", sessionInfo.CreatedAt)
	fmt.Printf("     更新时间: %v\n", sessionInfo.UpdatedAt)
	fmt.Printf("     会话数据: %v\n", sessionInfo.Data)
}

// 更新Session示例
func updateSessionExample(service session.SessionService, db *gorm.DB, sessionID string) {
	if sessionID == "" {
		fmt.Println("  ❌ SessionID为空")
		return
	}

	// 更新会话数据
	newData := map[string]interface{}{
		"last_activity": time.Now().Format("2006-01-02 15:04:05"),
		"page":          "/dashboard",
		"action":        "view",
	}

	err := service.UpdateSession(&auth.GORMContext{DB: db}, sessionID, newData)
	if err != nil {
		fmt.Printf("  ❌ 更新Session失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Session更新成功\n")

	// 验证更新
	sessionInfo, _ := service.GetSession(&auth.GORMContext{DB: db}, sessionID)
	fmt.Printf("     新的会话数据: %v\n", sessionInfo.Data)
}

// 刷新Session示例
func refreshSessionExample(service session.SessionService, db *gorm.DB, sessionID string) {
	if sessionID == "" {
		fmt.Println("  ❌ SessionID为空")
		return
	}

	newDuration := 2 * time.Hour // 延长2小时
	err := service.RefreshSession(&auth.GORMContext{DB: db}, sessionID, newDuration)
	if err != nil {
		fmt.Printf("  ❌ 刷新Session失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Session刷新成功\n")

	// 验证刷新
	sessionInfo, _ := service.GetSession(&auth.GORMContext{DB: db}, sessionID)
	fmt.Printf("     新的过期时间: %v\n", sessionInfo.ExpiresAt)
	fmt.Printf("     剩余时间: %v\n", sessionInfo.ExpiresAt.Sub(time.Now()))
}

// 获取用户Session列表示例
func getUserSessionsExample(service session.SessionService, db *gorm.DB, userID uint) {
	sessions, err := service.GetUserSessions(&auth.GORMContext{DB: db}, userID)
	if err != nil {
		fmt.Printf("  ❌ 获取用户Session失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 用户Session列表 (共%d个):\n", len(sessions))
	for i, session := range sessions {
		fmt.Printf("     [%d] SessionID: %s\n", i+1, session.SessionID)
		fmt.Printf("         过期时间: %v\n", session.ExpiresAt)
		fmt.Printf("         创建时间: %v\n", session.CreatedAt)
	}
}

// 验证Session有效性示例
func checkSessionValidExample(service session.SessionService, db *gorm.DB, sessionID string) {
	if sessionID == "" {
		fmt.Println("  ❌ SessionID为空")
		return
	}

	isValid, err := service.IsSessionValid(&auth.GORMContext{DB: db}, sessionID)
	if err != nil {
		fmt.Printf("  ❌ 验证失败: %v\n", err)
		return
	}

	if isValid {
		fmt.Printf("  ✅ Session有效\n")
	} else {
		fmt.Printf("  ❌ Session无效或已过期\n")
	}
}

// 清理过期Session示例
func cleanExpiredSessionsExample(service session.SessionService, db *gorm.DB) {
	deletedCount, err := service.CleanExpiredSessions(&auth.GORMContext{DB: db})
	if err != nil {
		fmt.Printf("  ❌ 清理失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 清理完成\n")
	fmt.Printf("     删除了 %d 个过期Session\n", deletedCount)
}

// 删除Session示例
func deleteSessionExample(service session.SessionService, db *gorm.DB, sessionID string) {
	if sessionID == "" {
		fmt.Println("  ❌ SessionID为空")
		return
	}

	err := service.DeleteSession(&auth.GORMContext{DB: db}, sessionID)
	if err != nil {
		fmt.Printf("  ❌ 删除Session失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Session删除成功: %s\n", sessionID)

	// 验证删除
	_, err = service.GetSession(&auth.GORMContext{DB: db}, sessionID)
	if err != nil {
		fmt.Printf("  ✅ 验证: Session已不存在\n")
	}
}

// 删除用户所有Session示例
func deleteUserSessionsExample(service session.SessionService, db *gorm.DB, userID uint) {
	err := service.DeleteUserSessions(&auth.GORMContext{DB: db}, userID)
	if err != nil {
		fmt.Printf("  ❌ 删除用户Session失败: %v\n", err)
		return
	}

	fmt.Printf("  ✅ 用户所有Session删除成功: %d\n", userID)

	// 验证删除
	sessions, _ := service.GetUserSessions(&auth.GORMContext{DB: db}, userID)
	fmt.Printf("     验证: 用户剩余Session数: %d\n", len(sessions))
}
