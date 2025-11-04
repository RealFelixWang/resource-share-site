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

// ArticleStatus 文章状态枚举
type ArticleStatus string

const (
	ArticleStatusDraft   ArticleStatus = "draft"   // 草稿
	ArticleStatusPending ArticleStatus = "pending" // 待审核
	ArticleStatusPublished ArticleStatus = "published" // 已发布
	ArticleStatusArchived  ArticleStatus = "archived"  // 已归档
)

// Article 文章模型
type Article struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Title   string `gorm:"not null;size:200;index" json:"title" binding:"required,min=1,max=200"`
	Slug    string `gorm:"not null;uniqueIndex;size:200" json:"slug" binding:"required"`
	Content string `gorm:"not null;type:longtext" json:"content" binding:"required"`
	Excerpt string `gorm:"size:500" json:"excerpt"`

	// 媒体资源
	FeaturedImage string `gorm:"size:500" json:"featured_image"`
	Tags          string `gorm:"size:200" json:"tags"` // 逗号分隔的标签

	// 分类
	Category string `gorm:"size:100;index" json:"category"`

	// 状态
	Status ArticleStatus `gorm:"default:'draft';not null;size:20" json:"status"`

	// 作者信息
	AuthorID uint  `gorm:"not null;index" json:"author_id"`
	Author   *User `gorm:"foreignKey:AuthorID" json:"author"`

	// 发布时间
	PublishedAt *time.Time `gorm:"index" json:"published_at"`

	// 统计数据
	ViewCount    uint `gorm:"default:0;not null" json:"view_count"`
	LikeCount    uint `gorm:"default:0;not null" json:"like_count"`
	CommentCount uint `gorm:"default:0;not null" json:"comment_count"`

	// 审核信息
	ReviewedByID *uint      `gorm:"index" json:"reviewed_by_id"`
	ReviewedBy   *User      `gorm:"foreignKey:ReviewedByID" json:"-"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
	ReviewNotes  string     `gorm:"size:500" json:"review_notes"`

	// SEO字段
	MetaTitle       string `gorm:"size:200" json:"meta_title"`
	MetaDescription string `gorm:"size:500" json:"meta_description"`
	MetaKeywords    string `gorm:"size:200" json:"meta_keywords"`
}

// TableName 指定表名
func (Article) TableName() string {
	return "articles"
}

// BeforeCreate 创建前钩子
func (a *Article) BeforeCreate(tx *gorm.DB) error {
	if a.Status == "" {
		a.Status = ArticleStatusDraft
	}
	if a.ViewCount == 0 {
		a.ViewCount = 0
	}
	if a.LikeCount == 0 {
		a.LikeCount = 0
	}
	if a.CommentCount == 0 {
		a.CommentCount = 0
	}
	return nil
}

// ArticleComment 文章评论模型
type ArticleComment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 基本信息
	Content string `gorm:"not null;type:text;size:1000" json:"content" binding:"required,min=1,max=1000"`

	// 关联信息
	UserID uint  `gorm:"not null;index" json:"user_id"`
	User   *User `gorm:"foreignKey:UserID" json:"user"`

	ArticleID uint   `gorm:"not null;index" json:"article_id"`
	Article   *Article `gorm:"foreignKey:ArticleID" json:"-"`

	// 父评论（用于回复）
	ParentID *uint         `gorm:"index" json:"parent_id"`
	Parent   *ArticleComment `gorm:"foreignKey:ParentID" json:"-"`

	// 状态
	Status CommentStatus `gorm:"default:'pending';not null;size:20" json:"status"`

	// 审核信息
	ReviewedByID *uint      `gorm:"index" json:"reviewed_by_id"`
	ReviewedBy   *User      `gorm:"foreignKey:ReviewedByID" json:"-"`
	ReviewedAt   *time.Time `json:"reviewed_at"`
	ReviewNotes  string     `gorm:"size:500" json:"review_notes"`

	// 点赞数
	LikeCount uint `gorm:"default:0;not null" json:"like_count"`
}

// TableName 指定表名
func (ArticleComment) TableName() string {
	return "article_comments"
}

// BeforeCreate 创建前钩子
func (ac *ArticleComment) BeforeCreate(tx *gorm.DB) error {
	if ac.Status == "" {
		ac.Status = CommentStatusPending
	}
	if ac.LikeCount == 0 {
		ac.LikeCount = 0
	}
	return nil
}
