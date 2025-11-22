package routes

import (
	"net/http"

	contractHTTP "github.com/donnigundala/dgcore/contracts/http"
	"github.com/donnigundala/dgcore/ctxutil"
	"github.com/donnigundala/dgcore/http/response"
)

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

	// Example: Pagination
	router.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		// Mock users data
		users := []map[string]interface{}{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
			{"id": 3, "name": "Bob Johnson", "email": "bob@example.com"},
		}

		// Pagination metadata
		meta := response.NewPaginationMeta(1, 20, 100)

		response.Paginated(w, users, meta)
	})

	// Example: Error responses
	router.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		response.BadRequest(w, "This is an example error")
	})

	// Example: Not found
	router.Get("/not-found", func(w http.ResponseWriter, r *http.Request) {
		response.NotFound(w, "Resource not found")
	})

	// Example: Validation error
	router.Post("/validate", func(w http.ResponseWriter, r *http.Request) {
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
	})
}
