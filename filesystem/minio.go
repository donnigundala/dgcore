package filesystem

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage implements Storage interface for MinIO
type MinIOStorage struct {
	client  *minio.Client
	bucket  string
	baseURL string
}

// MinIOConfig holds MinIO-specific configuration
type MinIOConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	UseSSL          bool   `json:"useSSL"`
	Bucket          string `json:"bucket"`
	BaseURL         string `json:"baseURL"`
}

// NewMinIOStorage creates a new MinIO storage instance
func NewMinIOStorage(cfg MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &MinIOStorage{
		client:  client,
		bucket:  cfg.Bucket,
		baseURL: cfg.BaseURL,
	}, nil
}

// Upload stores data to MinIO with specified visibility
func (ms *MinIOStorage) Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error {
	opts := minio.PutObjectOptions{
		UserMetadata: map[string]string{"visibility": string(visibility)},
	}

	_, err := ms.client.PutObject(ctx, ms.bucket, key, data, size, opts)
	if err != nil {
		return fmt.Errorf("failed to upload to MinIO: %w", err)
	}
	return nil
}

// Download retrieves data from MinIO
func (ms *MinIOStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	object, err := ms.client.GetObject(ctx, ms.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download from MinIO: %w", err)
	}
	return object, nil
}

// Delete removes an object from MinIO
func (ms *MinIOStorage) Delete(ctx context.Context, key string) error {
	err := ms.client.RemoveObject(ctx, ms.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete from MinIO: %w", err)
	}
	return nil
}

// Exists checks if an object exists in MinIO
func (ms *MinIOStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := ms.client.StatObject(ctx, ms.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// GetURL generates a URL for accessing the object.
func (ms *MinIOStorage) GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error) {
	if visibility == Public {
		return ms.getPublicURL(key), nil
	}
	return ms.getSignedURL(ctx, key, duration)
}

// List returns all keys with the given prefix
func (ms *MinIOStorage) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string

	objectCh := ms.client.ListObjects(ctx, ms.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		keys = append(keys, object.Key)
	}

	return keys, nil
}

// getPublicURL returns the direct URL for a public object.
func (ms *MinIOStorage) getPublicURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", ms.baseURL, ms.bucket, key)
}

// getSignedURL generates a temporary signed URL for a private object.
func (ms *MinIOStorage) getSignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	url, err := ms.client.PresignedGetObject(ctx, ms.bucket, key, duration, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	return url.String(), nil
}
