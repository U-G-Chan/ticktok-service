package worker

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"ticktok-service/internal/content/repository"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/ffmpeg"
	"ticktok-service/pkg/kafka"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/minio"

	"go.uber.org/zap"
)

type CoverWorker struct {
	repo repository.VideoRepository
}

func NewCoverWorker(repo repository.VideoRepository) *CoverWorker {
	return &CoverWorker{repo: repo}
}

func (w *CoverWorker) Start(ctx context.Context) {
	topic := config.Config.Kafka.VideoPublishTopic
	if topic == "" {
		topic = "video_publish_events"
	}
	groupID := config.Config.Kafka.VideoGroupID
	if groupID == "" {
		groupID = "group_video_worker"
	}

	reader := kafka.NewConsumerForTopic(topic, groupID)
	defer reader.Close()

	logger.Log.Info("Cover worker started, listening on topic: " + topic)

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Cover worker stopping...")
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				logger.Log.Error("Cover worker failed to read message", zap.Error(err))
				time.Sleep(time.Second) // backoff
				continue
			}

			videoIDStr := string(msg.Value)
			videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
			if err != nil {
				logger.Log.Error("Cover worker invalid video id format", zap.String("id", videoIDStr))
				continue
			}

			w.processVideoCover(ctx, videoID)
		}
	}
}

func (w *CoverWorker) processVideoCover(ctx context.Context, videoID int64) {
	video, err := w.repo.GetByID(ctx, videoID)
	if err != nil {
		logger.Log.Error("Cover worker failed to get video from db", zap.Int64("video_id", videoID), zap.Error(err))
		return
	}

	tempVideoFile := fmt.Sprintf("%s/%d.mp4", os.TempDir(), videoID)
	tempCoverFile := fmt.Sprintf("%s/%d.jpg", os.TempDir(), videoID)
	defer os.Remove(tempVideoFile)
	defer os.Remove(tempCoverFile)

	videoObjectName := fmt.Sprintf("%d.mp4", videoID)
	if err := minio.DownloadToLocalFile(ctx, videoObjectName, tempVideoFile); err != nil {
		logger.Log.Error("Cover worker failed to download video from minio", zap.Int64("video_id", videoID), zap.Error(err))
		return
	}

	logger.Log.Info("Starting to extract cover", zap.Int64("video_id", videoID), zap.String("video_file", tempVideoFile))
	err = ffmpeg.ExtractCover(tempVideoFile, tempCoverFile)
	if err != nil {
		logger.Log.Error("Cover worker ffmpeg failed", zap.Int64("video_id", videoID), zap.Error(err))
		return
	}

	objectName := fmt.Sprintf("%d.jpg", videoID)
	_, err = minio.UploadLocalFile(ctx, objectName, tempCoverFile, "image/jpeg")
	if err != nil {
		logger.Log.Error("Cover worker failed to upload cover", zap.Int64("video_id", videoID), zap.Error(err))
		return
	}

	if video.CoverURL == "" {
		video.CoverURL = minio.GetObjectURL(objectName)
		if err := w.repo.Update(ctx, video); err != nil {
			logger.Log.Error("Cover worker failed to update video cover url in db", zap.Int64("video_id", videoID), zap.Error(err))
			return
		}
	}

	logger.Log.Info("Successfully processed cover", zap.Int64("video_id", videoID), zap.String("cover_url", video.CoverURL))
}
