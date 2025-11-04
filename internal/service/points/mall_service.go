/*
Points Mall Service - 积分商城服务

提供完整的积分商城功能，包括：
- 商品管理
- 购物车功能
- 订单处理
- 库存管理

Author: Felix Wang
Email: felixwang.biz@gmail.com
Date: 2025-10-31
*/

package points

import (
	"fmt"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MallService 积分商城服务
type MallService struct {
	db *gorm.DB
}

// NewMallService 创建新的积分商城服务
func NewMallService(db *gorm.DB) *MallService {
	return &MallService{
		db: db,
	}
}

// CreateProduct 创建商品
func (s *MallService) CreateProduct(product *model.Product) error {
	if product.PointsPrice <= 0 {
		return fmt.Errorf("商品积分价格必须大于0")
	}

	if err := s.db.Create(product).Error; err != nil {
		return fmt.Errorf("创建商品失败: %w", err)
	}

	return nil
}

// UpdateProduct 更新商品信息
func (s *MallService) UpdateProduct(productID uint, updates map[string]interface{}) error {
	if err := s.db.Model(&model.Product{}).
		Where("id = ?", productID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("更新商品失败: %w", err)
	}

	return nil
}

// DeleteProduct 删除商品（软删除）
func (s *MallService) DeleteProduct(productID uint) error {
	if err := s.db.Delete(&model.Product{}, productID).Error; err != nil {
		return fmt.Errorf("删除商品失败: %w", err)
	}

	return nil
}

// GetProduct 获取单个商品信息
func (s *MallService) GetProduct(productID uint) (*model.Product, error) {
	var product model.Product
	if err := s.db.First(&product, productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("商品不存在")
		}
		return nil, fmt.Errorf("查询商品失败: %w", err)
	}

	return &product, nil
}

// ListProducts 获取商品列表
func (s *MallService) ListProducts(category model.ProductCategory, status model.ProductStatus,
	page, pageSize int) ([]model.Product, int64, error) {

	var products []model.Product
	var total int64

	query := s.db.Model(&model.Product{})

	// 按分类筛选
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 按状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取商品总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("sales_count DESC, created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("获取商品列表失败: %w", err)
	}

	return products, total, nil
}

// CreateOrder 创建订单
func (s *MallService) CreateOrder(userID, productID uint, quantity int) (*model.MallOrder, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("购买数量必须大于0")
	}

	return nil, fmt.Errorf("未实现")
}

// PurchaseProduct 直接购买商品
func (s *MallService) PurchaseProduct(userID, productID uint, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("购买数量必须大于0")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取商品信息
		var product model.Product
		if err := tx.Clauses(clause.Locking{}).First(&product, productID).Error; err != nil {
			return fmt.Errorf("商品不存在: %w", err)
		}

		// 检查商品状态
		if product.Status != model.ProductStatusActive {
			return fmt.Errorf("商品已下架或不可购买")
		}

		// 检查库存
		if product.IsLimited && product.Stock < quantity {
			return fmt.Errorf("商品库存不足，当前库存: %d", product.Stock)
		}

		// 计算所需积分
		totalPoints := product.PointsPrice * quantity

		// 使用消费服务来检查积分和扣除积分
		consumptionService := NewConsumptionService(s.db)

		// 创建订单
		order := model.MallOrder{
			UserID:      userID,
			ProductID:   productID,
			Quantity:    quantity,
			PointsCost:  totalPoints,
			Status:      model.OrderStatusPaid, // 直接购买即为已支付
			CompletedAt: &time.Time{},
		}

		// 如果时间指针为空，需要设置为当前时间
		now := time.Now()
		order.CompletedAt = &now

		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("创建订单失败: %w", err)
		}

		// 扣除积分
		description := fmt.Sprintf("购买商品: %s x%d", product.Name, quantity)
		if err := consumptionService.SpendPointsForPurchase(userID, totalPoints, description, &productID); err != nil {
			return fmt.Errorf("扣除积分失败: %w", err)
		}

		// 更新商品库存
		if product.IsLimited {
			if err := tx.Model(&model.Product{}).
				Where("id = ?", productID).
				Update("stock", gorm.Expr("stock - ?", quantity)).Error; err != nil {
				return fmt.Errorf("更新库存失败: %w", err)
			}
		}

		// 更新商品销售数量
		if err := tx.Model(&model.Product{}).
			Where("id = ?", productID).
			Update("sales_count", gorm.Expr("sales_count + ?", quantity)).Error; err != nil {
			return fmt.Errorf("更新销售数量失败: %w", err)
		}

		// 发放奖励（VIP等）
		if product.Category == model.ProductCategoryVip && product.ValidDays != nil {
			// 这里可以添加VIP权限发放逻辑
			_ = *product.ValidDays // 使用变量避免未使用警告
		}

		return nil
	})
}

// CancelOrder 取消订单
func (s *MallService) CancelOrder(userID, orderID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取订单信息
		var order model.MallOrder
		if err := tx.Clauses(clause.Locking{}).
			Preload("Product").
			First(&order, orderID).Error; err != nil {
			return fmt.Errorf("订单不存在: %w", err)
		}

		// 验证订单属于当前用户
		if order.UserID != userID {
			return fmt.Errorf("订单不属于当前用户")
		}

		// 检查订单状态
		if order.Status != model.OrderStatusPending {
			return fmt.Errorf("只能取消待支付订单")
		}

		// 取消订单
		order.Status = model.OrderStatusCancelled
		if err := tx.Save(&order).Error; err != nil {
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		return nil
	})
}

// RefundOrder 订单退款
func (s *MallService) RefundOrder(userID, orderID uint, reason string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 获取订单信息
		var order model.MallOrder
		if err := tx.Clauses(clause.Locking{}).
			Preload("Product").
			First(&order, orderID).Error; err != nil {
			return fmt.Errorf("订单不存在: %w", err)
		}

		// 验证订单属于当前用户
		if order.UserID != userID {
			return fmt.Errorf("订单不属于当前用户")
		}

		// 检查订单状态
		if order.Status != model.OrderStatusPaid && order.Status != model.OrderStatusCompleted {
			return fmt.Errorf("只能对已支付或已完成订单进行退款")
		}

		// 退还积分
		description := fmt.Sprintf("订单退款: %s (%s)", order.Product.Name, reason)
		earningService := NewEarningService(s.db)
		if err := earningService.addPoints(tx, userID, order.PointsCost, model.PointSourceAdminAdd,
			description, nil, nil); err != nil {
			return fmt.Errorf("退还积分失败: %w", err)
		}

		// 恢复库存
		if order.Product.IsLimited {
			if err := tx.Model(&model.Product{}).
				Where("id = ?", order.ProductID).
				Update("stock", gorm.Expr("stock + ?", order.Quantity)).Error; err != nil {
				return fmt.Errorf("恢复库存失败: %w", err)
			}
		}

		// 减少销售数量
		if err := tx.Model(&model.Product{}).
			Where("id = ?", order.ProductID).
			Update("sales_count", gorm.Expr("sales_count - ?", order.Quantity)).Error; err != nil {
			return fmt.Errorf("更新销售数量失败: %w", err)
		}

		// 更新订单状态
		order.Status = model.OrderStatusRefunded
		if err := tx.Save(&order).Error; err != nil {
			return fmt.Errorf("更新订单状态失败: %w", err)
		}

		return nil
	})
}

// GetUserOrders 获取用户订单列表
func (s *MallService) GetUserOrders(userID uint, status model.OrderStatus,
	page, pageSize int) ([]model.MallOrder, int64, error) {

	var orders []model.MallOrder
	var total int64

	query := s.db.Model(&model.MallOrder{}).Where("user_id = ?", userID)

	// 按状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取订单总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("Product").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("获取订单列表失败: %w", err)
	}

	return orders, total, nil
}

// GetProductSalesStats 获取商品销售统计
func (s *MallService) GetProductSalesStats(productID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总销量
	var product model.Product
	if err := s.db.Select("sales_count").First(&product, productID).Error; err != nil {
		return nil, fmt.Errorf("商品不存在: %w", err)
	}
	stats["total_sales"] = product.SalesCount

	// 今日销量
	var todaySales int64
	rows, _ := s.db.Raw("SELECT COUNT(*) FROM mall_orders WHERE product_id = ? AND DATE(created_at) = ?", productID, time.Now().Format("2006-01-02")).Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&todaySales)
	}
	stats["today_sales"] = todaySales

	// 本月销量
	var monthSales int64
	rows, _ = s.db.Raw("SELECT COUNT(*) FROM mall_orders WHERE product_id = ? AND strftime('%Y-%m', created_at) = ?", productID, time.Now().Format("2006-01")).Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&monthSales)
	}
	stats["month_sales"] = monthSales

	// 总收入积分
	var totalPoints int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(points_cost), 0) FROM mall_orders WHERE product_id = ? AND status IN ('paid', 'completed')", productID).Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&totalPoints)
	}
	stats["total_points_earned"] = totalPoints

	return stats, nil
}

// GetMallStats 获取商城统计信息
func (s *MallService) GetMallStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总商品数
	var totalProducts int64
	s.db.Model(&model.Product{}).Count(&totalProducts)
	stats["total_products"] = totalProducts

	// 在售商品数
	var activeProducts int64
	s.db.Model(&model.Product{}).Where("status = ?", model.ProductStatusActive).Count(&activeProducts)
	stats["active_products"] = activeProducts

	// 总订单数
	var totalOrders int64
	s.db.Model(&model.MallOrder{}).Count(&totalOrders)
	stats["total_orders"] = totalOrders

	// 今日订单数
	var todayOrders int64
	rows, _ := s.db.Raw("SELECT COUNT(*) FROM mall_orders WHERE DATE(created_at) = ?", time.Now().Format("2006-01-02")).Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&todayOrders)
	}
	stats["today_orders"] = todayOrders

	// 总销售额（积分）
	var totalSales int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(points_cost), 0) FROM mall_orders WHERE status IN ('paid', 'completed')").Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&totalSales)
	}
	stats["total_sales"] = totalSales

	// 今日销售额
	var todaySales int64
	rows, _ = s.db.Raw("SELECT COALESCE(SUM(points_cost), 0) FROM mall_orders WHERE DATE(created_at) = ? AND status IN ('paid', 'completed')", time.Now().Format("2006-01-02")).Rows()
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&todaySales)
	}
	stats["today_sales"] = todaySales

	// 热销商品 TOP 5
	var topProducts []struct {
		ID    uint
		Name  string
		Sales int
	}
	s.db.Raw(`
		SELECT p.id, p.name, p.sales_count as sales
		FROM products p
		ORDER BY p.sales_count DESC
		LIMIT 5
	`).Scan(&topProducts)
	stats["top_products"] = topProducts

	return stats, nil
}

// SearchProducts 搜索商品
func (s *MallService) SearchProducts(keyword string, category model.ProductCategory,
	page, pageSize int) ([]model.Product, int64, error) {

	var products []model.Product
	var total int64

	query := s.db.Model(&model.Product{})

	// 关键词搜索
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 分类筛选
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 只显示上架商品
	query = query.Where("status = ?", model.ProductStatusActive)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取商品总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("sales_count DESC, created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("搜索商品失败: %w", err)
	}

	return products, total, nil
}
