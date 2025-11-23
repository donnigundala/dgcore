package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contractHTTP "github.com/donnigundala/dgcore/contracts/http"
)

// HTTPClient is a test HTTP client for making requests to a router.
type HTTPClient struct {
	router  contractHTTP.Router
	cookies []*http.Cookie
	headers map[string]string
}

// NewHTTPClient creates a new HTTP test client.
func NewHTTPClient(router contractHTTP.Router) *HTTPClient {
	return &HTTPClient{
		router:  router,
		cookies: make([]*http.Cookie, 0),
		headers: make(map[string]string),
	}
}

// WithHeader sets a header for all subsequent requests.
func (c *HTTPClient) WithHeader(key, value string) *HTTPClient {
	c.headers[key] = value
	return c
}

// WithCookie sets a cookie for all subsequent requests.
func (c *HTTPClient) WithCookie(cookie *http.Cookie) *HTTPClient {
	c.cookies = append(c.cookies, cookie)
	return c
}

// Get makes a GET request.
func (c *HTTPClient) Get(path string) *Response {
	return c.request(http.MethodGet, path, nil)
}

// Post makes a POST request with JSON body.
func (c *HTTPClient) Post(path string, body interface{}) *Response {
	return c.requestJSON(http.MethodPost, path, body)
}

// Put makes a PUT request with JSON body.
func (c *HTTPClient) Put(path string, body interface{}) *Response {
	return c.requestJSON(http.MethodPut, path, body)
}

// Patch makes a PATCH request with JSON body.
func (c *HTTPClient) Patch(path string, body interface{}) *Response {
	return c.requestJSON(http.MethodPatch, path, body)
}

// Delete makes a DELETE request.
func (c *HTTPClient) Delete(path string) *Response {
	return c.request(http.MethodDelete, path, nil)
}

// request makes an HTTP request.
func (c *HTTPClient) request(method, path string, body io.Reader) *Response {
	req := httptest.NewRequest(method, path, body)

	// Set headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Set cookies
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}

	// Create response recorder
	rec := httptest.NewRecorder()

	// Serve the request
	c.router.ServeHTTP(rec, req)

	// Extract cookies from response
	responseCookies := rec.Result().Cookies()

	return &Response{
		StatusCode: rec.Code,
		Body:       rec.Body.Bytes(),
		Headers:    rec.Header(),
		Cookies:    responseCookies,
	}
}

// requestJSON makes an HTTP request with JSON body.
func (c *HTTPClient) requestJSON(method, path string, body interface{}) *Response {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return &Response{
			StatusCode: 0,
			Body:       []byte(err.Error()),
			err:        err,
		}
	}

	// Set Content-Type header
	c.WithHeader("Content-Type", "application/json")

	return c.request(method, path, bytes.NewReader(jsonBody))
}

// Response represents an HTTP response.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Cookies    []*http.Cookie
	err        error
	t          *testing.T
}

// WithT sets the testing.T for assertions.
func (r *Response) WithT(t *testing.T) *Response {
	r.t = t
	return r
}

// AssertStatus asserts the response status code.
func (r *Response) AssertStatus(expected int) *Response {
	if r.t != nil {
		if r.StatusCode != expected {
			r.t.Errorf("Expected status %d, got %d", expected, r.StatusCode)
		}
	}
	return r
}

// AssertJSON asserts the response body matches the expected JSON.
func (r *Response) AssertJSON(expected interface{}) *Response {
	if r.t != nil {
		var actual interface{}
		if err := json.Unmarshal(r.Body, &actual); err != nil {
			r.t.Errorf("Failed to decode JSON: %v", err)
			return r
		}

		expectedJSON, _ := json.Marshal(expected)
		actualJSON, _ := json.Marshal(actual)

		if string(expectedJSON) != string(actualJSON) {
			r.t.Errorf("Expected JSON:\n%s\n\nGot:\n%s", expectedJSON, actualJSON)
		}
	}
	return r
}

// AssertHeader asserts a response header value.
func (r *Response) AssertHeader(key, expected string) *Response {
	if r.t != nil {
		actual := r.Headers.Get(key)
		if actual != expected {
			r.t.Errorf("Expected header %s: %s, got: %s", key, expected, actual)
		}
	}
	return r
}

// AssertHeaderContains asserts a response header contains a value.
func (r *Response) AssertHeaderContains(key, expected string) *Response {
	if r.t != nil {
		actual := r.Headers.Get(key)
		if !strings.Contains(actual, expected) {
			r.t.Errorf("Expected header %s to contain: %s, got: %s", key, expected, actual)
		}
	}
	return r
}

// AssertCookie asserts a cookie value.
func (r *Response) AssertCookie(name, expected string) *Response {
	if r.t != nil {
		for _, cookie := range r.Cookies {
			if cookie.Name == name {
				if cookie.Value != expected {
					r.t.Errorf("Expected cookie %s: %s, got: %s", name, expected, cookie.Value)
				}
				return r
			}
		}
		r.t.Errorf("Cookie %s not found", name)
	}
	return r
}

// AssertBodyContains asserts the response body contains a string.
func (r *Response) AssertBodyContains(expected string) *Response {
	if r.t != nil {
		if !strings.Contains(string(r.Body), expected) {
			r.t.Errorf("Expected body to contain: %s\n\nGot:\n%s", expected, r.Body)
		}
	}
	return r
}

// DecodeJSON decodes the response body into the given value.
func (r *Response) DecodeJSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// String returns the response body as a string.
func (r *Response) String() string {
	return string(r.Body)
}
