/*
Package model defines all data models for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package model

import (
	"time"
)

// PointType 积分类型枚举
type PointType string

const (
	PointTypeIncome  PointType = "income"  // 收入
	PointTypeExpense PointType = "expense" // 支出
)

// PointSource 积分来源枚举
type PointSource string

const (
	PointSourceInviteReward     PointSource = "invite_reward"     // 邀请奖励
	PointSourceResourceDownload PointSource = "resource_download" // 资源下载
	PointSourceAdminAdd         PointSource = "admin_add"         // 管理员添加
	PointSourceDailyCheckin     PointSource = "daily_checkin"     // 每日签到
	PointSourceUploadReward     PointSource = "upload_reward"     // 上传奖励
)

// PointRecord 积分记录模型
type PointRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// 用户信息
	UserID uint  `gorm:"not null;index" json:"user_id"`
	User   *User `gorm:"foreignKey:UserID" json:"user"`

	// 积分变动
	Type         PointType `gorm:"not null;size:10" json:"type"`  // income 或 expense
	Points       int       `gorm:"not null" json:"points"`        // 正数为收入，负数为支出
	BalanceAfter int       `gorm:"not null" json:"balance_after"` // 变动后的余额

	// 来源信息
	Source PointSource `gorm:"not null;size:30" json:"source"`

	// 关联信息（可选）
	ResourceID *uint     `gorm:"index" json:"resource_id"`
	Resource   *Resource `gorm:"foreignKey:ResourceID" json:"-"`

	InvitationID *uint       `gorm:"index" json:"invitation_id"`
	Invitation   *Invitation `gorm:"foreignKey:InvitationID" json:"-"`

	// 描述信息
	Description string `gorm:"size:255" json:"description"`

	// 操作人（管理员操作时）
	OperatedByID *uint `gorm:"index" json:"operated_by_id"`
	OperatedBy   *User `gorm:"foreignKey:OperatedByID" json:"-"`
}

// TableName 指定表名
func (PointRecord) TableName() string {
	return "point_records"
}
