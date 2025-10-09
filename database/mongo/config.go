package mongo

import (
	"time"
)

// Config holds MongoDB configuration.
// It supports multiple hosts for replica sets and various connection options.
type Config struct {
	// Cluster host list
	Hosts []string `mapstructure:"hosts" json:"hosts" yaml:"hosts"`

	// Basic auth
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	Database string `mapstructure:"database" json:"database" yaml:"database"`

	// Options
	AuthSource string `mapstructure:"auth_source" json:"auth_source" yaml:"auth_source"`
	ReplicaSet string `mapstructure:"replica_set" json:"replica_set" yaml:"replica_set"`
	UseSRV     bool   `mapstructure:"use_srv" json:"use_srv" yaml:"use_srv"`
	URI        string `mapstructure:"uri" json:"uri" yaml:"uri"` // Optional URI override (e.g., mongodb+srv://...)

	// Pool & timeout
	MaxPoolSize        int           `mapstructure:"max_pool_size" json:"max_pool_size" yaml:"max_pool_size"`
	MinPoolSize        int           `mapstructure:"min_pool_size" json:"min_pool_size" yaml:"min_pool_size"`
	ConnectTimeout     time.Duration `mapstructure:"connect_timeout" json:"connect_timeout" yaml:"connect_timeout"`
	ServerSelectionTTL time.Duration `mapstructure:"server_selection_ttl" json:"server_selection_ttl" yaml:"server_selection_ttl"`

	// Monitoring
	EnableMonitor bool   `mapstructure:"enable_monitor" json:"enable_monitor" yaml:"enable_monitor"`
	Debug         bool   `mapstructure:"debug" json:"debug" yaml:"debug"`
	Driver        string `mapstructure:"driver" json:"driver" yaml:"driver"`
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
