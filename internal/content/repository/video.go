package repository

import (
	"context"
	"ticktok-service/internal/content/model"

	"gorm.io/gorm"
)

type VideoRepository interface {
	Create(ctx context.Context, video *model.Video) error
	Update(ctx context.Context, video *model.Video) error
	UpdateInteractionCounts(ctx context.Context, videoID int64, favoriteCount, commentCount int64) error
	GetByID(ctx context.Context, id int64) (*model.Video, error)
	GetByIDs(ctx context.Context, ids []int64) ([]*model.Video, error)
	GetFeed(ctx context.Context, lastScore int32, lastID int64, limit int) ([]*model.Video, error)
	GetByAuthorID(ctx context.Context, authorID int64) ([]*model.Video, error)
}

type videoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) VideoRepository {
	return &videoRepository{db: db}
}

func (r *videoRepository) Create(ctx context.Context, video *model.Video) error {
	return r.db.WithContext(ctx).Create(video).Error
}

func (r *videoRepository) Update(ctx context.Context, video *model.Video) error {
	return r.db.WithContext(ctx).Save(video).Error
}

func (r *videoRepository) UpdateInteractionCounts(ctx context.Context, videoID int64, favoriteCount, commentCount int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Video{}).
		Where("id = ?", videoID).
		Updates(map[string]interface{}{
			"favorite_count": favoriteCount,
			"comment_count":  commentCount,
		}).Error
}

func (r *videoRepository) GetByID(ctx context.Context, id int64) (*model.Video, error) {
	var video model.Video
	err := r.db.WithContext(ctx).First(&video, id).Error
	return &video, err
}

func (r *videoRepository) GetByIDs(ctx context.Context, ids []int64) ([]*model.Video, error) {
	var videos []*model.Video
	if len(ids) == 0 {
		return videos, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&videos).Error
	return videos, err
}

func (r *videoRepository) GetFeed(ctx context.Context, lastScore int32, lastID int64, limit int) ([]*model.Video, error) {
	var videos []*model.Video
	db := r.db.WithContext(ctx).Where("status = ?", 1) // 只查已发布的

	if lastScore > 0 || lastID > 0 {
		// 复合游标分页逻辑: (score < lastScore) OR (score == lastScore AND id < lastID)
		db = db.Where("recommend_score < ? OR (recommend_score = ? AND id < ?)", lastScore, lastScore, lastID)
	}

	err := db.Order("recommend_score DESC, id DESC").Limit(limit).Find(&videos).Error
	return videos, err
}

func (r *videoRepository) GetByAuthorID(ctx context.Context, authorID int64) ([]*model.Video, error) {
	var videos []*model.Video
	err := r.db.WithContext(ctx).Where("author_id = ? AND status = ?", authorID, 1).Order("created_at DESC").Find(&videos).Error
	return videos, err
}
