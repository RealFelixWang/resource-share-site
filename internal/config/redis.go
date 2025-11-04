/*
Package config provides configuration management for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package config

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisConfig Redis配置结构
type RedisConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     string `mapstructure:"port" json:"port"`
	Password string `mapstructure:"password" json:"password"`
	DB       int    `mapstructure:"db" json:"db"`
}

// InitRedisClient 初始化Redis客户端
func InitRedisClient(cfg *RedisConfig) (*redis.Client, error) {
	if cfg == nil {
		return nil, ErrRedisConfigNil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		// 连接池配置
		PoolSize:     10, // 连接池大小
		MinIdleConns: 5,  // 最小空闲连接数
		MaxRetries:   3,  // 最大重试次数
		// 超时配置
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		// 空闲超时
		IdleCheckFrequency: 60 * time.Second,
		IdleTimeout:        5 * time.Minute,
		// 熔断器
		MaxRetryBackoff: 8 * time.Second,
		MinRetryBackoff: 100 * time.Millisecond,
		Dialer:          nil,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}

	return client, nil
}

// CloseRedisClient 关闭Redis客户端连接
func CloseRedisClient(client *redis.Client) error {
	if client == nil {
		return nil
	}
	return client.Close()
}

// GetRedisClient 安全获取Redis客户端
func GetRedisClient() *redis.Client {
	// TODO: 这里可以实现单例模式或从容器中获取
	// 为了简化示例，暂时返回nil
	return nil
}

// RedisKeyBuilder Redis键构建器
type RedisKeyBuilder struct {
	prefix string
}

// NewRedisKeyBuilder 创建新的键构建器
func NewRedisKeyBuilder(prefix string) *RedisKeyBuilder {
	return &RedisKeyBuilder{
		prefix: prefix,
	}
}

// Build 构建Redis键名
func (rb *RedisKeyBuilder) Build(key ...string) string {
	if len(key) == 0 {
		return rb.prefix
	}
	return fmt.Sprintf("%s:%s", rb.prefix, key[0])
}

// BuildWithSeparator 使用指定分隔符构建键
func (rb *RedisKeyBuilder) BuildWithSeparator(separator string, key ...string) string {
	if len(key) == 0 {
		return rb.prefix
	}
	return fmt.Sprintf("%s%s%s", rb.prefix, separator, key[0])
}

// BuildUserSessionKey 构建用户会话键
func (rb *RedisKeyBuilder) BuildUserSessionKey(sessionID string) string {
	return rb.Build("session", sessionID)
}

// BuildUserInfoKey 构建用户信息缓存键
func (rb *RedisKeyBuilder) BuildUserInfoKey(userID uint) string {
	return rb.Build("user:info", fmt.Sprintf("%d", userID))
}

// BuildResourceCacheKey 构建资源缓存键
func (rb *RedisKeyBuilder) BuildResourceCacheKey(resourceID uint) string {
	return rb.Build("resource", fmt.Sprintf("%d", resourceID))
}

// BuildCategoryCacheKey 构建分类缓存键
func (rb *RedisKeyBuilder) BuildCategoryCacheKey() string {
	return rb.Build("categories")
}

// BuildHotResourcesKey 构建热门资源键
func (rb *RedisKeyBuilder) BuildHotResourcesKey() string {
	return rb.Build("hot:resources")
}

// BuildVisitCountKey 构建访问计数键
func (rb *RedisKeyBuilder) BuildVisitCountKey(date string) string {
	return rb.Build("visit:count", date)
}

// RedisCache Redis缓存接口
type RedisCache interface {
	// 基本操作
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, expiration time.Duration) error
	Del(ctx context.Context, key ...string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 键过期
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// 批量操作
	GetMultiple(ctx context.Context, keys []string) (map[string]string, error)

	// 计数器
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)

	// 哈希操作
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key, field, value string) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) error

	// 列表操作
	LPush(ctx context.Context, key string, values ...string) (int64, error)
	RPush(ctx context.Context, key string, values ...string) (int64, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	LRem(ctx context.Context, key string, count int64, value string) (int64, error)

	// 有序集合操作
	ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error)
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZScore(ctx context.Context, key, member string) (float64, error)
	ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)

	// 管道操作
	Pipeline() redis.Pipeliner

	// 关闭连接
	Close() error
}

// RedisClient Redis客户端包装器
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient 创建新的Redis客户端包装器
func NewRedisClient(client *redis.Client) RedisCache {
	return &RedisClient{
		client: client,
	}
}

// Get 获取值
func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return rc.client.Get(ctx, key).Result()
}

// Set 设置值
func (rc *RedisClient) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return rc.client.Set(ctx, key, value, expiration).Err()
}

// Del 删除键
func (rc *RedisClient) Del(ctx context.Context, key ...string) error {
	return rc.client.Del(ctx, key...).Err()
}

// Exists 检查键是否存在
func (rc *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Expire 设置过期时间
func (rc *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return rc.client.Expire(ctx, key, expiration).Err()
}

// GetMultiple 批量获取
func (rc *RedisClient) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, key := range keys {
		value, err := rc.client.Get(ctx, key).Result()
		if err == nil {
			result[key] = value
		}
	}
	return result, nil
}

// Incr 增加1
func (rc *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return rc.client.Incr(ctx, key).Result()
}

// IncrBy 增加指定值
func (rc *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return rc.client.IncrBy(ctx, key, value).Result()
}

// Decr 减1
func (rc *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return rc.client.Decr(ctx, key).Result()
}

// HGet 获取哈希字段值
func (rc *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return rc.client.HGet(ctx, key, field).Result()
}

// HSet 设置哈希字段
func (rc *RedisClient) HSet(ctx context.Context, key, field, value string) error {
	return rc.client.HSet(ctx, key, field, value).Err()
}

// HGetAll 获取所有哈希字段
func (rc *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return rc.client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (rc *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return rc.client.HDel(ctx, key, fields...).Err()
}

// LPush 左侧插入
func (rc *RedisClient) LPush(ctx context.Context, key string, values ...string) (int64, error) {
	interfaceValues := make([]interface{}, len(values))
	for i, v := range values {
		interfaceValues[i] = v
	}
	return rc.client.LPush(ctx, key, interfaceValues...).Result()
}

// RPush 右侧插入
func (rc *RedisClient) RPush(ctx context.Context, key string, values ...string) (int64, error) {
	interfaceValues := make([]interface{}, len(values))
	for i, v := range values {
		interfaceValues[i] = v
	}
	return rc.client.RPush(ctx, key, interfaceValues...).Result()
}

// LRange 获取列表范围
func (rc *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return rc.client.LRange(ctx, key, start, stop).Result()
}

// LRem 删除列表中的值
func (rc *RedisClient) LRem(ctx context.Context, key string, count int64, value string) (int64, error) {
	return rc.client.LRem(ctx, key, count, value).Result()
}

// ZAdd 添加有序集合成员
func (rc *RedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	pointerMembers := make([]*redis.Z, len(members))
	for i := range members {
		pointerMembers[i] = &members[i]
	}
	return rc.client.ZAdd(ctx, key, pointerMembers...).Result()
}

// ZRange 获取有序集合范围
func (rc *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return rc.client.ZRange(ctx, key, start, stop).Result()
}

// ZRevRange 获取有序集合逆序范围
func (rc *RedisClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return rc.client.ZRevRange(ctx, key, start, stop).Result()
}

// ZScore 获取成员分数
func (rc *RedisClient) ZScore(ctx context.Context, key, member string) (float64, error) {
	return rc.client.ZScore(ctx, key, member).Result()
}

// ZRem 删除有序集合成员
func (rc *RedisClient) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return rc.client.ZRem(ctx, key, members...).Result()
}

// Pipeline 创建管道
func (rc *RedisClient) Pipeline() redis.Pipeliner {
	return rc.client.Pipeline()
}

// Close 关闭连接
func (rc *RedisClient) Close() error {
	return rc.client.Close()
}

// 默认Redis键前缀常量
const (
	SessionKeyPrefix      = "rss:session"       // 会话
	UserInfoKeyPrefix     = "rss:user:info"     // 用户信息
	ResourceKeyPrefix     = "rss:resource"      // 资源
	CategoryKeyPrefix     = "rss:category"      // 分类
	VisitKeyPrefix        = "rss:visit"         // 访问记录
	HotResourcesKeyPrefix = "rss:hot:resources" // 热门资源
	RateLimitKeyPrefix    = "rss:ratelimit"     // 限流
	CacheKeyPrefix        = "rss:cache"         // 通用缓存
)
