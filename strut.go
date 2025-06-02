package strut

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/modfin/strut/schema"
	"gopkg.in/yaml.v3"
	"log/slog"
	"net/http"
	"path/filepath"
	"reflect"
)

type Strut struct {
	Definition *Definition
	mux        *chi.Mux
}

func (s *Strut) AddServer(url string, description string) *Strut {
	s.Definition.Servers = append(s.Definition.Servers, Server{
		URL:         url,
		Description: "the main server",
	})
	return s
}

func (s *Strut) Title(title string) *Strut {
	s.Definition.Info.Title = title
	return s

}
func (s *Strut) Description(description string) *Strut {
	s.Definition.Info.Description = description
	return s

}
func (s *Strut) Version(version string) *Strut {
	s.Definition.Info.Version = version
	return s

}

func (s *Strut) SchemaHandlerYAML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	err := yaml.NewEncoder(w).Encode(s.Definition)
	if err != nil {
		slog.Error("error encoding schema", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Strut) SchemaHandlerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(s.Definition)
	if err != nil {
		slog.Error("error encoding schema", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func New(chi *chi.Mux) *Strut {
	return &Strut{
		mux: chi,
		Definition: &Definition{
			OpenAPI: "3.0.3",
			Info: Info{
				Title:       "strut",
				Description: "strut",
				Version:     "v0.0.1",
			},
			Paths: map[string]Path{},
			Components: &Components{
				Schemas: map[string]*schema.JSON{},
			},
		},
	}
}

func HTTPRequest(ctx context.Context) *http.Request {
	return ctx.Value("http-request").(*http.Request)
}

func PathParam(ctx context.Context, param string) string {
	return chi.URLParamFromCtx(ctx, param)
}

func QueryParam(ctx context.Context, param string) string {
	r := HTTPRequest(ctx)
	return r.URL.Query().Get(param)
}

type OpConfig func(op *Operation)

func Post[REQ any, RES any](s *Strut, path string,
	fn func(ctx context.Context, req REQ) (res RES, err error),
	ops ...OpConfig,
) {

	op := Operation{}
	for _, o := range ops {
		o(&op)
	}

	s.Definition.Paths[path] = Path{
		Post: &op,
	}

	var req REQ
	reqSchema := schema.From(req)
	reqType := reflect.TypeOf(req)
	reqName := reqType.Name()
	reqPkg := filepath.Base(reqType.PkgPath())
	reqUri := fmt.Sprintf("%s_%s", reqPkg, reqName)

	reqRef := "#/components/schemas/" + reqUri
	s.Definition.Components.Schemas[reqUri] = reqSchema

	if op.RequestBody == nil { // Defaulting stuff...
		op.RequestBody = &RequestBody{}
	}
	if op.RequestBody.Content == nil {
		op.RequestBody.Content = map[string]MediaType{}
	}
	op.RequestBody.Content["application/json"] = MediaType{
		Schema: &schema.JSON{Ref: reqRef},
	}

	var res RES
	resSchema := schema.From(res)

	resType := reflect.TypeOf(res)
	resName := resType.Name()
	resPkg := filepath.Base(resType.PkgPath())
	resUri := fmt.Sprintf("%s_%s", resPkg, resName)
	resRef := "#/components/schemas/" + resUri
	s.Definition.Components.Schemas[resUri] = resSchema
	if op.Responses["200"] == nil {
		op.Responses["200"] = &Response{}
	}
	if op.Responses["200"].Content == nil {
		op.Responses["200"].Content = map[string]MediaType{}
	}
	op.Responses["200"].Content["application/json"] = MediaType{
		Schema: &schema.JSON{Ref: resRef},
	}

	s.mux.Post(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "http-request", r)

		var req REQ
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			// TODO pass error logging to chi middleware?
			slog.Default().Error("error decoding request", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := fn(ctx, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			// TODO pass error logging to chi middleware?
			slog.Default().Error("error encoding response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

}

func Get[RES any](s *Strut, path string,
	fn func(ctx context.Context) (res RES, err error),
	ops ...OpConfig,
) {

	op := Operation{}
	for _, o := range ops {
		o(&op)
	}

	s.Definition.Paths[path] = Path{
		Post: &op,
	}

	var res RES
	resSchema := schema.From(res)

	resType := reflect.TypeOf(res)
	resName := resType.Name()
	resPkg := filepath.Base(resType.PkgPath())
	resUri := fmt.Sprintf("%s_%s", resPkg, resName)
	resRef := "#/components/schemas/" + resUri
	s.Definition.Components.Schemas[resUri] = resSchema
	if op.Responses["200"] == nil {
		op.Responses["200"] = &Response{}
	}
	if op.Responses["200"].Content == nil {
		op.Responses["200"].Content = map[string]MediaType{}
	}
	op.Responses["200"].Content["application/json"] = MediaType{
		Schema: &schema.JSON{Ref: resRef},
	}

	s.mux.Get(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "http-request", r)

		res, err := fn(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			// TODO pass error logging to chi middleware?
			slog.Default().Error("error encoding response", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

}
