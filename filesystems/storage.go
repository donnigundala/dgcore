package filesystems

import (
	"context"
	"io"
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
