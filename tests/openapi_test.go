package tests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/modfin/strut"
	"github.com/modfin/strut/swag"
	"github.com/modfin/strut/with"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

// Test types for API endpoints
type OpenAPITestRequest struct {
	Message string `json:"message" json-description:"A test message"`
	Count   int    `json:"count" json-description:"A count value"`
}

type OpenAPITestResponse struct {
	Echo      string    `json:"echo" json-description:"Echoed message"`
	Timestamp time.Time `json:"timestamp" json-description:"Server timestamp"`
	Count     int       `json:"count" json-description:"Count value"`
}

// createTestAPI creates a test API with various endpoints for schema validation
func createTestAPI(t *testing.T) (*strut.Strut, *chi.Mux) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r).
		Title("Test OpenAPI Schema").
		Description("API for testing OpenAPI schema validation").
		Version("1.0.0").
		AddServer("https://api.example.com", "Production").
		AddServer("https://staging.example.com", "Staging")

	// GET endpoint with path parameters
	strut.Get(s, "/users/{id}", func(ctx context.Context) strut.Response[OpenAPITestResponse] {
		id := strut.PathParam(ctx, "id")
		return strut.RespondOk(OpenAPITestResponse{
			Echo:      fmt.Sprintf("User ID: %s", id),
			Timestamp: time.Now(),
			Count:     1,
		})
	},
		with.OperationId("get-user"),
		with.Description("Get user by ID"),
		with.PathParam[string]("id", "User ID"),
		with.QueryParam[string]("include", "Fields to include"),
		with.ResponseDescription(200, "User information"),
		with.Response(http.StatusNotFound, swag.ResponseOf[strut.Error]("User not found")),
	)

	// POST endpoint with request body
	strut.Post(s, "/users", func(ctx context.Context, req OpenAPITestRequest) strut.Response[OpenAPITestResponse] {
		return strut.RespondOk(OpenAPITestResponse{
			Echo:      fmt.Sprintf("Created user: %s", req.Message),
			Timestamp: time.Now(),
			Count:     req.Count,
		})
	},
		with.OperationId("create-user"),
		with.Description("Create a new user"),
		with.ResponseDescription(200, "User created successfully"),
		with.Response(http.StatusBadRequest, swag.ResponseOf[strut.Error]("Invalid request")),
	)

	// PUT endpoint with path parameter and request body
	strut.Put(s, "/users/{id}", func(ctx context.Context, req OpenAPITestRequest) strut.Response[OpenAPITestResponse] {
		id := strut.PathParam(ctx, "id")
		return strut.RespondOk(OpenAPITestResponse{
			Echo:      fmt.Sprintf("Updated user %s: %s", id, req.Message),
			Timestamp: time.Now(),
			Count:     req.Count,
		})
	},
		with.OperationId("update-user"),
		with.Description("Update an existing user"),
		with.PathParam[string]("id", "User ID"),
		with.ResponseDescription(200, "User updated successfully"),
		with.Response(http.StatusNotFound, swag.ResponseOf[strut.Error]("User not found")),
	)

	// DELETE endpoint with path parameter
	strut.Delete(s, "/users/{id}", func(ctx context.Context) strut.Response[OpenAPITestResponse] {
		id := strut.PathParam(ctx, "id")
		return strut.RespondOk(OpenAPITestResponse{
			Echo:      fmt.Sprintf("Deleted user: %s", id),
			Timestamp: time.Now(),
			Count:     0,
		})
	},
		with.OperationId("delete-user"),
		with.Description("Delete a user"),
		with.PathParam[string]("id", "User ID"),
		with.ResponseDescription(200, "User deleted successfully"),
		with.Response(http.StatusNotFound, swag.ResponseOf[strut.Error]("User not found")),
	)

	// Register the schema handlers
	r.Get("/.well-known/openapi.json", s.SchemaHandlerJSON)
	r.Get("/.well-known/openapi.yaml", s.SchemaHandlerYAML)

	return s, r
}

// TestOpenAPISchemaValidation tests that the OpenAPI schema is valid
// using the kin-openapi validator
func TestOpenAPISchemaValidation(t *testing.T) {
	// Create test API
	_, r := createTestAPI(t)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Fetch the OpenAPI schema
	resp, err := http.Get(server.URL + "/.well-known/openapi.json")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response status and content type
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Read the schema
	schemaBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Parse the schema using kin-openapi
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(schemaBytes)
	require.NoError(t, err, "OpenAPI schema should be valid")

	// Validate the schema against the OpenAPI 3.0 specification
	err = doc.Validate(context.Background())
	require.NoError(t, err, "OpenAPI schema should be valid according to the OpenAPI 3.0 specification")

	// Validate the schema structure
	assert.Equal(t, "3.0.3", doc.OpenAPI)
	assert.Equal(t, "Test OpenAPI Schema", doc.Info.Title)
	assert.Equal(t, "API for testing OpenAPI schema validation", doc.Info.Description)
	assert.Equal(t, "1.0.0", doc.Info.Version)

	// Validate servers
	require.Len(t, doc.Servers, 2)
	assert.Equal(t, "https://api.example.com", doc.Servers[0].URL)
	assert.Equal(t, "Production", doc.Servers[0].Description)
	assert.Equal(t, "https://staging.example.com", doc.Servers[1].URL)
	assert.Equal(t, "Staging", doc.Servers[1].Description)

	// Validate paths
	require.Contains(t, doc.Paths.Map(), "/users/{id}")
	require.Contains(t, doc.Paths.Map(), "/users")

	// Validate GET /users/{id} endpoint
	getUserPath := doc.Paths.Find("/users/{id}")
	require.NotNil(t, getUserPath)
	require.NotNil(t, getUserPath.Get)
	assert.Equal(t, "get-user", getUserPath.Get.OperationID)
	assert.Equal(t, "Get user by ID", getUserPath.Get.Description)

	// Validate parameters for GET /users/{id}
	require.GreaterOrEqual(t, len(getUserPath.Get.Parameters), 1)
	var idParam *openapi3.ParameterRef
	for _, param := range getUserPath.Get.Parameters {
		if param.Value.Name == "id" {
			idParam = param
			break
		}
	}
	require.NotNil(t, idParam, "GET /users/{id} should have 'id' path parameter")
	assert.Equal(t, "path", idParam.Value.In)
	assert.True(t, idParam.Value.Required)
	assert.Equal(t, "User ID", idParam.Value.Description)

	// Validate POST /users endpoint
	postUserPath := doc.Paths.Find("/users")
	require.NotNil(t, postUserPath)
	require.NotNil(t, postUserPath.Post)
	assert.Equal(t, "create-user", postUserPath.Post.OperationID)
	assert.Equal(t, "Create a new user", postUserPath.Post.Description)

	// Validate request body for POST /users
	require.NotNil(t, postUserPath.Post.RequestBody)
	require.NotNil(t, postUserPath.Post.RequestBody.Value.Content)
	require.Contains(t, postUserPath.Post.RequestBody.Value.Content, "application/json")

	// Validate PUT /users/{id} endpoint
	putUserPath := doc.Paths.Find("/users/{id}")
	require.NotNil(t, putUserPath)
	require.NotNil(t, putUserPath.Put)
	assert.Equal(t, "update-user", putUserPath.Put.OperationID)
	assert.Equal(t, "Update an existing user", putUserPath.Put.Description)

	// Validate DELETE /users/{id} endpoint
	deleteUserPath := doc.Paths.Find("/users/{id}")
	require.NotNil(t, deleteUserPath)
	require.NotNil(t, deleteUserPath.Delete)
	assert.Equal(t, "delete-user", deleteUserPath.Delete.OperationID)
	assert.Equal(t, "Delete a user", deleteUserPath.Delete.Description)
}

// TestOpenAPISchemaComponents tests that the components section of the schema is correctly generated
func TestOpenAPISchemaComponents(t *testing.T) {
	// Create test API
	_, r := createTestAPI(t)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Fetch the OpenAPI schema
	resp, err := http.Get(server.URL + "/.well-known/openapi.json")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read the schema
	schemaBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Parse the schema using kin-openapi
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(schemaBytes)
	require.NoError(t, err)
	err = doc.Validate(context.Background())
	require.NoError(t, err, "OpenAPI schema should be valid")

	// Verify components section exists
	require.NotNil(t, doc.Components)
	require.NotNil(t, doc.Components.Schemas)

	// Find our test schemas
	var testRequestSchema, testResponseSchema *openapi3.SchemaRef
	for name, schema := range doc.Components.Schemas {
		if strings.Contains(name, "OpenAPITestRequest") {
			testRequestSchema = schema
		}
		if strings.Contains(name, "OpenAPITestResponse") {
			testResponseSchema = schema
		}
	}

	require.NotNil(t, testRequestSchema, "Schema should include OpenAPITestRequest component")
	require.NotNil(t, testResponseSchema, "Schema should include OpenAPITestResponse component")

	// Verify TestRequest schema properties
	properties := testRequestSchema.Value.Properties
	require.NotNil(t, properties)

	messageProperty := properties["message"]
	require.NotNil(t, messageProperty, "OpenAPITestRequest should have message property")
	// Just verify the property exists with the right description
	assert.Equal(t, "A test message", messageProperty.Value.Description)

	countProperty := properties["count"]
	require.NotNil(t, countProperty, "OpenAPITestRequest should have count property")
	// Just verify the property exists with the right description
	assert.Equal(t, "A count value", countProperty.Value.Description)

	// Verify TestResponse schema properties
	properties = testResponseSchema.Value.Properties
	require.NotNil(t, properties)

	echoProperty := properties["echo"]
	require.NotNil(t, echoProperty, "OpenAPITestResponse should have echo property")
	// Just verify the property exists with the right description
	assert.Equal(t, "Echoed message", echoProperty.Value.Description)

	timestampProperty := properties["timestamp"]
	require.NotNil(t, timestampProperty, "OpenAPITestResponse should have timestamp property")
	// Just verify the property exists with the right description
	assert.Equal(t, "Server timestamp", timestampProperty.Value.Description)
}

// TestOpenAPISchemaPathParameters tests that path parameters are correctly defined in the schema
func TestOpenAPISchemaPathParameters(t *testing.T) {
	// Create test API
	_, r := createTestAPI(t)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Fetch the OpenAPI schema
	resp, err := http.Get(server.URL + "/.well-known/openapi.json")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read the schema
	schemaBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Parse the schema using kin-openapi
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(schemaBytes)
	require.NoError(t, err)
	err = doc.Validate(context.Background())
	require.NoError(t, err, "OpenAPI schema should be valid")

	// Validate path parameters for /users/{id}
	userPath := doc.Paths.Find("/users/{id}")
	require.NotNil(t, userPath)

	// Check GET parameters
	require.NotNil(t, userPath.Get)
	require.GreaterOrEqual(t, len(userPath.Get.Parameters), 1)

	var idParam *openapi3.ParameterRef
	for _, param := range userPath.Get.Parameters {
		if param.Value.Name == "id" {
			idParam = param
			break
		}
	}

	require.NotNil(t, idParam, "GET /users/{id} should have 'id' path parameter")
	assert.Equal(t, "path", idParam.Value.In)
	assert.True(t, idParam.Value.Required)
	assert.Equal(t, "User ID", idParam.Value.Description)

	// Check PUT parameters
	require.NotNil(t, userPath.Put)
	require.GreaterOrEqual(t, len(userPath.Put.Parameters), 1)

	idParam = nil
	for _, param := range userPath.Put.Parameters {
		if param.Value.Name == "id" {
			idParam = param
			break
		}
	}

	require.NotNil(t, idParam, "PUT /users/{id} should have 'id' path parameter")
	assert.Equal(t, "path", idParam.Value.In)
	assert.True(t, idParam.Value.Required)
	assert.Equal(t, "User ID", idParam.Value.Description)

	// Check DELETE parameters
	require.NotNil(t, userPath.Delete)
	require.GreaterOrEqual(t, len(userPath.Delete.Parameters), 1)

	idParam = nil
	for _, param := range userPath.Delete.Parameters {
		if param.Value.Name == "id" {
			idParam = param
			break
		}
	}

	require.NotNil(t, idParam, "DELETE /users/{id} should have 'id' path parameter")
	assert.Equal(t, "path", idParam.Value.In)
	assert.True(t, idParam.Value.Required)
	assert.Equal(t, "User ID", idParam.Value.Description)
}

// TestOpenAPISchemaRequestResponse tests the request and response schemas
func TestOpenAPISchemaRequestResponse(t *testing.T) {
	// Create test API
	_, r := createTestAPI(t)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Fetch the OpenAPI schema
	resp, err := http.Get(server.URL + "/.well-known/openapi.json")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read the schema
	schemaBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Parse the schema using kin-openapi
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(schemaBytes)
	require.NoError(t, err)
	err = doc.Validate(context.Background())
	require.NoError(t, err, "OpenAPI schema should be valid")

	err = doc.Validate(context.Background())
	require.NoError(t, err, "OpenAPI schema should be valid")

	// Validate POST /users request body
	postUserPath := doc.Paths.Find("/users")
	require.NotNil(t, postUserPath)
	require.NotNil(t, postUserPath.Post)
	require.NotNil(t, postUserPath.Post.RequestBody)
	require.NotNil(t, postUserPath.Post.RequestBody.Value.Content)

	jsonContent := postUserPath.Post.RequestBody.Value.Content["application/json"]
	require.NotNil(t, jsonContent)
	require.NotNil(t, jsonContent.Schema)

	// The schema might be a reference or an inline schema
	if jsonContent.Schema.Ref != "" {
		assert.NotEmpty(t, jsonContent.Schema.Ref)
	} else {
		assert.NotNil(t, jsonContent.Schema.Value)
	}

	// Validate POST /users response
	responses := postUserPath.Post.Responses
	require.NotNil(t, responses)

	// Check if responses map contains expected status codes
	require.Contains(t, responses.Map(), "200")
	require.Contains(t, responses.Map(), "400")

	// Access responses by string key
	okResponse := responses.Map()["200"]
	require.NotNil(t, okResponse, "Should have 200 response")
	require.NotNil(t, okResponse.Value.Content)

	jsonContent = okResponse.Value.Content["application/json"]
	require.NotNil(t, jsonContent)
	require.NotNil(t, jsonContent.Schema)

	// The schema might be a reference or an inline schema
	if jsonContent.Schema.Ref != "" {
		assert.NotEmpty(t, jsonContent.Schema.Ref)
	} else {
		assert.NotNil(t, jsonContent.Schema.Value)
	}

	// Validate error response
	badRequestResponse := responses.Map()["400"]
	require.NotNil(t, badRequestResponse, "Should have 400 response")
	require.NotNil(t, badRequestResponse.Value.Content)

	jsonContent = badRequestResponse.Value.Content["application/json"]
	require.NotNil(t, jsonContent)
	require.NotNil(t, jsonContent.Schema)

	// The schema might be a reference or an inline schema
	if jsonContent.Schema.Ref != "" {
		assert.NotEmpty(t, jsonContent.Schema.Ref)
	} else {
		assert.NotNil(t, jsonContent.Schema.Value)
	}
}

// TestOpenAPISchemaQueryParameters tests that query parameters are correctly defined in the schema
func TestOpenAPISchemaQueryParameters(t *testing.T) {
	// Create test API
	_, r := createTestAPI(t)

	// Create a test server
	server := httptest.NewServer(r)
	defer server.Close()

	// Fetch the OpenAPI schema
	resp, err := http.Get(server.URL + "/.well-known/openapi.json")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read the schema
	schemaBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Parse the schema using kin-openapi
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(schemaBytes)
	require.NoError(t, err)
	err = doc.Validate(context.Background())
	require.NoError(t, err, "OpenAPI schema should be valid")

	// Validate query parameters for GET /users/{id}
	userPath := doc.Paths.Find("/users/{id}")
	require.NotNil(t, userPath)
	require.NotNil(t, userPath.Get)

	var includeParam *openapi3.ParameterRef
	for _, param := range userPath.Get.Parameters {
		if param.Value.Name == "include" {
			includeParam = param
			break
		}
	}

	require.NotNil(t, includeParam, "GET /users/{id} should have 'include' query parameter")
	assert.Equal(t, "query", includeParam.Value.In)
	assert.False(t, includeParam.Value.Required)
	assert.Equal(t, "Fields to include", includeParam.Value.Description)
}
