# Strut

Strut is a lightweight, type-safe Go library for building somewwhat OpenAPI-compliant REST APIs with minimal boilerplate. It leverages Go's generics to provide a clean, fluent interface for defining API endpoints and automatically generates OpenAPI 3.0.3 documentation.

## Features

- **Type-safe API definitions**: Use Go's type system to define request and response structures
- **Automatic OpenAPI schema generation**: Generate OpenAPI 3.0.3 documentation from your Go code
- **Fluent API**: Chainable methods for configuring your API
- **Chi Router integration**: Built on top of the popular [chi router](https://github.com/go-chi/chi)
- **Minimal boilerplate**: Define endpoints with clean, readable code

## Installation

```bash
go get github.com/modfin/strut
```

## Quick Start

```go
package main

import (
    "context"
    "github.com/go-chi/chi/v5"
    "github.com/modfin/strut"
    "github.com/modfin/strut/with"
    "log/slog"
    "net/http"
)

// Define your request and response types
type UserRequest struct {
    Name string `json:"name" json-description:"User's name"`
}

type UserResponse struct {
    Name    string `json:"name" json-description:"User's name"`
    Greeting string `json:"greeting" json-description:"Personalized greeting"`
}

// Handler function
func GreetUser(ctx context.Context, req UserRequest) (UserResponse, error) {
    return UserResponse{
        Name:    req.Name,
        Greeting: "Hello, " + req.Name + "!",
    }, nil
}

func main() {
    // Create a chi router
    r := chi.NewRouter()
    
    // Create a new Strut instance
    s := strut.New(slog.Default(), r).
        Title("Greeting API").
        Description("A simple API for greeting users").
        Version("1.0.0").
        AddServer("http://localhost:8080", "Development server")
    
    // Define a POST endpoint
    strut.Post(s, "/greet", GreetUser,
        with.OperationId("greet-user"),
        with.Description("Greet a user by name"),
        with.Tags("greetings"),
        with.RequestDescription("User information"),
        with.ResponseDescription(200, "Personalized greeting"),
    )
    
    // Add OpenAPI schema endpoints
    r.Get("/.well-known/openapi.yaml", s.SchemaHandlerYAML)
    r.Get("/.well-known/openapi.json", s.SchemaHandlerJSON)
    
    // Start the server
    http.ListenAndServe(":8080", r)
}
```

## How It Works

Strut uses Go's generics and reflection to automatically:

1. Generate OpenAPI schemas from your Go structs
2. Map HTTP requests to your handler functions
3. Convert handler responses to HTTP responses
4. It is opinionated and only handles json

## Core Concepts

### Defining Endpoints

Strut provides functions for defining HTTP endpoints:

- `strut.Get` - Define a GET endpoint
- `strut.Post` - Define a POST endpoint
- `strut.Put` - Define a PUT endpoint
- `strut.Delete` - Define a DELETE endpoint

Each function takes:
- A Strut instance
- An endpoint path
- A handler function
- Optional configuration functions

### Handler Functions

Handler functions follow a simple pattern:

```go
// For endpoints with a request body (POST, PUT)
func HandlerInOut(ctx context.Context, req RequestType) (ResponseType, error)

// For endpoints without a request body (GET, DELETE)
func HandlerOut(ctx context.Context) (ResponseType, error)
```

If a handler returns an error then the response is expected to be set manually or using the `strut.NewError` function.

ex 

```go 
func(ctx context.Context) (Person, error) {
    w := HTTPResponseWriter(ctx)
    w.WriteHeader(http.StatusTeapot)
    w.Write([]byte("im a teapot"))
    
    return Person{}, errors.New("I respond my self")
}

// or 

func(ctx context.Context) (Person, error) {
    return strut.NewError[Person](ctx, http.StatusTeapot, "im a teapot")
}

```

### Configuration

The `with` package provides functions for configuring endpoints:

```go
strut.Post(s, "/users", CreateUser,
    with.OperationId("create-user"),
    with.Description("Create a new user"),
    with.Tags("users"),
    with.QueryParam[string]("source", "Source of the user creation"),
    with.RequestDescription("User to create"),
    with.ResponseDescription(200, "Created user"),
    with.Response(400, strut.ResponseOf[strut.Error]("Bad request")),
)
```

### Path and Query Parameters

Access path and query parameters using the provided helper functions:

```go
// Get a path parameter
name := strut.PathParam(ctx, "name")

// Get a query parameter
source := strut.QueryParam(ctx, "source")
```

## Example

Here's a complete example of a simple API with multiple endpoints:

```go
package main

import (
    "context"
    "github.com/go-chi/chi/v5"
    "github.com/modfin/strut"
    "github.com/modfin/strut/with"
    "log/slog"
    "net/http"
    "strconv"
)

type Person struct {
    Name string `json:"name" json-description:"Person's name"`
}

type PersonWithAge struct {
    Person
    Age int `json:"age" json-description:"Person's age"`
}

// Add age to a person
func AddAge(ctx context.Context, req Person) (PersonWithAge, error) {
    ageStr := strut.QueryParam(ctx, "age")
    age, _ := strconv.Atoi(ageStr)
    
    if age == 0 {
        age = 30 // Default age
    }
    
    return PersonWithAge{
        Person: req,
        Age:    age,
    }, nil
}

// Get a person by name
func GetPerson(ctx context.Context) (Person, error) {
    name := strut.PathParam(ctx, "name")
    return Person{
        Name: name,
    }, nil
}

    
   

func main() {
    r := chi.NewRouter()
    
    s := strut.New(slog.Default(), r).
        Title("Person API").
        Description("API for managing person information").
        Version("1.0.0").
        AddServer("http://localhost:8080", "Development server")
    
    // POST endpoint with request body and query parameter
    strut.Post(s, "/add-age", AddAge,
        with.OperationId("add-age"),
        with.Description("Add age to a person"),
        with.Tags("person"),
        with.QueryParam[int]("age", "Person's age"),
        with.RequestDescription("Person to add age to"),
        with.ResponseDescription(200, "Person with age"),
    )
    
    // GET endpoint with path parameter
    strut.Get(s, "/person/{name}", GetPerson,
        with.OperationId("get-person"),
        with.Description("Get a person by name"),
        with.PathParam[string]("name", "Person's name"),
        with.ResponseDescription(200, "Person information"),
    )

	// Failing example
	strut.Get(s, "/im-a-teapot", func(ctx context.Context) (Person, error) {
		return strut.NewError[Person](ctx, http.StatusTeapot, "im a teapot")
	},
    with.Description("a endpoint that always returns 418"),
	with.Response(418, strut.ResponseOf[strut.Error]("im a teapot")),
	)


	// Manual response
	strut.Get(s, "/im-a-teapot", func(ctx context.Context) (Person, error) {
		
        w := HTTPResponseWriter(ctx)
        w.WriteHeader(http.StatusTeapot)
        w.Write([]byte("im a teapot"))
		
		return Person{}, errors.New("I handle this by my self")
	},
		with.Description("a endpoint that always returns 418"),
		with.Response(418, strut.ResponseOf[strut.Error]("im a teapot")),
	)
    
    // Add OpenAPI schema endpoints
    r.Get("/.well-known/openapi.yaml", s.SchemaHandlerYAML)
    r.Get("/.well-known/openapi.json", s.SchemaHandlerJSON)
    
    http.ListenAndServe(":8080", r)
}
```

## License

Strut is released under the MIT License. See the LICENSE file for details.
