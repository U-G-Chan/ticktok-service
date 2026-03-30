package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	contentv1 "ticktok-service/api/content/v1"
	userv1 "ticktok-service/api/user/v1"
	"ticktok-service/internal/content/model"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/errno"
	"ticktok-service/pkg/kafka"
	"ticktok-service/pkg/redis"
	"ticktok-service/pkg/snowflake"
	"ticktok-service/pkg/util"

	goredis "github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

const (
	favoriteActionLike    = 1
	favoriteActionUnlike  = 2
	commentActionCreate   = 1
	commentActionDelete   = 2
	interactionCacheTTL   = 24 * time.Hour
	dirtyVideoStatsKey    = "dirty:video:stats"
	defaultCommentListCap = 100
)

type favoriteEvent struct {
	EventID int64 `json:"event_id"`
	VideoID int64 `json:"video_id"`
	UserID  int64 `json:"user_id"`
	Status  int8  `json:"status"`
}

func parseUserIDFromToken(token string) int64 {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0
	}
	claims, err := util.ParseToken(token)
	if err != nil {
		return 0
	}
	return claims.UserID
}

func favoriteStatusKey(videoID, userID int64) string {
	return "like:status:" + strconv.FormatInt(videoID, 10) + ":" + strconv.FormatInt(userID, 10)
}

func favoriteCountKey(videoID int64) string {
	return "cnt:video:like:" + strconv.FormatInt(videoID, 10)
}

func commentCountKey(videoID int64) string {
	return "cnt:video:comment:" + strconv.FormatInt(videoID, 10)
}

func commentListKey(videoID int64) string {
	return "comment:list:" + strconv.FormatInt(videoID, 10)
}

func commentDetailKey(commentID int64) string {
	return "comment:detail:" + strconv.FormatInt(commentID, 10)
}

func (s *contentService) buildPBVideo(ctx context.Context, video *model.Video, author *userv1.User, viewerID int64) (*contentv1.Video, error) {
	favoriteCount, err := s.getCachedCount(ctx, favoriteCountKey(video.ID), video.FavoriteCount)
	if err != nil {
		return nil, err
	}
	commentCount, err := s.getCachedCount(ctx, commentCountKey(video.ID), video.CommentCount)
	if err != nil {
		return nil, err
	}
	isFavorite, err := s.getFavoriteStatus(ctx, video.ID, viewerID)
	if err != nil {
		return nil, err
	}

	return &contentv1.Video{
		Id:             video.ID,
		Author:         author,
		PlayUrl:        video.PlayURL,
		CoverUrl:       video.CoverURL,
		FavoriteCount:  favoriteCount,
		CommentCount:   commentCount,
		IsFavorite:     isFavorite,
		Title:          video.Title,
		RecommendScore: video.RecommendScore,
	}, nil
}

func (s *contentService) FavoriteAction(ctx context.Context, userID, videoID int64, actionType int32) error {
	if userID == 0 {
		return &errno.ErrUnauthorized
	}

	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return &errno.ErrVideoNotFound
	}

	currentLiked, err := s.getFavoriteStatus(ctx, videoID, userID)
	if err != nil {
		return err
	}

	targetStatus := int8(0)
	countDelta := int64(0)
	switch actionType {
	case favoriteActionLike:
		targetStatus = 1
		if !currentLiked {
			countDelta = 1
		}
	case favoriteActionUnlike:
		targetStatus = 0
		if currentLiked {
			countDelta = -1
		}
	default:
		return &errno.ErrInvalidAction
	}

	if err := s.setFavoriteStatus(ctx, videoID, userID, targetStatus == 1); err != nil {
		return err
	}
	if countDelta != 0 {
		if _, err := s.adjustCachedCount(ctx, favoriteCountKey(videoID), video.FavoriteCount, countDelta); err != nil {
			return err
		}
		if err := s.markVideoStatsDirty(ctx, videoID); err != nil {
			return err
		}
	}

	event := favoriteEvent{
		EventID: snowflake.GenerateMsgID(),
		VideoID: videoID,
		UserID:  userID,
		Status:  targetStatus,
	}
	if err := s.sendFavoriteEvent(ctx, event); err != nil {
		favorite := &model.VideoFavorite{
			ID:      event.EventID,
			VideoID: videoID,
			UserID:  userID,
			Status:  targetStatus,
		}
		if upsertErr := s.favoriteRepo.Upsert(ctx, favorite); upsertErr != nil {
			return upsertErr
		}
		if err := s.markVideoStatsDirty(ctx, videoID); err != nil {
			return err
		}
	}

	return nil
}

func (s *contentService) CommentAction(ctx context.Context, userID, videoID int64, actionType int32, commentText string, commentID int64) (*contentv1.Comment, error) {
	if userID == 0 {
		return nil, &errno.ErrUnauthorized
	}

	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return nil, &errno.ErrVideoNotFound
	}

	switch actionType {
	case commentActionCreate:
		commentText = strings.TrimSpace(commentText)
		if commentText == "" {
			return nil, &errno.ErrValidation
		}

		comment := &model.VideoComment{
			ID:      snowflake.GenerateMsgID(),
			VideoID: videoID,
			UserID:  userID,
			Content: commentText,
			Status:  1,
		}
		if err := s.commentRepo.Create(ctx, comment); err != nil {
			return nil, err
		}
		if err := s.cacheComment(ctx, comment); err != nil {
			return nil, err
		}
		if _, err := s.adjustCachedCount(ctx, commentCountKey(videoID), video.CommentCount, 1); err != nil {
			return nil, err
		}
		if err := s.markVideoStatsDirty(ctx, videoID); err != nil {
			return nil, err
		}
		return s.buildPBComment(ctx, comment)
	case commentActionDelete:
		comment, err := s.commentRepo.GetByID(ctx, commentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, &errno.ErrCommentNotFound
			}
			return nil, err
		}
		if comment.UserID != userID || comment.Status != 1 {
			return nil, &errno.ErrCommentNotFound
		}
		if err := s.commentRepo.SoftDelete(ctx, commentID, userID); err != nil {
			return nil, err
		}
		if err := s.evictComment(ctx, comment); err != nil {
			return nil, err
		}
		videoForDelete := video
		if comment.VideoID != videoID {
			videoForDelete, err = s.repo.GetByID(ctx, comment.VideoID)
			if err != nil {
				return nil, err
			}
		}
		if _, err := s.adjustCachedCount(ctx, commentCountKey(comment.VideoID), videoForDelete.CommentCount, -1); err != nil {
			return nil, err
		}
		if err := s.markVideoStatsDirty(ctx, comment.VideoID); err != nil {
			return nil, err
		}
		return s.buildPBComment(ctx, comment)
	default:
		return nil, &errno.ErrInvalidAction
	}
}

func (s *contentService) GetCommentList(ctx context.Context, videoID int64) ([]*contentv1.Comment, error) {
	commentIDs, err := redis.RDB.ZRevRange(ctx, commentListKey(videoID), 0, defaultCommentListCap-1).Result()
	if err == nil && len(commentIDs) > 0 {
		comments := make([]*model.VideoComment, 0, len(commentIDs))
		cacheComplete := true
		for _, commentIDStr := range commentIDs {
			commentID, parseErr := strconv.ParseInt(commentIDStr, 10, 64)
			if parseErr != nil {
				cacheComplete = false
				break
			}
			comment, cacheErr := s.getCachedComment(ctx, commentID)
			if cacheErr != nil {
				cacheComplete = false
				break
			}
			comments = append(comments, comment)
		}
		if cacheComplete {
			return s.buildPBComments(ctx, comments)
		}
	}

	comments, err := s.commentRepo.ListByVideoID(ctx, videoID, defaultCommentListCap)
	if err != nil {
		return nil, err
	}
	for _, comment := range comments {
		if cacheErr := s.cacheComment(ctx, comment); cacheErr != nil {
			return nil, cacheErr
		}
	}
	return s.buildPBComments(ctx, comments)
}

func (s *contentService) sendFavoriteEvent(ctx context.Context, event favoriteEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	topic := "video_favorite_events"
	if config.Config.Kafka.FavoriteTopic != "" {
		topic = config.Config.Kafka.FavoriteTopic
	}
	return kafka.SendMessageToTopic(ctx, topic, []byte(strconv.FormatInt(event.VideoID, 10)), payload)
}

func (s *contentService) getFavoriteStatus(ctx context.Context, videoID, userID int64) (bool, error) {
	if userID == 0 {
		return false, nil
	}

	value, err := redis.RDB.Get(ctx, favoriteStatusKey(videoID, userID)).Result()
	if err == nil {
		return value == "1", nil
	}
	if err != goredis.Nil {
		return false, err
	}

	isFavorite, err := s.favoriteRepo.IsFavorite(ctx, videoID, userID)
	if err != nil {
		return false, err
	}
	if setErr := s.setFavoriteStatus(ctx, videoID, userID, isFavorite); setErr != nil {
		return false, setErr
	}
	return isFavorite, nil
}

func (s *contentService) setFavoriteStatus(ctx context.Context, videoID, userID int64, liked bool) error {
	value := "0"
	if liked {
		value = "1"
	}
	return redis.RDB.Set(ctx, favoriteStatusKey(videoID, userID), value, interactionCacheTTL).Err()
}

func (s *contentService) getCachedCount(ctx context.Context, key string, fallback int64) (int64, error) {
	value, err := redis.RDB.Get(ctx, key).Result()
	if err == nil {
		count, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr == nil {
			return count, nil
		}
	}
	if err != nil && err != goredis.Nil {
		return 0, err
	}
	if setErr := redis.RDB.SetNX(ctx, key, strconv.FormatInt(fallback, 10), interactionCacheTTL).Err(); setErr != nil {
		return 0, setErr
	}
	return fallback, nil
}

func (s *contentService) adjustCachedCount(ctx context.Context, key string, fallback int64, delta int64) (int64, error) {
	if err := redis.RDB.SetNX(ctx, key, strconv.FormatInt(fallback, 10), interactionCacheTTL).Err(); err != nil {
		return 0, err
	}
	count, err := redis.RDB.IncrBy(ctx, key, delta).Result()
	if err != nil {
		return 0, err
	}
	if count < 0 {
		count = 0
		if err := redis.RDB.Set(ctx, key, "0", interactionCacheTTL).Err(); err != nil {
			return 0, err
		}
	}
	if err := redis.RDB.Expire(ctx, key, interactionCacheTTL).Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *contentService) cacheComment(ctx context.Context, comment *model.VideoComment) error {
	payload, err := json.Marshal(comment)
	if err != nil {
		return err
	}
	if err := redis.RDB.Set(ctx, commentDetailKey(comment.ID), payload, interactionCacheTTL).Err(); err != nil {
		return err
	}
	if err := redis.RDB.ZAdd(ctx, commentListKey(comment.VideoID), &goredis.Z{
		Score:  float64(comment.CreatedAt.UnixNano()),
		Member: strconv.FormatInt(comment.ID, 10),
	}).Err(); err != nil {
		return err
	}
	return redis.RDB.Expire(ctx, commentListKey(comment.VideoID), interactionCacheTTL).Err()
}

func (s *contentService) evictComment(ctx context.Context, comment *model.VideoComment) error {
	if err := redis.RDB.Del(ctx, commentDetailKey(comment.ID)).Err(); err != nil {
		return err
	}
	return redis.RDB.ZRem(ctx, commentListKey(comment.VideoID), strconv.FormatInt(comment.ID, 10)).Err()
}

func (s *contentService) getCachedComment(ctx context.Context, commentID int64) (*model.VideoComment, error) {
	payload, err := redis.RDB.Get(ctx, commentDetailKey(commentID)).Result()
	if err != nil {
		return nil, err
	}
	var comment model.VideoComment
	if err := json.Unmarshal([]byte(payload), &comment); err != nil {
		return nil, err
	}
	return &comment, nil
}

func (s *contentService) buildPBComment(ctx context.Context, comment *model.VideoComment) (*contentv1.Comment, error) {
	comments, err := s.buildPBComments(ctx, []*model.VideoComment{comment})
	if err != nil {
		return nil, err
	}
	if len(comments) == 0 {
		return nil, nil
	}
	return comments[0], nil
}

func (s *contentService) buildPBComments(ctx context.Context, comments []*model.VideoComment) ([]*contentv1.Comment, error) {
	if len(comments) == 0 {
		return []*contentv1.Comment{}, nil
	}

	userIDMap := make(map[int64]struct{})
	userIDs := make([]int64, 0, len(comments))
	for _, comment := range comments {
		if _, ok := userIDMap[comment.UserID]; !ok {
			userIDMap[comment.UserID] = struct{}{}
			userIDs = append(userIDs, comment.UserID)
		}
	}

	authorMap := make(map[int64]*userv1.User)
	resp, err := s.userClient.MGetUserInfo(ctx, &userv1.MGetUserInfoRequest{UserIds: userIDs})
	if err == nil && resp.Code == int32(errno.Success.Code) {
		for _, user := range resp.Users {
			authorMap[user.Id] = user
		}
	}

	result := make([]*contentv1.Comment, 0, len(comments))
	for _, comment := range comments {
		author, ok := authorMap[comment.UserID]
		if !ok {
			author = &userv1.User{Id: comment.UserID}
		}
		result = append(result, &contentv1.Comment{
			Id:         comment.ID,
			User:       author,
			Content:    comment.Content,
			CreateDate: comment.CreatedAt.Format("01-02"),
		})
	}
	return result, nil
}

func (s *contentService) markVideoStatsDirty(ctx context.Context, videoID int64) error {
	return redis.RDB.SAdd(ctx, dirtyVideoStatsKey, strconv.FormatInt(videoID, 10)).Err()
}
