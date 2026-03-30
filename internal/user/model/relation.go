package model

import "time"

type Relation struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	UserID    int64     `gorm:"column:user_id;not null;uniqueIndex:uk_user_to_user" json:"user_id"`
	ToUserID  int64     `gorm:"column:to_user_id;not null;uniqueIndex:uk_user_to_user;index:idx_to_user_id" json:"to_user_id"`
	Status    int8      `gorm:"column:status;not null;default:1" json:"status"` // 1-follow, 0-unfollow
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Relation) TableName() string {
	return "relations"
}
