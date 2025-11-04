/*
Package resource provides resource statistics services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package resource

import (
	"fmt"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// ResourceStatistics 资源统计信息
type ResourceStatistics struct {
	// 总体统计
	TotalResources int64 `json:"total_resources"`
	TotalDownloads int64 `json:"total_downloads"`
	TotalViews     int64 `json:"total_views"`
	TotalPoints    int64 `json:"total_points"`

	// 状态统计
	PendingResources  int64 `json:"pending_resources"`
	ApprovedResources int64 `json:"approved_resources"`
	RejectedResources int64 `json:"rejected_resources"`

	// 趋势统计
	DailyUploads   []int64  `json:"daily_uploads"`
	DailyDownloads []int64  `json:"daily_downloads"`
	DailyViews     []int64  `json:"daily_views"`
	DateLabels     []string `json:"date_labels"`

	// 用户统计
	ActiveUploaders int64 `json:"active_uploaders"`
	TotalUploaders  int64 `json:"total_uploaders"`

	// 分类统计
	TopCategories []map[string]interface{} `json:"top_categories"`

	// 时间范围
	PeriodStart *time.Time `json:"period_start"`
	PeriodEnd   *time.Time `json:"period_end"`
	ComputedAt  time.Time  `json:"computed_at"`
}

// UploaderStatistics 上传者统计
type UploaderStatistics struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`

	// 资源统计
	TotalResources    int64 `json:"total_resources"`
	PendingResources  int64 `json:"pending_resources"`
	ApprovedResources int64 `json:"approved_resources"`
	RejectedResources int64 `json:"rejected_resources"`

	// 互动统计
	TotalDownloads int64 `json:"total_downloads"`
	TotalViews     int64 `json:"total_views"`
	TotalPoints    int64 `json:"total_points"`

	// 质量指标
	ApprovalRate float64 `json:"approval_rate"`
	AvgDownloads float64 `json:"avg_downloads"`
	AvgViews     float64 `json:"avg_views"`
	Score        float64 `json:"score"` // 排行榜评分

	// 时间信息
	FirstUploadAt *time.Time `json:"first_upload_at"`
	LastUploadAt  *time.Time `json:"last_upload_at"`
}

// PopularResource 热门资源
type PopularResource struct {
	ResourceID     uint      `json:"resource_id"`
	Title          string    `json:"title"`
	CategoryName   string    `json:"category_name"`
	UploaderName   string    `json:"uploader_name"`
	DownloadsCount int64     `json:"downloads_count"`
	ViewsCount     int64     `json:"views_count"`
	PointsPrice    int       `json:"points_price"`
	CreatedAt      time.Time `json:"created_at"`
	Score          float64   `json:"score"`
}

// ResourceTrends 资源趋势
type ResourceTrends struct {
	Date         string  `json:"date"`
	NewResources int64   `json:"new_resources"`
	NewDownloads int64   `json:"new_downloads"`
	NewViews     int64   `json:"new_views"`
	GrowthRate   float64 `json:"growth_rate"`
}

// StatisticsService 统计服务
type StatisticsService struct {
	db *gorm.DB
}

// NewStatisticsService 创建新的统计服务
func NewStatisticsService(db *gorm.DB) *StatisticsService {
	return &StatisticsService{
		db: db,
	}
}

// GetOverallStatistics 获取总体统计信息
// 参数：
//   - startDate: 统计开始时间（可选）
//   - endDate: 统计结束时间（可选）
//
// 返回：
//   - 统计信息
//   - 错误信息
func (s *StatisticsService) GetOverallStatistics(startDate, endDate *time.Time) (*ResourceStatistics, error) {
	stats := &ResourceStatistics{
		ComputedAt: time.Now(),
	}

	// 基础查询条件
	query := s.db.Model(&model.Resource{}).Where("deleted_at IS NULL")

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
		stats.PeriodStart = startDate
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
		stats.PeriodEnd = endDate
	}

	// 总体统计
	if err := query.Count(&stats.TotalResources).Error; err != nil {
		return nil, fmt.Errorf("查询总资源数失败: %w", err)
	}

	var totalDownloads int64
	if err := query.Pluck("downloads_count", &[]int64{}).Error; err != nil {
		return nil, fmt.Errorf("查询总下载数失败: %w", err)
	}
	stats.TotalDownloads = totalDownloads

	var totalViews int64
	if err := query.Pluck("views_count", &[]int64{}).Error; err != nil {
		return nil, fmt.Errorf("查询总浏览数失败: %w", err)
	}
	stats.TotalViews = totalViews

	var totalPoints int64
	if err := query.Pluck("points_price", &[]int64{}).Error; err != nil {
		return nil, fmt.Errorf("查询总积分失败: %w", err)
	}
	stats.TotalPoints = totalPoints

	// 状态统计
	if err := query.Where("status = ?", model.ResourceStatusPending).Count(&stats.PendingResources).Error; err != nil {
		return nil, fmt.Errorf("查询待审核数失败: %w", err)
	}

	if err := query.Where("status = ?", model.ResourceStatusApproved).Count(&stats.ApprovedResources).Error; err != nil {
		return nil, fmt.Errorf("查询已通过数失败: %w", err)
	}

	if err := query.Where("status = ?", model.ResourceStatusRejected).Count(&stats.RejectedResources).Error; err != nil {
		return nil, fmt.Errorf("查询已拒绝数失败: %w", err)
	}

	// 用户统计
	var totalUploaders int64
	if err := query.Distinct("uploaded_by_id").Count(&totalUploaders).Error; err != nil {
		return nil, fmt.Errorf("查询总上传者数失败: %w", err)
	}
	stats.TotalUploaders = totalUploaders

	// 活跃上传者（最近30天有上传的用户）
	var activeUploaders int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	activeQuery := s.db.Model(&model.Resource{}).
		Where("created_at >= ? AND deleted_at IS NULL", thirtyDaysAgo)
	if startDate != nil {
		activeQuery = activeQuery.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		activeQuery = activeQuery.Where("created_at <= ?", *endDate)
	}
	if err := activeQuery.Distinct("uploaded_by_id").Count(&activeUploaders).Error; err != nil {
		return nil, fmt.Errorf("查询活跃上传者数失败: %w", err)
	}
	stats.ActiveUploaders = activeUploaders

	// 获取趋势数据（最近30天）
	stats.DailyUploads, stats.DateLabels = s.getDailyUploads(30, startDate, endDate)
	stats.DailyDownloads, _ = s.getDailyDownloads(30, startDate, endDate)
	stats.DailyViews, _ = s.getDailyViews(30, startDate, endDate)

	// 获取热门分类
	topCategories, err := s.getTopCategories(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("查询热门分类失败: %w", err)
	}
	stats.TopCategories = topCategories

	return stats, nil
}

// GetUploaderStatistics 获取上传者统计信息
// 参数：
//   - userID: 用户ID
//   - startDate: 统计开始时间（可选）
//   - endDate: 统计结束时间（可选）
//
// 返回：
//   - 统计信息
//   - 错误信息
func (s *StatisticsService) GetUploaderStatistics(userID uint, startDate, endDate *time.Time) (*UploaderStatistics, error) {
	// 获取用户信息
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	stats := &UploaderStatistics{
		UserID:   userID,
		Username: user.Username,
		Email:    user.Email,
	}

	// 基础查询条件
	query := s.db.Model(&model.Resource{}).
		Where("uploaded_by_id = ? AND deleted_at IS NULL", userID)

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	// 资源统计
	if err := query.Count(&stats.TotalResources).Error; err != nil {
		return nil, fmt.Errorf("查询总资源数失败: %w", err)
	}

	if err := query.Where("status = ?", model.ResourceStatusPending).Count(&stats.PendingResources).Error; err != nil {
		return nil, fmt.Errorf("查询待审核数失败: %w", err)
	}

	if err := query.Where("status = ?", model.ResourceStatusApproved).Count(&stats.ApprovedResources).Error; err != nil {
		return nil, fmt.Errorf("查询已通过数失败: %w", err)
	}

	if err := query.Where("status = ?", model.ResourceStatusRejected).Count(&stats.RejectedResources).Error; err != nil {
		return nil, fmt.Errorf("查询已拒绝数失败: %w", err)
	}

	// 互动统计
	var totalDownloads int64
	if err := query.Pluck("downloads_count", &[]int64{}).Error; err != nil {
		return nil, fmt.Errorf("查询总下载数失败: %w", err)
	}
	stats.TotalDownloads = totalDownloads

	var totalViews int64
	if err := query.Pluck("views_count", &[]int64{}).Error; err != nil {
		return nil, fmt.Errorf("查询总浏览数失败: %w", err)
	}
	stats.TotalViews = totalViews

	var totalPoints int64
	if err := query.Pluck("points_price", &[]int64{}).Error; err != nil {
		return nil, fmt.Errorf("查询总积分失败: %w", err)
	}
	stats.TotalPoints = totalPoints

	// 质量指标
	if stats.TotalResources > 0 {
		stats.ApprovalRate = float64(stats.ApprovedResources) / float64(stats.TotalResources) * 100
		stats.AvgDownloads = float64(stats.TotalDownloads) / float64(stats.TotalResources)
		stats.AvgViews = float64(stats.TotalViews) / float64(stats.TotalResources)
	}

	// 时间信息
	var firstUploadAt time.Time
	result := query.Order("created_at ASC").Pluck("created_at", &firstUploadAt)
	if result.Error == nil && !firstUploadAt.IsZero() {
		stats.FirstUploadAt = &firstUploadAt
	}

	var lastUploadAt time.Time
	result = query.Order("created_at DESC").Pluck("created_at", &lastUploadAt)
	if result.Error == nil && !lastUploadAt.IsZero() {
		stats.LastUploadAt = &lastUploadAt
	}

	return stats, nil
}

// GetPopularResources 获取热门资源排行
// 参数：
//   - rankingType: 排行类型（downloads, views, latest）
//   - period: 时间周期（day, week, month, year, all）
//   - categoryID: 分类ID筛选（可选）
//   - limit: 限制数量
//   - offset: 偏移量
//
// 返回：
//   - 热门资源列表
//   - 错误信息
func (s *StatisticsService) GetPopularResources(rankingType, period string, categoryID *uint, limit, offset int) ([]*PopularResource, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// 计算时间范围
	periodStart, err := s.calculatePeriodStart(period)
	if err != nil {
		return nil, fmt.Errorf("计算时间范围失败: %w", err)
	}

	// 基础查询
	query := s.db.Table("resources").
		Select(`resources.id as resource_id, resources.title, resources.downloads_count,
				resources.views_count, resources.points_price, resources.created_at,
				categories.name as category_name, users.username as uploader_name`).
		Joins("LEFT JOIN categories ON resources.category_id = categories.id").
		Joins("LEFT JOIN users ON resources.uploaded_by_id = users.id").
		Where("resources.deleted_at IS NULL AND resources.status = ?", model.ResourceStatusApproved)

	if periodStart != nil {
		query = query.Where("resources.created_at >= ?", *periodStart)
	}

	if categoryID != nil {
		query = query.Where("resources.category_id = ?", *categoryID)
	}

	// 应用排序
	switch rankingType {
	case "downloads":
		query = query.Order("resources.downloads_count DESC, resources.created_at DESC")
	case "views":
		query = query.Order("resources.views_count DESC, resources.created_at DESC")
	case "latest":
		query = query.Order("resources.created_at DESC")
	default:
		query = query.Order("resources.downloads_count DESC, resources.created_at DESC")
	}

	// 获取结果
	var results []map[string]interface{}
	if err := query.Offset(offset).Limit(limit).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("查询热门资源失败: %w", err)
	}

	// 转换为结构体
	popularResources := make([]*PopularResource, len(results))
	for i, result := range results {
		popularResources[i] = &PopularResource{
			ResourceID:     uint(result["resource_id"].(int64)),
			Title:          result["title"].(string),
			CategoryName:   result["category_name"].(string),
			UploaderName:   result["uploader_name"].(string),
			DownloadsCount: result["downloads_count"].(int64),
			ViewsCount:     result["views_count"].(int64),
			PointsPrice:    int(result["points_price"].(int64)),
			CreatedAt:      result["created_at"].(time.Time),
		}

		// 计算评分（综合下载量和浏览量）
		downloadsScore := float64(popularResources[i].DownloadsCount) * 0.6
		viewsScore := float64(popularResources[i].ViewsCount) * 0.3
		timeScore := 0.0
		if periodStart != nil {
			// 越新的资源时间分数越高
			daysSinceCreation := int(time.Since(popularResources[i].CreatedAt).Hours() / 24)
			if daysSinceCreation < 30 {
				timeScore = float64(30-daysSinceCreation) * 0.1
			}
		}
		popularResources[i].Score = downloadsScore + viewsScore + timeScore
	}

	return popularResources, nil
}

// GetResourceTrends 获取资源趋势数据
// 参数：
//   - days: 天数
//   - categoryID: 分类ID筛选（可选）
//   - metric: 指标类型（uploads, downloads, views）
//
// 返回：
//   - 趋势数据
//   - 错误信息
func (s *StatisticsService) GetResourceTrends(days int, categoryID *uint, metric string) ([]*ResourceTrends, error) {
	if days <= 0 || days > 365 {
		days = 30
	}

	var trends []*ResourceTrends

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.AddDate(0, 0, 1)

		trend := &ResourceTrends{
			Date: startOfDay.Format("2006-01-02"),
		}

		// 基础查询
		query := s.db.Model(&model.Resource{}).
			Where("created_at >= ? AND created_at < ? AND deleted_at IS NULL", startOfDay, endOfDay)

		if categoryID != nil {
			query = query.Where("category_id = ?", *categoryID)
		}

		// 根据指标类型获取数据
		switch metric {
		case "uploads":
			var count int64
			if err := query.Count(&count).Error; err != nil {
				return nil, fmt.Errorf("查询上传数失败: %w", err)
			}
			trend.NewResources = count
		case "downloads":
			// 下载数需要从资源表中获取
			var totalDownloads int64
			if err := s.db.Model(&model.Resource{}).
				Where("deleted_at IS NULL AND created_at >= ? AND created_at < ?", startOfDay, endOfDay).
				Pluck("downloads_count", &[]int64{}).Error; err != nil {
				return nil, fmt.Errorf("查询下载数失败: %w", err)
			}
			trend.NewDownloads = totalDownloads
		case "views":
			// 浏览数需要从资源表中获取
			var totalViews int64
			if err := s.db.Model(&model.Resource{}).
				Where("deleted_at IS NULL AND created_at >= ? AND created_at < ?", startOfDay, endOfDay).
				Pluck("views_count", &[]int64{}).Error; err != nil {
				return nil, fmt.Errorf("查询浏览数失败: %w", err)
			}
			trend.NewViews = totalViews
		}

		trends = append(trends, trend)
	}

	// 计算增长率
	for i := 1; i < len(trends); i++ {
		var current, previous int64
		switch metric {
		case "uploads":
			current = trends[i].NewResources
			previous = trends[i-1].NewResources
		case "downloads":
			current = trends[i].NewDownloads
			previous = trends[i-1].NewDownloads
		case "views":
			current = trends[i].NewViews
			previous = trends[i-1].NewViews
		}

		if previous > 0 {
			trends[i].GrowthRate = float64(current-previous) / float64(previous) * 100
		}
	}

	return trends, nil
}

// GetUploadersRanking 获取上传者排行榜
// 参数：
//   - rankingType: 排行类型（resources, downloads, views）
//   - period: 时间周期
//   - limit: 限制数量
//   - offset: 偏移量
//
// 返回：
//   - 上传者排行列表
//   - 错误信息
func (s *StatisticsService) GetUploadersRanking(rankingType, period string, limit, offset int) ([]*UploaderStatistics, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// 计算时间范围
	periodStart, err := s.calculatePeriodStart(period)
	if err != nil {
		return nil, fmt.Errorf("计算时间范围失败: %w", err)
	}

	// 获取上传者统计
	var uploaders []*UploaderStatistics

	if periodStart != nil {
		// 有时间范围限制，需要逐个查询用户统计
		var users []model.User
		query := s.db.Model(&model.User{})

		if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
			return nil, fmt.Errorf("查询用户列表失败: %w", err)
		}

		for _, user := range users {
			uploaderStats, err := s.GetUploaderStatistics(user.ID, periodStart, nil)
			if err != nil {
				continue // 跳过查询失败的用户
			}

			// 根据排行类型排序
			switch rankingType {
			case "resources":
				uploaderStats.Score = float64(uploaderStats.TotalResources)
			case "downloads":
				uploaderStats.Score = float64(uploaderStats.TotalDownloads)
			case "views":
				uploaderStats.Score = float64(uploaderStats.TotalViews)
			default:
				uploaderStats.Score = float64(uploaderStats.TotalResources)
			}

			uploaders = append(uploaders, uploaderStats)
		}

		// 按分数排序
		for i := 0; i < len(uploaders)-1; i++ {
			for j := i + 1; j < len(uploaders); j++ {
				if uploaders[i].Score < uploaders[j].Score {
					uploaders[i], uploaders[j] = uploaders[j], uploaders[i]
				}
			}
		}
	} else {
		// 无时间限制，直接查询总体统计
		var users []model.User
		if err := s.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
			return nil, fmt.Errorf("查询用户列表失败: %w", err)
		}

		for _, user := range users {
			uploaderStats, err := s.GetUploaderStatistics(user.ID, nil, nil)
			if err != nil {
				continue
			}

			// 根据排行类型设置分数
			switch rankingType {
			case "resources":
				uploaderStats.Score = float64(uploaderStats.TotalResources)
			case "downloads":
				uploaderStats.Score = float64(uploaderStats.TotalDownloads)
			case "views":
				uploaderStats.Score = float64(uploaderStats.TotalViews)
			default:
				uploaderStats.Score = float64(uploaderStats.TotalResources)
			}

			uploaders = append(uploaders, uploaderStats)
		}
	}

	return uploaders, nil
}

// getDailyUploads 获取每日上传数
func (s *StatisticsService) getDailyUploads(days int, startDate, endDate *time.Time) ([]int64, []string) {
	uploads := make([]int64, days)
	labels := make([]string, days)

	// 如果有自定义时间范围，调整天数
	if startDate != nil && endDate != nil {
		days = int(endDate.Sub(*startDate).Hours() / 24)
		if days > 365 {
			days = 365
		}
		uploads = make([]int64, days)
		labels = make([]string, days)
	}

	for i := 0; i < days; i++ {
		var date time.Time
		if startDate != nil {
			date = startDate.AddDate(0, 0, i)
		} else {
			date = time.Now().AddDate(0, 0, -days+i+1)
		}

		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.AddDate(0, 0, 1)

		var count int64
		query := s.db.Model(&model.Resource{}).
			Where("created_at >= ? AND created_at < ? AND deleted_at IS NULL", startOfDay, endOfDay)

		if err := query.Count(&count).Error; err == nil {
			uploads[i] = count
		}

		labels[i] = startOfDay.Format("01-02")
	}

	return uploads, labels
}

// getDailyDownloads 获取每日下载数
func (s *StatisticsService) getDailyDownloads(days int, startDate, endDate *time.Time) ([]int64, []string) {
	downloads := make([]int64, days)
	labels := make([]string, days)

	// 如果有自定义时间范围，调整天数
	if startDate != nil && endDate != nil {
		days = int(endDate.Sub(*startDate).Hours() / 24)
		if days > 365 {
			days = 365
		}
		downloads = make([]int64, days)
		labels = make([]string, days)
	}

	for i := 0; i < days; i++ {
		var date time.Time
		if startDate != nil {
			date = startDate.AddDate(0, 0, i)
		} else {
			date = time.Now().AddDate(0, 0, -days+i+1)
		}

		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.AddDate(0, 0, 1)

		var totalDownloads int64
		query := s.db.Model(&model.Resource{}).
			Where("created_at >= ? AND created_at < ? AND deleted_at IS NULL", startOfDay, endOfDay)

		if err := query.Pluck("downloads_count", &[]int64{}).Error; err == nil {
			downloads[i] = totalDownloads
		}

		labels[i] = startOfDay.Format("01-02")
	}

	return downloads, labels
}

// getDailyViews 获取每日浏览数
func (s *StatisticsService) getDailyViews(days int, startDate, endDate *time.Time) ([]int64, []string) {
	views := make([]int64, days)
	labels := make([]string, days)

	// 如果有自定义时间范围，调整天数
	if startDate != nil && endDate != nil {
		days = int(endDate.Sub(*startDate).Hours() / 24)
		if days > 365 {
			days = 365
		}
		views = make([]int64, days)
		labels = make([]string, days)
	}

	for i := 0; i < days; i++ {
		var date time.Time
		if startDate != nil {
			date = startDate.AddDate(0, 0, i)
		} else {
			date = time.Now().AddDate(0, 0, -days+i+1)
		}

		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.AddDate(0, 0, 1)

		var totalViews int64
		query := s.db.Model(&model.Resource{}).
			Where("created_at >= ? AND created_at < ? AND deleted_at IS NULL", startOfDay, endOfDay)

		if err := query.Pluck("views_count", &[]int64{}).Error; err == nil {
			views[i] = totalViews
		}

		labels[i] = startOfDay.Format("01-02")
	}

	return views, labels
}

// getTopCategories 获取热门分类
func (s *StatisticsService) getTopCategories(startDate, endDate *time.Time) ([]map[string]interface{}, error) {
	query := s.db.Table("categories").
		Select(`categories.id, categories.name, COUNT(resources.id) as resource_count,
				SUM(resources.downloads_count) as total_downloads,
				SUM(resources.views_count) as total_views`).
		Joins("LEFT JOIN resources ON categories.id = resources.category_id AND resources.deleted_at IS NULL").
		Group("categories.id")

	if startDate != nil {
		query = query.Where("resources.created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("resources.created_at <= ?", *endDate)
	}

	var results []map[string]interface{}
	if err := query.Order("resource_count DESC").Limit(10).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("查询热门分类失败: %w", err)
	}

	return results, nil
}

// calculatePeriodStart 计算时间范围的开始时间
func (s *StatisticsService) calculatePeriodStart(period string) (*time.Time, error) {
	now := time.Now()
	var start time.Time

	switch period {
	case "day":
		start = now.AddDate(0, 0, -1)
	case "week":
		start = now.AddDate(0, 0, -7)
	case "month":
		start = now.AddDate(0, -1, 0)
	case "year":
		start = now.AddDate(-1, 0, 0)
	case "all":
		return nil, nil
	default:
		return nil, fmt.Errorf("未知的周期: %s", period)
	}

	return &start, nil
}
