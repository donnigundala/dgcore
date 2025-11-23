package request

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/donnigundala/dg-core/validation"
)

// JSON decodes the request body as JSON into the given value.
func JSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

// JSONWithLimit decodes the request body as JSON with a size limit.
func JSONWithLimit(r *http.Request, v interface{}, maxBytes int64) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)
	return JSON(r, v)
}

// JSONWithValidation decodes the request body as JSON and validates it.
func JSONWithValidation(r *http.Request, v interface{}, validator validation.IValidator) error {
	if err := JSON(r, v); err != nil {
		return err
	}
	return validator.ValidateStruct(context.Background(), v)
}

// JSONWithLimitAndValidation decodes with size limit and validation.
func JSONWithLimitAndValidation(r *http.Request, v interface{}, maxBytes int64, validator validation.IValidator) error {
	if err := JSONWithLimit(r, v, maxBytes); err != nil {
		return err
	}
	return validator.ValidateStruct(context.Background(), v)
}

// Body returns the raw request body as bytes.
func Body(r *http.Request) ([]byte, error) {
	return io.ReadAll(r.Body)
}

// BodyString returns the raw request body as a string.
func BodyString(r *http.Request) (string, error) {
	body, err := Body(r)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
