package model

import (
	"time"

	"gorm.io/gorm"
)

// ListType 列表类型
type ListType string

const (
	ListTypeWorks      ListType = "works"      // 作品
	ListTypeRecommend  ListType = "recommend"  // 推荐
	ListTypeCollection ListType = "collection" // 收藏
	ListTypeLikes      ListType = "likes"      // 点赞
)

// WorkType 作品类型
type WorkType string

const (
	WorkTypeDraft     WorkType = "draft"     // 草稿
	WorkTypePublished WorkType = "published" // 已发布
	WorkTypePrivate   WorkType = "private"   // 私密
)

// UserContent 用户内容表
type UserContent struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ItemID      string         `json:"itemId" gorm:"type:varchar(255);uniqueIndex;not null"`
	UserID      string         `json:"userId" gorm:"type:varchar(100);index;not null"`
	ListType    ListType       `json:"listType" gorm:"type:varchar(50);not null"`
	Thumbnail   string         `json:"thumbnail" gorm:"type:text"`
	Likes       int            `json:"likes" gorm:"default:0"`
	WorkType    WorkType       `json:"workType" gorm:"type:varchar(20);default:'published'"`
	Title       string         `json:"title" gorm:"type:varchar(500)"`
	Description string         `json:"description" gorm:"type:text"`
	PublishTime *time.Time     `json:"publishTime"`
	Duration    int            `json:"duration" gorm:"default:0"`
	Tags        string         `json:"tags" gorm:"type:json"` // JSON字符串存储标签数组
	IsPublic    bool           `json:"isPublic" gorm:"default:true"`
	Other       string         `json:"other" gorm:"type:json"` // JSON字符串存储其他扩展字段
	CreatedAt   time.Time      `json:"createTime"`
	UpdatedAt   time.Time      `json:"updateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// ContentItem 前端API对应的内容项结构
type ContentItem struct {
	ItemID    string    `json:"itemId"`
	ListType  ListType  `json:"listType"`
	Thumbnail string    `json:"thumbnail"`
	Likes     int       `json:"likes"`
	Other     *OtherInfo `json:"other,omitempty"`
}

// OtherInfo 其他信息结构
type OtherInfo struct {
	WorkType    WorkType   `json:"workType,omitempty"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	PublishTime *time.Time `json:"publishTime,omitempty"`
	Duration    int        `json:"duration,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	IsPublic    bool       `json:"isPublic"`
	CreateTime  time.Time  `json:"createTime"`
	UpdateTime  time.Time  `json:"updateTime"`
}

// ContentLike 用户点赞表
type ContentLike struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    string         `json:"userId" gorm:"type:varchar(100);index;not null"`
	ItemID    string         `json:"itemId" gorm:"type:varchar(255);index;not null"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 设置表名
func (UserContent) TableName() string {
	return "user_contents"
}

// TableName 设置表名
func (ContentLike) TableName() string {
	return "content_likes"
}

// CreateContentParams 创建内容参数
type CreateContentParams struct {
	ItemID      string            `json:"itemId"`
	UserID      string            `json:"userId" binding:"required"`
	ListType    ListType          `json:"listType" binding:"required"`
	Thumbnail   string            `json:"thumbnail"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	WorkType    WorkType          `json:"workType"`
	Duration    int               `json:"duration"`
	Likes       int               `json:"likes"`
	Tags        []string          `json:"tags"`
	IsPublic    *bool             `json:"isPublic"`
	Other       map[string]interface{} `json:"other"`
}

// UpdateContentParams 更新内容参数
type UpdateContentParams struct {
	ItemID      string            `json:"itemId" binding:"required"`
	UserID      string            `json:"userId" binding:"required"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Thumbnail   string            `json:"thumbnail"`
	WorkType    WorkType          `json:"workType"`
	IsPublic    *bool             `json:"isPublic"`
	Tags        []string          `json:"tags"`
	Other       map[string]interface{} `json:"other"`
}

// ContentListParams 内容列表查询参数
type ContentListParams struct {
	UserID   string   `form:"userId" binding:"required"`
	ListType ListType `form:"listType" binding:"required"`
	Page     int      `form:"page"`
	PageSize int      `form:"pageSize"`
}

// ContentListResult 内容列表返回结果
type ContentListResult struct {
	Items    []ContentItem `json:"items"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	HasMore  bool          `json:"hasMore"`
}

// ToggleLikeParams 切换点赞参数
type ToggleLikeParams struct {
	ItemID   string `json:"itemId" binding:"required"`
	UserID   string `json:"userId" binding:"required"`
	IsLiked  bool   `json:"isLiked"`
}

// ToggleLikeResult 切换点赞结果
type ToggleLikeResult struct {
	Likes   int  `json:"likes"`
	IsLiked bool `json:"isLiked"`
} 