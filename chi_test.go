package strut_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/modfin/strut"
	"github.com/modfin/strut/with"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"
)

type Person struct {
	Name string `in:"body" json:"name" json-description:"User's name"`
}

type PersonWithAge struct {
	Person
	Age int `json:"age" description:"User's age"`
}

func SetAge(ctx context.Context, req Person) (res PersonWithAge, err error) {

	agestr := strut.QueryParam(ctx, "age")
	age, err := strconv.Atoi(agestr)

	if age == 0 {
		age = rand.Intn(100) + 1
	}

	return PersonWithAge{
		Person: req,
		Age:    age,
	}, nil
}

func GetPerson(ctx context.Context) (res Person, err error) {
	name := strut.PathParam(ctx, "name")
	return Person{
		Name: name,
	}, nil
}

func StartServer(t *testing.T) *http.Server {
	r := chi.NewRouter()

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	s := strut.New(r).
		Title("A Strut API, openapi for agents").
		Description("A Strut API, openapi for agents implementing demo endpoints").
		Version("1.0.0").
		AddServer("http://localhost:8080", "The main server")

	strut.Post(s, "/add-age", SetAge,
		with.OperationId("add-age"),
		with.Description("Adds a random age to the user"),
		with.Tags("users", "user-age"),

		// Query params
		with.QueryParam[int]("age", "User's age"),

		with.RequestDescription("User to add age to"),
		// Since generics is used. The description of the
		with.ResponseDescription("User with added age"),
	)

	strut.Get(s, "/person/{name}", GetPerson,
		with.OperationId("get-person"),
		with.Description("Get a person by name"),
		// Query params
		with.PathParam[string]("name", "User's name"),

		with.ResponseDescription("A person"),
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

func TestFrom_Simple(t *testing.T) {
	ss := StartServer(t)
	defer func() {
		cc, cancel := context.WithTimeout(context.Background(), time.Second)
		ss.Shutdown(cc)
		cancel()
	}()

	for i := 0; i < 100; i++ {
		resp, err := http.Post("http://localhost:8080/add-age", "application/json", bytes.NewBuffer([]byte(`{"name": "John Doe"}`)))
		if err != nil {
			t.Fatal(err)
		}

		var result PersonWithAge
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()

		assert.Equal(t, "John Doe", result.Name)
		assert.NotEqual(t, 0, result.Age)
	}

}

func TestFrom_Fixed(t *testing.T) {
	ss := StartServer(t)
	defer func() {
		cc, cancel := context.WithTimeout(context.Background(), time.Second)
		ss.Shutdown(cc)
		cancel()
	}()

	for i := 0; i < 10; i++ {
		resp, err := http.Post("http://localhost:8080/add-age?age=17", "application/json", bytes.NewBuffer([]byte(`{"name": "John Doe"}`)))
		if err != nil {
			t.Fatal(err)
		}

		var result PersonWithAge
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()

		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, 17, result.Age)
	}

}

func TestFrom_spec_yaml(t *testing.T) {
	ss := StartServer(t)
	defer func() {
		cc, cancel := context.WithTimeout(context.Background(), time.Second)
		ss.Shutdown(cc)
		cancel()
	}()

	y, err := http.Get("http://localhost:8080/.well-known/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	text, err := io.ReadAll(y.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(text))

}
