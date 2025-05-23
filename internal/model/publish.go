package model

import (
	"time"
)

// MediaFile 媒体文件模型
type MediaFile struct {
	ID        string    `json:"id" gorm:"primaryKey;size:50"`
	Type      string    `json:"type" gorm:"column:type;type:enum('photo','video');not null"`
	URL       string    `json:"url" gorm:"column:url;size:255;not null"`
	FilePath  string    `json:"-" gorm:"column:file_path;size:500;not null"` // 实际文件路径
	FileName  string    `json:"-" gorm:"column:file_name;size:255;not null"` // 原始文件名
	FileSize  int64     `json:"-" gorm:"column:file_size;not null"`          // 文件大小
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;not null"`
}

// Draft 草稿模型
type Draft struct {
	ID          string    `json:"id" gorm:"primaryKey;size:50"`
	UserID      uint      `json:"userId" gorm:"column:user_id;not null;default:1"`
	Title       string    `json:"title" gorm:"column:title;size:255;not null"`
	Description string    `json:"description" gorm:"column:description;type:text"`
	MediaItems  string    `json:"-" gorm:"column:media_items;type:json"`       // JSON存储媒体项
	Topics      string    `json:"-" gorm:"column:topics;type:json"`            // JSON存储话题
	Mentions    string    `json:"-" gorm:"column:mentions;type:json"`          // JSON存储提及
	Tags        string    `json:"-" gorm:"column:tags;type:json"`              // JSON存储标签
	Visibility  string    `json:"visibility" gorm:"column:visibility;type:enum('public','friends','private');default:'public'"`
	IsDaily     bool      `json:"isDaily" gorm:"column:is_daily;default:false"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at;not null"`
}

// Content 发布内容模型
type Content struct {
	ID          string    `json:"id" gorm:"primaryKey;size:50"`
	UserID      uint      `json:"userId" gorm:"column:user_id;not null;default:1"`
	Title       string    `json:"title" gorm:"column:title;size:255;not null"`
	Description string    `json:"description" gorm:"column:description;type:text"`
	MediaItems  string    `json:"-" gorm:"column:media_items;type:json"`       // JSON存储媒体项
	Topics      string    `json:"-" gorm:"column:topics;type:json"`            // JSON存储话题
	Mentions    string    `json:"-" gorm:"column:mentions;type:json"`          // JSON存储提及
	Tags        string    `json:"-" gorm:"column:tags;type:json"`              // JSON存储标签
	Visibility  string    `json:"visibility" gorm:"column:visibility;type:enum('public','friends','private');default:'public'"`
	IsDaily     bool      `json:"isDaily" gorm:"column:is_daily;default:false"`
	ViewCount   int64     `json:"viewCount" gorm:"column:view_count;default:0"`
	LikeCount   int64     `json:"likeCount" gorm:"column:like_count;default:0"`
	CommentCount int64    `json:"commentCount" gorm:"column:comment_count;default:0"`
	ShareCount  int64     `json:"shareCount" gorm:"column:share_count;default:0"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at;not null"`
}

// MediaItem 前端媒体项结构
type MediaItem struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// PublishMessage 发布消息请求结构
type PublishMessage struct {
	Title       string      `json:"title" binding:"required"`
	Description string      `json:"description"`
	MediaItems  []MediaItem `json:"mediaItems"`
	Topics      []string    `json:"topics"`
	Mentions    []string    `json:"mentions"`
	Tags        []string    `json:"tags"`
	Visibility  string      `json:"visibility" binding:"required,oneof=public friends private"`
	IsDaily     bool        `json:"isDaily"`
	CreatedAt   int64       `json:"createdAt,omitempty"`
	UpdatedAt   int64       `json:"updatedAt,omitempty"`
}

// UploadResponse 文件上传响应
type UploadResponse struct {
	Code int                   `json:"code"`
	Msg  string                `json:"msg"`
	Data []UploadResponseData  `json:"data"`
}

// UploadResponseData 上传响应数据
type UploadResponseData struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// SaveDraftResponse 保存草稿响应
type SaveDraftResponse struct {
	Success bool   `json:"success"`
	DraftID string `json:"draftId,omitempty"`
}

// PublishContentResponse 发布内容响应
type PublishContentResponse struct {
	Success   bool   `json:"success"`
	PublishID string `json:"publishId,omitempty"`
}

// TopicsResponse 话题列表响应
type TopicsResponse struct {
	Topics []string `json:"topics"`
} 