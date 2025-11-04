/*
Test Program

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"resource-share-site/internal/config"

	"github.com/go-redis/redis/v8"
)

// Redis连接测试程序
func main() {
	fmt.Println("=== Redis连接测试 ===\n")

	// 1. 测试Redis配置
	fmt.Println("1. 初始化Redis配置...")
	redisConfig := &config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "", // 如果有密码请填写
		DB:       0,  // 默认数据库
	}
	fmt.Printf("Redis地址: %s:%s\n", redisConfig.Host, redisConfig.Port)
	fmt.Printf("数据库ID: %d\n\n", redisConfig.DB)

	// 2. 连接Redis
	fmt.Println("2. 连接Redis...")
	client, err := config.InitRedisClient(redisConfig)
	if err != nil {
		log.Fatalf("Redis连接失败: %v", err)
	}
	defer config.CloseRedisClient(client)
	fmt.Println("✅ Redis连接成功\n")

	// 3. 测试基本操作
	fmt.Println("3. 测试基本操作...")
	ctx := context.Background()

	// 测试Set/Get
	key := "test:key"
	value := "test:value"
	if err := client.Set(ctx, key, value, 10*time.Second).Err(); err != nil {
		log.Fatalf("Set操作失败: %v", err)
	}
	fmt.Printf("  ✅ Set操作: %s = %s\n", key, value)

	// 测试Get
	getValue, err := client.Get(ctx, key).Result()
	if err != nil {
		log.Fatalf("Get操作失败: %v", err)
	}
	if getValue != value {
		log.Fatalf("值不匹配: 期望 %s，实际 %s", value, getValue)
	}
	fmt.Printf("  ✅ Get操作: %s = %s\n", key, getValue)

	// 4. 测试键操作
	fmt.Println("\n4. 测试键操作...")

	// 测试Exists
	exists, err := client.Exists(ctx, key).Result()
	if err != nil {
		log.Fatalf("Exists操作失败: %v", err)
	}
	fmt.Printf("  ✅ Exists操作: %s 存在 = %v\n", key, exists > 0)

	// 测试Expire
	if err := client.Expire(ctx, key, 5*time.Second).Err(); err != nil {
		log.Fatalf("Expire操作失败: %v", err)
	}
	fmt.Printf("  ✅ Expire操作: %s 设置5秒过期\n", key)

	// 5. 测试计数器
	fmt.Println("\n5. 测试计数器操作...")

	counterKey := "test:counter"
	if err := client.Set(ctx, counterKey, "0", 0).Err(); err != nil {
		log.Fatalf("设置计数器失败: %v", err)
	}

	// Incr
	count, err := client.Incr(ctx, counterKey).Result()
	if err != nil {
		log.Fatalf("Incr操作失败: %v", err)
	}
	fmt.Printf("  ✅ Incr操作: %s = %d\n", counterKey, count)

	// IncrBy
	count, err = client.IncrBy(ctx, counterKey, 10).Result()
	if err != nil {
		log.Fatalf("IncrBy操作失败: %v", err)
	}
	fmt.Printf("  ✅ IncrBy操作: %s = %d\n", counterKey, count)

	// 6. 测试哈希操作
	fmt.Println("\n6. 测试哈希操作...")

	hashKey := "test:user:1"
	fields := map[string]string{
		"username": "admin",
		"email":    "admin@example.com",
		"role":     "admin",
	}

	for field, value := range fields {
		if err := client.HSet(ctx, hashKey, field, value).Err(); err != nil {
			log.Fatalf("HSet操作失败: %v", err)
		}
	}
	fmt.Printf("  ✅ HSet操作: 设置 %d 个字段\n", len(fields))

	// HGetAll
	allFields, err := client.HGetAll(ctx, hashKey).Result()
	if err != nil {
		log.Fatalf("HGetAll操作失败: %v", err)
	}
	fmt.Printf("  ✅ HGetAll操作: 获取 %d 个字段\n", len(allFields))
	for field, value := range allFields {
		fmt.Printf("    - %s: %s\n", field, value)
	}

	// 7. 测试列表操作
	fmt.Println("\n7. 测试列表操作...")

	listKey := "test:list"
	items := []string{"item1", "item2", "item3"}

	for _, item := range items {
		if err := client.RPush(ctx, listKey, item).Err(); err != nil {
			log.Fatalf("RPush操作失败: %v", err)
		}
	}
	fmt.Printf("  ✅ RPush操作: 插入 %d 个元素\n", len(items))

	// LRange
	listItems, err := client.LRange(ctx, listKey, 0, -1).Result()
	if err != nil {
		log.Fatalf("LRange操作失败: %v", err)
	}
	fmt.Printf("  ✅ LRange操作: 获取 %d 个元素\n", len(listItems))
	for i, item := range listItems {
		fmt.Printf("    [%d]: %s\n", i, item)
	}

	// 8. 测试有序集合操作
	fmt.Println("\n8. 测试有序集合操作...")

	zsetKey := "test:ranking"
	members := []redis.Z{
		{Score: 100, Member: "user1"},
		{Score: 200, Member: "user2"},
		{Score: 300, Member: "user3"},
	}

	// 转换为指针切片
	pointerMembers := make([]*redis.Z, len(members))
	for i := range members {
		pointerMembers[i] = &members[i]
	}
	if err := client.ZAdd(ctx, zsetKey, pointerMembers...).Err(); err != nil {
		log.Fatalf("ZAdd操作失败: %v", err)
	}
	fmt.Printf("  ✅ ZAdd操作: 添加 %d 个成员\n", len(members))

	// ZRevRange
	ranking, err := client.ZRevRange(ctx, zsetKey, 0, -1).Result()
	if err != nil {
		log.Fatalf("ZRevRange操作失败: %v", err)
	}
	fmt.Printf("  ✅ ZRevRange操作: 获取排名\n")
	for i, member := range ranking {
		score, _ := client.ZScore(ctx, zsetKey, member).Result()
		fmt.Printf("    [%d] %s: %.0f分\n", i+1, member, score)
	}

	// 9. 测试管道操作
	fmt.Println("\n9. 测试管道操作...")

	pipe := client.Pipeline()
	pipe.Set(ctx, "pipe:key1", "value1", 0)
	pipe.Set(ctx, "pipe:key2", "value2", 0)
	pipe.Incr(ctx, "pipe:counter")
	pipe.Expire(ctx, "pipe:key1", 60*time.Second)

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Fatalf("Pipeline操作失败: %v", err)
	}
	fmt.Println("  ✅ Pipeline操作: 批量执行3个命令")

	// 10. 测试连接池状态
	fmt.Println("\n10. 测试连接池状态...")
	stats := client.PoolStats()
	fmt.Printf("  ✅ 连接池状态:\n")
	fmt.Printf("    - 总连接数: %d\n", stats.TotalConns)
	fmt.Printf("    - 空闲连接数: %d\n", stats.IdleConns)
	fmt.Printf("    - 命中次数: %d\n", stats.Hits)
	fmt.Printf("    - 未命中次数: %d\n", stats.Misses)
	fmt.Printf("    - 超时次数: %d\n", stats.Timeouts)
	fmt.Printf("    - 拒绝次数: %d\n", stats.StaleConns)

	// 11. 清理测试数据
	fmt.Println("\n11. 清理测试数据...")
	keys := []string{
		key, counterKey, hashKey, listKey, zsetKey,
		"pipe:key1", "pipe:key2", "pipe:counter",
	}
	client.Del(ctx, keys...)
	fmt.Printf("  ✅ 清理 %d 个测试键\n", len(keys))

	// 12. 测试键构建器
	fmt.Println("\n12. 测试键构建器...")
	keyBuilder := config.NewRedisKeyBuilder("rss")
	testKeys := []string{
		keyBuilder.Build("session", "abc123"),
		keyBuilder.BuildUserInfoKey(100),
		keyBuilder.BuildResourceCacheKey(200),
		keyBuilder.BuildCategoryCacheKey(),
		keyBuilder.BuildHotResourcesKey(),
		keyBuilder.BuildVisitCountKey("2025-10-31"),
	}
	fmt.Printf("  ✅ 键构建器测试: 生成 %d 个键\n", len(testKeys))
	for _, k := range testKeys {
		fmt.Printf("    - %s\n", k)
	}

	// 13. 测试缓存模拟
	fmt.Println("\n13. 模拟缓存操作...")
	模拟缓存操作 := []struct {
		key    string
		value  string
		expire time.Duration
	}{
		{"user:100", "{\"id\":100,\"username\":\"admin\"}", 5 * time.Minute},
		{"resource:200", "{\"id\":200,\"title\":\"测试资源\"}", 10 * time.Minute},
		{"hot:categories", "[\"软件工具\",\"电子资料\"]", 15 * time.Minute},
	}

	for _, item := range 模拟缓存操作 {
		if err := client.Set(ctx, item.key, item.value, item.expire).Err(); err != nil {
			log.Fatalf("缓存设置失败: %v", err)
		}
		fmt.Printf("  ✅ 设置缓存: %s (过期: %v)\n", item.key, item.expire)
	}

	// 获取并验证缓存
	for _, item := range 模拟缓存操作 {
		value, err := client.Get(ctx, item.key).Result()
		if err != nil {
			log.Fatalf("缓存获取失败: %v", err)
		}
		if len(value) > 50 {
			value = value[:50] + "..."
		}
		fmt.Printf("  ✅ 获取缓存: %s = %s\n", item.key, value)
	}

	fmt.Println("\n=== 所有Redis测试通过! ===")
}
