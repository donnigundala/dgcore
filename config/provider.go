package config

import "github.com/spf13/viper"

type provider struct {
	v *viper.Viper
}

// GetViper returns the underlying Viper instance.
func (p *provider) GetViper() *viper.Viper {
	return p.v
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Viper will check in the following order:
// override, flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func (p *provider) Get(key string) interface{} {
	return p.v.Get(key)
}

// MustGet retrieves a configuration value by key and panics if the key is not set.
//func (p *provider) MustGet(key string) interface{} {
//	if !p.v.IsSet(key) {
//		panic("missing config key: " + key)
//	}
//	return p.v.Get(key)
//}

// GetString returns the value associated with the key as a string.
func (p *provider) GetString(key string) string {
	return p.v.GetString(key)
}

// GetInt returns the value associated with the key as an integer.
func (p *provider) GetInt(key string) int {
	return p.v.GetInt(key)
}

// GetBool returns the value associated with the key as a boolean.
func (p *provider) GetBool(key string) bool {
	return p.v.GetBool(key)
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (p *provider) GetStringSlice(key string) []string {
	return p.v.GetStringSlice(key)
}
