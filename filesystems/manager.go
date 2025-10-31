package filesystems

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"time"
)

// ManagerConfig holds the configuration for the filesystem manager.
type ManagerConfig struct {
	Default string
	Disks   map[string]Disk
}

// Disk represents the configuration for a storage disk.
type Disk struct {
	Driver string
	Config map[string]interface{}
}

// FileSystem is the top-level manager for all storage disks.
type FileSystem struct {
	disks       map[string]Storage
	defaultDisk string
	logger      *slog.Logger
	traceIDKey  any
}

// Option configures a FileSystem.
type Option func(*FileSystem)

// WithTraceIDKey provides a context key to look for a trace ID. The value
// associated with this key should be a string.
func WithTraceIDKey(key any) Option {
	return func(fs *FileSystem) {
		fs.traceIDKey = key
	}
}

// WithLogger provides a slog logger for the filesystem manager.
// If not provided, logs will be discarded.
func WithLogger(logger *slog.Logger) Option {
	return func(fs *FileSystem) {
		if logger != nil {
			fs.logger = logger
		}
	}
}

// New creates a new FileSystem manager from the given configuration.
func New(config ManagerConfig, opts ...Option) (*FileSystem, error) {
	fs := &FileSystem{
		// Default to a silent logger. This can be overridden by the WithLogger option.
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	for _, opt := range opts {
		opt(fs)
	}

	if len(config.Disks) == 0 {
		return nil, fmt.Errorf("no disks configured")
	}

	fs.logger.Debug("Initializing filesystem manager", "default_disk", config.Default, "disk_count", len(config.Disks))

	disks := make(map[string]Storage)
	factory := NewFactory()

	for name, disk := range config.Disks {
		driverConfig, err := convertConfig(disk.Driver, disk.Config)
		if err != nil {
			fs.logger.Error("Failed to convert config for disk", "disk", name, "driver", disk.Driver, "error", err)
			return nil, fmt.Errorf("failed to create config for disk '%s': %w", name, err)
		}

		storageDriver, err := factory.Create(disk.Driver, driverConfig, fs.logger.With("disk", name, "driver", disk.Driver), fs.traceIDKey)
		if err != nil {
			fs.logger.Error("Failed to create driver for disk", "disk", name, "driver", disk.Driver, "error", err)
			return nil, fmt.Errorf("failed to create driver for disk '%s': %w", name, err)
		}
		disks[name] = storageDriver
	}

	// If default not provided, choose one deterministically
	defaultDisk := config.Default
	if defaultDisk == "" {
		names := make([]string, 0, len(disks))
		for name := range disks {
			names = append(names, name)
		}
		sort.Strings(names)
		if len(names) > 0 {
			defaultDisk = names[0]
		}
	}

	if _, ok := disks[defaultDisk]; !ok {
		if len(disks) == 0 {
			return nil, fmt.Errorf("no disks configured to select a default from")
		}
		return nil, fmt.Errorf("default disk '%s' not found in configured disks", defaultDisk)
	}

	fs.disks = disks
	fs.defaultDisk = defaultDisk

	fs.logger.Info("Filesystem manager initialized", "default_disk", fs.defaultDisk)

	return fs, nil
}

// Disk returns a specific storage driver by name. Panics if not configured.
func (fs *FileSystem) Disk(name string) Storage {
	disk, ok := fs.disks[name]
	if !ok {
		fs.logger.Error("Requested disk not configured", "disk", name)
		panic(fmt.Sprintf("disk '%s' not configured", name))
	}
	return disk
}

// GetDisk returns the disk and an error if not found (non-panicking).
func (fs *FileSystem) GetDisk(name string) (Storage, error) {
	disk, ok := fs.disks[name]
	if !ok {
		fs.logger.Warn("Requested disk not configured", "disk", name)
		return nil, fmt.Errorf("disk '%s' not configured", name)
	}
	return disk, nil
}

// Default-disk convenience methods

func (fs *FileSystem) Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error {
	return fs.Disk(fs.defaultDisk).Upload(ctx, key, data, size, visibility)
}

func (fs *FileSystem) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return fs.Disk(fs.defaultDisk).Download(ctx, key)
}

func (fs *FileSystem) Delete(ctx context.Context, key string) error {
	return fs.Disk(fs.defaultDisk).Delete(ctx, key)
}

func (fs *FileSystem) Exists(ctx context.Context, key string) (bool, error) {
	return fs.Disk(fs.defaultDisk).Exists(ctx, key)
}

func (fs *FileSystem) GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error) {
	return fs.Disk(fs.defaultDisk).GetURL(ctx, key, visibility, duration)
}

func (fs *FileSystem) List(ctx context.Context, prefix string) ([]string, error) {
	return fs.Disk(fs.defaultDisk).List(ctx, prefix)
}
