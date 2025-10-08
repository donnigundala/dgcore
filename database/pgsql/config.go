package pgsql

// Config holds PostgreSQL configuration.
type Config struct {
	Host                  string
	Port                  string
	Username              string
	Password              string
	Name                  string
	Debug                 bool
	Timezone              string
	SSLMode               string
	MaxOpenConnection     string
	MaxConnectionLifetime string
	MaxIdleLifetime       string
	Driver                string
}

//func NewConfig(appCfg *newconfig.PgsqlConfig) *Config {
//	if appCfg == nil {
//		return &Config{}
//	}
//
//	return &Config{
//		Host:                  appCfg.Host,
//		Port:                  appCfg.Port,
//		Username:              appCfg.Username,
//		Password:              appCfg.Password,
//		Name:                  appCfg.Name,
//		Debug:                 appCfg.Debug,
//		Timezone:              appCfg.Timezone,
//		SSLMode:               appCfg.SSLMode,
//		MaxOpenConnection:     appCfg.MaxOpenConnection,
//		MaxConnectionLifetime: appCfg.MaxConnectionLifetime,
//		MaxIdleLifetime:       appCfg.MaxIdleLifetime,
//		Driver:                "pgsql",
//	}
//}

func defaultConfig() *Config {
	return &Config{
		Host:                  "localhost",
		Port:                  "5432",
		Username:              "postgres",
		Password:              "",
		Name:                  "postgres",
		Debug:                 false,
		Timezone:              "UTC",
		SSLMode:               "disable",
		MaxOpenConnection:     "10",
		MaxConnectionLifetime: "5m",
		MaxIdleLifetime:       "5m",
		Driver:                "pgsql",
	}
}
