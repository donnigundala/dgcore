package filesystems

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

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

	RegisterDriver("local", func(config interface{}) (Storage, error) {
		cfg, ok := config.(LocalConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config for local driver")
		}
		return NewLocalDriver(cfg)
	})
}

// LocalDriver implements Storage using the filesystem.
type LocalDriver struct {
	cfg LocalConfig
}

func NewLocalDriver(cfg LocalConfig) (*LocalDriver, error) {
	if cfg.BasePath == "" {
		return nil, fmt.Errorf("basePath is required")
	}
	// Ensure the base path exists and is a directory.
	info, err := os.Stat(cfg.BasePath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cfg.BasePath, 0o755); err != nil {
				return nil, fmt.Errorf("local: could not create base path: %w", err)
			}
		} else {
			return nil, fmt.Errorf("local: could not stat base path: %w", err)
		}
	} else if !info.IsDir() {
		return nil, fmt.Errorf("local: base path is not a directory")
	}
	return &LocalDriver{cfg: cfg}, nil
}

func (l *LocalDriver) resolvePath(key string) (string, error) {
	// Clean key to prevent directory traversal.
	cleanKey := filepath.Clean(key)
	if strings.Contains(cleanKey, "..") {
		return "", fmt.Errorf("local: invalid key, contains '..'")
	}
	// Join with base path and check if it's still within the base path.
	path := filepath.Join(l.cfg.BasePath, cleanKey)
	rel, err := filepath.Rel(l.cfg.BasePath, path)
	if err != nil {
		return "", fmt.Errorf("local: could not determine relative path: %w", err)
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("local: resolved path is outside of the base path")
	}
	return path, nil
}

func (l *LocalDriver) Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error {
	path, err := l.resolvePath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("local: mkdir: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("local: create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	if err != nil {
		return fmt.Errorf("local: write file: %w", err)
	}
	return nil
}

func (l *LocalDriver) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	path, err := l.resolvePath(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("local: open: %w", err)
	}
	return f, nil
}

func (l *LocalDriver) Delete(ctx context.Context, key string) error {
	path, err := l.resolvePath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("local: remove: %w", err)
	}
	return nil
}

func (l *LocalDriver) Exists(ctx context.Context, key string) (bool, error) {
	path, err := l.resolvePath(key)
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
	return false, fmt.Errorf("local: stat: %w", err)
}

func (l *LocalDriver) GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error) {
	if l.cfg.BaseURL == "" {
		return "", fmt.Errorf("local: baseURL not configured")
	}

	baseURL := strings.TrimRight(l.cfg.BaseURL, "/") + "/" + strings.TrimLeft(key, "/")

	if visibility == VisibilityPublic {
		return baseURL, nil
	}

	// For private visibility, generate a temporary signed URL.
	if l.cfg.Secret == "" {
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
	root, err := l.resolvePath("")
	if err != nil {
		return nil, err
	}
	searchBase, err := l.resolvePath(prefix)
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
			return []string{}, nil
		}
		return nil, fmt.Errorf("local: walk: %w", err)
	}
	return out, nil
}
