package filesystem

import (
	"context"
	"fmt"
	"io"
	"time"
)

// FileSystem is the top-level manager for all storage disks.
type FileSystem struct {
	disks       map[string]Storage
	defaultDisk string
}

// ManagerConfig holds the configuration for the filesystem manager.
type ManagerConfig struct {
	Default string
	Disks   map[string]Disk
}

// Disk represents the configuration for a single storage driver.
type Disk struct {
	Driver string
	Config map[string]interface{}
}

// New creates a new FileSystem manager from the given configuration.
func New(config ManagerConfig) (*FileSystem, error) {
	factory := NewFactory()
	disks := make(map[string]Storage)

	for name, disk := range config.Disks {
		driverConfig, err := convertConfig(disk.Driver, disk.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create config for disk '%s': %w", name, err)
		}

		storageDriver, err := factory.Create(disk.Driver, driverConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create driver for disk '%s': %w", name, err)
		}
		disks[name] = storageDriver
	}

	if _, ok := disks[config.Default]; !ok && len(disks) > 0 {
		return nil, fmt.Errorf("default disk '%s' not found in configured disks", config.Default)
	}

	return &FileSystem{
		disks:       disks,
		defaultDisk: config.Default,
	}, nil
}

// Disk returns a specific storage driver by name.
// It panics if the disk is not configured, as this is considered a programming error.
func (fs *FileSystem) Disk(name string) Storage {
	disk, ok := fs.disks[name]
	if !ok {
		panic(fmt.Sprintf("disk '%s' not configured", name))
	}
	return disk
}

// --- Default Disk Methods ---

// Upload uses the default disk to store data.
func (fs *FileSystem) Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error {
	return fs.Disk(fs.defaultDisk).Upload(ctx, key, data, size, visibility)
}

// Download uses the default disk to retrieve data.
func (fs *FileSystem) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return fs.Disk(fs.defaultDisk).Download(ctx, key)
}

// Delete uses the default disk to remove an object.
func (fs *FileSystem) Delete(ctx context.Context, key string) error {
	return fs.Disk(fs.defaultDisk).Delete(ctx, key)
}

// Exists uses the default disk to check if an object exists.
func (fs *FileSystem) Exists(ctx context.Context, key string) (bool, error) {
	return fs.Disk(fs.defaultDisk).Exists(ctx, key)
}

// GetURL uses the default disk to generate a URL.
func (fs *FileSystem) GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error) {
	return fs.Disk(fs.defaultDisk).GetURL(ctx, key, visibility, duration)
}

// List uses the default disk to return all keys with a given prefix.
func (fs *FileSystem) List(ctx context.Context, prefix string) ([]string, error) {
	return fs.Disk(fs.defaultDisk).List(ctx, prefix)
}

// convertConfig converts a map[string]interface{} to the specific config struct needed by the factory.
func convertConfig(driver string, config map[string]interface{}) (interface{}, error) {
	switch driver {
	case "local":
		cfg := LocalConfig{}
		if v, ok := config["basePath"].(string); ok {
			cfg.BasePath = v
		}
		if v, ok := config["baseURL"].(string); ok {
			cfg.BaseURL = v
		}
		if v, ok := config["secret"].(string); ok {
			cfg.Secret = v
		}
		return cfg, nil
	case "s3":
		cfg := S3ConfigWithAuth{}
		if v, ok := config["bucket"].(string); ok {
			cfg.Bucket = v
		}
		if v, ok := config["region"].(string); ok {
			cfg.Region = v
		}
		if v, ok := config["accessKey"].(string); ok {
			cfg.AccessKey = v
		}
		if v, ok := config["secretKey"].(string); ok {
			cfg.SecretKey = v
		}
		return cfg, nil
	case "minio":
		cfg := MinIOConfig{}
		if v, ok := config["endpoint"].(string); ok {
			cfg.Endpoint = v
		}
		if v, ok := config["accessKeyID"].(string); ok {
			cfg.AccessKeyID = v
		}
		if v, ok := config["secretAccessKey"].(string); ok {
			cfg.SecretAccessKey = v
		}
		if v, ok := config["useSSL"].(bool); ok {
			cfg.UseSSL = v
		}
		if v, ok := config["bucket"].(string); ok {
			cfg.Bucket = v
		}
		if v, ok := config["baseURL"].(string); ok {
			cfg.BaseURL = v
		}
		return cfg, nil
	default:
		return nil, fmt.Errorf("unsupported driver for config conversion: %s", driver)
	}
}
