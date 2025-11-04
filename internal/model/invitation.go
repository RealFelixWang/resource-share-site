/*
Package model defines all data models for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package model

import (
	"time"

	"gorm.io/gorm"
)

// InvitationStatus 邀请状态枚举
type InvitationStatus string

const (
	InvitationStatusPending   InvitationStatus = "pending"   // 待注册
	InvitationStatusCompleted InvitationStatus = "completed" // 已完成
	InvitationStatusExpired   InvitationStatus = "expired"   // 已过期
)

// Invitation 邀请模型
type Invitation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 邀请关系
	InviterID uint  `gorm:"not null;index" json:"inviter_id"`
	Inviter   *User `gorm:"foreignKey:InviterID" json:"inviter"`

	InviteeID *uint `gorm:"index" json:"invitee_id"`
	Invitee   *User `gorm:"foreignKey:InviteeID" json:"invitee"`

	// 邀请码
	InviteCode string `gorm:"not null;uniqueIndex;size:36" json:"invite_code"`

	// 状态和积分奖励
	Status        InvitationStatus `gorm:"default:'pending';not null;size:20" json:"status"`
	PointsAwarded int              `gorm:"default:0;not null" json:"points_awarded"`
	AwardedAt     *time.Time       `json:"awarded_at"`

	// 过期时间
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
}

// TableName 指定表名
func (Invitation) TableName() string {
	return "invitations"
}

// BeforeCreate 创建钩子
func (i *Invitation) BeforeCreate(tx *gorm.DB) error {
	// 设置默认状态为待注册
	if i.Status == "" {
		i.Status = InvitationStatusPending
	}

	// 设置默认过期时间为 30 天后
	if i.ExpiresAt.IsZero() {
		i.ExpiresAt = time.Now().AddDate(0, 0, 30)
	}

	return nil
}
