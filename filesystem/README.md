# Filesystem Abstraction Layer

A flexible, extensible filesystem abstraction layer for Go, inspired by Laravel's filesystem. It supports multiple storage drivers like local, MinIO, and AWS S3 with a unified, expressive API.

## Features

- **Multiple Drivers**: Switch between `local`, `minio`, and `s3` storage.
- **Unified Interface**: A single `Storage` interface for all drivers.
- **Public & Private Visibility**: Control file visibility with `filesystem.Public` and `filesystem.Private`.
- **Unified URL Generation**: A single `GetURL` method that returns a public URL for public files or a temporary signed URL for private files.
- **Easy to Extend**: Add new drivers by implementing the `Storage` interface.
- **Factory Pattern**: Simple driver instantiation via a factory.

## Installation

```bash
go get github.com/aws/aws-sdk-go-v2
go get github.com/minio/minio-go/v7
```

## Quick Start

### 1. Create a Factory

```go
import "path/to/your/project/core/filesystem"

factory := filesystem.NewFactory()
```

### 2. Configure a Driver

#### Local Storage

```go
localCfg := filesystem.LocalConfig{
    BasePath: "./storage",
    BaseURL:  "http://localhost:8080",
    Secret:   "a-very-secure-secret-key",
}

store, err := factory.Create("local", localCfg)
```

#### AWS S3 Storage

```go
s3Cfg := filesystem.S3ConfigWithAuth{
    Bucket:    "my-s3-bucket",
    Region:    "us-east-1",
    AccessKey: "YOUR_AWS_ACCESS_KEY",
    SecretKey: "YOUR_AWS_SECRET_KEY",
}

store, err := factory.Create("s3", s3Cfg)
```

#### MinIO Storage

```go
minioCfg := filesystem.MinIOConfig{
    Endpoint:        "localhost:9000",
    AccessKeyID:     "minioadmin",
    SecretAccessKey: "minioadmin",
    UseSSL:          false,
    Bucket:          "my-minio-bucket",
    BaseURL:         "http://localhost:9000",
}

store, err := factory.Create("minio", minioCfg)
```

### 3. Use the Storage Driver

```go
ctx := context.Background()
data := strings.NewReader("This is a test file.")

// Upload a public file
err = store.Upload(ctx, "public/avatar.txt", data, int64(data.Len()), filesystem.Public)

// Upload a private file
err = store.Upload(ctx, "private/document.pdf", data, int64(data.Len()), filesystem.Private)

// Get URL for the public file
// Output: http://localhost:8080/public/avatar.txt (for local driver)
publicURL, err := store.GetURL(ctx, "public/avatar.txt", filesystem.Public, 0)

// Get a temporary signed URL for the private file (valid for 1 hour)
// Output: A signed URL with an expiration
privateURL, err := store.GetURL(ctx, "private/document.pdf", filesystem.Private, 1*time.Hour)

// Download a file
reader, err := store.Download(ctx, "public/avatar.txt")
if err == nil {
    defer reader.Close()
    // ... read the content
}
```

## Storage Interface

All drivers implement this interface:

```go
type Storage interface {
    Upload(ctx context.Context, key string, data io.Reader, size int64, visibility Visibility) error
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    GetURL(ctx context.Context, key string, visibility Visibility, duration time.Duration) (string, error)
    List(ctx context.Context, prefix string) ([]string, error)
}
```

## Visibility and URLs

The `GetURL` method provides a unified way to get a URL for a file, abstracting away the details of public vs. private access.

- **`filesystem.Public`**: When you upload a file with `Public` visibility, `GetURL` will return a permanent, publicly accessible URL.
- **`filesystem.Private`**: When you upload a file with `Private` visibility, `GetURL` will return a temporary, signed URL that grants access for a limited time.

```go
// Get a direct public URL
// The duration parameter is ignored for public files
url, err := store.GetURL(ctx, "image.jpg", filesystem.Public, 0)

// Get a signed URL valid for 15 minutes
url, err := store.GetURL(ctx, "report.pdf", filesystem.Private, 15*time.Minute)
```

### Validating Local Storage Signed URLs

For the `local` driver, you can validate incoming signed URLs in your HTTP handler:

```go
// Assume `localStore` is your instantiated LocalStorage driver
func handleRequest(w http.ResponseWriter, r *http.Request) {
    expires := r.URL.Query().Get("expires")
    signature := r.URL.Query().Get("signature")
    urlPath := r.URL.Path

    if localStore.VerifySignedURL(urlPath, expires, signature) {
        // URL is valid and not expired - serve the file
    } else {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
    }
}
```

## Adding a New Driver

To add a new storage driver (e.g., Google Cloud Storage):

1.  **Implement the `Storage` interface** in a new file (e.g., `filesystem/gcs.go`).
2.  **Update the factory** in `filesystem/factory.go` to include a new case for your driver.
3.  **Use it** by passing the new driver type and config to `factory.Create()`.

## License

MIT
