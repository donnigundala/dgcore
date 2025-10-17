// package config - production-grade config engine with Auto-Discovery & Auto-Inject
// Put this file in: pkg/config/extended_config.go

package dgconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	mu            sync.RWMutex
	viperInstance = viper.New()
	registry      = make(map[string]any)
	debugMode     = os.Getenv("CONFIG_DEBUG") == "true"
)

// ------------------------- Basic registry & helpers -------------------------

// Add registers a map of key->value under a prefix, e.g. Add("app", map[string]any{"name": "x"})
// It also sets viper default for each registered value and binds the env key.
func Add(prefix string, data map[string]any) {
	mu.Lock()
	defer mu.Unlock()

	flattenAndRegister(prefix, data)
}

func flattenAndRegister(prefix string, data map[string]any) {
	for k, v := range data {
		fullKey := prefix + "." + k
		switch val := v.(type) {
		case map[string]any:
			// recursive flatten of nested maps and register
			flattenAndRegister(fullKey, val)
		default:
			if debugMode {
				log.Printf("[CONFIG] Register %s = %v", fullKey, v)
			}
			registry[fullKey] = v

			// set default value if not already set in viper (from config file or env)
			if !viperInstance.IsSet(fullKey) {
				viperInstance.SetDefault(fullKey, v)
			}

			// Bind environment variable automatically: prefix_key in upper case
			envKey := toEnvKey(fullKey)
			_ = viperInstance.BindEnv(fullKey, envKey)
		}
	}
}

// Env helper reads from OS environment with fallback
func Env(key string, defaultVal any) any {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func Get(key string) any {
	mu.RLock()
	defer mu.RUnlock()

	// If exact key exists in viper
	//if viperInstance.IsSet(key) {
	//	return viperInstance.Get(key)
	//}

	// Special handling: if viper has a subtree for this key (config file nested map),
	// we must not return viper.Get(key) directly because that map may not include
	// environment overrides for nested keys. Instead, fall through to the merged
	// path below which combines registry defaults, YAML, and ENV.
	sub := viperInstance.Sub(key)
	if sub == nil {
		// No subtree in viper: safe to return scalar from viper or registry.
		if viperInstance.IsSet(key) {
			return viperInstance.Get(key)
		}
		if val, ok := registry[key]; ok {
			return val
		}
	}

	// If exact key exists in registry
	//if val, ok := registry[key]; ok {
	//	return val
	//}

	// If key is a prefix (e.g., app.nested), build a merged nested map from all sources
	prefix := key + "."
	hasNested := false
	m := map[string]any{}

	// Collect from registry defaults
	for k, v := range registry {
		if strings.HasPrefix(k, prefix) {
			short := k[len(prefix):]
			assignNested(m, short, v)
			hasNested = true
		}
	}

	// Collect from viper (YAML + ENV)
	for _, k := range viperInstance.AllKeys() {
		if strings.HasPrefix(k, prefix) {
			if viperInstance.IsSet(k) {
				short := k[len(prefix):]
				assignNested(m, short, viperInstance.Get(k))
				hasNested = true
			}
		}
	}

	// Ensure env-bound keys that may not appear in AllKeys are checked
	for regKey := range registry {
		if strings.HasPrefix(regKey, prefix) {
			if viperInstance.IsSet(regKey) {
				short := regKey[len(prefix):]
				assignNested(m, short, viperInstance.Get(regKey))
				hasNested = true
			}
		}
	}

	if hasNested {
		return m
	}

	// Fallback: not found
	return nil
}

func GetString(key string) string {
	if v := Get(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
		// fallback: try viper GetString
		return viperInstance.GetString(key)
	}
	return ""
}

func GetBool(key string) bool {
	if v := Get(key); v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
		return viperInstance.GetBool(key)
	}
	return false
}

// AllKeys returns a copy of registered keys
func AllKeys() []string {
	mu.RLock()
	defer mu.RUnlock()

	keys := make([]string, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}
	return keys
}

// PrintAll prints all registered keys and resolved values
func PrintAll() {
	log.Print("[CONFIG] ==== Registered Configs Value ====")
	for _, k := range AllKeys() {
		log.Printf("[CONFIG] %s = %v\n", k, Get(k))
	}
	log.Print("[CONFIG] ============================")
}

// ------------------------- Helpers -------------------------

// syncEnv binds environment variables for all keys under the given prefix.
// This ensures that any new keys added to the registry are also bound to their env vars.
func syncEnv(prefix string) {
	replacer := strings.NewReplacer(".", "_")
	for key := range registry {
		if strings.HasPrefix(key, prefix+".") {
			envKey := strings.ToUpper(replacer.Replace(key))
			_ = viperInstance.BindEnv(key, envKey)
		}
	}
}

func debugPrint(key string, val any, tag string) {
	if os.Getenv("CONFIG_DEBUG") == "true" {
		fmt.Printf("[CONFIG][%s] %s => %v\n", tag, key, val)
	}
}

func toEnvKey(fullKey string) string {
	// e.g. app.name -> APP_NAME
	k := fullKey
	k = replaceNonAlphaNumeric(k, '_')
	k = upperAndReplaceDots(k)
	return k
}

func upperAndReplaceDots(s string) string {
	return stringsToUpper(stringsReplace(s, ".", "_"))
}

func stringsToUpper(s string) string { return strings.ToUpper(s) }

func stringsReplace(s, old, new string) string { return strings.ReplaceAll(s, old, new) }

func replaceNonAlphaNumeric(s string, repl rune) string {
	// keep simple: only replace dots with underscore earlier; we won't attempt full sanitization.
	return s
}

func isTestFile(p string) bool { return filepath.Ext(p) == ".go" && stringsHasSuffix(p, "_test.go") }

func hasPrefix(s, pre string) bool { return len(s) >= len(pre) && s[:len(pre)] == pre }

// Minimal wrappers for strings functions to avoid heavy imports in this single-file example
func stringsHasSuffix(s, suf string) bool { return len(s) >= len(suf) && s[len(s)-len(suf):] == suf }

// ------------------------- End of package -------------------------
