package logger

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// SlogGormLogger is a GORM logger that uses slog for structured logging.
type SlogGormLogger struct {
	slog   *slog.Logger
	config gormlogger.Config
}

// NewSlogGormLogger creates a new SlogGormLogger.
func NewSlogGormLogger(slog *slog.Logger, config gormlogger.Config) gormlogger.Interface {
	return &SlogGormLogger{
		slog:   slog,
		config: config,
	}
}

// LogMode sets the log level.
func (l *SlogGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newlogger := *l
	newlogger.config.LogLevel = level
	return &newlogger
}

// Info logs an informational message.
func (l *SlogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gormlogger.Info {
		l.slog.InfoContext(ctx, msg, "data", data)
	}
}

// Warn logs a warning message.
func (l *SlogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gormlogger.Warn {
		l.slog.WarnContext(ctx, msg, "data", data)
	}
}

// Error logs an error message.
func (l *SlogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.config.LogLevel >= gormlogger.Error {
		l.slog.ErrorContext(ctx, msg, "data", data)
	}
}

// Trace logs a SQL query.
func (l *SlogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.config.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.config.LogLevel >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		l.slog.ErrorContext(ctx, "GORM query error",
			"error", err,
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
	case elapsed > l.config.SlowThreshold && l.config.SlowThreshold != 0 && l.config.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		l.slog.WarnContext(ctx, "GORM slow query",
			"threshold", l.config.SlowThreshold,
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
	case l.config.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		l.slog.InfoContext(ctx, "GORM query",
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
	}
}
