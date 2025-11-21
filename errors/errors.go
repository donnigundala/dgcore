package errors

import (
	"fmt"
	"net/http"

	pkgErrors "github.com/pkg/errors"
)

// Error represents an enhanced error with additional context.
type Error struct {
	err        error
	message    string
	code       string
	httpStatus int
	fields     map[string]interface{}
}

// New creates a new Error.
func New(message string) *Error {
	return &Error{
		err:        pkgErrors.New(message),
		message:    message,
		httpStatus: http.StatusInternalServerError,
		fields:     make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context.
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		err:        pkgErrors.Wrap(err, message),
		message:    message,
		httpStatus: http.StatusInternalServerError,
		fields:     make(map[string]interface{}),
	}
}

// Wrapf wraps an error with a formatted message.
func Wrapf(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf(format, args...)
	return &Error{
		err:        pkgErrors.Wrap(err, message),
		message:    message,
		httpStatus: http.StatusInternalServerError,
		fields:     make(map[string]interface{}),
	}
}

// WithCode sets the error code.
func (e *Error) WithCode(code string) *Error {
	e.code = code
	return e
}

// WithStatus sets the HTTP status code.
func (e *Error) WithStatus(status int) *Error {
	e.httpStatus = status
	return e
}

// WithField adds a field to the error.
func (e *Error) WithField(key string, value interface{}) *Error {
	e.fields[key] = value
	return e
}

// WithFields adds multiple fields to the error.
func (e *Error) WithFields(fields map[string]interface{}) *Error {
	for k, v := range fields {
		e.fields[k] = v
	}
	return e
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return e.message
}

// Unwrap returns the underlying error for error chain unwrapping.
func (e *Error) Unwrap() error {
	return e.err
}

// Code returns the error code.
func (e *Error) Code() string {
	return e.code
}

// HTTPStatus returns the HTTP status code.
func (e *Error) HTTPStatus() int {
	return e.httpStatus
}

// Fields returns the error fields.
func (e *Error) Fields() map[string]interface{} {
	return e.fields
}

// Message returns the error message.
func (e *Error) Message() string {
	return e.message
}

// StackTrace returns the stack trace if available.
func (e *Error) StackTrace() string {
	type stackTracer interface {
		StackTrace() pkgErrors.StackTrace
	}

	if st, ok := e.err.(stackTracer); ok {
		return fmt.Sprintf("%+v", st.StackTrace())
	}
	return ""
}

// Is checks if the error matches the target error.
func Is(err, target error) bool {
	return pkgErrors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
func As(err error, target interface{}) bool {
	return pkgErrors.As(err, target)
}

// Cause returns the underlying cause of the error.
func Cause(err error) error {
	return pkgErrors.Cause(err)
}

// Common sentinel errors
var (
	ErrNotFound           = New("resource not found").WithStatus(http.StatusNotFound).WithCode("NOT_FOUND")
	ErrUnauthorized       = New("unauthorized").WithStatus(http.StatusUnauthorized).WithCode("UNAUTHORIZED")
	ErrForbidden          = New("forbidden").WithStatus(http.StatusForbidden).WithCode("FORBIDDEN")
	ErrBadRequest         = New("bad request").WithStatus(http.StatusBadRequest).WithCode("BAD_REQUEST")
	ErrInternalServer     = New("internal server error").WithStatus(http.StatusInternalServerError).WithCode("INTERNAL_ERROR")
	ErrConflict           = New("conflict").WithStatus(http.StatusConflict).WithCode("CONFLICT")
	ErrUnprocessable      = New("unprocessable entity").WithStatus(http.StatusUnprocessableEntity).WithCode("UNPROCESSABLE")
	ErrTooManyRequests    = New("too many requests").WithStatus(http.StatusTooManyRequests).WithCode("TOO_MANY_REQUESTS")
	ErrServiceUnavailable = New("service unavailable").WithStatus(http.StatusServiceUnavailable).WithCode("SERVICE_UNAVAILABLE")
)
