/*
Package resource provides resource review services.

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

// ReviewStatus 审核状态
type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"  // 待审核
	ReviewStatusApproved ReviewStatus = "approved" // 已通过
	ReviewStatusRejected ReviewStatus = "rejected" // 已拒绝
)

// ReviewAction 审核动作
type ReviewAction string

const (
	ReviewActionApprove ReviewAction = "approve" // 通过
	ReviewActionReject  ReviewAction = "reject"  // 拒绝
	ReviewActionRevert  ReviewAction = "revert"  // 撤回
)

// ReviewLog 审核日志模型
type ReviewLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	ResourceID uint            `gorm:"not null;index" json:"resource_id"`
	Resource   *model.Resource `gorm:"foreignKey:ResourceID" json:"resource"`

	ReviewerID uint        `gorm:"not null;index" json:"reviewer_id"`
	Reviewer   *model.User `gorm:"foreignKey:ReviewerID" json:"reviewer"`

	Action     ReviewAction         `gorm:"not null;size:20" json:"action"`
	Notes      string               `gorm:"size:500" json:"notes"`
	FromStatus model.ResourceStatus `gorm:"not null;size:20" json:"from_status"`
	ToStatus   model.ResourceStatus `gorm:"not null;size:20" json:"to_status"`
}

// TableName 指定表名
func (ReviewLog) TableName() string {
	return "review_logs"
}

// ReviewService 审核服务
type ReviewService struct {
	db *gorm.DB
}

// NewReviewService 创建新的审核服务
func NewReviewService(db *gorm.DB) *ReviewService {
	return &ReviewService{
		db: db,
	}
}

// ReviewResource 审核资源
// 参数：
//   - resourceID: 资源ID
//   - reviewerID: 审核者ID
//   - action: 审核动作（通过/拒绝）
//   - notes: 审核备注
//
// 返回：
//   - 审核日志
//   - 错误信息
func (s *ReviewService) ReviewResource(resourceID, reviewerID uint, action ReviewAction, notes string) (*ReviewLog, error) {
	// 获取资源
	var resource model.Resource
	if err := s.db.First(&resource, resourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("查询资源失败: %w", err)
	}

	// 检查审核者是否有权限
	var reviewer model.User
	if err := s.db.First(&reviewer, reviewerID).Error; err != nil {
		return nil, fmt.Errorf("查询审核者失败: %w", err)
	}

	// 检查审核者权限（管理员或版主）
	if reviewer.Role != "admin" && reviewer.Role != "moderator" {
		return nil, errors.New("没有审核权限")
	}

	// 记录原状态
	oldStatus := resource.Status

	// 确定新状态
	var newStatus model.ResourceStatus
	switch action {
	case ReviewActionApprove:
		newStatus = model.ResourceStatusApproved
	case ReviewActionReject:
		newStatus = model.ResourceStatusRejected
	case ReviewActionRevert:
		// 撤回审核，恢复到待审核状态
		newStatus = model.ResourceStatusPending
	default:
		return nil, errors.New("未知的审核动作")
	}

	// 检查资源状态是否允许该操作
	if oldStatus == model.ResourceStatusApproved && action == ReviewActionApprove {
		return nil, errors.New("资源已经审核通过")
	}
	if oldStatus == model.ResourceStatusRejected && action == ReviewActionReject {
		return nil, errors.New("资源已经审核拒绝")
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

	// 更新资源状态
	if err := tx.Model(&resource).
		Updates(map[string]interface{}{
			"status":         newStatus,
			"reviewed_by_id": reviewerID,
			"reviewed_at":    time.Now(),
			"review_notes":   notes,
			"updated_at":     time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新资源状态失败: %w", err)
	}

	// 创建审核日志
	reviewLog := &ReviewLog{
		ResourceID: resourceID,
		ReviewerID: reviewerID,
		Action:     action,
		Notes:      notes,
		FromStatus: oldStatus,
		ToStatus:   newStatus,
	}

	if err := tx.Create(reviewLog).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建审核日志失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return reviewLog, nil
}

// BatchReviewResources 批量审核资源
// 参数：
//   - resourceIDs: 资源ID列表
//   - reviewerID: 审核者ID
//   - action: 审核动作（通过/拒绝）
//   - notes: 审核备注
//
// 返回：
//   - 成功的数量
//   - 失败的资源ID列表
//   - 错误信息
func (s *ReviewService) BatchReviewResources(resourceIDs []uint, reviewerID uint, action ReviewAction, notes string) (int, []uint, error) {
	if len(resourceIDs) == 0 {
		return 0, nil, nil
	}

	// 检查审核者权限
	var reviewer model.User
	if err := s.db.First(&reviewer, reviewerID).Error; err != nil {
		return 0, nil, fmt.Errorf("查询审核者失败: %w", err)
	}

	if reviewer.Role != "admin" && reviewer.Role != "moderator" {
		return 0, nil, errors.New("没有审核权限")
	}

	// 确定新状态
	var newStatus model.ResourceStatus
	switch action {
	case ReviewActionApprove:
		newStatus = model.ResourceStatusApproved
	case ReviewActionReject:
		newStatus = model.ResourceStatusRejected
	default:
		return 0, nil, errors.New("批量审核只支持通过和拒绝操作")
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

		oldStatus := resource.Status

		// 检查状态
		if oldStatus == model.ResourceStatusApproved && action == ReviewActionApprove {
			failedIDs = append(failedIDs, resourceID)
			continue
		}
		if oldStatus == model.ResourceStatusRejected && action == ReviewActionReject {
			failedIDs = append(failedIDs, resourceID)
			continue
		}

		// 更新资源状态
		if err := tx.Model(&resource).
			Updates(map[string]interface{}{
				"status":         newStatus,
				"reviewed_by_id": reviewerID,
				"reviewed_at":    time.Now(),
				"review_notes":   notes,
				"updated_at":     time.Now(),
			}).Error; err != nil {
			failedIDs = append(failedIDs, resourceID)
			continue
		}

		// 创建审核日志
		reviewLog := &ReviewLog{
			ResourceID: resourceID,
			ReviewerID: reviewerID,
			Action:     action,
			Notes:      notes,
			FromStatus: oldStatus,
			ToStatus:   newStatus,
		}

		if err := tx.Create(reviewLog).Error; err != nil {
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

// GetPendingResources 获取待审核资源列表
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//   - categoryID: 分类ID筛选（可选）
//   - uploadedByID: 上传者ID筛选（可选）
//   - source: 资源来源筛选（可选）
//
// 返回：
//   - 资源列表
//   - 总数
//   - 错误信息
func (s *ReviewService) GetPendingResources(page, pageSize int, categoryID *uint, uploadedByID *uint, source *model.ResourceSource) ([]*model.Resource, int64, error) {
	var resources []*model.Resource
	var total int64

	query := s.db.Model(&model.Resource{}).
		Where("status = ? AND deleted_at IS NULL", model.ResourceStatusPending)

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if uploadedByID != nil {
		query = query.Where("uploaded_by_id = ?", *uploadedByID)
	}
	if source != nil {
		query = query.Where("source = ?", *source)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询待审核资源总数失败: %w", err)
	}

	// 获取列表
	if err := query.Preload("Category").Preload("UploadedBy").
		Order("created_at ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&resources).Error; err != nil {
		return nil, 0, fmt.Errorf("查询待审核资源列表失败: %w", err)
	}

	return resources, total, nil
}

// GetReviewedResources 获取已审核资源列表
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//   - status: 审核状态（已通过/已拒绝）
//   - reviewerID: 审核者ID筛选（可选）
//   - categoryID: 分类ID筛选（可选）
//   - startDate: 审核开始时间（可选）
//   - endDate: 审核结束时间（可选）
//
// 返回：
//   - 资源列表
//   - 总数
//   - 错误信息
func (s *ReviewService) GetReviewedResources(page, pageSize int, status model.ResourceStatus, reviewerID *uint, categoryID *uint, startDate, endDate *time.Time) ([]*model.Resource, int64, error) {
	var resources []*model.Resource
	var total int64

	query := s.db.Model(&model.Resource{}).
		Where("status = ? AND deleted_at IS NULL", status)

	if reviewerID != nil {
		query = query.Where("reviewed_by_id = ?", *reviewerID)
	}
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if startDate != nil {
		query = query.Where("reviewed_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("reviewed_at <= ?", *endDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询已审核资源总数失败: %w", err)
	}

	// 获取列表
	if err := query.Preload("Category").Preload("UploadedBy").Preload("ReviewedBy").
		Order("reviewed_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&resources).Error; err != nil {
		return nil, 0, fmt.Errorf("查询已审核资源列表失败: %w", err)
	}

	return resources, total, nil
}

// GetReviewLogs 获取审核日志
// 参数：
//   - resourceID: 资源ID（可选）
//   - reviewerID: 审核者ID（可选）
//   - page: 页码
//   - pageSize: 每页数量
//   - startDate: 开始时间（可选）
//   - endDate: 结束时间（可选）
//
// 返回：
//   - 审核日志列表
//   - 总数
//   - 错误信息
func (s *ReviewService) GetReviewLogs(resourceID, reviewerID *uint, page, pageSize int, startDate, endDate *time.Time) ([]*ReviewLog, int64, error) {
	var reviewLogs []*ReviewLog
	var total int64

	query := s.db.Model(&ReviewLog{})

	if resourceID != nil {
		query = query.Where("resource_id = ?", *resourceID)
	}
	if reviewerID != nil {
		query = query.Where("reviewer_id = ?", *reviewerID)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询审核日志总数失败: %w", err)
	}

	// 获取列表
	if err := query.Preload("Resource").Preload("Reviewer").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&reviewLogs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询审核日志列表失败: %w", err)
	}

	return reviewLogs, total, nil
}

// GetReviewStatistics 获取审核统计信息
// 参数：
//   - startDate: 统计开始时间（可选）
//   - endDate: 统计结束时间（可选）
//
// 返回：
//   - 统计信息
//   - 错误信息
func (s *ReviewService) GetReviewStatistics(startDate, endDate *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := s.db.Model(&model.Resource{}).Where("deleted_at IS NULL")

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	// 总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询总数失败: %w", err)
	}
	stats["total"] = total

	// 待审核数
	var pending int64
	if err := query.Where("status = ?", model.ResourceStatusPending).Count(&pending).Error; err != nil {
		return nil, fmt.Errorf("查询待审核数失败: %w", err)
	}
	stats["pending"] = pending

	// 已通过数
	var approved int64
	if err := query.Where("status = ?", model.ResourceStatusApproved).Count(&approved).Error; err != nil {
		return nil, fmt.Errorf("查询已通过数失败: %w", err)
	}
	stats["approved"] = approved

	// 已拒绝数
	var rejected int64
	if err := query.Where("status = ?", model.ResourceStatusRejected).Count(&rejected).Error; err != nil {
		return nil, fmt.Errorf("查询已拒绝数失败: %w", err)
	}
	stats["rejected"] = rejected

	// 通过率
	var approvalRate float64
	if total > 0 {
		approvalRate = float64(approved) / float64(total) * 100
	}
	stats["approval_rate"] = approvalRate

	// 平均审核时间（需要从审核日志计算）
	var avgReviewTime float64
	if startDate != nil && endDate != nil {
		// 这里可以添加更复杂的平均审核时间计算逻辑
		avgReviewTime = 0
	}
	stats["avg_review_time"] = avgReviewTime

	return stats, nil
}

// AutoApproveResource 自动审核资源（简单的规则审核）
// 参数：
//   - resourceID: 资源ID
//
// 返回：
//   - 是否自动通过
//   - 错误信息
func (s *ReviewService) AutoApproveResource(resourceID uint) (bool, error) {
	// 获取资源
	var resource model.Resource
	if err := s.db.First(&resource, resourceID).Error; err != nil {
		return false, fmt.Errorf("查询资源失败: %w", err)
	}

	// 检查是否已经是审核状态
	if resource.Status != model.ResourceStatusPending {
		return false, nil
	}

	// 简单的自动审核规则
	// 1. 检查标题和描述是否完整
	if resource.Title == "" || resource.Description == "" {
		return false, nil
	}

	// 2. 检查是否有有效的网盘链接
	if resource.NetdiskURL == "" {
		return false, nil
	}

	// 3. 检查是否属于系统允许的分类
	// 这里可以添加更复杂的分类检查逻辑

	// 4. 检查上传者信誉（这里简化处理）
	var uploader model.User
	if err := s.db.First(&uploader, resource.UploadedByID).Error; err != nil {
		return false, fmt.Errorf("查询上传者失败: %w", err)
	}

	// 如果上传者是管理员或信誉良好，自动通过
	if uploader.Role == "admin" || uploader.Role == "moderator" {
		return true, nil
	}

	// 默认不自动通过
	return false, nil
}

// RevertReview 撤回审核（将资源恢复到待审核状态）
// 参数：
//   - resourceID: 资源ID
//   - reviewerID: 审核者ID
//   - notes: 备注
//
// 返回：
//   - 审核日志
//   - 错误信息
func (s *ReviewService) RevertReview(resourceID, reviewerID uint, notes string) (*ReviewLog, error) {
	return s.ReviewResource(resourceID, reviewerID, ReviewActionRevert, notes)
}
