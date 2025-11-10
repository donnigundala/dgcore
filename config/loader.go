package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// ------------------------- Loader (env + yaml) -------------------------

// Load searches for and loads configuration files from default paths.
// It looks for .env files and any .yaml/.yml files in ./ and ./config/
func Load() {
	// Default search paths for config files
	defaultPaths := []string{"./", "./config/"}
	LoadWithPaths(defaultPaths...)
}

// LoadWithPaths loads configuration from the specified directory paths.
// It loads a .env file, then merges all .yaml/.yml files found in the paths,
// and finally enables environment variable overrides.
func LoadWithPaths(paths ...string) {
	// 1. Load .env file. Errors are ignored if the file doesn't exist.
	_ = godotenv.Load()

	// 2. Set up Viper for environment variable overrides (highest priority after explicit Set)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viperInstance.AutomaticEnv()

	// 3. Load and merge all YAML/YML files from the specified paths.
	// These will be overridden by environment variables.
	mergedFiles := 0
	for _, path := range paths {
		files, err := os.ReadDir(path)
		if err != nil {
			// Silently ignore paths that don't exist
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			fileName := file.Name()
			if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
				// Set the config file to be merged
				viperInstance.SetConfigFile(filepath.Join(path, fileName))

				// Merge the config file. This adds its values without overwriting
				// values from previously merged files or environment variables
				// that have higher priority.
				if err := viperInstance.MergeInConfig(); err != nil {
					log.Printf("[CONFIG] Warning: could not merge config file %s: %v", fileName, err)
				} else {
					log.Printf("[CONFIG] Merged config from %s", viperInstance.ConfigFileUsed())
					mergedFiles++
				}
			}
		}
	}

	if mergedFiles == 0 {
		log.Printf("[CONFIG] No config files found, using defaults and environment variables.")
	}
}
