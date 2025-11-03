package dgsql

const (
	DriverMySQL    Driver = "mysql"
	DriverPostgres Driver = "postgres"
	DriverPgSQL    Driver = "pgsql"
	DriverSQLite   Driver = "sqlite"

	LogSilent LogLevel = "silent"
	LogError  LogLevel = "error"
	LogWarn   LogLevel = "warn"
	LogInfo   LogLevel = "info"

	MySQLDefaultPort DefaultPort = 3306
	PgSQLDefaultPort DefaultPort = 5432
)

type Driver string
type LogLevel string
type DefaultPort int
