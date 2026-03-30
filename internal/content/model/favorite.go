package model

import "time"

type VideoFavorite struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	VideoID   int64     `gorm:"column:video_id;not null;uniqueIndex:uk_video_user" json:"video_id"`
	UserID    int64     `gorm:"column:user_id;not null;uniqueIndex:uk_video_user" json:"user_id"`
	Status    int8      `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (VideoFavorite) TableName() string {
	return "video_favorites"
}
