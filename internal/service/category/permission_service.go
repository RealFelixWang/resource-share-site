/*
Package category provides category permission control services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package category

import (
	"errors"
	"fmt"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// PermissionType 权限类型
type PermissionType string

const (
	PermissionView   PermissionType = "view"   // 查看
	PermissionCreate PermissionType = "create" // 创建
	PermissionEdit   PermissionType = "edit"   // 编辑
	PermissionDelete PermissionType = "delete" // 删除
	PermissionManage PermissionType = "manage" // 管理
)

// CategoryPermission 分类权限模型
type CategoryPermission struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	CategoryID uint            `gorm:"not null;index" json:"category_id"`
	Category   *model.Category `gorm:"foreignKey:CategoryID" json:"category"`
	UserID     uint            `gorm:"not null;index" json:"user_id"`
	User       *model.User     `gorm:"foreignKey:UserID" json:"user"`
	Permission PermissionType  `gorm:"not null;size:20" json:"permission"`
	GrantedAt  int64           `gorm:"not null" json:"granted_at"`
	GrantedBy  uint            `gorm:"not null" json:"granted_by"` // 授予者ID
}

// TableName 指定表名
func (CategoryPermission) TableName() string {
	return "category_permissions"
}

// PermissionService 权限服务
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService 创建新的权限服务
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{
		db: db,
	}
}

// HasPermission 检查用户是否有指定权限
// 参数：
//   - userID: 用户ID
//   - categoryID: 分类ID
//   - permission: 权限类型
//
// 返回：
//   - 是否有权限
//   - 错误信息
func (s *PermissionService) HasPermission(userID, categoryID uint, permission PermissionType) (bool, error) {
	// 检查直接权限
	var count int64
	err := s.db.Model(&CategoryPermission{}).
		Where("user_id = ? AND category_id = ? AND permission = ?", userID, categoryID, permission).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("查询权限失败: %w", err)
	}
	if count > 0 {
		return true, nil
	}

	// 检查继承权限（从父级分类继承）
	hasInherited, err := s.checkInheritedPermission(userID, categoryID, permission)
	if err != nil {
		return false, fmt.Errorf("检查继承权限失败: %w", err)
	}
	if hasInherited {
		return true, nil
	}

	// 检查全局权限（管理员或超级用户）
	isAdmin, err := s.checkGlobalPermission(userID, permission)
	if err != nil {
		return false, fmt.Errorf("检查全局权限失败: %w", err)
	}

	return isAdmin, nil
}

// GrantPermission 授予权限
// 参数：
//   - userID: 用户ID
//   - categoryID: 分类ID
//   - permission: 权限类型
//   - grantedBy: 授予者ID
//
// 返回：
//   - 权限对象
//   - 错误信息
func (s *PermissionService) GrantPermission(userID, categoryID uint, permission PermissionType, grantedBy uint) (*CategoryPermission, error) {
	// 检查分类是否存在
	var category model.Category
	if err := s.db.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}

	// 检查用户是否存在
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查是否已存在相同权限
	var existingPermission CategoryPermission
	if err := s.db.Where("user_id = ? AND category_id = ? AND permission = ?", userID, categoryID, permission).First(&existingPermission).Error; err == nil {
		// 已存在，更新授予时间和授予者
		existingPermission.GrantedAt = 0 // GORM会自动设置为当前时间
		existingPermission.GrantedBy = grantedBy
		if err := s.db.Save(&existingPermission).Error; err != nil {
			return nil, fmt.Errorf("更新权限失败: %w", err)
		}
		return &existingPermission, nil
	}

	// 创建新权限
	categoryPermission := &CategoryPermission{
		CategoryID: categoryID,
		UserID:     userID,
		Permission: permission,
		GrantedBy:  grantedBy,
	}

	if err := s.db.Create(categoryPermission).Error; err != nil {
		return nil, fmt.Errorf("创建权限失败: %w", err)
	}

	return categoryPermission, nil
}

// RevokePermission 撤销权限
// 参数：
//   - userID: 用户ID
//   - categoryID: 分类ID
//   - permission: 权限类型
//
// 返回：
//   - 错误信息
func (s *PermissionService) RevokePermission(userID, categoryID uint, permission PermissionType) error {
	result := s.db.Where("user_id = ? AND category_id = ? AND permission = ?", userID, categoryID, permission).
		Delete(&CategoryPermission{})

	if result.Error != nil {
		return fmt.Errorf("撤销权限失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("权限不存在")
	}

	return nil
}

// GetUserPermissions 获取用户的所有分类权限
// 参数：
//   - userID: 用户ID
//   - categoryID: 分类ID（可选，为nil则获取用户所有权限）
//
// 返回：
//   - 权限列表
//   - 错误信息
func (s *PermissionService) GetUserPermissions(userID uint, categoryID *uint) ([]*CategoryPermission, error) {
	var permissions []*CategoryPermission
	query := s.db.Preload("Category").Preload("User").Where("user_id = ?", userID)

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	if err := query.Order("granted_at DESC").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("查询用户权限失败: %w", err)
	}

	return permissions, nil
}

// GetCategoryPermissions 获取分类的所有权限
// 参数：
//   - categoryID: 分类ID
//
// 返回：
//   - 权限列表
//   - 错误信息
func (s *PermissionService) GetCategoryPermissions(categoryID uint) ([]*CategoryPermission, error) {
	var permissions []*CategoryPermission

	if err := s.db.Preload("Category").Preload("User").
		Where("category_id = ?", categoryID).
		Order("granted_at DESC").
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("查询分类权限失败: %w", err)
	}

	return permissions, nil
}

// GetAccessibleCategories 获取用户有权限访问的分类列表
// 参数：
//   - userID: 用户ID
//   - permission: 权限类型
//
// 返回：
//   - 分类列表
//   - 错误信息
func (s *PermissionService) GetAccessibleCategories(userID uint, permission PermissionType) ([]*model.Category, error) {
	var categories []*model.Category

	// 获取有直接权限的分类
	var directCategories []*model.Category
	if err := s.db.Table("category_permissions").
		Select("categories.*").
		Joins("JOIN categories ON category_permissions.category_id = categories.id").
		Where("category_permissions.user_id = ? AND category_permissions.permission = ?", userID, permission).
		Find(&directCategories).Error; err != nil {
		return nil, fmt.Errorf("查询直接权限分类失败: %w", err)
	}

	// 合并结果
	categories = append(categories, directCategories...)

	// 如果有管理权限，可以访问所有分类
	hasManagePerm, err := s.HasPermission(userID, 0, PermissionManage)
	if err != nil {
		return nil, fmt.Errorf("检查管理权限失败: %w", err)
	}
	if hasManagePerm {
		if err := s.db.Find(&categories).Error; err != nil {
			return nil, fmt.Errorf("查询所有分类失败: %w", err)
		}
	}

	// 去重
	uniqueCategories := make([]*model.Category, 0, len(categories))
	seen := make(map[uint]bool)
	for _, cat := range categories {
		if !seen[cat.ID] {
			uniqueCategories = append(uniqueCategories, cat)
			seen[cat.ID] = true
		}
	}

	return uniqueCategories, nil
}

// BatchGrantPermission 批量授予权限
// 参数：
//   - userIDs: 用户ID列表
//   - categoryID: 分类ID
//   - permission: 权限类型
//   - grantedBy: 授予者ID
//
// 返回：
//   - 成功授予的数量
//   - 错误信息
func (s *PermissionService) BatchGrantPermission(userIDs []uint, categoryID uint, permission PermissionType, grantedBy uint) (int, error) {
	// 检查分类是否存在
	var category model.Category
	if err := s.db.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrCategoryNotFound
		}
		return 0, fmt.Errorf("查询分类失败: %w", err)
	}

	// 批量创建权限
	permissions := make([]*CategoryPermission, 0, len(userIDs))
	for _, userID := range userIDs {
		permissions = append(permissions, &CategoryPermission{
			CategoryID: categoryID,
			UserID:     userID,
			Permission: permission,
			GrantedBy:  grantedBy,
		})
	}

	if err := s.db.CreateInBatches(permissions, 100).Error; err != nil {
		return 0, fmt.Errorf("批量创建权限失败: %w", err)
	}

	return len(permissions), nil
}

// BatchRevokePermission 批量撤销权限
// 参数：
//   - userIDs: 用户ID列表
//   - categoryID: 分类ID
//   - permission: 权限类型
//
// 返回：
//   - 成功撤销的数量
//   - 错误信息
func (s *PermissionService) BatchRevokePermission(userIDs []uint, categoryID uint, permission PermissionType) (int, error) {
	result := s.db.Where("category_id = ? AND permission = ? AND user_id IN ?", categoryID, permission, userIDs).
		Delete(&CategoryPermission{})

	if result.Error != nil {
		return 0, fmt.Errorf("批量撤销权限失败: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}

// checkInheritedPermission 检查继承权限
func (s *PermissionService) checkInheritedPermission(userID, categoryID uint, permission PermissionType) (bool, error) {
	// 获取分类的父级
	var category model.Category
	if err := s.db.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("查询分类失败: %w", err)
	}

	// 如果没有父级，返回false
	if category.ParentID == nil {
		return false, nil
	}

	// 递归检查父级的权限
	return s.HasPermission(userID, *category.ParentID, permission)
}

// checkGlobalPermission 检查全局权限（管理员权限）
func (s *PermissionService) checkGlobalPermission(userID uint, permission PermissionType) (bool, error) {
	// 检查用户是否是管理员
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("用户不存在")
		}
		return false, fmt.Errorf("查询用户失败: %w", err)
	}

	// 如果是管理员，拥有所有权限
	if user.Role == "admin" {
		return true, nil
	}

	// 检查是否有超级用户权限
	if permission == PermissionView {
		// 查看权限对于活跃用户是开放的
		return user.Status == "active", nil
	}

	return false, nil
}

// ClearUserPermissions 清除用户的所有分类权限
// 参数：
//   - userID: 用户ID
//
// 返回：
//   - 错误信息
func (s *PermissionService) ClearUserPermissions(userID uint) error {
	if err := s.db.Where("user_id = ?", userID).Delete(&CategoryPermission{}).Error; err != nil {
		return fmt.Errorf("清除用户权限失败: %w", err)
	}

	return nil
}

// ClearCategoryPermissions 清除分类的所有权限
// 参数：
//   - categoryID: 分类ID
//
// 返回：
//   - 错误信息
func (s *PermissionService) ClearCategoryPermissions(categoryID uint) error {
	if err := s.db.Where("category_id = ?", categoryID).Delete(&CategoryPermission{}).Error; err != nil {
		return fmt.Errorf("清除分类权限失败: %w", err)
	}

	return nil
}
