package database

import (
	"context"
)

// ProviderType defines the type of database provider.
type ProviderType string

const (
	// ProviderSQL represents the SQL database provider.
	ProviderSQL ProviderType = "sql"
	// ProviderMongo represents the MongoDB database provider.
	ProviderMongo ProviderType = "mongo"
	// ProviderRedis represents the Redis database provider (for future use).
	ProviderRedis ProviderType = "redis"
)

// MetricsProvider defines the interface for a metrics provider.
// This allows for integration with any metrics library (e.g., Prometheus, StatsD).
type MetricsProvider interface {
	// Inc increments a counter metric.
	Inc(name string, labels ...string)
	// SetGauge sets the value of a gauge metric.
	SetGauge(name string, value float64, labels ...string)
	// Observe records a value in a histogram or summary metric.
	Observe(name string, value float64, labels ...string)
}

// nopMetricsProvider is a no-op implementation of MetricsProvider.
type nopMetricsProvider struct{}

func (n *nopMetricsProvider) Inc(name string, labels ...string)                     {}
func (n *nopMetricsProvider) SetGauge(name string, value float64, labels ...string) {}
func (n *nopMetricsProvider) Observe(name string, value float64, labels ...string)  {}

// NewNopMetricsProvider returns a metrics provider that does nothing.
func NewNopMetricsProvider() MetricsProvider {
	return &nopMetricsProvider{}
}

// Provider is the interface that all database providers must implement.
type Provider interface {
	Ping(ctx context.Context) error
	Close() error
	// GetDB is deprecated. Use GetWriter() instead.
	GetDB() any
	GetWriter() any
	GetReader() any
}

// Config holds the configuration for all database providers.
type Config struct {
	Driver     ProviderType    `json:"driver" mapstructure:"driver"`
	SQL        *SQLConfig      `json:"sql" mapstructure:"sql"`
	Mongo      *MongoConfig    `json:"mongo" mapstructure:"mongo"`
	Policy     *HealthPolicy   `json:"policy" mapstructure:"policy"`
	Metrics    MetricsProvider `json:"-" mapstructure:"-"` // Metrics provider is not configurable via file
	TraceIDKey string          `json:"trace_id_key" mapstructure:"trace_id_key"`
}

// ManagerConfig is used to inject the full database configuration map.
type ManagerConfig map[string]*Config
