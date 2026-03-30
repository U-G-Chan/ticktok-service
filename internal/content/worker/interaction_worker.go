package worker

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"ticktok-service/internal/content/model"
	"ticktok-service/internal/content/repository"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/kafka"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/redis"

	"go.uber.org/zap"
)

const (
	dirtyVideoStatsKey  = "dirty:video:stats"
	interactionCacheTTL = 24 * time.Hour
)

type favoriteEvent struct {
	EventID int64 `json:"event_id"`
	VideoID int64 `json:"video_id"`
	UserID  int64 `json:"user_id"`
	Status  int8  `json:"status"`
}

type InteractionWorker struct {
	videoRepo    repository.VideoRepository
	favoriteRepo repository.FavoriteRepository
	commentRepo  repository.CommentRepository
}

func NewInteractionWorker(
	videoRepo repository.VideoRepository,
	favoriteRepo repository.FavoriteRepository,
	commentRepo repository.CommentRepository,
) *InteractionWorker {
	return &InteractionWorker{
		videoRepo:    videoRepo,
		favoriteRepo: favoriteRepo,
		commentRepo:  commentRepo,
	}
}

func (w *InteractionWorker) StartFavoriteConsumer(ctx context.Context) {
	topic := config.Config.Kafka.FavoriteTopic
	if topic == "" {
		topic = "video_favorite_events"
	}
	groupID := config.Config.Kafka.FavoriteGroupID
	if groupID == "" {
		groupID = "group_favorite_worker"
	}

	reader := kafka.NewConsumerForTopic(topic, groupID)
	defer reader.Close()

	logger.Log.Info("Favorite worker started", zap.String("topic", topic))
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Favorite worker stopping")
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				logger.Log.Error("Favorite worker failed to read message", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			var event favoriteEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				logger.Log.Error("Favorite worker failed to unmarshal message", zap.Error(err))
				continue
			}

			favorite := &model.VideoFavorite{
				ID:      event.EventID,
				VideoID: event.VideoID,
				UserID:  event.UserID,
				Status:  event.Status,
			}
			if err := w.favoriteRepo.Upsert(ctx, favorite); err != nil {
				logger.Log.Error("Favorite worker failed to upsert favorite", zap.Error(err), zap.Int64("video_id", event.VideoID), zap.Int64("user_id", event.UserID))
				continue
			}
			if err := redis.RDB.SAdd(ctx, dirtyVideoStatsKey, strconv.FormatInt(event.VideoID, 10)).Err(); err != nil {
				logger.Log.Error("Favorite worker failed to mark dirty video stats", zap.Error(err), zap.Int64("video_id", event.VideoID))
			}
		}
	}
}

func (w *InteractionWorker) StartStatsFlusher(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	logger.Log.Info("Interaction stats flusher started")
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Interaction stats flusher stopping")
			return
		case <-ticker.C:
			w.flushDirtyVideoStats(ctx)
		}
	}
}

func (w *InteractionWorker) flushDirtyVideoStats(ctx context.Context) {
	videoIDStrs, err := redis.RDB.SMembers(ctx, dirtyVideoStatsKey).Result()
	if err != nil {
		logger.Log.Error("Stats flusher failed to read dirty set", zap.Error(err))
		return
	}

	for _, videoIDStr := range videoIDStrs {
		videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
		if err != nil {
			continue
		}

		favoriteCount, err := w.favoriteRepo.CountActiveByVideoID(ctx, videoID)
		if err != nil {
			logger.Log.Error("Stats flusher failed to count favorites", zap.Error(err), zap.Int64("video_id", videoID))
			continue
		}
		commentCount, err := w.commentRepo.CountActiveByVideoID(ctx, videoID)
		if err != nil {
			logger.Log.Error("Stats flusher failed to count comments", zap.Error(err), zap.Int64("video_id", videoID))
			continue
		}
		if err := w.videoRepo.UpdateInteractionCounts(ctx, videoID, favoriteCount, commentCount); err != nil {
			logger.Log.Error("Stats flusher failed to update video stats", zap.Error(err), zap.Int64("video_id", videoID))
			continue
		}
		if err := redis.RDB.Set(ctx, "cnt:video:like:"+videoIDStr, strconv.FormatInt(favoriteCount, 10), interactionCacheTTL).Err(); err != nil {
			logger.Log.Error("Stats flusher failed to refresh favorite cache", zap.Error(err), zap.Int64("video_id", videoID))
			continue
		}
		if err := redis.RDB.Set(ctx, "cnt:video:comment:"+videoIDStr, strconv.FormatInt(commentCount, 10), interactionCacheTTL).Err(); err != nil {
			logger.Log.Error("Stats flusher failed to refresh comment cache", zap.Error(err), zap.Int64("video_id", videoID))
			continue
		}
		if err := redis.RDB.SRem(ctx, dirtyVideoStatsKey, videoIDStr).Err(); err != nil {
			logger.Log.Error("Stats flusher failed to clear dirty video", zap.Error(err), zap.Int64("video_id", videoID))
		}
	}
}
