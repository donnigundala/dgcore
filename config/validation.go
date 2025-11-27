package config

import (
	"fmt"
	"reflect"
)

// validateStruct validates a struct using go-playground/validator if available.
// This is kept separate to make the validator dependency optional.
func validateStruct(target any) error {
	// Check if the target is a pointer to a struct
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("validation target must be a pointer to a struct")
	}

	// Check if validate tag is present in the struct
	// If no validate tags, skip validation silently
	hasValidateTags := false
	structType := v.Elem().Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if _, ok := field.Tag.Lookup("validate"); ok {
			hasValidateTags = true
			break
		}
	}

	if !hasValidateTags {
		// No validation tags found, skip validation
		return nil
	}

	// If validate tags exist but validator is not available,
	// return a helpful error message
	return fmt.Errorf(
		"configuration has 'validate' tags but validation is not available. " +
			"To use InjectAndValidate, import github.com/go-playground/validator/v10 in your application",
	)
}
