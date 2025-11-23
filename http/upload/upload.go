package upload

import (
	"fmt"
	"net/http"
)

// HandleUpload handles a single file upload.
func HandleUpload(r *http.Request, config Config) (*File, error) {
	// Set defaults
	if config.FieldName == "" {
		config.FieldName = "file"
	}
	if config.MaxSize == 0 {
		config.MaxSize = 10 * 1024 * 1024 // 10MB default
	}

	// Parse multipart form
	err := r.ParseMultipartForm(config.MaxSize)
	if err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// Get file from form
	_, header, err := r.FormFile(config.FieldName)
	if err != nil {
		return nil, fmt.Errorf("failed to get file from field '%s': %w", config.FieldName, err)
	}

	// Create File struct
	file, err := NewFile(config.FieldName, header)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Validate file
	validators := []Validator{}

	// Add size validator
	if config.MaxSize > 0 {
		validators = append(validators, SizeValidator(config.MaxSize))
	}

	// Add type validator
	if len(config.AllowedTypes) > 0 {
		validators = append(validators, TypeValidator(config.AllowedTypes))
	}

	// Validate
	if err := ValidateFile(file, validators...); err != nil {
		return nil, err
	}

	return file, nil
}

// HandleMultipleUploads handles multiple file uploads.
func HandleMultipleUploads(r *http.Request, config Config) ([]*File, error) {
	// Set defaults
	if config.FieldName == "" {
		config.FieldName = "file"
	}
	if config.MaxSize == 0 {
		config.MaxSize = 10 * 1024 * 1024 // 10MB default per file
	}
	if config.MaxFiles == 0 {
		config.MaxFiles = 10 // Default max 10 files
	}

	// Parse multipart form
	err := r.ParseMultipartForm(config.MaxSize * int64(config.MaxFiles))
	if err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// Get files from form
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, fmt.Errorf("no files uploaded")
	}

	headers := r.MultipartForm.File[config.FieldName]
	if len(headers) == 0 {
		return nil, fmt.Errorf("no files found in field '%s'", config.FieldName)
	}

	// Check max files
	if len(headers) > config.MaxFiles {
		return nil, fmt.Errorf("too many files: %d, maximum allowed: %d", len(headers), config.MaxFiles)
	}

	// Process each file
	files := make([]*File, 0, len(headers))
	validators := []Validator{}

	// Add validators
	if config.MaxSize > 0 {
		validators = append(validators, SizeValidator(config.MaxSize))
	}
	if len(config.AllowedTypes) > 0 {
		validators = append(validators, TypeValidator(config.AllowedTypes))
	}

	for _, header := range headers {
		file, err := NewFile(config.FieldName, header)
		if err != nil {
			return nil, fmt.Errorf("failed to read file '%s': %w", header.Filename, err)
		}

		// Validate
		if err := ValidateFile(file, validators...); err != nil {
			return nil, fmt.Errorf("validation failed for '%s': %w", header.Filename, err)
		}

		files = append(files, file)
	}

	return files, nil
}

// HandleUploadWithValidators handles upload with custom validators.
func HandleUploadWithValidators(r *http.Request, config Config, validators ...Validator) (*File, error) {
	// Get file without validation
	file, err := HandleUpload(r, Config{
		FieldName: config.FieldName,
		MaxSize:   config.MaxSize,
	})
	if err != nil {
		return nil, err
	}

	// Apply custom validators
	if err := ValidateFile(file, validators...); err != nil {
		return nil, err
	}

	return file, nil
}
