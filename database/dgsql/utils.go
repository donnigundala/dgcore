package dgsql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ----------------------------------------------------------------------
// Utility functions
// ------------------------------------------------------

// encodeParams encodes a map of string key-value pairs into a URL query string format.
func encodeParams(params map[string]string) string {
	first := true
	out := ""
	for k, v := range params {
		if first {
			out += fmt.Sprintf("%s=%s", k, v)
			first = false
			continue
		}
		out += fmt.Sprintf("&%s=%s", k, v)
	}
	return out
}

// pingWithTimeout attempts to ping the database with a specified timeout.
func pingWithTimeout(db *sql.DB, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return db.PingContext(ctx)
}

// sleep pauses execution for an exponentially increasing duration based on the attempt number.
func sleep(attempt int, base time.Duration) {
	// exponential backoff with jitter could be added; keep simple exponential for now
	delay := base << attempt
	if delay > 10*time.Second {
		delay = 10 * time.Second
	}
	time.Sleep(delay)
}

// applySQLitePragmas sets recommended PRAGMA settings for SQLite databases.
func applySQLitePragmas(db *gorm.DB) error {
	// WAL improves concurrency; busy_timeout prevents immediate SQLITE_BUSY errors
	if err := db.Exec("PRAGMA journal_mode = WAL;").Error; err != nil {
		return err
	}
	if err := db.Exec("PRAGMA busy_timeout = 5000;").Error; err != nil {
		return err
	}
	return nil
}

// nonEmpty returns s if it's non-empty; otherwise returns fallback.
func nonEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
