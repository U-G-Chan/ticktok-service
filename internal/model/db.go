package model

import (
	"fmt"
	"log"
	"ticktok-service/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

	log.Println("数据库连接成功")

	// 自动迁移模型
	err = DB.AutoMigrate(&Blog{}, &BlogImage{}, &BlogTag{}, &Comment{})
	if err != nil {
		return fmt.Errorf("自动迁移数据表失败: %w", err)
	}

	return nil
} 