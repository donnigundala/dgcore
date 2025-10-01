package config

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	viperInstance     *viper.Viper
	mu                sync.RWMutex
	registeredConfigs []interface{}
	subscribers       []func()
)

// Defaultable interface to provide default values for config struct
type Defaultable interface {
	Defaults() map[string]interface{}
}

// RegisterConfig add config struct so it can be un-marshaled & apply default values
// example: config.RegisterConfig(&config.AppConfig{})
func RegisterConfig(cfg interface{}) {
	mu.Lock()
	defer mu.Unlock()
	registeredConfigs = append(registeredConfigs, cfg)
}

// Viper instance getter
func Viper() *viper.Viper {
	return viperInstance
}

// Load read config file + .env + set defaults
func Load(configPath, configName, envPrefix string) error {
	// load .env first
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	// Support env
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// read main config file
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	viperInstance = v

	// apply defaults sebelum unmarshal
	applyDefaults()

	// unmarshal ke struct2 yang sudah register
	if err := unmarshalConfigs(); err != nil {
		return err
	}

	// setup hot reload
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("config file changed: %s", e.Name)
		if err := unmarshalConfigs(); err != nil {
			log.Printf("error reloading config: %v", err)
		}
		notifySubscribers()
	})

	return nil
}

// LoadFromViper accept custom viper instance (remote/merge/test)
func LoadFromViper(v *viper.Viper, envPrefix string) error {
	_ = godotenv.Load()

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	viperInstance = v

	// apply defaults sebelum unmarshal
	applyDefaults()

	if err := unmarshalConfigs(); err != nil {
		return err
	}

	// hot reload tetap jalan
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("config file changed: %s", e.Name)
		if err := unmarshalConfigs(); err != nil {
			log.Printf("error reloading config: %v", err)
		}
		notifySubscribers()
	})

	return nil
}

// applyDefaults apply default values from all registered configs
func applyDefaults() {
	mu.RLock()
	defer mu.RUnlock()
	for _, cfg := range registeredConfigs {
		if d, ok := cfg.(Defaultable); ok {
			for k, v := range d.Defaults() {
				viperInstance.SetDefault(k, v)
			}
		}
	}
}

// unmarshal reload everytime config changed
func unmarshalConfigs() error {
	mu.Lock()
	defer mu.Unlock()
	for _, cfg := range registeredConfigs {
		if err := viperInstance.Unmarshal(cfg); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}
	return nil
}

// Subscribe callback for config changes
func Subscribe(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	subscribers = append(subscribers, fn)
}

// notifySubscribers notify all subscribers
func notifySubscribers() {
	mu.RLock()
	defer mu.RUnlock()
	for _, fn := range subscribers {
		fn()
	}
}

// MustGet helper to get registered config pointer
func MustGet[T any](cfg T) T {
	return cfg
}

// Helper to load absolute path of .env (optional)
func loadEnvFile(path string) {
	if path == "" {
		_ = godotenv.Load()
		return
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		log.Printf("failed to resolve env file path: %v", err)
		return
	}
	_ = godotenv.Load(abs)
}
