package filesystems

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// LocalConfig holds configuration for the local driver.
type LocalConfig struct {
	BasePath string
	BaseURL  string
	Secret   string
}

// register converter and driver at package init
func init() {
	RegisterConfigConverter("local", func(cfg map[string]interface{}) (interface{}, error) {
		lc := LocalConfig{}
		if v, ok := cfg["basePath"].(string); ok {
			lc.BasePath = v
		}
		if v, ok := cfg["baseURL"].(string); ok {
			lc.BaseURL = v
		}
		if v, ok := cfg["secret"].(string); ok {
			lc.Secret = v
		}
		if lc.BasePath == "" {
			return nil, fmt.Errorf("local: basePath required")
		}
		return lc, nil
	})

	RegisterDriver("local", func(config interface{}, logger *slog.Logger, traceIDKey any) (Storage, error) {
		cfg, ok := config.(LocalConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config for local driver")
		}
		return NewLocalDriver(cfg, logger, traceIDKey)
	})
}

// LocalDriver implements Storage using the filesystem.
type LocalDriver struct {
	cfg        LocalConfig
	logger     *slog.Logger
	traceIDKey any
}

func NewLocalDriver(cfg LocalConfig, logger *slog.Logger, traceIDKey any) (*LocalDriver, error) {
	if cfg.BasePath == "" {
		return nil, fmt.Errorf("basePath is required")
	}
	// Ensure the base path exists and is a directory.
	info, err := os.Stat(cfg.BasePath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("Base path does not exist, creating it", "path", cfg.BasePath)
			if err := os.MkdirAll(cfg.BasePath, 0o755); err != nil {
				logger.Error("Could not create base path", "path", cfg.BasePath, "error", err)
				return nil, fmt.Errorf("local: could not create base path: %w", err)
			}
		} else {
			logger.Error("Could not stat base path", "path", cfg.BasePath, "error", err)
			return nil, fmt.Errorf("local: could not stat base path: %w", err)
		}
	} else if !info.IsDir() {
		logger.Error("Base path is not a directory", "path", cfg.BasePath)
		return nil, fmt.Errorf("local: base path is not a directory")
	}
	return &LocalDriver{cfg: cfg, logger: logger, traceIDKey: traceIDKey}, nil
}

func (l *LocalDriver) loggerFrom(ctx context.Context) *slog.Logger {
	logger := l.logger
	if l.traceIDKey != nil {
		if traceID, ok := ctx.Value(l.traceIDKey).(string); ok && traceID != "" {
			logger = logger.With(slog.String("trace_id", traceID))
		}
	}
	return logger
}

func (l *LocalDriver) resolvePath(key string, logger *slog.Logger) (string, error) {
	// Clean key to prevent directory traversal.
	cleanKey := filepath.Clean(key)
	if strings.Contains(cleanKey, "..") {
		logger.Warn("Invalid key contains directory traversal", "key", key)
		return "", fmt.Errorf("local: invalid key, contains '..'")
	}
	// Join with base path and check if it's still within the base path.
	path := filepath.Join(l.cfg.BasePath, cleanKey)
	rel, err := filepath.Rel(l.cfg.BasePath, path)
	if err != nil {
		return "", fmt.Errorf("local: could not determine relative path: %w", err)
	}
	if strings.HasPrefix(rel, "..") {
		logger.Warn("Resolved path is outside of the base path", "key", key, "path", path)
		return "", fmt.Errorf("local: resolved path is outside of the base path")
	}
	return path, nil
}

func (l *LocalDriver) Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error {
	logger := l.loggerFrom(ctx)
	path, err := l.resolvePath(key, logger)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		logger.Error("Failed to create directory for upload", "path", path, "error", err)
		return fmt.Errorf("local: mkdir: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		logger.Error("Failed to create file for upload", "path", path, "error", err)
		return fmt.Errorf("local: create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	if err != nil {
		logger.Error("Failed to write to file during upload", "path", path, "error", err)
		return fmt.Errorf("local: write file: %w", err)
	}
	logger.Debug("Uploaded file", "key", key, "path", path)
	return nil
}

func (l *LocalDriver) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	logger := l.loggerFrom(ctx)
	path, err := l.resolvePath(key, logger)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Debug("File not found for download", "key", key, "path", path)
			return nil, ErrNotFound
		}
		logger.Error("Failed to open file for download", "key", key, "path", path, "error", err)
		return nil, fmt.Errorf("local: open: %w", err)
	}
	return f, nil
}

func (l *LocalDriver) Delete(ctx context.Context, key string) error {
	logger := l.loggerFrom(ctx)
	path, err := l.resolvePath(key, logger)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			logger.Debug("File not found for deletion", "key", key, "path", path)
			return ErrNotFound
		}
		logger.Error("Failed to delete file", "key", key, "path", path, "error", err)
		return fmt.Errorf("local: remove: %w", err)
	}
	logger.Debug("Deleted file", "key", key, "path", path)
	return nil
}

func (l *LocalDriver) Exists(ctx context.Context, key string) (bool, error) {
	logger := l.loggerFrom(ctx)
	path, err := l.resolvePath(key, logger)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	logger.Warn("Failed to check existence of file", "key", key, "path", path, "error", err)
	return false, fmt.Errorf("local: stat: %w", err)
}

func (l *LocalDriver) GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error) {
	logger := l.loggerFrom(ctx)
	if l.cfg.BaseURL == "" {
		return "", fmt.Errorf("local: baseURL not configured")
	}

	baseURL := strings.TrimRight(l.cfg.BaseURL, "/") + "/" + strings.TrimLeft(key, "/")

	if visibility == VisibilityPublic {
		return baseURL, nil
	}

	// For private visibility, generate a temporary signed URL.
	if l.cfg.Secret == "" {
		logger.Error("Cannot generate signed URL, secret not configured", "key", key)
		return "", fmt.Errorf("local: secret not configured for signed URLs")
	}

	expires := time.Now().Add(duration).Unix()
	mac := hmac.New(sha256.New, []byte(l.cfg.Secret))
	mac.Write([]byte(key + strconv.FormatInt(expires, 10)))
	signature := hex.EncodeToString(mac.Sum(nil))

	q := url.Values{}
	q.Set("expires", strconv.FormatInt(expires, 10))
	q.Set("signature", signature)

	return baseURL + "?" + q.Encode(), nil
}

func (l *LocalDriver) List(ctx context.Context, prefix string) ([]string, error) {
	logger := l.loggerFrom(ctx)
	root, err := l.resolvePath("", logger)
	if err != nil {
		return nil, err
	}
	searchBase, err := l.resolvePath(prefix, logger)
	if err != nil {
		return nil, err
	}

	out := []string{}
	err = filepath.Walk(searchBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, rerr := filepath.Rel(root, path)
			if rerr != nil {
				return rerr
			}
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})

	if err != nil {
		// If the prefix doesn't exist, return an empty list instead of an error.
		if _, ok := err.(*os.PathError); ok {
			logger.Debug("Prefix not found for list, returning empty list", "prefix", prefix)
			return []string{}, nil
		}
		logger.Error("Failed to walk directory for list", "prefix", prefix, "error", err)
		return nil, fmt.Errorf("local: walk: %w", err)
	}
	return out, nil
}
