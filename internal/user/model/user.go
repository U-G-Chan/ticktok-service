package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username        string `gorm:"column:username;type:varchar(32);not null;uniqueIndex"`
	Password        string `gorm:"column:password;type:varchar(255);not null"`
	Role            string `gorm:"column:role;type:varchar(20);default:'user'"`
	Avatar          string `gorm:"column:avatar;type:varchar(255)"`
	BackgroundImage string `gorm:"column:background_image;type:varchar(255)"`
	Signature       string `gorm:"column:signature;type:varchar(255)"`
}

func (User) TableName() string {
	return "users"
}
