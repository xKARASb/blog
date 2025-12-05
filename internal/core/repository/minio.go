package repository

import (
	"context"
	"fmt"
	"io"

	minIO "github.com/minio/minio-go/v7"
	"github.com/xkarasb/blog/pkg/storage/minio"
)

type MinIORepository struct {
	Storage *minio.MinIOClient
}

func NewMinIORepository(storage *minio.MinIOClient) *MinIORepository {
	return &MinIORepository{storage}
}

func (rep *MinIORepository) PutImage(objectName string, file io.Reader, fileSize int64, contentType string) (string, error) {
	info, err := rep.Storage.Client.PutObject(
		context.Background(),
		rep.Storage.BucketName,
		objectName,
		file,
		fileSize,
		minIO.PutObjectOptions{
			ContentType: contentType,
		},
	)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/%s/%s", info.Bucket, objectName), nil
}

func (rep *MinIORepository) DeleteImage(objectName string) error {
	return rep.Storage.Client.RemoveObject(context.Background(), rep.Storage.BucketName, objectName, minIO.RemoveObjectOptions{})
}
