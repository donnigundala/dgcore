package main

import (
	"fmt"
	"log"

	"github.com/donnigundala/dg-core/errors"
)

// MockReporter is an example error reporter implementation.
type MockReporter struct {
	name string
}

func (m *MockReporter) Report(err error, context map[string]interface{}) error {
	fmt.Printf("[%s] Error reported: %v\n", m.name, err)
	if context != nil {
		fmt.Printf("[%s] Context: %+v\n", m.name, context)
	}
	return nil
}

func main() {
	fmt.Println("=== Error Reporter Example ===\n")

	// 1. Without reporter (default)
	fmt.Println("1. Without reporter:")
	err1 := errors.New("database connection failed")
	errors.Report(err1, map[string]interface{}{
		"user_id": 123,
	})
	fmt.Println("   (No output - no reporter configured)\n")

	// 2. With mock reporter
	fmt.Println("2. With mock reporter:")
	reporter := &MockReporter{name: "MockSentry"}
	errors.SetReporter(reporter)

	err2 := errors.New("user not found").WithCode("USER_NOT_FOUND")
	errors.Report(err2, map[string]interface{}{
		"user_id":    456,
		"request_id": "abc-123",
	})
	fmt.Println()

	// 3. Report with defaults
	fmt.Println("3. Report with defaults:")
	err3 := errors.New("validation failed")
	errors.ReportWithDefaults(err3)
	fmt.Println()

	// 4. Wrapped error
	fmt.Println("4. Wrapped error:")
	originalErr := fmt.Errorf("connection timeout")
	wrappedErr := errors.Wrap(originalErr, "failed to connect to database")
	errors.Report(wrappedErr, map[string]interface{}{
		"database": "postgres",
		"host":     "localhost",
	})
	fmt.Println()

	fmt.Println("=== Example Complete ===")
	fmt.Println("\nTo integrate with real error reporting services:")
	fmt.Println("1. Install plugin: go get github.com/donnigundala/dg-core-sentry")
	fmt.Println("2. Initialize: reporter := sentry.NewReporter(dsn)")
	fmt.Println("3. Set reporter: errors.SetReporter(reporter)")
	fmt.Println("4. All errors.Report() calls will now go to Sentry!")

	log.Println("\nDone!")
}
