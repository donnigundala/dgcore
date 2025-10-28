package filesystem

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Factory creates Storage instances based on configuration.
type Factory struct{}

// NewFactory creates a new storage factory.
func NewFactory() *Factory {
	return &Factory{}
}

// Create instantiates the appropriate storage provider.
func (f *Factory) Create(storageType string, cfg interface{}) (Storage, error) {
	storageType = strings.ToLower(strings.TrimSpace(storageType))

	switch storageType {
	case "local":
		localCfg, ok := cfg.(LocalConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for local storage")
		}
		return NewLocalStorage(localCfg)

	case "minio":
		minioCfg, ok := cfg.(MinIOConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for MinIO storage")
		}
		return NewMinIOStorage(minioCfg)

	case "s3":
		s3Cfg, ok := cfg.(S3ConfigWithAuth)
		if !ok {
			return nil, fmt.Errorf("invalid config type for S3 storage")
		}
		return NewS3StorageWithAuth(&s3Cfg)

	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// LocalConfig holds configuration for local storage.
type LocalConfig struct {
	BasePath string `json:"basePath"`
	BaseURL  string `json:"baseURL"`
	Secret   string `json:"secret"`
}

// S3ConfigWithAuth holds S3 configuration with authentication.
type S3ConfigWithAuth struct {
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

// NewS3StorageWithAuth creates S3 storage with AWS credentials.
func NewS3StorageWithAuth(cfg *S3ConfigWithAuth) (*S3Storage, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)

	return NewS3Storage(S3Config{
		Bucket: cfg.Bucket,
		Region: cfg.Region,
		Client: client,
	})
}
