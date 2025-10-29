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

	// 1. Define your storage configuration for multiple disks.
	config := filesystem.ManagerConfig{
		Default: "local", // Specify the default disk
		Disks: map[string]filesystem.Disk{
			"local": {
				Driver: "local",
				Config: map[string]interface{}{
					"basePath": "./storage",
					"baseURL":  "http://localhost:8080",
					"secret":   "a-very-secure-secret-key-for-hmac",
				},
			},
			"s3_public": {
				Driver: "s3",
				Config: map[string]interface{}{
					"bucket":    "your-public-s3-bucket", // <-- IMPORTANT: Change this
					"region":    "us-east-1",
					"accessKey": "YOUR_AWS_ACCESS_KEY", // <-- IMPORTANT: Change this
					"secretKey": "YOUR_AWS_SECRET_KEY", // <-- IMPORTANT: Change this
				},
			},
			"minio_private": {
				Driver: "minio",
				Config: map[string]interface{}{
					"endpoint":        "localhost:9000",
					"accessKeyID":     "minioadmin",
					"secretAccessKey": "minioadmin",
					"useSSL":          false,
					"bucket":          "private-files",
					"baseURL":         "http://localhost:9000",
				},
			},
		},
	}

	// 2. Create the FileSystem manager from the configuration.
	fs, err := filesystem.New(config)
	if err != nil {
		log.Fatalf("Failed to create filesystem manager: %v", err)
	}

	fmt.Println("=== Filesystem Manager Example ===")

	// 3. Use the manager to interact with different disks.

	// --- Use the default disk ("local") ---
	fmt.Println("\n--- Using default disk (local) ---")
	localKey := "default-disk-file.txt"
	localData := strings.NewReader("This file is on the default disk.")
	err = fs.Upload(ctx, localKey, localData, int64(localData.Len()), filesystem.Public)
	if err != nil {
		log.Printf("Failed to upload to default disk: %v", err)
	} else {
		fmt.Printf("✓ Uploaded '%s' to default disk.\n", localKey)
		fs.Delete(ctx, localKey)
	}

	// --- Use a specific disk by name ("s3_public") ---
	fmt.Println("\n--- Using named disk (s3_public) ---")
	s3Key := "images/avatar.jpg"
	s3Data := strings.NewReader("This is a public S3 file.")
	err = fs.Disk("s3_public").Upload(ctx, s3Key, s3Data, int64(s3Data.Len()), filesystem.Public)
	if err != nil {
		log.Printf("Failed to upload to s3_public disk: %v (check config)", err)
	} else {
		fmt.Printf("✓ Uploaded '%s' to s3_public disk.\n", s3Key)
		// Get public URL
		url, _ := fs.Disk("s3_public").GetURL(ctx, s3Key, filesystem.Public, 0)
		fmt.Printf("✓ Public S3 URL: %s\n", url)
		fs.Disk("s3_public").Delete(ctx, s3Key)
	}

	// --- Use another specific disk ("minio_private") ---
	fmt.Println("\n--- Using named disk (minio_private) ---")
	minioKey := "reports/2023-annual-report.pdf"
	minioData := strings.NewReader("This is a private MinIO document.")
	err = fs.Disk("minio_private").Upload(ctx, minioKey, minioData, int64(minioData.Len()), filesystem.Private)
	if err != nil {
		log.Printf("Failed to upload to minio_private disk: %v (check if MinIO is running)", err)
	} else {
		fmt.Printf("✓ Uploaded '%s' to minio_private disk.\n", minioKey)
		// Get signed URL
		url, _ := fs.Disk("minio_private").GetURL(ctx, minioKey, filesystem.Private, 15*time.Minute)
		fmt.Printf("✓ Signed MinIO URL (valid 15 mins): %s\n", url)
		fs.Disk("minio_private").Delete(ctx, minioKey)
	}
}
