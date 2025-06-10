# strut

A Go library for creating LLM-friendly APIs using chi and OpenAPI

## Overview

Strut is a specialized OpenAPI library built on top of Go's chi router that focuses on creating APIs that are optimized for interaction with Language Learning Models (LLMs) and agentic systems. It provides a streamlined way to expose your services in a format that's easy for AI agents to understand and work with.

## Features

- Built on top of chi router for efficient HTTP handling
- OpenAPI 3.0.3 compliant with automatic schema generation
- Type-safe API definitions using Go generics
- Simplified request and response handling
- Automatic OpenAPI documentation generation (JSON/YAML)
- Optimized for LLM interaction patterns
- Comprehensive error handling
- Path and query parameter support

## Installation

```bash
go get github.com/modfin/strut
```

## Basic Usage

### Creating a Simple API

```go
package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/modfin/strut"
	"github.com/modfin/strut/swag"
	"github.com/modfin/strut/with"
	"log/slog"
	"net/http"
)

// Define your request and response types
type Person struct {
	Name string `json:"name" json-description:"User's name"`
}

// Handler function for GET request
func GetPerson(ctx context.Context) strut.Response[Person] {
	name := strut.PathParam(ctx, "name")
	if name == "" {
		return strut.RespondError[Person](http.StatusBadRequest, "name is required")
	}
	return strut.RespondOk(Person{Name: name})
}

func main() {
	// Create a new chi router
	r := chi.NewRouter()

	// Create strut instance with logger and router
	s := strut.New(slog.Default(), r).
		Title("Person API").
		Description("API for retrieving person information").
		Version("1.0.0").
		AddServer("http://localhost:8080", "Development server")

	// Register GET endpoint
	strut.Get(s, "/person/{name}", GetPerson,
		with.OperationId("get-person"),
		with.Description("Get a person by name"),
		with.PathParam[string]("name", "Person's name"),
		with.ResponseDescription(200, "Person details"),
		with.Response(http.StatusBadRequest, swag.ResponseOf[strut.Error]("Error when name is missing")),
	)

	// Expose OpenAPI documentation
	r.Get("/.well-known/openapi.yaml", s.SchemaHandlerYAML)
	r.Get("/.well-known/openapi.json", s.SchemaHandlerJSON)

	// Start the server
	http.ListenAndServe(":8080", r)
}
```

### POST Request Example

```go
// Define request and response types
type CreatePersonRequest struct {
	Name    string `json:"name" json-description:"Person's name"`
	Age     int    `json:"age" json-description:"Person's age"`
	Country string `json:"country" json-description:"Person's country"`
}

type CreatePersonResponse struct {
	ID      string `json:"id" json-description:"Person's unique ID"`
	Name    string `json:"name" json-description:"Person's name"`
	Created string `json:"created" json-description:"Creation timestamp"`
}

// Handler function for POST request
func CreatePerson(ctx context.Context, req CreatePersonRequest) strut.Response[CreatePersonResponse] {
	// Validate request
	if req.Name == "" {
		return strut.RespondError[CreatePersonResponse](http.StatusBadRequest, "name is required")
	}
	
	// Process request (in a real app, you would save to database, etc.)
	response := CreatePersonResponse{
		ID:      "person-123", // In a real app, generate a unique ID
		Name:    req.Name,
		Created: time.Now().Format(time.RFC3339),
	}
	
	return strut.RespondOk(response)
}

// Register the endpoint
strut.Post(s, "/person", CreatePerson,
	with.OperationId("create-person"),
	with.Description("Create a new person"),
	with.RequestDescription("Person details to create"),
	with.ResponseDescription(200, "Created person details"),
	with.Response(http.StatusBadRequest, swag.ResponseOf[strut.Error]("Invalid request")),
)
```

## Advanced Usage

### Custom Response Handling

```go
func GetCustomResponse(ctx context.Context) strut.Response[any] {
	// Access HTTP writer to set custom headers
	w := strut.HTTPResponseWriter(ctx)
	w.Header().Set("X-Custom-Header", "custom-value")
	
	// Return custom response using RespondFunc
	return strut.RespondFunc[any](func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(map[string]string{
			"message": "Custom response",
			"time":    time.Now().Format(time.RFC3339),
		})
	})
}
```

### Error Handling

```go
func GetResource(ctx context.Context) strut.Response[Resource] {
	id := strut.PathParam(ctx, "id")
	
	// Return error response for invalid input
	if id == "0" {
		return strut.RespondError[Resource](http.StatusBadRequest, "Invalid ID")
	}
	
	// Return teapot status for special case
	if id == "teapot" {
		return strut.RespondError[Resource](http.StatusTeapot, "I'm a teapot")
	}
	
	// Normal successful response
	return strut.RespondOk(Resource{ID: id})
}

// Register with error responses documented
strut.Get(s, "/resource/{id}", GetResource,
	with.OperationId("get-resource"),
	with.Description("Get a resource by ID"),
	with.PathParam[string]("id", "Resource ID"),
	with.ResponseDescription(200, "Resource details"),
	with.Response(http.StatusBadRequest, swag.ResponseOf[strut.Error]("Invalid resource ID")),
	with.Response(http.StatusTeapot, swag.ResponseOf[strut.Error]("I'm a teapot")),
)
```

### Using Query Parameters

```go
func SearchPeople(ctx context.Context) strut.Response[PeopleList] {
	// Get query parameters
	name := strut.QueryParam(ctx, "name")
	country := strut.QueryParam(ctx, "country")
	
	// Use parameters for filtering (implementation details omitted)
	// ...
	
	return strut.RespondOk(results)
}

// Register endpoint with query parameters
strut.Get(s, "/people/search", SearchPeople,
	with.OperationId("search-people"),
	with.Description("Search for people"),
	with.QueryParam[string]("name", "Filter by name"),
	with.QueryParam[string]("country", "Filter by country"),
	with.ResponseDescription(200, "List of people matching criteria"),
)
```

## Schema and LLM-Friendly Descriptions

Strut uses the `schema` package to automatically generate OpenAPI documentation from your Go structs. This is particularly important for LLM agents, as it allows them to understand the purpose and constraints of each field in your API.

### Using JSON Description Tags

The `json-description` tag is a powerful way to communicate the purpose of each field to LLMs:

```go
type User struct {
    ID        string `json:"id" json-description:"Unique identifier for the user"`
    Name      string `json:"name" json-description:"User's full name"`
    Email     string `json:"email" json-description:"User's email address for notifications"`
    Age       int    `json:"age" json-description:"User's age in years"`
    IsActive  bool   `json:"is_active" json-description:"Whether the user account is currently active"`
    CreatedAt string `json:"created_at" json-description:"Timestamp when the user was created (RFC3339 format)"`
}
```

These descriptions are automatically included in the OpenAPI schema, making it easier for LLMs to:

1. Understand the purpose of each field
2. Generate appropriate sample values
3. Provide better assistance to users interacting with your API

### Additional Schema Tags

Beyond basic descriptions, you can use additional tags to provide more context:

```go
type Product struct {
    ID          string   `json:"id" json-description:"Unique product identifier"`
    Name        string   `json:"name" json-description:"Product name" json-min-length:"3" json-max-length:"100"`
    Price       float64  `json:"price" json-description:"Product price in USD" json-minimum:"0.01"`
    Categories  []string `json:"categories" json-description:"Product categories" json-min-items:"1"`
    StockLevel  int      `json:"stock_level" json-description:"Current inventory count" json-minimum:"0"`
    Status      string   `json:"status" json-description:"Product status" json-enum:"active,discontinued,out_of_stock"`
}
```

Available schema tags include:

| Tag | Type | Description |
|-----|------|-------------|
| `json-description` | All | Human-readable description of the field |
| `json-type` | All | Override the JSON type |
| `json-minimum` | Number/Integer | Minimum value (inclusive) |
| `json-maximum` | Number/Integer | Maximum value (inclusive) |
| `json-exclusive-minimum` | Number/Integer | Minimum value (exclusive) |
| `json-exclusive-maximum` | Number/Integer | Maximum value (exclusive) |
| `json-min-length` | String | Minimum string length |
| `json-max-length` | String | Maximum string length |
| `json-pattern` | String | Regular expression pattern |
| `json-format` | String | Format hint (e.g., "date-time", "email") |
| `json-min-items` | Array | Minimum array length |
| `json-max-items` | Array | Maximum array length |
| `json-enum` | String/Number/Integer/Boolean | Comma-separated list of allowed values |

### Why This Matters for LLM Agents

LLM agents can leverage these descriptions and constraints to:

1. **Generate Valid Requests**: Understand the expected format and constraints for each field
2. **Interpret Responses**: Correctly parse and understand the meaning of response fields
3. **Provide Better Assistance**: Help users construct valid requests with appropriate values
4. **Handle Errors**: Better understand error messages related to invalid inputs

### Example: Input and Output Descriptions

For optimal LLM interaction, describe both request and response objects clearly:

```go
// Request object with clear descriptions
type WeatherRequest struct {
    Location    string  `json:"location" json-description:"City name or coordinates (latitude,longitude)"`
    Units       string  `json:"units" json-description:"Temperature units: 'celsius' or 'fahrenheit'" json-enum:"celsius,fahrenheit"`
    IncludeUV   bool    `json:"include_uv" json-description:"Whether to include UV index in the response"`
}

// Response object with clear descriptions
type WeatherResponse struct {
    Location    string  `json:"location" json-description:"Location name that was resolved"`
    Temperature float64 `json:"temperature" json-description:"Current temperature in requested units"`
    Conditions  string  `json:"conditions" json-description:"Text description of weather conditions"`
    Humidity    int     `json:"humidity" json-description:"Current humidity percentage"`
    WindSpeed   float64 `json:"wind_speed" json-description:"Wind speed in km/h"`
    UVIndex     *int    `json:"uv_index,omitempty" json-description:"UV index (only included if requested)"`
    UpdatedAt   string  `json:"updated_at" json-description:"Timestamp of weather data (RFC3339 format)"`
}

// Handler with both input and output clearly described
func GetWeather(ctx context.Context, req WeatherRequest) strut.Response[WeatherResponse] {
    // Implementation...
}

// Register with descriptive operation details
strut.Post(s, "/weather", GetWeather,
    with.OperationId("get-weather"),
    with.Description("Get current weather conditions for a location"),
    with.RequestDescription("Weather request parameters"),
    with.ResponseDescription(200, "Current weather conditions"),
    with.Response(http.StatusBadRequest, swag.ResponseOf[strut.Error]("Invalid location or parameters")),
)
```

By providing clear descriptions for both input and output objects, you enable LLMs to better understand your API's purpose and usage patterns, making it easier for them to generate code that interacts with your API correctly.

## Why Strut for LLM Agents?

Strut is specifically designed to make it easier for LLMs and agentic systems to interact with your APIs:

1. **Clear Schema Documentation**: Automatically generates OpenAPI documentation that LLMs can parse to understand your API structure.

2. **Semantic Descriptions**: Uses `json-description` tags to provide semantic meaning to fields, making it easier for LLMs to understand the purpose of each field.

3. **Consistent Error Handling**: Standardized error responses make it easier for LLMs to handle and recover from errors.

4. **Self-Documenting Endpoints**: Operation IDs, descriptions, and parameter documentation help LLMs understand the purpose of each endpoint.

5. **Type Safety**: Go's type system ensures that your API contracts are consistent and well-defined.

## OpenAPI Documentation

Strut automatically generates OpenAPI documentation for your API, and it can be exposed 
through chi

```go 
    r := chi.NewRouter()
    // Create strut instance with logger and router
    s := strut.New(slog.Default(), r)...

    r.Get("/.well-known/openapi.yaml", s.SchemaHandlerYAML)
	r.Get("/.well-known/openapi.json", s.SchemaHandlerJSON)

```

LLM agents where clients can convert OpenAPI specs to Tools can use these 
endpoints to discover and understand your API structure.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.