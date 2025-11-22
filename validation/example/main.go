package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/donnigundala/dgcore/http/response"
	"github.com/donnigundala/dgcore/validation"
)

// CreateUserRequest represents the request body for creating a user.
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Age      int    `json:"age" validate:"required,gte=18,lte=120"`
}

func main() {
	// Initialize validator
	validator := validation.NewValidator()

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.BadRequest(w, "Method not allowed")
			return
		}

		// Parse request body
		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "Invalid JSON")
			return
		}

		// Validate request
		if err := validator.ValidateStruct(context.Background(), &req); err != nil {
			// Check if it's a validation error
			if valErr, ok := err.(*validation.Error); ok {
				response.ValidationError(w, valErr.Errors)
				return
			}
			response.InternalServerError(w, "Validation failed")
			return
		}

		// Success - create user (mock)
		response.Created(w, map[string]interface{}{
			"id":    123,
			"name":  req.Name,
			"email": req.Email,
			"age":   req.Age,
		}, "/users/123")
	})

	fmt.Println("Validation example server running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
