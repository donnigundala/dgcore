package main

import (
	"context"
	"fmt"

	"github.com/donnigundala/dgcore/validation"
)

// UserRegistration demonstrates custom validators
type UserRegistration struct {
	Username string `json:"username" validate:"required,username"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
	Phone    string `json:"phone" validate:"required,phone"`
}

// ProfileUpdate demonstrates more custom validators
type ProfileUpdate struct {
	FullName string `json:"full_name" validate:"required,alpha_space"`
	Bio      string `json:"bio" validate:"no_xss,no_sql"`
	Website  string `json:"website" validate:"url"`
}

// Settings demonstrates additional validators
type Settings struct {
	ThemeColor string `json:"theme_color" validate:"required,color_hex"`
	Timezone   string `json:"timezone" validate:"required,timezone"`
	Slug       string `json:"slug" validate:"required,slug"`
	UUID       string `json:"uuid" validate:"required,uuid"`
}

func main() {
	fmt.Println("=== Custom Validators Example ===\n")

	validator := validation.NewValidator()
	ctx := context.Background()

	// Example 1: Valid user registration
	fmt.Println("1. Valid User Registration:")
	validUser := UserRegistration{
		Username: "john_doe",
		Email:    "john@example.com",
		Password: "SecurePass123",
		Phone:    "+1234567890",
	}
	if err := validator.ValidateStruct(ctx, &validUser); err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Valid!")
	}

	// Example 2: Invalid username
	fmt.Println("\n2. Invalid Username (too short):")
	invalidUser := UserRegistration{
		Username: "ab", // Too short
		Email:    "test@example.com",
		Password: "SecurePass123",
		Phone:    "+1234567890",
	}
	if err := validator.ValidateStruct(ctx, &invalidUser); err != nil {
		if valErr, ok := err.(*validation.Error); ok {
			fmt.Printf("   ❌ Errors: %v\n", valErr.Errors)
		}
	}

	// Example 3: Weak password
	fmt.Println("\n3. Weak Password:")
	weakPassword := UserRegistration{
		Username: "john_doe",
		Email:    "test@example.com",
		Password: "weak", // No uppercase, no number
		Phone:    "+1234567890",
	}
	if err := validator.ValidateStruct(ctx, &weakPassword); err != nil {
		if valErr, ok := err.(*validation.Error); ok {
			fmt.Printf("   ❌ Errors: %v\n", valErr.Errors)
		}
	}

	// Example 4: Invalid phone
	fmt.Println("\n4. Invalid Phone:")
	invalidPhone := UserRegistration{
		Username: "john_doe",
		Email:    "test@example.com",
		Password: "SecurePass123",
		Phone:    "123", // Too short
	}
	if err := validator.ValidateStruct(ctx, &invalidPhone); err != nil {
		if valErr, ok := err.(*validation.Error); ok {
			fmt.Printf("   ❌ Errors: %v\n", valErr.Errors)
		}
	}

	// Example 5: Valid settings
	fmt.Println("\n5. Valid Settings:")
	validSettings := Settings{
		ThemeColor: "#FF5733",
		Timezone:   "America/New_York",
		Slug:       "my-awesome-post",
		UUID:       "550e8400-e29b-41d4-a716-446655440000",
	}
	if err := validator.ValidateStruct(ctx, &validSettings); err != nil {
		fmt.Printf("   ❌ Validation failed: %v\n", err)
	} else {
		fmt.Println("   ✅ Valid!")
	}

	// Example 6: Invalid slug
	fmt.Println("\n6. Invalid Slug (uppercase not allowed):")
	invalidSlug := Settings{
		ThemeColor: "#FF5733",
		Timezone:   "UTC",
		Slug:       "My-Post", // Uppercase not allowed
		UUID:       "550e8400-e29b-41d4-a716-446655440000",
	}
	if err := validator.ValidateStruct(ctx, &invalidSlug); err != nil {
		if valErr, ok := err.(*validation.Error); ok {
			fmt.Printf("   ❌ Errors: %v\n", valErr.Errors)
		}
	}

	// Example 7: XSS attempt
	fmt.Println("\n7. XSS Prevention:")
	xssAttempt := ProfileUpdate{
		FullName: "John Doe",
		Bio:      "<script>alert('xss')</script>", // XSS attempt
		Website:  "https://example.com",
	}
	if err := validator.ValidateStruct(ctx, &xssAttempt); err != nil {
		if valErr, ok := err.(*validation.Error); ok {
			fmt.Printf("   ❌ Errors: %v\n", valErr.Errors)
		}
	}

	fmt.Println("\n=== Example Complete ===")
}
