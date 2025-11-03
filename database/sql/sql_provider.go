package sql

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/donnigundala/dgcore/database/config"
	"github.com/donnigundala/dgcore/database/contracts"
	"github.com/donnigundala/dgcore/database/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type SQLProvider struct {
	cfg        *config.SQLConfig
	policy     *config.HealthPolicy
	metrics    contracts.MetricsProvider
	baseLogger *slog.Logger // Base logger for the provider
	traceIDKey string       // Key to extract trace ID from context
	primary    *gorm.DB
	replicas   []*gorm.DB
	mu         sync.RWMutex

	isPrimaryHealthy bool
	lastReplicaIndex int
	failureCount     int
	reconnectBackoff time.Duration
}

func NewSQLProvider(ctx context.Context, cfg *config.SQLConfig, policy *config.HealthPolicy, metrics contracts.MetricsProvider, baseLogger *slog.Logger, traceIDKey string) (contracts.Provider, error) {
	if cfg == nil {
		return nil, errors.New("SQL config cannot be nil")
	}
	if policy == nil {
		policy = config.DefaultPolicy()
	}
	if baseLogger == nil {
		baseLogger = slog.Default()
	}

	p := &SQLProvider{
		cfg:              cfg,
		policy:           policy,
		metrics:          metrics,
		baseLogger:       baseLogger.With("component", "sql_provider"),
		traceIDKey:       traceIDKey,
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
func (p *SQLProvider) logger(ctx context.Context) *slog.Logger {
	if p.traceIDKey != "" {
		if traceID, ok := ctx.Value(p.traceIDKey).(string); ok && traceID != "" {
			return p.baseLogger.With("trace_id", traceID)
		}
	}
	return p.baseLogger
}

func (p *SQLProvider) connectAll(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Connect to Primary
	if p.cfg.Primary != nil {
		primaryDB, err := p.connect(ctx, "primary", p.cfg.Primary)
		if err != nil {
			p.logger(ctx).Error("Failed to connect to primary DB. Entering failover mode.", "error", err)
			p.metrics.Inc("db_connection_errors", "primary")
		} else {
			p.primary = primaryDB
			p.isPrimaryHealthy = true
			p.metrics.SetGauge("db_primary_healthy", 1, "primary")
			p.logger(ctx).Info("Primary connected")
		}
	} else {
		return errors.New("primary SQL connection details are nil")
	}

	// Connect to Replicas
	for i, details := range p.cfg.Replicas {
		name := fmt.Sprintf("replica-%d", i)
		replicaDB, err := p.connect(ctx, name, details)
		if err != nil {
			p.logger(ctx).Error("Failed to connect to replica", "name", name, "error", err)
			p.metrics.Inc("db_connection_errors", name)
		} else {
			p.replicas = append(p.replicas, replicaDB)
			p.logger(ctx).Info("Replica connected", "name", name)
		}
	}

	if p.primary == nil && len(p.replicas) == 0 {
		return errors.New("failed to connect to any SQL database")
	}

	return nil
}

func (p *SQLProvider) connect(ctx context.Context, name string, details *config.SQLConnectionDetails) (*gorm.DB, error) {
	var dsn string
	var dialector gorm.Dialector
	var err error

	// Prioritize RawDSN if provided
	if rawDSN := details.RawDSN.Get(); rawDSN != "" {
		dsn = rawDSN
	} else {
		// Otherwise, construct DSN from individual parameters
		dsn, err = p.buildDSN(details)
		if err != nil {
			return nil, fmt.Errorf("failed to build DSN for %s: %w", name, err)
		}
	}

	// Select GORM Dialector based on DriverName
	switch p.cfg.DriverName {
	case "postgres":
		dialector = postgres.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported SQL driver: %s", p.cfg.DriverName)
	}

	gormConfig := &gorm.Config{}
	gormConfig.Logger = logger.NewSlogGormLogger(p.baseLogger.With("db_instance", name), gormlogger.Config{
		SlowThreshold:             200 * time.Millisecond, // Default slow query threshold
		LogLevel:                  p.gormLogLevel(),
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})

	gormDB, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm connection for %s: %w", name, err)
	}

	// Apply pool settings if provided
	if p.cfg.Pool != nil {
		sqlDB, err := gormDB.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get underlying sql.DB for %s: %w", name, err)
		}
		p.logger(ctx).Info("Applying pool settings", "name", name, "max_open_conns", p.cfg.Pool.MaxOpenConns)
		sqlDB.SetMaxOpenConns(p.cfg.Pool.MaxOpenConns)
		sqlDB.SetMaxIdleConns(p.cfg.Pool.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(p.cfg.Pool.ConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(p.cfg.Pool.ConnMaxIdleTime)
	}

	return gormDB, nil
}

func (p *SQLProvider) buildDSN(details *config.SQLConnectionDetails) (string, error) {
	sb := strings.Builder{}

	switch p.cfg.DriverName {
	case "postgres":
		if host := details.Host.Get(); host != "" {
			sb.WriteString(fmt.Sprintf("host=%s ", host))
		}
		if port := details.Port.Get(); port != "" {
			sb.WriteString(fmt.Sprintf("port=%s ", port))
		}
		if user := details.User.Get(); user != "" {
			sb.WriteString(fmt.Sprintf("user=%s ", user))
		}
		if password := details.Password.Get(); password != "" {
			sb.WriteString(fmt.Sprintf("password=%s ", password))
		}
		if dbName := details.DBName.Get(); dbName != "" {
			sb.WriteString(fmt.Sprintf("dbname=%s ", dbName))
		}

		// Add TLS settings for Postgres
		if p.cfg.TLS != nil && p.cfg.TLS.Enabled {
			if p.cfg.TLS.Insecure {
				sb.WriteString("sslmode=disable ")
			} else {
				sb.WriteString("sslmode=require ") // Or verify-full, depending on CA/Cert/Key
			}
			// For CA/Cert/Key, pgx requires registering custom TLS config, which is more complex
			// and might be better handled by RawDSN or specific driver options if needed.
			p.logger(context.Background()).Warn("Custom TLS certs/keys for Postgres are not fully supported via params. Consider RawDSN.")
		} else {
			sb.WriteString("sslmode=disable ") // Default to disable if TLS not enabled
		}

		// Add other params
		for k, v := range details.Params {
			sb.WriteString(fmt.Sprintf("%s=%s ", k, v.Get()))
		}
		return strings.TrimSpace(sb.String()), nil

	case "mysql":
		user := details.User.Get()
		password := details.Password.Get()
		host := details.Host.Get()
		port := details.Port.Get()
		dbName := details.DBName.Get()

		if user == "" || host == "" || dbName == "" {
			return "", errors.New("user, host, and db_name are required for MySQL")
		}

		addr := host
		if port != "" {
			addr = fmt.Sprintf("%s:%s", host, port)
		}

		sb.WriteString(fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, addr, dbName))

		params := url.Values{}
		params.Add("charset", "utf8mb4")
		params.Add("parseTime", "True")
		params.Add("loc", "Local")

		// Add TLS settings for MySQL
		if p.cfg.TLS != nil && p.cfg.TLS.Enabled {
			if p.cfg.TLS.Insecure {
				params.Add("tls", "skip-verify")
			} else {
				// For full TLS with CA/Cert/Key, MySQL driver requires registering custom TLS config
				// which is complex. For now, we assume 'true' or 'skip-verify'.
				params.Add("tls", "true")
				p.logger(context.Background()).Warn("Custom TLS certs/keys for MySQL are not fully supported via params. Consider RawDSN.")
			}
		}

		// Add other params
		for k, v := range details.Params {
			params.Add(k, v.Get())
		}

		return sb.String() + "?" + params.Encode(), nil

	case "sqlite":
		dbName := details.DBName.Get()
		if dbName == "" {
			return "", errors.New("db_name is required for SQLite")
		}
		sb.WriteString(fmt.Sprintf("file:%s", dbName))

		params := url.Values{}
		params.Add("_foreign_keys", "on")
		// Add other params
		for k, v := range details.Params {
			params.Add(k, v.Get())
		}
		return sb.String() + "?" + params.Encode(), nil

	default:
		return "", fmt.Errorf("unsupported SQL driver for DSN construction: %s", p.cfg.DriverName)
	}
}

func (p *SQLProvider) gormLogLevel() gormlogger.LogLevel {
	switch p.cfg.LogLevel {
	case config.LogLevelSilent:
		return gormlogger.Silent
	case config.LogLevelError:
		return gormlogger.Error
	case config.LogLevelWarn:
		return gormlogger.Warn
	case config.LogLevelInfo:
		return gormlogger.Info
	default:
		return gormlogger.Warn // Default to Warn if not specified
	}
}

func (p *SQLProvider) createTLSConfig() (*tls.Config, error) {
	caCert := []byte(p.cfg.TLS.CA.Get())
	cert := []byte(p.cfg.TLS.Cert.Get())
	key := []byte(p.cfg.TLS.Key.Get())

	caCertPool := x509.NewCertPool()
	if len(caCert) > 0 {
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to append CA cert")
		}
	}

	clientCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load client key pair: %w", err)
	}

	return &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{clientCert},
		InsecureSkipVerify: p.cfg.TLS.Insecure,
	}, nil
}

func (p *SQLProvider) startHealthMonitor(ctx context.Context) {
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

func (p *SQLProvider) startMetricsCollector(ctx context.Context) {
	if p.primary == nil {
		return
	}
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

func (p *SQLProvider) collectMetrics() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.primary != nil {
		sqlDB, err := p.primary.DB()
		if err == nil {
			stats := sqlDB.Stats()
			p.metrics.SetGauge("db_pool_open_conns", float64(stats.OpenConnections), "primary")
			p.metrics.SetGauge("db_pool_in_use_conns", float64(stats.InUse), "primary")
			p.metrics.SetGauge("db_pool_idle_conns", float64(stats.Idle), "primary")
		}
	}

	for i, replica := range p.replicas {
		name := fmt.Sprintf("replica-%d", i)
		sqlDB, err := replica.DB()
		if err == nil {
			stats := sqlDB.Stats()
			p.metrics.SetGauge("db_pool_open_conns", float64(stats.OpenConnections), name)
			p.metrics.SetGauge("db_pool_in_use_conns", float64(stats.InUse), name)
			p.metrics.SetGauge("db_pool_idle_conns", float64(stats.Idle), name)
		}
	}
}

func (p *SQLProvider) checkHealth(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.primary != nil {
		sqlDB, _ := p.primary.DB()
		if err := sqlDB.PingContext(ctx); err != nil {
			p.logger(ctx).Warn("Primary ping failed", "error", err)
			p.handlePrimaryFailure(ctx, err)
		} else {
			p.handlePrimaryRecovery(ctx)
		}
	} else {
		p.attemptReconnect(ctx)
	}
}

func (p *SQLProvider) handlePrimaryFailure(ctx context.Context, err error) {
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

func (p *SQLProvider) handlePrimaryRecovery(ctx context.Context) {
	if !p.isPrimaryHealthy {
		p.logger(ctx).Info("Primary has recovered.")
		p.isPrimaryHealthy = true
		p.metrics.SetGauge("db_primary_healthy", 1, "primary")
		p.failureCount = 0
		p.reconnectBackoff = p.policy.ReconnectBackoffBase
	}
}

func (p *SQLProvider) attemptReconnect(ctx context.Context) {
	p.logger(ctx).Info("Attempting to reconnect to primary", "backoff", p.reconnectBackoff)
	p.metrics.Inc("db_reconnect_attempts", "primary")
	db, err := p.connect(ctx, "primary", p.cfg.Primary)
	if err == nil {
		p.primary = db
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
func (p *SQLProvider) GetDB() any {
	return p.GetWriter()
}

func (p *SQLProvider) GetWriter() any {
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

func (p *SQLProvider) GetReader() any {
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

func (p *SQLProvider) Ping(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.primary == nil {
		return errors.New("primary database is not connected")
	}

	sqlDB, err := p.primary.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (p *SQLProvider) Close() error {
	ctx := context.Background() // Default context for closing
	p.mu.Lock()
	defer p.mu.Unlock()

	var allErrors []error
	if p.primary != nil {
		sqlDB, _ := p.primary.DB()
		if err := sqlDB.Close(); err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed closing primary: %w", err))
		}
	}

	for i, replica := range p.replicas {
		sqlDB, _ := replica.DB()
		if err := sqlDB.Close(); err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed closing replica %d: %w", i, err))
		}
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}

	p.logger(ctx).Info("All SQL connections closed.")
	return nil
}
