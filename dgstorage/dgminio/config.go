package dgminio

type Config struct {
	Endpoint  string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	AccessKey string `mapstructure:"access_key" json:"access_key" yaml:"access_key"`
	SecretKey string `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`
	Secure    bool   `mapstructure:"secure" json:"secure" yaml:"secure"`
}

func defaultConfig() *Config {
	return &Config{
		Endpoint:  "localhost:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		Secure:    false,
	}
}
