/*
Test Program

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package main

import (
	"fmt"
	"os"

	"resource-share-site/internal/config"
)

// 配置加载示例程序
func main() {
	fmt.Println("=== 配置加载测试 ===\n")

	// 方法1: 从默认配置文件加载
	fmt.Println("方法1: 从默认配置文件加载...")
	config1, err := config.LoadConfig(config.GetConfigFilePath())
	if err != nil {
		fmt.Printf("❌ 加载失败: %v\n", err)
	} else {
		fmt.Printf("✅ 加载成功\n")
		fmt.Printf("   应用: %s v%s\n", config1.App.Name, config1.App.Version)
		fmt.Printf("   环境: %s\n", config1.App.Environment)
		fmt.Printf("   服务器: %s\n", config1.GetServerAddress())
		fmt.Printf("   数据库: %s:%s\n", config1.Database.Type, config1.Database.Name)
		if config1.Redis != nil {
			fmt.Printf("   Redis: %s\n", config1.GetRedisAddr())
		} else {
			fmt.Printf("   Redis: 未配置\n")
		}
	}
	fmt.Println()

	// 方法2: 从环境变量加载
	fmt.Println("方法2: 从环境变量加载...")
	config2, err := config.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("❌ 加载失败: %v\n", err)
	} else {
		fmt.Printf("✅ 加载成功\n")
		fmt.Printf("   应用: %s v%s\n", config2.App.Name, config2.App.Version)
		fmt.Printf("   环境: %s\n", config2.App.Environment)
	}
	fmt.Println()

	// 方法3: 根据环境变量选择配置文件
	fmt.Println("方法3: 根据环境变量选择配置文件...")
	env := os.Getenv("RSS_APP_ENVIRONMENT")
	if env == "" {
		env = "default"
	}
	configFile := config.EnvironmentToConfigFile(env)
	fmt.Printf("   环境: %s\n", env)
	fmt.Printf("   配置文件: %s\n", configFile)

	config3, err := config.LoadConfigFromFileOrEnv("config/" + configFile)
	if err != nil {
		fmt.Printf("❌ 加载失败: %v\n", err)
	} else {
		fmt.Printf("✅ 加载成功\n")
		fmt.Printf("   数据库类型: %s\n", config3.Database.Type)
	}
	fmt.Println()

	// 演示配置验证
	fmt.Println("方法4: 演示配置验证...")
	testConfig := &config.AppConfig{
		Database: &config.DatabaseConfig{
			Type: "mysql",
			Host: "localhost",
			Port: "3306",
			Name: "test",
		},
		App: &config.AppSettings{
			Environment: "production",
			Port:        8080,
		},
	}

	if err := ValidateConfig(testConfig); err != nil {
		fmt.Printf("❌ 配置验证失败: %v\n", err)
	} else {
		fmt.Printf("✅ 配置验证通过\n")
	}
	fmt.Println()

	// 演示环境变量覆盖
	fmt.Println("方法5: 演示环境变量覆盖...")
	os.Setenv("RSS_APP_PORT", "9090")
	os.Setenv("RSS_DATABASE_TYPE", "postgres")

	config5, err := config.LoadConfigFromEnv()
	if err != nil {
		fmt.Printf("❌ 加载失败: %v\n", err)
	} else {
		fmt.Printf("✅ 加载成功\n")
		fmt.Printf("   应用端口(覆盖后): %d\n", config5.App.Port)
		fmt.Printf("   数据库类型(覆盖后): %s\n", config5.Database.Type)
	}
	fmt.Println()

	// 演示配置方法
	fmt.Println("方法6: 演示配置方法...")
	exampleConfig, err := config.LoadConfig(config.GetConfigFilePath())
	if err != nil {
		fmt.Printf("❌ 加载失败: %v\n", err)
	} else {
		fmt.Printf("✅ 配置方法测试\n")
		fmt.Printf("   是否开发环境: %v\n", exampleConfig.IsDevelopment())
		fmt.Printf("   是否生产环境: %v\n", exampleConfig.IsProduction())
		fmt.Printf("   服务器地址: %s\n", exampleConfig.GetServerAddress())
		if exampleConfig.Database.Type == "mysql" {
			fmt.Printf("   数据库DSN: %s\n", exampleConfig.GetDatabaseDSN())
		}
		if exampleConfig.Redis != nil {
			fmt.Printf("   Redis地址: %s\n", exampleConfig.GetRedisAddr())
		}
	}
	fmt.Println()

	// 配置优先级说明
	fmt.Println("=== 配置优先级 ===")
	fmt.Println("1. 环境变量 (最高优先级)")
	fmt.Println("2. 配置文件")
	fmt.Println("3. 默认值 (最低优先级)")
	fmt.Println()

	fmt.Println("=== 测试完成 ===")
}

// ValidateConfig 验证配置
func ValidateConfig(config *config.AppConfig) error {
	if config.App == nil {
		return fmt.Errorf("应用配置不能为空")
	}

	if config.App.Port <= 0 || config.App.Port > 65535 {
		return fmt.Errorf("端口号无效: %d", config.App.Port)
	}

	if config.Database == nil {
		return fmt.Errorf("数据库配置不能为空")
	}

	if config.Database.Type != "mysql" && config.Database.Type != "sqlite" {
		return fmt.Errorf("不支持的数据库类型: %s", config.Database.Type)
	}

	if config.Database.Type == "mysql" {
		if config.Database.Host == "" {
			return fmt.Errorf("数据库主机不能为空")
		}
		if config.Database.Port == "" {
			return fmt.Errorf("数据库端口不能为空")
		}
		if config.Database.Name == "" {
			return fmt.Errorf("数据库名称不能为空")
		}
	}

	return nil
}
