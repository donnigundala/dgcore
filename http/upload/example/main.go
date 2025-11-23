package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/donnigundala/dgcore/http/response"
	"github.com/donnigundala/dgcore/http/upload"
)

func main() {
	// Single file upload example
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.BadRequest(w, "Method not allowed")
			return
		}

		// Handle upload with validation
		file, err := upload.HandleUpload(r, upload.Config{
			MaxSize:      5 * 1024 * 1024, // 5MB
			AllowedTypes: []string{"image/jpeg", "image/png", "image/gif"},
			FieldName:    "avatar",
		})
		if err != nil {
			response.BadRequest(w, err.Error())
			return
		}

		// Save to local storage
		uploadDir := "./uploads"
		os.MkdirAll(uploadDir, 0755)

		err = os.WriteFile(uploadDir+"/"+file.Filename, file.Data(), 0644)
		if err != nil {
			response.InternalServerError(w, "Failed to save file")
			return
		}

		response.Success(w, map[string]interface{}{
			"filename":     file.Filename,
			"size":         file.Size,
			"content_type": file.ContentType,
			"extension":    file.Extension(),
			"is_image":     file.IsImage(),
		}, "File uploaded successfully")
	})

	// Multiple files upload example
	http.HandleFunc("/upload-multiple", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.BadRequest(w, "Method not allowed")
			return
		}

		// Handle multiple uploads
		files, err := upload.HandleMultipleUploads(r, upload.Config{
			MaxSize:   10 * 1024 * 1024, // 10MB per file
			MaxFiles:  5,
			FieldName: "documents",
		})
		if err != nil {
			response.BadRequest(w, err.Error())
			return
		}

		// Save all files
		uploadDir := "./uploads"
		os.MkdirAll(uploadDir, 0755)

		uploaded := []map[string]interface{}{}
		for _, file := range files {
			err = os.WriteFile(uploadDir+"/"+file.Filename, file.Data(), 0644)
			if err != nil {
				response.InternalServerError(w, "Failed to save file: "+file.Filename)
				return
			}

			uploaded = append(uploaded, map[string]interface{}{
				"filename":     file.Filename,
				"size":         file.Size,
				"content_type": file.ContentType,
			})
		}

		response.Success(w, map[string]interface{}{
			"files": uploaded,
			"count": len(uploaded),
		}, "Files uploaded successfully")
	})

	// Custom validation example
	http.HandleFunc("/upload-image", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.BadRequest(w, "Method not allowed")
			return
		}

		// Handle upload with custom validators
		file, err := upload.HandleUploadWithValidators(r,
			upload.DefaultConfig(),
			upload.SizeValidator(2*1024*1024), // 2MB
			upload.ImageValidator(),
			upload.ExtensionValidator([]string{".jpg", ".jpeg", ".png"}),
		)
		if err != nil {
			response.BadRequest(w, err.Error())
			return
		}

		// Save file
		uploadDir := "./uploads"
		os.MkdirAll(uploadDir, 0755)
		err = os.WriteFile(uploadDir+"/"+file.Filename, file.Data(), 0644)
		if err != nil {
			response.InternalServerError(w, "Failed to save file")
			return
		}

		response.Success(w, map[string]interface{}{
			"filename": file.Filename,
			"size":     file.Size,
		}, "Image uploaded successfully")
	})

	fmt.Println("Upload example server running on :8082")
	fmt.Println("Try:")
	fmt.Println("  curl -F 'avatar=@image.jpg' http://localhost:8082/upload")
	fmt.Println("  curl -F 'documents=@file1.pdf' -F 'documents=@file2.pdf' http://localhost:8082/upload-multiple")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
