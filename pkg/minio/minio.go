package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"wishlist/internal/utils/colors"
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

func NewClient(ctx context.Context, cfg Config) (*minio.Client, error) {
	fmt.Printf("Connecting to MinIO S3...")

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		fmt.Println()
		return nil, fmt.Errorf("MinIO: failed to initialize client: %w", err)
	}

	if _, err = client.BucketExists(ctx, cfg.BucketName); err != nil {
		fmt.Println()
		return nil, fmt.Errorf("MinIO: failed to connect: %w", err)
	}

	fmt.Println(colors.Green(" Done."))
	return client, nil
}
