package minio

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOConfig struct {
	Endpoint   string `env:"MINIO_ENDPOINT" env-default:"localhost:9000"`
	AccessKey  string `env:"MINIO_ACCESSKEY" env-default:"minioadmin"`
	Secret     string `env:"MINIO_SECRET" env-default:"minioadmin"`
	UseSSL     bool   `env:"MINIO_SSL" env-default:"FALSE"`
	BucketName string `env:"MINIO_BUCKET" env-default:"images"`
}

type MinIOClient struct {
	Client     *minio.Client
	BucketName string
	config     MinIOConfig
}

func NewMinIOClient(cfg MinIOConfig) (*MinIOClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.Secret, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	mc := &MinIOClient{
		Client:     client,
		BucketName: cfg.BucketName,
		config:     cfg,
	}

	if err := mc.ensureBucketExists(); err != nil {
		return nil, err
	}

	return mc, nil
}

func (mc *MinIOClient) ensureBucketExists() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := mc.Client.BucketExists(ctx, mc.BucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {

		err = mc.Client.MakeBucket(ctx, mc.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		log.Printf("Bucket '%s' created successfully", mc.BucketName)
	}
	policy := fmt.Sprintf(
		`{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {"AWS": ["*"]},
            "Action": ["s3:GetObject"],
            "Resource": ["arn:aws:s3:::%s/*"]
        }
    ]
}`, mc.BucketName)

	mc.Client.SetBucketPolicy(ctx, mc.BucketName, policy)

	return nil
}
