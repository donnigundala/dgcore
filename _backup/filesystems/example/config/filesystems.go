package config

//func init() {
//	config.Add("filesystem", map[string]any{
//		// Default filesystem disk name (can be "local", "s3", or "minio")
//		"default": config.Env("FILESYSTEM_DISK", "local"),
//
//		// All configured disks
//		"disks": map[string]any{
//
//			// ─────────────────────────────
//			// LOCAL DISK CONFIGURATION
//			// ─────────────────────────────
//			"local": map[string]any{
//				"driver": "local",
//				"config": map[string]any{
//					// Absolute or relative path to store uploaded files
//					"basePath": config.Env("FILESYSTEM_LOCAL_PATH", "./storage/app"),
//
//					// Public URL base (for generating GetURL())
//					"baseURL": config.Env("FILESYSTEM_LOCAL_URL", "http://localhost:8080/storage"),
//				},
//			},
//
//			// ─────────────────────────────
//			// MINIO DISK CONFIGURATION
//			// ─────────────────────────────
//			"minio": map[string]any{
//				"driver": "minio",
//				"config": map[string]any{
//					"endpoint":        config.Env("MINIO_ENDPOINT", "127.0.0.1:9000"),
//					"accessKeyID":     config.Env("MINIO_ACCESS_KEY", "minioadmin"),
//					"secretAccessKey": config.Env("MINIO_SECRET_KEY", "minioadmin"),
//					"useSSL":          config.Env("MINIO_USE_SSL", "false"),
//					"bucket":          config.Env("MINIO_BUCKET", "local-bucket"),
//					"baseURL":         config.Env("MINIO_BASE_URL", "http://127.0.0.1:9000/local-bucket"),
//				},
//			},
//
//			// ─────────────────────────────
//			// AWS S3 DISK CONFIGURATION
//			// ─────────────────────────────
//			"s3": map[string]any{
//				"driver": "s3",
//				"config": map[string]any{
//					"bucket":    config.Env("AWS_BUCKET", "my-app-bucket"),
//					"region":    config.Env("AWS_REGION", "ap-southeast-1"),
//					"accessKey": config.Env("AWS_ACCESS_KEY_ID", ""),
//					"secretKey": config.Env("AWS_SECRET_ACCESS_KEY", ""),
//					"baseURL":   config.Env("AWS_BASE_URL", "https://my-app-bucket.s3.amazonaws.com"),
//				},
//			},
//		},
//	})
//}
