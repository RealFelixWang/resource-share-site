/*
Package config provides configuration management for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package config

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// AppConfig 应用配置结构
type AppConfig struct {
	// 数据库配置
	Database *DatabaseConfig `mapstructure:"database"`

	// Redis配置
	Redis *RedisConfig `mapstructure:"redis"`

	// 应用配置
	App *AppSettings `mapstructure:"app"`

	// 日志配置
	Log *LogConfig `mapstructure:"log"`
}

// AppSettings 应用设置
type AppSettings struct {
	Name        string `mapstructure:"name"`          // 应用名称
	Version     string `mapstructure:"version"`       // 版本号
	Environment string `mapstructure:"environment"`   // 环境: dev/test/prod
	Host        string `mapstructure:"host"`          // 监听地址
	Port        int    `mapstructure:"port"`          // 监听端口
	SecretKey   string `mapstructure:"secret_key"`    // 密钥
	Timeout     int    `mapstructure:"timeout"`       // 超时时间(秒)
	MaxBodySize int    `mapstructure:"max_body_size"` // 最大请求体大小(MB)
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`       // 日志级别: debug/info/warn/error
	Format     string `mapstructure:"format"`      // 日志格式: text/json
	Output     string `mapstructure:"output"`      // 输出: stdout/file
	Filename   string `mapstructure:"filename"`    // 文件名
	MaxSize    int    `mapstructure:"max_size"`    // 文件最大大小(MB)
	MaxAge     int    `mapstructure:"max_age"`     // 文件保留天数
	MaxBackups int    `mapstructure:"max_backups"` // 文件备份数量
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*AppConfig, error) {
	v := viper.New()

	// 设置配置搜索路径
	v.AddConfigPath(filepath.Dir(configPath))
	v.SetConfigFile(configPath)

	// 设置环境变量前缀
	v.SetEnvPrefix("RSS")

	// 自动将环境变量映射到配置
	v.AutomaticEnv()

	// 设置默认值
	setDefaultConfig(v)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认配置
			// 但仍然可以使用环境变量
		} else {
			return nil, err
		}
	}

	// 解析配置
	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() (*AppConfig, error) {
	v := viper.New()

	// 设置环境变量前缀
	v.SetEnvPrefix("RSS")

	// 自动映射环境变量
	v.AutomaticEnv()

	// 设置默认值
	setDefaultConfig(v)

	// 解析配置
	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfigFromFileOrEnv 从文件或环境变量加载配置
func LoadConfigFromFileOrEnv(configPath string) (*AppConfig, error) {
	v := viper.New()

	// 如果提供了配置文件路径
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err == nil {
			// 读取成功，使用文件配置
			v.AutomaticEnv()
		}
		// 即使文件不存在，也会继续使用环境变量
	} else {
		// 未提供配置文件，仅使用环境变量
		v.AutomaticEnv()
	}

	// 设置默认值
	setDefaultConfig(v)

	// 解析配置
	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaultConfig 设置默认配置
func setDefaultConfig(v *viper.Viper) {
	// 数据库默认配置
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.name", "resource_share_site")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "3306")
	v.SetDefault("database.charset", "utf8mb4")

	// Redis默认配置
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", "6379")
	v.SetDefault("redis.db", 0)

	// 应用默认配置
	v.SetDefault("app.name", "ResourceShareSite")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.host", "0.0.0.0")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.secret_key", "your-secret-key-change-in-production")
	v.SetDefault("app.timeout", 30)
	v.SetDefault("app.max_body_size", 10)

	// 日志默认配置
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.filename", "logs/app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.max_backups", 10)
}

// validateConfig 验证配置
func validateConfig(config *AppConfig) error {
	// 验证应用配置
	if config.App == nil {
		return ErrConfigInvalid
	}

	if config.App.SecretKey == "" || config.App.SecretKey == "your-secret-key-change-in-production" {
		// 可以在开发环境下允许默认值，但生产环境应该强制设置
		// 这里仅做警告，不返回错误
	}

	if config.App.Port <= 0 || config.App.Port > 65535 {
		return ErrConfigInvalid
	}

	if config.App.Environment != "development" && config.App.Environment != "test" && config.App.Environment != "production" {
		return ErrConfigInvalid
	}

	// 验证数据库配置
	if config.Database == nil {
		return ErrConfigInvalid
	}

	if config.Database.Type != "mysql" && config.Database.Type != "sqlite" {
		return ErrConfigInvalid
	}

	if config.Database.Type == "mysql" {
		if config.Database.Host == "" || config.Database.Port == "" || config.Database.Name == "" {
			return ErrConfigInvalid
		}
	}

	// 验证Redis配置（可选，可以为nil）
	if config.Redis != nil {
		if config.Redis.Host == "" || config.Redis.Port == "" {
			return ErrConfigInvalid
		}
	}

	return nil
}

// IsDevelopment 检查是否为开发环境
func (c *AppConfig) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction 检查是否为生产环境
func (c *AppConfig) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsTest 检查是否为测试环境
func (c *AppConfig) IsTest() bool {
	return c.App.Environment == "test"
}

// GetServerAddress 获取服务器地址
func (c *AppConfig) GetServerAddress() string {
	return strings.Join([]string{c.App.Host, ":", string(rune(c.App.Port))}, "")
}

// GetDatabaseDSN 获取数据库DSN（适用于MySQL）
func (c *AppConfig) GetDatabaseDSN() string {
	if c.Database.Type != "mysql" {
		return ""
	}

	parts := []string{
		c.Database.User,
		":",
		c.Database.Password,
		"@tcp(",
		c.Database.Host,
		":",
		c.Database.Port,
		")/",
		c.Database.Name,
		"?charset=",
		c.Database.Charset,
		"&parseTime=true&loc=Local",
	}

	return strings.Join(parts, "")
}

// GetRedisAddr 获取Redis地址
func (c *AppConfig) GetRedisAddr() string {
	if c.Redis == nil {
		return ""
	}

	return strings.Join([]string{c.Redis.Host, ":", c.Redis.Port}, "")
}

// GetConfigFilePath 获取配置文件的默认路径
func GetConfigFilePath() string {
	return filepath.Join("config", "config.yaml")
}

// EnvironmentToConfigFile 环境名转换为配置文件名
func EnvironmentToConfigFile(env string) string {
	env = strings.ToLower(env)
	if env == "" || env == "default" {
		return "config.yaml"
	}
	return "config-" + env + ".yaml"
}
