package repository

import (
	"context"
	"fmt"
	"strconv"
	"ticktok-service/pkg/redis"

	go_redis "github.com/go-redis/redis/v8"
)

const (
	inboxKeyPrefix = "inbox:"
)

// AddToInbox adds a msgID to the user's inbox sorted set using msgID as the score
func AddToInbox(ctx context.Context, userID int64, msgID int64) error {
	key := fmt.Sprintf("%s%d", inboxKeyPrefix, userID)
	return redis.RDB.ZAdd(ctx, key, &go_redis.Z{
		Score:  float64(msgID),
		Member: msgID,
	}).Err()
}

// GetMissedMessageIDs gets all msgIDs greater than the syncKey
func GetMissedMessageIDs(ctx context.Context, userID int64, syncKey int64) ([]int64, error) {
	key := fmt.Sprintf("%s%d", inboxKeyPrefix, userID)
	min := fmt.Sprintf("(%d", syncKey) // strictly greater than syncKey
	max := "+inf"

	members, err := redis.RDB.ZRangeByScore(ctx, key, &go_redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()

	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, m := range members {
		id, _ := strconv.ParseInt(m, 10, 64)
		ids = append(ids, id)
	}
	return ids, nil
}
