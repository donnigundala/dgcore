package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Config is an instance-based configuration manager, providing an isolated config environment.
// This is useful for testing or for applications that need to manage multiple, separate configurations.
type Config struct {
	mu sync.RWMutex
	v  *viper.Viper
}

// New creates a new, isolated config instance.
func New() *Config {
	v := viper.New()
	// Configure viper instance for consistency with the global one.
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return &Config{
		v: v,
	}
}

// Load loads YAML configuration files from the specified paths into this instance.
// If no paths are provided, it uses default paths ("./", "./config/").
// Environment variables will override values from the loaded files.
func (c *Config) Load(paths ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(paths) == 0 {
		paths = []string{"./", "./config/"}
	}

	mergedFiles := 0
	for _, path := range paths {
		files, err := os.ReadDir(path)
		if err != nil {
			// Silently ignore paths that don't exist.
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			fileName := file.Name()
			if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
				fullPath := filepath.Join(path, fileName)
				c.v.SetConfigFile(fullPath)

				if err := c.v.MergeInConfig(); err != nil {
					return fmt.Errorf("instance failed to merge config file %s: %w", fullPath, err)
				}
				slog.Debug("Instance merged config file", "path", c.v.ConfigFileUsed())
				mergedFiles++
			}
		}
	}

	if mergedFiles == 0 {
		slog.Debug("Instance found no config files to load.")
	}

	return nil
}

// Get retrieves a value by key from this specific config instance.
func (c *Config) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.Get(key)
}

// GetString retrieves a string value by key from this specific config instance.
func (c *Config) GetString(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.GetString(key)
}

// GetInt retrieves an int value by key from this specific config instance.
func (c *Config) GetInt(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.GetInt(key)
}

// GetBool retrieves a boolean value by key from this specific config instance.
func (c *Config) GetBool(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.GetBool(key)
}

// IsSet checks if a key is set in this specific config instance.
func (c *Config) IsSet(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.IsSet(key)
}

// Inject unmarshals a config prefix from this instance into a struct.
// It first checks if the key exists before attempting to unmarshal.
func (c *Config) Inject(prefix string, out any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.v.IsSet(prefix) {
		return fmt.Errorf("prefix '%s' not found in config instance", prefix)
	}

	if err := c.v.UnmarshalKey(prefix, out); err != nil {
		return fmt.Errorf("failed to inject config for prefix '%s': %w", prefix, err)
	}
	return nil
}
