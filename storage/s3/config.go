package s3

type Config struct {
	Region          string `mapstructure:"region" json:"region" yaml:"region"`
	Bucket          string `mapstructure:"bucket" json:"bucket" yaml:"bucket"`
	AccessKeyID     string `mapstructure:"access_key" json:"access_key" yaml:"access_key"`
	SecretAccessKey string `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`
	BaseURL         string `mapstructure:"base_url" json:"base_url" yaml:"base_url"`
}

func defaultConfig() *Config {
	return &Config{
		Region:          "us-east-1",
		Bucket:          "",
		AccessKeyID:     "",
		SecretAccessKey: "",
		BaseURL:         "",
	}
}
