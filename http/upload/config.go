package upload

// Config holds upload configuration.
type Config struct {
	MaxSize      int64    // Max file size in bytes (default: 10MB)
	AllowedTypes []string // Allowed MIME types (empty = all allowed)
	MaxFiles     int      // Max number of files (default: 1)
	FieldName    string   // Form field name (default: "file")
}

// DefaultConfig returns the default upload configuration.
func DefaultConfig() Config {
	return Config{
		MaxSize:      10 * 1024 * 1024, // 10MB
		AllowedTypes: []string{},       // All types allowed
		MaxFiles:     1,
		FieldName:    "file",
	}
}

// WithMaxSize sets the maximum file size.
func (c Config) WithMaxSize(size int64) Config {
	c.MaxSize = size
	return c
}

// WithAllowedTypes sets the allowed MIME types.
func (c Config) WithAllowedTypes(types []string) Config {
	c.AllowedTypes = types
	return c
}

// WithMaxFiles sets the maximum number of files.
func (c Config) WithMaxFiles(max int) Config {
	c.MaxFiles = max
	return c
}

// WithFieldName sets the form field name.
func (c Config) WithFieldName(name string) Config {
	c.FieldName = name
	return c
}
