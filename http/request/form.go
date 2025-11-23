package request

import (
	"net/http"
	"strconv"
)

// Form returns a form value with a default.
func Form(r *http.Request, key, defaultValue string) string {
	if err := r.ParseForm(); err != nil {
		return defaultValue
	}
	value := r.FormValue(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// FormInt returns a form value as an integer with a default.
func FormInt(r *http.Request, key string, defaultValue int) int {
	if err := r.ParseForm(); err != nil {
		return defaultValue
	}
	value := r.FormValue(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// FormInt64 returns a form value as an int64 with a default.
func FormInt64(r *http.Request, key string, defaultValue int64) int64 {
	if err := r.ParseForm(); err != nil {
		return defaultValue
	}
	value := r.FormValue(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// FormBool returns a form value as a boolean with a default.
func FormBool(r *http.Request, key string, defaultValue bool) bool {
	if err := r.ParseForm(); err != nil {
		return defaultValue
	}
	value := r.FormValue(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

// FormFloat returns a form value as a float64 with a default.
func FormFloat(r *http.Request, key string, defaultValue float64) float64 {
	if err := r.ParseForm(); err != nil {
		return defaultValue
	}
	value := r.FormValue(key)
	if value == "" {
		return defaultValue
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

// FormArray returns all values for a form field.
func FormArray(r *http.Request, key string) []string {
	if err := r.ParseForm(); err != nil {
		return []string{}
	}
	return r.Form[key]
}

// FormAll returns all form values.
func FormAll(r *http.Request) map[string][]string {
	if err := r.ParseForm(); err != nil {
		return map[string][]string{}
	}
	return r.Form
}

// HasFile checks if a file was uploaded with the given field name.
func HasFile(r *http.Request, key string) bool {
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		return false
	}
	if r.MultipartForm == nil {
		return false
	}
	_, ok := r.MultipartForm.File[key]
	return ok
}
