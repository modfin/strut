package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/modfin/strut"
	"github.com/modfin/strut/with"
	"github.com/stretchr/testify/assert"
)

// Test data structures for middleware tests
type MiddlewareTestRequest struct {
	Message string `json:"message"`
}

type MiddlewareTestResponse struct {
	Message string            `json:"message"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Test handlers
func middlewareTestGetHandler(ctx context.Context) strut.Response[MiddlewareTestResponse] {
	return strut.RespondOk(MiddlewareTestResponse{Message: "GET response"})
}

func middlewareTestPostHandler(ctx context.Context, req MiddlewareTestRequest) strut.Response[MiddlewareTestResponse] {
	return strut.RespondOk(MiddlewareTestResponse{Message: fmt.Sprintf("POST response: %s", req.Message)})
}

// Middleware functions for testing
func headerMiddleware(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-"+name, "executed")
			next.ServeHTTP(w, r)
		})
	}
}

func authMiddleware(requiredToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token != requiredToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Middleware", "applied")
		next.ServeHTTP(w, r)
	})
}

func TestGlobalMiddleware(t *testing.T) {
	r := chi.NewRouter()

	// Apply CORS middleware directly to the router for OPTIONS testing
	r.Use(corsMiddleware)

	s := strut.New(slog.Default(), r)

	// Add global middleware
	s.Use(headerMiddleware("Global"))

	// Register endpoints
	strut.Get(s, "/test", middlewareTestGetHandler, with.OperationId("test-get"))
	strut.Post(s, "/test", middlewareTestPostHandler, with.OperationId("test-post"))

	// Add OPTIONS handler for CORS testing
	r.Options("/test", func(w http.ResponseWriter, r *http.Request) {
		// This will be handled by the CORS middleware
	})

	// Test GET endpoint with global middleware
	t.Run("GET with global middleware", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Global"))
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

		var response MiddlewareTestResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "GET response", response.Message)
	})

	// Test POST endpoint with global middleware
	t.Run("POST with global middleware", func(t *testing.T) {
		reqBody := `{"message": "test"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Global"))
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

		var response MiddlewareTestResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "POST response: test", response.Message)
	})

	// Test OPTIONS request (CORS preflight)
	t.Run("OPTIONS request with CORS middleware", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	})
}

func TestGroupMiddleware(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Add global middleware
	s.Use(headerMiddleware("Global"))

	// Register endpoint without group
	strut.Get(s, "/global", middlewareTestGetHandler, with.OperationId("global-get"))

	// Create a group with additional middleware
	s.Group(func(gs *strut.Strut) {
		gs.Use(headerMiddleware("Group"))
		gs.Use(contentTypeMiddleware)

		strut.Get(gs, "/group", middlewareTestGetHandler, with.OperationId("group-get"))
		strut.Post(gs, "/group", middlewareTestPostHandler, with.OperationId("group-post"))
	})

	// Test global endpoint (only global middleware)
	t.Run("Global endpoint with global middleware only", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/global", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Global"))
		assert.Empty(t, w.Header().Get("X-Middleware-Group"))
		assert.Empty(t, w.Header().Get("X-Content-Type-Middleware"))
	})

	// Test group endpoint (global + group middleware)
	t.Run("Group endpoint with global and group middleware", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/group", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Global"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Group"))
		assert.Equal(t, "applied", w.Header().Get("X-Content-Type-Middleware"))
	})

	// Test POST in group
	t.Run("POST in group with all middleware", func(t *testing.T) {
		reqBody := `{"message": "group test"}`
		req := httptest.NewRequest("POST", "/group", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Global"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Group"))
		assert.Equal(t, "applied", w.Header().Get("X-Content-Type-Middleware"))

		var response MiddlewareTestResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "POST response: group test", response.Message)
	})
}

func TestNestedGroups(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Add global middleware
	s.Use(headerMiddleware("Global"))

	// Create nested groups
	s.Group(func(level1 *strut.Strut) {
		level1.Use(headerMiddleware("Level1"))

		// Register endpoint at level 1
		strut.Get(level1, "/level1", middlewareTestGetHandler, with.OperationId("level1-get"))

		// Create nested group
		level1.Group(func(level2 *strut.Strut) {
			level2.Use(headerMiddleware("Level2"))

			strut.Get(level2, "/level2", middlewareTestGetHandler, with.OperationId("level2-get"))

			// Create another nested group
			level2.Group(func(level3 *strut.Strut) {
				level3.Use(headerMiddleware("Level3"))

				strut.Get(level3, "/level3", middlewareTestGetHandler, with.OperationId("level3-get"))
			})
		})
	})

	// Test level 1 endpoint
	t.Run("Level 1 endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/level1", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Global"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Level1"))
		assert.Empty(t, w.Header().Get("X-Middleware-Level2"))
		assert.Empty(t, w.Header().Get("X-Middleware-Level3"))
	})

	// Test level 2 endpoint
	t.Run("Level 2 endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/level2", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Global"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Level1"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Level2"))
		assert.Empty(t, w.Header().Get("X-Middleware-Level3"))
	})

	// Test level 3 endpoint
	t.Run("Level 3 endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/level3", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Global"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Level1"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Level2"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Level3"))
	})
}

func TestAuthenticationMiddleware(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Public endpoints (no auth required)
	strut.Get(s, "/public", middlewareTestGetHandler, with.OperationId("public-get"))

	// Protected endpoints
	s.Group(func(protected *strut.Strut) {
		protected.Use(authMiddleware("Bearer valid-token"))

		strut.Get(protected, "/protected", middlewareTestGetHandler, with.OperationId("protected-get"))
		strut.Post(protected, "/protected", middlewareTestPostHandler, with.OperationId("protected-post"))
	})

	// Test public endpoint (no auth required)
	t.Run("Public endpoint without auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/public", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test protected endpoint without auth
	t.Run("Protected endpoint without auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test protected endpoint with invalid auth
	t.Run("Protected endpoint with invalid auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test protected endpoint with valid auth
	t.Run("Protected endpoint with valid auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response MiddlewareTestResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "GET response", response.Message)
	})

	// Test protected POST endpoint with valid auth
	t.Run("Protected POST endpoint with valid auth", func(t *testing.T) {
		reqBody := `{"message": "protected test"}`
		req := httptest.NewRequest("POST", "/protected", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response MiddlewareTestResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "POST response: protected test", response.Message)
	})
}

func TestMiddlewareOrder(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Middleware that sets headers in order
	orderMiddleware := func(order string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				existing := w.Header().Get("X-Order")
				if existing == "" {
					w.Header().Set("X-Order", order)
				} else {
					w.Header().Set("X-Order", existing+","+order)
				}
				next.ServeHTTP(w, r)
			})
		}
	}

	// Add middleware in specific order
	s.Use(orderMiddleware("1"))
	s.Use(orderMiddleware("2"))

	s.Group(func(gs *strut.Strut) {
		gs.Use(orderMiddleware("3"))
		gs.Use(orderMiddleware("4"))

		strut.Get(gs, "/order", middlewareTestGetHandler, with.OperationId("order-get"))
	})

	t.Run("Middleware execution order", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/order", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "1,2,3,4", w.Header().Get("X-Order"))
	})
}

func TestMiddlewareWithAllHTTPMethods(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Add middleware to all methods
	s.Use(headerMiddleware("AllMethods"))

	// Register all HTTP methods
	strut.Get(s, "/test", middlewareTestGetHandler, with.OperationId("test-get"))
	strut.Post(s, "/test", middlewareTestPostHandler, with.OperationId("test-post"))
	strut.Put(s, "/test", middlewareTestPostHandler, with.OperationId("test-put"))
	strut.Delete(s, "/test", middlewareTestGetHandler, with.OperationId("test-delete"))

	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		t.Run(fmt.Sprintf("%s method with middleware", method), func(t *testing.T) {
			var req *http.Request
			if method == "GET" || method == "DELETE" {
				req = httptest.NewRequest(method, "/test", nil)
			} else {
				reqBody := `{"message": "test"}`
				req = httptest.NewRequest(method, "/test", strings.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "executed", w.Header().Get("X-Middleware-AllMethods"))
		})
	}
}

func TestMiddlewareIsolation(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)
	s.Use(headerMiddleware("Group0"))

	// Create two separate groups with different middleware
	s.Group(func(group1 *strut.Strut) {
		group1.Use(headerMiddleware("Group1"))
		strut.Get(group1, "/group1", middlewareTestGetHandler, with.OperationId("group1-get"))
	})

	s.Group(func(group2 *strut.Strut) {
		group2.Use(headerMiddleware("Group2"))
		strut.Get(group2, "/group2", middlewareTestGetHandler, with.OperationId("group2-get"))
	})

	// Test that group1 middleware doesn't affect group2
	t.Run("Group1 endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/group1", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Group0"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Group1"))
		assert.Empty(t, w.Header().Get("X-Middleware-Group2"))
	})

	// Test that group2 middleware doesn't affect group1
	t.Run("Group2 endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/group2", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Group2"))
		assert.Empty(t, w.Header().Get("X-Middleware-Group1"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Group0"))

	})
}

func TestMiddlewareIsolationWith(t *testing.T) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)
	s.Use(headerMiddleware("Group0"))

	// Create two separate groups with different middleware

	strut.Get(s.With(headerMiddleware("Group1")),
		"/group1", middlewareTestGetHandler, with.OperationId("group1-get"))

	strut.Get(s.With(headerMiddleware("Group2")),
		"/group2", middlewareTestGetHandler, with.OperationId("group2-get"))

	// Test that group1 middleware doesn't affect group2
	t.Run("Group1 endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/group1", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		fmt.Println(w.Header())
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Group0"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Group1"))
		assert.Empty(t, w.Header().Get("X-Middleware-Group2"))
	})

	// Test that group2 middleware doesn't affect group1
	t.Run("Group2 endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/group2", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		fmt.Println(w.Header())

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Group2"))
		assert.Empty(t, w.Header().Get("X-Middleware-Group1"))
		assert.Equal(t, "executed", w.Header().Get("X-Middleware-Group0"))

	})
}

// Benchmark tests for middleware performance
func BenchmarkMiddlewareOverhead(b *testing.B) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r)

	// Add multiple middleware layers
	s.Use(headerMiddleware("1"))
	s.Use(headerMiddleware("2"))
	s.Use(headerMiddleware("3"))
	s.Use(corsMiddleware)
	s.Use(contentTypeMiddleware)

	strut.Get(s, "/benchmark", middlewareTestGetHandler, with.OperationId("benchmark-get"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/benchmark", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
