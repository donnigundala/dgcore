package mysql

type Config struct {
	Host                  string `mapstructure:"host" json:"host" yaml:"host"`
	Port                  string `mapstructure:"port" json:"port" yaml:"port"`
	Username              string `mapstructure:"username" json:"username" yaml:"username"`
	Password              string `mapstructure:"password" json:"password" yaml:"password"`
	Name                  string `mapstructure:"name" json:"name" yaml:"name"`
	Debug                 bool   `mapstructure:"debug" json:"debug" yaml:"debug"`
	Timezone              string `mapstructure:"timezone" json:"timezone" yaml:"timezone"`
	SSLMode               string `mapstructure:"ssl_mode" json:"ssl_mode" yaml:"ssl_mode"`
	MaxOpenConnection     string `mapstructure:"max_open_connection" json:"max_open_connection" yaml:"max_open_connection"`
	MaxConnectionLifetime string `mapstructure:"max_connection_lifetime" json:"max_connection_lifetime" yaml:"max_connection_lifetime"`
	MaxIdleLifetime       string `mapstructure:"max_idle_lifetime" json:"max_idle_lifetime" yaml:"max_idle_lifetime"`
	Driver                string `mapstructure:"driver" json:"driver" yaml:"driver"`
}

//func NewConfig(appCfg *configs.MysqlConfig) *Config {
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
//		Driver:                "mysql",
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
		Driver:                "mysql",
	}
}
