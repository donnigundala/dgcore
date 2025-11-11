package config

// Config holds the main application configuration.
type Config struct {
	Name     string `mapstructure:"name"`
	Env      string `mapstructure:"env"`
	Timezone string `mapstructure:"timezone"`
	Debug    string `mapstructure:"debug"`
	Secret   string `mapstructure:"secret"`
}
