/*
Points Earning Service - 积分获取机制服务

提供用户通过各种活动获取积分的机制，包括：
- 邀请用户奖励
- 资源上传奖励
- 资源下载奖励
- 每日签到奖励
- 管理员手动添加积分

Author: Felix Wang
Email: felixwang.biz@gmail.com
Date: 2025-10-31
*/

package points

import (
	"database/sql"
	"fmt"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EarningService 积分获取服务
type EarningService struct {
	db *gorm.DB
}

// NewEarningService 创建新的积分获取服务
func NewEarningService(db *gorm.DB) *EarningService {
	return &EarningService{
		db: db,
	}
}

// EarnPointsByInvite 用户邀请奖励
func (s *EarningService) EarnPointsByInvite(inviterID, inviteeID uint, points int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取邀请关系
		var invitation model.Invitation
		if err := tx.Where("inviter_id = ? AND invitee_id = ?", inviterID, inviteeID).
			First(&invitation).Error; err != nil {
			return fmt.Errorf("邀请关系不存在: %w", err)
		}

		// 检查是否已经奖励过
		var count int64
		tx.Model(&model.PointRecord{}).
			Where("user_id = ? AND source = ? AND invitation_id = ?",
				inviterID, model.PointSourceInviteReward, invitation.ID).
			Count(&count)
		if count > 0 {
			return fmt.Errorf("该邀请已经奖励过积分")
		}

		// 添加积分
		if err := s.addPoints(tx, inviterID, points, model.PointSourceInviteReward,
			"邀请用户奖励", &invitation.ID, nil); err != nil {
			return err
		}

		// 更新邀请关系状态
		invitation.PointsAwarded = points
		now := time.Now()
		invitation.AwardedAt = &now
		if err := tx.Save(&invitation).Error; err != nil {
			return fmt.Errorf("更新邀请状态失败: %w", err)
		}

		return nil
	})
}

// EarnPointsByResourceUpload 资源上传奖励
func (s *EarningService) EarnPointsByResourceUpload(uploaderID, resourceID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取上传奖励规则
		rule, err := getRuleByKey(tx, model.PointSourceUploadReward)
		if err != nil {
			return err
		}

		if !rule.IsEnabled {
			return fmt.Errorf("资源上传奖励已禁用")
		}

		// 检查是否已经奖励过
		var count int64
		tx.Model(&model.PointRecord{}).
			Where("user_id = ? AND source = ? AND resource_id = ?",
				uploaderID, model.PointSourceUploadReward, resourceID).
			Count(&count)
		if count > 0 {
			return fmt.Errorf("该资源上传已经奖励过积分")
		}

		// 添加积分
		description := fmt.Sprintf("资源上传奖励: %s", rule.RuleName)
		if err := s.addPoints(tx, uploaderID, rule.Points, model.PointSourceUploadReward,
			description, nil, &resourceID); err != nil {
			return err
		}

		return nil
	})
}

// EarnPointsByResourceDownload 资源下载奖励
func (s *EarningService) EarnPointsByResourceDownload(downloaderID, resourceID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取下载奖励规则
		rule, err := getRuleByKey(tx, model.PointSourceResourceDownload)
		if err != nil {
			return err
		}

		if !rule.IsEnabled {
			return fmt.Errorf("资源下载奖励已禁用")
		}

		// 检查是否已经奖励过
		var count int64
		tx.Model(&model.PointRecord{}).
			Where("user_id = ? AND source = ? AND resource_id = ?",
				downloaderID, model.PointSourceResourceDownload, resourceID).
			Count(&count)
		if count > 0 {
			return fmt.Errorf("该资源下载已经奖励过积分")
		}

		// 添加积分
		description := fmt.Sprintf("资源下载奖励: %s", rule.RuleName)
		if err := s.addPoints(tx, downloaderID, rule.Points, model.PointSourceResourceDownload,
			description, nil, &resourceID); err != nil {
			return err
		}

		return nil
	})
}

// EarnPointsByDailyCheckin 每日签到奖励
func (s *EarningService) EarnPointsByDailyCheckin(userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取签到奖励规则
		rule, err := getRuleByKey(tx, model.PointSourceDailyCheckin)
		if err != nil {
			return err
		}

		if !rule.IsEnabled {
			return fmt.Errorf("每日签到奖励已禁用")
		}

		// 检查今天是否已经签到
		today := time.Now().Format("2006-01-02")
		var count int64
		tx.Model(&model.PointRecord{}).
			Where("user_id = ? AND source = ? AND DATE(created_at) = ?",
				userID, model.PointSourceDailyCheckin, today).
			Count(&count)
		if count > 0 {
			return fmt.Errorf("今天已经签到过了")
		}

		// 添加积分
		description := fmt.Sprintf("每日签到奖励: %s", rule.RuleName)
		if err := s.addPoints(tx, userID, rule.Points, model.PointSourceDailyCheckin,
			description, nil, nil); err != nil {
			return err
		}

		return nil
	})
}

// EarnPointsByAdmin 管理员手动添加积分
func (s *EarningService) EarnPointsByAdmin(userID uint, points int, description string, operatedByID *uint) error {
	if points <= 0 {
		return fmt.Errorf("积分数量必须大于0")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 添加积分
		if err := s.addPoints(tx, userID, points, model.PointSourceAdminAdd,
			description, nil, nil); err != nil {
			return err
		}

		// 记录操作人
		if operatedByID != nil {
			var record model.PointRecord
			if err := tx.Where("user_id = ? AND source = ? AND description = ?",
				userID, model.PointSourceAdminAdd, description).
				Last(&record).Error; err == nil {
				record.OperatedByID = operatedByID
				if err := tx.Save(&record).Error; err != nil {
					return fmt.Errorf("记录操作人失败: %w", err)
				}
			}
		}

		return nil
	})
}

// getRuleByKey 根据规则键获取规则
func getRuleByKey(tx *gorm.DB, key model.PointSource) (*model.PointsRule, error) {
	var rule model.PointsRule
	if err := tx.Where("rule_key = ?", key).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("规则 %s 不存在", key)
		}
		return nil, fmt.Errorf("查询规则失败: %w", err)
	}
	return &rule, nil
}

// addPoints 内部方法：添加积分
func (s *EarningService) addPoints(tx *gorm.DB, userID uint, points int, source model.PointSource,
	description string, invitationID *uint, resourceID *uint) error {

	// 获取用户当前积分余额
	var user model.User
	if err := tx.Clauses(clause.Locking{}).First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 计算新的余额
	newBalance := user.PointsBalance + points
	if newBalance < 0 {
		return fmt.Errorf("积分余额不足")
	}

	// 创建积分记录
	record := model.PointRecord{
		UserID:       userID,
		Type:         model.PointTypeIncome,
		Points:       points,
		BalanceAfter: newBalance,
		Source:       source,
		Description:  description,
		InvitationID: invitationID,
		ResourceID:   resourceID,
	}

	// 如果有操作人信息，会在外层方法中设置

	if err := tx.Create(&record).Error; err != nil {
		return fmt.Errorf("创建积分记录失败: %w", err)
	}

	// 更新用户积分余额
	if err := tx.Model(&model.User{}).
		Where("id = ?", userID).
		Update("points_balance", newBalance).Error; err != nil {
		return fmt.Errorf("更新用户积分失败: %w", err)
	}

	return nil
}

// GetEarningRules 获取所有有效的积分获取规则
func (s *EarningService) GetEarningRules() ([]model.PointsRule, error) {
	var rules []model.PointsRule
	if err := s.db.Where("is_enabled = ?", true).
		Order("rule_key").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("获取规则列表失败: %w", err)
	}
	return rules, nil
}

// GetUserPointRecords 获取用户的积分记录
func (s *EarningService) GetUserPointRecords(userID uint, limit, offset int) ([]model.PointRecord, int64, error) {
	var records []model.PointRecord
	var total int64

	if err := s.db.Model(&model.PointRecord{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取记录总数失败: %w", err)
	}

	if err := s.db.Preload("Resource").
		Preload("Invitation.Inviter").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("获取积分记录失败: %w", err)
	}

	return records, total, nil
}

// GetUserPointsBalance 获取用户积分余额
func (s *EarningService) GetUserPointsBalance(userID uint) (int, error) {
	var user model.User
	if err := s.db.Select("points_balance").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("用户不存在")
		}
		return 0, fmt.Errorf("查询用户积分失败: %w", err)
	}
	return user.PointsBalance, nil
}

// CheckUserCanEarn 检查用户是否可以获取积分
func (s *EarningService) CheckUserCanEarn(userID uint, source model.PointSource) (bool, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("用户不存在")
		}
		return false, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		return false, fmt.Errorf("用户状态异常，无法获取积分")
	}

	// 检查特定来源的限制
	switch source {
	case model.PointSourceDailyCheckin:
		today := time.Now().Format("2006-01-02")
		var count int64
		s.db.Model(&model.PointRecord{}).
			Where("user_id = ? AND source = ? AND DATE(created_at) = ?",
				userID, source, today).
			Count(&count)
		return count == 0, nil

	case model.PointSourceResourceDownload:
		var count int64
		s.db.Model(&model.PointRecord{}).
			Where("user_id = ? AND source = ?", userID, source).
			Count(&count)
		return true, nil // 每次下载都可以获得积分

	case model.PointSourceInviteReward:
		return true, nil // 每次邀请都可以获得积分

	case model.PointSourceUploadReward:
		return true, nil // 每次上传都可以获得积分

	default:
		return false, fmt.Errorf("不支持的积分来源: %s", source)
	}
}

// BatchEarnPoints 批量给用户添加积分
func (s *EarningService) BatchEarnPoints(earnings []struct {
	UserID      uint
	Points      int
	Source      model.PointSource
	Description string
}) error {
	if len(earnings) == 0 {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, earning := range earnings {
			if err := s.addPoints(tx, earning.UserID, earning.Points, earning.Source,
				earning.Description, nil, nil); err != nil {
				return fmt.Errorf("用户 %d 添加积分失败: %w", earning.UserID, err)
			}
		}
		return nil
	})
}

// GetPointsStats 获取积分统计信息
func (s *EarningService) GetPointsStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总收入积分
	var totalIncome int64
	rows, err := s.db.Raw("SELECT COALESCE(SUM(points), 0) FROM point_records WHERE user_id = ? AND type = ?", userID, model.PointTypeIncome).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		var sum sql.NullInt64
		rows.Scan(&sum)
		if sum.Valid {
			totalIncome = sum.Int64
		}
	}
	stats["total_income"] = totalIncome

	// 总支出积分
	var totalExpense int64
	rows, err = s.db.Raw("SELECT COALESCE(SUM(points), 0) FROM point_records WHERE user_id = ? AND type = ?", userID, model.PointTypeExpense).Rows()
	if err != nil {
		// 如果查询失败，设置为0而不是panic
		stats["total_expense"] = 0
	} else {
		defer rows.Close()
		if rows.Next() {
			var sum sql.NullInt64
			rows.Scan(&sum)
			if sum.Valid {
				totalExpense = sum.Int64
			}
		}
		stats["total_expense"] = totalExpense
	}

	// 当前余额
	balance, err := s.GetUserPointsBalance(userID)
	if err != nil {
		return nil, err
	}
	stats["current_balance"] = balance

	// 今日获得积分
	today := time.Now().Format("2006-01-02")
	var todayIncome int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(points), 0) FROM point_records WHERE user_id = ? AND type = ? AND DATE(created_at) = ?", userID, model.PointTypeIncome, today).Rows()
	defer rows.Close()
	if rows.Next() {
		var sum sql.NullInt64
		rows.Scan(&sum)
		if sum.Valid {
			todayIncome = sum.Int64
		}
	}
	stats["today_income"] = todayIncome

	// 本月获得积分
	month := time.Now().Format("2006-01")
	var monthIncome int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(points), 0) FROM point_records WHERE user_id = ? AND type = ? AND strftime('%Y-%m', created_at) = ?", userID, model.PointTypeIncome, month).Rows()
	defer rows.Close()
	if rows.Next() {
		var sum sql.NullInt64
		rows.Scan(&sum)
		if sum.Valid {
			monthIncome = sum.Int64
		}
	}
	stats["month_income"] = monthIncome

	// 积分来源分布
	var sourceStats []struct {
		Source model.PointSource
		Total  int64
	}
	s.db.Model(&model.PointRecord{}).
		Select("source, SUM(points) as total").
		Where("user_id = ? AND type = ?", userID, model.PointTypeIncome).
		Group("source").
		Scan(&sourceStats)
	stats["source_distribution"] = sourceStats

	return stats, nil
}
