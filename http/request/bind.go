package request

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/donnigundala/dg-core/validation"
)

// Bind binds request data to a struct based on Content-Type.
// Supports JSON, form data, and query parameters.
func Bind(r *http.Request, v interface{}, validator validation.IValidator) error {
	contentType := r.Header.Get("Content-Type")

	var err error
	switch {
	case contentType == "application/json":
		err = JSON(r, v)
	case contentType == "application/x-www-form-urlencoded" || contentType == "multipart/form-data":
		err = bindForm(r, v)
	default:
		// Try JSON first, then form
		err = JSON(r, v)
		if err != nil {
			err = bindForm(r, v)
		}
	}

	if err != nil {
		return err
	}

	if validator != nil {
		return validator.ValidateStruct(context.Background(), v)
	}

	return nil
}

// BindJSON binds JSON request body to a struct with validation.
func BindJSON(r *http.Request, v interface{}, validator validation.IValidator) error {
	if err := JSON(r, v); err != nil {
		return err
	}
	if validator != nil {
		return validator.ValidateStruct(context.Background(), v)
	}
	return nil
}

// BindForm binds form data to a struct with validation.
func BindForm(r *http.Request, v interface{}, validator validation.IValidator) error {
	if err := bindForm(r, v); err != nil {
		return err
	}
	if validator != nil {
		return validator.ValidateStruct(context.Background(), v)
	}
	return nil
}

// BindQuery binds query parameters to a struct with validation.
func BindQuery(r *http.Request, v interface{}, validator validation.IValidator) error {
	if err := bindQuery(r, v); err != nil {
		return err
	}
	if validator != nil {
		return validator.ValidateStruct(context.Background(), v)
	}
	return nil
}

// bindForm binds form data to a struct using reflection.
func bindForm(r *http.Request, v interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return bindValues(r.Form, v, "form")
}

// bindQuery binds query parameters to a struct using reflection.
func bindQuery(r *http.Request, v interface{}) error {
	return bindValues(r.URL.Query(), v, "json")
}

// bindValues binds url.Values to a struct using reflection.
func bindValues(values url.Values, v interface{}, tag string) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return json.Unmarshal([]byte("{}"), v) // Return error
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return json.Unmarshal([]byte("{}"), v) // Return error
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if !field.CanSet() {
			continue
		}

		fieldType := rt.Field(i)
		tagValue := fieldType.Tag.Get(tag)
		if tagValue == "" || tagValue == "-" {
			continue
		}

		value := values.Get(tagValue)
		if value == "" {
			continue
		}

		if err := setField(field, value); err != nil {
			return err
		}
	}

	return nil
}

// setField sets a reflect.Value from a string.
func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	}
	return nil
}
