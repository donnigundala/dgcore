package dgminio

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client *minio.Client
	cfg    *Config
}

// New Create new Minio storage
func New(cfg *Config) (*Storage, error) {
	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.Secure,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	if err != nil {
		return nil, err
	}

	return &Storage{
		client: cli,
		cfg:    cfg,
	}, nil
}

// EnsureBucket Ensure bucket exists
func (m *Storage) EnsureBucket(bucket string) error {
	ctx := context.Background()
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		return m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	}
	return nil
}

// UploadFile Upload file to bucket
func (m *Storage) UploadFile(bucket, objectName, contentType string, file multipart.File) (string, error) {
	ctx := context.Background()
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return "", err
	}

	_, err := m.client.PutObject(ctx, bucket, objectName, bytes.NewReader(buf.Bytes()), int64(buf.Len()), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", m.cfg.Endpoint, bucket, objectName), nil
}

// DownloadFile Download file from bucket
func (m *Storage) DownloadFile(bucket, objectName string) ([]byte, error) {
	ctx := context.Background()
	obj, err := m.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer func(obj *minio.Object) {
		err := obj.Close()
		if err != nil {
			fmt.Println("Error closing object:", err)
		}
	}(obj)

	return io.ReadAll(obj)
}

// DeleteFile Delete file from bucket
func (m *Storage) DeleteFile(bucket, objectName string) error {
	ctx := context.Background()
	return m.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

// GetSignedURL Signed URL support
func (m *Storage) GetSignedURL(bucket, objectName string, expirySeconds int64, reqParams url.Values) (string, error) {
	ctx := context.Background()
	generatedUrl, err := m.client.PresignedGetObject(ctx, bucket, objectName, time.Duration(expirySeconds)*time.Second, reqParams)
	if err != nil {
		return "", err
	}
	return generatedUrl.String(), nil
}
