package model

import (
	"gorm.io/gorm"
)

type Video struct {
	gorm.Model
	AuthorID      int64  `gorm:"column:author_id;not null;index" json:"author_id"`
	PlayURL       string `gorm:"column:play_url;type:varchar(255);not null" json:"play_url"`
	CoverURL      string `gorm:"column:cover_url;type:varchar(255);not null" json:"cover_url"`
	FavoriteCount int64  `gorm:"column:favorite_count;default:0" json:"favorite_count"`
	CommentCount  int64  `gorm:"column:comment_count;default:0" json:"comment_count"`
	Title         string `gorm:"column:title;type:varchar(255);not null" json:"title"`
}

func (Video) TableName() string {
	return "videos"
}
