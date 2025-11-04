/*
Package config provides configuration management for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package config

import (
	"resource-share-site/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	Type     string `mapstructure:"type" json:"type"`
	Host     string `mapstructure:"host" json:"host"`
	Port     string `mapstructure:"port" json:"port"`
	Name     string `mapstructure:"name" json:"name"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
	Charset  string `mapstructure:"charset" json:"charset"`
}

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	// 根据数据库类型配置连接
	switch cfg.Type {
	case "mysql":
		dsn := cfg.User + ":" + cfg.Password + "@tcp(" + cfg.Host + ":" + cfg.Port + ")/" + cfg.Name + "?charset=" + cfg.Charset + "&parseTime=true&loc=Local"
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.Name + ".db")
	default:
		return nil, ErrUnsupportedDBType
	}

	// 打开数据库连接
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	// 自动迁移数据库表结构
	if err := AutoMigrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	// 注册所有数据模型
	return db.AutoMigrate(
		// 用户系统
		&model.User{},
		&model.Session{},

		// 分类系统
		&model.Category{},

		// 资源系统
		&model.Resource{},
		&model.Comment{},

		// 邀请系统
		&model.Invitation{},

		// 积分系统
		&model.PointsRule{},
		&model.PointRecord{},

		// 监控审计
		&model.VisitLog{},
		&model.IPBlacklist{},
		&model.AdminLog{},

		// 系统管理
		&model.Ad{},
		&model.Permission{},
		&model.ImportTask{},
	)
}
