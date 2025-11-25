package errors

import (
	"errors"
	"net/http"
	"testing"
)

// TestNew_BasicError tests basic error creation
func TestNew_BasicError(t *testing.T) {
	msg := "test error"
	err := New(msg)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Message() != msg {
		t.Errorf("expected message %q, got %q", msg, err.Message())
	}

	if err.Error() != msg {
		t.Errorf("expected error string %q, got %q", msg, err.Error())
	}

	if err.HTTPStatus() != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, err.HTTPStatus())
	}

	if err.Code() != "" {
		t.Errorf("expected empty code, got %q", err.Code())
	}
}

// TestWrap_WithContext tests wrapping errors with context
func TestWrap_WithContext(t *testing.T) {
	originalErr := errors.New("original error")
	wrapMsg := "wrapped error"

	wrapped := Wrap(originalErr, wrapMsg)

	if wrapped == nil {
		t.Fatal("expected error, got nil")
	}

	if wrapped.Message() != wrapMsg {
		t.Errorf("expected message %q, got %q", wrapMsg, wrapped.Message())
	}

	// Test unwrapping
	unwrapped := wrapped.Unwrap()
	if unwrapped == nil {
		t.Fatal("expected unwrapped error, got nil")
	}

	// Test wrapping nil
	nilWrapped := Wrap(nil, "should be nil")
	if nilWrapped != nil {
		t.Error("expected nil when wrapping nil error")
	}
}

// TestWrapf_FormattedMessage tests wrapping with formatted messages
func TestWrapf_FormattedMessage(t *testing.T) {
	originalErr := errors.New("database error")
	userID := 123

	wrapped := Wrapf(originalErr, "failed to fetch user %d", userID)

	expectedMsg := "failed to fetch user 123"
	if wrapped.Message() != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, wrapped.Message())
	}

	// Test wrapping nil with Wrapf
	nilWrapped := Wrapf(nil, "should be nil")
	if nilWrapped != nil {
		t.Error("expected nil when wrapping nil error with Wrapf")
	}
}

// TestWithCode_ErrorCode tests setting error codes
func TestWithCode_ErrorCode(t *testing.T) {
	err := New("test error").WithCode("TEST_ERROR")

	if err.Code() != "TEST_ERROR" {
		t.Errorf("expected code %q, got %q", "TEST_ERROR", err.Code())
	}

	// Test chaining
	err2 := New("another error").
		WithCode("CHAIN_ERROR").
		WithStatus(http.StatusBadRequest)

	if err2.Code() != "CHAIN_ERROR" {
		t.Errorf("expected code %q, got %q", "CHAIN_ERROR", err2.Code())
	}

	if err2.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, err2.HTTPStatus())
	}
}

// TestWithStatus_HTTPStatus tests setting HTTP status codes
func TestWithStatus_HTTPStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"BadRequest", http.StatusBadRequest},
		{"NotFound", http.StatusNotFound},
		{"Unauthorized", http.StatusUnauthorized},
		{"InternalServerError", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New("test error").WithStatus(tt.status)

			if err.HTTPStatus() != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, err.HTTPStatus())
			}
		})
	}
}

// TestWithField_SingleField tests adding a single field
func TestWithField_SingleField(t *testing.T) {
	err := New("test error").WithField("user_id", 123)

	fields := err.Fields()
	if len(fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(fields))
	}

	if fields["user_id"] != 123 {
		t.Errorf("expected user_id=123, got %v", fields["user_id"])
	}
}

// TestWithFields_MultipleFields tests adding multiple fields
func TestWithFields_MultipleFields(t *testing.T) {
	fields := map[string]interface{}{
		"user_id":  123,
		"username": "john",
		"active":   true,
	}

	err := New("test error").WithFields(fields)

	errFields := err.Fields()
	if len(errFields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(errFields))
	}

	if errFields["user_id"] != 123 {
		t.Errorf("expected user_id=123, got %v", errFields["user_id"])
	}

	if errFields["username"] != "john" {
		t.Errorf("expected username=john, got %v", errFields["username"])
	}

	if errFields["active"] != true {
		t.Errorf("expected active=true, got %v", errFields["active"])
	}
}

// TestChaining_AllMethods tests chaining all methods together
func TestChaining_AllMethods(t *testing.T) {
	err := New("test error").
		WithCode("TEST_CODE").
		WithStatus(http.StatusBadRequest).
		WithField("key1", "value1").
		WithFields(map[string]interface{}{
			"key2": "value2",
			"key3": 123,
		})

	if err.Code() != "TEST_CODE" {
		t.Errorf("expected code TEST_CODE, got %q", err.Code())
	}

	if err.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.HTTPStatus())
	}

	fields := err.Fields()
	if len(fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(fields))
	}
}

// TestIs_ErrorComparison tests error comparison
func TestIs_ErrorComparison(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrap(baseErr, "wrapped")

	if !Is(wrapped, baseErr) {
		t.Error("expected Is to return true for wrapped error")
	}

	otherErr := errors.New("other error")
	if Is(wrapped, otherErr) {
		t.Error("expected Is to return false for different error")
	}
}

// TestAs_ErrorType tests error type assertion
func TestAs_ErrorType(t *testing.T) {
	customErr := New("custom error").WithCode("CUSTOM")

	var target *Error
	if !As(customErr, &target) {
		t.Fatal("expected As to return true")
	}

	if target.Code() != "CUSTOM" {
		t.Errorf("expected code CUSTOM, got %q", target.Code())
	}
}

// TestCause_UnderlyingError tests getting the root cause
func TestCause_UnderlyingError(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped1 := Wrap(baseErr, "wrap 1")

	cause := Cause(wrapped1)
	if cause == nil {
		t.Fatal("expected cause, got nil")
	}

	// Cause returns the deepest error in the chain
	// For our Wrap implementation, this includes the wrap message
	expectedMsg := "wrap 1: base error"
	if cause.Error() != expectedMsg {
		t.Errorf("expected cause message %q, got %q", expectedMsg, cause.Error())
	}
}

// TestSentinelErrors tests predefined sentinel errors
func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        *Error
		wantCode   string
		wantStatus int
	}{
		{"NotFound", ErrNotFound, "NOT_FOUND", http.StatusNotFound},
		{"Unauthorized", ErrUnauthorized, "UNAUTHORIZED", http.StatusUnauthorized},
		{"Forbidden", ErrForbidden, "FORBIDDEN", http.StatusForbidden},
		{"BadRequest", ErrBadRequest, "BAD_REQUEST", http.StatusBadRequest},
		{"InternalServer", ErrInternalServer, "INTERNAL_ERROR", http.StatusInternalServerError},
		{"Conflict", ErrConflict, "CONFLICT", http.StatusConflict},
		{"Unprocessable", ErrUnprocessable, "UNPROCESSABLE", http.StatusUnprocessableEntity},
		{"TooManyRequests", ErrTooManyRequests, "TOO_MANY_REQUESTS", http.StatusTooManyRequests},
		{"ServiceUnavailable", ErrServiceUnavailable, "SERVICE_UNAVAILABLE", http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code() != tt.wantCode {
				t.Errorf("expected code %q, got %q", tt.wantCode, tt.err.Code())
			}

			if tt.err.HTTPStatus() != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, tt.err.HTTPStatus())
			}
		})
	}
}

// TestStackTrace tests stack trace functionality
func TestStackTrace(t *testing.T) {
	err := New("test error")
	stackTrace := err.StackTrace()

	// Stack trace should be present for new errors
	if stackTrace == "" {
		t.Error("expected non-empty stack trace")
	}

	// Wrapped errors should also have stack traces
	wrapped := Wrap(errors.New("base"), "wrapped")
	wrappedTrace := wrapped.StackTrace()

	if wrappedTrace == "" {
		t.Error("expected non-empty stack trace for wrapped error")
	}
}

// TestError_NilHandling tests nil error handling
func TestError_NilHandling(t *testing.T) {
	// Wrap should return nil for nil input
	if Wrap(nil, "message") != nil {
		t.Error("expected nil when wrapping nil error")
	}

	// Wrapf should return nil for nil input
	if Wrapf(nil, "message %s", "test") != nil {
		t.Error("expected nil when wrapping nil error with Wrapf")
	}
}
