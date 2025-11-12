# Validation Package

The `validation` package provides a robust, flexible, and production-grade solution for validating data structures in your Go applications. It leverages the powerful `go-playground/validator/v10` library and extends it with features crucial for professional development, such as custom error messages, internationalization (i18n) support, custom field name resolution, and easy registration of custom validation rules.

## Features

-   **Pure & Decoupled:** No direct dependencies on database layers (like GORM), making it highly reusable.
-   **Extensible:** Easily register custom field-level and struct-level validation rules.
-   **Customizable Error Messages:** Override default validation messages for specific tags.
-   **Internationalization (i18n) Support:** Provide different sets of error messages based on the request's locale.
-   **Flexible Field Naming:** Control how field names appear in error messages using custom struct tags.
-   **Smart Fallback:** If a custom message is not found, it gracefully falls back to the `go-playground/validator`'s default message.

## Installation

This package is part of the `dgcore` framework. Ensure you have `dgcore` installed:

```bash
go get github.com/donnigundala/dgcore
```

## Usage

### 1. Defining Your Structs

Define your data structures with `go-playground/validator` tags. You can also add a custom tag (e.g., `display`) for more human-readable field names in error messages.

```go
package main

import "github.com/go-playground/validator/v10"

type UserRegistration struct {
	Username             string `validate:"required,min=3,max=20" json:"username" display:"Username"`
	Email                string `validate:"required,email" json:"email" display:"Email Address"`
	Password             string `validate:"required,min=8" json:"password" display:"Password"`
	PasswordConfirmation string `validate:"required,eqfield=Password" json:"password_confirmation" display:"Password Confirmation"`
	Framework            string `validate:"required,is_awesome" json:"framework" display:"Framework Name"`
}
```

### 2. Initializing the Validator

The `validation.New()` function accepts functional options to configure the validator.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/donnigundala/dgcore/validation"
	"github.com/go-playground/validator/v10"
)

// Custom field-level validation logic
func isAwesome(fl validator.FieldLevel) bool {
	return fl.Field().String() == "dg-framework"
}

func main() {
	// Define custom messages for different locales
	enMessages := map[string]string{
		"required":   "The '%s' field is required.",
		"email":      "The value provided for '%s' is not a valid email.",
		"min":        "The '%s' field must be at least %s characters long.",
		"eqfield":    "The '%s' field must match the %s field.",
		"is_awesome": "The framework must be 'dg-framework'!",
	}

	idMessages := map[string]string{
		"required":   "Kolom '%s' wajib diisi.",
		"email":      "Format email untuk '%s' tidak valid.",
		"min":        "Kolom '%s' minimal harus %s karakter.",
		"eqfield":    "Kolom '%s' harus sama dengan kolom %s.",
		"is_awesome": "Framework harus 'dg-framework'!",
	}

	// Initialize the Validator with all desired features
	v := validation.New(
		// Use the "display" tag for custom field names in error messages
		validation.WithFieldNameTag("display"),
		// Set the default language for messages (used if no specific locale is found)
		validation.WithDefaultLocale("en"),
		// Register messages for English
		validation.WithLocaleMessages("en", enMessages),
		// Register messages for Indonesian
		validation.WithLocaleMessages("id", idMessages),
		// Register a custom field-level validation rule
		validation.WithCustomValidation("is_awesome", isAwesome),
		// Register a struct-level validation rule (example - uncomment if needed)
		// validation.WithStructValidation(passwordsMatch, UserRegistration{}),
	)

	// ... rest of your application logic
}
```

### 3. Performing Validation

Use the `ValidateStruct` method, passing a `context.Context` and the struct to be validated.

```go
// ... inside your main function or handler

// Scenario A: Failing validation with default locale (English)
fmt.Println("\n--- Scenario A: Failing Validation (English) ---")
invalidUser := UserRegistration{
	Username:             "dj",
	Email:                "invalid-email",
	Password:             "password123",
	PasswordConfirmation: "password456",
	Framework:            "another-framework",
}

// Create a standard context. The default locale "en" will be used.
ctx := context.Background()
err := v.ValidateStruct(ctx, &invalidUser)

if err != nil {
	// The custom error type can be asserted to access the map of errors
	if validationErr, ok := err.(*validation.Error); ok {
		log.Printf("Validation failed with %d errors:", len(validationErr.Errors))
		for field, msg := range validationErr.Errors {
			log.Printf("  - Field: %s, Error: %s", field, msg)
		}
	}
}
/* Expected Output:
Validation failed with 4 errors:
  - Field: Username, Error: The 'Username' field must be at least 3 characters long.
  - Field: Email Address, Error: The value provided for 'Email Address' is not a valid email.
  - Field: Password Confirmation, Error: The 'Password Confirmation' field must match the Password field.
  - Field: Framework Name, Error: The framework must be 'dg-framework'!
*/

// Scenario B: Failing validation with Indonesian locale
fmt.Println("\n--- Scenario B: Failing Validation (Indonesian) ---")
// Create a context with the "id" locale. In a real application, this would typically
// be done in a middleware based on request headers (e.g., Accept-Language).
idCtx := validation.ToContext(context.Background(), "id")
err = v.ValidateStruct(idCtx, &invalidUser)

if err != nil {
	if validationErr, ok := err.(*validation.Error); ok {
		log.Printf("Validasi gagal dengan %d error:", len(validationErr.Errors))
		for field, msg := range validationErr.Errors {
			log.Printf("  - Kolom: %s, Error: %s", field, msg)
		}
	}
}
/* Expected Output:
Validasi gagal dengan 4 error:
  - Kolom: Username, Error: Kolom 'Username' minimal harus 3 karakter.
  - Kolom: Email Address, Error: Format email untuk 'Email Address' tidak valid.
  - Kolom: Password Confirmation, Error: Kolom 'Password Confirmation' harus sama dengan kolom Password.
  - Kolom: Framework Name, Error: Framework harus 'dg-framework'!
*/

// Scenario C: Successful validation
fmt.Println("\n--- Scenario C: Successful Validation ---")
validUser := UserRegistration{
	Username:             "donnigundala",
	Email:                "donni@example.com",
	Password:             "a-very-secure-password",
	PasswordConfirmation: "a-very-secure-password",
	Framework:            "dg-framework",
}

err = v.ValidateStruct(ctx, &validUser)
if err == nil {
	fmt.Println("User data is valid!")
}
/* Expected Output:
User data is valid!
*/
```

### 4. Helper Functions

-   **`validation.ToContext(ctx context.Context, locale string) context.Context`**: This helper function allows you to embed a locale string into a `context.Context`. This is useful in HTTP middleware to set the desired language for validation messages based on incoming requests.

## Error Handling

The `ValidateStruct` method returns an `error`. If validation fails, this error can be asserted to `*validation.Error` to access a map of field-specific error messages:

```go
if err != nil {
    if validationErr, ok := err.(*validation.Error); ok {
        // validationErr.Errors is a map[string]string where key is field name and value is the error message.
        for field, msg := range validationErr.Errors {
            fmt.Printf("Field '%s' has error: %s\n", field, msg)
        }
    } else {
        // Handle other types of errors (e.g., internal errors during validation setup)
        fmt.Printf("An unexpected error occurred: %v\n", err)
    }
}
```

This concludes the enhancements and documentation for the `validation` package. It is now truly production-ready.
