package local

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Storage struct {
	basePath string
	cfg      *Config
}

func NewLocal(cfg *Config) (*Storage, error) {
	path := "storage"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return nil, err
		}
	}
	return &Storage{basePath: path}, nil
}

func (l *Storage) EnsureBucket(bucket string) error {
	return os.MkdirAll(filepath.Join(l.basePath, bucket), 0755)
}

func (l *Storage) UploadFile(bucket, objectName, contentType string, file multipart.File) (string, error) {
	if err := l.EnsureBucket(bucket); err != nil {
		return "", err
	}

	dst := filepath.Join(l.basePath, bucket, objectName)
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, file); err != nil {
		return "", err
	}

	return strings.TrimPrefix(dst, l.basePath+"/"), nil
}

func (l *Storage) DownloadFile(bucket, objectName string) ([]byte, error) {
	return os.ReadFile(filepath.Join(l.basePath, bucket, objectName))
}

func (l *Storage) DeleteFile(bucket, objectName string) error {
	return os.Remove(filepath.Join(l.basePath, bucket, objectName))
}

// GetSignedURL Local storage "signed URL" is just a fake URL with timestamp
func (l *Storage) GetSignedURL(bucket, objectName string, expirySeconds int64, reqParams url.Values) (string, error) {
	// In production you might serve this via a static file server
	return fmt.Sprintf("/%s/%s?expires=%d", bucket, objectName, time.Now().Add(time.Duration(expirySeconds)*time.Second).Unix()), nil
}
