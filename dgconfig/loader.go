package dgconfig

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
)

// ------------------------- Loader (env + yaml) -------------------------

// Load reads config.yaml (if exists) and enables AutomaticEnv()
// You may call LoadWithPaths to specify search paths.
func Load() {
	LoadWithPaths([]string{"./", "./configs"})
}

func LoadWithPaths(paths []string) {
	_ = godotenv.Load()
	viperInstance.SetConfigName("config")
	viperInstance.SetConfigType("yaml")
	for _, p := range paths {
		viperInstance.AddConfigPath(p)
	}

	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viperInstance.AutomaticEnv()

	if err := viperInstance.ReadInConfig(); err == nil {
		log.Printf("[CONFIG] Loaded from %s", viperInstance.ConfigFileUsed())
	} else {
		log.Printf("[CONFIG] No config file found, using defaults and environment")
	}
}
