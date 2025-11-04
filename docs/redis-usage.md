# Redis 使用指南

## 概述

Redis 是一个开源的内存数据库，可用作数据库、缓存和消息代理。资源分享平台使用 Redis 来存储会话、管理缓存、实现限流等场景。

## 目录

- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [连接管理](#连接管理)
- [键命名规范](#键命名规范)
- [使用示例](#使用示例)
- [最佳实践](#最佳实践)
- [性能优化](#性能优化)
- [监控维护](#监控维护)

## 快速开始

### 1. 基本用法

```go
package main

import (
    "context"
    "log"
    "time"
    
    "resource-share-site/internal/config"
)

func main() {
    // 配置Redis
    redisConfig := &config.RedisConfig{
        Host:     "localhost",
        Port:     "6379",
        Password: "",
        DB:       0,
    }
    
    // 连接Redis
    client, err := config.InitRedisClient(redisConfig)
    if err != nil {
        log.Fatal(err)
    }
    defer config.CloseRedisClient(client)
    
    ctx := context.Background()
    
    // 基本操作
    err = client.Set(ctx, "key", "value", 10*time.Second).Err()
    if err != nil {
        log.Fatal(err)
    }
    
    value, err := client.Get(ctx, "key").Result()
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Value:", value)
}
```

### 2. 使用缓存接口

```go
// 创建缓存客户端
cache := config.NewRedisClient(client)

// 设置缓存
err := cache.Set(ctx, "user:100", `{"id":100,"username":"admin"}`, 5*time.Minute)
if err != nil {
    log.Fatal(err)
}

// 获取缓存
value, err := cache.Get(ctx, "user:100")
if err != nil {
    log.Fatal(err)
}
log.Println("User:", value)
```

## 配置说明

### RedisConfig 结构

```go
type RedisConfig struct {
    Host     string `mapstructure:"host"`     // 主机地址 (默认: localhost)
    Port     string `mapstructure:"port"`     // 端口 (默认: 6379)
    Password string `mapstructure:"password"` // 密码 (默认: 空)
    DB       int    `mapstructure:"db"`       // 数据库ID (默认: 0)
}
```

### 配置参数详解

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| Host | string | localhost | Redis服务器地址 |
| Port | string | 6379 | Redis服务端口 |
| Password | string | "" | Redis密码（可选） |
| DB | int | 0 | 数据库编号（0-15） |

### 连接池配置

```go
client := redis.NewClient(&redis.Options{
    Addr:               "localhost:6379",
    Password:           "",
    DB:                 0,
    
    // 连接池大小
    PoolSize:     10,    // 最大连接数
    MinIdleConns: 5,     // 最小空闲连接数
    
    // 超时配置
    DialTimeout:  5 * time.Second,  // 连接超时
    ReadTimeout:  3 * time.Second,  // 读超时
    WriteTimeout: 3 * time.Second,  // 写超时
    
    // 空闲连接管理
    IdleCheckFrequency: 60 * time.Second,  // 空闲检查频率
    IdleTimeout:        5 * time.Minute,   // 空闲超时
    
    // 重试配置
    MaxRetries:        3,        // 最大重试次数
    MaxRetryBackoff:   8 * time.Second,  // 最大重试间隔
    MinRetryBackoff:   100 * time.Millisecond,  // 最小重试间隔
})
```

## 连接管理

### 1. 单例模式

```go
// 推荐方式：使用全局变量管理Redis连接
var redisClient *redis.Client

func InitRedis() {
    config := &config.RedisConfig{
        Host:     "localhost",
        Port:     "6379",
        Password: "",
        DB:       0,
    }
    
    var err error
    redisClient, err = config.InitRedisClient(config)
    if err != nil {
        panic(err)
    }
}

func GetRedisClient() *redis.Client {
    if redisClient == nil {
        panic("Redis未初始化")
    }
    return redisClient
}
```

### 2. 依赖注入

```go
// 创建结构体持有Redis客户端
type UserService struct {
    redisClient *redis.Client
    dbClient    *gorm.DB
}

func NewUserService(redisClient *redis.Client, dbClient *gorm.DB) *UserService {
    return &UserService{
        redisClient: redisClient,
        dbClient:    dbClient,
    }
}
```

### 3. 连接池监控

```go
stats := client.PoolStats()
fmt.Printf("连接池状态:\n")
fmt.Printf("  总连接数: %d\n", stats.TotalConns)
fmt.Printf("  空闲连接数: %d\n", stats.IdleConns)
fmt.Printf("  命中次数: %d\n", stats.Hits)
fmt.Printf("  未命中次数: %d\n", stats.Misses)
fmt.Printf("  超时次数: %d\n", stats.Timeouts)
```

## 键命名规范

### 命名约定

```
rss:{模块}:{子模块}:{标识}
```

### 常用键前缀

| 前缀 | 用途 | 示例 |
|------|------|------|
| `rss:session` | 会话管理 | `rss:session:abc123` |
| `rss:user:info` | 用户信息缓存 | `rss:user:info:100` |
| `rss:resource` | 资源缓存 | `rss:resource:200` |
| `rss:category` | 分类缓存 | `rss:categories` |
| `rss:visit` | 访问记录 | `rss:visit:2025-10-31` |
| `rss:hot:resources` | 热门资源 | `rss:hot:resources` |
| `rss:ratelimit` | 限流 | `rss:ratelimit:ip:192.168.1.1` |
| `rss:cache` | 通用缓存 | `rss:cache:dashboard:stats` |

### 使用键构建器

```go
// 创建键构建器
kb := config.NewRedisKeyBuilder("rss")

// 构建各种键
sessionKey := kb.BuildUserSessionKey("abc123")  // rss:session:abc123
userKey := kb.BuildUserInfoKey(100)             // rss:user:info:100
resourceKey := kb.BuildResourceCacheKey(200)    // rss:resource:200
categoryKey := kb.BuildCategoryCacheKey()       // rss:categories
hotKey := kb.BuildHotResourcesKey()             // rss:hot:resources
visitKey := kb.BuildVisitCountKey("2025-10-31") // rss:visit:count:2025-10-31
```

## 使用示例

### 1. 用户会话管理

```go
func SetSession(ctx context.Context, client *redis.Client, sessionID, userID string, expire time.Duration) error {
    kb := config.NewRedisKeyBuilder("rss")
    key := kb.BuildUserSessionKey(sessionID)
    
    return client.Set(ctx, key, userID, expire).Err()
}

func GetSession(ctx context.Context, client *redis.Client, sessionID string) (string, error) {
    kb := config.NewRedisKeyBuilder("rss")
    key := kb.BuildUserSessionKey(sessionID)
    
    return client.Get(ctx, key).Result()
}

func DeleteSession(ctx context.Context, client *redis.Client, sessionID string) error {
    kb := config.NewRedisKeyBuilder("rss")
    key := kb.BuildUserSessionKey(sessionID)
    
    return client.Del(ctx, key).Err()
}
```

### 2. 用户信息缓存

```go
func GetUserFromCache(ctx context.Context, client *redis.Client, userID uint) (*User, error) {
    kb := config.NewRedisKeyBuilder("rss")
    key := kb.BuildUserInfoKey(userID)
    
    value, err := client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, nil // 缓存不存在
    }
    if err != nil {
        return nil, err
    }
    
    var user User
    if err := json.Unmarshal([]byte(value), &user); err != nil {
        return nil, err
    }
    
    return &user, nil
}

func SetUserToCache(ctx context.Context, client *redis.Client, user *User, expire time.Duration) error {
    kb := config.NewRedisKeyBuilder("rss")
    key := kb.BuildUserInfoKey(user.ID)
    
    data, err := json.Marshal(user)
    if err != nil {
        return err
    }
    
    return client.Set(ctx, key, data, expire).Err()
}
```

### 3. 资源访问计数

```go
func IncResourceView(ctx context.Context, client *redis.Client, resourceID uint) error {
    kb := config.NewRedisKeyBuilder("rss")
    key := kb.BuildResourceCacheKey(resourceID)
    
    return client.Incr(ctx, key).Err()
}

func GetResourceViewCount(ctx context.Context, client *redis.Client, resourceID uint) (int64, error) {
    kb := config.NewRedisKeyBuilder("rss")
    key := kb.BuildResourceCacheKey(resourceID)
    
    return client.Get(ctx, key).Int64()
}
```

### 4. 热门资源排行

```go
func AddHotResource(ctx context.Context, client *redis.Client, resourceID uint, score float64) error {
    kb := config.NewRedisKeyBuilder("rss")
    key := kb.BuildHotResourcesKey()
    
    return client.ZAdd(ctx, key, &redis.Z{
        Score:  score,
        Member: fmt.Sprintf("%d", resourceID),
    }).Err()
}

func GetTopHotResources(ctx context.Context, client *redis.Client, limit int) ([]uint, error) {
    kb := config.NewRedisKeyBuilder("rss")
    key := kb.BuildHotResourcesKey()
    
    members, err := client.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
    if err != nil {
        return nil, err
    }
    
    var resourceIDs []uint
    for _, member := range members {
        if id, err := strconv.ParseUint(member, 10, 32); err == nil {
            resourceIDs = append(resourceIDs, uint(id))
        }
    }
    
    return resourceIDs, nil
}
```

### 5. IP限流

```go
func IsAllowed(ctx context.Context, client *redis.Client, ip string, limit int, window time.Duration) (bool, error) {
    kb := config.NewRedisKeyBuilder("rss")
    key := fmt.Sprintf("%s:ip:%s", kb.prefix, ip)
    
    pipe := client.Pipeline()
    
    // 当前计数
    pipe.Incr(ctx, key)
    // 设置过期
    pipe.Expire(ctx, key, window)
    
    results, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }
    
    count := results[0].(*redis.IntCmd).Val()
    return count <= int64(limit), nil
}
```

### 6. 缓存中间件

```go
func CacheMiddleware(client *redis.Client, expire time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.Request.URL.Path + "?" + c.Request.URL.RawQuery
        
        ctx := c.Request.Context()
        value, err := client.Get(ctx, key).Result()
        if err == nil {
            c.Data(200, "application/json", []byte(value))
            c.Abort()
            return
        }
        
        // 继续处理请求
        c.Next()
        
        // 缓存响应
        if c.Writer.Status() == 200 {
            response, _ := c.Get("response")
            if data, ok := response.(string); ok {
                client.Set(ctx, key, data, expire)
            }
        }
    }
}
```

### 7. 管道批量操作

```go
func BatchGetUsers(ctx context.Context, client *redis.Client, userIDs []uint) (map[uint]*User, error) {
    kb := config.NewRedisKeyBuilder("rss")
    pipe := client.Pipeline()
    
    keys := make([]string, len(userIDs))
    for i, id := range userIDs {
        keys[i] = kb.BuildUserInfoKey(id)
        pipe.Get(ctx, keys[i])
    }
    
    cmds, err := pipe.Exec(ctx)
    if err != nil {
        return nil, err
    }
    
    users := make(map[uint]*User)
    for i, cmd := range cmds {
        value, err := cmd.(*redis.StringCmd).Result()
        if err == nil {
            var user User
            if err := json.Unmarshal([]byte(value), &user); err == nil {
                users[userIDs[i]] = &user
            }
        }
    }
    
    return users, nil
}
```

## 最佳实践

### 1. 错误处理

```go
// 正确处理redis.Nil错误
value, err := client.Get(ctx, "key").Result()
if err == redis.Nil {
    // 键不存在，这是正常情况
    log.Println("键不存在")
    return nil, nil
}
if err != nil {
    // 其他错误
    return nil, err
}
```

### 2. 连接复用

```go
// 不要每次操作都创建新连接
var client *redis.Client

func Init() {
    client = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
}

// 使用全局客户端
func GetUser(ctx context.Context, userID uint) (*User, error) {
    // 使用全局client
    return getUserFromCache(ctx, client, userID)
}
```

### 3. 键过期策略

```go
// 1. 会话：设置明确的过期时间
client.Set(ctx, "session:abc", userID, 24*time.Hour)

// 2. 缓存：使用合理的过期时间
client.Set(ctx, "user:100", userData, 5*time.Minute)

// 3. 计数器：不设置过期（使用业务逻辑控制）
client.Incr(ctx, "visit:count:2025-10-31")
```

### 4. 内存优化

```go
// 1. 定期清理无用数据
client.Del(ctx, "old:key")

// 2. 使用哈希存储对象
client.HSet(ctx, "user:100", "username", "admin")
client.HSet(ctx, "user:100", "email", "admin@example.com")

// 3. 避免大对象
// 好的做法：将大对象拆分成多个键
for i := 0; i < 10; i++ {
    client.Set(ctx, fmt.Sprintf("batch:%d", i), data[i], 0)
}
```

### 5. 性能建议

```go
// 1. 使用管道批量操作
pipe := client.Pipeline()
for _, key := range keys {
    pipe.Get(ctx, key)
}
cmds, _ := pipe.Exec(ctx)

// 2. 使用批量接口
client.MGet(ctx, "key1", "key2", "key3")

// 3. 合理设置连接池大小
// CPU核心数 * 2 通常是一个好的起点
```

## 性能优化

### 1. 连接池调优

```go
client := redis.NewClient(&redis.Options{
    PoolSize:     100,           // 根据QPS调整
    MinIdleConns: 20,            // 根据峰值调整
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    IdleTimeout:  5 * time.Minute,
})
```

### 2. 内存优化

```go
// 1. 使用Redis内存分析命令
// INFO memory
// INFO stats

// 2. 定期清理过期键
// CONFIG SET maxmemory-policy allkeys-lru

// 3. 使用Redis集群分散负载
// cluster enabled yes
```

### 3. 持久化策略

```go
// 1. RDB快照（适合读多写少）
save 900 1
save 300 10
save 60 10000

// 2. AOF日志（适合数据安全性要求高）
appendonly yes
appendfsync everysec
```

### 4. 缓存预热

```go
func WarmupCache(client *redis.Client) error {
    ctx := context.Background()
    
    // 预热热门数据
    resources, _ := db.FindHotResources(100)
    for _, resource := range resources {
        SetResourceToCache(ctx, client, resource, 30*time.Minute)
    }
    
    // 预热分类数据
    categories, _ := db.FindAllCategories()
    SetCategoriesToCache(ctx, client, categories, 60*time.Minute)
    
    return nil
}
```

## 监控维护

### 1. 健康检查

```go
func HealthCheck(client *redis.Client) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    pong, err := client.Ping(ctx).Result()
    if err != nil {
        return fmt.Errorf("Redis Ping失败: %w", err)
    }
    
    if pong != "PONG" {
        return fmt.Errorf("Redis响应异常: %s", pong)
    }
    
    return nil
}
```

### 2. 性能监控

```go
// 获取Redis信息
info := client.Info(ctx)
fmt.Println("Redis Info:", info)

// 监控连接数
infoMap := client.Info(ctx, "clients").String()
fmt.Println("客户端连接:", infoMap)

// 监控内存使用
memInfo := client.Info(ctx, "memory").String()
fmt.Println("内存使用:", memInfo)

// 监控慢查询
slowLog := client.SlowlogGet(ctx, 10)
fmt.Println("慢查询:", slowLog)
```

### 3. 数据统计

```go
// 键数量统计
dbsize, err := client.DBSize(ctx).Result()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("数据库大小: %d 个键\n", dbsize)

// 键过期统计
expiredKeys, err := client.Info(ctx, "keyspace").Result()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("键空间信息: %s\n", expiredKeys)

// 命中率统计
stats := client.PoolStats()
hitRate := float64(stats.Hits) / float64(stats.Hits+stats.Misses) * 100
fmt.Printf("缓存命中率: %.2f%%\n", hitRate)
```

### 4. 备份恢复

```bash
# 备份（RDB快照）
redis-cli BGSAVE

# 恢复
# 1. 停止Redis
# 2. 复制dump.rdb文件
# 3. 重启Redis

# 使用AOF恢复
redis-cli --pipe < appendonly.aof
```

### 5. 运维脚本

```bash
#!/bin/bash
# redis-health-check.sh

REDIS_CLI="redis-cli -h localhost -p 6379"

# 检查连接
if ! $REDIS_CLI ping > /dev/null 2>&1; then
    echo "Redis连接失败"
    exit 1
fi

# 检查内存使用
MEMORY_USAGE=$($REDIS_CLI info memory | grep used_memory_human | cut -d: -f2 | tr -d '\r')
echo "内存使用: $MEMORY_USAGE"

# 检查键数量
KEY_COUNT=$($REDIS_CLI dbsize)
echo "键数量: $KEY_COUNT"

# 检查慢查询
SLOW_COUNT=$($REDIS_CLI slowlog len)
if [ $SLOW_COUNT -gt 10 ]; then
    echo "慢查询数量过多: $SLOW_COUNT"
fi

echo "Redis健康检查完成"
```

## 常见问题

### Q1: 连接超时

**A**: 检查网络和Redis服务状态

```go
// 增加超时时间
client := redis.NewClient(&redis.Options{
    DialTimeout: 10 * time.Second,  // 增加连接超时
    ReadTimeout: 5 * time.Second,   // 增加读超时
})
```

### Q2: 内存使用过高

**A**: 检查键过期策略和数据结构

```bash
# 查看内存使用详情
redis-cli info memory

# 查看大键
redis-cli --bigkeys

# 查看键空间分布
redis-cli --scan --pattern '*' | head -1000 | redis-cli --pipe
```

### Q3: 性能下降

**A**: 优化连接池和查询模式

```go
// 增加连接池大小
client := redis.NewClient(&redis.Options{
    PoolSize: 50,  // 根据负载调整
})

// 使用管道批量操作
pipe := client.Pipeline()
// ... 批量操作
```

### Q4: 数据丢失

**A**: 检查持久化配置

```bash
# 检查RDB配置
redis-cli config get save

# 检查AOF配置
redis-cli config get appendonly
```

## 相关文档

- [Redis官方文档](https://redis.io/documentation)
- [go-redis文档](https://pkg.go.dev/github.com/go-redis/redis/v8)
- [数据库设计文档](database-architecture.md)
- [缓存策略指南](cache-strategy.md)

---

**维护者**: Felix Wang  
**邮箱**: felixwang.biz@gmail.com  
**最后更新**: 2025-10-31
