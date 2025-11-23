package request

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Param returns a path parameter value.
// This works with Gin router by extracting from context.
func Param(r *http.Request, key string) string {
	// Try to get Gin context
	if ginCtx, ok := r.Context().Value("gin").(gin.Context); ok {
		return ginCtx.Param(key)
	}

	// Fallback: try to get from context directly
	if val := r.Context().Value(key); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}

	return ""
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
