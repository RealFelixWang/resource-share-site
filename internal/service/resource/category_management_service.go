/*
Package resource provides resource category management services.

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

// CategoryChangeLog 分类变更日志模型
type CategoryChangeLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	ResourceID uint            `gorm:"not null;index" json:"resource_id"`
	Resource   *model.Resource `gorm:"foreignKey:ResourceID" json:"resource"`

	OldCategoryID *uint           `json:"old_category_id"`
	OldCategory   *model.Category `gorm:"foreignKey:OldCategoryID" json:"old_category"`

	NewCategoryID uint            `gorm:"not null" json:"new_category_id"`
	NewCategory   *model.Category `gorm:"foreignKey:NewCategoryID" json:"new_category"`

	ChangedByID uint        `gorm:"not null;index" json:"changed_by_id"`
	ChangedBy   *model.User `gorm:"foreignKey:ChangedByID" json:"changed_by"`

	Reason string `gorm:"size:500" json:"reason"`
}

// TableName 指定表名
func (CategoryChangeLog) TableName() string {
	return "category_change_logs"
}

// CategoryManagementService 分类管理服务
type CategoryManagementService struct {
	db *gorm.DB
}

// NewCategoryManagementService 创建新的分类管理服务
func NewCategoryManagementService(db *gorm.DB) *CategoryManagementService {
	return &CategoryManagementService{
		db: db,
	}
}

// MoveResourceToCategory 将资源移动到指定分类
// 参数：
//   - resourceID: 资源ID
//   - newCategoryID: 新分类ID
//   - changedByID: 操作者ID
//   - reason: 变更原因
//
// 返回：
//   - 分类变更日志
//   - 错误信息
func (s *CategoryManagementService) MoveResourceToCategory(resourceID, newCategoryID, changedByID uint, reason string) (*CategoryChangeLog, error) {
	// 获取资源
	var resource model.Resource
	if err := s.db.First(&resource, resourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("查询资源失败: %w", err)
	}

	// 检查是否是新分类
	if resource.CategoryID == newCategoryID {
		return nil, errors.New("资源已在该分类中")
	}

	// 获取新分类
	var newCategory model.Category
	if err := s.db.First(&newCategory, newCategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("新分类不存在")
		}
		return nil, fmt.Errorf("查询新分类失败: %w", err)
	}

	// 获取旧分类
	if resource.CategoryID != 0 {
		// 记录旧分类ID
		_ = resource.CategoryID
	}

	// 检查操作者权限
	var changer model.User
	if err := s.db.First(&changer, changedByID).Error; err != nil {
		return nil, fmt.Errorf("查询操作者失败: %w", err)
	}

	if changer.Role != "admin" && changer.Role != "moderator" {
		return nil, errors.New("没有分类管理权限")
	}

	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("开始事务失败: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 记录旧分类ID
	oldCategoryID := resource.CategoryID

	// 更新资源的分类
	if err := tx.Model(&resource).
		Updates(map[string]interface{}{
			"category_id": newCategoryID,
			"updated_at":  time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新资源分类失败: %w", err)
	}

	// 创建分类变更日志
	changeLog := &CategoryChangeLog{
		ResourceID:    resourceID,
		OldCategoryID: &oldCategoryID,
		NewCategoryID: newCategoryID,
		ChangedByID:   changedByID,
		Reason:        reason,
	}

	if err := tx.Create(changeLog).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建分类变更日志失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return changeLog, nil
}

// BatchMoveResources 批量移动资源到指定分类
// 参数：
//   - resourceIDs: 资源ID列表
//   - newCategoryID: 新分类ID
//   - changedByID: 操作者ID
//   - reason: 变更原因
//
// 返回：
//   - 成功的数量
//   - 失败的资源ID列表
//   - 错误信息
func (s *CategoryManagementService) BatchMoveResources(resourceIDs []uint, newCategoryID, changedByID uint, reason string) (int, []uint, error) {
	if len(resourceIDs) == 0 {
		return 0, nil, nil
	}

	// 获取新分类
	var newCategory model.Category
	if err := s.db.First(&newCategory, newCategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil, errors.New("新分类不存在")
		}
		return 0, nil, fmt.Errorf("查询新分类失败: %w", err)
	}

	// 检查操作者权限
	var changer model.User
	if err := s.db.First(&changer, changedByID).Error; err != nil {
		return 0, nil, fmt.Errorf("查询操作者失败: %w", err)
	}

	if changer.Role != "admin" && changer.Role != "moderator" {
		return 0, nil, errors.New("没有分类管理权限")
	}

	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return 0, nil, fmt.Errorf("开始事务失败: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	successCount := 0
	failedIDs := []uint{}

	// 逐个处理资源
	for _, resourceID := range resourceIDs {
		var resource model.Resource
		if err := tx.First(&resource, resourceID).Error; err != nil {
			failedIDs = append(failedIDs, resourceID)
			continue
		}

		// 检查是否已是目标分类
		if resource.CategoryID == newCategoryID {
			failedIDs = append(failedIDs, resourceID)
			continue
		}

		oldCategoryID := resource.CategoryID

		// 更新资源分类
		if err := tx.Model(&resource).
			Updates(map[string]interface{}{
				"category_id": newCategoryID,
				"updated_at":  time.Now(),
			}).Error; err != nil {
			failedIDs = append(failedIDs, resourceID)
			continue
		}

		// 创建分类变更日志
		changeLog := &CategoryChangeLog{
			ResourceID:    resourceID,
			OldCategoryID: &oldCategoryID,
			NewCategoryID: newCategoryID,
			ChangedByID:   changedByID,
			Reason:        reason,
		}

		if err := tx.Create(changeLog).Error; err != nil {
			failedIDs = append(failedIDs, resourceID)
			continue
		}

		successCount++
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return successCount, failedIDs, nil
}

// GetCategoryResourceStats 获取分类下资源统计
// 参数：
//   - categoryID: 分类ID
//   - includeChildren: 是否包含子分类
//
// 返回：
//   - 统计信息
//   - 错误信息
func (s *CategoryManagementService) GetCategoryResourceStats(categoryID uint, includeChildren bool) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 基础查询条件
	query := s.db.Model(&model.Resource{}).Where("deleted_at IS NULL")

	if includeChildren {
		// 获取所有子分类ID
		var childCategoryIDs []uint
		if err := s.getAllChildCategoryIDs(categoryID, &childCategoryIDs); err != nil {
			return nil, fmt.Errorf("获取子分类失败: %w", err)
		}

		// 添加分类ID条件（包括自己和子分类）
		categoryIDs := append([]uint{categoryID}, childCategoryIDs...)
		query = query.Where("category_id IN ?", categoryIDs)
	} else {
		query = query.Where("category_id = ?", categoryID)
	}

	// 总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询总数失败: %w", err)
	}
	stats["total"] = total

	// 按状态统计
	var pending int64
	if err := query.Where("status = ?", model.ResourceStatusPending).Count(&pending).Error; err != nil {
		return nil, fmt.Errorf("查询待审核数失败: %w", err)
	}
	stats["pending"] = pending

	var approved int64
	if err := query.Where("status = ?", model.ResourceStatusApproved).Count(&approved).Error; err != nil {
		return nil, fmt.Errorf("查询已通过数失败: %w", err)
	}
	stats["approved"] = approved

	var rejected int64
	if err := query.Where("status = ?", model.ResourceStatusRejected).Count(&rejected).Error; err != nil {
		return nil, fmt.Errorf("查询已拒绝数失败: %w", err)
	}
	stats["rejected"] = rejected

	// 资源统计
	var totalDownloads int64
	if err := query.Pluck("downloads_count", &[]int64{}).Error; err != nil {
		return nil, fmt.Errorf("查询总下载数失败: %w", err)
	}
	stats["total_downloads"] = totalDownloads

	var totalViews int64
	if err := query.Pluck("views_count", &[]int64{}).Error; err != nil {
		return nil, fmt.Errorf("查询总浏览数失败: %w", err)
	}
	stats["total_views"] = totalViews

	var totalPoints int64
	if err := query.Pluck("points_price", &[]int64{}).Error; err != nil {
		return nil, fmt.Errorf("查询总积分失败: %w", err)
	}
	stats["total_points"] = totalPoints

	// 活跃资源数（最近7天有更新的）
	var activeCount int64
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := query.Where("updated_at >= ?", sevenDaysAgo).Count(&activeCount).Error; err != nil {
		return nil, fmt.Errorf("查询活跃资源数失败: %w", err)
	}
	stats["active_count"] = activeCount

	// 平均分（积分/资源数）
	var avgPoints float64
	if total > 0 {
		avgPoints = float64(totalPoints) / float64(total)
	}
	stats["avg_points"] = avgPoints

	// 上传者数量
	var uploaderCount int64
	if err := query.Distinct("uploaded_by_id").Count(&uploaderCount).Error; err != nil {
		return nil, fmt.Errorf("查询上传者数失败: %w", err)
	}
	stats["uploader_count"] = uploaderCount

	return stats, nil
}

// GetCategoryResourceRanking 获取分类下资源排行榜
// 参数：
//   - categoryID: 分类ID
//   - rankingType: 排行类型（downloads, views, points, latest）
//   - limit: 限制数量
//   - includeChildren: 是否包含子分类
//
// 返回：
//   - 排行列表
//   - 错误信息
func (s *CategoryManagementService) GetCategoryResourceRanking(categoryID uint, rankingType string, limit int, includeChildren bool) ([]*model.Resource, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	query := s.db.Model(&model.Resource{}).Where("deleted_at IS NULL AND status = ?", model.ResourceStatusApproved)

	if includeChildren {
		// 获取所有子分类ID
		var childCategoryIDs []uint
		if err := s.getAllChildCategoryIDs(categoryID, &childCategoryIDs); err != nil {
			return nil, fmt.Errorf("获取子分类失败: %w", err)
		}

		// 添加分类ID条件
		categoryIDs := append([]uint{categoryID}, childCategoryIDs...)
		query = query.Where("category_id IN ?", categoryIDs)
	} else {
		query = query.Where("category_id = ?", categoryID)
	}

	// 应用排序
	switch rankingType {
	case "downloads":
		query = query.Order("downloads_count DESC")
	case "views":
		query = query.Order("views_count DESC")
	case "points":
		query = query.Order("points_price DESC")
	case "latest":
		query = query.Order("created_at DESC")
	default:
		query = query.Order("downloads_count DESC")
	}

	var resources []*model.Resource
	if err := query.Preload("Category").Preload("UploadedBy").
		Limit(limit).
		Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("查询排行失败: %w", err)
	}

	return resources, nil
}

// GetCategoryChangeLogs 获取分类变更日志
// 参数：
//   - categoryID: 分类ID（可选）
//   - resourceID: 资源ID（可选）
//   - changedByID: 操作者ID（可选）
//   - page: 页码
//   - pageSize: 每页数量
//   - startDate: 开始时间（可选）
//   - endDate: 结束时间（可选）
//
// 返回：
//   - 变更日志列表
//   - 总数
//   - 错误信息
func (s *CategoryManagementService) GetCategoryChangeLogs(categoryID, resourceID, changedByID *uint, page, pageSize int, startDate, endDate *time.Time) ([]*CategoryChangeLog, int64, error) {
	var changeLogs []*CategoryChangeLog
	var total int64

	query := s.db.Model(&CategoryChangeLog{})

	if categoryID != nil {
		query = query.Where("old_category_id = ? OR new_category_id = ?", *categoryID, *categoryID)
	}
	if resourceID != nil {
		query = query.Where("resource_id = ?", *resourceID)
	}
	if changedByID != nil {
		query = query.Where("changed_by_id = ?", *changedByID)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询变更日志总数失败: %w", err)
	}

	// 获取列表
	if err := query.Preload("Resource").Preload("OldCategory").Preload("NewCategory").Preload("ChangedBy").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&changeLogs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询变更日志列表失败: %w", err)
	}

	return changeLogs, total, nil
}

// ReorganizeCategoryResources 重新整理分类资源（批量移动）
// 参数：
//   - oldCategoryID: 旧分类ID
//   - newCategoryID: 新分类ID
//   - changedByID: 操作者ID
//   - reason: 变更原因
//
// 返回：
//   - 移动的资源数量
//   - 错误信息
func (s *CategoryManagementService) ReorganizeCategoryResources(oldCategoryID, newCategoryID, changedByID uint, reason string) (int, error) {
	// 获取旧分类下的所有资源
	var resources []*model.Resource
	if err := s.db.Where("category_id = ? AND deleted_at IS NULL", oldCategoryID).Find(&resources).Error; err != nil {
		return 0, fmt.Errorf("查询旧分类资源失败: %w", err)
	}

	if len(resources) == 0 {
		return 0, nil
	}

	// 获取资源ID列表
	resourceIDs := make([]uint, 0, len(resources))
	for _, resource := range resources {
		resourceIDs = append(resourceIDs, resource.ID)
	}

	// 批量移动
	successCount, failedIDs, err := s.BatchMoveResources(resourceIDs, newCategoryID, changedByID, reason)
	if err != nil {
		return 0, fmt.Errorf("批量移动资源失败: %w", err)
	}

	if len(failedIDs) > 0 {
		return successCount, fmt.Errorf("部分资源移动失败: %v", failedIDs)
	}

	return successCount, nil
}

// UpdateCategoryResourceCount 更新分类的资源计数
// 参数：
//   - categoryID: 分类ID
//
// 返回：
//   - 错误信息
func (s *CategoryManagementService) UpdateCategoryResourceCount(categoryID uint) error {
	// 计算分类的资源总数
	var count int64
	if err := s.db.Model(&model.Resource{}).
		Where("category_id = ? AND deleted_at IS NULL", categoryID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("统计资源数量失败: %w", err)
	}

	// 更新分类的资源计数
	if err := s.db.Model(&model.Category{}).
		Where("id = ?", categoryID).
		Update("resources_count", count).Error; err != nil {
		return fmt.Errorf("更新分类资源计数失败: %w", err)
	}

	return nil
}

// BatchUpdateAllCategoryResourceCounts 批量更新所有分类的资源计数
// 返回：
//   - 更新的分类数量
//   - 错误信息
func (s *CategoryManagementService) BatchUpdateAllCategoryResourceCounts() (int, error) {
	// 获取所有分类
	var categories []*model.Category
	if err := s.db.Find(&categories).Error; err != nil {
		return 0, fmt.Errorf("查询分类列表失败: %w", err)
	}

	updatedCount := 0
	for _, category := range categories {
		if err := s.UpdateCategoryResourceCount(category.ID); err != nil {
			// 跳过更新失败的分类
			continue
		}
		updatedCount++
	}

	return updatedCount, nil
}

// getAllChildCategoryIDs 获取所有子分类ID（递归）
func (s *CategoryManagementService) getAllChildCategoryIDs(parentID uint, childIDs *[]uint) error {
	var children []*model.Category
	if err := s.db.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		*childIDs = append(*childIDs, child.ID)
		// 递归获取子分类的子分类
		if err := s.getAllChildCategoryIDs(child.ID, childIDs); err != nil {
			return err
		}
	}

	return nil
}

// FindSimilarResources 查找相似资源
// 参数：
//   - resourceID: 参考资源ID
//   - limit: 限制数量
//
// 返回：
//   - 相似资源列表
//   - 错误信息
func (s *CategoryManagementService) FindSimilarResources(resourceID uint, limit int) ([]*model.Resource, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	// 获取参考资源
	var referenceResource model.Resource
	if err := s.db.First(&referenceResource, resourceID).Error; err != nil {
		return nil, fmt.Errorf("查询参考资源失败: %w", err)
	}

	// 基于分类和标签查找相似资源
	var similarResources []*model.Resource
	query := s.db.Model(&model.Resource{}).
		Where("deleted_at IS NULL AND status = ? AND id != ?", model.ResourceStatusApproved, resourceID)

	// 同分类或相似标签
	if referenceResource.CategoryID != 0 {
		query = query.Where("category_id = ?", referenceResource.CategoryID)
	}

	if referenceResource.Tags != "" {
		// 这里可以添加更复杂的标签匹配逻辑
		// 简单起见，我们先按分类查找
	}

	if err := query.Preload("Category").Preload("UploadedBy").
		Order("downloads_count DESC, created_at DESC").
		Limit(limit).
		Find(&similarResources).Error; err != nil {
		return nil, fmt.Errorf("查询相似资源失败: %w", err)
	}

	return similarResources, nil
}
