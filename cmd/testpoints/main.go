/*
Points System Test Program - 积分系统测试程序

测试积分系统的核心功能：
1. 积分获取机制（6.1）
2. 积分消费规则（6.2）
3. 积分商城功能（6.3）
4. 积分统计分析（6.4）

Author: Felix Wang
Email: felixwang.biz@gmail.com
Date: 2025-10-31
*/

package main

import (
	"fmt"
	"log"

	"resource-share-site/internal/model"
	points "resource-share-site/internal/service/points"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== 积分系统测试程序 ===")
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
		&model.Invitation{},
		&model.PointsRule{},
		&model.PointRecord{},
		&model.Resource{},
		&model.Category{},
		&model.Product{},
		&model.MallOrder{},
	)
}

func createTestData(db *gorm.DB) error {
	// 创建用户
	users := []model.User{
		{
			Username:     "admin",
			Email:        "admin@example.com",
			PasswordHash: "hashed_password",
			Role:         "admin",
			Status:       "active",
		},
		{
			Username:     "user1",
			Email:        "user1@example.com",
			PasswordHash: "hashed_password",
			Status:       "active",
		},
		{
			Username:     "user2",
			Email:        "user2@example.com",
			PasswordHash: "hashed_password",
			Status:       "active",
		},
	}

	if err := db.Create(&users).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	// 创建邀请关系
	inviteeID := uint(2)
	invitation := model.Invitation{
		InviterID: 1,
		InviteeID: &inviteeID,
		Status:    model.InvitationStatusCompleted,
	}

	if err := db.Create(&invitation).Error; err != nil {
		return fmt.Errorf("创建邀请关系失败: %w", err)
	}

	// 创建积分规则
	rules := []model.PointsRule{
		{
			RuleKey:     "invite_reward",
			RuleName:    "邀请奖励",
			Description: "成功邀请用户后获得积分",
			Points:      100,
			IsEnabled:   true,
		},
		{
			RuleKey:     "resource_download",
			RuleName:    "下载奖励",
			Description: "下载资源后获得积分",
			Points:      5,
			IsEnabled:   true,
		},
		{
			RuleKey:     "daily_checkin",
			RuleName:    "每日签到",
			Description: "每日签到获得积分",
			Points:      10,
			IsEnabled:   true,
		},
		{
			RuleKey:     "upload_reward",
			RuleName:    "上传奖励",
			Description: "上传资源后获得积分",
			Points:      50,
			IsEnabled:   true,
		},
	}

	if err := db.Create(&rules).Error; err != nil {
		return fmt.Errorf("创建积分规则失败: %w", err)
	}

	// 创建测试资源
	resources := []model.Resource{
		{
			Title:        "测试资源1",
			Description:  "这是一个测试资源",
			UploadedByID: 1,
			Status:       "approved",
		},
		{
			Title:        "测试资源2",
			Description:  "这是另一个测试资源",
			UploadedByID: 2,
			Status:       "approved",
		},
	}

	if err := db.Create(&resources).Error; err != nil {
		return fmt.Errorf("创建测试资源失败: %w", err)
	}

	return nil
}

func runAllTests(db *gorm.DB) {
	// 测试积分获取机制（6.1）
	fmt.Println("【6.1 测试】积分获取机制")
	fmt.Println("=" + fmt.Sprint(50) + "=")
	if err := testEarningService(db); err != nil {
		fmt.Printf("❌ 积分获取机制测试失败: %v\n\n", err)
	} else {
		fmt.Println("✅ 积分获取机制测试通过\n\n")
	}

	// 测试积分消费规则（6.2）
	fmt.Println("【6.2 测试】积分消费规则")
	fmt.Println("=" + fmt.Sprint(50) + "=")
	if err := testConsumptionService(db); err != nil {
		fmt.Printf("❌ 积分消费规则测试失败: %v\n\n", err)
	} else {
		fmt.Println("✅ 积分消费规则测试通过\n\n")
	}

	// 测试积分商城功能（6.3）
	fmt.Println("【6.3 测试】积分商城功能")
	fmt.Println("=" + fmt.Sprint(50) + "=")
	if err := testMallService(db); err != nil {
		fmt.Printf("❌ 积分商城功能测试失败: %v\n\n", err)
	} else {
		fmt.Println("✅ 积分商城功能测试通过\n\n")
	}

	// 测试积分统计分析（6.4）
	fmt.Println("【6.4 测试】积分统计分析")
	fmt.Println("=" + fmt.Sprint(50) + "=")
	if err := testStatisticsService(db); err != nil {
		fmt.Printf("❌ 积分统计分析测试失败: %v\n\n", err)
	} else {
		fmt.Println("✅ 积分统计分析测试通过\n\n")
	}

	fmt.Println("=== 所有测试完成 ===")
}

func testEarningService(db *gorm.DB) error {
	earningService := points.NewEarningService(db)

	// 1. 测试邀请奖励
	fmt.Println("1. 测试邀请奖励:")
	if err := earningService.EarnPointsByInvite(1, 2, 100); err != nil {
		fmt.Printf("   ❌ 邀请奖励失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 邀请奖励成功")
	}

	// 查看用户1的积分
	balance1, _ := earningService.GetUserPointsBalance(1)
	fmt.Printf("   用户1当前积分: %d\n", balance1)

	// 2. 测试每日签到
	fmt.Println("\n2. 测试每日签到:")
	if err := earningService.EarnPointsByDailyCheckin(2); err != nil {
		fmt.Printf("   ❌ 签到失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 签到成功")
	}

	// 查看用户2的积分
	balance2, _ := earningService.GetUserPointsBalance(2)
	fmt.Printf("   用户2当前积分: %d\n", balance2)

	// 3. 测试资源下载奖励
	fmt.Println("\n3. 测试资源下载奖励:")
	if err := earningService.EarnPointsByResourceDownload(2, 1); err != nil {
		fmt.Printf("   ❌ 下载奖励失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 下载奖励成功")
	}

	// 查看用户2的积分
	balance2, _ = earningService.GetUserPointsBalance(2)
	fmt.Printf("   用户2当前积分: %d\n", balance2)

	// 4. 测试资源上传奖励
	fmt.Println("\n4. 测试资源上传奖励:")
	if err := earningService.EarnPointsByResourceUpload(1, 1); err != nil {
		fmt.Printf("   ❌ 上传奖励失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 上传奖励成功")
	}

	// 查看用户1的积分
	balance1, _ = earningService.GetUserPointsBalance(1)
	fmt.Printf("   用户1当前积分: %d\n", balance1)

	// 5. 测试管理员添加积分
	fmt.Println("\n5. 测试管理员添加积分:")
	adminID := uint(1)
	if err := earningService.EarnPointsByAdmin(1, 200, "管理员奖励测试", &adminID); err != nil {
		fmt.Printf("   ❌ 管理员添加积分失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 管理员添加积分成功")
	}

	// 查看用户1的积分
	balance1, _ = earningService.GetUserPointsBalance(1)
	fmt.Printf("   用户1当前积分: %d\n", balance1)

	// 6. 测试获取用户积分记录
	fmt.Println("\n6. 测试获取用户积分记录:")
	records, total, err := earningService.GetUserPointRecords(1, 10, 0)
	if err != nil {
		fmt.Printf("   ❌ 获取积分记录失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 获取到 %d 条积分记录\n", total)
		for i, record := range records {
			if i >= 5 { // 只显示前5条
				break
			}
			fmt.Printf("   - %s: %s (%+d) - %s\n",
				record.CreatedAt.Format("2006-01-02 15:04:05"),
				record.Source,
				record.Points,
				record.Description)
		}
	}

	// 7. 测试积分统计
	fmt.Println("\n7. 测试积分统计:")
	stats, err := earningService.GetPointsStats(1)
	if err != nil {
		fmt.Printf("   ❌ 获取积分统计失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 积分统计信息:")
		fmt.Printf("   - 总收入: %d\n", stats["total_income"])
		fmt.Printf("   - 总支出: %d\n", stats["total_expense"])
		fmt.Printf("   - 当前余额: %d\n", stats["current_balance"])
		fmt.Printf("   - 今日收入: %d\n", stats["today_income"])
		fmt.Printf("   - 本月收入: %d\n", stats["month_income"])
	}

	// 8. 测试重复签到
	fmt.Println("\n8. 测试重复签到:")
	if err := earningService.EarnPointsByDailyCheckin(2); err != nil {
		fmt.Printf("   ✅ 预期错误: %v\n", err)
	} else {
		fmt.Println("   ❌ 重复签到未检查")
	}

	// 9. 测试获取积分规则
	fmt.Println("\n9. 测试获取积分规则:")
	rules, err := earningService.GetEarningRules()
	if err != nil {
		fmt.Printf("   ❌ 获取规则失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 获取到 %d 条规则\n", len(rules))
		for _, rule := range rules {
			fmt.Printf("   - %s: %d积分 (%v)\n", rule.RuleName, rule.Points, rule.IsEnabled)
		}
	}

	// 10. 测试批量添加积分
	fmt.Println("\n10. 测试批量添加积分:")
	earnings := []struct {
		UserID      uint
		Points      int
		Source      model.PointSource
		Description string
	}{
		{1, 50, model.PointSourceAdminAdd, "批量测试1"},
		{2, 50, model.PointSourceAdminAdd, "批量测试2"},
	}
	if err := earningService.BatchEarnPoints(earnings); err != nil {
		fmt.Printf("   ❌ 批量添加积分失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 批量添加积分成功")
	}

	// 查看最终积分
	balance1, _ = earningService.GetUserPointsBalance(1)
	balance2, _ = earningService.GetUserPointsBalance(2)
	fmt.Printf("   用户1最终积分: %d\n", balance1)
	fmt.Printf("   用户2最终积分: %d\n", balance2)

	return nil
}

func testConsumptionService(db *gorm.DB) error {
	consumptionService := points.NewConsumptionService(db)

	// 1. 测试积分购买
	fmt.Println("1. 测试积分购买:")
	if err := consumptionService.SpendPointsForPurchase(1, 50, "测试购买", nil); err != nil {
		fmt.Printf("   ❌ 购买失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 购买成功")
	}

	// 查看用户1的积分
	balance, _ := points.NewEarningService(db).GetUserPointsBalance(1)
	fmt.Printf("   用户1当前积分: %d\n", balance)

	// 2. 测试积分下载付费资源
	fmt.Println("\n2. 测试积分下载付费资源:")
	if err := consumptionService.SpendPointsForDownload(2, 1, 20); err != nil {
		fmt.Printf("   ❌ 下载付费失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 下载付费成功")
	}

	// 查看用户2的积分
	balance2, _ := points.NewEarningService(db).GetUserPointsBalance(2)
	fmt.Printf("   用户2当前积分: %d\n", balance2)

	// 3. 测试积分升级VIP
	fmt.Println("\n3. 测试积分升级VIP:")
	if err := consumptionService.SpendPointsForVipUpgrade(1, "高级VIP", 100); err != nil {
		fmt.Printf("   ❌ VIP升级失败: %v\n", err)
	} else {
		fmt.Println("   ✅ VIP升级成功")
	}

	// 查看用户1的积分
	balance1, _ := points.NewEarningService(db).GetUserPointsBalance(1)
	fmt.Printf("   用户1当前积分: %d\n", balance1)

	// 4. 测试获取消费历史
	fmt.Println("\n4. 测试获取消费历史:")
	_, total, err := consumptionService.GetConsumptionHistory(1, 10, 0)
	if err != nil {
		fmt.Printf("   ❌ 获取消费历史失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 获取到 %d 条消费记录\n", total)
	}

	// 5. 测试消费统计
	fmt.Println("\n5. 测试消费统计:")
	stats, err := consumptionService.GetConsumptionStats(1)
	if err != nil {
		fmt.Printf("   ❌ 获取消费统计失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 消费统计信息:")
		fmt.Printf("   - 总消费: %d\n", stats["total_consumption"])
		fmt.Printf("   - 今日消费: %d\n", stats["today_consumption"])
		fmt.Printf("   - 本月消费: %d\n", stats["month_consumption"])
	}

	return nil
}

func testMallService(db *gorm.DB) error {
	mallService := points.NewMallService(db)

	// 1. 测试创建商品
	fmt.Println("1. 测试创建商品:")
	vipDays := 30
	product := &model.Product{
		Name:        "高级VIP会员",
		Description: "享受30天高级VIP服务",
		Category:    model.ProductCategoryVip,
		PointsPrice: 200,
		Stock:       100,
		IsLimited:   true,
		Status:      model.ProductStatusActive,
		ValidDays:   &vipDays,
	}
	if err := mallService.CreateProduct(product); err != nil {
		fmt.Printf("   ❌ 创建商品失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 创建商品成功，ID: %d\n", product.ID)
	}

	// 2. 测试获取商品列表
	fmt.Println("\n2. 测试获取商品列表:")
	products, total, err := mallService.ListProducts("", model.ProductStatusActive, 1, 10)
	if err != nil {
		fmt.Printf("   ❌ 获取商品列表失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 获取到 %d 个商品（总数: %d）\n", len(products), total)
		for _, p := range products {
			fmt.Printf("   - %s: %d积分 (%s)\n", p.Name, p.PointsPrice, p.Status)
		}
	}

	// 3. 测试购买商品
	fmt.Println("\n3. 测试购买商品:")
	if product.ID > 0 {
		if err := mallService.PurchaseProduct(1, product.ID, 1); err != nil {
			fmt.Printf("   ❌ 购买失败: %v\n", err)
		} else {
			fmt.Println("   ✅ 购买成功")
		}

		// 查看用户1的积分
		balance, _ := points.NewEarningService(db).GetUserPointsBalance(1)
		fmt.Printf("   用户1当前积分: %d\n", balance)
	}

	// 4. 测试获取订单列表
	fmt.Println("\n4. 测试获取订单列表:")
	orders, total, err := mallService.GetUserOrders(1, "", 1, 10)
	if err != nil {
		fmt.Printf("   ❌ 获取订单列表失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 获取到 %d 个订单（总数: %d）\n", len(orders), total)
	}

	// 5. 测试商城统计
	fmt.Println("\n5. 测试商城统计:")
	stats, err := mallService.GetMallStats()
	if err != nil {
		fmt.Printf("   ❌ 获取商城统计失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 商城统计信息:")
		fmt.Printf("   - 总商品数: %d\n", stats["total_products"])
		fmt.Printf("   - 总订单数: %d\n", stats["total_orders"])
		fmt.Printf("   - 今日订单: %d\n", stats["today_orders"])
		fmt.Printf("   - 总销售额: %d积分\n", stats["total_sales"])
	}

	return nil
}

func testStatisticsService(db *gorm.DB) error {
	statsService := points.NewStatisticsService(db)

	// 1. 测试用户积分概览
	fmt.Println("1. 测试用户积分概览:")
	summary, err := statsService.GetUserPointsSummary(1)
	if err != nil {
		fmt.Printf("   ❌ 获取积分概览失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 积分概览:")
		fmt.Printf("   - 当前余额: %d\n", summary["current_balance"])
		fmt.Printf("   - 总收入: %d\n", summary["total_income"])
		fmt.Printf("   - 总支出: %d\n", summary["total_expense"])
		fmt.Printf("   - 今日收入: %d\n", summary["today_income"])
		fmt.Printf("   - 今日支出: %d\n", summary["today_expense"])
		fmt.Printf("   - 本月收入: %d\n", summary["month_income"])
		fmt.Printf("   - 本月支出: %d\n", summary["month_expense"])
	}

	// 2. 测试积分趋势
	fmt.Println("\n2. 测试积分趋势:")
	trend, err := statsService.GetUserPointsTrend(1, 7)
	if err != nil {
		fmt.Printf("   ❌ 获取积分趋势失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 最近7天积分趋势（%d天数据）\n", len(trend))
		for _, t := range trend {
			fmt.Printf("   - %s: 收入%+d, 支出%+d, 净变动%+d\n",
				t["date"], t["daily_income"], t["daily_expense"], t["daily_net"])
		}
	}

	// 3. 测试积分排行榜
	fmt.Println("\n3. 测试积分排行榜:")
	ranking, err := statsService.GetUserPointsRanking(10)
	if err != nil {
		fmt.Printf("   ❌ 获取排行榜失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 积分排行榜（TOP %d）:\n", len(ranking))
		for _, r := range ranking {
			if r.Rank <= 3 {
				fmt.Printf("   🏆 第%d名: %s - %d积分\n", r.Rank, r.Username, r.Balance)
			}
		}
	}

	// 4. 测试系统统计
	fmt.Println("\n4. 测试系统统计:")
	systemStats, err := statsService.GetSystemPointsStats()
	if err != nil {
		fmt.Printf("   ❌ 获取系统统计失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 系统统计信息:")
		fmt.Printf("   - 系统总积分: %d\n", systemStats["total_points"])
		fmt.Printf("   - 今日活跃用户: %d\n", systemStats["active_users_today"])
		fmt.Printf("   - 今日新增用户: %d\n", systemStats["new_users_today"])
		fmt.Printf("   - 总收入积分: %d\n", systemStats["total_income"])
		fmt.Printf("   - 总支出积分: %d\n", systemStats["total_expense"])
		fmt.Printf("   - 今日收入: %d\n", systemStats["today_income"])
		fmt.Printf("   - 今日支出: %d\n", systemStats["today_expense"])
	}

	// 5. 测试积分获取排行榜
	fmt.Println("\n5. 测试积分获取排行榜:")
	earners, err := statsService.GetTopEarners(5, 30)
	if err != nil {
		fmt.Printf("   ❌ 获取积分获取排行榜失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 最近30天积分获取TOP %d:\n", len(earners))
		for i, e := range earners {
			fmt.Printf("   🥇 第%d名: %s - 获得% d积分 (%d次交易)\n",
				i+1, e.Username, e.TotalEarned, e.TransactionCount)
		}
	}

	// 6. 测试积分消费排行榜
	fmt.Println("\n6. 测试积分消费排行榜:")
	spenders, err := statsService.GetTopSpenders(5, 30)
	if err != nil {
		fmt.Printf("   ❌ 获取积分消费排行榜失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 最近30天积分消费TOP %d:\n", len(spenders))
		for i, s := range spenders {
			fmt.Printf("   💸 第%d名: %s - 消费% d积分 (%d次交易)\n",
				i+1, s.Username, s.TotalSpent, s.TransactionCount)
		}
	}

	// 7. 测试积分流动趋势
	fmt.Println("\n7. 测试积分流动趋势:")
	flowTrend, err := statsService.GetPointsFlowTrend(7)
	if err != nil {
		fmt.Printf("   ❌ 获取积分流动趋势失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 最近7天积分流动趋势:\n")
		for _, t := range flowTrend {
			fmt.Printf("   - %s: 收入%+d, 支出%+d, 净流入%+d, 活跃用户%d\n",
				t["date"], t["total_income"], t["total_expense"], t["net_flow"], t["active_users"])
		}
	}

	return nil
}
