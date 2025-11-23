package middleware

import (
	"net/http"

	"github.com/donnigundala/dg-core/errors"
)

// BodySizeLimit returns a middleware that limits the request body size.
func BodySizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to methods that can have a body
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				// Limit the request body size
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// BodySizeLimitWithError returns a middleware that limits body size and returns a proper error.
func BodySizeLimitWithError(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to methods that can have a body
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				// Check Content-Length header first
				if r.ContentLength > maxBytes {
					err := errors.New("request body too large").
						WithCode("BODY_TOO_LARGE").
						WithStatus(http.StatusRequestEntityTooLarge).
						WithField("max_bytes", maxBytes).
						WithField("content_length", r.ContentLength)
					errors.WriteHTTPError(w, err)
					return
				}

				// Limit the request body size
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}

			next.ServeHTTP(w, r)
		})
	}
}
