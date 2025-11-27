package foundation

import (
	"fmt"
	"reflect"

	"github.com/donnigundala/dg-core/config"
)

// InjectProviderConfig scans provider struct for `config:"key"` tags
// and automatically injects configuration.
//
// Example:
//
//	type MyProvider struct {
//	    Config MyConfig `config:"myapp" validate:"required"`
//	}
//
// The framework will automatically call config.Inject("myapp", &provider.Config)
// before Register() is called. If the validate tag is present, it will use
// config.InjectAndValidate() instead.
func InjectProviderConfig(provider interface{}) error {
	v := reflect.ValueOf(provider)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil // Not a struct, skip
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Check for config tag
		configKey := fieldType.Tag.Get("config")
		if configKey == "" {
			continue
		}

		// Field must be addressable
		if !field.CanAddr() {
			return fmt.Errorf("field %s is not addressable", fieldType.Name)
		}

		// Check if validation is required
		validateTag := fieldType.Tag.Get("validate")

		// Inject configuration
		var err error
		if validateTag != "" {
			err = config.InjectAndValidate(configKey, field.Addr().Interface())
		} else {
			err = config.Inject(configKey, field.Addr().Interface())
		}

		if err != nil {
			return fmt.Errorf("failed to inject config '%s' into %s: %w",
				configKey, fieldType.Name, err)
		}
	}

	return nil
}
