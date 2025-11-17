package filesystems

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"time"

	minio "github.com/minio/minio-go/v7"
	creds "github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/donnigundala/dgcore/ctxutil"
)

// MinIOConfig holds configuration for the MinIO driver.
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Bucket          string
	BaseURL         string
}

// newMinioDriver is the constructor for the MinIO driver, returning a Storage interface.
func newMinioDriver(config interface{}, logger *slog.Logger) (Storage, error) {
	cfg, ok := config.(MinIOConfig)
	if !ok {
		raw, ok := config.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid config type for minio driver: %T", config)
		}
		converted, err := convertMinioConfig(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to convert config for minio driver: %w", err)
		}
		cfg = converted
	}
	return newMinioDriverInternal(cfg, logger)
}

// convertMinioConfig handles the conversion from a map to MinIOConfig struct.
func convertMinioConfig(cfg map[string]interface{}) (MinIOConfig, error) {
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
		return m, fmt.Errorf("minio: endpoint, bucket, accessKeyID and secretAccessKey are required")
	}
	return m, nil
}

// MinioDriver implements Storage using MinIO.
type MinioDriver struct {
	client *minio.Client
	cfg    MinIOConfig
	logger *slog.Logger // Base logger for the driver
}

// newMinioDriverInternal is the internal constructor for MinioDriver.
func newMinioDriverInternal(cfg MinIOConfig, logger *slog.Logger) (*MinioDriver, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  creds.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		logger.Error("Failed to create MinIO client", "endpoint", cfg.Endpoint, "error", err)
		return nil, fmt.Errorf("minio: create client: %w", err)
	}

	// Ensure bucket exists (best-effort)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Debug("Checking if MinIO bucket exists", "bucket", cfg.Bucket)
	exists, err := minioClient.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		logger.Error("Failed to check if MinIO bucket exists", "bucket", cfg.Bucket, "error", err)
		return nil, fmt.Errorf("minio: bucket exists check failed: %w", err)
	}
	if !exists {
		logger.Info("MinIO bucket does not exist, creating it", "bucket", cfg.Bucket)
		if err := minioClient.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			logger.Error("Failed to create MinIO bucket", "bucket", cfg.Bucket, "error", err)
			return nil, fmt.Errorf("minio: create bucket: %w", err)
		}
	}

	return &MinioDriver{client: minioClient, cfg: cfg, logger: logger}, nil
}

// loggerFrom retrieves the context-aware logger.
func (m *MinioDriver) loggerFrom(ctx context.Context) *slog.Logger {
	// Use the ctxutil helper to get the logger from the context.
	// This logger will already contain the request_id if present.
	return ctxutil.LoggerFromContext(ctx)
}

func (m *MinioDriver) Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error {
	logger := m.loggerFrom(ctx)
	opts := minio.PutObjectOptions{}
	if visibility == VisibilityPublic {
		opts.UserMetadata = map[string]string{"x-amz-acl": "public-read"}
	}
	_, err := m.client.PutObject(ctx, m.cfg.Bucket, key, data, size, opts)
	if err != nil {
		logger.Error("Failed to upload to MinIO", "key", key, "bucket", m.cfg.Bucket, "error", err)
		return fmt.Errorf("minio: put object: %w", err)
	}
	logger.Debug("Uploaded to MinIO", "key", key, "bucket", m.cfg.Bucket)
	return nil
}

func (m *MinioDriver) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	logger := m.loggerFrom(ctx)
	obj, err := m.client.GetObject(ctx, m.cfg.Bucket, key, minio.GetObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.StatusCode == 404 {
			logger.Debug("Object not found in MinIO for download", "key", key, "bucket", m.cfg.Bucket)
			return nil, ErrNotFound
		}
		logger.Error("Failed to get object from MinIO", "key", key, "bucket", m.cfg.Bucket, "error", err)
		return nil, fmt.Errorf("minio: get object: %w", err)
	}
	// Stat the object to ensure it exists before returning.
	_, err = obj.Stat()
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.StatusCode == 404 {
			logger.Debug("Object not found in MinIO on stat after get", "key", key, "bucket", m.cfg.Bucket)
			return nil, ErrNotFound
		}
		logger.Error("Failed to stat object from MinIO after get", "key", key, "bucket", m.cfg.Bucket, "error", err)
		return nil, fmt.Errorf("minio: stat object: %w", err)
	}
	return obj, nil
}

func (m *MinioDriver) Delete(ctx context.Context, key string) error {
	logger := m.loggerFrom(ctx)
	err := m.client.RemoveObject(ctx, m.cfg.Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		logger.Error("Failed to delete object from MinIO", "key", key, "bucket", m.cfg.Bucket, "error", err)
		return fmt.Errorf("minio: remove object: %w", err)
	}
	logger.Debug("Deleted object from MinIO", "key", key, "bucket", m.cfg.Bucket)
	return nil
}

func (m *MinioDriver) Exists(ctx context.Context, key string) (bool, error) {
	logger := m.loggerFrom(ctx)
	_, err := m.client.StatObject(ctx, m.cfg.Bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.StatusCode == 404 {
			return false, nil
		}
		logger.Warn("Failed to check existence in MinIO", "key", key, "bucket", m.cfg.Bucket, "error", err)
		return false, fmt.Errorf("minio: stat object: %w", err)
	}
	return true, nil
}

func (m *MinioDriver) GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error) {
	logger := m.loggerFrom(ctx)
	if visibility == VisibilityPublic && m.cfg.BaseURL != "" {
		return strings.TrimRight(m.cfg.BaseURL, "/") + "/" + strings.TrimLeft(key, "/"), nil
	}

	// Generate a presigned URL for private files or if no base URL is set.
	reqParams := make(url.Values)
	signedUrl, err := m.client.PresignedGetObject(ctx, m.cfg.Bucket, key, duration, reqParams)
	if err != nil {
		logger.Error("Failed to generate presigned URL for MinIO", "key", key, "bucket", m.cfg.Bucket, "error", err)
		return "", fmt.Errorf("minio: presigned url: %w", err)
	}
	return signedUrl.String(), nil
}

func (m *MinioDriver) List(ctx context.Context, prefix string) ([]string, error) {
	logger := m.loggerFrom(ctx)
	// Use a channel to stream results and avoid high memory usage.
	objectCh := m.client.ListObjects(ctx, m.cfg.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var results []string
	for object := range objectCh {
		if object.Err != nil {
			logger.Error("Error while listing objects from MinIO", "prefix", prefix, "bucket", m.cfg.Bucket, "error", object.Err)
			return nil, fmt.Errorf("minio: list objects: %w", object.Err)
		}
		results = append(results, object.Key)
	}

	return results, nil
}
