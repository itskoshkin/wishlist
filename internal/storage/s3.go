package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"

	"wishlist/internal/config"
)

const (
	AvatarPrefix = "avatars/%s/%s"
)

type MinioServiceImpl struct {
	client *minio.Client
	bucket string
}

func NewMinioService(c *minio.Client) *MinioServiceImpl {
	return &MinioServiceImpl{client: c, bucket: viper.GetString(config.MinioBucketName)}
}

func (svc *MinioServiceImpl) GetBaseURL() string {
	return fmt.Sprintf("%s/%s", svc.client.EndpointURL().String(), svc.bucket)
}

func (svc *MinioServiceImpl) UploadObject(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := svc.client.PutObject(ctx, svc.bucket, objectName, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}

	return nil
}

func (svc *MinioServiceImpl) GetObjectURL(objectName string) string {
	return fmt.Sprintf("%s/%s/%s", svc.client.EndpointURL().String(), svc.bucket, objectName)
}

func (svc *MinioServiceImpl) DeleteObject(ctx context.Context, objectName string) error {
	if err := svc.client.RemoveObject(ctx, svc.bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}
