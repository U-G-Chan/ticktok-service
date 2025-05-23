package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"ticktok-service/internal/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UploadHandler 文件上传相关处理器
type UploadHandler struct {
	db            *gorm.DB
	baseUploadDir string
}

// NewUploadHandler 创建新的文件上传处理器
func NewUploadHandler(db *gorm.DB) *UploadHandler {
	// 确保上传目录存在
	baseDir := "./uploads"
	os.MkdirAll(filepath.Join(baseDir, "photo"), os.ModePerm)
	os.MkdirAll(filepath.Join(baseDir, "video"), os.ModePerm)
	
	return &UploadHandler{
		db:            db,
		baseUploadDir: baseDir,
	}
}

// UploadMedia 上传媒体文件
func (h *UploadHandler) UploadMedia(c *gin.Context) {
	// 1. 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, model.UploadResponse{
			Code: 400,
			Msg:  "未找到上传文件",
			Data: []model.UploadResponseData{},
		})
		return
	}
	
	// 2. 获取文件类型
	fileType := c.PostForm("type")
	if fileType != "photo" && fileType != "video" {
		c.JSON(http.StatusOK, model.UploadResponse{
			Code: 400,
			Msg:  "文件类型必须是photo或video",
			Data: []model.UploadResponseData{},
		})
		return
	}
	
	// 3. 生成唯一文件名和ID
	fileID := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s%s", fileID, ext)
	
	// 4. 确定保存路径
	relativePath := filepath.Join(fileType, time.Now().Format("2006/01/02"), fileName)
	fullPath := filepath.Join(h.baseUploadDir, relativePath)
	
	// 5. 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		c.JSON(http.StatusOK, model.UploadResponse{
			Code: 500,
			Msg:  "创建上传目录失败",
			Data: []model.UploadResponseData{},
		})
		return
	}
	
	// 6. 保存文件
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		c.JSON(http.StatusOK, model.UploadResponse{
			Code: 500,
			Msg:  "保存文件失败: " + err.Error(),
			Data: []model.UploadResponseData{},
		})
		return
	}
	
	// 7. 保存文件信息到数据库
	mediaFile := model.MediaFile{
		ID:        fileID,
		Type:      fileType,
		URL:       "/uploads/" + filepath.ToSlash(relativePath), // 使用正斜杠
		FilePath:  fullPath,
		FileName:  file.Filename,
		FileSize:  file.Size,
		CreatedAt: time.Now(),
	}
	
	if err := h.db.Create(&mediaFile).Error; err != nil {
		// 如果数据库保存失败，删除已上传的文件
		os.Remove(fullPath)
		c.JSON(http.StatusOK, model.UploadResponse{
			Code: 500,
			Msg:  "保存文件信息失败",
			Data: []model.UploadResponseData{},
		})
		return
	}
	
	// 8. 返回成功响应
	c.JSON(http.StatusOK, model.UploadResponse{
		Code: 200,
		Msg:  "上传成功",
		Data: []model.UploadResponseData{
			{
				ID:   fileID,
				Type: fileType,
				URL:  "/uploads/" + filepath.ToSlash(relativePath),
			},
		},
	})
} 