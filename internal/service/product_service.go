package service

import (
	"fmt"
	"ticktok-service/internal/model"
	"ticktok-service/internal/repository"

	"gorm.io/gorm"
)

// ProductService 商品服务接口
type ProductService interface {
	GetProductList(page, pageSize int) (*model.PageResult, error)
	GetProductDetail(id uint) (*model.ProductDetailResponse, error)
	GetLabels() (*model.LabelResponse, error)
}

// productService 商品服务实现
type productService struct {
	productRepo repository.ProductRepository
}

// NewProductService 创建商品服务
func NewProductService(db *gorm.DB) ProductService {
	return &productService{
		productRepo: repository.NewProductRepository(db),
	}
}

// GetProductList 获取商品列表
func (s *productService) GetProductList(page, pageSize int) (*model.PageResult, error) {
	// 获取商品数据
	products, total, err := s.productRepo.GetProducts(page, pageSize)
	if err != nil {
		return nil, err
	}
	
	// 转换为响应格式
	var productResponses []*model.ProductResponse
	for _, product := range products {
		// 收集标签
		var labels []string
		for _, label := range product.Labels {
			labels = append(labels, label.LabelContent)
		}
		
		// 创建响应
		response := &model.ProductResponse{
			ID:            product.ID,
			Image:         product.Image,
			Title:         product.Title,
			HeadLabel:     product.HeadLabel,
			Labels:        labels,
			OriginalPrice: fmt.Sprintf("%.2f", product.OriginalPrice),
			Price:         fmt.Sprintf("%.2f", product.Price),
			ShopInfo: model.ShopBrief{
				Name:  product.Shop.Name,
				Sales: product.Shop.Sales,
			},
		}
		productResponses = append(productResponses, response)
	}
	
	// 计算是否有更多数据
	hasMore := int64(page*pageSize) < total
	
	// 创建分页结果
	result := &model.PageResult{
		List:     productResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	}
	
	return result, nil
}

// GetProductDetail 获取商品详情
func (s *productService) GetProductDetail(id uint) (*model.ProductDetailResponse, error) {
	// 获取商品详情
	product, err := s.productRepo.GetProductByID(id)
	if err != nil {
		return nil, err
	}
	
	// 收集图片
	var imageList []string
	for _, img := range product.Images {
		imageList = append(imageList, img.ImageURL)
	}
	
	// 收集标签
	var labels []string
	for _, label := range product.Labels {
		labels = append(labels, label.LabelContent)
	}
	
	// 收集规格
	var specifications []model.SpecResponse
	for _, spec := range product.Specifications {
		var options []string
		for _, opt := range spec.Options {
			options = append(options, opt.OptionValue)
		}
		specifications = append(specifications, model.SpecResponse{
			Name:    spec.Name,
			Options: options,
		})
	}
	
	// 收集服务
	var services []string
	for _, service := range product.Services {
		services = append(services, service.ServiceContent)
	}
	
	// 创建响应
	response := &model.ProductDetailResponse{
		ID:              product.ID,
		Image:           product.Image,
		ImageList:       imageList,
		Title:           product.Title,
		HeadLabel:       product.HeadLabel,
		Labels:          labels,
		OriginalPrice:   fmt.Sprintf("%.2f", product.OriginalPrice),
		Price:           fmt.Sprintf("%.2f", product.Price),
		Description:     product.Description,
		CommentCount:    product.CommentCount,
		GoodCommentRate: product.GoodCommentRate,
		ShopInfo: model.ShopDetail{
			ID:        product.Shop.ID,
			Name:      product.Shop.Name,
			Sales:     product.Shop.Sales,
			Logo:      product.Shop.Logo,
			Rating:    product.Shop.Rating,
			Followers: product.Shop.Followers,
		},
		Specifications: specifications,
		Services:       services,
	}
	
	return response, nil
}

// GetLabels 获取所有标签
func (s *productService) GetLabels() (*model.LabelResponse, error) {
	// 获取标签数据
	labels, brandLabels, err := s.productRepo.GetAllLabels()
	if err != nil {
		return nil, err
	}
	
	// 创建响应
	response := &model.LabelResponse{
		Labels:      labels,
		BrandLabels: brandLabels,
	}
	
	return response, nil
} 