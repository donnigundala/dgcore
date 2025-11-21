package firebase

import "github.com/donnigundala/dgcore/config"

// Config struct for Firebase
type Config struct {
	// CredentialsFile path to the JSON credentials file.
	CredentialsFile string `mapstructure:"credentials_file" json:"credentials_file" yaml:"credentials_file"`
	// CredentialsJSON json string of the credentials.
	CredentialsJSON string `mapstructure:"credentials_json" json:"credentials_json" yaml:"credentials_json"`
}

// init registers default configuration values for a 'default' firebase app.
func init() {
	// The user is expected to override these with actual paths or JSON content,
	// typically via environment variables for security.
	config.Add("firebase", map[string]any{
		"default": map[string]any{
			"credentials_file": "",
			"credentials_json": "",
		},
	})
}
