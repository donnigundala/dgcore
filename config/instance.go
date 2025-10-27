package config

import (
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

// Config is an instance-based configuration manager.
type Config struct {
	mu       sync.RWMutex
	v        *viper.Viper
	registry map[string]any
}

// New creates a new config instance.
func New() *Config {
	return &Config{
		v:        viper.New(),
		registry: make(map[string]any),
	}
}

// Load loads YAML or env config files into this instance.
func (c *Config) Load(paths ...string) error {
	defaultPaths := []string{"./", "./configs"}
	if paths == nil {
		for _, path := range defaultPaths {
			paths = append(paths, path)
		}
	}

	LoadWithPaths(paths)
	return nil
}

// Add registers a config map under a prefix.
func (c *Config) Add(prefix string, data map[string]any) {
	Add(prefix, data)
}

// Env reads from environment variable or fallback.
func (c *Config) Env(key string, def any) any {
	return Env(key, def)
}

// Get retrieves a value by key.
func (c *Config) Get(key string) any {
	return Get(key)
}

// GetString retrieves a string value by key.
func (c *Config) GetString(key string) string {
	return GetString(key)
}

// AutoDiscover finds and prints config files in a directory.
func (c *Config) AutoDiscover(basePath string) error {
	files, err := filepath.Glob(filepath.Join(basePath, "*.go"))
	if err != nil {
		return err
	}
	for _, f := range files {
		debugPrint("DISCOVER", f, "FOUND (INSTANCE)")
	}
	return nil
}

// Inject unmarshals a config prefix into a struct.
func (c *Config) Inject(prefix string, out any) error {
	return Inject(prefix, out)
}
