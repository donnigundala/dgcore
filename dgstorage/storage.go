package dgstorage

import (
	"fmt"
	"mime/multipart"
	"net/url"

	"github.com/donnigundala/dgcore/dgstorage/dglocal"
	"github.com/donnigundala/dgcore/dgstorage/dgminio"
	"github.com/donnigundala/dgcore/dgstorage/dgs3"
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
		return dgminio.New(cfg.(*dgminio.Config))
	case "local":
		return dglocal.New(cfg.(*dglocal.Config))
	case "s3":
		return dgs3.New(cfg.(*dgs3.Config))
	// case "gcs":
	// 	return NewGCS(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", driver)
	}
}
