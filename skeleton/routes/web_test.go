package routes_test

import (
	"testing"

	coreHTTP "github.com/donnigundala/dgcore/http"
	dgTesting "github.com/donnigundala/dgcore/testing"

	"example-app/routes"
)

func TestHomeRoute(t *testing.T) {
	// Create router and register routes
	router := coreHTTP.NewRouter()
	routes.Register(router)

	// Create HTTP test client
	client := dgTesting.NewHTTPClient(router)

	// Make GET request to home route
	resp := client.Get("/").WithT(t)

	// Assert response
	resp.AssertStatus(200)
	resp.AssertHeader("Content-Type", "application/json")
	resp.AssertBodyContains("Hello from DG Framework!")
	resp.AssertBodyContains("version")
}

func TestUsersRoute(t *testing.T) {
	// Create router and register routes
	router := coreHTTP.NewRouter()
	routes.Register(router)

	// Create HTTP test client
	client := dgTesting.NewHTTPClient(router)

	// Make GET request to users route
	resp := client.Get("/users").WithT(t)

	// Assert response
	resp.AssertStatus(200)
	resp.AssertHeader("Content-Type", "application/json")

	// Decode JSON response
	var response map[string]interface{}
	err := resp.DecodeJSON(&response)
	dgTesting.Assert.NoError(t, err)

	// Assert JSON structure
	dgTesting.Assert.Equal(t, true, response["success"])
	dgTesting.Assert.NotNil(t, response["data"])
	dgTesting.Assert.NotNil(t, response["meta"])
}

func TestMetadataRoute(t *testing.T) {
	// Create router and register routes
	router := coreHTTP.NewRouter()
	routes.Register(router)

	// Create HTTP test client
	client := dgTesting.NewHTTPClient(router)

	// Make GET request with custom header
	resp := client.
		WithHeader("User-Agent", "Test Client").
		Get("/me").
		WithT(t)

	// Assert response
	resp.AssertStatus(200)

	// Decode JSON response
	var response map[string]interface{}
	err := resp.DecodeJSON(&response)
	dgTesting.Assert.NoError(t, err)

	// Assert metadata
	dgTesting.Assert.NotNil(t, response["client_ip"])
	dgTesting.Assert.Equal(t, "Test Client", response["user_agent"])
}

func TestErrorRoute(t *testing.T) {
	// Create router and register routes
	router := coreHTTP.NewRouter()
	routes.Register(router)

	// Create HTTP test client
	client := dgTesting.NewHTTPClient(router)

	// Make GET request to error route
	resp := client.Get("/error").WithT(t)

	// Assert error response
	resp.AssertStatus(400)
	resp.AssertBodyContains("error")
}

func TestNotFoundRoute(t *testing.T) {
	// Create router and register routes
	router := coreHTTP.NewRouter()
	routes.Register(router)

	// Create HTTP test client
	client := dgTesting.NewHTTPClient(router)

	// Make GET request to not-found route
	resp := client.Get("/not-found").WithT(t)

	// Assert not found response
	resp.AssertStatus(404)
	resp.AssertBodyContains("not found")
}

func TestHealthRoutes(t *testing.T) {
	// Create router and register routes
	router := coreHTTP.NewRouter()
	routes.Register(router)

	// Create HTTP test client
	client := dgTesting.NewHTTPClient(router)

	// Test API health endpoint
	resp := client.Get("/api/v1/health").WithT(t)
	resp.AssertStatus(200)
	resp.AssertBodyContains("healthy")
}

// Example of testing with assertions
func TestAssertions(t *testing.T) {
	// String assertions
	dgTesting.Assert.Equal(t, "hello", "hello")
	dgTesting.Assert.NotEqual(t, "hello", "world")
	dgTesting.Assert.Contains(t, "hello world", "world")
	dgTesting.Assert.HasPrefix(t, "hello world", "hello")
	dgTesting.Assert.HasSuffix(t, "hello world", "world")

	// Boolean assertions
	dgTesting.Assert.True(t, true)
	dgTesting.Assert.False(t, false)

	// Nil assertions
	dgTesting.Assert.Nil(t, nil)
	dgTesting.Assert.NotNil(t, "not nil")

	// Collection assertions
	slice := []string{"a", "b", "c"}
	dgTesting.Assert.Len(t, slice, 3)
	dgTesting.Assert.NotEmpty(t, slice)

	emptySlice := []string{}
	dgTesting.Assert.Empty(t, emptySlice)

	// Error assertions
	dgTesting.Assert.NoError(t, nil)
}
