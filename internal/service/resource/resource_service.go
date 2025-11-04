/*
Package resource provides resource upload and download services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package resource

import (
	"errors"
	"fmt"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// Errors 定义自定义错误
var (
	ErrResourceNotFound      = errors.New("资源不存在")
	ErrResourceNotApproved   = errors.New("资源未通过审核")
	ErrResourceDeleted       = errors.New("资源已删除")
	ErrInsufficientPoints    = errors.New("积分不足")
	ErrDownloadLimitExceeded = errors.New("超出下载限制")
)

// ResourceService 资源服务
type ResourceService struct {
	db *gorm.DB
}

// NewResourceService 创建新的资源服务
func NewResourceService(db *gorm.DB) *ResourceService {
	return &ResourceService{
		db: db,
	}
}

// CreateResource 创建资源
// 参数：
//   - title: 资源标题
//   - description: 资源描述
//   - categoryID: 分类ID
//   - netdiskURL: 网盘链接
//   - pointsPrice: 所需积分（0表示免费）
//   - tags: 标签（JSON格式）
//   - uploadedByID: 上传者ID
//   - source: 资源来源
//
// 返回：
//   - 资源对象
//   - 错误信息
func (s *ResourceService) CreateResource(title, description string, categoryID uint, netdiskURL string, pointsPrice int, tags string, uploadedByID uint, source model.ResourceSource) (*model.Resource, error) {
	// 检查分类是否存在
	var category model.Category
	if err := s.db.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分类不存在")
		}
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}

	// 检查用户是否存在
	var user model.User
	if err := s.db.First(&user, uploadedByID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 创建资源
	resource := &model.Resource{
		Title:        title,
		Description:  description,
		CategoryID:   categoryID,
		NetdiskURL:   netdiskURL,
		PointsPrice:  pointsPrice,
		Tags:         tags,
		UploadedByID: uploadedByID,
		Source:       source,
		Status:       model.ResourceStatusPending, // 默认待审核
	}

	if err := s.db.Create(resource).Error; err != nil {
		return nil, fmt.Errorf("创建资源失败: %w", err)
	}

	return resource, nil
}

// UpdateResource 更新资源
// 参数：
//   - resourceID: 资源ID
//   - title: 资源标题
//   - description: 资源描述
//   - categoryID: 分类ID
//   - netdiskURL: 网盘链接
//   - pointsPrice: 所需积分
//   - tags: 标签（JSON格式）
//
// 返回：
//   - 更新后的资源对象
//   - 错误信息
func (s *ResourceService) UpdateResource(resourceID uint, title, description string, categoryID uint, netdiskURL string, pointsPrice int, tags string) (*model.Resource, error) {
	// 获取原有资源
	var resource model.Resource
	if err := s.db.First(&resource, resourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("查询资源失败: %w", err)
	}

	// 检查资源是否已审核通过（已审核的资源不允许修改基本信息）
	if resource.Status == model.ResourceStatusApproved {
		return nil, errors.New("已审核通过的资源不允许修改")
	}

	// 更新资源
	updates := map[string]interface{}{
		"title":        title,
		"description":  description,
		"category_id":  categoryID,
		"netdisk_url":  netdiskURL,
		"points_price": pointsPrice,
		"tags":         tags,
		"updated_at":   time.Now(),
	}

	if err := s.db.Model(&resource).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新资源失败: %w", err)
	}

	return &resource, nil
}

// DeleteResource 删除资源
// 参数：
//   - resourceID: 资源ID
//   - force: 是否强制删除（软删除或硬删除）
//
// 返回：
//   - 错误信息
func (s *ResourceService) DeleteResource(resourceID uint, force bool) error {
	// 获取资源
	var resource model.Resource
	if err := s.db.First(&resource, resourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("查询资源失败: %w", err)
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

	if force {
		// 硬删除（删除评论等关联数据）
		if err := tx.Where("resource_id = ?", resourceID).Delete(&model.Comment{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("删除评论失败: %w", err)
		}

		if err := tx.Where("resource_id = ?", resourceID).Delete(&model.PointRecord{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("删除积分记录失败: %w", err)
		}

		if err := tx.Delete(&resource).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("删除资源失败: %w", err)
		}
	} else {
		// 软删除
		if err := tx.Delete(&resource).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("软删除资源失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// GetResourceByID 根据ID获取资源
// 参数：
//   - resourceID: 资源ID
//   - includeDeleted: 是否包含已删除的资源
//
// 返回：
//   - 资源对象
//   - 错误信息
func (s *ResourceService) GetResourceByID(resourceID uint, includeDeleted bool) (*model.Resource, error) {
	var resource model.Resource
	query := s.db

	if !includeDeleted {
		query = query.Unscoped().Where("deleted_at IS NULL")
	}

	if err := query.Preload("Category").Preload("UploadedBy").First(&resource, resourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("查询资源失败: %w", err)
	}

	return &resource, nil
}

// GetResources 获取资源列表
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//   - categoryID: 分类ID筛选（可选）
//   - status: 状态筛选（可选）
//   - uploadedByID: 上传者ID筛选（可选）
//   - minPrice: 最低价格筛选（可选）
//   - maxPrice: 最高价格筛选（可选）
//   - orderBy: 排序字段
//   - orderDesc: 是否降序
//
// 返回：
//   - 资源列表
//   - 总数
//   - 错误信息
func (s *ResourceService) GetResources(page, pageSize int, categoryID *uint, status *model.ResourceStatus, uploadedByID *uint, minPrice, maxPrice *int, orderBy string, orderDesc bool) ([]*model.Resource, int64, error) {
	var resources []*model.Resource
	var total int64

	query := s.db.Model(&model.Resource{}).Where("deleted_at IS NULL")

	// 应用筛选条件
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if uploadedByID != nil {
		query = query.Where("uploaded_by_id = ?", *uploadedByID)
	}
	if minPrice != nil {
		query = query.Where("points_price >= ?", *minPrice)
	}
	if maxPrice != nil {
		query = query.Where("points_price <= ?", *maxPrice)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询资源总数失败: %w", err)
	}

	// 应用排序
	if orderBy == "" {
		orderBy = "created_at"
	}
	if orderDesc {
		query = query.Order(fmt.Sprintf("%s DESC", orderBy))
	} else {
		query = query.Order(fmt.Sprintf("%s ASC", orderBy))
	}

	// 获取列表
	if err := query.Preload("Category").Preload("UploadedBy").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&resources).Error; err != nil {
		return nil, 0, fmt.Errorf("查询资源列表失败: %w", err)
	}

	return resources, total, nil
}

// SearchResources 搜索资源
// 参数：
//   - keyword: 搜索关键词
//   - page: 页码
//   - pageSize: 每页数量
//   - categoryID: 分类ID筛选（可选）
//   - status: 状态筛选（可选）
//
// 返回：
//   - 资源列表
//   - 总数
//   - 错误信息
func (s *ResourceService) SearchResources(keyword string, page, pageSize int, categoryID *uint, status *model.ResourceStatus) ([]*model.Resource, int64, error) {
	var resources []*model.Resource
	var total int64

	query := s.db.Model(&model.Resource{}).Where("deleted_at IS NULL")

	// 应用筛选条件
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// 搜索关键词（在标题和描述中搜索）
	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询资源总数失败: %w", err)
	}

	// 获取列表
	if err := query.Preload("Category").Preload("UploadedBy").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&resources).Error; err != nil {
		return nil, 0, fmt.Errorf("查询资源列表失败: %w", err)
	}

	return resources, total, nil
}

// DownloadResource 下载资源
// 参数：
//   - resourceID: 资源ID
//   - userID: 下载用户ID
//
// 返回：
//   - 资源下载链接
//   - 错误信息
func (s *ResourceService) DownloadResource(resourceID, userID uint) (string, error) {
	// 获取资源
	resource, err := s.GetResourceByID(resourceID, false)
	if err != nil {
		return "", err
	}

	// 检查资源是否已审核通过
	if resource.Status != model.ResourceStatusApproved {
		return "", ErrResourceNotApproved
	}

	// 如果需要积分，检查用户积分是否足够
	if resource.PointsPrice > 0 {
		var user model.User
		if err := s.db.First(&user, userID).Error; err != nil {
			return "", fmt.Errorf("查询用户失败: %w", err)
		}

		if user.PointsBalance < resource.PointsPrice {
			return "", ErrInsufficientPoints
		}

		// 扣除积分
		if err := s.db.Model(&user).
			Where("id = ?", userID).
			Update("points_balance", gorm.Expr("points_balance - ?", resource.PointsPrice)).Error; err != nil {
			return "", fmt.Errorf("扣除积分失败: %w", err)
		}

		// 记录积分变更
		pointRecord := &model.PointRecord{
			UserID:       userID,
			Points:       -resource.PointsPrice,
			Type:         model.PointTypeExpense,
			Source:       model.PointSourceResourceDownload,
			Description:  fmt.Sprintf("下载资源: %s", resource.Title),
			ResourceID:   &resourceID,
			BalanceAfter: user.PointsBalance - resource.PointsPrice,
		}
		if err := s.db.Create(pointRecord).Error; err != nil {
			return "", fmt.Errorf("记录积分变更失败: %w", err)
		}
	}

	// 增加下载次数
	if err := s.db.Model(&resource).
		Updates(map[string]interface{}{
			"downloads_count": gorm.Expr("downloads_count + 1"),
			"updated_at":      time.Now(),
		}).Error; err != nil {
		return "", fmt.Errorf("更新下载次数失败: %w", err)
	}

	return resource.NetdiskURL, nil
}

// ViewResource 浏览资源（增加浏览次数）
// 参数：
//   - resourceID: 资源ID
//
// 返回：
//   - 错误信息
func (s *ResourceService) ViewResource(resourceID uint) error {
	// 增加浏览次数
	result := s.db.Model(&model.Resource{}).
		Where("id = ? AND deleted_at IS NULL", resourceID).
		Updates(map[string]interface{}{
			"views_count": gorm.Expr("views_count + 1"),
			"updated_at":  time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("更新浏览次数失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// GetUserResources 获取用户上传的资源
// 参数：
//   - userID: 用户ID
//   - page: 页码
//   - pageSize: 每页数量
//   - status: 状态筛选（可选）
//
// 返回：
//   - 资源列表
//   - 总数
//   - 错误信息
func (s *ResourceService) GetUserResources(userID uint, page, pageSize int, status *model.ResourceStatus) ([]*model.Resource, int64, error) {
	return s.GetResources(page, pageSize, nil, status, &userID, nil, nil, "created_at", true)
}

// GetResourcesByCategory 获取指定分类的资源
// 参数：
//   - categoryID: 分类ID
//   - page: 页码
//   - pageSize: 每页数量
//   - status: 状态筛选（可选）
//   - orderBy: 排序字段
//   - orderDesc: 是否降序
//
// 返回：
//   - 资源列表
//   - 总数
//   - 错误信息
func (s *ResourceService) GetResourcesByCategory(categoryID uint, page, pageSize int, status *model.ResourceStatus, orderBy string, orderDesc bool) ([]*model.Resource, int64, error) {
	return s.GetResources(page, pageSize, &categoryID, status, nil, nil, nil, orderBy, orderDesc)
}

// GetFreeResources 获取免费资源
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//   - categoryID: 分类ID筛选（可选）
//
// 返回：
//   - 资源列表
//   - 总数
//   - 错误信息
func (s *ResourceService) GetFreeResources(page, pageSize int, categoryID *uint) ([]*model.Resource, int64, error) {
	zero := 0
	return s.GetResources(page, pageSize, categoryID, nil, nil, &zero, &zero, "created_at", true)
}

// GetPopularResources 获取热门资源（按下载量排序）
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//   - categoryID: 分类ID筛选（可选）
//   - limitDays: 限制天数（可选）
//
// 返回：
//   - 资源列表
//   - 总数
//   - 错误信息
func (s *ResourceService) GetPopularResources(page, pageSize int, categoryID *uint, limitDays *int) ([]*model.Resource, int64, error) {
	query := s.db.Model(&model.Resource{}).Where("deleted_at IS NULL AND status = ?", model.ResourceStatusApproved)

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	if limitDays != nil {
		since := time.Now().AddDate(0, 0, -*limitDays)
		query = query.Where("created_at >= ?", since)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询资源总数失败: %w", err)
	}

	var resources []*model.Resource
	if err := query.Preload("Category").Preload("UploadedBy").
		Order("downloads_count DESC, created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&resources).Error; err != nil {
		return nil, 0, fmt.Errorf("查询资源列表失败: %w", err)
	}

	return resources, total, nil
}

// UpdateResourceTags 更新资源标签
// 参数：
//   - resourceID: 资源ID
//   - tags: 标签（JSON格式）
//
// 返回：
//   - 错误信息
func (s *ResourceService) UpdateResourceTags(resourceID uint, tags string) error {
	result := s.db.Model(&model.Resource{}).
		Where("id = ? AND deleted_at IS NULL", resourceID).
		Updates(map[string]interface{}{
			"tags":       tags,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("更新资源标签失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// CountResources 统计资源数量
// 参数：
//   - categoryID: 分类ID筛选（可选）
//   - status: 状态筛选（可选）
//   - uploadedByID: 上传者ID筛选（可选）
//
// 返回：
//   - 资源数量
//   - 错误信息
func (s *ResourceService) CountResources(categoryID *uint, status *model.ResourceStatus, uploadedByID *uint) (int64, error) {
	query := s.db.Model(&model.Resource{}).Where("deleted_at IS NULL")

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if uploadedByID != nil {
		query = query.Where("uploaded_by_id = ?", *uploadedByID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计资源数量失败: %w", err)
	}

	return count, nil
}
