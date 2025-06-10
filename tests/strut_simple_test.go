package tests

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/modfin/strut"
	"github.com/modfin/strut/swag"
	"github.com/modfin/strut/with"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

type Person struct {
	Name string `in:"body" json:"name" json-description:"User's name"`
}

func GetPerson(ctx context.Context) strut.Response[Person] {
	nn := strut.PathParam(ctx, "name")
	if nn == "" {
		return strut.RespondError[any](http.StatusBadRequest, "name is required")
	}
	person := Person{Name: nn}
	return strut.RespondOk(person)
}

func GetTeapot(ctx context.Context) strut.Response[Person] {
	nn := strut.PathParam(ctx, "name")

	// RespondError allows you to return an error response, {"error": "im a teapot", "status_code": 418}
	return strut.RespondError[any](http.StatusTeapot, "im a teapot, "+nn)
}

type CustomData struct {
	Im     string `json:"im" json-description:"I'm"`
	Custom string `json:"custom" json-description:"Custom"`
	Data   string `json:"data" json-description:"Data"`
}

func GetCustom(ctx context.Context) strut.Response[CustomData] {

	w := strut.HTTPResponseWriter(ctx)
	w.Header().Set("X-Custom", "custom")

	// RespondFunc allows you to return a custom http response,
	// which can be used for binary data or custom structs for different responses codes or such
	return strut.RespondFunc[any](func(w http.ResponseWriter, r *http.Request) error {
		return json.NewEncoder(w).Encode(CustomData{
			Im:     "I'm",
			Custom: "custom",
			Data:   "data",
		})
	})
}

func StartServer(t *testing.T) *http.Server {
	r := chi.NewRouter()

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	s := strut.New(slog.Default(), r).
		Title("A Strut API, openapi for agents").
		Description("A Strut API, openapi for agents implementing demo endpoints").
		Version("1.0.0").
		AddServer("http://localhost:8080", "The main server")

	strut.Get(s, "/person/{name}", GetPerson,
		with.OperationId("get-person"),
		with.Description("Get a person by name"),
		// Path parameters
		with.PathParam[string]("name", "User's name"),
		with.ResponseDescription(200, "A person"),
		with.Response(http.StatusBadRequest, swag.ResponseOf[strut.Error]("error when name is missing")),
	)

	strut.Get(s, "/person/{name}/teapot", GetTeapot,
		with.OperationId("get-teapot"),
		with.Description("Get a teapot by name"),
		// Path parameters
		with.PathParam[string]("name", "User's name"),
		with.ResponseDescription(200, "A person"),
		with.Response(http.StatusTeapot, swag.ResponseOf[strut.Error]("im a teapot")),
	)

	strut.Get(s, "/custom/stuff", GetCustom,
		with.OperationId("get-person"),
		with.Description("Get a person by name"),
		with.ResponseDescription(200, "A set of custom data"),
	)

	r.Get("/.well-known/openapi.yaml", s.SchemaHandlerYAML)
	r.Get("/.well-known/openapi.json", s.SchemaHandlerJSON)

	ss := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		ss.ListenAndServe()
	}()

	// wait until endpoint is alive
	endpoint := "http://localhost:8080/ping"
	start := time.Now()
	for time.Since(start) < time.Second {
		resp, err := http.Get(endpoint)
		if err == nil && resp.StatusCode == http.StatusOK {
			return ss
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("endpoint %s is not alive after 1 second", endpoint)

	return ss
}

func TestFrom_SimpleGet(t *testing.T) {
	ss := StartServer(t)
	defer func() {
		cc, cancel := context.WithTimeout(context.Background(), time.Second)
		ss.Shutdown(cc)
		cancel()
	}()

	resp, err := http.Get("http://localhost:8080/person/John%20Doe")
	if err != nil {
		t.Fatal(err)
	}

	var result Person
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	assert.Equal(t, "John Doe", result.Name)

}

func TestFrom_SimpleGetTeapot(t *testing.T) {
	ss := StartServer(t)
	defer func() {
		cc, cancel := context.WithTimeout(context.Background(), time.Second)
		ss.Shutdown(cc)
		cancel()
	}()

	resp, err := http.Get("http://localhost:8080/person/John%20Doe/teapot")
	if err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	assert.Contains(t, string(data), "im a teapot, John Doe")

}

func TestFrom_SimpleGetCustom(t *testing.T) {
	ss := StartServer(t)
	defer func() {
		cc, cancel := context.WithTimeout(context.Background(), time.Second)
		ss.Shutdown(cc)
		cancel()
	}()

	resp, err := http.Get("http://localhost:8080/custom/stuff")
	if err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(data), "im")
	assert.Contains(t, string(data), "custom")
	assert.Contains(t, string(data), "data")

	assert.Equal(t, resp.Header.Get("x-custom"), "custom")

}
