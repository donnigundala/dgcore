package validation

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// RegisterCustomValidators registers all built-in custom validators.
// This is called automatically when creating a new validator.
func registerCustomValidators(v *validator.Validate) {
	// UUID validation
	v.RegisterValidation("uuid", validateUUID)

	// Slug validation (URL-friendly)
	v.RegisterValidation("slug", validateSlug)

	// Phone number validation (basic)
	v.RegisterValidation("phone", validatePhone)

	// Password strength validation
	v.RegisterValidation("password", validatePassword)

	// Username validation
	v.RegisterValidation("username", validateUsername)

	// Alpha with spaces
	v.RegisterValidation("alpha_space", validateAlphaSpace)

	// SQL injection prevention (basic)
	v.RegisterValidation("no_sql", validateNoSQL)

	// XSS prevention (basic)
	v.RegisterValidation("no_xss", validateNoXSS)

	// Hex color validation
	v.RegisterValidation("color_hex", validateColorHex)

	// Timezone validation
	v.RegisterValidation("timezone", validateTimezone)
}

// validateUUID checks if the field is a valid UUID.
// Usage: `validate:"uuid"`
func validateUUID(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

// validateSlug checks if the field is a valid URL slug.
// Allows: lowercase letters, numbers, hyphens
// Usage: `validate:"slug"`
func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	if slug == "" {
		return false
	}

	// Slug pattern: lowercase letters, numbers, hyphens
	// Must start and end with alphanumeric
	slugPattern := `^[a-z0-9]+(?:-[a-z0-9]+)*$`
	matched, _ := regexp.MatchString(slugPattern, slug)
	return matched
}

// validatePhone checks if the field is a valid phone number.
// Accepts: +1234567890, (123) 456-7890, 123-456-7890, etc.
// Usage: `validate:"phone"`
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return false
	}

	// Remove common phone number characters
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || r == '+' {
			return r
		}
		return -1
	}, phone)

	// Must have 10-15 digits (international format)
	length := len(cleaned)
	if strings.HasPrefix(cleaned, "+") {
		return length >= 11 && length <= 16 // +1234567890 to +123456789012345
	}
	return length >= 10 && length <= 15
}

// validatePassword checks password strength.
// Requirements: min 8 chars, at least 1 uppercase, 1 lowercase, 1 number
// Usage: `validate:"password"`
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper  = false
		hasLower  = false
		hasNumber = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		}
	}

	return hasUpper && hasLower && hasNumber
}

// validateUsername checks if the field is a valid username.
// Allows: letters, numbers, underscores, hyphens (3-20 chars)
// Usage: `validate:"username"`
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 20 {
		return false
	}

	// Username pattern: letters, numbers, underscores, hyphens
	usernamePattern := `^[a-zA-Z0-9_-]+$`
	matched, _ := regexp.MatchString(usernamePattern, username)
	return matched
}

// validateAlphaSpace allows letters and spaces only.
// Usage: `validate:"alpha_space"`
func validateAlphaSpace(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	for _, char := range value {
		if !unicode.IsLetter(char) && !unicode.IsSpace(char) {
			return false
		}
	}
	return true
}

// validateNoSQL performs basic SQL injection prevention.
// Rejects common SQL keywords and special characters.
// Usage: `validate:"no_sql"`
func validateNoSQL(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())

	// Common SQL injection patterns
	sqlKeywords := []string{
		"select", "insert", "update", "delete", "drop", "create",
		"alter", "exec", "execute", "union", "declare", "--", "/*", "*/",
		"xp_", "sp_", "0x", "char(", "nchar(", "varchar(", "nvarchar(",
	}

	for _, keyword := range sqlKeywords {
		if strings.Contains(value, keyword) {
			return false
		}
	}

	return true
}

// validateNoXSS performs basic XSS prevention.
// Rejects common XSS patterns.
// Usage: `validate:"no_xss"`
func validateNoXSS(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())

	// Common XSS patterns
	xssPatterns := []string{
		"<script", "</script", "javascript:", "onerror=", "onload=",
		"onclick=", "onmouseover=", "<iframe", "<object", "<embed",
		"eval(", "expression(", "vbscript:", "data:text/html",
	}

	for _, pattern := range xssPatterns {
		if strings.Contains(value, pattern) {
			return false
		}
	}

	return true
}

// validateColorHex checks if the field is a valid hex color code.
// Accepts: #FFF, #FFFFFF, #fff, #ffffff
// Usage: `validate:"color_hex"`
func validateColorHex(fl validator.FieldLevel) bool {
	color := fl.Field().String()
	hexPattern := `^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`
	matched, _ := regexp.MatchString(hexPattern, color)
	return matched
}

// validateTimezone checks if the field is a valid timezone string.
// Accepts: UTC, America/New_York, Asia/Tokyo, etc.
// Usage: `validate:"timezone"`
func validateTimezone(fl validator.FieldLevel) bool {
	tz := fl.Field().String()

	// Common timezone patterns
	// This is a basic check - for production, use time.LoadLocation
	if tz == "UTC" || tz == "GMT" {
		return true
	}

	// Format: Continent/City
	tzPattern := `^[A-Z][a-z]+/[A-Z][a-z_]+$`
	matched, _ := regexp.MatchString(tzPattern, tz)
	return matched
}
