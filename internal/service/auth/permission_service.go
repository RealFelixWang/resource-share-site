/*
Package auth provides authentication and permission services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package auth

import (
	"errors"
	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// PermissionService 权限服务
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{
		db: db,
	}
}

// IsAdmin 检查用户是否为管理员
func (s *PermissionService) IsAdmin(userID uint) (bool, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false, err
	}

	return user.Role == "admin", nil
}

// IsOwnerOrAdmin 检查用户是否是资源所有者或管理员
func (s *PermissionService) IsOwnerOrAdmin(userID uint, ownerID uint) (bool, error) {
	// 如果是资源所有者，允许操作
	if userID == ownerID {
		return true, nil
	}

	// 检查是否是管理员
	isAdmin, err := s.IsAdmin(userID)
	if err != nil {
		return false, err
	}

	return isAdmin, nil
}

// RequireAdmin 需要管理员权限
func (s *PermissionService) RequireAdmin(userID uint) error {
	isAdmin, err := s.IsAdmin(userID)
	if err != nil {
		return err
	}

	if !isAdmin {
		return errors.New("需要管理员权限")
	}

	return nil
}

// RequireOwnerOrAdmin 需要所有者或管理员权限
func (s *PermissionService) RequireOwnerOrAdmin(userID uint, ownerID uint) error {
	isOwnerOrAdmin, err := s.IsOwnerOrAdmin(userID, ownerID)
	if err != nil {
		return err
	}

	if !isOwnerOrAdmin {
		return errors.New("需要所有者或管理员权限")
	}

	return nil
}
