# Filesystem Abstraction Layer

A flexible, extensible filesystem abstraction layer for Go, inspired by Laravel's filesystem. It supports multiple storage drivers, public/private visibility, and a unified API for managing all your storage needs.

## Features

- **Multi-Disk Management**: Configure and manage multiple storage disks (`local`, `s3`, `minio`, etc.) at once.
- **Unified API**: Interact with any disk using a single, consistent API.
- **Public & Private Visibility**: Control file visibility with `filesystem.Public` and `filesystem.Private`.
- **Unified URL Generation**: A single `GetURL` method returns a public URL for public files or a temporary signed URL for private ones.
- **Easy to Extend**: Add new drivers by implementing the `Storage` interface.

## Installation

```bash
go get github.com/donnigundala/dg-framework/core/filesystem
```

## Quick Start

### 1. Define Your Configuration

Create a configuration that defines all your storage "disks". This can be loaded from a JSON/YAML file or defined in code.

```go
import "github.com/donnigundala/dg-framework/core/filesystem"

config := filesystem.ManagerConfig{
    Default: "local", // Specify the default disk
    Disks: map[string]filesystem.Disk{
        "local": {
            Driver: "local",
            Config: map[string]interface{}{
                "basePath": "./storage",
                "baseURL":  "http://localhost:8080",
                "secret":   "a-very-secure-secret-key",
            },
        },
        "s3_public": {
            Driver: "s3",
            Config: map[string]interface{}{
                "bucket":    "my-public-assets-bucket",
                "region":    "us-east-1",
                "accessKey": "YOUR_AWS_ACCESS_KEY",
                "secretKey": "YOUR_AWS_SECRET_KEY",
            },
        },
        "minio_private": {
            Driver: "minio",
            Config: map[string]interface{}{
                "endpoint":        "localhost:9000",
                "accessKeyID":     "minioadmin",
                "secretAccessKey": "minioadmin",
                "useSSL":          false,
                "bucket":          "private-documents",
                "baseURL":         "http://localhost:9000",
            },
        },
    },
}
```

### 2. Create the FileSystem Manager

Create a single `FileSystem` manager instance from your configuration.

```go
fs, err := filesystem.New(config)
if err != nil {
    log.Fatalf("Failed to create filesystem manager: %v", err)
}
```

### 3. Use the Manager

Now you can interact with any of your configured disks.

```go
ctx := context.Background()

// --- Use the default disk ("local") ---
data := strings.NewReader("Default disk content.")
err = fs.Upload(ctx, "default-file.txt", data, int64(data.Len()), filesystem.Public)

// --- Use a specific disk by name ---
data = strings.NewReader("Public S3 content.")
err = fs.Disk("s3_public").Upload(ctx, "images/avatar.jpg", data, int64(data.Len()), filesystem.Public)

// --- Get a URL from a specific disk ---
// This will be a public URL because the file was uploaded with Public visibility.
url, err := fs.Disk("s3_public").GetURL(ctx, "images/avatar.jpg", filesystem.Public, 0)

// --- Get a signed URL for a private file ---
data = strings.NewReader("Private MinIO content.")
err = fs.Disk("minio_private").Upload(ctx, "reports/annual.pdf", data, int64(data.Len()), filesystem.Private)

privateURL, err := fs.Disk("minio_private").GetURL(ctx, "reports/annual.pdf", filesystem.Private, 15*time.Minute)
```

## Architecture

The package is composed of two main components:

1.  **The `FileSystem` Manager (`manager.go`)**: The top-level entry point that holds and manages all configured storage disks.
2.  **The `Storage` Interface (`interface.go`)**: A standard interface that defines the contract for all storage drivers. Each driver (`local.go`, `s3.go`, `minio.go`) is an implementation of this interface.

The manager uses a `Factory` (`factory.go`) internally to create the individual driver instances based on your configuration.

## Visibility and URLs

The `GetURL` method provides a unified way to get a URL for a file, abstracting away the details of public vs. private access.

- **`filesystem.Public`**: When you upload a file with `Public` visibility, `GetURL` will return a permanent, publicly accessible URL.
- **`filesystem.Private`**: When you upload a file with `Private` visibility, `GetURL` will return a temporary, signed URL that grants access for a limited time.

## Adding a New Driver

To add a new storage driver (e.g., Google Cloud Storage):

1.  **Implement the `Storage` interface** in a new file (e.g., `filesystem/gcs.go`).
2.  **Update the factory** in `filesystem/factory.go` to include a `case` for your new driver.
3.  **Update the `convertConfig` function** in `filesystem/manager.go` to handle your new driver's configuration.
4.  You can now add the driver to your `ManagerConfig`.

## License

MIT
