/*
Points Consumption Service - 积分消费规则服务

提供积分消费的完整机制，包括：
- 积分扣除
- 消费规则验证
- 消费限制检查
- 事务安全保障

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
	"gorm.io/gorm/clause"
)

// ConsumptionService 积分消费服务
type ConsumptionService struct {
	db *gorm.DB
}

// NewConsumptionService 创建新的积分消费服务
func NewConsumptionService(db *gorm.DB) *ConsumptionService {
	return &ConsumptionService{
		db: db,
	}
}

// SpendPointsForPurchase 积分购买商品
func (s *ConsumptionService) SpendPointsForPurchase(userID uint, points int, description string, productID *uint) error {
	if points <= 0 {
		return fmt.Errorf("消费积分必须大于0")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查用户积分是否足够
		if err := s.checkSufficientPoints(tx, userID, points); err != nil {
			return fmt.Errorf("积分不足: %w", err)
		}

		// 扣除积分
		if err := s.deductPoints(tx, userID, points, "expense", description, productID); err != nil {
			return fmt.Errorf("扣除积分失败: %w", err)
		}

		// 如果有商品ID，可以在这里添加其他业务逻辑（如记录购买记录等）

		return nil
	})
}

// SpendPointsForDownload 积分下载付费资源
func (s *ConsumptionService) SpendPointsForDownload(userID, resourceID uint, cost int) error {
	if cost <= 0 {
		return fmt.Errorf("下载费用必须大于0")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查用户积分是否足够
		if err := s.checkSufficientPoints(tx, userID, cost); err != nil {
			return fmt.Errorf("积分不足，无法下载: %w", err)
		}

		// 扣除积分
		description := fmt.Sprintf("下载付费资源 #%d", resourceID)
		if err := s.deductPoints(tx, userID, cost, "expense", description, &resourceID); err != nil {
			return fmt.Errorf("扣除积分失败: %w", err)
		}

		// 更新资源下载次数
		if err := tx.Model(&model.Resource{}).
			Where("id = ?", resourceID).
			Update("downloads_count", gorm.Expr("downloads_count + 1")).Error; err != nil {
			return fmt.Errorf("更新下载次数失败: %w", err)
		}

		// 更新用户下载次数
		if err := tx.Model(&model.User{}).
			Where("id = ?", userID).
			Update("downloaded_resources_count", gorm.Expr("downloaded_resources_count + 1")).Error; err != nil {
			return fmt.Errorf("更新用户下载次数失败: %w", err)
		}

		return nil
	})
}

// SpendPointsForVipUpgrade 积分升级VIP
func (s *ConsumptionService) SpendPointsForVipUpgrade(userID uint, vipLevel string, cost int) error {
	if cost <= 0 {
		return fmt.Errorf("VIP升级费用必须大于0")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查用户积分是否足够
		if err := s.checkSufficientPoints(tx, userID, cost); err != nil {
			return fmt.Errorf("积分不足，无法升级VIP: %w", err)
		}

		// 扣除积分
		description := fmt.Sprintf("升级VIP: %s", vipLevel)
		if err := s.deductPoints(tx, userID, cost, "expense", description, nil); err != nil {
			return fmt.Errorf("扣除积分失败: %w", err)
		}

		// 这里可以添加其他VIP升级逻辑

		return nil
	})
}

// SpendPointsForAdvertisement 积分广告投放
func (s *ConsumptionService) SpendPointsForAdvertisement(userID uint, adType string, durationDays int, costPerDay int) error {
	totalCost := durationDays * costPerDay
	if totalCost <= 0 {
		return fmt.Errorf("广告费用计算错误")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查用户积分是否足够
		if err := s.checkSufficientPoints(tx, userID, totalCost); err != nil {
			return fmt.Errorf("积分不足，无法投放广告: %w", err)
		}

		// 扣除积分
		description := fmt.Sprintf("投放广告: %s (%d天)", adType, durationDays)
		if err := s.deductPoints(tx, userID, totalCost, "expense", description, nil); err != nil {
			return fmt.Errorf("扣除积分失败: %w", err)
		}

		// 这里可以添加广告投放逻辑

		return nil
	})
}

// RefundPoints 积分退款
func (s *ConsumptionService) RefundPoints(userID uint, originalRecordID uint, points int, reason string) error {
	if points <= 0 {
		return fmt.Errorf("退款积分必须大于0")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取原始消费记录
		var originalRecord model.PointRecord
		if err := tx.First(&originalRecord, originalRecordID).Error; err != nil {
			return fmt.Errorf("原始消费记录不存在: %w", err)
		}

		// 验证用户和记录匹配
		if originalRecord.UserID != userID {
			return fmt.Errorf("用户与消费记录不匹配")
		}

		if originalRecord.Type != model.PointTypeExpense {
			return fmt.Errorf("只能对支出记录进行退款")
		}

		// 退还积分
		description := fmt.Sprintf("退款: %s", reason)
		earningService := NewEarningService(s.db)
		if err := earningService.addPoints(tx, userID, points, model.PointSourceAdminAdd,
			description, nil, originalRecord.ResourceID); err != nil {
			return fmt.Errorf("退款失败: %w", err)
		}

		return nil
	})
}

// checkSufficientPoints 检查积分是否足够
func (s *ConsumptionService) checkSufficientPoints(tx *gorm.DB, userID uint, requiredPoints int) error {
	var user model.User
	if err := tx.Clauses(clause.Locking{}).First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	if user.PointsBalance < requiredPoints {
		return fmt.Errorf("当前积分 %d，所需积分 %d，差额 %d",
			user.PointsBalance, requiredPoints, requiredPoints-user.PointsBalance)
	}

	return nil
}

// deductPoints 内部方法：扣除积分
func (s *ConsumptionService) deductPoints(tx *gorm.DB, userID uint, points int, expenseType string,
	description string, resourceID *uint) error {

	// 获取用户当前积分余额
	var user model.User
	if err := tx.Clauses(clause.Locking{}).First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 计算新的余额
	newBalance := user.PointsBalance - points
	if newBalance < 0 {
		return fmt.Errorf("积分余额不足")
	}

	// 创建积分记录
	record := model.PointRecord{
		UserID:       userID,
		Type:         model.PointTypeExpense,
		Points:       -points, // 支出为负数
		BalanceAfter: newBalance,
		Source:       model.PointSource("expense_" + expenseType),
		Description:  description,
		ResourceID:   resourceID,
	}

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

// GetConsumptionHistory 获取消费历史
func (s *ConsumptionService) GetConsumptionHistory(userID uint, limit, offset int) ([]model.PointRecord, int64, error) {
	var records []model.PointRecord
	var total int64

	if err := s.db.Model(&model.PointRecord{}).
		Where("user_id = ? AND type = ?", userID, model.PointTypeExpense).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取记录总数失败: %w", err)
	}

	if err := s.db.Preload("Resource").
		Where("user_id = ? AND type = ?", userID, model.PointTypeExpense).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("获取消费记录失败: %w", err)
	}

	return records, total, nil
}

// GetConsumptionStats 获取消费统计
func (s *ConsumptionService) GetConsumptionStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总消费积分
	var totalConsumption int64
	rows, _ := s.db.Raw("SELECT COALESCE(SUM(ABS(points)), 0) FROM point_records WHERE user_id = ? AND type = ?", userID, model.PointTypeExpense).Rows()
	defer rows.Close()
	if rows.Next() {
		var sum *int64
		rows.Scan(&sum)
		if sum != nil {
			totalConsumption = *sum
		}
	}
	stats["total_consumption"] = totalConsumption

	// 今日消费
	today := time.Now().Format("2006-01-02")
	var todayConsumption int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(ABS(points)), 0) FROM point_records WHERE user_id = ? AND type = ? AND DATE(created_at) = ?", userID, model.PointTypeExpense, today).Rows()
	defer rows.Close()
	if rows.Next() {
		var sum *int64
		rows.Scan(&sum)
		if sum != nil {
			todayConsumption = *sum
		}
	}
	stats["today_consumption"] = todayConsumption

	// 本月消费
	month := time.Now().Format("2006-01")
	var monthConsumption int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(ABS(points)), 0) FROM point_records WHERE user_id = ? AND type = ? AND strftime('%Y-%m', created_at) = ?", userID, model.PointTypeExpense, month).Rows()
	defer rows.Close()
	if rows.Next() {
		var sum *int64
		rows.Scan(&sum)
		if sum != nil {
			monthConsumption = *sum
		}
	}
	stats["month_consumption"] = monthConsumption

	// 平均每日消费
	daysInMonth := time.Now().Day()
	if daysInMonth > 0 {
		stats["avg_daily_consumption"] = monthConsumption / int64(daysInMonth)
	} else {
		stats["avg_daily_consumption"] = int64(0)
	}

	// 消费来源分布
	var sourceStats []struct {
		Source string
		Total  int64
	}
	s.db.Model(&model.PointRecord{}).
		Select("source, SUM(ABS(points)) as total").
		Where("user_id = ? AND type = ?", userID, model.PointTypeExpense).
		Group("source").
		Scan(&sourceStats)
	stats["source_distribution"] = sourceStats

	return stats, nil
}

// CheckUserCanSpend 检查用户是否可以消费积分
func (s *ConsumptionService) CheckUserCanSpend(userID uint, points int) (bool, error) {
	if points <= 0 {
		return false, fmt.Errorf("消费积分必须大于0")
	}

	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("用户不存在")
		}
		return false, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		return false, fmt.Errorf("用户状态异常，无法消费积分")
	}

	// 检查积分是否足够
	return user.PointsBalance >= points, nil
}

// BatchSpendPoints 批量消费积分
func (s *ConsumptionService) BatchSpendPoints(spendings []struct {
	UserID      uint
	Points      int
	Description string
	ResourceID  *uint
}) error {
	if len(spendings) == 0 {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, spending := range spendings {
			if err := s.deductPoints(tx, spending.UserID, spending.Points, "batch",
				spending.Description, spending.ResourceID); err != nil {
				return fmt.Errorf("用户 %d 消费积分失败: %w", spending.UserID, err)
			}
		}
		return nil
	})
}
