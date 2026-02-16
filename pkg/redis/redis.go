package redis

import (
	"context"
	"fmt"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client

func Init() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Config.Redis.Host, config.Config.Redis.Port),
		Password: config.Config.Redis.Password,
		DB:       config.Config.Redis.DB,
	})

	if err := RDB.Ping(context.Background()).Err(); err != nil {
		logger.Log.Fatal("failed to connect redis: " + err.Error())
	}

	logger.Log.Info("Redis connected successfully")
}
