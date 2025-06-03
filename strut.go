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

type Mux interface {
	Post(path string, handler http.HandlerFunc)
	Get(path string, handler http.HandlerFunc)
	Put(path string, handler http.HandlerFunc)
	Delete(path string, handler http.HandlerFunc)
}

type Strut struct {
	Definition *Definition
	mux        Mux
	log        *slog.Logger
}

func (s *Strut) AddServer(url string, description string) *Strut {
	s.Definition.Servers = append(s.Definition.Servers, Server{
		URL:         url,
		Description: description,
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

func New(log *slog.Logger, mux Mux) *Strut {

	return &Strut{
		log: log,
		mux: mux,
		Definition: &Definition{
			OpenAPI: "3.0.3",
			Info: Info{
				Title:       "strut",
				Description: "strut",
				Version:     "v0.0.1",
			},
			Paths: map[string]*Path{},
			Components: &Components{
				Schemas: map[string]*schema.JSON{},
			},
		},
	}
}

// PathParam returns the value of the path parameter
// expects that chi is being used
func PathParam(ctx context.Context, param string) string {
	return chi.URLParamFromCtx(ctx, param)
}

func QueryParam(ctx context.Context, param string) string {
	r := HTTPRequest(ctx)
	return r.URL.Query().Get(param)
}

func HTTPRequest(ctx context.Context) *http.Request {
	r := ctx.Value("http-request")
	if r == nil {
		return nil
	}
	return r.(*http.Request)
}
func HTTPResponseWriter(ctx context.Context) http.ResponseWriter {
	w := ctx.Value("http-response-writer")
	if w == nil {
		return nil
	}
	return ctx.Value("http-response-writer").(http.ResponseWriter)
}

func decorateContext(req *http.Request, w http.ResponseWriter) context.Context {
	ctx := req.Context()
	ctx = context.WithValue(ctx, "http-request", req)
	ctx = context.WithValue(ctx, "http-response-writer", w)
	return ctx
}

type OpConfig func(op *Operation)

func assignOperation(ops ...OpConfig) *Operation {
	op := &Operation{}
	for _, o := range ops {
		o(op)
	}
	return op
}

func assignRequest[REQ any](s *Strut, op *Operation) {
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
}
func assignResponse[RES any](s *Strut, op *Operation) {
	var res RES
	resSchema := schema.From(res)

	resType := reflect.TypeOf(res)
	resName := resType.Name()
	resPkg := filepath.Base(resType.PkgPath())
	resUri := fmt.Sprintf("%s_%s", resPkg, resName)
	resRef := "#/components/schemas/" + resUri
	s.Definition.Components.Schemas[resUri] = resSchema
	if op.Responses == nil {
		op.Responses = map[string]*Response{}
	}
	if op.Responses["200"] == nil {
		op.Responses["200"] = &Response{}
	}
	if op.Responses["200"].Content == nil {
		op.Responses["200"].Content = map[string]MediaType{}
	}
	op.Responses["200"].Content["application/json"] = MediaType{
		Schema: &schema.JSON{Ref: resRef},
	}
}

func getPath(d *Definition, path string) *Path {
	if d.Paths == nil {
		d.Paths = map[string]*Path{}
	}
	if d.Paths[path] == nil {
		d.Paths[path] = &Path{}
	}
	return d.Paths[path]
}

func createResponse[RES any](s *Strut, ctx context.Context, res RES) {
	w := HTTPResponseWriter(ctx)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		s.log.Error("error encoding response", "error", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

}

// Handler for a GET and DELETE request
type HandlerOut[RES any] func(ctx context.Context) (res RES, err error)

// Handler for a POST and PUT request
type HandlerInOut[REQ any, RES any] func(ctx context.Context, req REQ) (res RES, err error)

func Post[REQ any, RES any](s *Strut, path string,
	handler HandlerInOut[REQ, RES],
	ops ...OpConfig,
) {

	op := assignOperation(ops...)
	getPath(s.Definition, path).Post = op
	assignRequest[REQ](s, op)
	assignResponse[RES](s, op)

	s.mux.Post(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := decorateContext(r, w)

		var req REQ
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			s.log.Error("error decoding request", "error", err)
			http.Error(w, "could not decode request", http.StatusBadRequest)
			return
		}

		res, err := handler(ctx, req)
		if err != nil {
			return
		}
		createResponse(s, ctx, res)

	})

}

func Get[RES any](s *Strut, path string,
	handler HandlerOut[RES],
	ops ...OpConfig,
) {

	op := assignOperation(ops...)
	getPath(s.Definition, path).Get = op
	assignResponse[RES](s, op)

	s.mux.Get(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := decorateContext(r, w)

		res, err := handler(ctx)
		if err != nil {
			return
		}

		createResponse(s, ctx, res)
	})

}

func Put[REQ any, RES any](s *Strut, path string,
	handler HandlerInOut[REQ, RES],
	ops ...OpConfig,
) {

	op := assignOperation(ops...)
	getPath(s.Definition, path).Put = op
	assignRequest[REQ](s, op)
	assignResponse[RES](s, op)

	s.mux.Put(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := decorateContext(r, w)

		var req REQ
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {

			s.log.Error("error decoding request", "error", err)
			http.Error(w, "could not decode request", http.StatusBadRequest)
			return
		}

		res, err := handler(ctx, req)
		if err != nil {
			return
		}

		createResponse(s, ctx, res)
	})
}

func Delete[RES any](s *Strut, path string,
	handler HandlerOut[RES],
	ops ...OpConfig,
) {

	op := assignOperation(ops...)
	getPath(s.Definition, path).Delete = op
	assignResponse[RES](s, op)

	s.mux.Delete(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := decorateContext(r, w)

		res, err := handler(ctx)
		if err != nil {
			return
		}

		createResponse(s, ctx, res)
	})
}
