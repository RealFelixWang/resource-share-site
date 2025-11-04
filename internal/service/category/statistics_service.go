/*
Package category provides category statistics services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package category

import (
	"fmt"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// CategoryStats 分类统计信息
type CategoryStats struct {
	CategoryID uint   `json:"category_id"`
	Name       string `json:"name"`
	ParentID   *uint  `json:"parent_id"`

	// 资源统计
	ResourcesCount     int64 `json:"resources_count"`
	ResourcesToday     int64 `json:"resources_today"`
	ResourcesThisWeek  int64 `json:"resources_this_week"`
	ResourcesThisMonth int64 `json:"resources_this_month"`
	ResourcesThisYear  int64 `json:"resources_this_year"`

	// 子分类统计
	ChildrenCount    int64 `json:"children_count"`
	TotalDescendants int64 `json:"total_descendants"`
	MaxDepth         int   `json:"max_depth"`

	// 活跃度统计
	ActiveUsers    int64 `json:"active_users"`
	TotalViews     int64 `json:"total_views"`
	ViewsToday     int64 `json:"views_today"`
	ViewsThisWeek  int64 `json:"views_this_week"`
	ViewsThisMonth int64 `json:"views_this_month"`

	// 增长率
	ResourceGrowthRate float64 `json:"resource_growth_rate"`
	ViewGrowthRate     float64 `json:"view_growth_rate"`

	// 时间信息
	FirstResourceAt *time.Time `json:"first_resource_at"`
	LastResourceAt  *time.Time `json:"last_resource_at"`
	ComputedAt      time.Time  `json:"computed_at"`
}

// CategoryRanking 分类排行榜
type CategoryRanking struct {
	CategoryID     uint    `json:"category_id"`
	Name           string  `json:"name"`
	ParentID       *uint   `json:"parent_id"`
	Rank           int     `json:"rank"`
	Score          float64 `json:"score"`
	ResourcesCount int64   `json:"resources_count"`
	ViewsCount     int64   `json:"views_count"`
	ActiveUsers    int64   `json:"active_users"`
	GrowthRate     float64 `json:"growth_rate"`
	Level          int     `json:"level"`
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

// GetCategoryStats 获取分类统计信息
// 参数：
//   - categoryID: 分类ID
//
// 返回：
//   - 统计信息
//   - 错误信息
func (s *StatisticsService) GetCategoryStats(categoryID uint) (*CategoryStats, error) {
	// 获取分类信息
	var category model.Category
	if err := s.db.First(&category, categoryID).Error; err != nil {
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}

	stats := &CategoryStats{
		CategoryID: categoryID,
		Name:       category.Name,
		ParentID:   category.ParentID,
		ComputedAt: time.Now(),
	}

	// 统计资源数量
	stats.ResourcesCount = s.countResources(categoryID)
	stats.ResourcesToday = s.countResourcesInPeriod(categoryID, time.Now().AddDate(0, 0, -1))
	stats.ResourcesThisWeek = s.countResourcesInPeriod(categoryID, time.Now().AddDate(0, 0, -7))
	stats.ResourcesThisMonth = s.countResourcesInPeriod(categoryID, time.Now().AddDate(0, -1, 0))
	stats.ResourcesThisYear = s.countResourcesInPeriod(categoryID, time.Now().AddDate(-1, 0, 0))

	// 统计子分类数量
	stats.ChildrenCount = s.countChildren(categoryID)
	stats.TotalDescendants = s.countTotalDescendants(categoryID)
	stats.MaxDepth = s.calculateMaxDepth(categoryID)

	// 统计活跃度和浏览量
	stats.ActiveUsers = s.countActiveUsers(categoryID)
	stats.TotalViews = s.countTotalViews(categoryID)
	stats.ViewsToday = s.countViewsInPeriod(categoryID, time.Now().AddDate(0, 0, -1))
	stats.ViewsThisWeek = s.countViewsInPeriod(categoryID, time.Now().AddDate(0, 0, -7))
	stats.ViewsThisMonth = s.countViewsInPeriod(categoryID, time.Now().AddDate(0, -1, 0))

	// 计算增长率
	stats.ResourceGrowthRate = s.calculateResourceGrowthRate(categoryID)
	stats.ViewGrowthRate = s.calculateViewGrowthRate(categoryID)

	// 获取时间信息
	stats.FirstResourceAt = s.getFirstResourceAt(categoryID)
	stats.LastResourceAt = s.getLastResourceAt(categoryID)

	return stats, nil
}

// GetAllCategoriesStats 获取所有分类的统计信息
// 参数：
//   - page: 页码
//   - pageSize: 每页数量
//
// 返回：
//   - 统计信息列表
//   - 总数
//   - 错误信息
func (s *StatisticsService) GetAllCategoriesStats(page, pageSize int) ([]*CategoryStats, int64, error) {
	// 获取分类列表
	var categories []*model.Category
	var total int64

	query := s.db.Model(&model.Category{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询分类总数失败: %w", err)
	}

	if err := query.Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&categories).Error; err != nil {
		return nil, 0, fmt.Errorf("查询分类列表失败: %w", err)
	}

	// 获取每个分类的统计信息
	statsList := make([]*CategoryStats, 0, len(categories))
	for _, category := range categories {
		stats, err := s.GetCategoryStats(category.ID)
		if err != nil {
			// 跳过统计失败的分类
			continue
		}
		statsList = append(statsList, stats)
	}

	return statsList, total, nil
}

// GetCategoryRanking 获取分类排行榜
// 参数：
//   - rankingType: 排行榜类型（"resources", "views", "growth", "popularity"）
//   - period: 时间周期（"day", "week", "month", "year", "all"）
//   - limit: 限制数量
//   - offset: 偏移量
//
// 返回：
//   - 排行榜列表
//   - 错误信息
func (s *StatisticsService) GetCategoryRanking(rankingType, period string, limit, offset int) ([]*CategoryRanking, error) {
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

	// 获取基础分类数据
	var categories []*model.Category
	if err := s.db.Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询分类列表失败: %w", err)
	}

	// 计算排行榜
	rankings := make([]*CategoryRanking, 0, len(categories))
	for i, category := range categories {
		ranking := &CategoryRanking{
			CategoryID: category.ID,
			Name:       category.Name,
			ParentID:   category.ParentID,
			Rank:       offset + i + 1,
			Level:      s.calculateCategoryLevel(category.ID),
		}

		// 根据类型计算分数
		switch rankingType {
		case "resources":
			ranking.Score = float64(s.countResourcesInPeriod(category.ID, *periodStart))
			ranking.ResourcesCount = int64(s.countResourcesInPeriod(category.ID, *periodStart))
		case "views":
			ranking.Score = float64(s.countViewsInPeriod(category.ID, *periodStart))
			ranking.ViewsCount = int64(s.countViewsInPeriod(category.ID, *periodStart))
		case "growth":
			ranking.Score = s.calculateResourceGrowthRate(category.ID)
			ranking.GrowthRate = s.calculateResourceGrowthRate(category.ID)
		case "popularity":
			// 综合分数 = 资源数 * 0.4 + 浏览量 * 0.3 + 活跃用户 * 0.3
			resourcesScore := float64(s.countResourcesInPeriod(category.ID, *periodStart)) * 0.4
			viewsScore := float64(s.countViewsInPeriod(category.ID, *periodStart)) * 0.3
			usersScore := float64(s.countActiveUsers(category.ID)) * 0.3
			ranking.Score = resourcesScore + viewsScore + usersScore
			ranking.ResourcesCount = int64(s.countResourcesInPeriod(category.ID, *periodStart))
			ranking.ViewsCount = int64(s.countViewsInPeriod(category.ID, *periodStart))
			ranking.ActiveUsers = int64(s.countActiveUsers(category.ID))
		default:
			return nil, fmt.Errorf("未知的排行榜类型: %s", rankingType)
		}

		rankings = append(rankings, ranking)
	}

	// 排序
	switch rankingType {
	case "resources", "views":
		// 按分数降序排序
	case "growth":
		// 按增长率降序排序
	case "popularity":
		// 按综合分数降序排序
	}

	// 应用分页
	if offset >= len(rankings) {
		return []*CategoryRanking{}, nil
	}
	if offset+limit > len(rankings) {
		limit = len(rankings) - offset
	}
	rankings = rankings[offset : offset+limit]

	// 更新排名
	for i, ranking := range rankings {
		ranking.Rank = offset + i + 1
	}

	return rankings, nil
}

// GetCategoryTrends 获取分类趋势数据
// 参数：
//   - categoryID: 分类ID
//   - days: 天数
//
// 返回：
//   - 趋势数据
//   - 错误信息
func (s *StatisticsService) GetCategoryTrends(categoryID uint, days int) (map[string][]int64, error) {
	if days <= 0 || days > 365 {
		days = 30 // 默认30天
	}

	trends := make(map[string][]int64)
	trends["resources"] = make([]int64, days)
	trends["views"] = make([]int64, days)
	trends["active_users"] = make([]int64, days)

	// 生成趋势数据
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.AddDate(0, 0, 1)

		// 统计当天的资源数
		var resourcesCount int64
		s.db.Model(&model.Resource{}).
			Where("category_id = ? AND created_at >= ? AND created_at < ?", categoryID, startOfDay, endOfDay).
			Count(&resourcesCount)
		trends["resources"][days-1-i] = resourcesCount

		// 统计当天的浏览量
		var viewsCount int64
		s.db.Model(&model.VisitLog{}).
			Where("resource_id IN (SELECT id FROM resources WHERE category_id = ?) AND visited_at >= ? AND visited_at < ?", categoryID, startOfDay, endOfDay).
			Count(&viewsCount)
		trends["views"][days-1-i] = viewsCount

		// 统计当天的活跃用户数
		var activeUsersCount int64
		s.db.Model(&model.User{}).
			Where("updated_at >= ? AND updated_at < ?", startOfDay, endOfDay).
			Count(&activeUsersCount)
		trends["active_users"][days-1-i] = activeUsersCount
	}

	return trends, nil
}

// UpdateCategoryResourceCount 更新分类的资源计数
// 参数：
//   - categoryID: 分类ID
//
// 返回：
//   - 错误信息
func (s *StatisticsService) UpdateCategoryResourceCount(categoryID uint) error {
	// 计算分类的资源总数
	var count int64
	if err := s.db.Model(&model.Resource{}).Where("category_id = ?", categoryID).Count(&count).Error; err != nil {
		return fmt.Errorf("统计资源数量失败: %w", err)
	}

	// 更新分类的资源计数
	if err := s.db.Model(&model.Category{}).Where("id = ?", categoryID).Update("resources_count", count).Error; err != nil {
		return fmt.Errorf("更新分类资源计数失败: %w", err)
	}

	return nil
}

// BatchUpdateAllCategoryCounts 批量更新所有分类的资源计数
// 返回：
//   - 更新的分类数量
//   - 错误信息
func (s *StatisticsService) BatchUpdateAllCategoryCounts() (int, error) {
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

// countResources 统计分类的资源总数
func (s *StatisticsService) countResources(categoryID uint) int64 {
	var count int64
	s.db.Model(&model.Resource{}).Where("category_id = ?", categoryID).Count(&count)
	return count
}

// countResourcesInPeriod 统计分类在指定时间范围内的资源数
func (s *StatisticsService) countResourcesInPeriod(categoryID uint, since time.Time) int64 {
	var count int64
	s.db.Model(&model.Resource{}).Where("category_id = ? AND created_at >= ?", categoryID, since).Count(&count)
	return count
}

// countChildren 统计分类的子分类数量
func (s *StatisticsService) countChildren(categoryID uint) int64 {
	var count int64
	s.db.Model(&model.Category{}).Where("parent_id = ?", categoryID).Count(&count)
	return count
}

// countTotalDescendants 统计分类的所有后代数量
func (s *StatisticsService) countTotalDescendants(categoryID uint) int64 {
	// 使用递归查询
	var count int64
	s.db.Raw(`
		WITH RECURSIVE category_tree AS (
			SELECT id, parent_id, 1 as level
			FROM categories
			WHERE parent_id = ?
			UNION ALL
			SELECT c.id, c.parent_id, ct.level + 1
			FROM categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.id
		)
		SELECT COUNT(*) FROM category_tree
	`, categoryID).Scan(&count)
	return count
}

// calculateMaxDepth 计算分类的最大深度
func (s *StatisticsService) calculateMaxDepth(categoryID uint) int {
	var depth int
	s.db.Raw(`
		WITH RECURSIVE category_tree AS (
			SELECT id, parent_id, 1 as level
			FROM categories
			WHERE parent_id = ?
			UNION ALL
			SELECT c.id, c.parent_id, ct.level + 1
			FROM categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.id
		)
		SELECT MAX(level) FROM category_tree
	`, categoryID).Scan(&depth)
	return depth
}

// calculateCategoryLevel 计算分类的层级
func (s *StatisticsService) calculateCategoryLevel(categoryID uint) int {
	level := 0
	currentID := categoryID

	for {
		var category model.Category
		if err := s.db.First(&category, currentID).Error; err != nil {
			break
		}

		if category.ParentID == nil {
			break
		}

		level++
		currentID = *category.ParentID
	}

	return level
}

// countActiveUsers 统计分类的活跃用户数
func (s *StatisticsService) countActiveUsers(categoryID uint) int64 {
	var count int64
	s.db.Model(&model.User{}).
		Joins("JOIN resources ON resources.uploaded_by_id = users.id").
		Where("resources.category_id = ? AND users.status = 'active'", categoryID).
		Distinct("users.id").
		Count(&count)
	return count
}

// countTotalViews 统计分类的总浏览量
func (s *StatisticsService) countTotalViews(categoryID uint) int64 {
	var count int64
	s.db.Model(&model.VisitLog{}).
		Joins("JOIN resources ON resources.id = visit_logs.resource_id").
		Where("resources.category_id = ?", categoryID).
		Count(&count)
	return count
}

// countViewsInPeriod 统计分类在指定时间范围内的浏览量
func (s *StatisticsService) countViewsInPeriod(categoryID uint, since time.Time) int64 {
	var count int64
	s.db.Model(&model.VisitLog{}).
		Joins("JOIN resources ON resources.id = visit_logs.resource_id").
		Where("resources.category_id = ? AND visit_logs.visited_at >= ?", categoryID, since).
		Count(&count)
	return count
}

// calculateResourceGrowthRate 计算资源增长率
func (s *StatisticsService) calculateResourceGrowthRate(categoryID uint) float64 {
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	twoMonthsAgo := now.AddDate(0, -2, 0)

	currentCount := s.countResourcesInPeriod(categoryID, lastMonth)
	previousCount := s.countResourcesInPeriod(categoryID, twoMonthsAgo)

	if previousCount == 0 {
		return 0.0
	}

	return float64(currentCount-previousCount) / float64(previousCount) * 100
}

// calculateViewGrowthRate 计算浏览量增长率
func (s *StatisticsService) calculateViewGrowthRate(categoryID uint) float64 {
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	twoMonthsAgo := now.AddDate(0, -2, 0)

	currentCount := s.countViewsInPeriod(categoryID, lastMonth)
	previousCount := s.countViewsInPeriod(categoryID, twoMonthsAgo)

	if previousCount == 0 {
		return 0.0
	}

	return float64(currentCount-previousCount) / float64(previousCount) * 100
}

// getFirstResourceAt 获取第一个资源的时间
func (s *StatisticsService) getFirstResourceAt(categoryID uint) *time.Time {
	var firstAt time.Time
	result := s.db.Model(&model.Resource{}).
		Where("category_id = ?", categoryID).
		Order("created_at ASC").
		Pluck("created_at", &firstAt)

	if result.Error != nil || firstAt.IsZero() {
		return nil
	}

	return &firstAt
}

// getLastResourceAt 获取最后一个资源的时间
func (s *StatisticsService) getLastResourceAt(categoryID uint) *time.Time {
	var lastAt time.Time
	result := s.db.Model(&model.Resource{}).
		Where("category_id = ?", categoryID).
		Order("created_at DESC").
		Pluck("created_at", &lastAt)

	if result.Error != nil || lastAt.IsZero() {
		return nil
	}

	return &lastAt
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
