/*
Package category provides category hierarchy management services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package category

import (
	"errors"
	"fmt"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// Errors 定义自定义错误
var (
	ErrCategoryNotFound         = errors.New("分类不存在")
	ErrCategoryNameExists       = errors.New("分类名称已存在")
	ErrParentNotFound           = errors.New("父级分类不存在")
	ErrCannotDeleteWithChildren = errors.New("无法删除包含子分类的分类")
	ErrCannotMoveToSelf         = errors.New("不能将分类移动到自己下面")
	ErrCannotMoveToDescendant   = errors.New("不能将分类移动到自己的子分类下面")
)

// CategoryService 分类服务
type CategoryService struct {
	db *gorm.DB
}

// NewCategoryService 创建新的分类服务
func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{
		db: db,
	}
}

// CreateCategory 创建分类
// 参数：
//   - name: 分类名称
//   - description: 分类描述
//   - icon: 图标
//   - color: 颜色
//   - parentID: 父级ID
//   - sortOrder: 排序
//
// 返回：
//   - 分类对象
//   - 错误信息
func (s *CategoryService) CreateCategory(name, description, icon, color string, parentID *uint, sortOrder int) (*model.Category, error) {
	// 检查分类名称是否已存在（同级下）
	query := s.db.Model(&model.Category{}).Where("name = ?", name)
	if parentID != nil {
		query = query.Where("parent_id = ?", parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	var existingCategory model.Category
	if err := query.First(&existingCategory).Error; err == nil {
		return nil, ErrCategoryNameExists
	}

	// 检查父级分类是否存在
	if parentID != nil {
		var parentCategory model.Category
		if err := s.db.First(&parentCategory, *parentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrParentNotFound
			}
			return nil, fmt.Errorf("查询父级分类失败: %w", err)
		}
	}

	// 创建分类
	category := &model.Category{
		Name:        name,
		Description: description,
		Icon:        icon,
		Color:       color,
		ParentID:    parentID,
		SortOrder:   sortOrder,
	}

	if err := s.db.Create(category).Error; err != nil {
		return nil, fmt.Errorf("创建分类失败: %w", err)
	}

	return category, nil
}

// UpdateCategory 更新分类
// 参数：
//   - categoryID: 分类ID
//   - name: 分类名称
//   - description: 分类描述
//   - icon: 图标
//   - color: 颜色
//   - sortOrder: 排序
//
// 返回：
//   - 更新后的分类对象
//   - 错误信息
func (s *CategoryService) UpdateCategory(categoryID uint, name, description, icon, color string, sortOrder int) (*model.Category, error) {
	// 获取原有分类
	var category model.Category
	if err := s.db.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}

	// 检查分类名称是否已存在（同级下，排除当前分类）
	query := s.db.Model(&model.Category{}).
		Where("name = ? AND id != ?", name, categoryID)
	if category.ParentID != nil {
		query = query.Where("parent_id = ?", *category.ParentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	var existingCategory model.Category
	if err := query.First(&existingCategory).Error; err == nil {
		return nil, ErrCategoryNameExists
	}

	// 更新分类
	updates := map[string]interface{}{
		"name":        name,
		"description": description,
		"icon":        icon,
		"color":       color,
		"sort_order":  sortOrder,
		"updated_at":  time.Now(),
	}

	if err := s.db.Model(&category).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新分类失败: %w", err)
	}

	return &category, nil
}

// DeleteCategory 删除分类
// 参数：
//   - categoryID: 分类ID
//   - force: 是否强制删除（包含子分类）
//
// 返回：
//   - 错误信息
func (s *CategoryService) DeleteCategory(categoryID uint, force bool) error {
	// 获取分类
	var category model.Category
	if err := s.db.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("查询分类失败: %w", err)
	}

	// 检查是否有子分类
	var childrenCount int64
	if err := s.db.Model(&model.Category{}).Where("parent_id = ?", categoryID).Count(&childrenCount).Error; err != nil {
		return fmt.Errorf("查询子分类数量失败: %w", err)
	}

	if childrenCount > 0 && !force {
		return ErrCannotDeleteWithChildren
	}

	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开始事务失败: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 如果是强制删除，先删除所有子分类
	if force && childrenCount > 0 {
		if err := tx.Where("parent_id = ?", categoryID).Delete(&model.Category{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("删除子分类失败: %w", err)
		}
	}

	// 删除当前分类
	if err := tx.Delete(&category).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除分类失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// GetCategoryByID 根据ID获取分类
// 参数：
//   - categoryID: 分类ID
//
// 返回：
//   - 分类对象
//   - 错误信息
func (s *CategoryService) GetCategoryByID(categoryID uint) (*model.Category, error) {
	var category model.Category
	if err := s.db.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}

	return &category, nil
}

// GetAllCategories 获取所有分类（平铺）
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//
// 返回：
//   - 分类列表
//   - 总数
//   - 错误信息
func (s *CategoryService) GetAllCategories(page, pageSize int) ([]*model.Category, int64, error) {
	var categories []*model.Category
	var total int64

	query := s.db.Model(&model.Category{})

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询分类总数失败: %w", err)
	}

	// 获取列表
	if err := query.Preload("Parent").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("parent_id ASC, sort_order ASC, created_at ASC").
		Find(&categories).Error; err != nil {
		return nil, 0, fmt.Errorf("查询分类列表失败: %w", err)
	}

	return categories, total, nil
}

// GetCategoryTree 获取分类树
// 参数：
//   - rootParentID: 根父级ID，nil表示获取所有顶级分类
//   - maxDepth: 最大深度
//
// 返回：
//   - 分类树
//   - 错误信息
func (s *CategoryService) GetCategoryTree(rootParentID *uint, maxDepth int) ([]*model.Category, error) {
	if maxDepth <= 0 || maxDepth > 10 {
		maxDepth = 5 // 默认最大5层
	}

	var categories []*model.Category
	query := s.db.Model(&model.Category{})

	if rootParentID != nil {
		query = query.Where("parent_id = ?", *rootParentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if err := query.Preload("Children.Children.Children.Children.Children").
		Order("sort_order ASC, created_at ASC").
		Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询分类树失败: %w", err)
	}

	return categories, nil
}

// GetCategoryPath 获取分类路径（从根到当前分类）
// 参数：
//   - categoryID: 分类ID
//
// 返回：
//   - 分类路径
//   - 错误信息
func (s *CategoryService) GetCategoryPath(categoryID uint) ([]*model.Category, error) {
	var path []*model.Category
	currentID := categoryID

	for {
		var category model.Category
		if err := s.db.First(&category, currentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break // 已到达根节点
			}
			return nil, fmt.Errorf("查询分类失败: %w", err)
		}

		// 将当前分类添加到路径开头
		path = append([]*model.Category{&category}, path...)

		// 如果没有父级，则到达根节点
		if category.ParentID == nil {
			break
		}

		// 移动到父级
		currentID = *category.ParentID
	}

	return path, nil
}

// MoveCategory 移动分类（修改父级）
// 参数：
//   - categoryID: 要移动的分类ID
//   - newParentID: 新的父级ID，nil表示移动到顶级
//
// 返回：
//   - 错误信息
func (s *CategoryService) MoveCategory(categoryID uint, newParentID *uint) error {
	// 获取要移动的分类
	var category model.Category
	if err := s.db.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("查询分类失败: %w", err)
	}

	// 检查不能移动到自己下面
	if newParentID != nil && *newParentID == categoryID {
		return ErrCannotMoveToSelf
	}

	// 检查不能移动到自己的子分类下面
	if newParentID != nil {
		if err := s.checkIsDescendant(categoryID, *newParentID); err == nil {
			return ErrCannotMoveToDescendant
		}
	}

	// 检查新父级是否存在
	if newParentID != nil {
		var parentCategory model.Category
		if err := s.db.First(&parentCategory, *newParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrParentNotFound
			}
			return fmt.Errorf("查询父级分类失败: %w", err)
		}
	}

	// 更新父级
	updates := map[string]interface{}{
		"parent_id":  newParentID,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(&category).Updates(updates).Error; err != nil {
		return fmt.Errorf("移动分类失败: %w", err)
	}

	return nil
}

// UpdateSortOrder 更新排序
// 参数：
//   - categoryID: 分类ID
//   - sortOrder: 新的排序值
//
// 返回：
//   - 错误信息
func (s *CategoryService) UpdateSortOrder(categoryID uint, sortOrder int) error {
	result := s.db.Model(&model.Category{}).
		Where("id = ?", categoryID).
		Updates(map[string]interface{}{
			"sort_order": sortOrder,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("更新排序失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// checkIsDescendant 检查targetID是否是sourceID的子分类
func (s *CategoryService) checkIsDescendant(sourceID, targetID uint) error {
	currentID := targetID
	for {
		if currentID == sourceID {
			return nil // target是source的子分类
		}

		var category model.Category
		if err := s.db.First(&category, currentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break // 已到达根节点，target不是source的子分类
			}
			return fmt.Errorf("查询分类失败: %w", err)
		}

		if category.ParentID == nil {
			break // 到达根节点
		}

		currentID = *category.ParentID
	}

	return errors.New("not descendant")
}

// GetRootCategories 获取顶级分类
// 参数：无
//
// 返回：
//   - 顶级分类列表
//   - 错误信息
func (s *CategoryService) GetRootCategories() ([]*model.Category, error) {
	var categories []*model.Category

	if err := s.db.Model(&model.Category{}).
		Where("parent_id IS NULL").
		Order("sort_order ASC, created_at ASC").
		Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询顶级分类失败: %w", err)
	}

	return categories, nil
}

// GetChildCategories 获取子分类
// 参数：
//   - parentID: 父级分类ID
//
// 返回：
//   - 子分类列表
//   - 错误信息
func (s *CategoryService) GetChildCategories(parentID uint) ([]*model.Category, error) {
	var categories []*model.Category

	if err := s.db.Model(&model.Category{}).
		Where("parent_id = ?", parentID).
		Order("sort_order ASC, created_at ASC").
		Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询子分类失败: %w", err)
	}

	return categories, nil
}

// CountCategories 统计分类数量
// 参数：
//   - parentID: 父级分类ID，nil表示统计所有
//
// 返回：
//   - 分类数量
//   - 错误信息
func (s *CategoryService) CountCategories(parentID *uint) (int64, error) {
	var count int64
	query := s.db.Model(&model.Category{})

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计分类数量失败: %w", err)
	}

	return count, nil
}
