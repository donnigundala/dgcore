package firebase

type Config struct {
	CredentialsFile string `mapstructure:"credentials_file" json:"credentials_file" yaml:"credentials_file"`
}

func defaultConfig() *Config {
	return &Config{
		CredentialsFile: "serviceAccountKey.json",
	}
}
