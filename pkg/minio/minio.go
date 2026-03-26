package minio

import (
	"context"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client

func Init() {
	var err error
	Client, err = minio.New(config.Config.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Config.MinIO.AccessKeyID, config.Config.MinIO.SecretAccessKey, ""),
		Secure: config.Config.MinIO.UseSSL,
	})
	if err != nil {
		panic(err)
	}

	// 检查 bucket 是否存在，不存在则创建
	ctx := context.Background()
	bucketName := config.Config.MinIO.BucketName
	exists, err := Client.BucketExists(ctx, bucketName)
	if err != nil {
		logger.Log.Fatal("Failed to check if bucket exists: " + err.Error())
	}
	if !exists {
		err = Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			logger.Log.Fatal("Failed to create bucket: " + err.Error())
		}
		logger.Log.Info("Successfully created bucket: " + bucketName)
		
		// 设置 bucket 策略为 public 只读
		policy := `{"Version":"2012-10-17","Statement":[{"Action":["s3:GetObject"],"Effect":"Allow","Principal":{"AWS":["*"]},"Resource":["arn:aws:s3:::` + bucketName + `/*"]}]}`
		err = Client.SetBucketPolicy(ctx, bucketName, policy)
		if err != nil {
			logger.Log.Fatal("Failed to set bucket policy: " + err.Error())
		}
	}
}

// GeneratePresignedPutURL 生成用于上传文件的预签名 URL
func GeneratePresignedPutURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	url, err := Client.PresignedPutObject(ctx, config.Config.MinIO.BucketName, objectName, expires)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

// GetObjectURL 获取文件的公开访问 URL
func GetObjectURL(objectName string) string {
	protocol := "http://"
	if config.Config.MinIO.UseSSL {
		protocol = "https://"
	}
	// 返回格式：http://endpoint/bucketName/objectName
	return protocol + config.Config.MinIO.Endpoint + "/" + config.Config.MinIO.BucketName + "/" + objectName
}
