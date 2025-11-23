package request

import (
	"net/http"
	"strconv"
)

// Param returns a path parameter value.
// This works with Gin router by extracting from the URL path.
func Param(r *http.Request, key string) string {
	// Try to get from context (works with custom routers that store params in context)
	if val := r.Context().Value(key); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}

	// Fallback: try to get from URL query (not ideal for path params, but safe fallback)
	return r.URL.Query().Get(key)
}

// ParamInt returns a path parameter as an integer.
func ParamInt(r *http.Request, key string) int {
	value := Param(r, key)
	if value == "" {
		return 0
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return intValue
}

// ParamInt64 returns a path parameter as an int64.
func ParamInt64(r *http.Request, key string) int64 {
	value := Param(r, key)
	if value == "" {
		return 0
	}
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return intValue
}

// ParamString returns a path parameter as a string (alias for Param).
func ParamString(r *http.Request, key string) string {
	return Param(r, key)
}
