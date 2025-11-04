/*
Package model defines all data models for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package model

import (
	"resource-share-site/pkg/utils"
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Username     string `gorm:"uniqueIndex;not null;size:50" json:"username" binding:"required,min=3,max=50"`
	Email        string `gorm:"uniqueIndex;not null;size:100" json:"email" binding:"required,email"`
	PasswordHash string `gorm:"not null;size:255" json:"-"`

	// 用户状态
	Role   string `gorm:"default:'user';not null;size:20" json:"role"`     // user, admin
	Status string `gorm:"default:'active';not null;size:20" json:"status"` // active, banned

	// 上传权限
	CanUpload bool `gorm:"default:false" json:"can_upload"`

	// 邀请相关
	InviteCode   string `gorm:"uniqueIndex;size:36" json:"invite_code"`
	InvitedByID  *uint  `gorm:"index" json:"invited_by_id"`
	InvitedBy    *User  `gorm:"foreignKey:InvitedByID" json:"-"` // 自引用
	InvitedUsers []User `gorm:"foreignKey:InvitedByID" json:"-"`

	// 积分
	PointsBalance int `gorm:"default:0;not null" json:"points_balance"`

	// 关联关系
	Resources           []Resource    `gorm:"foreignKey:UploadedByID" json:"-"`
	Comments            []Comment     `gorm:"foreignKey:UserID" json:"-"`
	PointRecords        []PointRecord `gorm:"foreignKey:UserID" json:"-"`
	SentInvitations     []Invitation  `gorm:"foreignKey:InviterID" json:"-"`
	ReceivedInvitations []Invitation  `gorm:"foreignKey:InviteeID" json:"-"`

	// 审核相关
	UploadedResourcesCount   uint `gorm:"default:0" json:"uploaded_resources_count"`
	DownloadedResourcesCount uint `gorm:"default:0" json:"downloaded_resources_count"`

	// 最后登录时间
	LastLoginAt *time.Time `json:"last_login_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 设置默认状态为 active
	if u.Status == "" {
		u.Status = "active"
	}

	// 设置默认角色为 user
	if u.Role == "" {
		u.Role = "user"
	}

	// 设置默认邀请码（如果没有）
	if u.InviteCode == "" {
		u.InviteCode = utils.GenerateInviteCode()
	}

	return nil
}
