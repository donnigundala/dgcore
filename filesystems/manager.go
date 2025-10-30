package filesystems

import (
	"context"
	"fmt"
	"io"
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
}

// New creates a new FileSystem manager from the given configuration.
func New(config ManagerConfig) (*FileSystem, error) {
	if len(config.Disks) == 0 {
		return nil, fmt.Errorf("no disks configured")
	}

	disks := make(map[string]Storage)

	for name, disk := range config.Disks {
		driverConfig, err := convertConfig(disk.Driver, disk.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create config for disk '%s': %w", name, err)
		}

		storageDriver, err := NewFactory().Create(disk.Driver, driverConfig)
		if err != nil {
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
		// This case can happen if there are no disks, and defaultDisk is ""
		if len(disks) == 0 {
			return nil, fmt.Errorf("no disks configured to select a default from")
		}
		return nil, fmt.Errorf("default disk '%s' not found in configured disks", defaultDisk)
	}

	return &FileSystem{
		disks:       disks,
		defaultDisk: defaultDisk,
	}, nil
}

// Disk returns a specific storage driver by name. Panics if not configured.
func (fs *FileSystem) Disk(name string) Storage {
	disk, ok := fs.disks[name]
	if !ok {
		panic(fmt.Sprintf("disk '%s' not configured", name))
	}
	return disk
}

// GetDisk returns the disk and an error if not found (non-panicking).
func (fs *FileSystem) GetDisk(name string) (Storage, error) {
	disk, ok := fs.disks[name]
	if !ok {
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
