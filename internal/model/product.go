package model

import (
	"time"
)

// Product 商品模型
type Product struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Title           string    `json:"title" gorm:"size:255;not null"`
	Image           string    `json:"image" gorm:"size:255;not null"`
	HeadLabel       string    `json:"headLabel" gorm:"column:head_label;size:50"`
	OriginalPrice   float64   `json:"originalPrice" gorm:"column:original_price;type:decimal(10,2);not null"`
	Price           float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	ShopID          uint      `json:"shopId" gorm:"column:shop_id;not null"`
	Description     string    `json:"description" gorm:"type:text"`
	CommentCount    int       `json:"commentCount" gorm:"column:comment_count;default:0"`
	GoodCommentRate string    `json:"goodCommentRate" gorm:"column:good_comment_rate;size:10;default:'0%'"`
	CreatedAt       time.Time `json:"createdAt" gorm:"not null"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"not null"`
	
	// 关联
	Shop          Shop            `json:"shop" gorm:"foreignKey:ShopID"`
	Images        []ProductImage  `json:"images" gorm:"foreignKey:ProductID"`
	Labels        []ProductLabel  `json:"labels" gorm:"foreignKey:ProductID"`
	Specifications []ProductSpec  `json:"specifications" gorm:"foreignKey:ProductID"`
	Services      []ProductService `json:"services" gorm:"foreignKey:ProductID"`
}

// ProductImage 商品图片模型
type ProductImage struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	ProductID uint   `json:"productId" gorm:"column:product_id;not null"`
	ImageURL  string `json:"imageUrl" gorm:"column:image_url;size:255;not null"`
	SortOrder int    `json:"sortOrder" gorm:"column:sort_order;default:0"`
}

// ProductLabel 商品标签模型
type ProductLabel struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	ProductID    uint   `json:"productId" gorm:"column:product_id;not null"`
	LabelContent string `json:"labelContent" gorm:"column:label_content;size:50;not null"`
}

// ProductSpec 商品规格模型
type ProductSpec struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	ProductID uint   `json:"productId" gorm:"column:product_id;not null"`
	Name      string `json:"name" gorm:"size:50;not null"`
	
	// 关联
	Options []SpecOption `json:"options" gorm:"foreignKey:SpecID"`
}

// SpecOption 规格选项模型
type SpecOption struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	SpecID      uint   `json:"specId" gorm:"column:spec_id;not null"`
	OptionValue string `json:"optionValue" gorm:"column:option_value;size:50;not null"`
}

// ProductService 商品服务模型
type ProductService struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	ProductID      uint   `json:"productId" gorm:"column:product_id;not null"`
	ServiceContent string `json:"serviceContent" gorm:"column:service_content;size:50;not null"`
}

// ProductResponse 商品列表响应项
type ProductResponse struct {
	ID            uint       `json:"id"`
	Image         string     `json:"image"`
	Title         string     `json:"title"`
	HeadLabel     string     `json:"headLabel"`
	Labels        []string   `json:"labels"`
	OriginalPrice string     `json:"originalPrice"`
	Price         string     `json:"price"`
	ShopInfo      ShopBrief  `json:"shopInfo"`
}

// ProductDetailResponse 商品详情响应
type ProductDetailResponse struct {
	ID              uint              `json:"id"`
	Image           string            `json:"image"`
	ImageList       []string          `json:"imageList"`
	Title           string            `json:"title"`
	HeadLabel       string            `json:"headLabel"`
	Labels          []string          `json:"labels"`
	OriginalPrice   string            `json:"originalPrice"`
	Price           string            `json:"price"`
	ShopInfo        ShopDetail        `json:"shopInfo"`
	Specifications  []SpecResponse    `json:"specifications"`
	Description     string            `json:"description"`
	Services        []string          `json:"services"`
	CommentCount    int               `json:"commentCount"`
	GoodCommentRate string            `json:"goodCommentRate"`
}

// SpecResponse 规格响应
type SpecResponse struct {
	Name    string   `json:"name"`
	Options []string `json:"options"`
} 