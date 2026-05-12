package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	contentmodel "ticktok-service/internal/content/model"
	usermodel "ticktok-service/internal/user/model"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/minio"
	"ticktok-service/pkg/mysql"
	"ticktok-service/pkg/redis"
	"ticktok-service/pkg/snowflake"
	"ticktok-service/pkg/util"

	goredis "github.com/go-redis/redis/v8"
	miniogo "github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type feedItem struct {
	ItemID      string   `json:"itemId"`
	ContentType string   `json:"contentType"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Likes       int64    `json:"likes"`
	Comments    int64    `json:"comments"`
	VideoURL    string   `json:"videoUrl"`
	Avatar      string   `json:"avatar"`
	Cover       string   `json:"cover"`
	Album       []string `json:"album"`
}

func main() {
	var (
		configPath = flag.String("config", "config/config.yaml", "")
		assetRoot  = flag.String("assets", "assets/web-public", "")
		limit      = flag.Int("limit", 0, "")
		dryRun     = flag.Bool("dry-run", false, "")
	)
	flag.Parse()

	if err := config.Init(*configPath); err != nil {
		panic(err)
	}
	logger.Init(config.Config.LogLevel)
	snowflake.Init()
	mysql.Init()
	redis.Init()
	minio.Init()

	_ = usermodel.AutoMigrate(mysql.DB)
	_ = contentmodel.AutoMigrate(mysql.DB)

	ctx := context.Background()

	dataPath := filepath.Join(*assetRoot, "media", "data.json")
	raw, err := os.ReadFile(dataPath)
	if err != nil {
		logger.Log.Fatal("read data.json failed: " + err.Error())
	}

	var items []feedItem
	if err := json.Unmarshal(raw, &items); err != nil {
		logger.Log.Fatal("parse data.json failed: " + err.Error())
	}

	passwordHash, err := util.HashPassword("seed_password")
	if err != nil {
		logger.Log.Fatal("hash seed password failed: " + err.Error())
	}

	authorCache := make(map[string]*usermodel.User)

	max := len(items)
	if *limit > 0 && *limit < max {
		max = *limit
	}

	processed := 0
	for i := 0; i < max; i++ {
		it := items[i]
		if it.ContentType != "video" {
			continue
		}

		videoID, err := strconv.ParseInt(it.ItemID, 10, 64)
		if err != nil || videoID <= 0 {
			continue
		}

		author, err := ensureAuthor(ctx, mysql.DB, authorCache, passwordHash, *assetRoot, it.Author, it.Avatar, *dryRun)
		if err != nil {
			logger.Log.Warn("ensure author failed: " + err.Error())
			continue
		}

		playURL, err := ensureObjectFromPublicPath(ctx, *assetRoot, it.VideoURL, "seed", *dryRun)
		if err != nil {
			logger.Log.Warn("upload video failed: " + err.Error())
			continue
		}

		coverURL, err := ensureObjectFromPublicPath(ctx, *assetRoot, it.Cover, "seed", *dryRun)
		if err != nil {
			logger.Log.Warn("upload cover failed: " + err.Error())
			continue
		}

		title := strings.TrimSpace(it.Title)
		if title == "" {
			title = "无标题"
		}

		recommendScore := int32(it.Likes / 1000)
		if recommendScore < 0 {
			recommendScore = 0
		}

		v := &contentmodel.Video{
			ID:             videoID,
			AuthorID:       author.ID,
			Title:          title,
			PlayURL:        playURL,
			CoverURL:       coverURL,
			RecommendScore: recommendScore,
			FavoriteCount:  it.Likes,
			CommentCount:   it.Comments,
			Status:         1,
		}

		if err := upsertVideo(mysql.DB, v, *dryRun); err != nil {
			logger.Log.Warn("upsert video failed: " + err.Error())
			continue
		}

		if !*dryRun {
			if err := redis.RDB.ZAdd(ctx, "feed:recommend_score", &goredis.Z{
				Score:  float64(v.RecommendScore),
				Member: fmt.Sprintf("%d", v.ID),
			}).Err(); err != nil {
				logger.Log.Warn("seed redis zset failed: " + err.Error())
			}
		}

		processed++
	}

	logger.Log.Info(fmt.Sprintf("seed done, processed=%d", processed))
}

func ensureAuthor(
	ctx context.Context,
	db *gorm.DB,
	cache map[string]*usermodel.User,
	passwordHash string,
	assetRoot string,
	authorName string,
	avatarPublicPath string,
	dryRun bool,
) (*usermodel.User, error) {
	key := strings.TrimSpace(authorName) + "|" + strings.TrimSpace(avatarPublicPath)
	if u, ok := cache[key]; ok {
		return u, nil
	}

	username := "seed_" + sha1Short(key)

	var u usermodel.User
	result := db.Where("username = ?", username).Limit(1).Find(&u)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected > 0 {
		cache[key] = &u
		return &u, nil
	}

	avatarURL := ""
	if avatarPublicPath != "" {
		url, err := ensureObjectFromPublicPath(ctx, assetRoot, avatarPublicPath, "seed", dryRun)
		if err == nil {
			avatarURL = url
		}
	}

	u = usermodel.User{
		ID:       snowflake.GenerateMsgID(),
		Username: username,
		Nickname: strings.TrimSpace(authorName),
		Password: passwordHash,
		Role:     "user",
		Avatar:   avatarURL,
	}

	if dryRun {
		cache[key] = &u
		return &u, nil
	}

	if err := db.Create(&u).Error; err != nil {
		return nil, err
	}

	cache[key] = &u
	return &u, nil
}

func upsertVideo(db *gorm.DB, v *contentmodel.Video, dryRun bool) error {
	if dryRun {
		return nil
	}

	var existing contentmodel.Video
	result := db.Where("id = ?", v.ID).Limit(1).Find(&existing)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		existing.AuthorID = v.AuthorID
		existing.Title = v.Title
		existing.PlayURL = v.PlayURL
		existing.CoverURL = v.CoverURL
		existing.RecommendScore = v.RecommendScore
		existing.FavoriteCount = v.FavoriteCount
		existing.CommentCount = v.CommentCount
		existing.Status = v.Status
		return db.Save(&existing).Error
	}
	return db.Create(v).Error
}

func ensureObjectFromPublicPath(ctx context.Context, assetRoot string, publicPath string, prefix string, dryRun bool) (string, error) {
	rel := strings.TrimPrefix(strings.TrimSpace(publicPath), "/")
	if rel == "" {
		return "", fmt.Errorf("empty public path")
	}

	localPath := filepath.Join(assetRoot, filepath.FromSlash(rel))
	if _, err := os.Stat(localPath); err != nil {
		return "", err
	}

	objectName := prefix + "/" + rel
	objectName = strings.ReplaceAll(objectName, "\\", "/")

	if dryRun {
		return minio.GetObjectURL(objectName), nil
	}

	_, err := minio.Client.StatObject(ctx, config.Config.MinIO.BucketName, objectName, miniogo.StatObjectOptions{})
	if err == nil {
		return minio.GetObjectURL(objectName), nil
	}

	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(localPath)))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	url, err := minio.UploadLocalFile(ctx, objectName, localPath, contentType)
	if err != nil {
		return "", err
	}
	return url, nil
}

func sha1Short(s string) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:])[:12]
}
