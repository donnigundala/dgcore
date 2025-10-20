package dghttp

import "log"

// logPrintf logging methods
// These methods prepend a [SEEDER] tag to log messages for clarity.
func logPrintf(format string, v ...any) {
	log.Printf("[SERVER] "+format, v...)
}

// logPrintln logs a message with a [SERVER] prefix.
// This is a convenience method to standardize seeder log output.
func logPrintln(v ...any) {
	log.Println(append([]any{"[SERVER]"}, v...)...)
}

// logPrint logs a message with a [SERVER] prefix.
// This is a convenience method to standardize seeder log output.
func logPrint(v ...any) {
	log.Print(append([]any{"[SERVER]"}, v...)...)
}

// logFatalf logs a formatted message with a [SERVER] prefix and exits the application.
// This is useful for critical errors during seeding that should halt execution.
func logFatalf(format string, v ...any) {
	log.Fatalf("[SERVER] "+format, v...)
}
