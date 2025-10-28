package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Storage implements Storage interface for AWS S3
type S3Storage struct {
	client     *s3.Client
	bucket     string
	region     string
	uploader   *manager.Uploader
	downloader *manager.Downloader
}

// S3Config holds S3-specific configuration
type S3Config struct {
	Bucket string
	Region string
	Client *s3.Client // Pre-configured S3 client
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	if cfg.Client == nil {
		return nil, fmt.Errorf("S3 client is required")
	}

	return &S3Storage{
		client:     cfg.Client,
		bucket:     cfg.Bucket,
		region:     cfg.Region,
		uploader:   manager.NewUploader(cfg.Client),
		downloader: manager.NewDownloader(cfg.Client),
	}, nil
}

// Upload stores data to S3 with specified visibility
func (s3s *S3Storage) Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error {
	acl := types.ObjectCannedACLPublicRead
	if visibility == Private {
		acl = types.ObjectCannedACLPrivate
	}

	_, err := s3s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
		Body:   data,
		ACL:    acl,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}
	return nil
}

// Download retrieves data from S3
func (s3s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s3s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	return result.Body, nil
}

// Delete removes an object from S3
func (s3s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s3s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	return nil
}

// Exists checks if an object exists in S3
func (s3s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s3s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// GetURL generates a URL for accessing the object.
func (s3s *S3Storage) GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error) {
	if visibility == Public {
		return s3s.getPublicURL(key), nil
	}
	return s3s.getSignedURL(ctx, key, duration)
}

// List returns all keys with the given prefix
func (s3s *S3Storage) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string

	paginator := s3.NewListObjectsV2Paginator(s3s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3s.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
	}

	return keys, nil
}

// getPublicURL returns the direct URL for a public object.
func (s3s *S3Storage) getPublicURL(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3s.bucket, s3s.region, key)
}

// getSignedURL generates a temporary signed URL for a private object.
func (s3s *S3Storage) getSignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3s.client)

	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return presignResult.URL, nil
}
