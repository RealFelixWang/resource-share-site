# 配置快速参考

## 配置文件

| 文件 | 说明 |
|------|------|
| `config.yaml` | 默认配置 |
| `config-development.yaml` | 开发环境 |
| `config-test.yaml` | 测试环境 |
| `config-production.yaml` | 生产环境 |
| `.env.example` | 环境变量示例 |

## 环境变量

### 前缀: `RSS_`

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `RSS_APP_ENVIRONMENT` | development | 环境 |
| `RSS_APP_PORT` | 8080 | 端口 |
| `RSS_DATABASE_TYPE` | sqlite | 数据库类型 |
| `RSS_REDIS_HOST` | localhost | Redis地址 |

## 快速使用

```go
// 加载配置
cfg, err := config.LoadConfig("config/config.yaml")
if err != nil {
    log.Fatal(err)
}

// 使用配置
fmt.Printf("服务器: %s\n", cfg.GetServerAddress())
fmt.Printf("环境: %s\n", cfg.App.Environment)
```

## 配置优先级

1. 环境变量
2. 配置文件
3. 默认值

## 环境示例

```bash
# 开发
export RSS_APP_ENVIRONMENT=development
export RSS_DATABASE_TYPE=sqlite

# 生产
export RSS_APP_ENVIRONMENT=production
export RSS_DATABASE_TYPE=mysql
export RSS_APP_SECRET_KEY=random-secret
```

---

**参考**: [完整配置文档](configuration-guide.md)
