/*
Package model defines all data models for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package model

import (
	"time"
)

// PointsRule 积分规则模型
type PointsRule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	RuleKey     string `gorm:"uniqueIndex;not null;size:50" json:"rule_key"` // 例如：invite_reward
	RuleName    string `gorm:"not null;size:100" json:"rule_name"`           // 例如：邀请奖励
	Description string `gorm:"size:255" json:"description"`
	Points      int    `gorm:"not null" json:"points"`         // 积分数量
	IsEnabled   bool   `gorm:"default:true" json:"is_enabled"` // 是否启用
}

// TableName 指定表名
func (PointsRule) TableName() string {
	return "points_rules"
}
