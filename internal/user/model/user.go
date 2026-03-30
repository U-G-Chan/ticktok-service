package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID              int64          `gorm:"primaryKey;autoIncrement:false" json:"id"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
	Username        string         `gorm:"column:username;type:varchar(32);not null;uniqueIndex"`
	Password        string         `gorm:"column:password;type:varchar(255);not null"`
	Role            string         `gorm:"column:role;type:varchar(20);default:'user'"`
	Avatar          string         `gorm:"column:avatar;type:varchar(255);default:''"`
	BackgroundImage string         `gorm:"column:background_image;type:varchar(255);default:''"`
	Signature       string         `gorm:"column:signature;type:varchar(255);default:''"`
}

func (User) TableName() string {
	return "users"
}
