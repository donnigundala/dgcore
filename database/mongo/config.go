package mongo

import (
	"time"
)

// Config holds MongoDB configuration.
// It supports multiple hosts for replica sets and various connection options.
type Config struct {
	// Cluster host list
	Hosts []string

	// Basic auth
	Username string
	Password string
	Database string

	// Options
	AuthSource string
	ReplicaSet string
	UseSRV     bool
	URI        string

	// Pool & timeout
	MaxPoolSize        int
	MinPoolSize        int
	ConnectTimeout     time.Duration
	ServerSelectionTTL time.Duration

	// Monitoring
	EnableMonitor bool
	Debug         bool
	Driver        string
}

//func NewConfig(appCfg *newconfig.MongoConfig) *Config {
//	if appCfg == nil {
//		return &Config{}
//	}
//
//	return &Config{
//		// Cluster host list
//		Hosts: appCfg.Hosts,
//
//		// basic auth
//		Username: appCfg.Username,
//		Password: appCfg.Password,
//		Database: appCfg.Database,
//
//		// options
//		AuthSource: appCfg.AuthSource,
//		ReplicaSet: appCfg.ReplicaSet,
//		UseSRV:     appCfg.UseSRV,
//		URI:        appCfg.URI, // Optional URI override (e.g., mongodb+srv://...)
//
//		// pool & timeout
//		MaxPoolSize:        appCfg.MaxPoolSize,
//		MinPoolSize:        appCfg.MinPoolSize,
//		ConnectTimeout:     appCfg.ConnectTimeout,
//		ServerSelectionTTL: appCfg.ServerSelectionTTL,
//
//		// monitoring
//		EnableMonitor: appCfg.EnableMonitor,
//		Debug:         appCfg.Debug,
//		Driver:        "mongo",
//	}
//}
