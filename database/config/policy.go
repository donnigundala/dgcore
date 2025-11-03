package config

import "time"

type HealthPolicy struct {
	PingInterval         time.Duration `json:"ping_interval" mapstructure:"ping_interval"`
	MaxFailures          int           `json:"max_failures" mapstructure:"max_failures"`
	ReconnectBackoffBase time.Duration `json:"reconnect_backoff_base" mapstructure:"reconnect_backoff_base"`
	ReconnectBackoffMax  time.Duration `json:"reconnect_backoff_max" mapstructure:"reconnect_backoff_max"`
	FallbackDuration     time.Duration `json:"fallback_duration" mapstructure:"fallback_duration"`
	UseHybridFailover    bool          `json:"use_hybrid_failover" mapstructure:"use_hybrid_failover"`
}

func DefaultPolicy() *HealthPolicy {
	return &HealthPolicy{
		PingInterval:         10 * time.Second,
		MaxFailures:          3,
		ReconnectBackoffBase: 2 * time.Second,
		ReconnectBackoffMax:  30 * time.Second,
		FallbackDuration:     1 * time.Minute,
		UseHybridFailover:    true,
	}
}
