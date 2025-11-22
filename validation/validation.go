package validation

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// contextKey is a private type to avoid context key collisions.
type contextKey string

const localeKey contextKey = "locale"

// Validator is the concrete implementation of IValidator.
type Validator struct {
	validate      *validator.Validate
	messages      map[string]map[string]string // map[locale]map[tag]message
	defaultLocale string
	fieldNameTag  string
}

// IValidator defines the interface for the validation service.
type IValidator interface {
	ValidateStruct(ctx context.Context, i interface{}) error
}

// Option configures the Validator.
type Option func(*Validator)

// defaultMessages returns a map of default validation messages for the "en" locale.
func defaultMessages() map[string]string {
	return map[string]string{
		"required": "is required",
		"email":    "is not a valid email format",
		"gte":      "must be greater than or equal to %s",
	}
}

// --- Public Options for Configuration ---

// WithCustomValidation allows registering a custom validation function.
func WithCustomValidation(tag string, fn validator.Func) Option {
	return func(v *Validator) {
		if err := v.validate.RegisterValidation(tag, fn); err != nil {
			panic(fmt.Sprintf("failed to register custom validation '%s': %v", tag, err))
		}
	}
}

// WithStructValidation registers a function for struct-level validation.
func WithStructValidation(fn validator.StructLevelFunc, types ...interface{}) Option {
	return func(v *Validator) {
		v.validate.RegisterStructValidation(fn, types...)
	}
}

// WithLocaleMessages registers or overrides messages for a specific locale.
func WithLocaleMessages(locale string, messages map[string]string) Option {
	return func(v *Validator) {
		if _, ok := v.messages[locale]; !ok {
			v.messages[locale] = make(map[string]string)
		}
		for key, msg := range messages {
			v.messages[locale][key] = msg
		}
	}
}

// WithDefaultLocale sets the fallback locale if the requested locale is not found.
func WithDefaultLocale(locale string) Option {
	return func(v *Validator) {
		v.defaultLocale = locale
	}
}

// WithFieldNameTag sets the struct tag to use for extracting field names in error messages.
func WithFieldNameTag(tag string) Option {
	return func(v *Validator) {
		v.fieldNameTag = tag
	}
}

// NewValidator creates a new, extensible, and production-grade validator instance.
func NewValidator(opts ...Option) IValidator {
	v := &Validator{
		validate:      validator.New(),
		messages:      make(map[string]map[string]string),
		defaultLocale: "en",   // Sensible default
		fieldNameTag:  "json", // Sensible default
	}

	// Load the framework's default messages for the default locale.
	v.messages[v.defaultLocale] = defaultMessages()

	// Apply all functional options passed by the consumer.
	for _, opt := range opts {
		opt(v)
	}

	return v
}

// --- Core Methods ---

// ValidateStruct performs validation on a struct, aware of context for localization.
func (v *Validator) ValidateStruct(ctx context.Context, i interface{}) error {
	err := v.validate.Struct(i)
	if err != nil {
		return v.parseErrors(ctx, err, i)
	}
	return nil
}

// parseErrors processes validation errors and returns a structured, localized error.
func (v *Validator) parseErrors(ctx context.Context, err error, obj interface{}) error {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return err
	}

	// 1. Determine the locale for this request.
	currentLocale := v.getLocale(ctx)
	messageSet, ok := v.messages[currentLocale]
	if !ok {
		// Fallback to the default locale if the requested one isn't registered.
		messageSet = v.messages[v.defaultLocale]
	}

	customErrors := make(map[string]string)
	objType := reflect.TypeOf(obj).Elem()

	for _, e := range validationErrors { // e is a validator.FieldError
		// 2. Resolve the field name using the configured tag.
		field, _ := objType.FieldByName(e.Field())
		fieldName := v.resolveFieldName(field)

		// 3. Find the message template.
		msgTemplate, ok := messageSet[e.Tag()]

		var finalMessage string
		if ok {
			// A custom or framework-default message was found. Use it.
			finalMessage = fmt.Sprintf(msgTemplate, fieldName, e.Param())
		} else {
			// No message was found in our maps. Fallback to the library's built-in default.
			finalMessage = e.Error()
		}
		customErrors[fieldName] = finalMessage
	}

	return &Error{Errors: customErrors}
}

// --- Helper Methods ---

// getLocale extracts the locale string from the context.
func (v *Validator) getLocale(ctx context.Context) string {
	if locale, ok := ctx.Value(localeKey).(string); ok && locale != "" {
		return locale
	}
	return v.defaultLocale
}

// resolveFieldName finds the best possible name for a field based on configuration.
func (v *Validator) resolveFieldName(field reflect.StructField) string {
	// 1. Try the custom configured tag first.
	name := field.Tag.Get(v.fieldNameTag)
	if name != "" && name != "-" {
		return name
	}
	// 2. Fallback to "json" tag.
	name = field.Tag.Get("json")
	if name != "" && name != "-" {
		return name
	}
	// 3. Fallback to the actual struct field name.
	return field.Name
}

// ToContext is a helper function to embed a locale within a context.
// This would typically be used in a middleware in the consumer application.
func ToContext(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeKey, locale)
}
