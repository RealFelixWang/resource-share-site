package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims JWT声明
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID uint, username string) (string, error) {
	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小时过期
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	// 签名token
	return token.SignedString([]byte("your-secret-key"))
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的token")
}

// GenerateInviteCode 生成邀请码
func GenerateInviteCode() string {
	// 使用 UUID 生成唯一邀请码
	return uuid.New().String()
}

// HashPassword 密码哈希
func HashPassword(password string) (string, error) {
	// 使用 SHA256 生成密码哈希（简单实现，实际应使用 bcrypt 或 Argon2）
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:]), nil
}

// VerifyPassword 验证密码
func VerifyPassword(password, hash string) error {
	// 验证密码哈希
	newHash, err := HashPassword(password)
	if err != nil {
		return err
	}

	if newHash != hash {
		return fmt.Errorf("密码不匹配")
	}

	return nil
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(n int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)

	// 使用加密随机数生成器
	randomBytes := make([]byte, n)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	for i := 0; i < n; i++ {
		result[i] = charset[int(randomBytes[i])%len(charset)]
	}

	return string(result), nil
}

// GetCurrentDate 获取当前日期（格式：YYYY-MM-DD）
func GetCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

// GetCurrentMonth 获取当前月份（格式：YYYY-MM）
func GetCurrentMonth() string {
	return time.Now().Format("2006-01")
}

// GetCurrentQuarter 获取当前季度
func GetCurrentQuarter() string {
	now := time.Now()
	month := now.Month()
	quarter := (int(month)-1)/3 + 1
	return fmt.Sprintf("%d-Q%d", now.Year(), quarter)
}

// GetStartOfDay 获取一天的开始时间
func GetStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetEndOfDay 获取一天的结束时间
func GetEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// GetStartOfMonth 获取一个月的开始时间
func GetStartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// GetEndOfMonth 获取一个月的结束时间
func GetEndOfMonth(t time.Time) time.Time {
	return GetStartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// GetStartOfQuarter 获取一个季度的开始时间
func GetStartOfQuarter(t time.Time) time.Time {
	month := t.Month()
	quarterStartMonth := ((int(month)-1)/3)*3 + 1
	return time.Date(t.Year(), time.Month(quarterStartMonth), 1, 0, 0, 0, 0, t.Location())
}

// GetEndOfQuarter 获取一个季度的结束时间
func GetEndOfQuarter(t time.Time) time.Time {
	startOfQuarter := GetStartOfQuarter(t)
	return startOfQuarter.AddDate(0, 3, 0).Add(-time.Nanosecond)
}

// Paginate 分页参数
type Paginate struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=10" binding:"min=1,max=100"`
}

// GetOffset 计算偏移量
func (p *Paginate) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *Paginate) GetLimit() int {
	return p.PageSize
}
