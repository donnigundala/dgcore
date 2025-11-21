package filesystems

// Visibility describes object visibility
type Visibility int

const (
	VisibilityPrivate Visibility = iota
	VisibilityPublic
)

// Common errors
var (
	ErrNotFound     = NewFilesystemError("not found")
	ErrInvalidInput = NewFilesystemError("invalid input")
	ErrInternal     = NewFilesystemError("internal error")
)

// FilesystemError wraps simple error strings to allow type checks.
type FilesystemError struct {
	Msg string
}

func (e *FilesystemError) Error() string { return e.Msg }

func NewFilesystemError(msg string) error { return &FilesystemError{Msg: msg} }
