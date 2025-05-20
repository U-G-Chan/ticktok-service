package model

import (
	"time"
)

// Blog 博客模型
type Blog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	AuthorID     string    `gorm:"size:50;not null" json:"authorId"`
	AuthorName   string    `gorm:"size:100;not null" json:"authorName"`
	AuthorAvatar string    `gorm:"size:255;not null" json:"authorAvatar"`
	Title        string    `gorm:"size:200;not null" json:"title"`
	CoverImg     string    `gorm:"size:255;not null" json:"coverImg"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	CreatedAt    time.Time `gorm:"not null" json:"-"`
	CreatedAtStr string    `gorm:"-" json:"createdAt"`
	Likes        int       `gorm:"default:0" json:"likes"`
	Forwards     int       `gorm:"default:0" json:"forwards"`
	Stars        int       `gorm:"default:0" json:"stars"`
	IsFollowing  bool      `gorm:"-" json:"isFollowing"`
	
	// 关联
	Images   []BlogImage `gorm:"foreignKey:BlogID" json:"images"`
	Tags     []BlogTag   `gorm:"foreignKey:BlogID" json:"tags"`
	Comments []Comment   `gorm:"foreignKey:BlogID" json:"comments"`
}

// BlogImage 博客图片模型
type BlogImage struct {
	ID       uint   `gorm:"primaryKey" json:"-"`
	BlogID   uint   `gorm:"not null" json:"-"`
	ImageURL string `gorm:"size:255;not null" json:"-"`
}

// BlogTag 博客标签模型
type BlogTag struct {
	ID         uint   `gorm:"primaryKey" json:"-"`
	BlogID     uint   `gorm:"not null" json:"-"`
	TagContent string `gorm:"size:50;not null" json:"-"`
}

// Comment 评论模型
type Comment struct {
	ID           string    `gorm:"primaryKey;size:50" json:"id"`
	BlogID       uint      `gorm:"not null" json:"-"`
	AuthorID     string    `gorm:"size:50;not null" json:"authorId"`
	AuthorName   string    `gorm:"size:100;not null" json:"authorName"`
	AuthorAvatar string    `gorm:"size:255;not null" json:"authorAvatar"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	CreatedAt    time.Time `gorm:"not null" json:"-"`
	CreatedAtStr string    `gorm:"-" json:"createdAt"`
	Location     string    `gorm:"size:100" json:"location"`
	Likes        int       `gorm:"default:0" json:"likes"`
}

// AfterFind GORM的钩子，用于格式化创建时间
func (b *Blog) AfterFind() error {
	b.CreatedAtStr = b.CreatedAt.Format("01-02")
	return nil
}

// AfterFind GORM的钩子，用于格式化评论创建时间
func (c *Comment) AfterFind() error {
	c.CreatedAtStr = c.CreatedAt.Format("01-02")
	return nil
} 