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

// ResourceStatus 资源状态枚举
type ResourceStatus string

const (
	ResourceStatusPending  ResourceStatus = "pending"  // 待审核
	ResourceStatusApproved ResourceStatus = "approved" // 已发布
	ResourceStatusRejected ResourceStatus = "rejected" // 已拒绝
)

// ResourceSource 资源来源枚举
type ResourceSource string

const (
	ResourceSourceManual  ResourceSource = "manual"  // 管理员手动添加
	ResourceSourceUser    ResourceSource = "user"    // 用户上传
	ResourceSourceCrawler ResourceSource = "crawler" // 爬虫抓取
	ResourceSourceExcel   ResourceSource = "excel"   // Excel 导入
)

// Resource 资源模型
type Resource struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Title       string    `gorm:"not null;size:200" json:"title" binding:"required,min=1,max=200"`
	Description string    `gorm:"type:text" json:"description"`
	CategoryID  uint      `gorm:"not null;index" json:"category_id"`
	Category    *Category `gorm:"foreignKey:CategoryID" json:"category"`

	// 资源信息
	NetdiskURL  string `gorm:"not null;size:500" json:"netdisk_url" binding:"required,url"`
	PointsPrice int    `gorm:"default:0;not null" json:"points_price"` // 0 表示免费

	// 来源信息
	Source ResourceSource `gorm:"not null;size:20" json:"source"`

	// 上传者
	UploadedByID uint  `gorm:"not null;index" json:"uploaded_by_id"`
	UploadedBy   *User `gorm:"foreignKey:UploadedByID" json:"uploader"`

	// 状态
	Status ResourceStatus `gorm:"default:'pending';not null;size:20" json:"status"`

	// 审核信息
	ReviewedByID *uint      `gorm:"index" json:"reviewed_by_id"`
	ReviewedBy   *User      `gorm:"foreignKey:ReviewedByID" json:"-"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
	ReviewNotes  string     `gorm:"size:500" json:"review_notes"`

	// 统计信息
	DownloadsCount uint `gorm:"default:0" json:"downloads_count"`
	ViewsCount     uint `gorm:"default:0" json:"views_count"`

	// 标签（使用 JSON 格式存储）
	Tags string `gorm:"type:text;size:1000" json:"tags"`

	// 关联关系
	Comments []Comment `gorm:"foreignKey:ResourceID" json:"-"`

	// 导入任务 ID（如果是导入的）
	ImportTaskID *uint `gorm:"index" json:"import_task_id"`
}

// TableName 指定表名
func (Resource) TableName() string {
	return "resources"
}

// BeforeCreate 创建钩子
func (r *Resource) BeforeCreate(tx *gorm.DB) error {
	// 设置默认状态为待审核
	if r.Status == "" {
		r.Status = ResourceStatusPending
	}

	// 设置默认来源为用户上传
	if r.Source == "" {
		r.Source = ResourceSourceUser
	}

	return nil
}
