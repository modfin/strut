package strut

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Responder[T any] interface {
	Respond(w http.ResponseWriter, r *http.Request) error
}
type responseHandler[T any] struct {
	handler func(w http.ResponseWriter, r *http.Request) error
}

func (r *responseHandler[T]) Respond(wri http.ResponseWriter, req *http.Request) error {
	return r.handler(wri, req)
}

func RespondWith[T any](status int, response any) Responder[T] {
	return &responseHandler[T]{
		handler: func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			return json.NewEncoder(w).Encode(response)
		},
	}
}

type Error struct {
	StatusCode int    `json:"status_code" json-description:"Error code"`
	Error      string `json:"error" json-description:"Error message"`
}

func Respond[T any](response T) Responder[T] {
	return RespondWith[T](http.StatusOK, response)
}

func RespondError[T any](statusCode int, message string) Responder[T] {
	return RespondWith[T](statusCode, Error{StatusCode: statusCode, Error: message})
}

func RespondFunc[T any](handler func(w http.ResponseWriter, r *http.Request) error) Responder[T] {
	return &responseHandler[T]{
		handler: handler,
	}
}

// PathParam returns the value of the path parameter
// expects that chi is being used
func PathParam(ctx context.Context, param string) string {
	return chi.URLParamFromCtx(ctx, param)
}

func QueryParam(ctx context.Context, param string) string {
	r := HTTPRequest(ctx)
	if r == nil {
		return ""
	}
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
