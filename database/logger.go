package db

import (
	"log"
	"os"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

func newGormLogger(level LogLevel, slowThreshold time.Duration) gormlogger.Interface {
	var gormLevel gormlogger.LogLevel
	switch level {
	case LogSilent:
		gormLevel = gormlogger.Silent
	case LogError:
		gormLevel = gormlogger.Error
	case LogWarn:
		gormLevel = gormlogger.Warn
	default:
		gormLevel = gormlogger.Info
	}

	return gormlogger.New(
		log.New(os.Stdout, "[gorm] ", log.LstdFlags|log.Lmicroseconds),
		gormlogger.Config{
			SlowThreshold:             slowThreshold,
			LogLevel:                  gormLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
}
