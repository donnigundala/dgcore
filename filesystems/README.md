# Filesystem Abstraction Layer

A flexible, extensible filesystem abstraction layer for Go, inspired by Laravel's filesystem. It supports multiple storage drivers, public/private visibility, and a unified API for managing all your storage needs.

## Features

- **Multi-Disk Management**: Configure and manage multiple storage disks (`local`, `s3`, etc.) at once.
- **Unified API**: Interact with any disk using a single, consistent API.
- **Public & Private Visibility**: Control file visibility with `filesystems.Public` and `filesystems.Private`.
- **Unified URL Generation**: A single `GetURL` method returns a public URL for public files or a temporary signed URL for private ones.
- **Easy to Extend**: Add new drivers by implementing the `Storage` interface.

## Installation

```bash
go get github.com/donnigundala/dg-framework/core/filesystems
```

## Quick Start

### 1. Define Your Configuration

Create a configuration that defines all your storage "disks". This can be loaded from a JSON/YAML file or defined in code.

```go
import "github.com/donnigundala/dg-framework/core/filesystems"

config := filesystems.ManagerConfig{
    Default: "local", // Specify the default disk
    Disks: map[string]filesystems.Disk{
        "local": {
            Driver: "local",
            Config: map[string]interface{}{
                "basePath": "./storage/app",
                "baseURL":  "http://localhost:8080/storage",
            },
        },
        "s3_public": {
            Driver: "s3",
            Config: map[string]interface{}{
                "bucket":    "my-public-assets-bucket",
                "region":    "us-east-1",
                "accessKey": "YOUR_AWS_ACCESS_KEY",
                "secretKey": "YOUR_AWS_SECRET_KEY",
                "baseURL":   "https://my-public-assets-bucket.s3.us-east-1.amazonaws.com",
            },
        },
        "s3_private": {
            Driver: "s3",
            Config: map[string]interface{}{
                "bucket":    "my-private-files-bucket",
                "region":    "us-east-1",
                "accessKey": "YOUR_AWS_ACCESS_KEY",
                "secretKey": "YOUR_AWS_SECRET_KEY",
            },
        },
    },
}
```

### 2. Create the FileSystem Manager

Create a single `FileSystem` manager instance from your configuration.

```go
fs, err := filesystems.New(config)
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
err = fs.Upload(ctx, "default-file.txt", data, int64(data.Len()), filesystems.Public)

// --- Use a specific disk by name ---
data = strings.NewReader("Public S3 content.")
err = fs.Disk("s3_public").Upload(ctx, "images/avatar.jpg", data, int64(data.Len()), filesystems.Public)

// --- Get a URL from a public disk ---
// This will be a public URL because the file was uploaded with Public visibility and a baseURL is set.
url, err := fs.Disk("s3_public").GetURL(ctx, "images/avatar.jpg", filesystems.Public, 15*time.Minute)

// --- Get a signed URL for a private file ---
data = strings.NewReader("Private S3 content.")
err = fs.Disk("s3_private").Upload(ctx, "reports/annual.pdf", data, int64(data.Len()), filesystems.Private)

privateURL, err := fs.Disk("s3_private").GetURL(ctx, "reports/annual.pdf", filesystems.Private, 15*time.Minute)
```

## Architecture

The package is composed of two main components:

1.  **The `FileSystem` Manager (`manager.go`)**: The top-level entry point that holds and manages all configured storage disks.
2.  **The `Storage` Interface (`interface.go`)**: A standard interface that defines the contract for all storage drivers. Each driver (`local.go`, `s3.go`) is an implementation of this interface.

The manager uses a `Factory` (`factory.go`) internally to create the individual driver instances based on your configuration.

## Visibility and URLs

The `GetURL` method provides a unified way to get a URL for a file, abstracting away the details of public vs. private access.

- **`filesystems.Public`**: When you upload a file with `Public` visibility, `GetURL` will return a permanent, publicly accessible URL if a `baseURL` is configured for the disk. Otherwise, it may return a signed URL.
- **`filesystems.Private`**: When you upload a file with `Private` visibility, `GetURL` will return a temporary, signed URL that grants access for a limited time.

## Adding a New Driver

To add a new storage driver (e.g., Google Cloud Storage):

1.  **Implement the `Storage` interface** in a new file (e.g., `filesystems/gcs.go`).
2.  **Register a config converter** in your driver file to process its specific configuration map.
3.  **Register the driver constructor** in your driver file, which will be used by the factory.

See `s3.go` for an example of how to register a driver and its config converter.

## License

MIT
