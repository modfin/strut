package strut_test

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/modfin/strut"
	"github.com/modfin/strut/swag"
	"github.com/modfin/strut/with"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestStrut_New demonstrates how to create a new Strut instance
// and configure its basic properties.
func TestStrut_New(t *testing.T) {
	// Create a new router and logger
	r := chi.NewRouter()
	logger := slog.Default()

	// Create a new Strut instance
	s := strut.New(logger, r)

	// Configure the API metadata
	s.Title("Test API")
	s.Description("API for testing purposes")
	s.Version("1.0.0")
	s.AddServer("http://localhost:8080", "Test server")

	// Verify the configuration
	assert.Equal(t, "Test API", s.Definition.Info.Title)
	assert.Equal(t, "API for testing purposes", s.Definition.Info.Description)
	assert.Equal(t, "1.0.0", s.Definition.Info.Version)
	assert.Equal(t, "http://localhost:8080", s.Definition.Servers[0].URL)
	assert.Equal(t, "Test server", s.Definition.Servers[0].Description)
}

// TestStrut_SchemaHandlers tests the OpenAPI schema handlers
// that serve the API documentation in YAML and JSON formats.
func TestStrut_SchemaHandlers(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)
	s.Title("Test API")

	// Test YAML schema handler
	req := httptest.NewRequest("GET", "/.well-known/openapi.yaml", nil)
	w := httptest.NewRecorder()
	s.SchemaHandlerYAML(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/yaml", resp.Header.Get("Content-Type"))

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Test API")

	// Test JSON schema handler
	req = httptest.NewRequest("GET", "/.well-known/openapi.json", nil)
	w = httptest.NewRecorder()
	s.SchemaHandlerJSON(w, req)

	resp = w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, _ = io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Test API")
}

// Simple types for testing endpoints
type TestRequest struct {
	Message string `json:"message" json-description:"A test message"`
}

type TestResponse struct {
	Echo string `json:"echo" json-description:"Echoed message"`
	Time string `json:"time" json-description:"Server time"`
}

// TestStrut_Get demonstrates how to create a GET endpoint
// and test its functionality.
func TestStrut_Get(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Define a handler for GET requests
	handler := func(ctx context.Context) strut.Response[TestResponse] {
		name := strut.PathParam(ctx, "name")
		return strut.RespondOk(TestResponse{
			Echo: "Hello, " + name,
			Time: time.Now().Format(time.RFC3339),
		})
	}

	// Register the GET endpoint
	strut.Get(s, "/greet/{name}", handler,
		with.OperationId("get-greeting"),
		with.Description("Get a greeting by name"),
		with.PathParam[string]("name", "Person's name"),
		with.ResponseDescription(200, "A greeting response"),
	)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Make a request to the endpoint
	resp, err := http.Get(server.URL + "/greet/World")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result TestResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "Hello, World", result.Echo)
	assert.NotEmpty(t, result.Time)

	// Verify the OpenAPI definition
	path := s.Definition.Paths["/greet/{name}"]
	assert.NotNil(t, path)
	assert.NotNil(t, path.Get)
	assert.Equal(t, "get-greeting", path.Get.OperationID)
	assert.Equal(t, "Get a greeting by name", path.Get.Description)
}

// TestStrut_Post demonstrates how to create a POST endpoint
// and test its functionality with request body.
func TestStrut_Post(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Define a handler for POST requests
	handler := func(ctx context.Context, req TestRequest) strut.Response[TestResponse] {
		return strut.RespondOk(TestResponse{
			Echo: "You said: " + req.Message,
			Time: time.Now().Format(time.RFC3339),
		})
	}

	// Register the POST endpoint
	strut.Post(s, "/echo", handler,
		with.OperationId("post-echo"),
		with.Description("Echo a message"),
		with.ResponseDescription(200, "Echoed message"),
	)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Create a request body
	reqBody := strings.NewReader(`{"message": "Hello, Server!"}`)

	// Make a request to the endpoint
	resp, err := http.Post(server.URL+"/echo", "application/json", reqBody)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result TestResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "You said: Hello, Server!", result.Echo)
	assert.NotEmpty(t, result.Time)

	// Verify the OpenAPI definition
	path := s.Definition.Paths["/echo"]
	assert.NotNil(t, path)
	assert.NotNil(t, path.Post)
	assert.Equal(t, "post-echo", path.Post.OperationID)
}

// TestStrut_Put demonstrates how to create a PUT endpoint
// and test its functionality with request body.
func TestStrut_Put(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Define a handler for PUT requests
	handler := func(ctx context.Context, req TestRequest) strut.Response[TestResponse] {
		return strut.RespondOk(TestResponse{
			Echo: "Updated: " + req.Message,
			Time: time.Now().Format(time.RFC3339),
		})
	}

	// Register the PUT endpoint
	strut.Put(s, "/update/{id}", handler,
		with.OperationId("put-update"),
		with.Description("Update a resource"),
		with.PathParam[string]("id", "Resource ID"),
		with.ResponseDescription(200, "Updated resource"),
	)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Create a request body
	reqBody := strings.NewReader(`{"message": "New content"}`)

	// Create a PUT request
	req, err := http.NewRequest(http.MethodPut, server.URL+"/update/123", reqBody)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result TestResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "Updated: New content", result.Echo)
	assert.NotEmpty(t, result.Time)

	// Verify the OpenAPI definition
	path := s.Definition.Paths["/update/{id}"]
	assert.NotNil(t, path)
	assert.NotNil(t, path.Put)
	assert.Equal(t, "put-update", path.Put.OperationID)
}

// TestStrut_Delete demonstrates how to create a DELETE endpoint
// and test its functionality.
func TestStrut_Delete(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Define a handler for DELETE requests
	handler := func(ctx context.Context) strut.Response[TestResponse] {
		id := strut.PathParam(ctx, "id")
		return strut.RespondOk(TestResponse{
			Echo: "Deleted resource " + id,
			Time: time.Now().Format(time.RFC3339),
		})
	}

	// Register the DELETE endpoint
	strut.Delete(s, "/resource/{id}", handler,
		with.OperationId("delete-resource"),
		with.Description("Delete a resource"),
		with.PathParam[string]("id", "Resource ID"),
		with.ResponseDescription(200, "Deletion confirmation"),
	)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Create a DELETE request
	req, err := http.NewRequest(http.MethodDelete, server.URL+"/resource/456", nil)
	require.NoError(t, err)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result TestResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "Deleted resource 456", result.Echo)
	assert.NotEmpty(t, result.Time)

	// Verify the OpenAPI definition
	path := s.Definition.Paths["/resource/{id}"]
	assert.NotNil(t, path)
	assert.NotNil(t, path.Delete)
	assert.Equal(t, "delete-resource", path.Delete.OperationID)
}

// TestStrut_ErrorHandling demonstrates how to handle errors
// in endpoint handlers.
func TestStrut_ErrorHandling(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Define a handler that returns an error
	handler := func(ctx context.Context) strut.Response[TestResponse] {
		id := strut.PathParam(ctx, "id")
		if id == "0" {
			return strut.RespondError[TestResponse](http.StatusBadRequest, "Invalid ID")
		}
		return strut.RespondOk(TestResponse{
			Echo: "Resource " + id,
			Time: time.Now().Format(time.RFC3339),
		})
	}

	// Register the endpoint
	strut.Get(s, "/resource/{id}", handler,
		with.OperationId("get-resource"),
		with.Description("Get a resource"),
		with.PathParam[string]("id", "Resource ID"),
		with.ResponseDescription(200, "Resource details"),
		with.Response(http.StatusBadRequest, swag.ResponseOf[strut.Error]("Invalid resource ID")),
	)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Test successful request
	resp, err := http.Get(server.URL + "/resource/123")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test error response
	resp, err = http.Get(server.URL + "/resource/0")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errorResp strut.Error
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Equal(t, "Invalid ID", errorResp.Error)
	assert.Equal(t, http.StatusBadRequest, errorResp.StatusCode)
}

// TestStrut_CustomResponse demonstrates how to use custom response handlers.
func TestStrut_CustomResponse(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Define a handler with custom response
	handler := func(ctx context.Context) strut.Response[TestResponse] {
		// Set custom header
		w := strut.HTTPResponseWriter(ctx)
		w.Header().Set("X-Custom-Header", "custom-value")

		// Return a custom response handler
		return strut.RespondFunc[TestResponse](func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			return json.NewEncoder(w).Encode(map[string]string{
				"custom": "response",
				"time":   time.Now().Format(time.RFC3339),
			})
		})
	}

	// Register the endpoint
	strut.Get(s, "/custom", handler,
		with.OperationId("get-custom"),
		with.Description("Get a custom response"),
		with.ResponseDescription(200, "Custom response"),
	)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL + "/custom")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "custom-value", resp.Header.Get("X-Custom-Header"))

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "response", result["custom"])
	assert.NotEmpty(t, result["time"])
}

// TestStrut_ContextHelpers tests the context helper functions.
func TestStrut_ContextHelpers(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Define a handler that uses context helpers
	handler := func(ctx context.Context) strut.Response[TestResponse] {
		// Get path parameter
		name := strut.PathParam(ctx, "name")

		// Get query parameter
		greeting := strut.QueryParam(ctx, "greeting")
		if greeting == "" {
			greeting = "Hello"
		}

		// Access HTTP request
		req := strut.HTTPRequest(ctx)
		userAgent := req.Header.Get("User-Agent")

		return strut.RespondOk(TestResponse{
			Echo: greeting + ", " + name + " (from " + userAgent + ")",
			Time: time.Now().Format(time.RFC3339),
		})
	}

	// Register the endpoint
	strut.Get(s, "/hello/{name}", handler,
		with.OperationId("get-hello"),
		with.Description("Get a personalized greeting"),
		with.PathParam[string]("name", "Person's name"),
		with.QueryParam[string]("greeting", "Custom greeting"),
		with.ResponseDescription(200, "Personalized greeting"),
	)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Create a custom request
	req, err := http.NewRequest("GET", server.URL+"/hello/World?greeting=Hola", nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "TestClient")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result TestResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result.Echo, "Hola, World")
	assert.Contains(t, result.Echo, "TestClient")
	assert.NotEmpty(t, result.Time)
}
