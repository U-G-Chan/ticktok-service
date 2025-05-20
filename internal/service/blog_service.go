package service

import (
	"ticktok-service/internal/model"
	"ticktok-service/internal/repository"
)

// BlogService 博客服务接口
type BlogService interface {
	GetBlogs(page, pageSize int) ([]model.Blog, int64, error)
	GetBlogByID(id uint) (*model.Blog, error)
	SearchBlogs(keyword string, page, pageSize int) ([]model.Blog, int64, error)
}

// blogService 博客服务实现
type blogService struct {
	blogRepo repository.BlogRepository
}

// NewBlogService 创建新的博客服务
func NewBlogService() BlogService {
	return &blogService{
		blogRepo: repository.NewBlogRepository(),
	}
}

// GetBlogs 获取博客列表
func (s *blogService) GetBlogs(page, pageSize int) ([]model.Blog, int64, error) {
	return s.blogRepo.GetBlogs(page, pageSize)
}

// GetBlogByID 根据ID获取博客
func (s *blogService) GetBlogByID(id uint) (*model.Blog, error) {
	return s.blogRepo.GetBlogByID(id)
}

// SearchBlogs 搜索博客
func (s *blogService) SearchBlogs(keyword string, page, pageSize int) ([]model.Blog, int64, error) {
	return s.blogRepo.SearchBlogs(keyword, page, pageSize)
} 