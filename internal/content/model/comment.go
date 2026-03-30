package model

import "time"

type VideoComment struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	VideoID   int64     `gorm:"column:video_id;not null;index:idx_video_status_created,priority:1" json:"video_id"`
	UserID    int64     `gorm:"column:user_id;not null;index" json:"user_id"`
	Content   string    `gorm:"column:content;type:text;not null" json:"content"`
	Status    int8      `gorm:"column:status;not null;default:1;index:idx_video_status_created,priority:2" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;index:idx_video_status_created,priority:3,sort:desc" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (VideoComment) TableName() string {
	return "video_comments"
}
