/*
Package model defines all data models for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package model

import (
	"time"
)

// SEOConfigType SEO配置类型枚举
type SEOConfigType string

const (
	SEOConfigTypeHome     SEOConfigType = "home"     // 首页
	SEOConfigTypeList     SEOConfigType = "list"     // 列表页
	SEOConfigTypeDetail   SEOConfigType = "detail"   // 详情页
	SEOConfigTypeCategory SEOConfigType = "category" // 分类页
	SEOConfigTypeResource SEOConfigType = "resource" // 资源页
)

// SEOConfig SEO配置模型
type SEOConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 配置标识
	ConfigType SEOConfigType `gorm:"not null;size:20;index" json:"config_type"` // 配置类型
	TargetID   *uint         `gorm:"index" json:"target_id"`                    // 目标ID（如分类ID、资源ID等）

	// Meta标签配置
	MetaTitle       string `gorm:"size:255" json:"meta_title"`       // 页面标题
	MetaDescription string `gorm:"size:500" json:"meta_description"` // 页面描述
	MetaKeywords    string `gorm:"size:500" json:"meta_keywords"`    // 关键词
	MetaAuthor      string `gorm:"size:100" json:"meta_author"`      // 作者

	// Open Graph标签
	OGTitle       string `gorm:"size:255" json:"og_title"`       // Open Graph标题
	OGDescription string `gorm:"size:500" json:"og_description"` // Open Graph描述
	OGImage       string `gorm:"size:255" json:"og_image"`       // Open Graph图片
	OGType        string `gorm:"size:50" json:"og_type"`         // Open Graph类型
	OGUrl         string `gorm:"size:255" json:"og_url"`         // Open Graph URL

	// Twitter Card标签
	TwitterCard        string `gorm:"size:50" json:"twitter_card"`         // Twitter Card类型
	TwitterTitle       string `gorm:"size:255" json:"twitter_title"`       // Twitter标题
	TwitterDescription string `gorm:"size:500" json:"twitter_description"` // Twitter描述
	TwitterImage       string `gorm:"size:255" json:"twitter_image"`       // Twitter图片

	// 结构化数据
	StructuredData string `gorm:"type:text" json:"structured_data"` // JSON-LD结构化数据

	// SEO优化
	CanonicalURL string  `gorm:"size:255" json:"canonical_url"`   // 规范URL
	Robots       string  `gorm:"size:100" json:"robots"`          // robots指令
	Priority     float64 `gorm:"default:0.5" json:"priority"`     // 优先级 (0.0-1.0)
	ChangeFreq   string  `gorm:"size:50" json:"change_frequency"` // 更新频率

	// 状态
	IsActive bool `gorm:"default:true" json:"is_active"` // 是否启用

	// 额外数据
	ExtraData string `gorm:"type:text" json:"extra_data"` // 额外的JSON数据
}

// TableName 指定表名
func (SEOConfig) TableName() string {
	return "seo_configs"
}

// MetaTag Meta标签模型
type MetaTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// 标签信息
	Name      string `gorm:"not null;size:100;index" json:"name"` // 标签名称
	Content   string `gorm:"not null;size:1000" json:"content"`   // 标签内容
	Property  string `gorm:"size:100" json:"property"`            // 属性 (name/property)
	HTTPEquiv string `gorm:"size:100" json:"http_equiv"`          // http-equiv属性

	// 关联信息
	PageType   SEOConfigType `gorm:"size:20;index" json:"page_type"` // 页面类型
	PageID     *uint         `gorm:"index" json:"page_id"`           // 页面ID
	ResourceID *uint         `gorm:"index" json:"resource_id"`       // 资源ID
	CategoryID *uint         `gorm:"index" json:"category_id"`       // 分类ID
}

// TableName 指定表名
func (MetaTag) TableName() string {
	return "meta_tags"
}

// SitemapUrl Sitemap URL模型
type SitemapUrl struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// URL信息
	Loc        string     `gorm:"not null;size:500;index" json:"loc"` // URL地址
	LastMod    *time.Time `json:"lastmod"`                            // 最后修改时间
	ChangeFreq string     `gorm:"size:50" json:"changefreq"`          // 更新频率
	Priority   float64    `gorm:"default:0.5" json:"priority"`        // 优先级

	// 关联信息
	PageType SEOConfigType `gorm:"size:20;index" json:"page_type"` // 页面类型
	TargetID *uint         `gorm:"index" json:"target_id"`         // 目标ID

	// 状态
	IsActive bool   `gorm:"default:true" json:"is_active"` // 是否启用
	Note     string `gorm:"size:255" json:"note"`          // 备注
}

// TableName 指定表名
func (SitemapUrl) TableName() string {
	return "sitemap_urls"
}

// SEOKeyword SEO关键词模型
type SEOKeyword struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关键词信息
	Keyword      string `gorm:"not null;size:100;uniqueIndex" json:"keyword"` // 关键词
	Language     string `gorm:"size:10;default:'zh'" json:"language"`         // 语言
	SearchVolume int    `gorm:"default:0" json:"search_volume"`               // 搜索量
	Difficulty   int    `gorm:"default:0" json:"difficulty"`                  // 难度 (0-100)

	// 分类
	Category string `gorm:"size:50" json:"category"` // 关键词分类
	Tags     string `gorm:"size:255" json:"tags"`    // 标签（逗号分隔）

	// 关联统计
	ClickRate   float64 `gorm:"default:0" json:"click_rate"`   // 点击率
	AvgPosition float64 `gorm:"default:0" json:"avg_position"` // 平均排名

	// 状态
	IsActive bool   `gorm:"default:true" json:"is_active"` // 是否启用
	Note     string `gorm:"size:255" json:"note"`          // 备注
}

// TableName 指定表名
func (SEOKeyword) TableName() string {
	return "seo_keywords"
}

// SEORank SEO排名追踪模型
type SEORank struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// 排名信息
	KeywordID uint        `gorm:"not null;index" json:"keyword_id"` // 关键词ID
	Keyword   *SEOKeyword `gorm:"foreignKey:KeywordID" json:"-"`    // 关键词

	SearchEngine string `gorm:"not null;size:20" json:"search_engine"` // 搜索引擎 (google, baidu, bing)
	Rank         int    `gorm:"not null" json:"rank"`                  // 排名位置
	URL          string `gorm:"not null;size:500" json:"url"`          // 目标URL

	// 额外信息
	Title       string `gorm:"size:255" json:"title"`       // 页面标题
	Description string `gorm:"size:500" json:"description"` // 页面描述

	// 竞争对手
	CompetitorURL  string `gorm:"size:500" json:"competitor_url"`   // 竞争URL
	CompetitorRank int    `gorm:"default:0" json:"competitor_rank"` // 竞争对手排名

	Note string `gorm:"size:255" json:"note"` // 备注
}

// TableName 指定表名
func (SEORank) TableName() string {
	return "seo_ranks"
}

// SEOReport SEO报告模型
type SEOReport struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// 报告信息
	ReportType string `gorm:"not null;size:50" json:"report_type"` // 报告类型 (daily, weekly, monthly)
	Period     string `gorm:"not null;size:50" json:"period"`      // 报告周期

	// 统计数据
	TotalPages    int     `gorm:"default:0" json:"total_pages"`    // 总页面数
	IndexedPages  int     `gorm:"default:0" json:"indexed_pages"`  // 已索引页面数
	TotalKeywords int     `gorm:"default:0" json:"total_keywords"` // 总关键词数
	AvgRank       float64 `gorm:"default:0" json:"avg_rank"`       // 平均排名

	// 性能指标
	OrganicTraffic     int     `gorm:"default:0" json:"organic_traffic"`      // 自然流量
	ClickThroughRate   float64 `gorm:"default:0" json:"click_through_rate"`   // 点击率
	BounceRate         float64 `gorm:"default:0" json:"bounce_rate"`          // 跳出率
	AvgSessionDuration float64 `gorm:"default:0" json:"avg_session_duration"` // 平均会话时长

	// SEO得分
	SEOScore int `gorm:"default:0" json:"seo_score"` // SEO得分 (0-100)

	// 报告数据
	ReportData string `gorm:"type:text" json:"report_data"` // 报告详细数据 (JSON)

	// 建议
	Recommendations string `gorm:"type:text" json:"recommendations"` // SEO建议
}

// TableName 指定表名
func (SEOReport) TableName() string {
	return "seo_reports"
}

// SEOEvent SEO事件追踪模型
type SEOEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// 事件信息
	EventType string `gorm:"not null;size:50" json:"event_type"`  // 事件类型 (crawl, index, rank_change)
	EventName string `gorm:"not null;size:100" json:"event_name"` // 事件名称

	// 关联信息
	PageURL   string `gorm:"size:500" json:"page_url"`   // 页面URL
	KeywordID *uint  `gorm:"index" json:"keyword_id"`    // 关键词ID
	OldValue  string `gorm:"size:1000" json:"old_value"` // 旧值
	NewValue  string `gorm:"size:1000" json:"new_value"` // 新值

	// 详细信息
	Details string `gorm:"type:text" json:"details"`   // 详细信息
	Source  string `gorm:"size:100" json:"source"`     // 来源 (google, baidu, manual)
	UserID  *uint  `gorm:"index" json:"user_id"`       // 操作人ID
	User    *User  `gorm:"foreignKey:UserID" json:"-"` // 操作人
}

// TableName 指定表名
func (SEOEvent) TableName() string {
	return "seo_events"
}
