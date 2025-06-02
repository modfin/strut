package strut_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/modfin/strut"
	"github.com/modfin/strut/schema"
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

// Test models
type Product struct {
	ID          string    `json:"id" json-description:"Unique product identifier"`
	Name        string    `json:"name" json-description:"Product name" json-min-length:"3" json-max-length:"100"`
	Description string    `json:"description,omitempty" json-description:"Product description"`
	Price       float64   `json:"price" json-description:"Product price" json-minimum:"0.01"`
	Categories  []string  `json:"categories" json-description:"Product categories" json-min-items:"1"`
	CreatedAt   time.Time `json:"created_at" json-description:"Creation timestamp"`
	InStock     bool      `json:"in_stock" json-description:"Whether the product is in stock"`
}

type ProductList struct {
	Products []Product `json:"products" json-description:"List of products"`
	Total    int       `json:"total" json-description:"Total number of products"`
}

type CreateProductRequest struct {
	Name        string   `json:"name" json-description:"Product name" json-min-length:"3" json-max-length:"100"`
	Description string   `json:"description,omitempty" json-description:"Product description"`
	Price       float64  `json:"price" json-description:"Product price" json-minimum:"0.01"`
	Categories  []string `json:"categories" json-description:"Product categories" json-min-items:"1"`
}

// Handler functions
func GetProducts(ctx context.Context) (res ProductList, err error) {
	// In a real application, this would fetch from a database
	products := []Product{
		{
			ID:          "prod-1",
			Name:        "Test Product 1",
			Description: "This is a test product",
			Price:       19.99,
			Categories:  []string{"test", "example"},
			CreatedAt:   time.Now(),
			InStock:     true,
		},
		{
			ID:          "prod-2",
			Name:        "Test Product 2",
			Description: "Another test product",
			Price:       29.99,
			Categories:  []string{"test", "premium"},
			CreatedAt:   time.Now(),
			InStock:     false,
		},
	}

	return ProductList{
		Products: products,
		Total:    len(products),
	}, nil
}

func GetProductByID(ctx context.Context) (res Product, err error) {
	id := strut.PathParam(ctx, "id")

	// In a real application, this would fetch from a database
	if id == "prod-1" {
		return Product{
			ID:          "prod-1",
			Name:        "Test Product 1",
			Description: "This is a test product",
			Price:       19.99,
			Categories:  []string{"test", "example"},
			CreatedAt:   time.Now(),
			InStock:     true,
		}, nil
	}

	// Return a custom error that will be handled as a 404
	return strut.NewError[Product](ctx, http.StatusNotFound, "product not found")
}

func CreateProduct(ctx context.Context, req CreateProductRequest) (res Product, err error) {
	// In a real application, this would save to a database
	return Product{
		ID:          "new-prod",
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Categories:  req.Categories,
		CreatedAt:   time.Now(),
		InStock:     true,
	}, nil
}

func SearchProducts(ctx context.Context) (res ProductList, err error) {
	query := strut.QueryParam(ctx, "query")
	category := strut.QueryParam(ctx, "category")
	minPrice := strut.QueryParam(ctx, "min_price")
	maxPrice := strut.QueryParam(ctx, "max_price")

	// In a real application, this would search in a database
	// For testing, we'll just return some mock data
	products := []Product{
		{
			ID:          "prod-1",
			Name:        "Test Product 1",
			Description: "This is a test product",
			Price:       19.99,
			Categories:  []string{"test", "example"},
			CreatedAt:   time.Now(),
			InStock:     true,
		},
	}

	// Log the search parameters for test verification
	fmt.Printf("Search params - query: %s, category: %s, minPrice: %s, maxPrice: %s\n",
		query, category, minPrice, maxPrice)

	return ProductList{
		Products: products,
		Total:    len(products),
	}, nil
}

// Test helper to create a test server
func setupTestServer(t *testing.T) (*strut.Strut, *chi.Mux, *httptest.Server) {
	r := chi.NewRouter()

	// Add custom error handler for 404 errors - MUST be added before routes
	//r.Use(func(next http.Handler) http.Handler {
	//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//		defer func() {
	//			if err := recover(); err != nil {
	//				if httpErr, ok := err.(strut.Error); ok {
	//					w.Header().Set("Content-Type", "application/json")
	//					w.WriteHeader(httpErr.StatusCode)
	//					json.NewEncoder(w).Encode(ErrorResponse{
	//						Error:   httpErr.Error(),
	//						Code:    httpErr.StatusCode,
	//						Details: httpErr.Details,
	//					})
	//					return
	//				}
	//				// Re-panic for other errors
	//				panic(err)
	//			}
	//		}()
	//		next.ServeHTTP(w, r)
	//	})
	//})

	s := strut.New(slog.Default(), r).
		Title("Product API").
		Description("API for managing products").
		Version("1.0.0").
		AddServer("http://localhost:8080", "Development server")

	// Register endpoints
	strut.Get(s, "/products", GetProducts,
		with.OperationId("get-products"),
		with.Description("Get all products"),
		with.Summary("List all products"),
		with.Tags("products"),
		with.QueryParam[int]("page", "Page number"),
		with.QueryParam[int]("limit", "Items per page"),
		with.ResponseDescription(200, "List of products"),
	)

	strut.Get(s, "/products/{id}", GetProductByID,
		with.OperationId("get-product-by-id"),
		with.Description("Get a specific product by its ID"),
		with.Summary("Get product details"),
		with.Tags("products"),
		with.PathParam[string]("id", "Product ID"),
		with.ResponseDescription(200, "Product details"),
		with.Response(404, swag.ResponseOf[strut.Error]("Product not found")),
	)

	strut.Post(s, "/products", CreateProduct,
		with.OperationId("create-product"),
		with.Description("Create a new product"),
		with.Summary("Create product"),
		with.Tags("products"),
		with.RequestDescription("Product to create"),
		with.ResponseDescription(200, "Created product"),
		with.Response(400, swag.ResponseOf[strut.Error]("Invalid product data")),
	)

	strut.Get(s, "/products/search", SearchProducts,
		with.OperationId("search-products"),
		with.Description("Search for products by query parameters"),
		with.Summary("Search products"),
		with.Tags("products"),
		with.QueryParam[string]("query", "Search query"),
		with.QueryParam[string]("category", "Filter by category"),
		with.QueryParam[float64]("min_price", "Minimum price"),
		with.QueryParam[float64]("max_price", "Maximum price"),
		with.ResponseDescription(200, "Search results"),
	)

	r.Get("/.well-known/openapi.yaml", s.SchemaHandlerYAML)
	r.Get("/.well-known/openapi.json", s.SchemaHandlerJSON)

	ts := httptest.NewServer(r)

	return s, r, ts
}

// Tests
func TestGetProducts(t *testing.T) {
	_, _, ts := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/products")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result ProductList
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Products, 2)
	assert.Equal(t, "Test Product 1", result.Products[0].Name)
	assert.Equal(t, "Test Product 2", result.Products[1].Name)
}

func TestGetProductByID(t *testing.T) {
	_, _, ts := setupTestServer(t)
	defer ts.Close()

	// Test successful retrieval
	resp, err := http.Get(ts.URL + "/products/prod-1")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var product Product
	err = json.NewDecoder(resp.Body).Decode(&product)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "prod-1", product.ID)
	assert.Equal(t, "Test Product 1", product.Name)

	// Test product not found
	resp, err = http.Get(ts.URL + "/products/non-existent")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCreateProduct(t *testing.T) {
	_, _, ts := setupTestServer(t)
	defer ts.Close()

	createReq := CreateProductRequest{
		Name:        "New Product",
		Description: "A brand new product",
		Price:       39.99,
		Categories:  []string{"new", "test"},
	}

	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	resp, err := http.Post(ts.URL+"/products", "application/json", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var product Product
	err = json.NewDecoder(resp.Body).Decode(&product)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "new-prod", product.ID)
	assert.Equal(t, createReq.Name, product.Name)
	assert.Equal(t, createReq.Description, product.Description)
	assert.Equal(t, createReq.Price, product.Price)
	assert.Equal(t, createReq.Categories, product.Categories)
	assert.True(t, product.InStock)
}

func TestSearchProducts(t *testing.T) {
	_, _, ts := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/products/search?query=test&category=example&min_price=10&max_price=50")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result ProductList
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 1, result.Total)
	assert.Len(t, result.Products, 1)
	assert.Equal(t, "Test Product 1", result.Products[0].Name)
}

func TestOpenAPISpecJSON(t *testing.T) {
	_, _, ts := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/.well-known/openapi.json")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var spec map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&spec)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Validate basic structure of the OpenAPI spec
	assert.Equal(t, "3.0.3", spec["openapi"])

	info, ok := spec["info"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Product API", info["title"])
	assert.Equal(t, "API for managing products", info["description"])
	assert.Equal(t, "1.0.0", info["version"])

	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok)

	// Check that all our endpoints are in the spec
	assert.Contains(t, paths, "/products")
	assert.Contains(t, paths, "/products/{id}")
	assert.Contains(t, paths, "/products/search")

	// Validate components section
	components, ok := spec["components"].(map[string]interface{})
	require.True(t, ok)

	schemas, ok := components["schemas"].(map[string]interface{})
	require.True(t, ok)

	// Check that our models are in the schemas
	assert.Contains(t, schemas, "strut_test_Product")
	assert.Contains(t, schemas, "strut_test_ProductList")
	assert.Contains(t, schemas, "strut_test_CreateProductRequest")
}

func TestOpenAPISpecYAML(t *testing.T) {
	_, _, ts := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/.well-known/openapi.yaml")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/yaml", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Basic validation of YAML content
	yamlContent := string(body)
	assert.True(t, strings.Contains(yamlContent, "openapi: 3.0.3"))
	assert.True(t, strings.Contains(yamlContent, "title: Product API"))
	assert.True(t, strings.Contains(yamlContent, "version: 1.0.0"))

	// Check that paths are included
	assert.True(t, strings.Contains(yamlContent, "/products:"))
	assert.True(t, strings.Contains(yamlContent, "/products/{id}:"))
	assert.True(t, strings.Contains(yamlContent, "/products/search:"))
}

func TestPathParameters(t *testing.T) {
	s, _, ts := setupTestServer(t)
	defer ts.Close()

	// Get the OpenAPI spec
	spec := s.Definition

	// Check that the path parameter is correctly defined
	pathItem, ok := spec.Paths["/products/{id}"]
	require.True(t, ok)
	require.NotNil(t, pathItem.Get)

	// Verify the path parameter
	var foundParam bool
	for _, param := range pathItem.Get.Parameters {
		if param.Name == "id" && param.In == "path" {
			foundParam = true
			assert.True(t, param.Required)
			assert.Equal(t, "Product ID", param.Description)
			assert.Equal(t, schema.String, param.Schema.Type)
		}
	}
	assert.True(t, foundParam, "Path parameter 'id' not found")
}

func TestQueryParameters(t *testing.T) {
	s, _, ts := setupTestServer(t)
	defer ts.Close()

	// Get the OpenAPI spec
	spec := s.Definition

	// Check that query parameters are correctly defined
	pathItem, ok := spec.Paths["/products/search"]
	require.True(t, ok)
	require.NotNil(t, pathItem.Get)

	// Expected parameters
	expectedParams := map[string]struct{}{
		"query":     {},
		"category":  {},
		"min_price": {},
		"max_price": {},
	}

	// Verify all query parameters
	for _, param := range pathItem.Get.Parameters {
		if param.In == "query" {
			_, exists := expectedParams[param.Name]
			assert.True(t, exists, "Unexpected query parameter: %s", param.Name)
			delete(expectedParams, param.Name)

			// Check specific parameter types
			if param.Name == "min_price" || param.Name == "max_price" {
				assert.Equal(t, schema.Number, param.Schema.Type)
			} else {
				assert.Equal(t, schema.String, param.Schema.Type)
			}
		}
	}

	assert.Empty(t, expectedParams, "Missing query parameters: %v", expectedParams)
}

func TestRequestBodySchema(t *testing.T) {
	s, _, ts := setupTestServer(t)
	defer ts.Close()

	// Get the OpenAPI spec
	spec := s.Definition

	// Check that the request body is correctly defined for POST /products
	pathItem, ok := spec.Paths["/products"]
	require.True(t, ok)
	require.NotNil(t, pathItem.Post)
	require.NotNil(t, pathItem.Post.RequestBody)

	// Verify request body content
	content, ok := pathItem.Post.RequestBody.Content["application/json"]
	require.True(t, ok)
	require.NotNil(t, content.Schema)

	// The schema should be a reference to the CreateProductRequest schema
	assert.Contains(t, content.Schema.Ref, "CreateProductRequest")
}

func TestResponseSchema(t *testing.T) {
	s, _, ts := setupTestServer(t)
	defer ts.Close()

	// Get the OpenAPI spec
	spec := s.Definition

	// Check responses for GET /products
	pathItem, ok := spec.Paths["/products"]
	require.True(t, ok)
	require.NotNil(t, pathItem.Get)

	// Verify 200 response
	response, ok := pathItem.Get.Responses["200"]
	require.True(t, ok)
	require.NotNil(t, response)

	// The response should have application/json content
	content, ok := response.Content["application/json"]
	require.True(t, ok)
	require.NotNil(t, content.Schema, "Content schema should not be nil")

	// The schema should be a reference to the ProductList schema
	assert.Contains(t, content.Schema.Ref, "ProductList")

	// Check error response for GET /products/{id}
	pathItem, ok = spec.Paths["/products/{id}"]
	require.True(t, ok)
	require.NotNil(t, pathItem.Get)

	// Verify 404 response
	response, ok = pathItem.Get.Responses["404"]
	require.True(t, ok)
	require.NotNil(t, response)
	assert.Equal(t, "Product not found", response.Description)

	// The response should have application/json content
	content, ok = response.Content["application/json"]
	require.True(t, ok)
	require.NotNil(t, content.Schema, "Error response schema should not be nil")

	// The schema should be a reference to the ErrorResponse schema
	assert.Contains(t, content.Schema.Properties["message"].Description, "Error message")
}

func TestSchemaGeneration(t *testing.T) {
	s, _, ts := setupTestServer(t)
	defer ts.Close()

	// Get the OpenAPI spec
	spec := s.Definition

	// Check that all schemas are correctly generated
	require.NotNil(t, spec.Components)
	require.NotNil(t, spec.Components.Schemas)

	// Check Product schema
	productSchema, ok := spec.Components.Schemas["strut_test_Product"]
	require.True(t, ok)
	assert.Equal(t, schema.Object, productSchema.Type)

	// Verify Product properties
	props := productSchema.Properties
	assert.Equal(t, schema.String, props["id"].Type)
	assert.Equal(t, "Unique product identifier", props["id"].Description)

	assert.Equal(t, schema.String, props["name"].Type)
	assert.Equal(t, 3, *props["name"].MinLength)
	assert.Equal(t, 100, *props["name"].MaxLength)

	assert.Equal(t, schema.Number, props["price"].Type)
	assert.Equal(t, 0.01, *props["price"].Minimum)

	assert.Equal(t, schema.Array, props["categories"].Type)
	assert.Equal(t, 1, *props["categories"].MinItems)
	assert.Equal(t, schema.String, props["categories"].Items.Type)

	assert.Equal(t, schema.Boolean, props["in_stock"].Type)

	// Check required fields
	assert.Contains(t, productSchema.Required, "id")
	assert.Contains(t, productSchema.Required, "name")
	assert.Contains(t, productSchema.Required, "price")
	assert.Contains(t, productSchema.Required, "categories")
	assert.Contains(t, productSchema.Required, "created_at")
	assert.Contains(t, productSchema.Required, "in_stock")
	assert.NotContains(t, productSchema.Required, "description") // This is optional
}

func TestOperationMetadata(t *testing.T) {
	s, _, ts := setupTestServer(t)
	defer ts.Close()

	// Get the OpenAPI spec
	spec := s.Definition

	// Check operation metadata for GET /products
	pathItem, ok := spec.Paths["/products"]
	require.True(t, ok)
	require.NotNil(t, pathItem.Get)

	op := pathItem.Get
	assert.Equal(t, "get-products", op.OperationID)
	assert.Equal(t, "Get all products", op.Description)
	assert.Equal(t, "List all products", op.Summary)
	assert.Contains(t, op.Tags, "products")

	// Check operation metadata for POST /products
	require.NotNil(t, pathItem.Post)
	op = pathItem.Post
	assert.Equal(t, "create-product", op.OperationID)
	assert.Equal(t, "Create a new product", op.Description)
	assert.Equal(t, "Create product", op.Summary)
	assert.Contains(t, op.Tags, "products")

	// Check request body description
	require.NotNil(t, op.RequestBody, "Request body should not be nil")
	assert.Equal(t, "Product to create", op.RequestBody.Description)

	// Check response description
	require.NotNil(t, op.Responses["200"], "Response 200 should not be nil")
	assert.Equal(t, "Created product", op.Responses["200"].Description)
}

func TestServerInformation(t *testing.T) {
	s, _, ts := setupTestServer(t)
	defer ts.Close()

	// Get the OpenAPI spec
	spec := s.Definition

	// Check API information
	assert.Equal(t, "Product API", spec.Info.Title)
	assert.Equal(t, "API for managing products", spec.Info.Description)
	assert.Equal(t, "1.0.0", spec.Info.Version)

	// Check server information
	require.Len(t, spec.Servers, 1)
	assert.Equal(t, "http://localhost:8080", spec.Servers[0].URL)
	assert.Equal(t, "Development server", spec.Servers[0].Description)
}
