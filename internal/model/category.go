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

// Category 分类模型
type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Name        string `gorm:"uniqueIndex;not null;size:50" json:"name" binding:"required"`
	Description string `gorm:"size:255" json:"description"`
	Icon        string `gorm:"size:100" json:"icon"`
	Color       string `gorm:"size:20" json:"color"`

	// 层级关系
	ParentID *uint      `gorm:"index" json:"parent_id"`
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"-"`
	Children []Category `gorm:"foreignKey:ParentID" json:"-"`

	// 排序
	SortOrder int `gorm:"default:0" json:"sort_order"`

	// 统计信息
	ResourcesCount uint `gorm:"default:0" json:"resources_count"`

	// 关联关系
	Resources []Resource `gorm:"foreignKey:CategoryID" json:"-"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}
