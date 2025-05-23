package model

import (
	"fmt"
	"ticktok-service/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接
var DB *gorm.DB

// SetupDB 初始化数据库连接
func SetupDB() error {
	dbConfig := config.AppConfig.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 自动迁移数据库表
	err = DB.AutoMigrate(
		// 商品相关表
		&Product{},
		&ProductImage{},
		&ProductLabel{},
		&ProductSpec{},
		&SpecOption{},
		&ProductService{},
		&Shop{},
		// 博客相关表
		&Blog{},
		&BlogImage{},
		&BlogTag{},
		&Comment{},
		// 聊天相关表
		&User{},
		&Friendship{},
		&Message{},
		&Session{},
		&UnreadMessage{},
		// 轮播内容相关表
		&SlideItem{},
		&SlideItemLabel{},
		&SlideAlbumImage{},
		// 发布相关表
		&MediaFile{},
		&Draft{},
		&Content{},
	)
	if err != nil {
		return fmt.Errorf("自动迁移数据库表失败: %w", err)
	}

	return nil
} 