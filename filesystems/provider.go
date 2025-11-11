package filesystems

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"
)

// Storage is the driver interface that each disk driver must implement.
type Storage interface {
	Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error)
	List(ctx context.Context, prefix string) ([]string, error)
}

// newStorage acts as an internal factory for creating a Storage provider.
// It is called by the Manager.
func newStorage(driver string, config interface{}, logger *slog.Logger, traceIDKey any) (Storage, error) {
	switch driver {
	case "local":
		// Assuming newLocalDriver is the constructor for the local driver
		return newLocalDriver(config, logger, traceIDKey)
	case "s3":
		// Assuming newS3Driver is the constructor for the s3 driver
		return newS3Driver(config, logger, traceIDKey)
	case "minio":
		// Assuming newMinioDriver is the constructor for the minio driver
		return newMinioDriver(config, logger, traceIDKey)
	default:
		return nil, fmt.Errorf("unsupported filesystem driver: %s", driver)
	}
}
