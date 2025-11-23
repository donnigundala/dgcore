package response

import (
	"net/http"

	"github.com/donnigundala/dg-core/errors"
)

// ErrorResponse represents an error API response.
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error"`
	Code    string      `json:"code,omitempty"`
	Fields  interface{} `json:"fields,omitempty"`
}

// Error writes an error response.
func Error(w http.ResponseWriter, err error, status int) {
	// Check if it's our custom Error type
	if e, ok := err.(*errors.Error); ok {
		JSON(w, e.HTTPStatus(), ErrorResponse{
			Success: false,
			Error:   e.Message(),
			Code:    e.Code(),
			Fields:  e.Fields(),
		})
		return
	}

	// Default error response
	JSON(w, status, ErrorResponse{
		Success: false,
		Error:   err.Error(),
	})
}

// ValidationError writes a validation error response.
func ValidationError(w http.ResponseWriter, validationErrors interface{}) {
	JSON(w, http.StatusUnprocessableEntity, ErrorResponse{
		Success: false,
		Error:   "Validation failed",
		Code:    "VALIDATION_ERROR",
		Fields:  validationErrors,
	})
}

// NotFound writes a 404 Not Found response.
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	JSON(w, http.StatusNotFound, ErrorResponse{
		Success: false,
		Error:   message,
		Code:    "NOT_FOUND",
	})
}

// Unauthorized writes a 401 Unauthorized response.
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	JSON(w, http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error:   message,
		Code:    "UNAUTHORIZED",
	})
}

// Forbidden writes a 403 Forbidden response.
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Forbidden"
	}
	JSON(w, http.StatusForbidden, ErrorResponse{
		Success: false,
		Error:   message,
		Code:    "FORBIDDEN",
	})
}

// BadRequest writes a 400 Bad Request response.
func BadRequest(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Bad request"
	}
	JSON(w, http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error:   message,
		Code:    "BAD_REQUEST",
	})
}

// InternalServerError writes a 500 Internal Server Error response.
func InternalServerError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Internal server error"
	}
	JSON(w, http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error:   message,
		Code:    "INTERNAL_ERROR",
	})
}
