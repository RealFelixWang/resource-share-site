/*
Package user provides user management services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package user

import (
	"errors"
	"time"

	"resource-share-site/internal/model"
	"resource-share-site/internal/service/auth"

	"gorm.io/gorm"
)

// UserStatus 用户状态管理服务
type UserStatusService interface {
	// 获取用户状态
	GetUserStatus(ctx *auth.GORMContext, userID uint) (*UserStatus, error)

	// 封禁用户
	BanUser(ctx *auth.GORMContext, adminID uint, userID uint, reason string, duration time.Duration) error

	// 解封用户
	UnbanUser(ctx *auth.GORMContext, adminID uint, userID uint, reason string) error

	// 激活用户
	ActivateUser(ctx *auth.GORMContext, adminID uint, userID uint, reason string) error

	// 禁用用户
	DeactivateUser(ctx *auth.GORMContext, adminID uint, userID uint, reason string) error

	// 批量封禁
	BatchBanUsers(ctx *auth.GORMContext, adminID uint, userIDs []uint, reason string, duration time.Duration) error

	// 批量解封
	BatchUnbanUsers(ctx *auth.GORMContext, adminID uint, userIDs []uint, reason string) error

	// 检查用户是否可以登录
	IsUserActive(ctx *auth.GORMContext, userID uint) (bool, error)

	// 获取被封禁的用户列表
	GetBannedUsers(ctx *auth.GORMContext, page, pageSize int) ([]UserStatus, int64, error)

	// 获取用户状态变更历史
	GetUserStatusHistory(ctx *auth.GORMContext, userID uint, limit int) ([]StatusChangeLog, error)
}

// UserStatus 用户状态信息
type UserStatus struct {
	ID            uint       `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	Status        string     `json:"status"`         // active, banned, inactive
	CanUpload     bool       `json:"can_upload"`     // 是否有上传权限
	PointsBalance int        `json:"points_balance"` // 积分余额
	InviteCode    string     `json:"invite_code"`    // 邀请码
	LastLoginAt   *time.Time `json:"last_login_at"`  // 最后登录时间
	CreatedAt     time.Time  `json:"created_at"`     // 创建时间
	UpdatedAt     time.Time  `json:"updated_at"`     // 更新时间

	// 封禁相关信息
	IsBanned  bool       `json:"is_banned"`  // 是否被封禁
	BanReason string     `json:"ban_reason"` // 封禁原因
	BannedAt  *time.Time `json:"banned_at"`  // 封禁时间
	BannedBy  string     `json:"banned_by"`  // 封禁者
	UnbanAt   *time.Time `json:"unban_at"`   // 解封时间
	AutoUnban bool       `json:"auto_unban"` // 是否自动解封
}

// StatusChangeLog 状态变更日志
type StatusChangeLog struct {
	ID         uint          `json:"id"`
	UserID     uint          `json:"user_id"`
	FromStatus string        `json:"from_status"` // 原状态
	ToStatus   string        `json:"to_status"`   // 新状态
	Action     string        `json:"action"`      // 操作类型: ban, unban, activate, deactivate
	Reason     string        `json:"reason"`      // 变更原因
	OperatorID uint          `json:"operator_id"` // 操作人ID
	Operator   string        `json:"operator"`    // 操作人用户名
	Duration   time.Duration `json:"duration"`    // 封禁时长
	CreatedAt  time.Time     `json:"created_at"`  // 变更时间
}

// UserStatusServiceImpl 用户状态管理服务实现
type UserStatusServiceImpl struct {
	db *gorm.DB
}

// NewUserStatusService 创建用户状态管理服务
func NewUserStatusService(db *gorm.DB) UserStatusService {
	return &UserStatusServiceImpl{
		db: db,
	}
}

// GetUserStatus 获取用户状态
func (s *UserStatusServiceImpl) GetUserStatus(ctx *auth.GORMContext, userID uint) (*UserStatus, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// TODO: 获取封禁信息（如果有）
	// 可以从封禁表或管理员日志中获取

	return &UserStatus{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Status:        user.Status,
		CanUpload:     user.CanUpload,
		PointsBalance: user.PointsBalance,
		InviteCode:    user.InviteCode,
		LastLoginAt:   user.LastLoginAt,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
		IsBanned:      user.Status == "banned",
		BanReason:     "",    // TODO: 从封禁表中获取
		BannedAt:      nil,   // TODO: 从封禁表中获取
		BannedBy:      "",    // TODO: 从封禁表中获取
		UnbanAt:       nil,   // TODO: 从封禁表中获取
		AutoUnban:     false, // TODO: 从封禁表中获取
	}, nil
}

// BanUser 封禁用户
func (s *UserStatusServiceImpl) BanUser(ctx *auth.GORMContext, adminID uint, userID uint, reason string, duration time.Duration) error {
	// 开启事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取用户
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		// 检查用户状态
		if user.Status == "banned" {
			return errors.New("用户已被封禁")
		}

		// 更新用户状态
		if err := tx.Model(&user).Update("status", "banned").Error; err != nil {
			return err
		}

		// 记录管理员操作日志
		adminLog := model.AdminLog{
			AdminID:    adminID,
			Action:     "ban_user",
			TargetType: "user",
			TargetID:   userID,
			BeforeData: `{"status": "` + user.Status + `"}`,
			AfterData:  `{"status": "banned", "reason": "` + reason + `", "duration": "` + duration.String() + `"}`,
			CreatedAt:  time.Now(),
		}
		if err := tx.Create(&adminLog).Error; err != nil {
			return err
		}

		// TODO: 如果需要持久化封禁信息，可以创建封禁记录
		// 例如：创建Ban记录到专门的表或使用admin_logs

		return nil
	})
}

// UnbanUser 解封用户
func (s *UserStatusServiceImpl) UnbanUser(ctx *auth.GORMContext, adminID uint, userID uint, reason string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取用户
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		// 检查用户状态
		if user.Status != "banned" {
			return errors.New("用户未被封禁")
		}

		// 更新用户状态
		if err := tx.Model(&user).Update("status", "active").Error; err != nil {
			return err
		}

		// 记录管理员操作日志
		adminLog := model.AdminLog{
			AdminID:    adminID,
			Action:     "unban_user",
			TargetType: "user",
			TargetID:   userID,
			BeforeData: `{"status": "banned"}`,
			AfterData:  `{"status": "active", "reason": "` + reason + `"}`,
			CreatedAt:  time.Now(),
		}
		if err := tx.Create(&adminLog).Error; err != nil {
			return err
		}

		return nil
	})
}

// ActivateUser 激活用户
func (s *UserStatusServiceImpl) ActivateUser(ctx *auth.GORMContext, adminID uint, userID uint, reason string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取用户
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		// 检查用户状态
		if user.Status == "active" {
			return errors.New("用户已经是激活状态")
		}

		// 记录原状态
		oldStatus := user.Status

		// 更新用户状态
		if err := tx.Model(&user).Update("status", "active").Error; err != nil {
			return err
		}

		// 记录管理员操作日志
		adminLog := model.AdminLog{
			AdminID:    adminID,
			Action:     "activate_user",
			TargetType: "user",
			TargetID:   userID,
			BeforeData: `{"status": "` + oldStatus + `"}`,
			AfterData:  `{"status": "active", "reason": "` + reason + `"}`,
			CreatedAt:  time.Now(),
		}
		if err := tx.Create(&adminLog).Error; err != nil {
			return err
		}

		return nil
	})
}

// DeactivateUser 禁用用户
func (s *UserStatusServiceImpl) DeactivateUser(ctx *auth.GORMContext, adminID uint, userID uint, reason string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取用户
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		// 检查用户状态
		if user.Status == "inactive" {
			return errors.New("用户已经是禁用状态")
		}

		// 记录原状态
		oldStatus := user.Status

		// 更新用户状态
		if err := tx.Model(&user).Update("status", "inactive").Error; err != nil {
			return err
		}

		// 记录管理员操作日志
		adminLog := model.AdminLog{
			AdminID:    adminID,
			Action:     "deactivate_user",
			TargetType: "user",
			TargetID:   userID,
			BeforeData: `{"status": "` + oldStatus + `"}`,
			AfterData:  `{"status": "inactive", "reason": "` + reason + `"}`,
			CreatedAt:  time.Now(),
		}
		if err := tx.Create(&adminLog).Error; err != nil {
			return err
		}

		return nil
	})
}

// BatchBanUsers 批量封禁用户
func (s *UserStatusServiceImpl) BatchBanUsers(ctx *auth.GORMContext, adminID uint, userIDs []uint, reason string, duration time.Duration) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 批量更新用户状态
		if err := tx.Model(&model.User{}).Where("id IN ?", userIDs).Update("status", "banned").Error; err != nil {
			return err
		}

		// 批量记录管理员操作日志
		adminLogs := make([]model.AdminLog, len(userIDs))
		for i, userID := range userIDs {
			adminLogs[i] = model.AdminLog{
				AdminID:    adminID,
				Action:     "batch_ban_users",
				TargetType: "user",
				TargetID:   userID,
				BeforeData: `{"status": "active"}`,
				AfterData:  `{"status": "banned", "reason": "` + reason + `", "duration": "` + duration.String() + `"}`,
				CreatedAt:  time.Now(),
			}
		}

		if err := tx.CreateInBatches(adminLogs, 100).Error; err != nil {
			return err
		}

		return nil
	})
}

// BatchUnbanUsers 批量解封用户
func (s *UserStatusServiceImpl) BatchUnbanUsers(ctx *auth.GORMContext, adminID uint, userIDs []uint, reason string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 批量更新用户状态
		if err := tx.Model(&model.User{}).Where("id IN ?", userIDs).Update("status", "active").Error; err != nil {
			return err
		}

		// 批量记录管理员操作日志
		adminLogs := make([]model.AdminLog, len(userIDs))
		for i, userID := range userIDs {
			adminLogs[i] = model.AdminLog{
				AdminID:    adminID,
				Action:     "batch_unban_users",
				TargetType: "user",
				TargetID:   userID,
				BeforeData: `{"status": "banned"}`,
				AfterData:  `{"status": "active", "reason": "` + reason + `"}`,
				CreatedAt:  time.Now(),
			}
		}

		if err := tx.CreateInBatches(adminLogs, 100).Error; err != nil {
			return err
		}

		return nil
	})
}

// IsUserActive 检查用户是否可以登录
func (s *UserStatusServiceImpl) IsUserActive(ctx *auth.GORMContext, userID uint) (bool, error) {
	var user model.User
	if err := s.db.Select("status").First(&user, userID).Error; err != nil {
		return false, err
	}

	return user.Status == "active", nil
}

// GetBannedUsers 获取被封禁的用户列表
func (s *UserStatusServiceImpl) GetBannedUsers(ctx *auth.GORMContext, page, pageSize int) ([]UserStatus, int64, error) {
	var users []model.User
	var total int64

	// 查询被封禁的用户
	if err := s.db.Model(&model.User{}).Where("status = ?", "banned").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Where("status = ?", "banned").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// 转换为UserStatus
	result := make([]UserStatus, len(users))
	for i, user := range users {
		result[i] = UserStatus{
			ID:            user.ID,
			Username:      user.Username,
			Email:         user.Email,
			Status:        user.Status,
			CanUpload:     user.CanUpload,
			PointsBalance: user.PointsBalance,
			InviteCode:    user.InviteCode,
			LastLoginAt:   user.LastLoginAt,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
			IsBanned:      true,
		}
	}

	return result, total, nil
}

// GetUserStatusHistory 获取用户状态变更历史
func (s *UserStatusServiceImpl) GetUserStatusHistory(ctx *auth.GORMContext, userID uint, limit int) ([]StatusChangeLog, error) {
	var logs []model.AdminLog

	// 查询状态变更相关的日志
	if err := s.db.Where("target_type = ? AND target_id = ? AND action IN ?", "user", userID, []string{"ban_user", "unban_user", "activate_user", "deactivate_user"}).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}

	// 转换为StatusChangeLog
	result := make([]StatusChangeLog, len(logs))
	for i, log := range logs {
		// 从BeforeData和AfterData中提取信息
		var fromStatus, toStatus string
		// TODO: 解析JSON数据获取详细状态信息

		result[i] = StatusChangeLog{
			ID:         log.ID,
			UserID:     userID,
			FromStatus: fromStatus,
			ToStatus:   toStatus,
			Action:     log.Action,
			Reason:     "", // TODO: 从AfterData中提取
			OperatorID: log.AdminID,
			Operator:   "", // TODO: 查询操作人用户名
			Duration:   0,  // TODO: 从AfterData中提取
			CreatedAt:  log.CreatedAt,
		}
	}

	return result, nil
}
