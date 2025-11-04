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

// CommentStatus 评论状态枚举
type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"  // 待审核
	CommentStatusApproved CommentStatus = "approved" // 已通过
	CommentStatusRejected CommentStatus = "rejected" // 已拒绝
)

// Comment 评论模型
type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Content string `gorm:"not null;type:text;size:1000" json:"content" binding:"required,min=1,max=1000"`

	// 关联信息
	UserID uint  `gorm:"not null;index" json:"user_id"`
	User   *User `gorm:"foreignKey:UserID" json:"user"`

	ResourceID uint      `gorm:"not null;index" json:"resource_id"`
	Resource   *Resource `gorm:"foreignKey:ResourceID" json:"-"`

	// 状态
	Status CommentStatus `gorm:"default:'pending';not null;size:20" json:"status"`

	// 审核信息
	ReviewedByID *uint      `gorm:"index" json:"reviewed_by_id"`
	ReviewedBy   *User      `gorm:"foreignKey:ReviewedByID" json:"-"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
	ReviewNotes  string     `gorm:"size:500" json:"review_notes"`

	// 回复评论 ID（用于支持嵌套评论）
	ParentID *uint     `gorm:"index" json:"parent_id"`
	Parent   *Comment  `gorm:"foreignKey:ParentID" json:"-"`
	Replies  []Comment `gorm:"foreignKey:ParentID" json:"-"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}

// BeforeCreate 创建钩子
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	// 设置默认状态为待审核
	if c.Status == "" {
		c.Status = CommentStatusPending
	}

	return nil
}
