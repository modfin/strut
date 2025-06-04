package strut

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/modfin/strut/schema"
	"github.com/modfin/strut/swag"
	"gopkg.in/yaml.v3"
	"log/slog"
	"net/http"
	"path/filepath"
	"reflect"
)

func New(log *slog.Logger, mux *chi.Mux) *Strut {

	return &Strut{
		log: log,
		mux: mux,
		Definition: &swag.Definition{
			OpenAPI: "3.0.3",
			Info: swag.Info{
				Title:       "strut",
				Description: "strut",
				Version:     "v0.0.1",
			},
			Paths: map[string]*swag.Path{},
			Components: &swag.Components{
				Schemas: map[string]*schema.JSON{},
			},
		},
	}
}

type Strut struct {
	Definition *swag.Definition
	mux        *chi.Mux
	log        *slog.Logger
}

func (s *Strut) AddServer(url string, description string) *Strut {
	s.Definition.Servers = append(s.Definition.Servers, swag.Server{
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

func decorateContext(req *http.Request, w http.ResponseWriter) context.Context {
	ctx := req.Context()
	ctx = context.WithValue(ctx, "http-request", req)
	ctx = context.WithValue(ctx, "http-response-writer", w)
	return ctx
}

type OpConfig func(op *swag.Operation)

func assignOperation(ops ...OpConfig) *swag.Operation {
	op := &swag.Operation{}
	for _, o := range ops {
		o(op)
	}
	return op
}

func assignRequest[REQ any](s *Strut, op *swag.Operation) {
	var req REQ
	reqSchema := schema.From(req)
	reqType := reflect.TypeOf(req)
	reqName := reqType.Name()
	reqPkg := filepath.Base(reqType.PkgPath())
	reqUri := fmt.Sprintf("%s_%s", reqPkg, reqName)

	reqRef := "#/components/schemas/" + reqUri
	s.Definition.Components.Schemas[reqUri] = reqSchema

	if op.RequestBody == nil { // Defaulting stuff...
		op.RequestBody = &swag.RequestBody{}
	}
	if op.RequestBody.Content == nil {
		op.RequestBody.Content = map[string]swag.MediaType{}
	}
	op.RequestBody.Content["application/json"] = swag.MediaType{
		Schema: &schema.JSON{Ref: reqRef},
	}
}
func assignResponse[RES any](s *Strut, op *swag.Operation) {
	var res RES
	resSchema := schema.From(res)

	resType := reflect.TypeOf(res)
	resName := resType.Name()
	resPkg := filepath.Base(resType.PkgPath())
	resUri := fmt.Sprintf("%s_%s", resPkg, resName)
	resRef := "#/components/schemas/" + resUri
	s.Definition.Components.Schemas[resUri] = resSchema
	if op.Responses == nil {
		op.Responses = map[string]*swag.OpResponse{}
	}
	if op.Responses["200"] == nil {
		op.Responses["200"] = &swag.OpResponse{}
	}
	if op.Responses["200"].Content == nil {
		op.Responses["200"].Content = map[string]swag.MediaType{}
	}
	op.Responses["200"].Content["application/json"] = swag.MediaType{
		Schema: &schema.JSON{Ref: resRef},
	}
}

func getPath(d *swag.Definition, path string) *swag.Path {
	if d.Paths == nil {
		d.Paths = map[string]*swag.Path{}
	}
	if d.Paths[path] == nil {
		d.Paths[path] = &swag.Path{}
	}
	return d.Paths[path]
}

func createResponse(s *Strut, ctx context.Context, responder Responder[any]) {
	w, r := HTTPResponseWriter(ctx), HTTPRequest(ctx)
	err := responder.Respond(w, r)
	if err != nil {
		s.log.Error("error responding", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandlerOut Handler for a GET and DELETE request
type HandlerOut[RES any] func(ctx context.Context) Responder[RES]

// HandlerInOut Handler for a POST and PUT request
type HandlerInOut[REQ any, RES any] func(ctx context.Context, req REQ) Responder[RES]

func Post[REQ any, RES any](s *Strut, path string, handler HandlerInOut[REQ, RES], ops ...OpConfig) {

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

		res := handler(ctx, req)
		createResponse(s, ctx, res)

	})

}

func Get[RES any](s *Strut, path string, handler HandlerOut[RES], ops ...OpConfig) {

	op := assignOperation(ops...)
	getPath(s.Definition, path).Get = op
	assignResponse[RES](s, op)

	s.mux.Get(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := decorateContext(r, w)

		res := handler(ctx)
		createResponse(s, ctx, res)
	})

}

func Put[REQ any, RES any](s *Strut, path string, handler HandlerInOut[REQ, RES], ops ...OpConfig) {

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

		res := handler(ctx, req)
		createResponse(s, ctx, res)
	})
}

func Delete[RES any](s *Strut, path string, handler HandlerOut[RES], ops ...OpConfig) {

	op := assignOperation(ops...)
	getPath(s.Definition, path).Delete = op
	assignResponse[RES](s, op)

	s.mux.Delete(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := decorateContext(r, w)

		res := handler(ctx)
		createResponse(s, ctx, res)
	})
}
