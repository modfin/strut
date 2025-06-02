package strut

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type Error struct {
	Status  int    `json:"status" json-description:"HTTP status code"`
	Message string `json:"message" json-description:"Error message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Status, e.Message)
}

func NewError[T any](ctx context.Context, status int, message string) (T, error) {

	w := HTTPResponseWriter(ctx)

	err := &Error{
		Status:  status,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	encErr := json.NewEncoder(w).Encode(err)

	var z T
	return z, errors.Join(err, encErr)
}
