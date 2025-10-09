package validation

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// Error validationError is a custom error type that holds field-specific errors.
type Error struct {
	Errors map[string]string
}

// ParseValidationErrors processes validation errors and returns a map with field names and custom error messages.
func ParseValidationErrors(err error, obj interface{}) error {
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		customErrors := make(map[string]string)
		// Get the reflection type of the object
		objType := reflect.TypeOf(obj).Elem()

		for _, e := range errs {
			// Get the field, and find the struct field by name
			field, found := objType.FieldByName(e.Field())
			if !found {
				// If the field is not found, use the default field name
				customErrors[e.Field()] = fmt.Sprintf("validation failed on %s", e.Tag())
				continue
			}

			// Get the JSON tag for the field, fallback to field name if not present
			// change e.Field() to jsonTag if you prefer to use json tag
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				jsonTag = e.Field()
			}

			// Customize error messages
			switch e.Tag() {
			case "required":
				customErrors[e.Field()] = fmt.Sprintf("wajib di isi")
			case "email":
				customErrors[e.Field()] = fmt.Sprintf("format email tidak valid")
			case "gte":
				customErrors[e.Field()] = fmt.Sprintf("%s harus lebih besar atau sama dengan %s", e.Field(), e.Param())
			case "unique":
				customErrors[e.Field()] = fmt.Sprintf("sudah digunakan")
			case "exists":
				customErrors[e.Field()] = fmt.Sprintf("tidak di temukan")
			default:
				customErrors[e.Field()] = fmt.Sprintf("data tidak valid")
			}
		}

		return &Error{Errors: customErrors}
	}

	return nil
}

func (v *Error) Error() string {
	return "Validation failed"
}
