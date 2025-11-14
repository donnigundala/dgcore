package filesystems

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"time"
)

// Manager is the top-level manager for all storage disks.
type Manager struct {
	disks       map[string]Storage
	defaultDisk string
	logger      *slog.Logger
}

// ManagerOption configures a Manager.
type ManagerOption func(*Manager)

// WithLogger provides a slog logger for the filesystem manager.
// If not provided, logs will be discarded.
func WithLogger(logger *slog.Logger) ManagerOption {
	return func(m *Manager) {
		if logger != nil {
			m.logger = logger
		}
	}
}

// NewManager creates a new Manager from the given configuration.
func NewManager(config Config, opts ...ManagerOption) (*Manager, error) {
	m := &Manager{
		// Default to a silent logger. This can be overridden by the WithLogger option.
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	for _, opt := range opts {
		opt(m)
	}

	if len(config.Disks) == 0 {
		return nil, fmt.Errorf("no disks configured")
	}

	m.logger.Debug("Initializing filesystem manager", "default_disk", config.Default, "disk_count", len(config.Disks))

	disks := make(map[string]Storage)

	for name, disk := range config.Disks {
		// The traceIDKey is no longer passed down.
		storageDriver, err := newStorage(disk.Driver, disk.Config, m.logger.With("disk", name, "driver", disk.Driver))
		if err != nil {
			m.logger.Error("Failed to create driver for disk", "disk", name, "driver", disk.Driver, "error", err)
			return nil, fmt.Errorf("failed to create driver for disk '%s': %w", name, err)
		}
		disks[name] = storageDriver
	}

	// If default not provided, choose one deterministically.
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

	m.disks = disks
	m.defaultDisk = defaultDisk

	m.logger.Info("Filesystem manager initialized", "default_disk", m.defaultDisk)

	return m, nil
}

// Disk returns a specific storage driver by name. Panics if not configured.
func (m *Manager) Disk(name string) Storage {
	disk, ok := m.disks[name]
	if !ok {
		m.logger.Error("Requested disk not configured", "disk", name)
		panic(fmt.Sprintf("disk '%s' not configured", name))
	}
	return disk
}

// GetDisk returns the disk and an error if not found (non-panicking).
func (m *Manager) GetDisk(name string) (Storage, error) {
	disk, ok := m.disks[name]
	if !ok {
		m.logger.Warn("Requested disk not configured", "disk", name)
		return nil, fmt.Errorf("disk '%s' not configured", name)
	}
	return disk, nil
}

// Default-disk convenience methods. These methods already accept a context,
// which will be passed down to the underlying storage driver.

func (m *Manager) Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error {
	return m.Disk(m.defaultDisk).Upload(ctx, key, data, size, visibility)
}

func (m *Manager) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return m.Disk(m.defaultDisk).Download(ctx, key)
}

func (m *Manager) Delete(ctx context.Context, key string) error {
	return m.Disk(m.defaultDisk).Delete(ctx, key)
}

func (m *Manager) Exists(ctx context.Context, key string) (bool, error) {
	return m.Disk(m.defaultDisk).Exists(ctx, key)
}

func (m *Manager) GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error) {
	return m.Disk(m.defaultDisk).GetURL(ctx, key, visibility, duration)
}

func (m *Manager) List(ctx context.Context, prefix string) ([]string, error) {
	return m.Disk(m.defaultDisk).List(ctx, prefix)
}
