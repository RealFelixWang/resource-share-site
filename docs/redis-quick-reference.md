# Redis 快速参考

## 概述

Redis 是一个开源的内存数据库，用于缓存、会话管理、排行榜等场景。

## 配置参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| Host | localhost | Redis服务器地址 |
| Port | 6379 | 端口号 |
| Password | "" | 密码（可选） |
| DB | 0 | 数据库编号 |

## 核心键前缀

```
rss:session        # 会话管理
rss:user:info      # 用户信息缓存
rss:resource       # 资源缓存
rss:category       # 分类缓存
rss:visit          # 访问记录
rss:hot:resources  # 热门资源
rss:ratelimit      # 限流控制
```

## 基本操作

### String 操作

```go
// 设置
client.Set(ctx, "key", "value", time.Second*10)

// 获取
value, err := client.Get(ctx, "key").Result()

// 检查存在
exists, err := client.Exists(ctx, "key").Result()

// 设置过期
client.Expire(ctx, "key", time.Hour)
```

### Hash 操作

```go
// 设置字段
client.HSet(ctx, "user:1", "username", "admin")

// 获取字段
username, err := client.HGet(ctx, "user:1", "username").Result()

// 获取所有
all, err := client.HGetAll(ctx, "user:1").Result()
```

### List 操作

```go
// 右侧插入
client.RPush(ctx, "list", "item1", "item2")

// 获取范围
items, err := client.LRange(ctx, "list", 0, -1).Result()
```

### ZSet 操作（排行榜）

```go
// 添加成员
client.ZAdd(ctx, "ranking", &redis.Z{
    Score:  100,
    Member: "user1",
})

// 获取Top 10
top, err := client.ZRevRange(ctx, "ranking", 0, 9).Result()
```

### 计数器

```go
// +1
client.Incr(ctx, "counter")

// +10
client.IncrBy(ctx, "counter", 10)
```

## 使用模式

### 1. 用户会话

```go
kb := config.NewRedisKeyBuilder("rss")
key := kb.BuildUserSessionKey(sessionID)
client.Set(ctx, key, userID, 24*time.Hour)
```

### 2. 用户缓存

```go
kb := config.NewRedisKeyBuilder("rss")
key := kb.BuildUserInfoKey(userID)
client.Set(ctx, key, userJSON, 5*time.Minute)
```

### 3. 限流

```go
kb := config.NewRedisKeyBuilder("rss")
key := fmt.Sprintf("%s:ip:%s", kb.prefix, ip)
pipe := client.Pipeline()
pipe.Incr(ctx, key)
pipe.Expire(ctx, key, time.Minute)
pipe.Exec(ctx)
```

## 性能建议

- **连接池大小**: CPU核心数 × 2
- **过期策略**: 合理设置TTL
- **批量操作**: 使用Pipeline
- **内存控制**: 定期清理无用数据

## 监控指标

```go
stats := client.PoolStats()
fmt.Printf("命中率: %.2f%%\n", 
    float64(stats.Hits)/float64(stats.Hits+stats.Misses)*100)
```

---

**参考**: [详细Redis文档](redis-usage.md)
