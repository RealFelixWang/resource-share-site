/*
Package model defines all data models for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package model

import (
	"time"
)

// Permission 权限配置模型
type Permission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Key         string `gorm:"uniqueIndex;not null;size:50" json:"key"` // 权限键名
	Name        string `gorm:"not null;size:100" json:"name"`           // 权限名称
	Description string `gorm:"size:255" json:"description"`             // 权限描述
	IsEnabled   bool   `gorm:"default:false" json:"is_enabled"`         // 是否启用
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// Session 会话模型
type Session struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID    uint      `gorm:"not null;index" json:"user_id"`
	SessionID string    `gorm:"uniqueIndex;not null;size:100" json:"session_id"`
	Data      string    `gorm:"type:text" json:"data"`   // 会话数据（JSON 格式）
	ExpiresAt time.Time `gorm:"index" json:"expires_at"` // 过期时间
	IP        string    `gorm:"size:45" json:"ip"`       // IP 地址
}

// TableName 指定表名
func (Session) TableName() string {
	return "sessions"
}

// AdminLog 管理员操作日志模型
type AdminLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	AdminID uint  `gorm:"not null;index" json:"admin_id"` // 管理员 ID
	Admin   *User `gorm:"foreignKey:AdminID" json:"admin"`

	Action     string `gorm:"not null;size:50" json:"action"`      // 操作类型
	TargetType string `gorm:"not null;size:50" json:"target_type"` // 目标类型
	TargetID   uint   `gorm:"not null" json:"target_id"`           // 目标 ID

	// 详细信息
	BeforeData string `gorm:"type:text" json:"before_data"` // 操作前数据
	AfterData  string `gorm:"type:text" json:"after_data"`  // 操作后数据
	IP         string `gorm:"size:45" json:"ip"`            // IP 地址
	UserAgent  string `gorm:"size:500" json:"user_agent"`   // 用户代理
}

// TableName 指定表名
func (AdminLog) TableName() string {
	return "admin_logs"
}

// ImportTask 导入任务模型
type ImportTask struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	TaskType string `gorm:"not null;size:30" json:"task_type"` // 任务类型：crawler, excel
	Status   string `gorm:"not null;size:20" json:"status"`    // 状态：pending, running, completed, failed

	// 统计信息
	TotalCount   uint `gorm:"default:0" json:"total_count"`   // 总数
	SuccessCount uint `gorm:"default:0" json:"success_count"` // 成功数
	FailCount    uint `gorm:"default:0" json:"fail_count"`    // 失败数

	// 详细信息
	ConfigData string `gorm:"type:text" json:"config_data"` // 配置信息（JSON）
	ErrorLog   string `gorm:"type:text" json:"error_log"`   // 错误日志

	// 执行信息
	StartedAt   *time.Time `json:"started_at"`   // 开始时间
	CompletedAt *time.Time `json:"completed_at"` // 完成时间
}

// TableName 指定表名
func (ImportTask) TableName() string {
	return "import_tasks"
}
