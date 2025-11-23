package upload

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Validator defines the interface for file validation.
type Validator interface {
	Validate(file *File) error
}

// ValidatorFunc is a function adapter for Validator interface.
type ValidatorFunc func(file *File) error

func (f ValidatorFunc) Validate(file *File) error {
	return f(file)
}

// SizeValidator validates file size.
func SizeValidator(maxSize int64) Validator {
	return ValidatorFunc(func(file *File) error {
		if file.Size > maxSize {
			return fmt.Errorf("file size %d exceeds maximum %d bytes", file.Size, maxSize)
		}
		return nil
	})
}

// TypeValidator validates MIME type.
func TypeValidator(allowedTypes []string) Validator {
	return ValidatorFunc(func(file *File) error {
		if len(allowedTypes) == 0 {
			return nil // All types allowed
		}

		for _, allowedType := range allowedTypes {
			if file.ContentType == allowedType {
				return nil
			}
		}

		return fmt.Errorf("file type %s not allowed, allowed types: %v", file.ContentType, allowedTypes)
	})
}

// ExtensionValidator validates file extension.
func ExtensionValidator(allowedExts []string) Validator {
	return ValidatorFunc(func(file *File) error {
		if len(allowedExts) == 0 {
			return nil // All extensions allowed
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		for _, allowedExt := range allowedExts {
			if ext == strings.ToLower(allowedExt) {
				return nil
			}
		}

		return fmt.Errorf("file extension %s not allowed, allowed extensions: %v", ext, allowedExts)
	})
}

// ImageValidator validates that the file is an image.
func ImageValidator() Validator {
	return ValidatorFunc(func(file *File) error {
		if !file.IsImage() {
			return fmt.Errorf("file must be an image, got: %s", file.ContentType)
		}
		return nil
	})
}

// ValidateFile validates a file using multiple validators.
func ValidateFile(file *File, validators ...Validator) error {
	for _, validator := range validators {
		if err := validator.Validate(file); err != nil {
			return err
		}
	}
	return nil
}
