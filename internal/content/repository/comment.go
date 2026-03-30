package repository

import (
	"context"
	"ticktok-service/internal/content/model"

	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *model.VideoComment) error
	SoftDelete(ctx context.Context, commentID, userID int64) error
	GetByID(ctx context.Context, commentID int64) (*model.VideoComment, error)
	ListByVideoID(ctx context.Context, videoID int64, limit int) ([]*model.VideoComment, error)
	CountActiveByVideoID(ctx context.Context, videoID int64) (int64, error)
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *model.VideoComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepository) SoftDelete(ctx context.Context, commentID, userID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.VideoComment{}).
		Where("id = ? AND user_id = ? AND status = ?", commentID, userID, 1).
		Updates(map[string]interface{}{
			"status":     2,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

func (r *commentRepository) GetByID(ctx context.Context, commentID int64) (*model.VideoComment, error) {
	var comment model.VideoComment
	err := r.db.WithContext(ctx).First(&comment, commentID).Error
	return &comment, err
}

func (r *commentRepository) ListByVideoID(ctx context.Context, videoID int64, limit int) ([]*model.VideoComment, error) {
	var comments []*model.VideoComment
	err := r.db.WithContext(ctx).
		Where("video_id = ? AND status = ?", videoID, 1).
		Order("created_at DESC").
		Limit(limit).
		Find(&comments).Error
	return comments, err
}

func (r *commentRepository) CountActiveByVideoID(ctx context.Context, videoID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.VideoComment{}).
		Where("video_id = ? AND status = ?", videoID, 1).
		Count(&count).Error
	return count, err
}
