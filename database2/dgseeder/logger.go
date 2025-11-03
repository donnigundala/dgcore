package dgseeder

import (
	"log"
	"os"
)

var (
	defaultLogger = &SeederLogger{log.New(os.Stdout, "", log.LstdFlags)}
	logger        = defaultLogger
)

// SeederLogger wraps the standard log.Logger and adds custom logging helpers.
type SeederLogger struct {
	*log.Logger
}

// LogPrint returns the current SeederLogger instance.
func (l *SeederLogger) LogPrint(v ...interface{}) {
	LogPrint(v...)
}

// LogPrintf logs a formatted message using the Seeder logger.
func (l *SeederLogger) LogPrintf(format string, v ...interface{}) {
	LogPrintf(format, v...)
}

// LogPrintln logs a message with a newline using the Seeder logger.
func (l *SeederLogger) LogPrintln(v ...interface{}) {
	LogPrintln(v...)
}

// LogFatalf logs a formatted message with a [SEEDER] prefix and exits the application.
func (l *SeederLogger) LogFatalf(format string, v ...interface{}) {
	LogFatalf(format, v...)
}

// Success logs a success message with ✅ prefix.
func (l *SeederLogger) Success(msg string) {
	LogPrintln("✅", msg)
}

// Error logs an error message with ❌ prefix.
func (l *SeederLogger) Error(msg string) {
	LogPrintln("❌", msg)
}

// Warn logs a warning message with ⚠️ prefix.
func (l *SeederLogger) Warn(msg string) {
	LogPrintln("⚠️", msg)
}

// LogPrint logs a message with a [SEEDER] prefix.
// This is a convenience method to standardize seeder log output.
func LogPrint(v ...interface{}) {
	logger.Print(append([]interface{}{"[SEEDER] "}, v...)...)
}

// LogPrintf logging methods
// These methods prepend a [SEEDER] tag to log messages for clarity.
func LogPrintf(format string, v ...interface{}) {
	logger.Printf("[SEEDER] "+format, v...)
}

// LogPrintln logs a message with a [SEEDER] prefix.
// This is a convenience method to standardize seeder log output.
func LogPrintln(v ...interface{}) {
	logger.Println(append([]interface{}{"[SEEDER]"}, v...)...)
}

// LogFatalf logs a formatted message with a [SEEDER] prefix and exits the application.
// This is useful for critical errors during seeding that should halt execution.
func LogFatalf(format string, v ...interface{}) {
	logger.Fatalf("[SEEDER] "+format, v...)
}

// Logger returns the current SeederLogger instance.
func Logger() *SeederLogger {
	return logger
}

// SetLogger allows overriding the default logger.
func SetLogger(l *SeederLogger) {
	if l != nil {
		logger = l
	}
}
