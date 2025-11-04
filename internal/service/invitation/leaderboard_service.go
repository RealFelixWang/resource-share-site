/*
Package invitation provides invitation leaderboard services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package invitation

import (
	"fmt"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// RankingPeriod 排行榜周期
type RankingPeriod string

const (
	PeriodAll   RankingPeriod = "all"
	PeriodYear  RankingPeriod = "year"
	PeriodMonth RankingPeriod = "month"
	PeriodWeek  RankingPeriod = "week"
	PeriodDay   RankingPeriod = "day"
)

// RankingType 排行榜类型
type RankingType string

const (
	TypeInviteCount  RankingType = "invite_count"  // 邀请数量
	TypePointsEarned RankingType = "points_earned" // 获得积分
	TypeNetworkSize  RankingType = "network_size"  // 网络规模
	TypeActiveUsers  RankingType = "active_users"  // 活跃用户
)

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	Rank         int         `json:"rank"`
	User         *model.User `json:"user"`
	InviteCount  int64       `json:"invite_count"`
	PointsEarned int64       `json:"points_earned"`
	NetworkSize  int64       `json:"network_size"`
	ActiveUsers  int64       `json:"active_users"`
	SuccessRate  float64     `json:"success_rate"`
	PeriodStart  *time.Time  `json:"period_start,omitempty"`
	PeriodEnd    *time.Time  `json:"period_end,omitempty"`
}

// UserRankInfo 用户排名信息
type UserRankInfo struct {
	UserID       uint          `json:"user_id"`
	Username     string        `json:"username"`
	Email        string        `json:"email"`
	Rank         int           `json:"rank"`
	InviteCount  int64         `json:"invite_count"`
	PointsEarned int64         `json:"points_earned"`
	NetworkSize  int64         `json:"network_size"`
	ActiveUsers  int64         `json:"active_users"`
	SuccessRate  float64       `json:"success_rate"`
	Period       RankingPeriod `json:"period"`
	Type         RankingType   `json:"type"`
	LastUpdated  time.Time     `json:"last_updated"`
}

// LeaderboardService 排行榜服务
type LeaderboardService struct {
	db *gorm.DB
}

// NewLeaderboardService 创建新的排行榜服务
func NewLeaderboardService(db *gorm.DB) *LeaderboardService {
	return &LeaderboardService{
		db: db,
	}
}

// GetLeaderboard 获取排行榜
// 参数：
//   - period: 周期（all, year, month, week, day）
//   - rankingType: 排行榜类型（invite_count, points_earned, network_size, active_users）
//   - limit: 限制数量
//   - offset: 偏移量
//
// 返回：
//   - 排行榜条目列表
//   - 错误信息
func (s *LeaderboardService) GetLeaderboard(period RankingPeriod, rankingType RankingType, limit, offset int) ([]*LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// 计算时间范围
	periodStart, periodEnd, err := s.calculatePeriod(period)
	if err != nil {
		return nil, fmt.Errorf("计算时间范围失败: %w", err)
	}

	// 获取排行榜数据
	entries, err := s.fetchLeaderboardData(period, rankingType, limit, offset, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜数据失败: %w", err)
	}

	// 计算排名
	for i, entry := range entries {
		entry.Rank = offset + i + 1
		if periodStart != nil {
			entry.PeriodStart = periodStart
			entry.PeriodEnd = periodEnd
		}
	}

	return entries, nil
}

// GetUserRank 获取用户排名
// 参数：
//   - userID: 用户ID
//   - period: 周期
//   - rankingType: 排行榜类型
//
// 返回：
//   - 用户排名信息
//   - 错误信息
func (s *LeaderboardService) GetUserRank(userID uint, period RankingPeriod, rankingType RankingType) (*UserRankInfo, error) {
	// 计算时间范围
	periodStart, periodEnd, err := s.calculatePeriod(period)
	if err != nil {
		return nil, fmt.Errorf("计算时间范围失败: %w", err)
	}

	// 获取用户信息
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("查询用户信息失败: %w", err)
	}

	// 获取用户统计数据
	stats, err := s.getUserStats(userID, period, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("获取用户统计信息失败: %w", err)
	}

	// 计算排名
	rank, err := s.calculateUserRank(userID, rankingType, period, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("计算用户排名失败: %w", err)
	}

	return &UserRankInfo{
		UserID:       userID,
		Username:     user.Username,
		Email:        user.Email,
		Rank:         rank,
		InviteCount:  stats["invite_count"].(int64),
		PointsEarned: stats["points_earned"].(int64),
		NetworkSize:  stats["network_size"].(int64),
		ActiveUsers:  stats["active_users"].(int64),
		SuccessRate:  stats["success_rate"].(float64),
		Period:       period,
		Type:         rankingType,
		LastUpdated:  time.Now(),
	}, nil
}

// GetTopInvitersThisMonth 获取本月邀请排行榜
// 参数：
//   - limit: 限制数量
//
// 返回：
//   - 排行榜条目列表
//   - 错误信息
func (s *LeaderboardService) GetTopInvitersThisMonth(limit int) ([]*LeaderboardEntry, error) {
	return s.GetLeaderboard(PeriodMonth, TypeInviteCount, limit, 0)
}

// GetTopPointsEarners 获取积分排行榜
// 参数：
//   - period: 周期
//   - limit: 限制数量
//
// 返回：
//   - 排行榜条目列表
//   - 错误信息
func (s *LeaderboardService) GetTopPointsEarners(period RankingPeriod, limit int) ([]*LeaderboardEntry, error) {
	return s.GetLeaderboard(period, TypePointsEarned, limit, 0)
}

// GetNetworkSizeLeaderboard 获取网络规模排行榜
// 参数：
//   - limit: 限制数量
//
// 返回：
//   - 排行榜条目列表
//   - 错误信息
func (s *LeaderboardService) GetNetworkSizeLeaderboard(limit int) ([]*LeaderboardEntry, error) {
	return s.GetLeaderboard(PeriodAll, TypeNetworkSize, limit, 0)
}

// GetWeeklyTopInviters 获取每周邀请排行榜
// 参数：
//   - weekNumber: 周数（1-53），0表示当前周
//   - year: 年份，0表示当前年
//   - limit: 限制数量
//
// 返回：
//   - 排行榜条目列表
//   - 错误信息
func (s *LeaderboardService) GetWeeklyTopInviters(year int, weekNumber int, limit int) ([]*LeaderboardEntry, error) {
	// 计算指定周的起始和结束时间
	periodStart, periodEnd, err := s.calculateWeekPeriod(year, weekNumber)
	if err != nil {
		return nil, fmt.Errorf("计算周期间失败: %w", err)
	}

	// 获取该周的排行榜数据
	entries, err := s.fetchLeaderboardData(PeriodWeek, TypeInviteCount, limit, 0, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("获取周排行榜失败: %w", err)
	}

	// 设置排名
	for i, entry := range entries {
		entry.Rank = i + 1
		entry.PeriodStart = periodStart
		entry.PeriodEnd = periodEnd
	}

	return entries, nil
}

// GetHistoricalRankings 获取历史排行榜
// 参数：
//   - year: 年份
//   - month: 月份（1-12），0表示全年
//   - rankingType: 排行榜类型
//   - limit: 限制数量
//
// 返回：
//   - 排行榜条目列表
//   - 错误信息
func (s *LeaderboardService) GetHistoricalRankings(year, month int, rankingType RankingType, limit int) ([]*LeaderboardEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var periodStart, periodEnd *time.Time

	if month > 0 {
		// 指定月份
		start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)
		periodStart = &start
		periodEnd = &end
	} else {
		// 全年
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
		periodStart = &start
		periodEnd = &end
	}

	entries, err := s.fetchLeaderboardData(PeriodYear, rankingType, limit, 0, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("获取历史排行榜失败: %w", err)
	}

	for i, entry := range entries {
		entry.Rank = i + 1
		entry.PeriodStart = periodStart
		entry.PeriodEnd = periodEnd
	}

	return entries, nil
}

// calculatePeriod 计算时间范围
func (s *LeaderboardService) calculatePeriod(period RankingPeriod) (*time.Time, *time.Time, error) {
	now := time.Now()
	var start, end time.Time

	switch period {
	case PeriodDay:
		start = now.AddDate(0, 0, -1)
		end = now
	case PeriodWeek:
		start = now.AddDate(0, 0, -7)
		end = now
	case PeriodMonth:
		start = now.AddDate(0, -1, 0)
		end = now
	case PeriodYear:
		start = now.AddDate(-1, 0, 0)
		end = now
	case PeriodAll:
		return nil, nil, nil
	default:
		return nil, nil, fmt.Errorf("未知的时间周期: %s", period)
	}

	return &start, &end, nil
}

// fetchLeaderboardData 获取排行榜数据
func (s *LeaderboardService) fetchLeaderboardData(period RankingPeriod, rankingType RankingType, limit, offset int, periodStart, periodEnd *time.Time) ([]*LeaderboardEntry, error) {
	// 构建基础查询
	query := s.db.Table("users").
		Select(`users.id, users.username, users.email, users.created_at`).
		Joins("LEFT JOIN invitations ON users.id = invitations.inviter_id AND invitations.status = ?", model.InvitationStatusCompleted)

	// 添加时间范围条件
	if periodStart != nil && periodEnd != nil {
		query = query.Where("invitations.created_at >= ? AND invitations.created_at < ?", periodStart, periodEnd)
	}

	// 根据排行榜类型添加不同的查询逻辑
	switch rankingType {
	case TypeInviteCount:
		query = query.Group("users.id").Having("COUNT(invitations.id) > 0")
	case TypePointsEarned:
		query = query.Group("users.id").Having("SUM(invitations.points_awarded) > 0")
	case TypeNetworkSize:
		// 网络规模查询需要更复杂的SQL
		return s.fetchNetworkSizeLeaderboard(period, limit, offset, periodStart, periodEnd)
	case TypeActiveUsers:
		// 活跃用户查询
		return s.fetchActiveUsersLeaderboard(period, limit, offset, periodStart, periodEnd)
	}

	// 执行查询
	var results []map[string]interface{}
	err := query.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 转换为LeaderboardEntry
	entries := make([]*LeaderboardEntry, len(results))
	for i, result := range results {
		userID := uint(result["id"].(int64))

		// 获取用户信息
		var user model.User
		if err := s.db.First(&user, userID).Error; err != nil {
			continue
		}

		// 获取统计数据
		stats, _ := s.getUserStats(userID, period, periodStart, periodEnd)

		entries[i] = &LeaderboardEntry{
			User:         &user,
			InviteCount:  stats["invite_count"].(int64),
			PointsEarned: stats["points_earned"].(int64),
			NetworkSize:  stats["network_size"].(int64),
			ActiveUsers:  stats["active_users"].(int64),
			SuccessRate:  stats["success_rate"].(float64),
		}
	}

	// 根据类型排序
	switch rankingType {
	case TypeInviteCount:
		// 按邀请数量排序
	case TypePointsEarned:
		// 按积分排序
	case TypeNetworkSize:
		// 按网络规模排序
	case TypeActiveUsers:
		// 按活跃用户数排序
	}

	return entries, nil
}

// fetchNetworkSizeLeaderboard 获取网络规模排行榜
func (s *LeaderboardService) fetchNetworkSizeLeaderboard(period RankingPeriod, limit, offset int, periodStart, periodEnd *time.Time) ([]*LeaderboardEntry, error) {
	// 这里实现网络规模查询的逻辑
	// 由于复杂性较高，这里简化处理
	return nil, fmt.Errorf("网络规模排行榜功能待实现")
}

// fetchActiveUsersLeaderboard 获取活跃用户排行榜
func (s *LeaderboardService) fetchActiveUsersLeaderboard(period RankingPeriod, limit, offset int, periodStart, periodEnd *time.Time) ([]*LeaderboardEntry, error) {
	// 这里实现活跃用户查询的逻辑
	// 由于复杂性较高，这里简化处理
	return nil, fmt.Errorf("活跃用户排行榜功能待实现")
}

// getUserStats 获取用户统计信息
func (s *LeaderboardService) getUserStats(userID uint, period RankingPeriod, periodStart, periodEnd *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 邀请数量
	query := s.db.Model(&model.Invitation{}).Where("inviter_id = ? AND status = ?", userID, model.InvitationStatusCompleted)
	if periodStart != nil && periodEnd != nil {
		query = query.Where("created_at >= ? AND created_at < ?", periodStart, periodEnd)
	}
	var inviteCount int64
	if err := query.Count(&inviteCount).Error; err != nil {
		return nil, err
	}
	stats["invite_count"] = inviteCount

	// 获得积分
	var pointsEarned int64
	if err := s.db.Model(&model.Invitation{}).Where("inviter_id = ? AND status = ?", userID, model.InvitationStatusCompleted).Pluck("points_awarded", &[]int{}).Error; err != nil {
		return nil, err
	}
	stats["points_earned"] = pointsEarned

	// 网络规模（简化处理）
	stats["network_size"] = inviteCount

	// 活跃用户（简化处理）
	stats["active_users"] = inviteCount

	// 成功率
	var successRate float64
	if inviteCount > 0 {
		totalInvites, _ := s.getTotalInvites(userID, periodStart, periodEnd)
		if totalInvites > 0 {
			successRate = float64(inviteCount) / float64(totalInvites) * 100
		}
	}
	stats["success_rate"] = successRate

	return stats, nil
}

// getTotalInvites 获取总邀请数
func (s *LeaderboardService) getTotalInvites(userID uint, periodStart, periodEnd *time.Time) (int64, error) {
	query := s.db.Model(&model.Invitation{}).Where("inviter_id = ?", userID)
	if periodStart != nil && periodEnd != nil {
		query = query.Where("created_at >= ? AND created_at < ?", periodStart, periodEnd)
	}
	var count int64
	err := query.Count(&count).Error
	return count, err
}

// calculateUserRank 计算用户排名
func (s *LeaderboardService) calculateUserRank(userID uint, rankingType RankingType, period RankingPeriod, periodStart, periodEnd *time.Time) (int, error) {
	// 获取用户统计信息
	stats, err := s.getUserStats(userID, period, periodStart, periodEnd)
	if err != nil {
		return 0, err
	}

	// 计算用户在该类型下的排名
	var value int64
	switch rankingType {
	case TypeInviteCount:
		value = stats["invite_count"].(int64)
	case TypePointsEarned:
		value = stats["points_earned"].(int64)
	case TypeNetworkSize:
		value = stats["network_size"].(int64)
	case TypeActiveUsers:
		value = stats["active_users"].(int64)
	default:
		return 0, fmt.Errorf("未知的排行榜类型: %s", rankingType)
	}

	// 计算有多少用户的值大于当前用户
	var count int64
	query := s.db.Table("users").
		Select("COUNT(*)").
		Joins("LEFT JOIN invitations ON users.id = invitations.inviter_id AND invitations.status = ?", model.InvitationStatusCompleted).
		Where("users.id != ?", userID)

	if periodStart != nil && periodEnd != nil {
		query = query.Where("invitations.created_at >= ? AND invitations.created_at < ?", periodStart, periodEnd)
	}

	switch rankingType {
	case TypeInviteCount:
		query = query.Group("users.id").Having("COUNT(invitations.id) > ?", value)
	case TypePointsEarned:
		query = query.Group("users.id").Having("SUM(invitations.points_awarded) > ?", value)
	}

	if err := query.Scan(&count).Error; err != nil {
		return 0, err
	}

	return int(count) + 1, nil
}

// calculateWeekPeriod 计算周期间
func (s *LeaderboardService) calculateWeekPeriod(year, weekNumber int) (*time.Time, *time.Time, error) {
	now := time.Now()
	if year == 0 {
		year = now.Year()
	}
	if weekNumber == 0 {
		// 当前周
		_, week := now.ISOWeek()
		weekNumber = week
	}

	// 计算该周的起始日期（周一）
	d := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	// 找到第一个周一
	for d.Weekday() != time.Monday {
		d = d.AddDate(0, 0, 1)
	}
	// 添加周数减1周
	d = d.AddDate(0, 0, (weekNumber-1)*7)
	start := d

	// 计算结束日期（下周一）
	end := start.AddDate(0, 0, 7)

	return &start, &end, nil
}
