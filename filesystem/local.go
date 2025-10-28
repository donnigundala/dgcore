package filesystem

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

// LocalStorage implements the Storage interface for a local filesystem.
// It supports public and private visibility for objects, storing them in
// separate subdirectories.
type LocalStorage struct {
	config LocalConfig
}

// NewLocalStorage creates a new LocalStorage instance.
func NewLocalStorage(config LocalConfig) (*LocalStorage, error) {
	// Create base directories if they don't exist
	if err := os.MkdirAll(filepath.Join(config.BasePath, string(Public)), 0755); err != nil {
		return nil, fmt.Errorf("failed to create public directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(config.BasePath, string(Private)), 0755); err != nil {
		return nil, fmt.Errorf("failed to create private directory: %w", err)
	}

	return &LocalStorage{config: config}, nil
}

// Upload stores data at the given key with the specified visibility.
func (ls *LocalStorage) Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error {
	path := filepath.Join(ls.config.BasePath, string(visibility), key)

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, data); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// Download retrieves data from the given key.
func (ls *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	path, _, err := ls.findFile(key)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete removes the object at the given key.
func (ls *LocalStorage) Delete(ctx context.Context, key string) error {
	path, _, err := ls.findFile(key)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists checks if an object exists at the given key.
func (ls *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, _, err := ls.findFile(key)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetURL generates a URL for accessing the object.
func (ls *LocalStorage) GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error) {
	if visibility == Public {
		return ls.getPublicURL(key), nil
	}
	return ls.getSignedURL(ctx, key, duration)
}

// List returns all keys with the given prefix.
func (ls *LocalStorage) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string

	for _, visibility := range []Visibility{Public, Private} {
		basePath := filepath.Join(ls.config.BasePath, string(visibility), prefix)

		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				relPath, err := filepath.Rel(filepath.Join(ls.config.BasePath, string(visibility)), path)
				if err != nil {
					return err
				}
				keys = append(keys, relPath)
			}
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return keys, nil
}

// findFile searches for a file in both public and private storage and returns its full path and visibility.
func (ls *LocalStorage) findFile(key string) (string, Visibility, error) {
	// Check public storage first
	publicPath := filepath.Join(ls.config.BasePath, string(Public), key)
	if _, err := os.Stat(publicPath); err == nil {
		return publicPath, Public, nil
	}

	// Then check private storage
	privatePath := filepath.Join(ls.config.BasePath, string(Private), key)
	if _, err := os.Stat(privatePath); err == nil {
		return privatePath, Private, nil
	}

	return "", "", os.ErrNotExist
}

// getPublicURL returns the direct URL for a public object.
func (ls *LocalStorage) getPublicURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", ls.config.BaseURL, string(Public), key)
}

// getSignedURL generates a temporary signed URL for a private object.
func (ls *LocalStorage) getSignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	path, visibility, err := ls.findFile(key)
	if err != nil {
		return "", err
	}
	if visibility != Private {
		return "", fmt.Errorf("cannot generate signed URL for public object")
	}

	// Generate expiration timestamp
	expiresAt := time.Now().Add(duration).Unix()

	// Create the path component for the URL
	urlPath := "/" + string(Private) + "/" + strings.ReplaceAll(key, "\\", "/")

	// Create signature payload: method + path + expires
	payload := fmt.Sprintf("GET\n%s\n%d", urlPath, expiresAt)

	// Generate HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(ls.config.Secret))
	h.Write([]byte(payload))
	signature := hex.EncodeToString(h.Sum(nil))

	// Build the signed URL
	signedURL := fmt.Sprintf(
		"%s%s?expires=%d&signature=%s",
		ls.config.BaseURL,
		urlPath,
		expiresAt,
		url.QueryEscape(signature),
	)

	return signedURL, nil
}

// VerifySignedURL validates a signed URL.
func (ls *LocalStorage) VerifySignedURL(urlPath string, expiresStr, signature string) bool {
	// Parse expiration timestamp
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return false
	}

	// Check if URL has expired
	if time.Now().Unix() > expires {
		return false
	}

	// Recreate the signature
	payload := fmt.Sprintf("GET\n%s\n%d", urlPath, expires)
	h := hmac.New(sha256.New, []byte(ls.config.Secret))
	h.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
