package filesystems

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	minio "github.com/minio/minio-go/v7"
	creds "github.com/minio/minio-go/v7/pkg/credentials"
)

// register converter and driver
func init() {
	RegisterConfigConverter("minio", func(cfg map[string]interface{}) (interface{}, error) {
		m := MinIOConfig{}
		if v, ok := cfg["endpoint"].(string); ok {
			m.Endpoint = v
		}
		if v, ok := cfg["accessKeyID"].(string); ok {
			m.AccessKeyID = v
		}
		if v, ok := cfg["secretAccessKey"].(string); ok {
			m.SecretAccessKey = v
		}
		if v, ok := cfg["useSSL"].(bool); ok {
			m.UseSSL = v
		}
		if v, ok := cfg["bucket"].(string); ok {
			m.Bucket = v
		}
		if v, ok := cfg["baseURL"].(string); ok {
			m.BaseURL = v
		}
		if m.Endpoint == "" || m.Bucket == "" || m.AccessKeyID == "" || m.SecretAccessKey == "" {
			return nil, fmt.Errorf("minio: endpoint, bucket, accessKeyID and secretAccessKey are required")
		}
		return m, nil
	})

	RegisterDriver("minio", func(cfg interface{}) (Storage, error) {
		c, ok := cfg.(MinIOConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for minio")
		}
		return NewMinioDriver(c)
	})
}

// MinioDriver implements Storage using MinIO.
type MinioDriver struct {
	client *minio.Client
	cfg    MinIOConfig
}

func NewMinioDriver(cfg MinIOConfig) (*MinioDriver, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  creds.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: create client: %w", err)
	}

	// Ensure bucket exists (best-effort)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	exists, err := minioClient.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("minio: bucket exists check failed: %w", err)
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			// non-fatal in some environments: just return error
			return nil, fmt.Errorf("minio: create bucket: %w", err)
		}
	}

	return &MinioDriver{client: minioClient, cfg: cfg}, nil
}

func (m *MinioDriver) Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error {
	opts := minio.PutObjectOptions{}
	if visibility == VisibilityPublic {
		opts.UserMetadata = map[string]string{"x-amz-acl": "public-read"}
	}
	_, err := m.client.PutObject(ctx, m.cfg.Bucket, key, data, size, opts)
	if err != nil {
		return fmt.Errorf("minio: put object: %w", err)
	}
	return nil
}

func (m *MinioDriver) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := m.client.GetObject(ctx, m.cfg.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		// Check if the error is a not-found error before trying to parse it.
		errResponse := minio.ToErrorResponse(err)
		if errResponse.StatusCode == 404 {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("minio: get object: %w", err)
	}
	// Stat the object to ensure it exists before returning.
	_, err = obj.Stat()
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.StatusCode == 404 {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("minio: stat object: %w", err)
	}
	return obj, nil
}

func (m *MinioDriver) Delete(ctx context.Context, key string) error {
	err := m.client.RemoveObject(ctx, m.cfg.Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio: remove object: %w", err)
	}
	return nil
}

func (m *MinioDriver) Exists(ctx context.Context, key string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.cfg.Bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("minio: stat object: %w", err)
	}
	return true, nil
}

func (m *MinioDriver) GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error) {
	if visibility == VisibilityPublic && m.cfg.BaseURL != "" {
		return strings.TrimRight(m.cfg.BaseURL, "/") + "/" + strings.TrimLeft(key, "/"), nil
	}

	// Generate a presigned URL for private files or if no base URL is set.
	reqParams := make(url.Values)
	signedUrl, err := m.client.PresignedGetObject(ctx, m.cfg.Bucket, key, duration, reqParams)
	if err != nil {
		return "", fmt.Errorf("minio: presigned url: %w", err)
	}
	return signedUrl.String(), nil
}

func (m *MinioDriver) List(ctx context.Context, prefix string) ([]string, error) {
	// Use a channel to stream results and avoid high memory usage.
	objectCh := m.client.ListObjects(ctx, m.cfg.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var results []string
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("minio: list objects: %w", object.Err)
		}
		results = append(results, object.Key)
	}

	return results, nil
}
