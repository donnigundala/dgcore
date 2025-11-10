package database

import (
	"os"
)

// Secret is a string value that can be provided directly or loaded from an environment variable.
type Secret struct {
	Value   string `json:"value,omitempty" mapstructure:"value"`
	FromEnv string `json:"from_env,omitempty" mapstructure:"from_env"`
}

// Get resolves the secret, preferring the environment variable if specified.
func (s *Secret) Get() string {
	if s.FromEnv != "" {
		return os.Getenv(s.FromEnv)
	}
	return s.Value
}
