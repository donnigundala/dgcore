package seeder

import "log"

// LogPrintf logging methods
// These methods prepend a [SEEDER] tag to log messages for clarity.
func LogPrintf(format string, v ...any) {
	log.Printf("[SEEDER] "+format, v...)
}

// LogPrintln logs a message with a [SEEDER] prefix.
// This is a convenience method to standardize seeder log output.
func LogPrintln(v ...any) {
	log.Println(append([]any{"[SEEDER]"}, v...)...)
}

// LogPrint logs a message with a [SEEDER] prefix.
// This is a convenience method to standardize seeder log output.
func LogPrint(v ...any) {
	log.Print(append([]any{"[SEEDER]"}, v...)...)
}

// LogFatalf logs a formatted message with a [SEEDER] prefix and exits the application.
// This is useful for critical errors during seeding that should halt execution.
func LogFatalf(format string, v ...any) {
	log.Fatalf("[SEEDER] "+format, v...)
}
