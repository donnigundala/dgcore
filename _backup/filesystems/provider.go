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
func newStorage(driver string, config interface{}, logger *slog.Logger) (Storage, error) {
	switch driver {
	case "local":
		// The traceIDKey is no longer needed.
		return newLocalDriver(config, logger)
	case "s3":
		// The traceIDKey is no longer needed.
		return newS3Driver(config, logger)
	case "minio":
		// The traceIDKey is no longer needed.
		return newMinioDriver(config, logger)
	default:
		return nil, fmt.Errorf("unsupported filesystem driver: %s", driver)
	}
}
