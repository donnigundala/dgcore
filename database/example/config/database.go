package config

// This file serves as a template for consumers of the database library.
// It should be copied to the consumer's own `config` directory.

//func init() {
//	// Register a default configuration for the 'databases' block.
//	// The consumer's application configuration will override these defaults.
//	config.Add("databases", map[string]any{
//		"default_trace_id_key": "X-Trace-ID", // Default trace ID key for context
//
//		// --- Example PostgreSQL Configuration ---
//		"my_postgres": map[string]any{
//			"driver": "sql",
//			"policy": map[string]any{
//				"ping_interval":          "10s",
//				"max_failures":           3,
//				"reconnect_backoff_base": "1s",
//				"reconnect_backoff_max":  "30s",
//			},
//			"sql": map[string]any{
//				"driver_name": "postgres",
//				"primary": map[string]any{
//					"host":     config.Env("POSTGRES_PRIMARY_HOST", "localhost"),
//					"port":     config.Env("POSTGRES_PRIMARY_PORT", "5432"),
//					"user":     config.Env("POSTGRES_PRIMARY_USER", "postgres"),
//					"password": config.Env("POSTGRES_PRIMARY_PASSWORD", "password"),
//					"db_name":  config.Env("POSTGRES_PRIMARY_DBNAME", "test"),
//				},
//				"replicas": []map[string]any{
//					{
//						"host":     config.Env("POSTGRES_REPLICA_1_HOST", "localhost"),
//						"port":     config.Env("POSTGRES_REPLICA_1_PORT", "5433"),
//						"user":     config.Env("POSTGRES_REPLICA_1_USER", "postgres"),
//						"password": config.Env("POSTGRES_REPLICA_1_PASSWORD", "password"),
//						"db_name":  config.Env("POSTGRES_REPLICA_1_DBNAME", "test"),
//					},
//				},
//				"pool": map[string]any{
//					"max_open_conns":    25,
//					"max_idle_conns":    10,
//					"conn_max_lifetime": "1h",
//				},
//				"tls": map[string]any{
//					"enabled": false,
//				},
//				"log_level": "warn", // Default GORM log level
//			},
//		},
//
//		// --- Example MySQL Configuration ---
//		"my_mysql": map[string]any{
//			"driver": "sql",
//			"sql": map[string]any{
//				"driver_name": "mysql",
//				"primary": map[string]any{
//					"host":     config.Env("MYSQL_PRIMARY_HOST", "localhost"),
//					"port":     config.Env("MYSQL_PRIMARY_PORT", "3306"),
//					"user":     config.Env("MYSQL_PRIMARY_USER", "root"),
//					"password": config.Env("MYSQL_PRIMARY_PASSWORD", "password"),
//					"db_name":  config.Env("MYSQL_PRIMARY_DBNAME", "test"),
//				},
//				"pool": map[string]any{
//					"max_open_conns":    25,
//					"max_idle_conns":    10,
//					"conn_max_lifetime": "1h",
//				},
//				"log_level": "warn",
//			},
//		},
//
//		// --- Example SQLite Configuration ---
//		"my_sqlite": map[string]any{
//			"driver": "sql",
//			"sql": map[string]any{
//				"driver_name": "sqlite",
//				"primary": map[string]any{
//					"db_name": config.Env("SQLITE_DBNAME", "test.db"), // SQLite uses file path as db_name
//				},
//				"log_level": "warn",
//			},
//		},
//
//		// --- Example MongoDB Configuration ---
//		"my_mongo": map[string]any{
//			"driver": "mongo",
//			"mongo": map[string]any{
//				"primary_uri": map[string]any{
//					"from_env": "MONGO_PRIMARY_URI",
//				},
//				"database": "my_app_db",
//				"pool": map[string]any{
//					"max_pool_size": 50,
//				},
//				"log_level": "info", // Default Mongo driver log level
//			},
//		},
//	})
//}
