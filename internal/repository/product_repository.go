package repository

import (
	"ticktok-service/internal/model"

	"gorm.io/gorm"
)

// ProductRepository 商品数据仓库接口
type ProductRepository interface {
	GetProducts(page, pageSize int) ([]*model.Product, int64, error)
	GetProductByID(id uint) (*model.Product, error)
	GetAllLabels() ([]string, []string, error)
}

// productRepository 商品数据仓库实现
type productRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建商品数据仓库
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{
		db: db,
	}
}

// GetProducts 获取商品列表
func (r *productRepository) GetProducts(page, pageSize int) ([]*model.Product, int64, error) {
	var products []*model.Product
	var total int64
	
	offset := (page - 1) * pageSize
	
	// 查询商品总数
	if err := r.db.Model(&model.Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 查询商品列表，预加载关联数据
	if err := r.db.Preload("Shop").
		Preload("Labels").
		Offset(offset).
		Limit(pageSize).
		Order("id DESC").
		Find(&products).Error; err != nil {
		return nil, 0, err
	}
	
	return products, total, nil
}

// GetProductByID 根据ID获取商品详情
func (r *productRepository) GetProductByID(id uint) (*model.Product, error) {
	var product model.Product
	
	// 查询商品详情，预加载所有关联数据
	if err := r.db.Preload("Shop").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Labels").
		Preload("Specifications").
		Preload("Specifications.Options").
		Preload("Services").
		First(&product, id).Error; err != nil {
		return nil, err
	}
	
	return &product, nil
}

// GetAllLabels 获取所有标签
func (r *productRepository) GetAllLabels() ([]string, []string, error) {
	// 获取普通标签
	var labels []string
	if err := r.db.Model(&model.ProductLabel{}).
		Select("DISTINCT label_content").
		Where("label_content NOT LIKE ?", "%品牌%"). // 排除品牌标签
		Pluck("label_content", &labels).Error; err != nil {
		return nil, nil, err
	}
	
	// 获取品牌标签
	var brandLabels []string
	if err := r.db.Model(&model.ProductLabel{}).
		Select("DISTINCT label_content").
		Where("label_content LIKE ?", "%品牌%").
		Or("label_content LIKE ?", "%旗舰%").
		Or("label_content LIKE ?", "%官方%").
		Or("label_content LIKE ?", "%好店%").
		Or("label_content LIKE ?", "%精选%").
		Pluck("label_content", &brandLabels).Error; err != nil {
		return nil, nil, err
	}
	
	return labels, brandLabels, nil
} 