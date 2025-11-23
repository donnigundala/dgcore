package validation

// Error is a custom error type that holds field-specific validation errors.
type Error struct {
	Errors map[string]string
}

// Error implements the error interface.
// This provides a default error message for the whole validation failure.
func (v *Error) Error() string {
	// You can make this more descriptive if you want, e.g., by joining all messages.
	// For now, a generic message is often sufficient as the structured `Errors` map is what's usually consumed.
	return "Validation failed"
}
