/*
Package model defines all data models for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package model

import (
	"time"

	"gorm.io/gorm"
)

// ProductStatus 商品状态枚举
type ProductStatus string

const (
	ProductStatusActive     ProductStatus = "active"       // 上架
	ProductStatusInactive   ProductStatus = "inactive"     // 下架
	ProductStatusOutOfStock ProductStatus = "out_of_stock" // 缺货
)

// ProductCategory 商品分类枚举
type ProductCategory string

const (
	ProductCategoryVip      ProductCategory = "vip"      // VIP会员
	ProductCategoryResource ProductCategory = "resource" // 资源包
	ProductCategoryService  ProductCategory = "service"  // 服务
	ProductCategoryGift     ProductCategory = "gift"     // 礼品
)

// Product 商品模型
type Product struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 商品基本信息
	Name        string          `gorm:"not null;size:100" json:"name"`
	Description string          `gorm:"size:255" json:"description"`
	Category    ProductCategory `gorm:"not null;size:20" json:"category"`

	// 价格信息
	PointsPrice   int `gorm:"not null" json:"points_price"`    // 积分价格
	OriginalPrice int `gorm:"default:0" json:"original_price"` // 原价（虚拟）

	// 库存信息
	Stock     int  `gorm:"default:0" json:"stock"`
	IsLimited bool `gorm:"default:false" json:"is_limited"` // 是否限量

	// 状态信息
	Status ProductStatus `gorm:"default:'active';not null;size:20" json:"status"`

	// 销售信息
	SalesCount int `gorm:"default:0" json:"sales_count"` // 销售数量

	// 有效期（VIP等）
	ValidDays *int `gorm:"json:"valid_days"` // 有效天数，VIP使用

	// 关联关系
	Orders []MallOrder `gorm:"foreignKey:ProductID" json:"-"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

// BeforeCreate 创建钩子
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	// 设置默认状态为上架
	if p.Status == "" {
		p.Status = ProductStatusActive
	}

	return nil
}

// OrderStatus 订单状态枚举
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"   // 待支付
	OrderStatusPaid      OrderStatus = "paid"      // 已支付
	OrderStatusCompleted OrderStatus = "completed" // 已完成
	OrderStatusCancelled OrderStatus = "cancelled" // 已取消
	OrderStatusRefunded  OrderStatus = "refunded"  // 已退款
)

// MallOrder 积分商城订单模型
type MallOrder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 订单基本信息
	OrderNo string `gorm:"uniqueIndex;not null;size:32" json:"order_no"` // 订单号
	UserID  uint   `gorm:"not null;index" json:"user_id"`
	User    *User  `gorm:"foreignKey:UserID" json:"user"`

	// 商品信息
	ProductID uint     `gorm:"not null;index" json:"product_id"`
	Product   *Product `gorm:"foreignKey:ProductID" json:"product"`

	// 数量和价格
	Quantity   int `gorm:"not null;default:1" json:"quantity"`
	PointsCost int `gorm:"not null" json:"points_cost"` // 所需积分

	// 订单状态
	Status OrderStatus `gorm:"default:'pending';not null;size:20" json:"status"`

	// 完成信息
	CompletedAt *time.Time `json:"completed_at"`

	// 备注
	Note string `gorm:"size:255" json:"note"`
}

// TableName 指定表名
func (MallOrder) TableName() string {
	return "mall_orders"
}

// BeforeCreate 创建钩子
func (o *MallOrder) BeforeCreate(tx *gorm.DB) error {
	// 生成订单号
	if o.OrderNo == "" {
		o.OrderNo = generateOrderNo()
	}

	// 设置默认状态为待支付
	if o.Status == "" {
		o.Status = OrderStatusPending
	}

	return nil
}

// 生成订单号
func generateOrderNo() string {
	return "MALL" + time.Now().Format("20060102150405")
}
