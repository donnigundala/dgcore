package upload

import (
	"bytes"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
)

// File represents an uploaded file.
type File struct {
	FieldName   string               // Form field name
	Filename    string               // Original filename
	Size        int64                // File size in bytes
	ContentType string               // MIME type
	Header      multipart.FileHeader // Full multipart header
	Reader      io.Reader            // File content reader
	data        []byte               // Internal buffer
}

// NewFile creates a new File from multipart.FileHeader.
func NewFile(fieldName string, header *multipart.FileHeader) (*File, error) {
	// Open the uploaded file
	src, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Read file data
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	return &File{
		FieldName:   fieldName,
		Filename:    header.Filename,
		Size:        header.Size,
		ContentType: header.Header.Get("Content-Type"),
		Header:      *header,
		Reader:      bytes.NewReader(data),
		data:        data,
	}, nil
}

// Extension returns the file extension.
func (f *File) Extension() string {
	return filepath.Ext(f.Filename)
}

// IsImage checks if the file is an image.
func (f *File) IsImage() bool {
	return strings.HasPrefix(f.ContentType, "image/")
}

// IsVideo checks if the file is a video.
func (f *File) IsVideo() bool {
	return strings.HasPrefix(f.ContentType, "video/")
}

// IsAudio checks if the file is audio.
func (f *File) IsAudio() bool {
	return strings.HasPrefix(f.ContentType, "audio/")
}

// IsPDF checks if the file is a PDF.
func (f *File) IsPDF() bool {
	return f.ContentType == "application/pdf"
}

// Data returns the file data as bytes.
func (f *File) Data() []byte {
	return f.data
}

// Reset resets the reader to the beginning.
func (f *File) Reset() {
	f.Reader = bytes.NewReader(f.data)
}
