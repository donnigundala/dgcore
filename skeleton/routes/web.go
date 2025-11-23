package routes

import (
	"net/http"

	contractHTTP "github.com/donnigundala/dgcore/contracts/http"
	"github.com/donnigundala/dgcore/ctxutil"
	"github.com/donnigundala/dgcore/http/request"
	"github.com/donnigundala/dgcore/http/response"
	"github.com/donnigundala/dgcore/validation"
)

// CreateUserRequest represents a user creation request.
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=18"`
}

// SearchRequest represents a search request.
type SearchRequest struct {
	Query   string `json:"query" form:"query"`
	Page    int    `json:"page" form:"page"`
	PerPage int    `json:"per_page" form:"per_page"`
}

// Register registers the web routes.
func Register(router contractHTTP.Router) {
	// Simple hello route
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, map[string]string{
			"message": "Hello from DG Framework!",
			"version": "1.0.0",
		}, "Welcome to DG Framework")
	})

	// Example: User context and metadata
	router.Get("/me", func(w http.ResponseWriter, r *http.Request) {
		// Demonstrate ctxutil metadata helpers
		clientIP := ctxutil.ClientIP(r)
		userAgent := ctxutil.UserAgent(r)
		requestPath := ctxutil.RequestPath(r)

		response.JSON(w, http.StatusOK, map[string]interface{}{
			"client_ip":   clientIP,
			"user_agent":  userAgent,
			"path":        requestPath,
			"method":      ctxutil.RequestMethod(r),
			"request_url": ctxutil.RequestURL(r),
		})
	})

	// Example: Query parameters
	router.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		// Using request helpers
		query := request.Query(r, "q", "")
		page := request.QueryInt(r, "page", 1)
		perPage := request.QueryInt(r, "per_page", 20)
		sortBy := request.Query(r, "sort", "created_at")

		response.JSON(w, http.StatusOK, map[string]interface{}{
			"query":    query,
			"page":     page,
			"per_page": perPage,
			"sort_by":  sortBy,
		})
	})

	// Example: Pagination with query params
	router.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		// Get pagination params
		page := request.QueryInt(r, "page", 1)
		perPage := request.QueryInt(r, "per_page", 20)

		// Mock users data
		users := []map[string]interface{}{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
			{"id": 3, "name": "Bob Johnson", "email": "bob@example.com"},
		}

		// Pagination metadata
		meta := response.NewPaginationMeta(page, perPage, 100)

		response.Paginated(w, users, meta)
	})

	// Example: JSON body with validation
	router.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserRequest
		validator := validation.NewValidator()

		// Parse and validate JSON
		err := request.JSONWithValidation(r, &req, validator)
		if err != nil {
			if valErr, ok := err.(*validation.Error); ok {
				response.ValidationError(w, valErr.Errors)
				return
			}
			response.BadRequest(w, err.Error())
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

	// Example: Bind query params to struct
	router.Get("/api/search", func(w http.ResponseWriter, r *http.Request) {
		var req SearchRequest
		err := request.BindQuery(r, &req, nil)
		if err != nil {
			response.BadRequest(w, err.Error())
			return
		}

		// Set defaults
		if req.Page == 0 {
			req.Page = 1
		}
		if req.PerPage == 0 {
			req.PerPage = 20
		}

		response.JSON(w, http.StatusOK, map[string]interface{}{
			"query":    req.Query,
			"page":     req.Page,
			"per_page": req.PerPage,
		})
	})

	// Example: Error responses
	router.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		response.BadRequest(w, "This is an example error")
	})

	// Example: Not found
	router.Get("/not-found", func(w http.ResponseWriter, r *http.Request) {
		response.NotFound(w, "Resource not found")
	})

	// Example: Validation error (manual)
	router.Post("/validate", func(w http.ResponseWriter, r *http.Request) {
		// Example validation errors (in real app, this would come from validator)
		validationErrors := map[string]string{
			"email":    "Email is required",
			"password": "Password must be at least 8 characters",
		}
		response.ValidationError(w, validationErrors)
	})

	// API group example
	router.Group(contractHTTP.GroupAttributes{
		Prefix: "/api/v1",
	}, func(r contractHTTP.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusOK, map[string]interface{}{
				"status":    "healthy",
				"timestamp": "2024-01-01T00:00:00Z",
			})
		})

		// Example: Path parameters
		r.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
			userID := request.ParamInt(r, "id")

			response.JSON(w, http.StatusOK, map[string]interface{}{
				"id":    userID,
				"name":  "John Doe",
				"email": "john@example.com",
			})
		})
	})
}
