package request

import (
	"net/http"
	"strconv"
)

// Query returns a query parameter value with a default.
func Query(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// QueryInt returns a query parameter as an integer with a default.
func QueryInt(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// QueryInt64 returns a query parameter as an int64 with a default.
func QueryInt64(r *http.Request, key string, defaultValue int64) int64 {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// QueryBool returns a query parameter as a boolean with a default.
func QueryBool(r *http.Request, key string, defaultValue bool) bool {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

// QueryFloat returns a query parameter as a float64 with a default.
func QueryFloat(r *http.Request, key string, defaultValue float64) float64 {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

// QueryArray returns all values for a query parameter.
func QueryArray(r *http.Request, key string) []string {
	return r.URL.Query()[key]
}

// QueryAll returns all query parameters.
func QueryAll(r *http.Request) map[string][]string {
	return r.URL.Query()
}

// Has checks if a query parameter exists.
func Has(r *http.Request, key string) bool {
	return r.URL.Query().Has(key)
}
