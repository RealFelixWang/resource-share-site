# 配置管理指南

## 概述

资源分享平台使用 Viper 进行配置管理，支持多种配置源：配置文件、环境变量、默认值。配置加载遵循优先级：环境变量 > 配置文件 > 默认值。

## 配置结构

### 1. 配置文件结构

```yaml
app:
  name: "ResourceShareSite"
  version: "1.0.0"
  environment: "development"
  host: "0.0.0.0"
  port: 8080
  secret_key: "your-secret-key"
  timeout: 30
  max_body_size: 10

database:
  type: "sqlite"  # mysql/sqlite
  host: "localhost"
  port: "3306"
  name: "resource_share_site"
  user: "resource_share"
  password: "password"
  charset: "utf8mb4"

redis:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0

log:
  level: "info"
  format: "text"
  output: "stdout"
  filename: "logs/app.log"
  max_size: 100
  max_age: 30
  max_backups: 10
```

### 2. 环境变量映射

所有配置都可以通过环境变量设置，以 `RSS_` 开头：

| 配置项 | 环境变量名 |
|--------|-----------|
| app.environment | RSS_APP_ENVIRONMENT |
| app.port | RSS_APP_PORT |
| app.secret_key | RSS_APP_SECRET_KEY |
| database.type | RSS_DATABASE_TYPE |
| database.host | RSS_DATABASE_HOST |
| database.port | RSS_DATABASE_PORT |
| database.name | RSS_DATABASE_NAME |
| redis.host | RSS_REDIS_HOST |
| redis.port | RSS_REDIS_PORT |
| log.level | RSS_LOG_LEVEL |

## 使用方法

### 1. 加载配置

```go
package main

import (
    "log"
    
    "resource-share-site/internal/config"
)

func main() {
    // 方法1: 从文件加载
    cfg, err := config.LoadConfig("config/config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // 方法2: 从环境变量加载
    cfg, err := config.LoadConfigFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    
    // 方法3: 优先从文件，加载失败使用环境变量
    cfg, err := config.LoadConfigFromFileOrEnv("config/config.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // 使用配置
    fmt.Printf("服务器地址: %s\n", cfg.GetServerAddress())
}
```

### 2. 不同环境的配置

```go
// 开发环境
export RSS_APP_ENVIRONMENT=development
export RSS_DATABASE_TYPE=sqlite
export RSS_DATABASE_NAME=resource_share_site_dev
cfg, _ := config.LoadConfigFromEnv()

// 测试环境
export RSS_APP_ENVIRONMENT=test
export RSS_DATABASE_TYPE=sqlite
export RSS_DATABASE_NAME=resource_share_site_test
cfg, _ := config.LoadConfigFromEnv()

// 生产环境
export RSS_APP_ENVIRONMENT=production
export RSS_DATABASE_TYPE=mysql
export RSS_DATABASE_HOST=prod-db.example.com
export RSS_DATABASE_NAME=resource_share_site
export RSS_APP_SECRET_KEY=your-production-secret-key
cfg, _ := config.LoadConfigFromEnv()
```

### 3. 在应用中使用

```go
type App struct {
    config *config.AppConfig
    db     *gorm.DB
    redis  *redis.Client
}

func NewApp(cfg *config.AppConfig) (*App, error) {
    // 初始化数据库
    db, err := config.InitDatabase(cfg.Database)
    if err != nil {
        return nil, err
    }
    
    // 初始化Redis(可选)
    var redisClient *redis.Client
    if cfg.Redis != nil {
        redisClient, err = config.InitRedisClient(cfg.Redis)
        if err != nil {
            return nil, err
        }
    }
    
    return &App{
        config: cfg,
        db:     db,
        redis:  redisClient,
    }, nil
}

func (app *App) Run() error {
    // 根据环境配置日志
    if app.config.IsProduction() {
        // 生产环境: 输出到文件，使用JSON格式
        setupFileLogger(app.config.Log)
    } else {
        // 开发/测试环境: 输出到控制台，使用文本格式
        setupConsoleLogger(app.config.Log)
    }
    
    // 启动服务器
    r := gin.Default()
    r.Run(app.config.GetServerAddress())
    
    return nil
}
```

## 配置验证

### 1. 自动验证

加载配置时自动验证必要字段：

```go
cfg, err := config.LoadConfig("config/config.yaml")
if err != nil {
    log.Fatal(err)
}
// 配置已验证
```

### 2. 手动验证

```go
func validateCustomConfig(cfg *config.AppConfig) error {
    // 检查必填字段
    if cfg.App.SecretKey == "" {
        return errors.New("secret_key不能为空")
    }
    
    // 检查端口范围
    if cfg.App.Port < 1024 || cfg.App.Port > 65535 {
        return errors.New("端口号必须在1024-65535之间")
    }
    
    // 检查数据库密码强度
    if cfg.Database.Type == "mysql" && len(cfg.Database.Password) < 8 {
        return errors.New("数据库密码长度至少8位")
    }
    
    return nil
}
```

### 3. 生产环境特殊验证

```go
func validateProductionConfig(cfg *config.AppConfig) error {
    if cfg.IsProduction() {
        // 检查是否使用了默认密钥
        if cfg.App.SecretKey == "your-secret-key-change-in-production" {
            return errors.New("生产环境必须修改默认密钥")
        }
        
        // 检查数据库密码
        if cfg.Database.Password == "password" || cfg.Database.Password == "123456" {
            return errors.New("生产环境必须使用强密码")
        }
        
        // 检查Redis配置
        if cfg.Redis == nil {
            return errors.New("生产环境建议启用Redis缓存")
        }
    }
    
    return nil
}
```

## 环境变量管理

### 1. 使用.env文件

```bash
# 创建 .env 文件
cp .env.example .env

# 编辑 .env 文件
vim .env
```

### 2. 加载.env文件(需要安装gopkg.in/ini.v1)

```go
import (
    "github.com/spf13/viper"
    "github.com/go-ini/ini"
)

func LoadEnvFile(filename string) error {
    // 读取.env文件
    cfg, err := ini.Load(filename)
    if err != nil {
        return err
    }
    
    // 转换为环境变量
    for _, section := range cfg.Sections() {
        for _, key := range section.Keys() {
            os.Setenv(key.Name(), key.Value())
        }
    }
    
    return nil
}
```

### 3. 使用direnv管理环境变量

```bash
# 安装direnv
# .envrc文件
export RSS_APP_ENVIRONMENT=development
export RSS_DATABASE_TYPE=sqlite
```

## 配置加密

### 1. 敏感信息加密

```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
)

func encryptPassword(password string, key []byte) (string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}
```

### 2. 从环境变量解密

```go
func loadEncryptedConfig() (*config.AppConfig, error) {
    cfg, err := config.LoadConfigFromEnv()
    if err != nil {
        return nil, err
    }
    
    // 解密数据库密码
    encryptedPassword := os.Getenv("RSS_DATABASE_PASSWORD_ENCRYPTED")
    if encryptedPassword != "" {
        password, err := decryptPassword(encryptedPassword, getEncryptionKey())
        if err != nil {
            return nil, err
        }
        cfg.Database.Password = password
    }
    
    return cfg, nil
}
```

## 热加载配置

### 1. 监听配置文件变化

```go
func watchConfig(filename string, onChange func(*config.AppConfig)) {
    viper.WatchConfig()
    viper.OnConfigFileChanged(func(e fsnotify.Event) {
        fmt.Println("配置文件已更改，重新加载...")
        cfg, err := config.LoadConfig(filename)
        if err != nil {
            fmt.Printf("重新加载失败: %v\n", err)
            return
        }
        onChange(cfg)
    })
}
```

### 2. 动态更新配置

```go
type ConfigManager struct {
    config  *config.AppConfig
    mutex   sync.RWMutex
    watchers []func(*config.AppConfig)
}

func (cm *ConfigManager) UpdateConfig(newConfig *config.AppConfig) {
    cm.mutex.Lock()
    defer cm.mutex.Unlock()
    
    cm.config = newConfig
    
    // 通知所有观察者
    for _, watcher := range cm.watchers {
        watcher(newConfig)
    }
}

func (cm *ConfigManager) GetConfig() *config.AppConfig {
    cm.mutex.RLock()
    defer cm.mutex.RUnlock()
    return cm.config
}
```

## 最佳实践

### 1. 配置文件组织

```
config/
├── config.yaml          # 默认配置
├── config-development.yaml  # 开发环境
├── config-test.yaml     # 测试环境
└── config-production.yaml  # 生产环境
```

### 2. 环境变量命名规范

- 使用 `RSS_` 前缀
- 使用大写字母和下划线
- 按模块分组：`RSS_APP_*`, `RSS_DATABASE_*`, `RSS_REDIS_*`

### 3. 默认值设置

```go
// 设置合理的默认值
v.SetDefault("app.port", 8080)
v.SetDefault("app.timeout", 30)
v.SetDefault("database.type", "sqlite")

// 敏感信息不设置默认值，强制用户提供
// v.SetDefault("app.secret_key", "")  // 错误!
```

### 4. 配置文档化

```yaml
# 配置文件
app:
  port: 8080  # 服务监听端口 (1-65535)
  timeout: 30 # 请求超时时间(秒)
```

### 5. 安全建议

1. **敏感信息**：
   - 生产环境使用强密码
   - 使用密钥管理系统(如Vault)
   - 不要将密码写在配置文件中

2. **权限控制**：
   - 配置文件权限设为600
   - 限制访问配置文件的用户

3. **加密传输**：
   - 使用HTTPS传输配置
   - 加密存储敏感配置

## 测试配置

### 1. 单元测试

```go
func TestLoadConfig(t *testing.T) {
    // 创建临时配置文件
    tempFile := "test_config.yaml"
    defer os.Remove(tempFile)
    
    content := `
app:
  port: 8080
database:
  type: sqlite
  name: test
`
    os.WriteFile(tempFile, []byte(content), 0644)
    
    // 加载配置
    cfg, err := config.LoadConfig(tempFile)
    assert.NoError(t, err)
    assert.Equal(t, 8080, cfg.App.Port)
    assert.Equal(t, "sqlite", cfg.Database.Type)
}
```

### 2. 配置测试

```go
func TestConfigValidation(t *testing.T) {
    // 测试无效配置
    invalidConfig := &config.AppConfig{
        App: &config.AppSettings{
            Port: 99999, // 无效端口
        },
    }
    
    err := validateConfig(invalidConfig)
    assert.Error(t, err)
}
```

## 故障排除

### Q1: 配置未加载

**A**: 检查文件路径和权限

```go
// 检查文件是否存在
if _, err := os.Stat("config/config.yaml"); os.IsNotExist(err) {
    log.Fatal("配置文件不存在")
}

// 检查环境变量
fmt.Printf("环境变量: %v\n", os.Environ())
```

### Q2: 环境变量未生效

**A**: 检查变量名和前缀

```go
// 确保以RSS_开头
// RSS_APP_PORT 而不是 APP_PORT

// 检查变量值
fmt.Printf("RSS_APP_PORT=%s\n", os.Getenv("RSS_APP_PORT"))
```

### Q3: 配置验证失败

**A**: 逐步验证配置

```go
// 打印配置详情
fmt.Printf("应用配置: %+v\n", cfg.App)
fmt.Printf("数据库配置: %+v\n", cfg.Database)
```

## 相关资源

- [Viper文档](https://github.com/spf13/viper)
- [环境变量最佳实践](https://12factor.net/config)
- [配置文件设计模式](https://github.com/spf13/viper#working-with-multiple-vipers)

---

**维护者**: Felix Wang  
**邮箱**: felixwang.biz@gmail.com  
**最后更新**: 2025-10-31
