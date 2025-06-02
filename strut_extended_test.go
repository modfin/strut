package strut_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/modfin/strut"
	"github.com/modfin/strut/swag"
	"github.com/modfin/strut/with"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Additional test models
type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty" json-description:"Updated product name" json-min-length:"3" json-max-length:"100"`
	Description *string  `json:"description,omitempty" json-description:"Updated product description"`
	Price       *float64 `json:"price,omitempty" json-description:"Updated product price" json-minimum:"0.01"`
	Categories  []string `json:"categories,omitempty" json-description:"Updated product categories" json-min-items:"1"`
	InStock     *bool    `json:"in_stock,omitempty" json-description:"Updated stock Status"`
}

type DeleteResponse struct {
	Success bool   `json:"success" json-description:"Whether the operation was successful"`
	Message string `json:"Message" json-description:"Operation result Message"`
}

// Handler functions for additional HTTP methods
func UpdateProduct(ctx context.Context, req UpdateProductRequest) (res Product, err error) {
	id := strut.PathParam(ctx, "id")

	// Simulate updating a product
	return Product{
		ID:          id,
		Name:        stringValue(req.Name, "Updated Product"),
		Description: stringValue(req.Description, "Updated description"),
		Price:       floatValue(req.Price, 49.99),
		Categories:  req.Categories,
		InStock:     boolValue(req.InStock, true),
	}, nil
}

func DeleteProduct(ctx context.Context) (res DeleteResponse, err error) {
	id := strut.PathParam(ctx, "id")

	// Simulate deleting a product
	return DeleteResponse{
		Success: true,
		Message: "Product " + id + " successfully deleted",
	}, nil
}

// Helper functions for handling nullable fields
func stringValue(ptr *string, defaultVal string) string {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func floatValue(ptr *float64, defaultVal float64) float64 {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

func boolValue(ptr *bool, defaultVal bool) bool {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// Test helper to create a test server with additional endpoints
func setupExtendedTestServer(t *testing.T) (*strut.Strut, *chi.Mux, *httptest.Server) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r).
		Title("Extended Product API").
		Description("Extended API for managing products with PUT and DELETE").
		Version("1.0.0").
		AddServer("http://localhost:8080", "Development server")

	// Register basic endpoints
	strut.Get(s, "/products", GetProducts,
		with.OperationId("get-products"),
		with.Description("Get all products"),
		with.Tags("products"),
		with.ResponseDescription(200, "List of products"),
	)

	strut.Get(s, "/products/{id}", GetProductByID,
		with.OperationId("get-product-by-id"),
		with.Description("Get a product by its ID"),
		with.Tags("products"),
		with.PathParam[string]("id", "Product ID"),
		with.ResponseDescription(200, "Product details"),
	)

	strut.Post(s, "/products", CreateProduct,
		with.OperationId("create-product"),
		with.Description("Create a new product"),
		with.Tags("products"),
		with.RequestDescription("Product to create"),
		with.ResponseDescription(200, "Created product"),
	)

	// Add PUT endpoint for updating a product
	strut.Put(s, "/products/{id}", UpdateProduct,
		with.OperationId("update-product"),
		with.Description("Update an existing product"),
		with.Summary("Update product"),
		with.Tags("products"),
		with.PathParam[string]("id", "Product ID"),
		with.RequestDescription("Product fields to update"),
		with.ResponseDescription(200, "Updated product"),
		with.Response(404, swag.ResponseOf[strut.Error]("Product not found")),
	)

	// Add DELETE endpoint for deleting a product
	strut.Delete(s, "/products/{id}", DeleteProduct,
		with.OperationId("delete-product"),
		with.Description("Delete an existing product"),
		with.Summary("Delete product"),
		with.Tags("products"),
		with.PathParam[string]("id", "Product ID"),
		with.ResponseDescription(200, "Deletion result"),
		with.Response(404, swag.ResponseOf[strut.Error]("Product not found")),
	)

	r.Get("/.well-known/openapi.yaml", s.SchemaHandlerYAML)
	r.Get("/.well-known/openapi.json", s.SchemaHandlerJSON)

	ts := httptest.NewServer(r)

	return s, r, ts
}

// Tests for PUT and DELETE methods
func TestUpdateProduct(t *testing.T) {
	_, _, ts := setupExtendedTestServer(t)
	defer ts.Close()

	// Create update request
	name := "Updated Test Product"
	price := 59.99
	updateReq := UpdateProductRequest{
		Name:       &name,
		Price:      &price,
		Categories: []string{"updated", "premium"},
	}

	reqBody, err := json.Marshal(updateReq)
	require.NoError(t, err)

	// Create PUT request
	req, err := http.NewRequest(http.MethodPut, ts.URL+"/products/prod-1", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var product Product
	err = json.NewDecoder(resp.Body).Decode(&product)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, "prod-1", product.ID)
	assert.Equal(t, name, product.Name)
	assert.Equal(t, price, product.Price)
	assert.Equal(t, []string{"updated", "premium"}, product.Categories)
}

func TestDeleteProduct(t *testing.T) {
	_, _, ts := setupExtendedTestServer(t)
	defer ts.Close()

	// Create DELETE request
	req, err := http.NewRequest(http.MethodDelete, ts.URL+"/products/prod-1", nil)
	require.NoError(t, err)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var result DeleteResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.True(t, result.Success)
	assert.Equal(t, "Product prod-1 successfully deleted", result.Message)
}

// Tests for OpenAPI spec generation with PUT and DELETE methods
func TestOpenAPISpecWithAllMethods(t *testing.T) {
	_, _, ts := setupExtendedTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/.well-known/openapi.json")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var spec map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&spec)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check paths section
	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok)

	// Check product path with ID parameter
	productPath, ok := paths["/products/{id}"].(map[string]interface{})
	require.True(t, ok)

	// Verify all HTTP methods are present
	assert.Contains(t, productPath, "get")
	assert.Contains(t, productPath, "put")
	assert.Contains(t, productPath, "delete")

	// Check PUT operation
	putOp, ok := productPath["put"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "update-product", putOp["operationId"])

	// Check DELETE operation
	deleteOp, ok := productPath["delete"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "delete-product", deleteOp["operationId"])
}

// Test for validating the complete OpenAPI spec structure
func TestCompleteOpenAPISpec(t *testing.T) {
	s, _, ts := setupExtendedTestServer(t)
	defer ts.Close()

	// Get the OpenAPI spec in YAML format
	resp, err := http.Get(ts.URL + "/.well-known/openapi.yaml")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Parse the YAML
	var spec map[string]interface{}
	err = yaml.Unmarshal(body, &spec)
	require.NoError(t, err)

	// Validate OpenAPI version
	assert.Equal(t, "3.0.3", spec["openapi"])

	// Validate info section
	info, ok := spec["info"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Extended Product API", info["title"])
	assert.Equal(t, "Extended API for managing products with PUT and DELETE", info["description"])
	assert.Equal(t, "1.0.0", info["version"])

	// Validate servers section
	servers, ok := spec["servers"].([]interface{})
	require.True(t, ok)
	require.Len(t, servers, 1)
	server := servers[0].(map[string]interface{})
	assert.Equal(t, "http://localhost:8080", server["url"])
	assert.Equal(t, "Development server", server["description"])

	// Validate components section
	components, ok := spec["components"].(map[string]interface{})
	require.True(t, ok)
	schemas, ok := components["schemas"].(map[string]interface{})
	require.True(t, ok)

	// Check that our models are in the schemas
	assert.Contains(t, schemas, "strut_test_Product")
	assert.Contains(t, schemas, "strut_test_UpdateProductRequest")
	assert.Contains(t, schemas, "strut_test_DeleteResponse")

	// Validate that the direct struct access matches the parsed YAML
	assert.Equal(t, s.Definition.OpenAPI, spec["openapi"])
	assert.Equal(t, s.Definition.Info.Title, info["title"])
	assert.Equal(t, s.Definition.Info.Description, info["description"])
	assert.Equal(t, s.Definition.Info.Version, info["version"])
}

// Test for nullable fields in OpenAPI spec
func TestNullableFieldsInOpenAPISpec(t *testing.T) {
	s, _, _ := setupExtendedTestServer(t)

	// Get the schema for UpdateProductRequest
	updateReqSchema, ok := s.Definition.Components.Schemas["strut_test_UpdateProductRequest"]
	require.True(t, ok)

	// Check that pointer fields are marked as nullable
	props := updateReqSchema.Properties
	assert.True(t, props["name"].Nullable)
	assert.True(t, props["description"].Nullable)
	assert.True(t, props["price"].Nullable)
	assert.True(t, props["in_stock"].Nullable)

	// Array fields should not be nullable (they can be empty)
	assert.False(t, props["categories"].Nullable)
}

// Test for path parameter validation
func TestPathParameterValidation(t *testing.T) {
	s, _, _ := setupExtendedTestServer(t)

	// Get the path item for /products/{id}
	pathItem, ok := s.Definition.Paths["/products/{id}"]
	require.True(t, ok)

	// Check GET operation parameters
	require.NotNil(t, pathItem.Get)
	var getParamFound bool
	for _, param := range pathItem.Get.Parameters {
		if param.Name == "id" && param.In == "path" {
			getParamFound = true
			assert.True(t, param.Required)
		}
	}
	assert.True(t, getParamFound)

	// Check PUT operation parameters
	require.NotNil(t, pathItem.Put)
	var putParamFound bool
	for _, param := range pathItem.Put.Parameters {
		if param.Name == "id" && param.In == "path" {
			putParamFound = true
			assert.True(t, param.Required)
		}
	}
	assert.True(t, putParamFound)

	// Check DELETE operation parameters
	require.NotNil(t, pathItem.Delete)
	var deleteParamFound bool
	for _, param := range pathItem.Delete.Parameters {
		if param.Name == "id" && param.In == "path" {
			deleteParamFound = true
			assert.True(t, param.Required)
		}
	}
	assert.True(t, deleteParamFound)
}

// Test for response Status codes
func TestResponseStatusCodes(t *testing.T) {
	s, _, _ := setupExtendedTestServer(t)

	// Check PUT operation responses
	pathItem, ok := s.Definition.Paths["/products/{id}"]
	require.True(t, ok)
	require.NotNil(t, pathItem.Put)

	// Should have 200 and 404 responses
	assert.Contains(t, pathItem.Put.Responses, "200")
	assert.Contains(t, pathItem.Put.Responses, "404")

	// Check 404 response
	notFoundResp := pathItem.Put.Responses["404"]
	assert.Equal(t, "Product not found", notFoundResp.Description)

	// Check content type
	content, ok := notFoundResp.Content["application/json"]
	require.True(t, ok)
	assert.Contains(t, content.Schema.Properties["message"].Description, "Error message")

	// Check DELETE operation responses
	require.NotNil(t, pathItem.Delete)
	assert.Contains(t, pathItem.Delete.Responses, "200")
	assert.Contains(t, pathItem.Delete.Responses, "404")
}

// Test for tags consistency
func TestTagsConsistency(t *testing.T) {
	s, _, _ := setupExtendedTestServer(t)

	// All operations should have the "products" tag
	for path, pathItem := range s.Definition.Paths {
		if strings.Contains(path, "/products") {
			if pathItem.Get != nil {
				assert.Contains(t, pathItem.Get.Tags, "products")
			}
			if pathItem.Post != nil {
				assert.Contains(t, pathItem.Post.Tags, "products")
			}
			if pathItem.Put != nil {
				assert.Contains(t, pathItem.Put.Tags, "products")
			}
			if pathItem.Delete != nil {
				assert.Contains(t, pathItem.Delete.Tags, "products")
			}
		}
	}
}

// Test for operation IDs uniqueness
func TestOperationIDsUniqueness(t *testing.T) {
	s, _, _ := setupExtendedTestServer(t)

	// Collect all operation IDs
	operationIDs := make(map[string]bool)

	for _, pathItem := range s.Definition.Paths {
		if pathItem.Get != nil {
			assert.False(t, operationIDs[pathItem.Get.OperationID], "Duplicate operation ID: %s", pathItem.Get.OperationID)
			operationIDs[pathItem.Get.OperationID] = true
		}
		if pathItem.Post != nil {
			assert.False(t, operationIDs[pathItem.Post.OperationID], "Duplicate operation ID: %s", pathItem.Post.OperationID)
			operationIDs[pathItem.Post.OperationID] = true
		}
		if pathItem.Put != nil {
			assert.False(t, operationIDs[pathItem.Put.OperationID], "Duplicate operation ID: %s", pathItem.Put.OperationID)
			operationIDs[pathItem.Put.OperationID] = true
		}
		if pathItem.Delete != nil {
			assert.False(t, operationIDs[pathItem.Delete.OperationID], "Duplicate operation ID: %s", pathItem.Delete.OperationID)
			operationIDs[pathItem.Delete.OperationID] = true
		}
	}
}
