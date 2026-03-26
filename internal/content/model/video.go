package model

import (
	"time"
)

type Video struct {
	ID             int64     `gorm:"primaryKey;autoIncrement:false;index:idx_score_id,priority:2,sort:desc" json:"id"`
	AuthorID       int64     `gorm:"column:author_id;not null;index:idx_author" json:"author_id"`
	Title          string    `gorm:"column:title;type:varchar(255);not null" json:"title"`
	PlayURL        string    `gorm:"column:play_url;type:varchar(255);not null" json:"play_url"`
	CoverURL       string    `gorm:"column:cover_url;type:varchar(255);not null" json:"cover_url"`
	RecommendScore int32     `gorm:"column:recommend_score;not null;default:0;index:idx_score_id,priority:1,sort:desc" json:"recommend_score"`
	FavoriteCount  int64     `gorm:"column:favorite_count;not null;default:0" json:"favorite_count"`
	CommentCount   int64     `gorm:"column:comment_count;not null;default:0" json:"comment_count"`
	Status         int8      `gorm:"column:status;not null;default:0" json:"status"` // 0-Pending, 1-Published
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Video) TableName() string {
	return "videos"
}
