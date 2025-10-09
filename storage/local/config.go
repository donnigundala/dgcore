package local

type Config struct {
	BasePath string `mapstructure:"base_path" json:"base_path" yaml:"base_path"`
	BaseURL  string `mapstructure:"base_url" json:"base_url" yaml:"base_url"`
}

func defaultConfig() *Config {
	return &Config{
		BasePath: "./uploads",
		BaseURL:  "http://localhost:8080/uploads",
	}
}
