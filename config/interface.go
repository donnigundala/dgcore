package config

type IConfig interface {
	Get(key string) interface{}
	// MustGet(name string) interface{}

	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetStringSlice(key string) []string
}

type IRegistrable interface {
	Defaults() map[string]interface{} // default values
	Prefix() string                   // e.g: "database"
}
