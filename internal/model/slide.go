package model

import (
	"time"
)

// SlideItem 轮播内容模型
type SlideItem struct {
	ID          string    `json:"id" gorm:"primaryKey;size:20"`
	ItemID      string    `json:"itemId" gorm:"column:item_id;size:20;not null;uniqueIndex"`
	ContentType string    `json:"contentType" gorm:"column:content_type;type:enum('video','picture');not null"`
	Title       string    `json:"title" gorm:"type:text;not null"`
	Author      string    `json:"author" gorm:"size:100;not null"`
	Likes       int64     `json:"likes" gorm:"default:0"`
	Comments    int64     `json:"comments" gorm:"default:0"`
	Stars       int64     `json:"stars" gorm:"default:0"`
	Forwards    int64     `json:"forwards" gorm:"default:0"`
	VideoURL    string    `json:"videoUrl,omitempty" gorm:"column:video_url;size:255"`
	Avatar      string    `json:"avatar" gorm:"size:255;not null"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at;not null"`
	
	// 关联
	Labels []SlideItemLabel `json:"labels" gorm:"foreignKey:ItemID;references:ItemID"`
	Album  []SlideAlbumImage `json:"album,omitempty" gorm:"foreignKey:ItemID;references:ItemID"`
}

// SlideItemLabel 轮播内容标签模型
type SlideItemLabel struct {
	ID           int    `json:"id" gorm:"primaryKey;autoIncrement"`
	ItemID       string `json:"-" gorm:"column:item_id;size:20;not null"`
	LabelContent string `json:"labelContent" gorm:"column:label_content;size:50;not null"`
}

// SlideAlbumImage 轮播内容相册图片模型
type SlideAlbumImage struct {
	ID        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	ItemID    string `json:"-" gorm:"column:item_id;size:20;not null"`
	ImageURL  string `json:"imageUrl" gorm:"column:image_url;size:255;not null"`
	SortOrder int    `json:"sortOrder" gorm:"column:sort_order;default:0"`
}

// SlideItemResponse 轮播内容响应
type SlideItemResponse struct {
	ID          string   `json:"id"`
	ItemID      string   `json:"itemId"`
	ContentType string   `json:"contentType"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Likes       int64    `json:"likes"`
	Comments    int64    `json:"comments"`
	Stars       int64    `json:"stars"`
	Forwards    int64    `json:"forwards"`
	Labels      []string `json:"labels"`
	VideoURL    string   `json:"videoUrl,omitempty"`
	Album       []string `json:"album,omitempty"`
	Avatar      string   `json:"avatar"`
}

// SlideResponse 轮播内容列表响应
type SlideResponse struct {
	Items   []*SlideItemResponse `json:"data"`
	Total   int64                `json:"total"`
	HasMore bool                 `json:"hasMore"`
} 