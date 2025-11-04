/*
Package invitation provides invitation reward mechanism services.

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

// RewardRule 奖励规则
type RewardRule struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	BasePoints  int    `json:"base_points"` // 基础奖励积分
	Multiplier  int    `json:"multiplier"`  // 奖励倍数
	MaxRewards  int    `json:"max_rewards"` // 最大奖励次数
	MinLevel    int    `json:"min_level"`   // 最小邀请层级
	MaxLevel    int    `json:"max_level"`   // 最大邀请层级
	IsActive    bool   `json:"is_active"`   // 是否启用
	Description string `json:"description"` // 描述
}

// RewardRecord 奖励记录
type RewardRecord struct {
	ID        uint      `json:"id"`
	InviterID uint      `json:"inviter_id"`
	InviteeID uint      `json:"invitee_id"`
	Points    int       `json:"points"`
	RuleID    uint      `json:"rule_id"`
	RuleName  string    `json:"rule_name"`
	Level     int       `json:"level"` // 邀请层级
	CreatedAt time.Time `json:"created_at"`
}

// RewardService 奖励服务
type RewardService struct {
	db *gorm.DB
}

// NewRewardService 创建新的奖励服务
func NewRewardService(db *gorm.DB) *RewardService {
	return &RewardService{
		db: db,
	}
}

// GetRewardRule 获取奖励规则
// 参数：
//   - ruleID: 规则ID
//
// 返回：
//   - 奖励规则
//   - 错误信息
func (s *RewardService) GetRewardRule(ruleID uint) (*RewardRule, error) {
	var rule RewardRule
	if err := s.db.First(&rule, ruleID).Error; err != nil {
		return nil, fmt.Errorf("查询奖励规则失败: %w", err)
	}
	return &rule, nil
}

// GetDefaultRewardRule 获取默认奖励规则
// 如果没有配置奖励规则，返回默认规则
//
// 返回：
//   - 默认奖励规则
func (s *RewardService) GetDefaultRewardRule() *RewardRule {
	return &RewardRule{
		ID:          0,
		Name:        "默认奖励规则",
		BasePoints:  100,  // 默认奖励100积分
		Multiplier:  1,    // 默认倍数1
		MaxRewards:  -1,   // 无限制
		MinLevel:    1,    // 最小层级1
		MaxLevel:    1,    // 最大层级1
		IsActive:    true, // 默认启用
		Description: "完成邀请后奖励100积分",
	}
}

// CalculateReward 计算奖励积分
// 参数：
//   - inviterID: 邀请者ID
//   - inviteeID: 被邀请者ID
//   - level: 邀请层级
//   - rule: 奖励规则
//
// 返回：
//   - 奖励积分
//   - 错误信息
func (s *RewardService) CalculateReward(inviterID, inviteeID uint, level int, rule *RewardRule) (int, error) {
	// 检查规则是否有效
	if rule == nil || !rule.IsActive {
		rule = s.GetDefaultRewardRule()
	}

	// 检查层级是否在范围内
	if level < rule.MinLevel || level > rule.MaxLevel {
		return 0, fmt.Errorf("邀请层级 %d 不在规则范围内 (%d-%d)", level, rule.MinLevel, rule.MaxLevel)
	}

	// 检查是否超过最大奖励次数
	if rule.MaxRewards > 0 {
		var rewardCount int64
		if err := s.db.Model(&model.Invitation{}).
			Where("inviter_id = ? AND status = ?", inviterID, model.InvitationStatusCompleted).
			Count(&rewardCount).Error; err != nil {
			return 0, fmt.Errorf("查询奖励次数失败: %w", err)
		}
		if rewardCount >= int64(rule.MaxRewards) {
			return 0, fmt.Errorf("已达到最大奖励次数 %d", rule.MaxRewards)
		}
	}

	// 计算奖励积分
	rewardPoints := rule.BasePoints * rule.Multiplier

	return rewardPoints, nil
}

// ApplyReward 应用奖励
// 参数：
//   - inviterID: 邀请者ID
//   - inviteeID: 被邀请者ID
//   - level: 邀请层级
//   - ruleID: 奖励规则ID，0表示使用默认规则
//
// 返回：
//   - 奖励记录
//   - 错误信息
func (s *RewardService) ApplyReward(inviterID, inviteeID uint, level int, ruleID uint) (*RewardRecord, error) {
	// 获取奖励规则
	var rule *RewardRule
	var err error

	if ruleID > 0 {
		rule, err = s.GetRewardRule(ruleID)
		if err != nil {
			return nil, fmt.Errorf("获取奖励规则失败: %w", err)
		}
	} else {
		rule = s.GetDefaultRewardRule()
	}

	// 计算奖励积分
	rewardPoints, err := s.CalculateReward(inviterID, inviteeID, level, rule)
	if err != nil {
		return nil, fmt.Errorf("计算奖励失败: %w", err)
	}

	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("开始事务失败: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新邀请者积分
	if rewardPoints > 0 {
		if err := tx.Model(&model.User{}).
			Where("id = ?", inviterID).
			Update("points_balance", gorm.Expr("points_balance + ?", rewardPoints)).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新邀请者积分失败: %w", err)
		}

		// 记录积分变更
		pointRecord := &model.PointRecord{
			UserID:      inviterID,
			Points:      rewardPoints,
			Type:        model.PointTypeIncome,
			Source:      model.PointSourceInviteReward,
			Description: fmt.Sprintf("邀请奖励（%s）", rule.Name),
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(pointRecord).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("记录积分变更失败: %w", err)
		}
	}

	// 创建奖励记录
	rewardRecord := &RewardRecord{
		InviterID: inviterID,
		InviteeID: inviteeID,
		Points:    rewardPoints,
		RuleID:    ruleID,
		RuleName:  rule.Name,
		Level:     level,
		CreatedAt: time.Now(),
	}
	if err := tx.Create(rewardRecord).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建奖励记录失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return rewardRecord, nil
}

// GetRewardHistory 获取奖励历史
// 参数：
//   - inviterID: 邀请者ID
//   - page: 页码
//   - pageSize: 每页数量
//
// 返回：
//   - 奖励记录列表
//   - 总数
//   - 错误信息
func (s *RewardService) GetRewardHistory(inviterID uint, page, pageSize int) ([]*RewardRecord, int64, error) {
	query := s.db.Model(&RewardRecord{}).Where("inviter_id = ?", inviterID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询奖励记录总数失败: %w", err)
	}

	var records []*RewardRecord
	if err := query.Preload("Invitee").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("查询奖励记录列表失败: %w", err)
	}

	return records, total, nil
}

// GetRewardStats 获取奖励统计信息
// 参数：
//   - inviterID: 邀请者ID
//
// 返回：
//   - 统计信息
//   - 错误信息
func (s *RewardService) GetRewardStats(inviterID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总奖励次数
	var totalRewards int64
	if err := s.db.Model(&RewardRecord{}).Where("inviter_id = ?", inviterID).Count(&totalRewards).Error; err != nil {
		return nil, fmt.Errorf("查询总奖励次数失败: %w", err)
	}
	stats["total_rewards"] = totalRewards

	// 总奖励积分
	var totalPoints int64
	err := s.db.Model(&RewardRecord{}).Where("inviter_id = ?", inviterID).Pluck("points", &[]int{}).Error
	if err != nil {
		return nil, fmt.Errorf("查询总奖励积分失败: %w", err)
	}
	stats["total_points"] = totalPoints

	// 本月奖励次数
	var monthRewards int64
	monthStart := time.Now().AddDate(0, -1, 0)
	if err := s.db.Model(&RewardRecord{}).Where("inviter_id = ? AND created_at >= ?", inviterID, monthStart).Count(&monthRewards).Error; err != nil {
		return nil, fmt.Errorf("查询本月奖励次数失败: %w", err)
	}
	stats["month_rewards"] = monthRewards

	// 本月奖励积分
	var monthPoints int64
	if err := s.db.Model(&model.Invitation{}).Where("inviter_id = ? AND status = ? AND created_at >= ?", inviterID, model.InvitationStatusCompleted, monthStart).Pluck("points_awarded", &[]int{}).Error; err != nil {
		return nil, fmt.Errorf("查询本月奖励积分失败: %w", err)
	}
	stats["month_points"] = monthPoints

	// 平均每次奖励
	var avgPoints float64
	if totalRewards > 0 {
		avgPoints = float64(totalPoints) / float64(totalRewards)
	}
	stats["average_points_per_reward"] = avgPoints

	return stats, nil
}

// GetMultiLevelRewards 获取多层级奖励配置
// 参数：
//   - inviterID: 邀请者ID
//
// 返回：
//   - 各层级奖励配置
//   - 错误信息
func (s *RewardService) GetMultiLevelRewards(inviterID uint) (map[int]map[string]interface{}, error) {
	rewards := make(map[int]map[string]interface{})

	// 第一层奖励
	var level1Points int64
	if err := s.db.Model(&model.Invitation{}).Where("inviter_id = ? AND status = ?", inviterID, model.InvitationStatusCompleted).Pluck("points_awarded", &[]int{}).Error; err != nil {
		return nil, fmt.Errorf("查询第一层奖励失败: %w", err)
	}
	rewards[1] = map[string]interface{}{
		"level":   1,
		"points":  level1Points,
		"name":    "直接邀请奖励",
		"rule_id": 1,
	}

	// 第二层奖励
	var level2Points int64
	if err := s.db.Model(&model.Invitation{}).Where("inviter_id = ? AND status = ?", inviterID, model.InvitationStatusCompleted).Pluck("points_awarded", &[]int{}).Error; err != nil {
		return nil, fmt.Errorf("查询第二层奖励失败: %w", err)
	}
	rewards[2] = map[string]interface{}{
		"level":   2,
		"points":  level2Points,
		"name":    "二级邀请奖励",
		"rule_id": 2,
	}

	// 第三层奖励
	var level3Points int64
	if err := s.db.Model(&model.Invitation{}).Where("inviter_id = ? AND status = ?", inviterID, model.InvitationStatusCompleted).Pluck("points_awarded", &[]int{}).Error; err != nil {
		return nil, fmt.Errorf("查询第三层奖励失败: %w", err)
	}
	rewards[3] = map[string]interface{}{
		"level":   3,
		"points":  level3Points,
		"name":    "三级邀请奖励",
		"rule_id": 3,
	}

	return rewards, nil
}

// CheckRewardEligibility 检查奖励资格
// 参数：
//   - inviterID: 邀请者ID
//   - rule: 奖励规则
//
// 返回：
//   - 是否有资格
//   - 错误信息
func (s *RewardService) CheckRewardEligibility(inviterID uint, rule *RewardRule) (bool, error) {
	if rule == nil || !rule.IsActive {
		return true, nil // 默认有资格
	}

	// 检查最大奖励次数
	if rule.MaxRewards > 0 {
		var rewardCount int64
		if err := s.db.Model(&RewardRecord{}).
			Where("inviter_id = ?", inviterID).
			Count(&rewardCount).Error; err != nil {
			return false, fmt.Errorf("查询奖励次数失败: %w", err)
		}
		if rewardCount >= int64(rule.MaxRewards) {
			return false, nil
		}
	}

	return true, nil
}
