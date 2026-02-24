package mysql

import (
	"fmt"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Config.MySQL.User,
		config.Config.MySQL.Password,
		config.Config.MySQL.Host,
		config.Config.MySQL.Port,
		config.Config.MySQL.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal("failed to connect database: " + err.Error())
	}

	logger.Log.Info("MySQL connected successfully")
}
