package router

import (
	"context"
	"fmt"
	"ticktok-service/pkg/redis"
	"time"
)

const (
	routeKeyPrefix = "im:route:user:"
	routeTTL       = 120 * time.Second
)

// SetRoute sets the user's current node IP in Redis
func SetRoute(ctx context.Context, userID int64, nodeIP string) error {
	key := fmt.Sprintf("%s%d", routeKeyPrefix, userID)
	return redis.RDB.Set(ctx, key, nodeIP, routeTTL).Err()
}

// GetRoute gets the user's current node IP from Redis
func GetRoute(ctx context.Context, userID int64) (string, error) {
	key := fmt.Sprintf("%s%d", routeKeyPrefix, userID)
	return redis.RDB.Get(ctx, key).Result()
}

// DeleteRoute removes the user's routing info
func DeleteRoute(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("%s%d", routeKeyPrefix, userID)
	return redis.RDB.Del(ctx, key).Err()
}

// RenewRoute extends the TTL of the user's routing info (Heartbeat)
func RenewRoute(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("%s%d", routeKeyPrefix, userID)
	return redis.RDB.Expire(ctx, key, routeTTL).Err()
}
