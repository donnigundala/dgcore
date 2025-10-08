package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	viperInstance  *viper.Viper
	mu             sync.RWMutex
	configRegistry = make(map[string]interface{})
	lazyConfigs    []func() interface{}
	subscribers    []func()
)

var debugEnabled = false

// Defaultable defines config structs that can set default values.
type Defaultable interface {
	Defaults() map[string]interface{}
}

// RegisterConfig registers a config struct or a factory (lazy load).
// Example:
//
//	config.RegisterConfig(&AppConfig{})          // direct
//	config.RegisterConfig(func() interface{} {   // lazy
//	    return &MongoConfig{}
//	})
func RegisterConfig(factory interface{}) {
	mu.Lock()
	defer mu.Unlock()

	switch f := factory.(type) {
	case func() interface{}:
		lazyConfigs = append(lazyConfigs, f)
		log.Printf("Registering lazy config: %T", f())
	default:
		key := reflect.TypeOf(factory).String()
		log.Printf("Registering config: %s", key)
		configRegistry[key] = factory
	}
}

// Viper returns the active viper instance.
func Viper() *viper.Viper {
	return viperInstance
}

// Load loads config from file, .env, and applies defaults.
func Load(configPaths []string, configName string, envPrefix string) error {
	// Load .env first
	_ = godotenv.Load()

	// Enable debug mode if DEBUG_CONFIG=true
	if strings.ToLower(os.Getenv("DEBUG_CONFIG")) == "true" {
		debugEnabled = true
		log.Println("[CONFIG] Debug mode enabled")
	}

	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType("yaml")

	// Add all possible paths
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file if exists
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	viperInstance = v

	// Apply defaults
	applyDefaults()

	// Process lazy configs (instantiate after viper ready)
	processLazyConfigs()

	// Unmarshal to registered configs
	if err := unmarshalConfigs(); err != nil {
		return err
	}

	// Setup hot reload
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

func applyDefaults() {
	mu.RLock()
	defer mu.RUnlock()

	for _, cfg := range configRegistry {
		if d, ok := cfg.(Defaultable); ok {
			for k, v := range d.Defaults() {
				if !viperInstance.IsSet(k) {
					viperInstance.SetDefault(k, v)
					debugPrint(k, v, "DEFAULT")
				}
				if err := viperInstance.BindEnv(k); err != nil {
					log.Printf("warning: failed to bind env for key %s: %v", k, err)
				}
			}
		}
	}
}

func processLazyConfigs() {
	mu.Lock()
	defer mu.Unlock()

	for _, factory := range lazyConfigs {
		cfg := factory()
		key := reflect.TypeOf(cfg).String()
		log.Printf("[CONFIG] Processing lazy config: %s", key)
		configRegistry[key] = cfg
	}
}

// unmarshalConfigs loads all registered configs into structs
func unmarshalConfigs() error {
	mu.Lock()
	defer mu.Unlock()

	for key, cfg := range configRegistry {
		t := fmt.Sprintf("%T", cfg)
		parts := strings.Split(t, ".")
		typeName := strings.TrimSuffix(parts[len(parts)-1], "Config")
		prefix := strings.ToLower(typeName)

		allSettings := viperInstance.AllSettings()
		if _, ok := allSettings[prefix]; ok {
			if err := viperInstance.UnmarshalKey(prefix, cfg); err != nil {
				return fmt.Errorf("failed to unmarshal section %s: %w", prefix, err)
			}
			//continue
		}

		val := reflect.ValueOf(cfg).Elem()
		typ := val.Type()

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			tag := field.Tag.Get("mapstructure")
			if tag == "" {
				tag = strings.ToLower(field.Name)
			}

			fullKey := fmt.Sprintf("%s.%s", prefix, tag)
			envKey := strings.ToUpper(strings.ReplaceAll(fullKey, ".", "_"))

			var finalVal interface{}
			source := "DEFAULT-FALLBACK"

			if viperInstance.IsSet(fullKey) {
				finalVal = viperInstance.Get(fullKey)
				if os.Getenv(envKey) != "" {
					source = "ENV"
				} else {
					source = "YAML"
				}
			} else if d, ok := cfg.(Defaultable); ok {
				defaults := d.Defaults()
				if defVal, ok := defaults[fullKey]; ok {
					finalVal = defVal
				}
			}

			if finalVal == nil {
				debugPrint(fullKey, "<nil>", "EMPTY")
				continue
			}

			debugPrint(fullKey, finalVal, source)

			fieldVal := val.Field(i)
			switch fieldVal.Kind() {
			case reflect.Slice:
				elemType := fieldVal.Type().Elem() // tipe elemen slice

				switch v := finalVal.(type) {
				case string:
					parts := strings.Split(v, ",")
					slice := reflect.MakeSlice(fieldVal.Type(), len(parts), len(parts))
					for j, p := range parts {
						elem := reflect.ValueOf(strings.TrimSpace(p))
						if !elem.Type().AssignableTo(elemType) {
							// fallback ke string
							elem = reflect.ValueOf(fmt.Sprintf("%v", p))
						}
						slice.Index(j).Set(elem)
					}
					fieldVal.Set(slice)
				case []string:
					fieldVal.Set(reflect.ValueOf(v))
				case []interface{}:
					slice := reflect.MakeSlice(fieldVal.Type(), len(v), len(v))
					for j, item := range v {
						elem := reflect.ValueOf(fmt.Sprintf("%v", item))
						slice.Index(j).Set(elem)
					}
					fieldVal.Set(slice)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch v := finalVal.(type) {
				case int:
					fieldVal.SetInt(int64(v))
				case float64:
					fieldVal.SetInt(int64(v))
				case string:
					if iv, err := strconv.ParseInt(v, 10, 64); err == nil {
						fieldVal.SetInt(iv)
					}
				case fmt.Stringer:
					if iv, err := strconv.ParseInt(v.String(), 10, 64); err == nil {
						fieldVal.SetInt(iv)
					}
				}

			case reflect.Bool:
				switch v := finalVal.(type) {
				case bool:
					fieldVal.SetBool(v)
				case string:
					if bv, err := strconv.ParseBool(v); err == nil {
						fieldVal.SetBool(bv)
					}
				}

			case reflect.Float32, reflect.Float64:
				switch v := finalVal.(type) {
				case float64:
					fieldVal.SetFloat(v)
				case string:
					if fv, err := strconv.ParseFloat(v, 64); err == nil {
						fieldVal.SetFloat(fv)
					}
				}

			default:
				v := reflect.ValueOf(finalVal)
				if v.Type().AssignableTo(fieldVal.Type()) {
					fieldVal.Set(v)
				} else {
					fieldVal.Set(reflect.ValueOf(fmt.Sprintf("%v", finalVal)))
				}
			}
		}

		configRegistry[key] = cfg
	}
	return nil
}

// Subscribe allows functions to react on config reload.
func Subscribe(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	subscribers = append(subscribers, fn)
}

func notifySubscribers() {
	mu.RLock()
	defer mu.RUnlock()
	for _, fn := range subscribers {
		fn()
	}
}

func MustGet[T any](cfg T) T {
	mu.RLock()
	defer mu.RUnlock()

	key := fmt.Sprintf("%T", cfg)
	if cfg, ok := configRegistry[key]; ok {
		return cfg.(T)
	}
	panic(fmt.Sprintf("config %s not found", key))
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

func debugPrint(key string, val interface{}, source string) {
	if debugEnabled {
		log.Printf("[CONFIG] %-30s = %-30v (%s)", key, val, source)
	}
}
