package s3

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage struct {
	client *s3.Client
	ps     *s3.PresignClient
	cfg    *Config
}

func NewS3(cfg *Config) (*Storage, error) {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	ps := s3.NewPresignClient(client)

	return &Storage{client: client, ps: ps, cfg: cfg}, nil
}

// EnsureBucket checks if the bucket exists and creates it if it doesn't.
func (s *Storage) EnsureBucket(bucket string) error {
	ctx := context.TODO()
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err == nil {
		// Bucket already exists
		return nil
	}

	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}
	return nil
}

// DownloadFile downloads a file from the specified S3 bucket.
func (s *Storage) DownloadFile(bucket string, objectName string) ([]byte, error) {
	ctx := context.TODO()
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close object body: %v", err)
		}
	}(out.Body)

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}
	return data, nil
}

// DeleteFile deletes a file from the specified S3 bucket.
func (s *Storage) DeleteFile(bucket string, objectName string) error {
	ctx := context.TODO()
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	// Wait for the object to actually be deleted using AWS SDK v2 waiter
	waiter := s3.NewObjectNotExistsWaiter(s.client)
	err = waiter.Wait(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectName),
	}, 30*time.Second)

	if err != nil {
		return fmt.Errorf("waiting for object deletion failed: %w", err)
	}
	return nil
}

// UploadFile uploads a file to the specified S3 bucket and returns its URL.
//func (s *Storage) UploadFile(objectName string, data []byte) (string, error) {
//	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
//		Bucket: aws.String(s.cfg.Bucket),
//		Key:    aws.String(objectName),
//		Body:   bytes.NewReader(data),
//	})
//	if err != nil {
//		return "", fmt.Errorf("failed to upload to s3: %w", err)
//	}
//
//	if s.cfg.BaseURL != "" {
//		return fmt.Sprintf("%s/%s", s.cfg.BaseURL, objectName), nil
//	}
//
//	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.Bucket, s.cfg.Region, objectName), nil
//}

func (s *Storage) UploadFile(bucket string, objectName string, contentType string, file multipart.File) (string, error) {
	ctx := context.Background()
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(objectName),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("s3://%s/%s", bucket, objectName), nil
}

// GetSignedURL generates a presigned URL for accessing an object in S3.
func (s *Storage) GetSignedURL(bucket, objectName string, expirySeconds int64, reqParams url.Values) (string, error) {
	req, err := s.ps.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectName),
	}, s3.WithPresignExpires(time.Duration(expirySeconds)*time.Second))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return req.URL, nil
}
