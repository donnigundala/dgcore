package main

import (
	"context"
	"fmt"
	"log"

	"github.com/donnigundala/dgcore/validation"
	"github.com/go-playground/validator/v10"
)

// UserRegistration --- 1. Define Structs with Validation Tags ---
// We use a `display` tag for custom field names in error messages.
type UserRegistration struct {
	Username             string `validate:"required,min=3,max=20" json:"username" display:"Username"`
	Email                string `validate:"required,email" json:"email" display:"Email Address"`
	Password             string `validate:"required,min=8" json:"password" display:"Password"`
	PasswordConfirmation string `validate:"required,eqfield=Password" json:"password_confirmation" display:"Password Confirmation"`
	Framework            string `validate:"required,is_awesome" json:"framework" display:"Framework Name"`
}

// --- 2. Define Custom Validation Logic ---

// passwordsMatch is a custom struct-level validation function.
// It's commented out here because the `eqfield` tag already handles this.
// This is just to show how you would register a struct-level rule.
func passwordsMatch(sl validator.StructLevel) {
	user := sl.Current().Interface().(UserRegistration)
	if user.Password != user.PasswordConfirmation {
		// ReportError(field, fieldName, structFieldName, tag, param)
		sl.ReportError(user.PasswordConfirmation, "PasswordConfirmation", "password_confirmation", "passwords_match", "")
	}
}

// isAwesome is a custom field-level validation function.
func isAwesome(fl validator.FieldLevel) bool {
	return fl.Field().String() == "dg-framework"
}

func main() {
	fmt.Println("--- Professional Grade Validator Example ---")

	// --- 3. Define Custom Messages for Different Locales ---
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

	// --- 4. Initialize the Validator with All Features ---
	v := validation.New(
		// Use the "display" tag for field names in error messages
		validation.WithFieldNameTag("display"),
		// Set the default language for messages
		validation.WithDefaultLocale("en"),
		// Register messages for English
		validation.WithLocaleMessages("en", enMessages),
		// Register messages for Indonesian
		validation.WithLocaleMessages("id", idMessages),
		// Register a custom field-level validation
		validation.WithCustomValidation("is_awesome", isAwesome),
		// Register a struct-level validation (example)
		// validation.WithStructValidation(passwordsMatch, UserRegistration{}),
	)

	// --- 5. Run Validation Scenarios ---

	// Scenario A: Failing validation with default locale (English)
	fmt.Println("\n--- Scenario A: Failing Validation (English) ---")
	invalidUser := UserRegistration{
		Username:             "dj",
		Email:                "invalid-email",
		Password:             "password123",
		PasswordConfirmation: "password456",
		Framework:            "another-framework",
	}

	// Create a standard context
	ctx := context.Background()
	err := v.ValidateStruct(ctx, &invalidUser)

	if err != nil {
		// The custom error type can be asserted to access the map
		if validationErr, ok := err.(*validation.Error); ok {
			log.Printf("Validation failed with %d errors:", len(validationErr.Errors))
			for field, msg := range validationErr.Errors {
				log.Printf("  - Field: %s, Error: %s", field, msg)
			}
		}
	}

	// Scenario B: Failing validation with Indonesian locale
	fmt.Println("\n--- Scenario B: Failing Validation (Indonesian) ---")
	// Create a context with the "id" locale. In a real app, this would be done in a middleware.
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
}
