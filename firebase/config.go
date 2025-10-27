package firebase

// Config struct for Firebase
type Config struct {
	// CredentialsFile path to the JSON credentials file.
	CredentialsFile string `mapstructure:"credentials_file" json:"credentials_file" yaml:"credentials_file"`
	// CredentialsJSON json string of the credentials.
	CredentialJSON string `mapstructure:"credential_json" json:"credential_json" yaml:"credential_json"`
}
