/*
Package model defines all data models for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package model

import (
	"time"
)

// IPBlacklist IP 黑名单模型
type IPBlacklist struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// IP 地址
	IP string `gorm:"uniqueIndex;not null;size:45" json:"ip"` // IPv4 或 IPv6 地址

	// 禁止信息
	Reason     string `gorm:"not null;size:255" json:"reason"`    // 禁止原因
	BannedByID uint   `gorm:"not null;index" json:"banned_by_id"` // 禁止者（管理员）
	BannedBy   *User  `gorm:"foreignKey:BannedByID" json:"banned_by"`

	// 时间信息
	BannedAt  time.Time  `gorm:"not null" json:"banned_at"`
	ExpiresAt *time.Time `gorm:"index" json:"expires_at"` // 过期时间，nil 表示永久禁止

	// 统计信息
	AccessCount  uint       `gorm:"default:0" json:"access_count"` // 访问次数
	LastAccessAt *time.Time `json:"last_access_at"`                // 最后访问时间
}

// TableName 指定表名
func (IPBlacklist) TableName() string {
	return "ip_blacklists"
}
