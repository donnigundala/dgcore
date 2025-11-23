package errors

// Reporter defines the interface for error reporting plugins (e.g., Sentry, Rollbar).
type Reporter interface {
	// Report sends an error to the external reporting service.
	// The context map can contain additional information like user_id, request_id, etc.
	Report(err error, context map[string]interface{}) error
}

// globalReporter holds the configured error reporter (if any).
var globalReporter Reporter

// SetReporter sets the global error reporter.
// This is typically called during application initialization.
// Example:
//
//	reporter := sentry.NewReporter(dsn)
//	errors.SetReporter(reporter)
func SetReporter(reporter Reporter) {
	globalReporter = reporter
}

// GetReporter returns the currently configured reporter (or nil if none).
func GetReporter() Reporter {
	return globalReporter
}

// Report reports an error to the configured reporter (if any).
// If no reporter is configured, this is a no-op.
// The context map can contain additional information to enrich the error report.
// Example:
//
//	errors.Report(err, map[string]interface{}{
//	    "user_id": 123,
//	    "request_id": "abc-123",
//	})
func Report(err error, context map[string]interface{}) {
	if globalReporter != nil {
		globalReporter.Report(err, context)
	}
}

// ReportWithDefaults reports an error with default context.
// This is a convenience function for simple error reporting.
func ReportWithDefaults(err error) {
	Report(err, nil)
}
