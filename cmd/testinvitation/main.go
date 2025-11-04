/*
Test program for invitation system.

This program tests all invitation services:
- InvitationService: Invitation code generation and validation
- RelationshipService: Invitation relationship tracking
- RewardService: Invitation reward mechanism
- LeaderboardService: Invitation leaderboard

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"resource-share-site/internal/config"
	"resource-share-site/internal/database"

	"resource-share-site/internal/model"
	"resource-share-site/internal/service/auth"
	"resource-share-site/internal/service/invitation"
	"resource-share-site/internal/service/session"
	"resource-share-site/internal/service/user"

	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== 邀请系统测试开始 ===\n")

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
	authService := auth.NewAuthService(db)
	userStatusService := user.NewUserStatusService(db)
	sessionService := session.NewSessionService(db, nil) // Redis 可选
	invitationService := invitation.NewInvitationService(db)
	relationshipService := invitation.NewRelationshipService(db)
	rewardService := invitation.NewRewardService(db)
	leaderboardService := invitation.NewLeaderboardService(db)
	fmt.Println("✅ 服务初始化成功\n")

	// 1. 测试用户注册
	fmt.Println("3. 测试用户注册...")
	testUserRegistration(db, authService, userStatusService)
	time.Sleep(100 * time.Millisecond)

	// 2. 测试邀请码生成和验证
	fmt.Println("\n4. 测试邀请码生成和验证...")
	testInviteCodeGeneration(db, invitationService, userStatusService)
	time.Sleep(100 * time.Millisecond)

	// 3. 测试邀请关系追踪
	fmt.Println("\n5. 测试邀请关系追踪...")
	testInvitationRelationships(db, relationshipService, invitationService, userStatusService)
	time.Sleep(100 * time.Millisecond)

	// 4. 测试邀请奖励机制
	fmt.Println("\n6. 测试邀请奖励机制...")
	testInvitationRewards(db, rewardService, invitationService, userStatusService)
	time.Sleep(100 * time.Millisecond)

	// 5. 测试邀请排行榜
	fmt.Println("\n7. 测试邀请排行榜...")
	testInvitationLeaderboard(db, leaderboardService)
	time.Sleep(100 * time.Millisecond)

	// 6. 测试邀请统计
	fmt.Println("\n8. 测试邀请统计...")
	testInvitationStats(db, invitationService, relationshipService, rewardService)
	time.Sleep(100 * time.Millisecond)

	// 7. 测试邀请树结构
	fmt.Println("\n9. 测试邀请树结构...")
	testInvitationTree(db, relationshipService, invitationService, userStatusService)
	time.Sleep(100 * time.Millisecond)

	// 8. 测试邀请路径
	fmt.Println("\n10. 测试邀请路径...")
	testInvitationPath(db, relationshipService, invitationService, userStatusService)

	// 清理数据（可选）
	if os.Getenv("CLEANUP") == "true" {
		fmt.Println("\n\n清理测试数据...")
		cleanupTestData(db)
	}

	fmt.Println("\n=== 邀请系统测试完成 ===")
}

// testUserRegistration 测试用户注册
func testUserRegistration(db *gorm.DB, authService auth.AuthService, userStatusService user.UserStatusService) {
	ctx := context.Background()

	// 注册邀请者
	inviter, err := authService.Register("inviter1", "inviter1@example.com", "password123", "", "inviter")
	if err != nil {
		log.Printf("注册邀请者失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 注册邀请者成功: ID=%d, 用户名=%s\n", inviter.ID, inviter.Username)

	// 激活用户
	if err := userStatusService.ActivateUser(inviter.ID, true); err != nil {
		log.Printf("激活邀请者失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 激活邀请者成功\n")

	// 注册被邀请者
	invitee, err := authService.Register("invitee1", "invitee1@example.com", "password123", "", "invitee")
	if err != nil {
		log.Printf("注册被邀请者失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 注册被邀请者成功: ID=%d, 用户名=%s\n", invitee.ID, invitee.Username)

	_ = ctx
}

// testInviteCodeGeneration 测试邀请码生成和验证
func testInviteCodeGeneration(db *gorm.DB, invitationService *invitation.InvitationService, userStatusService user.UserStatusService) {
	// 获取第一个用户作为邀请者
	var inviter model.User
	if err := db.Raw("SELECT * FROM users WHERE username = ?", "inviter1").Scan(&inviter).Error; err != nil {
		log.Printf("查询邀请者失败: %v", err)
		return
	}

	// 生成邀请码
	inviteCode, expiresAt, err := invitationService.GenerateInviteCode(inviter.ID, 72) // 72小时过期
	if err != nil {
		log.Printf("生成邀请码失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 生成邀请码成功: %s, 过期时间: %v\n", inviteCode, expiresAt)

	// 创建邀请记录
	invitation, err := invitationService.CreateInvitation(inviter.ID, inviteCode, expiresAt)
	if err != nil {
		log.Printf("创建邀请记录失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 创建邀请记录成功: ID=%d, 状态=%s\n", invitation.ID, invitation.Status)

	// 验证邀请码
	inviterInfo, _, err := invitationService.ValidateInviteCode(inviteCode)
	if err != nil {
		log.Printf("验证邀请码失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 验证邀请码成功: 邀请者=%s\n", inviterInfo.Username)

	// 测试过期邀请码
	expiredCode, _, _ := invitationService.GenerateInviteCode(inviter.ID, 0)
	expiredInvitation, _ := invitationService.CreateInvitation(inviter.ID, expiredCode, time.Now().Add(-time.Hour))
	_, _, err = invitationService.ValidateInviteCode(expiredCode)
	if err != nil {
		fmt.Printf("  ✓ 过期邀请码验证失败（预期）: %v\n", err)
	}
	_ = expiredInvitation
}

// testInvitationRelationships 测试邀请关系追踪
func testInvitationRelationships(db *gorm.DB, relationshipService *invitation.RelationshipService, invitationService *invitation.InvitationService, userStatusService user.UserStatusService) {
	// 获取邀请者和被邀请者
	var inviter, invitee model.User
	if err := db.Raw("SELECT * FROM users WHERE username = ?", "inviter1").Scan(&inviter).Error; err != nil {
		log.Printf("查询邀请者失败: %v", err)
		return
	}
	if err := db.Raw("SELECT * FROM users WHERE username = ?", "invitee1").Scan(&invitee).Error; err != nil {
		log.Printf("查询被邀请者失败: %v", err)
		return
	}

	// 完成邀请
	inviteCode, _, _ := invitationService.GenerateInviteCode(inviter.ID, 72)
	if err := invitationService.CompleteInvitation(inviteCode, invitee.ID, 100); err != nil {
		log.Printf("完成邀请失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 完成邀请关系\n")

	// 获取邀请列表
	invitations, total, err := invitationService.GetInvitationsByInviter(inviter.ID, "", 1, 10)
	if err != nil {
		log.Printf("获取邀请列表失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取邀请列表: 总数=%d, 第一页=%d条\n", total, len(invitations))

	// 获取用户直接邀请的用户
	invitedUsers, total, err := relationshipService.GetUserInvitedUsers(inviter.ID, 1, 10)
	if err != nil {
		log.Printf("获取被邀请用户失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取被邀请用户: 总数=%d, 第一页=%d条\n", total, len(invitedUsers))

	// 获取网络统计
	networkStats, err := relationshipService.GetNetworkStats(inviter.ID)
	if err != nil {
		log.Printf("获取网络统计失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取网络统计: %+v\n", networkStats)

	// 获取各层级邀请数量
	counts, err := relationshipService.GetInvitationCountByLevel(inviter.ID, 3)
	if err != nil {
		log.Printf("获取层级邀请数量失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取层级邀请数量: %+v\n", counts)
}

// testInvitationRewards 测试邀请奖励机制
func testInvitationRewards(db *gorm.DB, rewardService *invitation.RewardService, invitationService *invitation.InvitationService, userStatusService user.UserStatusService) {
	// 创建更多测试用户
	inviter, _ := getUserByUsername(db, "inviter1")

	// 注册更多被邀请者
	authService := auth.NewAuthService(db)
	invitee2, _ := authService.Register("invitee2", "invitee2@example.com", "password123", "", "invitee")
	invitee3, _ := authService.Register("invitee3", "invitee3@example.com", "password123", "", "invitee")

	userStatusService.ActivateUser(invitee2.ID, true)
	userStatusService.ActivateUser(invitee3.ID, true)

	// 生成邀请码并完成邀请
	inviteCode2, _, _ := invitationService.GenerateInviteCode(inviter.ID, 72)
	invitationService.CreateInvitation(inviter.ID, inviteCode2, time.Now().Add(72*time.Hour))
	invitationService.CompleteInvitation(inviteCode2, invitee2.ID, 150)

	inviteCode3, _, _ := invitationService.GenerateInviteCode(inviter.ID, 72)
	invitationService.CreateInvitation(inviter.ID, inviteCode3, time.Now().Add(72*time.Hour))
	invitationService.CompleteInvitation(inviteCode3, invitee3.ID, 200)

	fmt.Printf("  ✓ 完成多用户邀请奖励\n")

	// 获取奖励历史
	rewardHistory, total, err := rewardService.GetRewardHistory(inviter.ID, 1, 10)
	if err != nil {
		log.Printf("获取奖励历史失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取奖励历史: 总数=%d, 第一页=%d条\n", total, len(rewardHistory))

	// 获取奖励统计
	rewardStats, err := rewardService.GetRewardStats(inviter.ID)
	if err != nil {
		log.Printf("获取奖励统计失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取奖励统计: %+v\n", rewardStats)

	// 获取多层级奖励
	multiLevelRewards, err := rewardService.GetMultiLevelRewards(inviter.ID)
	if err != nil {
		log.Printf("获取多层级奖励失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取多层级奖励: %+v\n", multiLevelRewards)

	// 应用奖励
	rewardRecord, err := rewardService.ApplyReward(inviter.ID, invitee2.ID, 2, 0) // 默认规则
	if err != nil {
		log.Printf("应用奖励失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 应用奖励成功: 奖励记录ID=%d, 积分=%d\n", rewardRecord.ID, rewardRecord.Points)

	// 检查奖励资格
	defaultRule := rewardService.GetDefaultRewardRule()
	isEligible, err := rewardService.CheckRewardEligibility(inviter.ID, defaultRule)
	if err != nil {
		log.Printf("检查奖励资格失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 检查奖励资格: 有资格=%v\n", isEligible)
}

// testInvitationLeaderboard 测试邀请排行榜
func testInvitationLeaderboard(db *gorm.DB, leaderboardService *invitation.LeaderboardService) {
	// 获取本月邀请排行榜
	topInviters, err := leaderboardService.GetTopInvitersThisMonth(10)
	if err != nil {
		log.Printf("获取本月邀请排行榜失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取本月邀请排行榜: %d条记录\n", len(topInviters))
	for i, entry := range topInviters {
		fmt.Printf("    排名%d: %s, 邀请数=%d\n", i+1, entry.User.Username, entry.InviteCount)
	}

	// 获取积分排行榜
	topPointsEarners, err := leaderboardService.GetTopPointsEarners(invitation.PeriodAll, 10)
	if err != nil {
		log.Printf("获取积分排行榜失败: %v", err)
		// 忽略错误，继续测试
	} else {
		fmt.Printf("  ✓ 获取积分排行榜: %d条记录\n", len(topPointsEarners))
	}

	// 获取用户排名
	var user model.User
	if err := db.Raw("SELECT * FROM users WHERE username = ?", "inviter1").Scan(&user).Error; err != nil {
		log.Printf("查询用户失败: %v", err)
		return
	}

	userRank, err := leaderboardService.GetUserRank(user.ID, invitation.PeriodAll, invitation.TypeInviteCount)
	if err != nil {
		log.Printf("获取用户排名失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取用户排名: 用户=%s, 排名=%d, 邀请数=%d\n",
		userRank.Username, userRank.Rank, userRank.InviteCount)

	// 获取周排行榜
	weeklyTop, err := leaderboardService.GetWeeklyTopInviters(time.Now().Year(), 0, 10)
	if err != nil {
		log.Printf("获取周排行榜失败: %v", err)
		// 忽略错误，继续测试
	} else {
		fmt.Printf("  ✓ 获取周排行榜: %d条记录\n", len(weeklyTop))
	}
}

// testInvitationStats 测试邀请统计
func testInvitationStats(db *gorm.DB, invitationService *invitation.InvitationService, relationshipService *invitation.RelationshipService, rewardService *invitation.RewardService) {
	var inviter model.User
	if err := db.Raw("SELECT * FROM users WHERE username = ?", "inviter1").Scan(&inviter).Error; err != nil {
		log.Printf("查询邀请者失败: %v", err)
		return
	}

	// 获取邀请统计
	stats, err := invitationService.GetInvitationStats(inviter.ID)
	if err != nil {
		log.Printf("获取邀请统计失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取邀请统计: %+v\n", stats)

	// 获取奖励统计
	rewardStats, err := rewardService.GetRewardStats(inviter.ID)
	if err != nil {
		log.Printf("获取奖励统计失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取奖励统计: %+v\n", rewardStats)

	// 获取网络统计
	networkStats, err := relationshipService.GetNetworkStats(inviter.ID)
	if err != nil {
		log.Printf("获取网络统计失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取网络统计: %+v\n", networkStats)
}

// testInvitationTree 测试邀请树结构
func testInvitationTree(db *gorm.DB, relationshipService *invitation.RelationshipService, invitationService *invitation.InvitationService, userStatusService user.UserStatusService) {
	// 创建多层级邀请关系
	inviter, _ := getUserByUsername(db, "inviter1")
	invitee1, _ := getUserByUsername(db, "invitee1")

	// 注册第二层用户
	authService := auth.NewAuthService(db)
	invitee2, _ := authService.Register("invitee2_child", "invitee2_child@example.com", "password123", "", "invitee")
	invitee3, _ := authService.Register("invitee3_child", "invitee3_child@example.com", "password123", "", "invitee")

	userStatusService.ActivateUser(invitee2.ID, true)
	userStatusService.ActivateUser(invitee3.ID, true)

	// 建立邀请关系
	inviteCode2, _, _ := invitationService.GenerateInviteCode(invitee1.ID, 72)
	invitationService.CreateInvitation(invitee1.ID, inviteCode2, time.Now().Add(72*time.Hour))
	invitationService.CompleteInvitation(inviteCode2, invitee2.ID, 50)

	inviteCode3, _, _ := invitationService.GenerateInviteCode(invitee1.ID, 72)
	invitationService.CreateInvitation(invitee1.ID, inviteCode3, time.Now().Add(72*time.Hour))
	invitationService.CompleteInvitation(inviteCode3, invitee3.ID, 50)

	fmt.Printf("  ✓ 创建多层级邀请关系\n")

	// 获取邀请树
	tree, err := relationshipService.GetUserInvitationTree(inviter.ID, 3)
	if err != nil {
		log.Printf("获取邀请树失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取邀请树成功: 根用户=%s\n", tree.User.Username)
	if len(tree.Children) > 0 {
		fmt.Printf("    第一层用户数: %d\n", len(tree.Children))
		if len(tree.Children[0].Children) > 0 {
			fmt.Printf("    第二层用户数: %d\n", len(tree.Children[0].Children))
		}
	}
}

// testInvitationPath 测试邀请路径
func testInvitationPath(db *gorm.DB, relationshipService *invitation.RelationshipService, invitationService *invitation.InvitationService, userStatusService user.UserStatusService) {
	// 获取最深层用户
	invitee2, _ := getUserByUsername(db, "invitee2_child")

	if invitee2.InvitedByID == nil {
		fmt.Printf("  ! 用户未被邀请，跳过路径测试\n")
		return
	}

	// 获取邀请路径
	path, err := relationshipService.GetUserInvitationPath(invitee2.ID)
	if err != nil {
		log.Printf("获取邀请路径失败: %v", err)
		return
	}
	fmt.Printf("  ✓ 获取邀请路径成功: 路径长度=%d\n", len(path))
	for i, hop := range path {
		fmt.Printf("    第%d跳: %s -> %s\n", i+1, hop.Inviter.Username, hop.Invitee.Username)
	}
}

// getUserByUsername 根据用户名获取用户
func getUserByUsername(db *gorm.DB, username string) (*model.User, error) {
	var user model.User
	if err := db.Raw("SELECT * FROM users WHERE username = ?", username).Scan(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// cleanupTestData 清理测试数据
func cleanupTestData(db *gorm.DB) {
	// 删除测试数据
	if err := db.Exec("DELETE FROM invitations WHERE inviter_id IN (SELECT id FROM users WHERE username LIKE 'inviter%' OR username LIKE 'invitee%' OR username LIKE '%_child')").Error; err != nil {
		log.Printf("清理邀请记录失败: %v", err)
	}
	if err := db.Exec("DELETE FROM point_records WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'inviter%' OR username LIKE 'invitee%' OR username LIKE '%_child')").Error; err != nil {
		log.Printf("清理积分记录失败: %v", err)
	}
	if err := db.Exec("DELETE FROM users WHERE username LIKE 'inviter%' OR username LIKE 'invitee%' OR username LIKE '%_child'").Error; err != nil {
		log.Printf("清理用户失败: %v", err)
	}
	fmt.Println("  ✓ 清理测试数据完成")
}
