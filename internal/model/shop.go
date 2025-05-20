package model

import (
	"time"
)

// Shop 店铺模型
type Shop struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Logo      string    `json:"logo" gorm:"size:255;not null"`
	Sales     string    `json:"sales" gorm:"size:20;not null"`
	Rating    float64   `json:"rating" gorm:"type:decimal(2,1);default:5.0"`
	Followers string    `json:"followers" gorm:"size:20;default:'0'"`
	CreatedAt time.Time `json:"createdAt" gorm:"not null"`
}

// ShopBrief 简要店铺信息
type ShopBrief struct {
	Name  string `json:"name"`
	Sales string `json:"sales"`
}

// ShopDetail 详细店铺信息
type ShopDetail struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Sales     string  `json:"sales"`
	Logo      string  `json:"logo"`
	Rating    float64 `json:"rating"`
	Followers string  `json:"followers"`
} 