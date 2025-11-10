package database

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// poolMetricsMonitor tracks connection pool statistics for a single mongo client.
type poolMetricsMonitor struct {
	openConnections int64
	slog            *slog.Logger
}

func (m *poolMetricsMonitor) Event(evt *event.PoolEvent) {
	switch evt.Type {
	case event.ConnectionCreated:
		atomic.AddInt64(&m.openConnections, 1)
		m.slog.Debug("Connection created", "connectionID", evt.ConnectionID, "pool", evt.PoolOptions)
	case event.ConnectionClosed:
		atomic.AddInt64(&m.openConnections, -1)
		m.slog.Debug("Connection closed", "connectionID", evt.ConnectionID, "reason", evt.Reason)
	}
}

func (m *poolMetricsMonitor) GetOpenConnections() int64 {
	return atomic.LoadInt64(&m.openConnections)
}

type MongoProvider struct {
	cfg        *MongoConfig
	policy     *HealthPolicy
	metrics    MetricsProvider
	baseLogger *slog.Logger // Base logger for the provider
	traceIDKey string       // Key to extract trace ID from context
	primary    *mongo.Client
	replicas   []*mongo.Client
	monitors   map[string]*poolMetricsMonitor
	mu         sync.RWMutex

	isPrimaryHealthy bool
	lastReplicaIndex int
	failureCount     int
	reconnectBackoff time.Duration
}

func NewMongoProvider(ctx context.Context, cfg *MongoConfig, policy *HealthPolicy, metrics MetricsProvider, baseLogger *slog.Logger, traceIDKey string) (Provider, error) {
	if cfg == nil {
		return nil, errors.New("mongo config cannot be nil")
	}
	if policy == nil {
		policy = DefaultPolicy()
	}
	if baseLogger == nil {
		baseLogger = slog.Default()
	}

	p := &MongoProvider{
		cfg:              cfg,
		policy:           policy,
		metrics:          metrics,
		baseLogger:       baseLogger.With("component", "mongo_provider"),
		traceIDKey:       traceIDKey,
		monitors:         make(map[string]*poolMetricsMonitor),
		lastReplicaIndex: -1,
		reconnectBackoff: policy.ReconnectBackoffBase,
	}

	if err := p.connectAll(ctx); err != nil {
		return nil, err
	}

	go p.startHealthMonitor(ctx)
	go p.startMetricsCollector(ctx)
	return p, nil
}

// logger returns a context-aware slog.Logger.
func (p *MongoProvider) logger(ctx context.Context) *slog.Logger {
	if p.traceIDKey != "" {
		if traceID, ok := ctx.Value(p.traceIDKey).(string); ok && traceID != "" {
			return p.baseLogger.With("trace_id", traceID)
		}
	}
	return p.baseLogger
}

func (p *MongoProvider) connectAll(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Connect to Primary
	primaryClient, err := p.connect(ctx, "primary", p.cfg.PrimaryURI, readpref.Primary())
	if err != nil {
		p.logger(ctx).Error("Failed to connect to primary mongo. Entering failover mode.", "error", err)
		p.metrics.Inc("db_connection_errors", "primary")
	} else {
		p.primary = primaryClient
		p.isPrimaryHealthy = true
		p.metrics.SetGauge("db_primary_healthy", 1, "primary")
		p.logger(ctx).Info("Primary connected")
	}

	// Connect to Replicas
	for i, uri := range p.cfg.Replicas {
		name := fmt.Sprintf("replica-%d", i)
		replicaClient, err := p.connect(ctx, name, uri, readpref.SecondaryPreferred())
		if err != nil {
			p.logger(ctx).Error("Failed to connect to replica", "name", name, "error", err)
			p.metrics.Inc("db_connection_errors", name)
		} else {
			p.replicas = append(p.replicas, replicaClient)
			p.logger(ctx).Info("Replica connected", "name", name)
		}
	}

	if p.primary == nil && len(p.replicas) == 0 {
		return errors.New("failed to connect to any mongo database")
	}

	return nil
}

func (p *MongoProvider) connect(ctx context.Context, name string, uri Secret, rp *readpref.ReadPref) (*mongo.Client, error) {
	resolvedURI := uri.Get()
	if resolvedURI == "" {
		return nil, errors.New("URI for " + name + " is empty")
	}

	// Create our custom monitor to track state
	customMonitor := &poolMetricsMonitor{slog: p.baseLogger.With("pool", name)}
	p.monitors[name] = customMonitor

	// Create the driver's PoolMonitor and assign our event handler
	driverMonitor := &event.PoolMonitor{
		Event: customMonitor.Event,
	}

	clientOptions := options.Client().ApplyURI(resolvedURI).SetPoolMonitor(driverMonitor)

	if p.cfg.TLS != nil && p.cfg.TLS.Enabled {
		tlsConfig := new(tls.Config)
		tlsConfig.InsecureSkipVerify = false
		if p.cfg.TLS.CAFile.Get() != "" {
			caFile := p.cfg.TLS.CAFile.Get()
			tlsConfig.RootCAs = x509.NewCertPool()
			if ok := tlsConfig.RootCAs.AppendCertsFromPEM([]byte(caFile)); !ok {
				return nil, errors.New("failed to append CA cert")
			}
		}
		if p.cfg.TLS.CertFile.Get() != "" && p.cfg.TLS.KeyFile.Get() != "" {
			certFile := p.cfg.TLS.CertFile.Get()
			keyFile := p.cfg.TLS.KeyFile.Get()
			cert, err := tls.X509KeyPair([]byte(certFile), []byte(keyFile))
			if err != nil {
				return nil, fmt.Errorf("failed to load client key pair: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		clientOptions.SetTLSConfig(tlsConfig)
	}

	if p.cfg.Pool != nil {
		p.logger(ctx).Info("Applying pool settings", "name", name, "max_pool_size", p.cfg.Pool.MaxPoolSize)
		clientOptions.SetMaxPoolSize(p.cfg.Pool.MaxPoolSize)
		clientOptions.SetMinPoolSize(p.cfg.Pool.MinPoolSize)
		clientOptions.SetMaxConnIdleTime(p.cfg.Pool.MaxConnIdleTime)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for %s: %w", name, err)
	}

	if err := client.Ping(ctx, rp); err != nil {
		return nil, fmt.Errorf("failed to ping %s: %w", name, err)
	}

	return client, nil
}

func (p *MongoProvider) startHealthMonitor(ctx context.Context) {
	ticker := time.NewTicker(p.policy.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.checkHealth(ctx)
		}
	}
}

func (p *MongoProvider) startMetricsCollector(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.collectMetrics()
		}
	}
}

func (p *MongoProvider) collectMetrics() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for name, monitor := range p.monitors {
		p.metrics.SetGauge("db_pool_open_conns", float64(monitor.GetOpenConnections()), name)
	}
}

func (p *MongoProvider) checkHealth(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.primary != nil {
		if err := p.primary.Ping(ctx, readpref.Primary()); err != nil {
			p.logger(ctx).Warn("Primary ping failed", "error", err)
			p.handlePrimaryFailure(ctx, err)
		} else {
			p.handlePrimaryRecovery(ctx)
		}
	} else {
		p.attemptReconnect(ctx)
	}
}

func (p *MongoProvider) handlePrimaryFailure(ctx context.Context, err error) {
	if p.isPrimaryHealthy {
		p.logger(ctx).Error("Primary ping failed. Marking as unhealthy.", "error", err)
		p.isPrimaryHealthy = false
		p.metrics.SetGauge("db_primary_healthy", 0, "primary")
		p.metrics.Inc("db_failover_triggered", "primary")
		p.failureCount = 0
		p.reconnectBackoff = p.policy.ReconnectBackoffBase
	}
	p.failureCount++
	if p.failureCount > p.policy.MaxFailures {
		p.logger(ctx).Warn("Max failures reached for primary. Will attempt to reconnect.")
		p.failureCount = 0 // Reset after attempting
	}
}

func (p *MongoProvider) handlePrimaryRecovery(ctx context.Context) {
	if !p.isPrimaryHealthy {
		p.logger(ctx).Info("Primary has recovered.")
		p.isPrimaryHealthy = true
		p.metrics.SetGauge("db_primary_healthy", 1, "primary")
		p.failureCount = 0
		p.reconnectBackoff = p.policy.ReconnectBackoffBase
	}
}

func (p *MongoProvider) attemptReconnect(ctx context.Context) {
	p.logger(ctx).Info("Attempting to reconnect to primary", "backoff", p.reconnectBackoff)
	p.metrics.Inc("db_reconnect_attempts", "primary")
	primaryClient, err := p.connect(ctx, "primary", p.cfg.PrimaryURI, readpref.Primary())
	if err == nil {
		p.primary = primaryClient
		p.isPrimaryHealthy = true
		p.logger(ctx).Info("Reconnected to primary.")
	} else {
		p.reconnectBackoff *= 2
		if p.reconnectBackoff > p.policy.ReconnectBackoffMax {
			p.reconnectBackoff = p.policy.ReconnectBackoffMax
		}
		p.logger(ctx).Error("Failed to reconnect to primary", "error", err, "next_backoff", p.reconnectBackoff)
	}
}

// GetDB is deprecated. Use GetWriter() instead.
func (p *MongoProvider) GetDB() any {
	return p.GetWriter()
}

func (p *MongoProvider) GetWriter() any {
	ctx := context.Background() // Default context if not provided by caller
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.isPrimaryHealthy && p.primary != nil {
		return p.primary
	}

	if len(p.replicas) > 0 {
		// Failover to a replica if the primary is down
		nextReplicaIndex := (p.lastReplicaIndex + 1) % len(p.replicas)
		p.lastReplicaIndex = nextReplicaIndex
		p.logger(ctx).Warn("Primary is down, failing over write operations to replica", "replica_index", nextReplicaIndex)
		p.metrics.Inc("db_failover_to_replica", fmt.Sprintf("replica-%d", nextReplicaIndex))
		return p.replicas[nextReplicaIndex]
	}

	return p.primary // Returns primary (which is nil) if no replicas are available
}

func (p *MongoProvider) GetReader() any {
	ctx := context.Background() // Default context if not provided by caller
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.replicas) == 0 {
		// No replicas available, so return the writer (which could be a replica in failover mode)
		return p.GetWriter()
	}

	// Round-robin between replicas for read operations
	nextReplicaIndex := (p.lastReplicaIndex + 1) % len(p.replicas)
	p.lastReplicaIndex = nextReplicaIndex
	p.logger(ctx).Debug("Serving read from replica", "replica_index", nextReplicaIndex)
	p.metrics.Inc("db_read_from_replica", fmt.Sprintf("replica-%d", nextReplicaIndex))
	return p.replicas[nextReplicaIndex]
}

func (p *MongoProvider) Ping(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.primary == nil {
		return errors.New("primary database is not connected")
	}

	return p.primary.Ping(ctx, readpref.Primary())
}

func (p *MongoProvider) Close() error {
	ctx := context.Background() // Default context for closing
	p.mu.Lock()
	defer p.mu.Unlock()

	var allErrors []error
	if p.primary != nil {
		if err := p.primary.Disconnect(ctx); err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed closing primary: %w", err))
		}
	}

	for i, replica := range p.replicas {
		if err := replica.Disconnect(ctx); err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed closing replica %d: %w", i, err))
		}
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}

	p.logger(ctx).Info("All Mongo connections closed.")
	return nil
}
