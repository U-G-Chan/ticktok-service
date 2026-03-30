package repository

import (
	"context"
	"ticktok-service/internal/content/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FavoriteRepository interface {
	Upsert(ctx context.Context, favorite *model.VideoFavorite) error
	UpdateStatus(ctx context.Context, videoID, userID int64, status int8) error
	IsFavorite(ctx context.Context, videoID, userID int64) (bool, error)
	CountActiveByVideoID(ctx context.Context, videoID int64) (int64, error)
}

type favoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepository{db: db}
}

func (r *favoriteRepository) Upsert(ctx context.Context, favorite *model.VideoFavorite) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "video_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status":     favorite.Status,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(favorite).Error
}

func (r *favoriteRepository) UpdateStatus(ctx context.Context, videoID, userID int64, status int8) error {
	return r.db.WithContext(ctx).
		Model(&model.VideoFavorite{}).
		Where("video_id = ? AND user_id = ?", videoID, userID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

func (r *favoriteRepository) IsFavorite(ctx context.Context, videoID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.VideoFavorite{}).
		Where("video_id = ? AND user_id = ? AND status = ?", videoID, userID, 1).
		Count(&count).Error
	return count > 0, err
}

func (r *favoriteRepository) CountActiveByVideoID(ctx context.Context, videoID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.VideoFavorite{}).
		Where("video_id = ? AND status = ?", videoID, 1).
		Count(&count).Error
	return count, err
}
