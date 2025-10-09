// Package database provides a flexible, driver-based database connector architecture.
package database

import (
	"fmt"

	"github.com/donnigundala/dg-frame/pkg/database/mongo"
	"github.com/donnigundala/dg-frame/pkg/database/mysql"
	"github.com/donnigundala/dg-frame/pkg/database/pgsql"
)

type Connector interface {
	Connect() (any, error)
	Close() error
}

func NewDatabase(driver string, cfg any) (Connector, error) {
	switch driver {
	case "postgres", "pgsql":
		pgsqlCfg, ok := cfg.(*pgsql.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for PgSQL")
		}

		// Create a new PgSQL connector
		db, err := pgsql.NewPostgres(pgsqlCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create PgSQL connector: %w", err)
		}
		return db, nil
	case "mysql":
		mysqlCfg, ok := cfg.(*mysql.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for MySQL")
		}

		// Create a new MySQL connector
		db, err := mysql.NewMysql(mysqlCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create MySQL connector: %w", err)
		}
		return db, nil
	case "mongo":
		mongoCfg, ok := cfg.(*mongo.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for MongoDB")
		}

		// Create a new MongoDB connector
		db, err := mongo.NewMongoDB(mongoCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create MongoDB connector: %w", err)
		}
		return db, nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
}
