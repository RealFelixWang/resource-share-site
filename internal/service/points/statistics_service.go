/*
Points Statistics Service - 积分统计分析服务

提供完整的积分系统统计和分析功能，包括：
- 用户积分统计
- 系统积分统计
- 趋势分析
- 排行榜
- 报表生成

Author: Felix Wang
Email: felixwang.biz@gmail.com
Date: 2025-10-31
*/

package points

import (
	"fmt"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// StatisticsService 积分统计分析服务
type StatisticsService struct {
	db *gorm.DB
}

// NewStatisticsService 创建新的积分统计分析服务
func NewStatisticsService(db *gorm.DB) *StatisticsService {
	return &StatisticsService{
		db: db,
	}
}

// GetUserPointsSummary 获取用户积分概览
func (s *StatisticsService) GetUserPointsSummary(userID uint) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// 当前积分余额
	balance, err := NewEarningService(s.db).GetUserPointsBalance(userID)
	if err != nil {
		return nil, fmt.Errorf("获取积分余额失败: %w", err)
	}
	summary["current_balance"] = balance

	// 总收入
	var totalIncome int64
	rows, _ := s.db.Raw("SELECT COALESCE(SUM(points), 0) FROM point_records WHERE user_id = ? AND type = ?", userID, model.PointTypeIncome).Rows()
	defer rows.Close()
	if rows.Next() {
		var sum *int64
		rows.Scan(&sum)
		if sum != nil {
			totalIncome = *sum
		}
	}
	summary["total_income"] = totalIncome

	// 总支出
	var totalExpense int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(ABS(points)), 0) FROM point_records WHERE user_id = ? AND type = ?", userID, model.PointTypeExpense).Rows()
	defer rows.Close()
	if rows.Next() {
		var sum *int64
		rows.Scan(&sum)
		if sum != nil {
			totalExpense = *sum
		}
	}
	summary["total_expense"] = totalExpense

	// 今日收入
	today := time.Now().Format("2006-01-02")
	var todayIncome int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(points), 0) FROM point_records WHERE user_id = ? AND type = ? AND DATE(created_at) = ?", userID, model.PointTypeIncome, today).Rows()
	defer rows.Close()
	if rows.Next() {
		var sum *int64
		rows.Scan(&sum)
		if sum != nil {
			todayIncome = *sum
		}
	}
	summary["today_income"] = todayIncome

	// 今日支出
	var todayExpense int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(ABS(points)), 0) FROM point_records WHERE user_id = ? AND type = ? AND DATE(created_at) = ?", userID, model.PointTypeExpense, today).Rows()
	defer rows.Close()
	if rows.Next() {
		var sum *int64
		rows.Scan(&sum)
		if sum != nil {
			todayExpense = *sum
		}
	}
	summary["today_expense"] = todayExpense

	// 本月收入
	month := time.Now().Format("2006-01")
	var monthIncome int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(points), 0) FROM point_records WHERE user_id = ? AND type = ? AND strftime('%Y-%m', created_at) = ?", userID, model.PointTypeIncome, month).Rows()
	defer rows.Close()
	if rows.Next() {
		var sum *int64
		rows.Scan(&sum)
		if sum != nil {
			monthIncome = *sum
		}
	}
	summary["month_income"] = monthIncome

	// 本月支出
	var monthExpense int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(ABS(points)), 0) FROM point_records WHERE user_id = ? AND type = ? AND strftime('%Y-%m', created_at) = ?", userID, model.PointTypeExpense, month).Rows()
	defer rows.Close()
	if rows.Next() {
		var sum *int64
		rows.Scan(&sum)
		if sum != nil {
			monthExpense = *sum
		}
	}
	summary["month_expense"] = monthExpense

	// 积分来源分布
	var sourceDistribution []struct {
		Source string
		Total  int64
		Count  int64
	}
	s.db.Raw(`
		SELECT
			source,
			SUM(ABS(points)) as total,
			COUNT(*) as count
		FROM point_records
		WHERE user_id = ? AND type = ?
		GROUP BY source
		ORDER BY total DESC
	`, userID, model.PointTypeIncome).Scan(&sourceDistribution)
	summary["income_sources"] = sourceDistribution

	// 积分消费分布
	var expenseDistribution []struct {
		Source string
		Total  int64
		Count  int64
	}
	s.db.Raw(`
		SELECT
			source,
			SUM(ABS(points)) as total,
			COUNT(*) as count
		FROM point_records
		WHERE user_id = ? AND type = ?
		GROUP BY source
		ORDER BY total DESC
	`, userID, model.PointTypeExpense).Scan(&expenseDistribution)
	summary["expense_sources"] = expenseDistribution

	// 连续签到天数
	consecutiveCheckins, err := s.getConsecutiveCheckins(userID)
	if err == nil {
		summary["consecutive_checkins"] = consecutiveCheckins
	}

	return summary, nil
}

// GetUserPointsTrend 获取用户积分趋势
func (s *StatisticsService) GetUserPointsTrend(userID uint, days int) ([]map[string]interface{}, error) {
	if days <= 0 || days > 365 {
		return nil, fmt.Errorf("查询天数必须在1-365之间")
	}

	var trends []map[string]interface{}

	// 查询最近N天的每日积分变动
	rows, err := s.db.Raw(`
		SELECT
			DATE(created_at) as date,
			SUM(CASE WHEN type = 'income' THEN points ELSE 0 END) as daily_income,
			SUM(CASE WHEN type = 'expense' THEN ABS(points) ELSE 0 END) as daily_expense,
			SUM(points) as daily_net
		FROM point_records
		WHERE user_id = ? AND created_at >= DATE('now', '-%d days')
		GROUP BY DATE(created_at)
		ORDER BY date
	`, userID, days).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询积分趋势失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var date string
		var dailyIncome, dailyExpense, dailyNet int64
		rows.Scan(&date, &dailyIncome, &dailyExpense, &dailyNet)

		trends = append(trends, map[string]interface{}{
			"date":          date,
			"daily_income":  dailyIncome,
			"daily_expense": dailyExpense,
			"daily_net":     dailyNet,
		})
	}

	return trends, nil
}

// GetUserPointsRanking 获取用户积分排行榜
func (s *StatisticsService) GetUserPointsRanking(limit int) ([]struct {
	UserID   uint
	Username string
	Balance  int
	Rank     int
}, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("排行榜限制必须在1-100之间")
	}

	var rankings []struct {
		UserID   uint
		Username string
		Balance  int
		Rank     int
	}

	rows, err := s.db.Raw(`
		SELECT
			id as user_id,
			username,
			points_balance as balance,
			RANK() OVER (ORDER BY points_balance DESC) as rank
		FROM users
		WHERE status = 'active'
		ORDER BY points_balance DESC
		LIMIT ?
	`, limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("获取排行榜失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ranking struct {
			UserID   uint
			Username string
			Balance  int
			Rank     int
		}
		rows.Scan(&ranking.UserID, &ranking.Username, &ranking.Balance, &ranking.Rank)
		rankings = append(rankings, ranking)
	}

	return rankings, nil
}

// GetSystemPointsStats 获取系统积分统计
func (s *StatisticsService) GetSystemPointsStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 系统总积分
	var totalPoints int64
	rows, _ := s.db.Raw("SELECT COALESCE(SUM(points_balance), 0) FROM users").Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&totalPoints)
	}
	stats["total_points"] = totalPoints

	// 活跃用户数（今天有积分变动的用户）
	var activeUsers int64
	today := time.Now().Format("2006-01-02")
	rows, _ = s.db.Raw("SELECT COUNT(DISTINCT user_id) FROM point_records WHERE DATE(created_at) = ?", today).Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&activeUsers)
	}
	stats["active_users_today"] = activeUsers

	// 新增用户数（今天注册的用户）
	var newUsers int64
	rows, _ = s.db.Raw("SELECT COUNT(*) FROM users WHERE DATE(created_at) = ?", today).Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&newUsers)
	}
	stats["new_users_today"] = newUsers

	// 总收入积分
	var totalIncome int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(points), 0) FROM point_records WHERE type = ?", model.PointTypeIncome).Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&totalIncome)
	}
	stats["total_income"] = totalIncome

	// 总支出积分
	var totalExpense int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(ABS(points)), 0) FROM point_records WHERE type = ?", model.PointTypeExpense).Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&totalExpense)
	}
	stats["total_expense"] = totalExpense

	// 今日收入
	var todayIncome int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(points), 0) FROM point_records WHERE type = ? AND DATE(created_at) = ?", model.PointTypeIncome, today).Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&todayIncome)
	}
	stats["today_income"] = todayIncome

	// 今日支出
	var todayExpense int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(ABS(points)), 0) FROM point_records WHERE type = ? AND DATE(created_at) = ?", model.PointTypeExpense, today).Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&todayExpense)
	}
	stats["today_expense"] = todayExpense

	// 积分来源统计
	var sourceStats []struct {
		Source string
		Total  int64
		Count  int64
	}
	s.db.Raw(`
		SELECT
			source,
			SUM(ABS(points)) as total,
			COUNT(*) as count
		FROM point_records
		WHERE type = ?
		GROUP BY source
		ORDER BY total DESC
	`, model.PointTypeIncome).Scan(&sourceStats)
	stats["income_sources"] = sourceStats

	// 积分消费统计
	var expenseStats []struct {
		Source string
		Total  int64
		Count  int64
	}
	s.db.Raw(`
		SELECT
			source,
			SUM(ABS(points)) as total,
			COUNT(*) as count
		FROM point_records
		WHERE type = ?
		GROUP BY source
		ORDER BY total DESC
	`, model.PointTypeExpense).Scan(&expenseStats)
	stats["expense_sources"] = expenseStats

	// 用户积分分布
	var distribution []struct {
		Range string
		Count int64
	}
	s.db.Raw(`
		SELECT
			CASE
				WHEN points_balance = 0 THEN '0'
				WHEN points_balance BETWEEN 1 AND 100 THEN '1-100'
				WHEN points_balance BETWEEN 101 AND 500 THEN '101-500'
				WHEN points_balance BETWEEN 501 AND 1000 THEN '501-1000'
				WHEN points_balance BETWEEN 1001 AND 5000 THEN '1001-5000'
				WHEN points_balance > 5000 THEN '5000+'
				ELSE '其他'
			END as range,
			COUNT(*) as count
		FROM users
		WHERE status = 'active'
		GROUP BY range
		ORDER BY MIN(points_balance)
	`).Scan(&distribution)
	stats["user_distribution"] = distribution

	return stats, nil
}

// GetPointsFlowTrend 获取积分流动趋势
func (s *StatisticsService) GetPointsFlowTrend(days int) ([]map[string]interface{}, error) {
	if days <= 0 || days > 365 {
		return nil, fmt.Errorf("查询天数必须在1-365之间")
	}

	var trends []map[string]interface{}

	// 查询最近N天的每日积分流动
	rows, err := s.db.Raw(`
		SELECT
			DATE(created_at) as date,
			SUM(CASE WHEN type = 'income' THEN points ELSE 0 END) as total_income,
			SUM(CASE WHEN type = 'expense' THEN ABS(points) ELSE 0 END) as total_expense,
			COUNT(DISTINCT user_id) as active_users
		FROM point_records
		WHERE created_at >= DATE('now', '-%d days')
		GROUP BY DATE(created_at)
		ORDER BY date
	`, days).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询积分流动趋势失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var date string
		var totalIncome, totalExpense, activeUsers int64
		rows.Scan(&date, &totalIncome, &totalExpense, &activeUsers)

		trends = append(trends, map[string]interface{}{
			"date":          date,
			"total_income":  totalIncome,
			"total_expense": totalExpense,
			"net_flow":      totalIncome - totalExpense,
			"active_users":  activeUsers,
		})
	}

	return trends, nil
}

// GetTopEarners 获取积分获取排行榜
func (s *StatisticsService) GetTopEarners(limit int, days int) ([]struct {
	UserID           uint
	Username         string
	TotalEarned      int64
	TransactionCount int64
}, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("排行榜限制必须在1-100之间")
	}

	if days <= 0 || days > 365 {
		days = 30 // 默认30天
	}

	var earners []struct {
		UserID           uint
		Username         string
		TotalEarned      int64
		TransactionCount int64
	}

	rows, err := s.db.Raw(`
		SELECT
			pr.user_id,
			u.username,
			SUM(pr.points) as total_earned,
			COUNT(*) as transaction_count
		FROM point_records pr
		JOIN users u ON pr.user_id = u.id
		WHERE pr.type = ? AND pr.created_at >= DATE('now', '-%d days')
		GROUP BY pr.user_id, u.username
		ORDER BY total_earned DESC
		LIMIT ?
	`, model.PointTypeIncome, days, limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("获取积分获取排行榜失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var earner struct {
			UserID           uint
			Username         string
			TotalEarned      int64
			TransactionCount int64
		}
		rows.Scan(&earner.UserID, &earner.Username, &earner.TotalEarned, &earner.TransactionCount)
		earners = append(earners, earner)
	}

	return earners, nil
}

// GetTopSpenders 获取积分消费排行榜
func (s *StatisticsService) GetTopSpenders(limit int, days int) ([]struct {
	UserID           uint
	Username         string
	TotalSpent       int64
	TransactionCount int64
}, error) {
	if limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("排行榜限制必须在1-100之间")
	}

	if days <= 0 || days > 365 {
		days = 30 // 默认30天
	}

	var spenders []struct {
		UserID           uint
		Username         string
		TotalSpent       int64
		TransactionCount int64
	}

	rows, err := s.db.Raw(`
		SELECT
			pr.user_id,
			u.username,
			SUM(ABS(pr.points)) as total_spent,
			COUNT(*) as transaction_count
		FROM point_records pr
		JOIN users u ON pr.user_id = u.id
		WHERE pr.type = ? AND pr.created_at >= DATE('now', '-%d days')
		GROUP BY pr.user_id, u.username
		ORDER BY total_spent DESC
		LIMIT ?
	`, model.PointTypeExpense, days, limit).Rows()
	if err != nil {
		return nil, fmt.Errorf("获取积分消费排行榜失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var spender struct {
			UserID           uint
			Username         string
			TotalSpent       int64
			TransactionCount int64
		}
		rows.Scan(&spender.UserID, &spender.Username, &spender.TotalSpent, &spender.TransactionCount)
		spenders = append(spenders, spender)
	}

	return spenders, nil
}

// getConsecutiveCheckins 获取连续签到天数
func (s *StatisticsService) getConsecutiveCheckins(userID uint) (int, error) {
	var consecutiveDays int

	rows, err := s.db.Raw(`
		WITH RECURSIVE dates(date) AS (
			SELECT DATE('now')
			UNION ALL
			SELECT DATE(date, '-1 day')
			FROM dates
			WHERE date > DATE('now', '-30 days')
		)
		SELECT COUNT(*)
		FROM dates d
		WHERE EXISTS (
			SELECT 1
			FROM point_records
			WHERE user_id = ?
			AND source = ?
			AND DATE(created_at) = d.date
		)
	`, userID, model.PointSourceDailyCheckin).Rows()
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&consecutiveDays)
	}

	return consecutiveDays, nil
}

// ExportUserPointsData 导出用户积分数据
func (s *StatisticsService) ExportUserPointsData(userID uint, startDate, endDate string) ([]model.PointRecord, error) {
	query := s.db.Where("user_id = ?", userID)

	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}

	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	var records []model.PointRecord
	if err := query.Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("导出积分数据失败: %w", err)
	}

	return records, nil
}
