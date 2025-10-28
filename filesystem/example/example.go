package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/donnigundala/dgcore/filesystem"
)

func main() {
	ctx := context.Background()

	// Example 1: Local Storage
	fmt.Println("=== Local Storage Example ===")
	demonstrateLocalStorage(ctx)

	// Example 2: MinIO Storage
	fmt.Println("\n=== MinIO Storage Example ===")
	demonstrateMinIOStorage(ctx)

	// Example 3: S3 Storage
	fmt.Println("\n=== S3 Storage Example ===")
	demonstrateS3Storage(ctx)
}

func demonstrateLocalStorage(ctx context.Context) {
	factory := filesystem.NewFactory()

	store, err := factory.Create("local", filesystem.LocalConfig{
		BasePath: "./storage",
		BaseURL:  "http://localhost:8080",
		Secret:   "a-very-secure-secret-key-for-hmac",
	})
	if err != nil {
		log.Fatalf("Failed to create local storage: %v", err)
	}

	// Upload a public file
	publicData := strings.NewReader("This is a public file.")
	publicKey := "public-image.jpg"
	err = store.Upload(ctx, publicKey, publicData, int64(publicData.Len()), filesystem.Public)
	if err != nil {
		log.Fatalf("Failed to upload public file: %v", err)
	}
	fmt.Printf("✓ Uploaded public file: %s\n", publicKey)

	// Upload a private file
	privateData := strings.NewReader("This is a top-secret document.")
	privateKey := "private-document.pdf"
	err = store.Upload(ctx, privateKey, privateData, int64(privateData.Len()), filesystem.Private)
	if err != nil {
		log.Fatalf("Failed to upload private file: %v", err)
	}
	fmt.Printf("✓ Uploaded private file: %s\n", privateKey)

	// Get URL for the public file
	publicURL, err := store.GetURL(ctx, publicKey, filesystem.Public, 0)
	if err != nil {
		log.Fatalf("Failed to get public URL: %v", err)
	}
	fmt.Printf("✓ Public URL: %s\n", publicURL)

	// Get a temporary signed URL for the private file
	privateURL, err := store.GetURL(ctx, privateKey, filesystem.Private, 1*time.Hour)
	if err != nil {
		log.Fatalf("Failed to get signed URL: %v", err)
	}
	fmt.Printf("✓ Signed URL (valid for 1 hour): %s\n", privateURL)

	// Clean up
	store.Delete(ctx, publicKey)
	store.Delete(ctx, privateKey)
	fmt.Println("✓ Cleaned up files")
}

func demonstrateMinIOStorage(ctx context.Context) {
	factory := filesystem.NewFactory()

	store, err := factory.Create("minio", filesystem.MinIOConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		UseSSL:          false,
		Bucket:          "my-test-bucket",
		BaseURL:         "http://localhost:9000",
	})
	if err != nil {
		log.Printf("Failed to create MinIO storage: %v (make sure MinIO is running)", err)
		return
	}

	// Upload a public file
	publicData := strings.NewReader("This is a public file for MinIO.")
	publicKey := "public-image.jpg"
	err = store.Upload(ctx, publicKey, publicData, int64(publicData.Len()), filesystem.Public)
	if err != nil {
		log.Printf("Failed to upload public file to MinIO: %v", err)
		return
	}
	fmt.Printf("✓ Uploaded public file to MinIO: %s\n", publicKey)

	// Get URL for the public file
	publicURL, err := store.GetURL(ctx, publicKey, filesystem.Public, 0)
	if err != nil {
		log.Printf("Failed to get public URL from MinIO: %v", err)
		return
	}
	fmt.Printf("✓ Public URL from MinIO: %s\n", publicURL)

	// Clean up
	store.Delete(ctx, publicKey)
	fmt.Println("✓ Cleaned up MinIO file")
}

func demonstrateS3Storage(ctx context.Context) {
	factory := filesystem.NewFactory()

	store, err := factory.Create("s3", filesystem.S3ConfigWithAuth{
		Bucket:    "your-s3-bucket-name", // <-- IMPORTANT: Change this to your bucket name
		Region:    "us-east-1",
		AccessKey: "YOUR_AWS_ACCESS_KEY", // <-- IMPORTANT: Change this
		SecretKey: "YOUR_AWS_SECRET_KEY", // <-- IMPORTANT: Change this
	})
	if err != nil {
		log.Printf("Failed to create S3 storage: %v (configure AWS credentials)", err)
		return
	}

	// Upload a private file
	privateData := strings.NewReader("This is a private S3 document.")
	privateKey := "private-document.pdf"
	err = store.Upload(ctx, privateKey, privateData, int64(privateData.Len()), filesystem.Private)
	if err != nil {
		log.Printf("Failed to upload private file to S3: %v", err)
		return
	}
	fmt.Printf("✓ Uploaded private file to S3: %s\n", privateKey)

	// Get a temporary signed URL for the private file
	privateURL, err := store.GetURL(ctx, privateKey, filesystem.Private, 15*time.Minute)
	if err != nil {
		log.Printf("Failed to get signed URL from S3: %v", err)
		return
	}
	fmt.Printf("✓ Signed URL from S3 (valid for 15 minutes): %s\n", privateURL)

	// Clean up
	store.Delete(ctx, privateKey)
	fmt.Println("✓ Cleaned up S3 file")
}
