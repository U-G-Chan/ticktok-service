package minio

import (
	"ticktok-service/pkg/config"

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
}
