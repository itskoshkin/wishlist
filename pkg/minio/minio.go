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

	var exists bool
	if exists, err = client.BucketExists(ctx, cfg.BucketName); err != nil {
		fmt.Println()
		return nil, fmt.Errorf("MinIO: failed to connect: %w", err)
	} else if !exists {
		if err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{}); err != nil {
			fmt.Println()
			return nil, fmt.Errorf("MinIO: failed to create bucket '%s': %w", cfg.BucketName, err)
		}
	}

	allowPublicReadPolicy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadObjects",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "*"
        ]
      },
      "Action": [
        "s3:GetObject"
      ],
      "Resource": [
        "arn:aws:s3:::%s/*"
      ]
    }
  ]
}`, cfg.BucketName)
	
	if err = client.SetBucketPolicy(ctx, cfg.BucketName, allowPublicReadPolicy); err != nil {
		fmt.Println()
		return nil, fmt.Errorf("MinIO: failed to set public read policy for bucket '%s': %w", cfg.BucketName, err)
	}

	fmt.Println(colors.Green(" Done."))
	return client, nil
}
