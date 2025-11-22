package testing

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Assert provides assertion helpers for tests.
var Assert = &Assertions{}

// Assertions provides test assertion methods.
type Assertions struct{}

// Equal asserts that two values are equal.
func (a *Assertions) Equal(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		msg := formatMessage("Expected values to be equal", msgAndArgs...)
		t.Errorf("%s\nExpected: %v\nActual:   %v", msg, expected, actual)
	}
}

// NotEqual asserts that two values are not equal.
func (a *Assertions) NotEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		msg := formatMessage("Expected values to be different", msgAndArgs...)
		t.Errorf("%s\nBoth:     %v", msg, expected)
	}
}

// Nil asserts that a value is nil.
func (a *Assertions) Nil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !isNil(value) {
		msg := formatMessage("Expected value to be nil", msgAndArgs...)
		t.Errorf("%s\nActual:   %v", msg, value)
	}
}

// NotNil asserts that a value is not nil.
func (a *Assertions) NotNil(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if isNil(value) {
		msg := formatMessage("Expected value to not be nil", msgAndArgs...)
		t.Error(msg)
	}
}

// True asserts that a condition is true.
func (a *Assertions) True(t *testing.T, condition bool, msgAndArgs ...interface{}) {
	t.Helper()
	if !condition {
		msg := formatMessage("Expected condition to be true", msgAndArgs...)
		t.Error(msg)
	}
}

// False asserts that a condition is false.
func (a *Assertions) False(t *testing.T, condition bool, msgAndArgs ...interface{}) {
	t.Helper()
	if condition {
		msg := formatMessage("Expected condition to be false", msgAndArgs...)
		t.Error(msg)
	}
}

// Contains asserts that a string contains a substring.
func (a *Assertions) Contains(t *testing.T, haystack, needle string, msgAndArgs ...interface{}) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		msg := formatMessage("Expected string to contain substring", msgAndArgs...)
		t.Errorf("%s\nHaystack: %s\nNeedle:   %s", msg, haystack, needle)
	}
}

// NotContains asserts that a string does not contain a substring.
func (a *Assertions) NotContains(t *testing.T, haystack, needle string, msgAndArgs ...interface{}) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		msg := formatMessage("Expected string to not contain substring", msgAndArgs...)
		t.Errorf("%s\nHaystack: %s\nNeedle:   %s", msg, haystack, needle)
	}
}

// HasPrefix asserts that a string has a prefix.
func (a *Assertions) HasPrefix(t *testing.T, str, prefix string, msgAndArgs ...interface{}) {
	t.Helper()
	if !strings.HasPrefix(str, prefix) {
		msg := formatMessage("Expected string to have prefix", msgAndArgs...)
		t.Errorf("%s\nString: %s\nPrefix: %s", msg, str, prefix)
	}
}

// HasSuffix asserts that a string has a suffix.
func (a *Assertions) HasSuffix(t *testing.T, str, suffix string, msgAndArgs ...interface{}) {
	t.Helper()
	if !strings.HasSuffix(str, suffix) {
		msg := formatMessage("Expected string to have suffix", msgAndArgs...)
		t.Errorf("%s\nString: %s\nSuffix: %s", msg, str, suffix)
	}
}

// NoError asserts that an error is nil.
func (a *Assertions) NoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		msg := formatMessage("Expected no error", msgAndArgs...)
		t.Errorf("%s\nError:    %v", msg, err)
	}
}

// Error asserts that an error is not nil.
func (a *Assertions) Error(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		msg := formatMessage("Expected an error", msgAndArgs...)
		t.Error(msg)
	}
}

// ErrorContains asserts that an error contains a message.
func (a *Assertions) ErrorContains(t *testing.T, err error, message string, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		msg := formatMessage("Expected an error", msgAndArgs...)
		t.Error(msg)
		return
	}
	if !strings.Contains(err.Error(), message) {
		msg := formatMessage("Expected error to contain message", msgAndArgs...)
		t.Errorf("%s\nError:   %v\nMessage: %s", msg, err, message)
	}
}

// Len asserts that a collection has a specific length.
func (a *Assertions) Len(t *testing.T, collection interface{}, length int, msgAndArgs ...interface{}) {
	t.Helper()
	v := reflect.ValueOf(collection)
	if v.Len() != length {
		msg := formatMessage("Expected collection to have specific length", msgAndArgs...)
		t.Errorf("%s\nExpected: %d\nActual:   %d", msg, length, v.Len())
	}
}

// Empty asserts that a collection is empty.
func (a *Assertions) Empty(t *testing.T, collection interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	v := reflect.ValueOf(collection)
	if v.Len() != 0 {
		msg := formatMessage("Expected collection to be empty", msgAndArgs...)
		t.Errorf("%s\nLength:   %d", msg, v.Len())
	}
}

// NotEmpty asserts that a collection is not empty.
func (a *Assertions) NotEmpty(t *testing.T, collection interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	v := reflect.ValueOf(collection)
	if v.Len() == 0 {
		msg := formatMessage("Expected collection to not be empty", msgAndArgs...)
		t.Error(msg)
	}
}

// Helper functions

func isNil(value interface{}) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func formatMessage(defaultMsg string, msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 {
		return defaultMsg
	}
	if len(msgAndArgs) == 1 {
		return fmt.Sprintf("%v", msgAndArgs[0])
	}
	return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
}
