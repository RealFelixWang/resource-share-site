/*
Package session provides session management services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package session

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"resource-share-site/internal/model"
	"resource-share-site/internal/service/auth"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// SessionService 会话管理服务接口
type SessionService interface {
	// 创建会话
	CreateSession(ctx *auth.GORMContext, userID uint, data map[string]interface{}, duration time.Duration) (*SessionInfo, error)

	// 获取会话
	GetSession(ctx *auth.GORMContext, sessionID string) (*SessionInfo, error)

	// 更新会话数据
	UpdateSession(ctx *auth.GORMContext, sessionID string, data map[string]interface{}) error

	// 刷新会话（延长过期时间）
	RefreshSession(ctx *auth.GORMContext, sessionID string, duration time.Duration) error

	// 删除会话
	DeleteSession(ctx *auth.GORMContext, sessionID string) error

	// 删除用户所有会话
	DeleteUserSessions(ctx *auth.GORMContext, userID uint) error

	// 获取用户会话列表
	GetUserSessions(ctx *auth.GORMContext, userID uint) ([]SessionInfo, error)

	// 清理过期会话
	CleanExpiredSessions(ctx *auth.GORMContext) (int64, error)

	// 检查会话是否存在且有效
	IsSessionValid(ctx *auth.GORMContext, sessionID string) (bool, error)
}

// SessionInfo 会话信息
type SessionInfo struct {
	ID        uint                   `json:"id"`
	SessionID string                 `json:"session_id"`
	UserID    uint                   `json:"user_id"`
	Data      map[string]interface{} `json:"data"` // 会话数据
	ExpiresAt time.Time              `json:"expires_at"`
	IP        string                 `json:"ip"` // IP地址
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// SessionServiceImpl 会话管理服务实现
type SessionServiceImpl struct {
	db    *gorm.DB
	redis *redis.Client // 可选的Redis客户端
}

// NewSessionService 创建会话管理服务
func NewSessionService(db *gorm.DB, redisClient *redis.Client) SessionService {
	return &SessionServiceImpl{
		db:    db,
		redis: redisClient,
	}
}

// CreateSession 创建会话
func (s *SessionServiceImpl) CreateSession(ctx *auth.GORMContext, userID uint, data map[string]interface{}, duration time.Duration) (*SessionInfo, error) {
	// 生成会话ID
	sessionID := generateSessionID()

	// 计算过期时间
	expiresAt := time.Now().Add(duration)

	// 序列化数据
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// 创建数据库记录
	session := model.Session{
		UserID:    userID,
		SessionID: sessionID,
		Data:      string(dataJSON),
		ExpiresAt: expiresAt,
		IP:        getClientIP(ctx), // 从上下文获取IP
	}

	if err := s.db.Create(&session).Error; err != nil {
		return nil, err
	}

	// 如果配置了Redis，也存储到Redis
	if s.redis != nil {
		s.storeInRedis(sessionID, data, duration)
	}

	// 返回会话信息
	return &SessionInfo{
		ID:        session.ID,
		SessionID: session.SessionID,
		UserID:    session.UserID,
		Data:      data,
		ExpiresAt: session.ExpiresAt,
		IP:        session.IP,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}, nil
}

// GetSession 获取会话
func (s *SessionServiceImpl) GetSession(ctx *auth.GORMContext, sessionID string) (*SessionInfo, error) {
	var session model.Session

	// 先尝试从Redis获取
	if s.redis != nil {
		if info := s.getFromRedis(sessionID); info != nil {
			return info, nil
		}
	}

	// 从数据库获取
	if err := s.db.Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会话不存在或已过期")
		}
		return nil, err
	}

	// 解析数据
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(session.Data), &data); err != nil {
		return nil, err
	}

	return &SessionInfo{
		ID:        session.ID,
		SessionID: session.SessionID,
		UserID:    session.UserID,
		Data:      data,
		ExpiresAt: session.ExpiresAt,
		IP:        session.IP,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}, nil
}

// UpdateSession 更新会话数据
func (s *SessionServiceImpl) UpdateSession(ctx *auth.GORMContext, sessionID string, data map[string]interface{}) error {
	// 序列化数据
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 更新数据库
	if err := s.db.Model(&model.Session{}).Where("session_id = ?", sessionID).Updates(map[string]interface{}{
		"data":       string(dataJSON),
		"updated_at": time.Now(),
	}).Error; err != nil {
		return err
	}

	// 如果配置了Redis，也更新Redis
	if s.redis != nil {
		s.redis.HSet(context.Background(), "session:"+sessionID, "data", string(dataJSON))
		s.redis.Expire(context.Background(), "session:"+sessionID, time.Hour*24)
	}

	return nil
}

// RefreshSession 刷新会话（延长过期时间）
func (s *SessionServiceImpl) RefreshSession(ctx *auth.GORMContext, sessionID string, duration time.Duration) error {
	expiresAt := time.Now().Add(duration)

	// 更新数据库
	if err := s.db.Model(&model.Session{}).Where("session_id = ?", sessionID).Updates(map[string]interface{}{
		"expires_at": expiresAt,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return err
	}

	// 如果配置了Redis，也更新Redis
	if s.redis != nil {
		s.redis.Expire(context.Background(), "session:"+sessionID, duration)
	}

	return nil
}

// DeleteSession 删除会话
func (s *SessionServiceImpl) DeleteSession(ctx *auth.GORMContext, sessionID string) error {
	// 从数据库删除
	if err := s.db.Where("session_id = ?", sessionID).Delete(&model.Session{}).Error; err != nil {
		return err
	}

	// 如果配置了Redis，也从Redis删除
	if s.redis != nil {
		s.redis.Del(context.Background(), "session:"+sessionID)
	}

	return nil
}

// DeleteUserSessions 删除用户所有会话
func (s *SessionServiceImpl) DeleteUserSessions(ctx *auth.GORMContext, userID uint) error {
	// 获取用户的会话ID列表
	var sessions []model.Session
	if err := s.db.Model(&model.Session{}).Where("user_id = ?", userID).Pluck("session_id", &sessions).Error; err != nil {
		return err
	}

	// 从数据库删除
	if err := s.db.Where("user_id = ?", userID).Delete(&model.Session{}).Error; err != nil {
		return err
	}

	// 如果配置了Redis，也从Redis删除
	if s.redis != nil {
		for _, session := range sessions {
			s.redis.Del(context.Background(), "session:"+session.SessionID)
		}
	}

	return nil
}

// GetUserSessions 获取用户会话列表
func (s *SessionServiceImpl) GetUserSessions(ctx *auth.GORMContext, userID uint) ([]SessionInfo, error) {
	var sessions []model.Session

	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&sessions).Error; err != nil {
		return nil, err
	}

	result := make([]SessionInfo, len(sessions))
	for i, session := range sessions {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(session.Data), &data); err != nil {
			return nil, err
		}

		result[i] = SessionInfo{
			ID:        session.ID,
			SessionID: session.SessionID,
			UserID:    session.UserID,
			Data:      data,
			ExpiresAt: session.ExpiresAt,
			IP:        session.IP,
			CreatedAt: session.CreatedAt,
			UpdatedAt: session.UpdatedAt,
		}
	}

	return result, nil
}

// CleanExpiredSessions 清理过期会话
func (s *SessionServiceImpl) CleanExpiredSessions(ctx *auth.GORMContext) (int64, error) {
	// 清理数据库中的过期会话
	result := s.db.Where("expires_at < ?", time.Now()).Delete(&model.Session{})
	if result.Error != nil {
		return 0, result.Error
	}
	deletedCount := result.RowsAffected

	// 如果配置了Redis，也清理Redis中的过期会话
	if s.redis != nil {
		// 获取所有会话键
		keys, err := s.redis.Keys(context.Background(), "session:*").Result()
		if err == nil {
			for _, key := range keys {
				// 检查会话是否过期
				ttl, err := s.redis.TTL(context.Background(), key).Result()
				if err == nil && ttl < 0 {
					s.redis.Del(context.Background(), key)
				}
			}
		}
	}

	return deletedCount, nil
}

// IsSessionValid 检查会话是否存在且有效
func (s *SessionServiceImpl) IsSessionValid(ctx *auth.GORMContext, sessionID string) (bool, error) {
	var count int64

	if err := s.db.Model(&model.Session{}).Where("session_id = ? AND expires_at > ?", sessionID, time.Now()).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// storeInRedis 将会话存储到Redis
func (s *SessionServiceImpl) storeInRedis(sessionID string, data map[string]interface{}, duration time.Duration) {
	ctx := context.Background()

	// 存储会话数据
	dataJSON, _ := json.Marshal(data)
	s.redis.HSet(ctx, "session:"+sessionID, "data", string(dataJSON))
	s.redis.Expire(ctx, "session:"+sessionID, duration)
}

// getFromRedis 从Redis获取会话
func (s *SessionServiceImpl) getFromRedis(sessionID string) *SessionInfo {
	ctx := context.Background()

	// 检查会话是否存在且未过期
	ttl, err := s.redis.TTL(ctx, "session:"+sessionID).Result()
	if err != nil || ttl < 0 {
		return nil
	}

	// 获取会话数据
	dataJSON, err := s.redis.HGet(ctx, "session:"+sessionID, "data").Result()
	if err != nil {
		return nil
	}

	// 解析数据
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
		return nil
	}

	// 计算过期时间
	expiresAt := time.Now().Add(time.Duration(ttl) * time.Second)

	return &SessionInfo{
		SessionID: sessionID,
		Data:      data,
		ExpiresAt: expiresAt,
	}
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	// 使用时间戳 + 随机数生成会话ID
	return time.Now().Format("20060102150405") + "-" + generateRandomString(16)
}

// generateRandomString 生成随机字符串
func generateRandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

// getClientIP 从上下文中获取客户端IP
func getClientIP(ctx *auth.GORMContext) string {
	// TODO: 从Gin上下文或其他地方获取真实的客户端IP
	// 这里暂时返回空字符串
	return ""
}
