/*
Package auth provides user authentication services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package auth

import (
	"errors"
	"time"

	"resource-share-site/internal/model"
	"resource-share-site/pkg/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GORMContext 认证上下文
type GORMContext struct {
	DB *gorm.DB
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // 用户名或邮箱
	Password   string `json:"password" binding:"required,min=6,max=100"`
	Remember   bool   `json:"remember"` // 记住登录状态
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo 用户信息（不包含敏感信息）
type UserInfo struct {
	ID                       uint   `json:"id"`
	Username                 string `json:"username"`
	Email                    string `json:"email"`
	Role                     string `json:"role"`
	Status                   string `json:"status"`
	CanUpload                bool   `json:"can_upload"`
	PointsBalance            int    `json:"points_balance"`
	InviteCode               string `json:"invite_code"`
	UploadedResourcesCount   uint   `json:"uploaded_resources_count"`
	DownloadedResourcesCount uint   `json:"downloaded_resources_count"`
}

// RegisterRequest 注册请求结构
type RegisterRequest struct {
	Username        string `json:"username" binding:"required,min=3,max=50"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6,max=100"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	InviteCode      string `json:"invite_code"` // 可选，有邀请码可以获得奖励
}

// RegisterResponse 注册响应结构
type RegisterResponse struct {
	ID             uint      `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	InviteCode     string    `json:"invite_code"`
	PointsBalance  int       `json:"points_balance"`
	InvitedBy      *UserInfo `json:"invited_by,omitempty"` // 邀请人信息
	RequiresInvite bool      `json:"requires_invite"`      // 是否需要邀请
	RegisteredAt   time.Time `json:"registered_at"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6,max=100"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"max=50"`
	Email    string `json:"email" binding:"email"`
}

// AuthService 认证服务接口
type AuthService interface {
	// 登录 - 支持用户名或邮箱
	Login(ctx *GORMContext, req *LoginRequest) (*LoginResponse, error)

	// 注册
	Register(ctx *GORMContext, req *RegisterRequest) (*RegisterResponse, error)

	// 修改密码
	ChangePassword(ctx *GORMContext, userID uint, req *ChangePasswordRequest) error

	// 更新用户资料
	UpdateProfile(ctx *GORMContext, userID uint, req *UpdateProfileRequest) error

	// 根据用户名或邮箱查找用户
	FindUserByIdentifier(ctx *GORMContext, identifier string) (*model.User, error)

	// 验证密码
	VerifyPassword(password, hash string) error

	// 密码加密
	HashPassword(password string) (string, error)
}

// AuthServiceImpl 认证服务实现
type AuthServiceImpl struct {
	db *gorm.DB
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB) AuthService {
	return &AuthServiceImpl{
		db: db,
	}
}

// Login 登录 - 支持用户名或邮箱
func (s *AuthServiceImpl) Login(ctx *GORMContext, req *LoginRequest) (*LoginResponse, error) {
	// 查找用户（支持用户名或邮箱）
	user, err := s.FindUserByIdentifier(ctx, req.Identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在或密码错误")
		}
		return nil, err
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, errors.New("账户已被禁用或未激活")
	}

	// 验证密码
	if err := s.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, errors.New("用户不存在或密码错误")
	}

	// 生成token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// 设置过期时间
	var expiresAt time.Time
	if req.Remember {
		expiresAt = time.Now().Add(30 * 24 * time.Hour) // 30天
	} else {
		expiresAt = time.Now().Add(24 * time.Hour) // 1天
	}

	// 更新最后登录时间
	s.db.Model(user).Update("last_login_at", time.Now())

	// 构建响应
	response := &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserInfo{
			ID:                       user.ID,
			Username:                 user.Username,
			Email:                    user.Email,
			Role:                     user.Role,
			Status:                   user.Status,
			CanUpload:                user.CanUpload,
			PointsBalance:            user.PointsBalance,
			InviteCode:               user.InviteCode,
			UploadedResourcesCount:   user.UploadedResourcesCount,
			DownloadedResourcesCount: user.DownloadedResourcesCount,
		},
	}

	return response, nil
}

// Register 注册
func (s *AuthServiceImpl) Register(ctx *GORMContext, req *RegisterRequest) (*RegisterResponse, error) {
	// 验证密码确认
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("两次输入的密码不一致")
	}

	// 检查用户名是否已存在
	var count int64
	s.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已被使用")
	}

	// 检查邮箱是否已存在
	s.db.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return nil, errors.New("邮箱已被使用")
	}

	// 加密密码
	passwordHash, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 处理邀请码（可选）
	var inviter *model.User
	if req.InviteCode != "" {
		// 查找邀请人
		s.db.Model(&model.User{}).Where("invite_code = ?", req.InviteCode).First(&inviter)
	}

	// 创建用户
	user := model.User{
		Username:      req.Username,
		Email:         req.Email,
		PasswordHash:  passwordHash,
		Role:          "user",
		Status:        "active",
		CanUpload:     false, // 默认没有上传权限
		InviteCode:    utils.GenerateInviteCode(),
		PointsBalance: 0,
		InvitedByID:   nil,
	}

	if inviter != nil {
		user.InvitedByID = &inviter.ID
	}

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 创建用户
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// 如果有邀请人，奖励积分
		if inviter != nil {
			// 获取邀请奖励规则
			var pointsRule model.PointsRule
			if err := tx.Where("rule_key = ?", "invite_reward").First(&pointsRule).Error; err == nil {
				if pointsRule.IsEnabled {
					// 更新邀请人积分
					newBalance := inviter.PointsBalance + pointsRule.Points
					if err := tx.Model(inviter).Update("points_balance", newBalance).Error; err != nil {
						return err
					}

					// 记录积分流水
					record := model.PointRecord{
						UserID:       inviter.ID,
						Type:         model.PointTypeIncome,
						Points:       pointsRule.Points,
						BalanceAfter: newBalance,
						Source:       model.PointSourceInviteReward,
						Description:  "邀请注册奖励",
					}
					if err := tx.Create(&record).Error; err != nil {
						return err
					}

					// 更新邀请记录
					var invitation model.Invitation
					if err := tx.Where("invite_code = ? AND invitee_id IS NULL", req.InviteCode).First(&invitation).Error; err == nil {
						tx.Model(&invitation).Updates(map[string]interface{}{
							"invitee_id":     user.ID,
							"status":         model.InvitationStatusCompleted,
							"points_awarded": pointsRule.Points,
							"awarded_at":     time.Now(),
						})
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 构建响应
	response := &RegisterResponse{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		InviteCode:     user.InviteCode,
		PointsBalance:  user.PointsBalance,
		RequiresInvite: false,
		RegisteredAt:   time.Now(),
	}

	// 设置邀请人信息（如果有）
	if inviter != nil {
		response.InvitedBy = &UserInfo{
			ID:       inviter.ID,
			Username: inviter.Username,
			Email:    inviter.Email,
		}
	}

	return response, nil
}

// ChangePassword 修改密码
func (s *AuthServiceImpl) ChangePassword(ctx *GORMContext, userID uint, req *ChangePasswordRequest) error {
	// 验证密码确认
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("两次输入的新密码不一致")
	}

	// 获取用户
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}

	// 验证旧密码
	if err := s.VerifyPassword(req.OldPassword, user.PasswordHash); err != nil {
		return errors.New("旧密码错误")
	}

	// 加密新密码
	newPasswordHash, err := s.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// 更新密码
	return s.db.Model(&user).Update("password_hash", newPasswordHash).Error
}

// UpdateProfile 更新用户资料
func (s *AuthServiceImpl) UpdateProfile(ctx *GORMContext, userID uint, req *UpdateProfileRequest) error {
	updates := make(map[string]interface{})

	// 检查用户名（如果提供）
	if req.Username != "" {
		// 检查用户名是否已被其他用户使用
		var count int64
		s.db.Model(&model.User{}).Where("id != ? AND username = ?", userID, req.Username).Count(&count)
		if count > 0 {
			return errors.New("用户名已被使用")
		}
		updates["username"] = req.Username
	}

	// 检查邮箱（如果提供）
	if req.Email != "" {
		// 检查邮箱是否已被其他用户使用
		var count int64
		s.db.Model(&model.User{}).Where("id != ? AND email = ?", userID, req.Email).Count(&count)
		if count > 0 {
			return errors.New("邮箱已被使用")
		}
		updates["email"] = req.Email
	}

	if len(updates) == 0 {
		return errors.New("没有需要更新的字段")
	}

	// 更新用户信息
	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// FindUserByIdentifier 根据用户名或邮箱查找用户
func (s *AuthServiceImpl) FindUserByIdentifier(ctx *GORMContext, identifier string) (*model.User, error) {
	var user model.User

	// 尝试按用户名查找
	err := s.db.Where("username = ?", identifier).First(&user).Error
	if err == nil {
		return &user, nil
	}

	// 如果按用户名没找到，尝试按邮箱查找
	err = s.db.Where("email = ?", identifier).First(&user).Error
	if err == nil {
		return &user, nil
	}

	return nil, err
}

// VerifyPassword 验证密码
func (s *AuthServiceImpl) VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// HashPassword 密码加密
func (s *AuthServiceImpl) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
