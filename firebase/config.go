package firebase

type Config struct {
	CredentialsFile string `mapstructure:"credentials_file" json:"credentials_file" yaml:"credentials_file"`
}

func DefaultConfig() *Config {
	return &Config{
		CredentialsFile: "serviceAccountKey.json",
	}
}

func SetViperDefaultConfig(prefix *string) map[string]interface{} {
	cfg := DefaultConfig()
	// Default prefix
	p := "firebase"
	if prefix != nil && *prefix != "" {
		p = *prefix
	}
	p = p + "."

	return map[string]interface{}{
		p + "credentials_file": cfg.CredentialsFile,
	}
}
