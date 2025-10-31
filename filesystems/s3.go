package filesystems

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3ConfigWithAuth defines configuration for AWS S3 driver.
type S3ConfigWithAuth struct {
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	BaseURL   string
}

// register converter and driver
func init() {
	RegisterConfigConverter("s3", func(cfg map[string]interface{}) (interface{}, error) {
		c := S3ConfigWithAuth{}
		if v, ok := cfg["bucket"].(string); ok {
			c.Bucket = v
		}
		if v, ok := cfg["region"].(string); ok {
			c.Region = v
		}
		if v, ok := cfg["accessKey"].(string); ok {
			c.AccessKey = v
		}
		if v, ok := cfg["secretKey"].(string); ok {
			c.SecretKey = v
		}
		if v, ok := cfg["baseURL"].(string); ok {
			c.BaseURL = v
		}
		if c.Bucket == "" || c.Region == "" {
			return nil, fmt.Errorf("s3: bucket and region are required")
		}
		return c, nil
	})

	RegisterDriver("s3", func(config interface{}, logger *slog.Logger, traceIDKey any) (Storage, error) {
		cfg, ok := config.(S3ConfigWithAuth)
		if !ok {
			return nil, fmt.Errorf("invalid config for s3 driver")
		}
		return NewS3Driver(cfg, logger, traceIDKey)
	})
}

// S3Driver implements the Storage interface for AWS S3.
type S3Driver struct {
	client     *s3.Client
	cfg        S3ConfigWithAuth
	logger     *slog.Logger
	traceIDKey any
}

func NewS3Driver(cfg S3ConfigWithAuth, logger *slog.Logger, traceIDKey any) (*S3Driver, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("s3: load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	return &S3Driver{client: client, cfg: cfg, logger: logger, traceIDKey: traceIDKey}, nil
}

func (s *S3Driver) loggerFrom(ctx context.Context) *slog.Logger {
	logger := s.logger
	if s.traceIDKey != nil {
		if traceID, ok := ctx.Value(s.traceIDKey).(string); ok && traceID != "" {
			logger = logger.With(slog.String("trace_id", traceID))
		}
	}
	return logger
}

func (s *S3Driver) Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error {
	logger := s.loggerFrom(ctx)
	uploader := manager.NewUploader(s.client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
		Body:   data,
	})
	if err != nil {
		logger.Error("Failed to upload to S3", "key", key, "error", err)
		return fmt.Errorf("s3: upload: %w", err)
	}
	logger.Debug("Uploaded to S3", "key", key)
	return nil
}

func (s *S3Driver) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	logger := s.loggerFrom(ctx)
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logger.Error("Failed to download from S3", "key", key, "error", err)
		return nil, fmt.Errorf("s3: get object: %w", err)
	}
	return out.Body, nil
}

func (s *S3Driver) Delete(ctx context.Context, key string) error {
	logger := s.loggerFrom(ctx)
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logger.Error("Failed to delete from S3", "key", key, "error", err)
		return fmt.Errorf("s3: delete: %w", err)
	}
	logger.Debug("Deleted from S3", "key", key)
	return nil
}

func (s *S3Driver) Exists(ctx context.Context, key string) (bool, error) {
	logger := s.loggerFrom(ctx)
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nfe *types.NotFound
		if errors.As(err, &nfe) {
			logger.Debug("Object not found in S3", "key", key)
			return false, nil
		}
		logger.Warn("Failed to check existence in S3", "key", key, "error", err)
		return false, fmt.Errorf("s3: head object: %w", err)
	}
	return true, nil
}

func (s *S3Driver) GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error) {
	logger := s.loggerFrom(ctx)
	if s.cfg.BaseURL != "" && visibility == VisibilityPublic {
		return strings.TrimRight(s.cfg.BaseURL, "/") + "/" + strings.TrimLeft(key, "/"), nil
	}
	presignClient := s3.NewPresignClient(s.client)
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(duration))
	if err != nil {
		logger.Error("Failed to generate presigned URL for S3", "key", key, "error", err)
		return "", fmt.Errorf("s3: presign url: %w", err)
	}
	return req.URL, nil
}

func (s *S3Driver) List(ctx context.Context, prefix string) ([]string, error) {
	logger := s.loggerFrom(ctx)
	out, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.cfg.Bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		logger.Error("Failed to list objects in S3", "prefix", prefix, "error", err)
		return nil, fmt.Errorf("s3: list objects: %w", err)
	}
	keys := make([]string, 0, len(out.Contents))
	for _, obj := range out.Contents {
		keys = append(keys, *obj.Key)
	}
	return keys, nil
}
