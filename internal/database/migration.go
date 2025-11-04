/*
Package database provides database initialization and migration functions.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package database

import (
	"fmt"

	"resource-share-site/internal/config"
	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// RunMigrations 执行数据库迁移
func RunMigrations(db *gorm.DB) error {
	fmt.Println("开始执行数据库迁移...")

	// 获取所有模型进行自动迁移
	if err := AutoMigrate(db); err != nil {
		return fmt.Errorf("自动迁移失败: %w", err)
	}

	fmt.Println("数据库迁移完成!")
	return nil
}

// AutoMigrate 自动迁移所有模型
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// 用户系统
		&model.User{},
		&model.Session{},

		// 分类系统
		&model.Category{},

		// 资源系统
		&model.Resource{},
		&model.Comment{},

		// 邀请系统
		&model.Invitation{},

		// 积分系统
		&model.PointsRule{},
		&model.PointRecord{},

		// 监控审计
		&model.VisitLog{},
		&model.IPBlacklist{},
		&model.AdminLog{},

		// 系统管理
		&model.Ad{},
		&model.Permission{},
		&model.ImportTask{},
	)
}

// InitDatabaseWithConfig 初始化数据库连接
func InitDatabaseWithConfig(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	db, err := config.InitDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库连接失败: %w", err)
	}

	// 执行迁移
	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	return db, nil
}

// CreateDefaultData 创建默认数据
func CreateDefaultData(db *gorm.DB) error {
	fmt.Println("开始创建默认数据...")

	// 检查是否已存在管理员用户
	var count int64
	db.Model(&model.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		fmt.Println("默认管理员已存在，跳过创建")
		return nil
	}

	// 创建默认管理员用户
	adminUser := model.User{
		Username:      "admin",
		Email:         "felixwang.biz@gmail.com",
		PasswordHash:  "$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iK.vpDh6Y5wR.MZCP.mC.bC5dv", // admin123
		Role:          "admin",
		Status:        "active",
		CanUpload:     true,
		InviteCode:    "ADMIN0000000000000000000000000000000000",
		PointsBalance: 0,
	}

	if err := db.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("创建默认管理员失败: %w", err)
	}

	// 创建默认积分规则
	pointsRules := []model.PointsRule{
		{RuleKey: "invite_reward", RuleName: "邀请奖励", Description: "成功邀请一个用户注册", Points: 50, IsEnabled: true},
		{RuleKey: "resource_download", RuleName: "资源下载", Description: "下载需要积分的资源", Points: -10, IsEnabled: true},
		{RuleKey: "daily_checkin", RuleName: "每日签到", Description: "每日登录奖励", Points: 5, IsEnabled: true},
		{RuleKey: "upload_reward", RuleName: "上传奖励", Description: "审核通过一个资源", Points: 10, IsEnabled: true},
	}

	for _, rule := range pointsRules {
		var existingRule model.PointsRule
		if err := db.Where("rule_key = ?", rule.RuleKey).First(&existingRule).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&rule).Error; err != nil {
					return fmt.Errorf("创建积分规则失败: %w", err)
				}
			} else {
				return fmt.Errorf("查询积分规则失败: %w", err)
			}
		}
	}

	// 创建默认权限
	permissions := []model.Permission{
		{Key: "user.upload", Name: "用户上传", Description: "允许用户上传资源", IsEnabled: true},
		{Key: "user.comment", Name: "用户评论", Description: "允许用户评论资源", IsEnabled: true},
		{Key: "admin.review", Name: "资源审核", Description: "允许审核资源", IsEnabled: true},
		{Key: "admin.ban_user", Name: "封禁用户", Description: "允许封禁/解封用户", IsEnabled: true},
		{Key: "admin.ip_ban", Name: "IP封禁", Description: "允许封禁IP地址", IsEnabled: true},
		{Key: "admin.manage_ads", Name: "广告管理", Description: "允许管理广告", IsEnabled: true},
		{Key: "admin.view_logs", Name: "查看日志", Description: "允许查看系统日志", IsEnabled: true},
		{Key: "admin.import", Name: "导入数据", Description: "允许导入资源数据", IsEnabled: true},
	}

	for _, perm := range permissions {
		var existingPerm model.Permission
		if err := db.Where("key = ?", perm.Key).First(&existingPerm).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&perm).Error; err != nil {
					return fmt.Errorf("创建权限失败: %w", err)
				}
			} else {
				return fmt.Errorf("查询权限失败: %w", err)
			}
		}
	}

	// 创建默认分类
	categories := []model.Category{
		{Name: "软件工具", Description: "各类实用软件和工具", Icon: "software.png", Color: "#3498db", SortOrder: 1},
		{Name: "电子资料", Description: "电子书、文档、教程等学习资料", Icon: "document.png", Color: "#2ecc71", SortOrder: 2},
		{Name: "多媒体", Description: "电影、音乐、图片等娱乐资源", Icon: "multimedia.png", Color: "#e74c3c", SortOrder: 3},
		{Name: "游戏", Description: "游戏资源、游戏补丁等", Icon: "game.png", Color: "#9b59b6", SortOrder: 4},
		{Name: "其他", Description: "其他类型的资源", Icon: "other.png", Color: "#95a5a6", SortOrder: 5},
	}

	for _, category := range categories {
		var existingCategory model.Category
		if err := db.Where("name = ?", category.Name).First(&existingCategory).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&category).Error; err != nil {
					return fmt.Errorf("创建分类失败: %w", err)
				}
			} else {
				return fmt.Errorf("查询分类失败: %w", err)
			}
		}
	}

	fmt.Println("默认数据创建完成!")
	return nil
}

// RollbackMigrations 回滚迁移（谨慎使用）
func RollbackMigrations(db *gorm.DB) error {
	fmt.Println("开始回滚数据库迁移...")

	// 注意：这是危险操作，会删除所有数据
	if err := db.Migrator().DropTable(
		"users",
		"sessions",
		"categories",
		"resources",
		"comments",
		"invitations",
		"points_rules",
		"point_records",
		"visit_logs",
		"ip_blacklists",
		"admin_logs",
		"ads",
		"permissions",
		"import_tasks",
	).Error; err != nil {
		return fmt.Errorf("回滚迁移失败: %w", err)
	}

	fmt.Println("数据库迁移回滚完成!")
	return nil
}

// GetMigrationStatus 获取迁移状态
func GetMigrationStatus(db *gorm.DB) error {
	// 获取所有表
	var tables []string
	if err := db.Raw("SHOW TABLES").Scan(&tables).Error; err != nil {
		return fmt.Errorf("查询表列表失败: %w", err)
	}

	fmt.Println("=== 数据库表列表 ===")
	for i, table := range tables {
		fmt.Printf("%d. %s\n", i+1, table)
	}

	// 获取用户数量
	var userCount int64
	if err := db.Model(&model.User{}).Count(&userCount).Error; err != nil {
		return fmt.Errorf("查询用户数量失败: %w", err)
	}
	fmt.Printf("\n用户总数: %d\n", userCount)

	// 获取资源数量
	var resourceCount int64
	if err := db.Model(&model.Resource{}).Count(&resourceCount).Error; err != nil {
		return fmt.Errorf("查询资源数量失败: %w", err)
	}
	fmt.Printf("资源总数: %d\n", resourceCount)

	// 获取分类数量
	var categoryCount int64
	if err := db.Model(&model.Category{}).Count(&categoryCount).Error; err != nil {
		return fmt.Errorf("查询分类数量失败: %w", err)
	}
	fmt.Printf("分类总数: %d\n", categoryCount)

	return nil
}
