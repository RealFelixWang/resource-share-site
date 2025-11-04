/*
Package model defines all data models for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package model

import (
	"time"
)

// VisitLog 访问记录模型
type VisitLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// 访问者信息
	UserID *uint `gorm:"index" json:"user_id"` // 可以为空（匿名访问）
	User   *User `gorm:"foreignKey:UserID" json:"-"`

	// 访问信息
	IP        string `gorm:"not null;size:45" json:"ip"`     // IPv4 或 IPv6 地址
	Path      string `gorm:"not null;size:500" json:"path"`  // 访问路径
	Method    string `gorm:"not null;size:10" json:"method"` // HTTP 方法
	UserAgent string `gorm:"size:500" json:"user_agent"`     // 用户代理
	Referer   string `gorm:"size:500" json:"referer"`        // 来源页面

	// 设备信息（可选）
	DeviceType string `gorm:"size:20" json:"device_type"` // desktop, mobile, tablet
	OS         string `gorm:"size:50" json:"os"`          // 操作系统
	Browser    string `gorm:"size:50" json:"browser"`     // 浏览器

	// 地理位置信息（可选）
	Country string `gorm:"size:50" json:"country"`
	City    string `gorm:"size:50" json:"city"`

	// 响应信息
	StatusCode   int   `gorm:"not null" json:"status_code"`   // HTTP 状态码
	ResponseTime int64 `gorm:"not null" json:"response_time"` // 响应时间（毫秒）

	// 会话信息
	SessionID string `gorm:"size:100" json:"session_id"`
}

// TableName 指定表名
func (VisitLog) TableName() string {
	return "visit_logs"
}
