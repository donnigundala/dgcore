package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestToHTTPError_CustomError tests converting custom Error to HTTPError
func TestToHTTPError_CustomError(t *testing.T) {
	err := New("test error").
		WithCode("TEST_ERROR").
		WithStatus(http.StatusBadRequest).
		WithField("user_id", 123)

	status, httpErr := ToHTTPError(err)

	if status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
	}

	if httpErr.Code != "TEST_ERROR" {
		t.Errorf("expected code %q, got %q", "TEST_ERROR", httpErr.Code)
	}

	if httpErr.Message != "test error" {
		t.Errorf("expected message %q, got %q", "test error", httpErr.Message)
	}

	if len(httpErr.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(httpErr.Fields))
	}

	if httpErr.Fields["user_id"] != 123 {
		t.Errorf("expected user_id=123, got %v", httpErr.Fields["user_id"])
	}
}

// TestToHTTPError_StandardError tests converting standard error to HTTPError
func TestToHTTPError_StandardError(t *testing.T) {
	err := errors.New("standard error")

	status, httpErr := ToHTTPError(err)

	if status != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, status)
	}

	if httpErr.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code %q, got %q", "INTERNAL_ERROR", httpErr.Code)
	}

	if httpErr.Message != "standard error" {
		t.Errorf("expected message %q, got %q", "standard error", httpErr.Message)
	}
}

// TestToHTTPError_NilError tests converting nil error
func TestToHTTPError_NilError(t *testing.T) {
	status, httpErr := ToHTTPError(nil)

	if status != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, status)
	}

	if httpErr != nil {
		t.Errorf("expected nil HTTPError, got %v", httpErr)
	}
}

// TestWriteHTTPError_Response tests writing error to HTTP response
func TestWriteHTTPError_Response(t *testing.T) {
	err := New("test error").
		WithCode("TEST_ERROR").
		WithStatus(http.StatusBadRequest)

	w := httptest.NewRecorder()
	WriteHTTPError(w, err)

	// Check status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content type %q, got %q", "application/json", contentType)
	}

	// Check response body
	var httpErr HTTPError
	if err := json.NewDecoder(w.Body).Decode(&httpErr); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if httpErr.Code != "TEST_ERROR" {
		t.Errorf("expected code %q, got %q", "TEST_ERROR", httpErr.Code)
	}

	if httpErr.Message != "test error" {
		t.Errorf("expected message %q, got %q", "test error", httpErr.Message)
	}
}

// TestWriteHTTPError_StandardError tests writing standard error
func TestWriteHTTPError_StandardError(t *testing.T) {
	err := errors.New("standard error")

	w := httptest.NewRecorder()
	WriteHTTPError(w, err)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var httpErr HTTPError
	if err := json.NewDecoder(w.Body).Decode(&httpErr); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if httpErr.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code %q, got %q", "INTERNAL_ERROR", httpErr.Code)
	}
}

// TestWriteHTTPError_WithFields tests writing error with fields
func TestWriteHTTPError_WithFields(t *testing.T) {
	err := New("validation error").
		WithCode("VALIDATION_ERROR").
		WithStatus(http.StatusUnprocessableEntity).
		WithFields(map[string]interface{}{
			"field":  "email",
			"reason": "invalid format",
		})

	w := httptest.NewRecorder()
	WriteHTTPError(w, err)

	var httpErr HTTPError
	if err := json.NewDecoder(w.Body).Decode(&httpErr); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(httpErr.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(httpErr.Fields))
	}

	if httpErr.Fields["field"] != "email" {
		t.Errorf("expected field=email, got %v", httpErr.Fields["field"])
	}

	if httpErr.Fields["reason"] != "invalid format" {
		t.Errorf("expected reason='invalid format', got %v", httpErr.Fields["reason"])
	}
}

// TestHTTPStatusFromError_CustomError tests extracting status from custom error
func TestHTTPStatusFromError_CustomError(t *testing.T) {
	tests := []struct {
		name       string
		err        *Error
		wantStatus int
	}{
		{"NotFound", ErrNotFound, http.StatusNotFound},
		{"BadRequest", ErrBadRequest, http.StatusBadRequest},
		{"Unauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"Custom", New("test").WithStatus(http.StatusTeapot), http.StatusTeapot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := HTTPStatusFromError(tt.err)

			if status != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, status)
			}
		})
	}
}

// TestHTTPStatusFromError_StandardError tests extracting status from standard error
func TestHTTPStatusFromError_StandardError(t *testing.T) {
	err := errors.New("standard error")

	status := HTTPStatusFromError(err)

	if status != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, status)
	}
}

// TestHTTPError_JSONSerialization tests JSON serialization of HTTPError
func TestHTTPError_JSONSerialization(t *testing.T) {
	httpErr := &HTTPError{
		Code:    "TEST_ERROR",
		Message: "test message",
		Fields: map[string]interface{}{
			"key": "value",
		},
	}

	data, err := json.Marshal(httpErr)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded HTTPError
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Code != httpErr.Code {
		t.Errorf("expected code %q, got %q", httpErr.Code, decoded.Code)
	}

	if decoded.Message != httpErr.Message {
		t.Errorf("expected message %q, got %q", httpErr.Message, decoded.Message)
	}

	if decoded.Fields["key"] != "value" {
		t.Errorf("expected key=value, got %v", decoded.Fields["key"])
	}
}

// TestHTTPError_EmptyFields tests HTTPError with no fields
func TestHTTPError_EmptyFields(t *testing.T) {
	err := New("test error").WithCode("TEST")

	_, httpErr := ToHTTPError(err)

	data, marshalErr := json.Marshal(httpErr)
	if marshalErr != nil {
		t.Fatalf("failed to marshal: %v", marshalErr)
	}

	// Empty fields should be omitted in JSON
	var result map[string]interface{}
	if unmarshalErr := json.Unmarshal(data, &result); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal: %v", unmarshalErr)
	}

	// Fields should not be present if empty
	if _, exists := result["fields"]; exists {
		fields := result["fields"].(map[string]interface{})
		if len(fields) > 0 {
			t.Error("expected fields to be omitted when empty")
		}
	}
}
