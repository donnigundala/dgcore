package storage

import (
	"fmt"
	"mime/multipart"
	"net/url"

	"github.com/donnigundala/dg-frame/pkg/storage/local"
	"github.com/donnigundala/dg-frame/pkg/storage/minio"
	"github.com/donnigundala/dg-frame/pkg/storage/s3"
)

type Storage interface {
	EnsureBucket(bucket string) error
	UploadFile(bucket, objectName, contentType string, file multipart.File) (string, error)
	DownloadFile(bucket, objectName string) ([]byte, error)
	DeleteFile(bucket, objectName string) error
	GetSignedURL(bucket, objectName string, expirySeconds int64, reqParams url.Values) (string, error)
}

// NewStorage - factory method to create a new Storage instance based on the driver type.
// TODO: Add support for GCS and Azure Blob Storage.
func NewStorage(driver string, cfg any) (Storage, error) {
	switch driver {
	case "minio":
		return minio.NewMinio(cfg.(*minio.Config))
	case "local":
		return local.NewLocal(cfg.(*local.Config))
	case "s3":
		return s3.NewS3(cfg.(*s3.Config))
		//return NewS3(cfg)
	// case "gcs":
	// 	return NewGCS(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", driver)
	}
}
