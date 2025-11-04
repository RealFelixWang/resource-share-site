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

// Ad 广告模型
type Ad struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title      string `gorm:"not null;size:100" json:"title"`      // 广告标题
	ImageURL   string `gorm:"not null;size:500" json:"image_url"`  // 图片链接
	LinkURL    string `gorm:"not null;size:500" json:"link_url"`   // 跳转链接
	AdPosition string `gorm:"not null;size:50" json:"ad_position"` // 广告位置

	// 显示设置
	IsActive   bool `gorm:"default:true" json:"is_active"` // 是否启用
	SortOrder  int  `gorm:"default:0" json:"sort_order"`   // 排序
	ClickCount uint `gorm:"default:0" json:"click_count"`  // 点击次数

	// 时间设置
	StartDate *time.Time `json:"start_date"` // 开始日期
	EndDate   *time.Time `json:"end_date"`   // 结束日期
}

// TableName 指定表名
func (Ad) TableName() string {
	return "ads"
}
