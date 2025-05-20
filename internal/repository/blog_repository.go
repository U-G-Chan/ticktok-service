package repository

import (
	"fmt"
	"ticktok-service/internal/model"
)

// BlogRepository 博客仓库接口
type BlogRepository interface {
	GetBlogs(page, pageSize int) ([]model.Blog, int64, error)
	GetBlogByID(id uint) (*model.Blog, error)
	SearchBlogs(keyword string, page, pageSize int) ([]model.Blog, int64, error)
}

// blogRepository 博客仓库实现
type blogRepository struct{}

// NewBlogRepository 创建新的博客仓库
func NewBlogRepository() BlogRepository {
	return &blogRepository{}
}

// GetBlogs 获取博客列表
func (r *blogRepository) GetBlogs(page, pageSize int) ([]model.Blog, int64, error) {
	var blogs []model.Blog
	var count int64

	// 获取总数
	if err := model.DB.Model(&model.Blog{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("计算博客总数失败: %w", err)
	}

	// 分页查询博客
	offset := (page - 1) * pageSize
	if err := model.DB.Offset(offset).Limit(pageSize).Find(&blogs).Error; err != nil {
		return nil, 0, fmt.Errorf("获取博客列表失败: %w", err)
	}

	// 加载关联数据
	for i := range blogs {
		model.DB.Model(&blogs[i]).Association("Images").Find(&blogs[i].Images)
		model.DB.Model(&blogs[i]).Association("Tags").Find(&blogs[i].Tags)
		model.DB.Model(&blogs[i]).Association("Comments").Find(&blogs[i].Comments)
	}

	return blogs, count, nil
}

// GetBlogByID 根据ID获取博客
func (r *blogRepository) GetBlogByID(id uint) (*model.Blog, error) {
	var blog model.Blog
	if err := model.DB.First(&blog, id).Error; err != nil {
		return nil, fmt.Errorf("获取博客详情失败: %w", err)
	}

	// 加载关联数据
	model.DB.Model(&blog).Association("Images").Find(&blog.Images)
	model.DB.Model(&blog).Association("Tags").Find(&blog.Tags)
	model.DB.Model(&blog).Association("Comments").Find(&blog.Comments)

	return &blog, nil
}

// SearchBlogs 搜索博客
func (r *blogRepository) SearchBlogs(keyword string, page, pageSize int) ([]model.Blog, int64, error) {
	var blogs []model.Blog
	var count int64

	query := model.DB.Model(&model.Blog{}).Where(
		"title LIKE ? OR content LIKE ?", 
		"%"+keyword+"%", 
		"%"+keyword+"%",
	)

	// 获取总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("计算搜索结果总数失败: %w", err)
	}

	// 分页查询博客
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&blogs).Error; err != nil {
		return nil, 0, fmt.Errorf("搜索博客失败: %w", err)
	}

	// 加载关联数据
	for i := range blogs {
		model.DB.Model(&blogs[i]).Association("Images").Find(&blogs[i].Images)
		model.DB.Model(&blogs[i]).Association("Tags").Find(&blogs[i].Tags)
		model.DB.Model(&blogs[i]).Association("Comments").Find(&blogs[i].Comments)
	}

	return blogs, count, nil
} 