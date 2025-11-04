/*
Package invitation provides invitation code generation, validation, and relationship tracking services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package invitation

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// Errors 定义自定义错误
var (
	ErrInvalidInviteCode  = errors.New("无效的邀请码")
	ErrInviteCodeExpired  = errors.New("邀请码已过期")
	ErrUserAlreadyInvited = errors.New("用户已被邀请")
	ErrUserNotFound       = errors.New("用户不存在")
	ErrSelfInvitation     = errors.New("不能邀请自己")
)

// InvitationService 邀请服务
type InvitationService struct {
	db *gorm.DB
}

// NewInvitationService 创建新的邀请服务
func NewInvitationService(db *gorm.DB) *InvitationService {
	return &InvitationService{
		db: db,
	}
}

// GenerateInviteCode 生成邀请码
// 参数：
//   - inviterID: 邀请者ID
//   - expiresIn: 过期时间（小时），0表示默认30天
//
// 返回：
//   - 邀请码
//   - 过期时间
//   - 错误信息
func (s *InvitationService) GenerateInviteCode(inviterID uint, expiresIn int) (string, time.Time, error) {
	// 检查用户是否存在
	var user model.User
	if err := s.db.First(&user, inviterID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", time.Time{}, ErrUserNotFound
		}
		return "", time.Time{}, fmt.Errorf("查询用户失败: %w", err)
	}

	// 计算过期时间
	expiresAt := time.Now().AddDate(0, 0, 30) // 默认30天
	if expiresIn > 0 {
		expiresAt = time.Now().Add(time.Hour * time.Duration(expiresIn))
	}

	// 生成随机邀请码
	inviteCode, err := s.generateRandomCode()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("生成邀请码失败: %w", err)
	}

	return inviteCode, expiresAt, nil
}

// ValidateInviteCode 验证邀请码
// 参数：
//   - inviteCode: 邀请码
//
// 返回：
//   - 邀请者信息
//   - 邀请信息
//   - 错误信息
func (s *InvitationService) ValidateInviteCode(inviteCode string) (*model.User, *model.Invitation, error) {
	// 查找邀请码对应的邀请记录
	var invitation model.Invitation
	result := s.db.Preload("Inviter").Where("invite_code = ?", inviteCode).First(&invitation)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil, ErrInvalidInviteCode
		}
		return nil, nil, fmt.Errorf("查询邀请记录失败: %w", result.Error)
	}

	// 检查是否已过期
	if time.Now().After(invitation.ExpiresAt) {
		return nil, nil, ErrInviteCodeExpired
	}

	// 检查状态
	if invitation.Status == model.InvitationStatusCompleted {
		return nil, nil, errors.New("邀请码已被使用")
	}

	if invitation.Status == model.InvitationStatusExpired {
		return nil, nil, ErrInviteCodeExpired
	}

	return invitation.Inviter, &invitation, nil
}

// CreateInvitation 创建邀请记录
// 参数：
//   - inviterID: 邀请者ID
//   - inviteCode: 邀请码
//   - expiresAt: 过期时间
//
// 返回：
//   - 邀请记录
//   - 错误信息
func (s *InvitationService) CreateInvitation(inviterID uint, inviteCode string, expiresAt time.Time) (*model.Invitation, error) {
	// 检查邀请者是否存在
	var inviter model.User
	if err := s.db.First(&inviter, inviterID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("查询邀请者失败: %w", err)
	}

	// 检查邀请码是否已存在
	var existingInvitation model.Invitation
	if err := s.db.Where("invite_code = ?", inviteCode).First(&existingInvitation).Error; err == nil {
		return nil, errors.New("邀请码已存在")
	}

	// 创建邀请记录
	invitation := &model.Invitation{
		InviterID:  inviterID,
		InviteCode: inviteCode,
		ExpiresAt:  expiresAt,
		Status:     model.InvitationStatusPending,
	}

	if err := s.db.Create(invitation).Error; err != nil {
		return nil, fmt.Errorf("创建邀请记录失败: %w", err)
	}

	return invitation, nil
}

// CompleteInvitation 完成邀请（用户注册时调用）
// 参数：
//   - inviteCode: 邀请码
//   - inviteeID: 被邀请者ID
//   - pointsAward: 奖励积分
//
// 返回：
//   - 错误信息
func (s *InvitationService) CompleteInvitation(inviteCode string, inviteeID uint, pointsAward int) error {
	// 验证邀请码
	inviter, invitation, err := s.ValidateInviteCode(inviteCode)
	if err != nil {
		return err
	}

	// 检查不能邀请自己
	if inviter.ID == inviteeID {
		return ErrSelfInvitation
	}

	// 检查被邀请者是否已存在
	var invitee model.User
	if err := s.db.First(&invitee, inviteeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("查询被邀请者失败: %w", err)
	}

	// 检查被邀请者是否已被邀请
	if invitee.InvitedByID != nil {
		return ErrUserAlreadyInvited
	}

	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开始事务失败: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新邀请记录
	if err := tx.Model(&model.Invitation{}).
		Where("id = ?", invitation.ID).
		Updates(map[string]interface{}{
			"invitee_id":     inviteeID,
			"status":         model.InvitationStatusCompleted,
			"points_awarded": pointsAward,
			"awarded_at":     time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新邀请记录失败: %w", err)
	}

	// 更新邀请者（被邀请者）
	if err := tx.Model(&model.User{}).
		Where("id = ?", inviteeID).
		Update("invited_by_id", inviter.ID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新被邀请者邀请者信息失败: %w", err)
	}

	// 给邀请者奖励积分
	if pointsAward > 0 {
		if err := tx.Model(&model.User{}).
			Where("id = ?", inviter.ID).
			Update("points_balance", gorm.Expr("points_balance + ?", pointsAward)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("奖励邀请者积分失败: %w", err)
		}

		// 记录积分变更
		pointRecord := &model.PointRecord{
			UserID:      inviter.ID,
			Points:      pointsAward,
			Type:        model.PointTypeIncome,
			Source:      model.PointSourceInviteReward,
			Description: fmt.Sprintf("邀请用户 %s", invitee.Username),
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(pointRecord).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("记录积分变更失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// GetInvitationsByInviter 获取用户的邀请列表
// 参数：
//   - inviterID: 邀请者ID
//   - status: 状态筛选，""表示全部
//   - page: 页码
//   - pageSize: 每页数量
//
// 返回：
//   - 邀请列表
//   - 总数
//   - 错误信息
func (s *InvitationService) GetInvitationsByInviter(inviterID uint, status string, page, pageSize int) ([]*model.Invitation, int64, error) {
	query := s.db.Model(&model.Invitation{}).Where("inviter_id = ?", inviterID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询邀请总数失败: %w", err)
	}

	var invitations []*model.Invitation
	if err := query.Preload("Invitee").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&invitations).Error; err != nil {
		return nil, 0, fmt.Errorf("查询邀请列表失败: %w", err)
	}

	return invitations, total, nil
}

// GetInvitationStats 获取邀请统计信息
// 参数：
//   - inviterID: 邀请者ID
//
// 返回：
//   - 统计信息
//   - 错误信息
func (s *InvitationService) GetInvitationStats(inviterID uint) (map[string]interface{}, error) {
	var stats = make(map[string]interface{})

	// 总邀请数
	var totalInvites int64
	if err := s.db.Model(&model.Invitation{}).Where("inviter_id = ?", inviterID).Count(&totalInvites).Error; err != nil {
		return nil, fmt.Errorf("查询总邀请数失败: %w", err)
	}
	stats["total_invites"] = totalInvites

	// 已完成邀请数
	var completedInvites int64
	if err := s.db.Model(&model.Invitation{}).Where("inviter_id = ? AND status = ?", inviterID, model.InvitationStatusCompleted).Count(&completedInvites).Error; err != nil {
		return nil, fmt.Errorf("查询已完成邀请数失败: %w", err)
	}
	stats["completed_invites"] = completedInvites

	// 待注册邀请数
	var pendingInvites int64
	if err := s.db.Model(&model.Invitation{}).Where("inviter_id = ? AND status = ?", inviterID, model.InvitationStatusPending).Count(&pendingInvites).Error; err != nil {
		return nil, fmt.Errorf("查询待注册邀请数失败: %w", err)
	}
	stats["pending_invites"] = pendingInvites

	// 过期邀请数
	var expiredInvites int64
	if err := s.db.Model(&model.Invitation{}).Where("inviter_id = ? AND status = ?", inviterID, model.InvitationStatusExpired).Count(&expiredInvites).Error; err != nil {
		return nil, fmt.Errorf("查询过期邀请数失败: %w", err)
	}
	stats["expired_invites"] = expiredInvites

	// 邀请成功率
	var successRate float64
	if totalInvites > 0 {
		successRate = float64(completedInvites) / float64(totalInvites) * 100
	}
	stats["success_rate"] = successRate

	// 总奖励积分
	var totalPoints int64
	err := s.db.Model(&model.Invitation{}).Where("inviter_id = ? AND status = ?", inviterID, model.InvitationStatusCompleted).Pluck("points_awarded", &[]int{}).Error
	if err != nil {
		return nil, fmt.Errorf("查询总奖励积分失败: %w", err)
	}
	stats["total_points_earned"] = totalPoints

	return stats, nil
}

// ExpireOldInvitations 清理过期的邀请码（定时任务）
// 返回：
//   - 处理的邀请数量
//   - 错误信息
func (s *InvitationService) ExpireOldInvitations() (int64, error) {
	result := s.db.Model(&model.Invitation{}).
		Where("status = ? AND expires_at < ?", model.InvitationStatusPending, time.Now()).
		Update("status", model.InvitationStatusExpired)

	if result.Error != nil {
		return 0, fmt.Errorf("更新过期邀请状态失败: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// generateRandomCode 生成随机邀请码
// 参数：无
//
// 返回：
//   - 随机邀请码
//   - 错误信息
func (s *InvitationService) generateRandomCode() (string, error) {
	// 生成32字节的随机数据
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// 计算SHA256哈希
	hash := sha256.Sum256(bytes)

	// 转换为十六进制字符串，取前16位作为邀请码
	code := hex.EncodeToString(hash[:])[:16]

	// 转换为大写
	code = hex.EncodeToString(hash[:8]) + hex.EncodeToString(hash[8:16])
	code = code[:16]

	return fmt.Sprintf("INV-%s", code), nil
}
