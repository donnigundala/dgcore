package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// ------------------------- Loader (env + yaml) -------------------------

// Load searches for and loads configuration files from default paths.
// It looks for .env files and any .yaml/.yml files in ./ and ./config/
// It returns an error if any config file is found but fails to parse.
func Load() error {
	// Default search paths for config files
	defaultPaths := []string{"./", "./config/"}
	return LoadWithPaths(defaultPaths...)
}

// LoadWithPaths loads configuration from the specified directory paths.
// It loads a .env file, then merges all .yaml/.yml files found in the paths,
// and finally enables environment variable overrides.
// It returns an error if any config file is found but fails to parse.
func LoadWithPaths(paths ...string) error {
	// 1. Load .env file. Errors are ignored if the file doesn't exist, which is standard.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		slog.Warn("Failed to load .env file", "error", err)
	}

	// 2. Set up Viper for environment variable overrides (highest priority).
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viperInstance.AutomaticEnv()

	// 3. Load and merge all YAML/YML files from the specified paths.
	mergedFiles := 0
	for _, path := range paths {
		files, err := os.ReadDir(path)
		if err != nil {
			// Silently ignore paths that don't exist. This is expected behavior.
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			fileName := file.Name()
			if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
				fullPath := filepath.Join(path, fileName)
				viperInstance.SetConfigFile(fullPath)

				// Merge the config file.
				if err := viperInstance.MergeInConfig(); err != nil {
					// This is a critical error. If a config file is present but malformed,
					// the application should fail fast.
					return fmt.Errorf("failed to merge config file %s: %w", fullPath, err)
				}
				slog.Info("Merged config file", "path", viperInstance.ConfigFileUsed())
				mergedFiles++
			}
		}
	}

	if mergedFiles == 0 {
		slog.Info("No config files found. Using defaults and environment variables only.")
	}

	return nil
}
