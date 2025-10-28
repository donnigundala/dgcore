package filesystem

import (
	"context"
	"io"
	"time"
)

// Visibility defines the access level for a stored object.
type Visibility string

const (
	// Public objects are accessible directly via URL.
	Public Visibility = "public"
	// Private objects require signed URLs or authentication for access.
	Private Visibility = "private"
)

// Storage defines the interface for all storage providers
type Storage interface {
	// Upload stores data at the given key with specified visibility.
	Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error

	// Download retrieves data from the given key.
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes the object at the given key.
	Delete(ctx context.Context, key string) error

	// Exists checks if an object exists at the given key.
	Exists(ctx context.Context, key string) (bool, error)

	// GetURL generates a URL for accessing the object. For private objects,
	// this will be a signed URL with the given duration.
	GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error)

	// List returns all keys with the given prefix.
	List(ctx context.Context, prefix string) ([]string, error)
}

// Config holds common configuration for storage providers
type Config struct {
	Type   string // "local", "minio", "s3"
	Bucket string
}

// Object represents metadata about a stored object
type Object struct {
	Key          string
	Size         int64
	LastModified int64 // Unix timestamp
	ContentType  string
}
