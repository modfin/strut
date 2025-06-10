package with

import (
	"fmt"
	"github.com/modfin/strut"
	"github.com/modfin/strut/schema"
	"github.com/modfin/strut/swag"
)

func Description(description string) strut.OpConfig {
	return func(op *swag.Operation) {
		op.Description = description
	}
}
func OperationId(operationId string) strut.OpConfig {
	return func(op *swag.Operation) {
		op.OperationID = operationId
	}
}

func Summary(summary string) strut.OpConfig {
	return func(op *swag.Operation) {
		op.Summary = summary
	}
}

func Param(param ...swag.Param) strut.OpConfig {
	return func(op *swag.Operation) {
		op.Parameters = append(op.Parameters, param...)
	}
}

func QueryParam[T any](name string, description string) strut.OpConfig {
	var ref T
	return func(op *swag.Operation) {
		op.Parameters = append(op.Parameters, swag.Param{
			Name:        name,
			In:          "query",
			Description: description,
			Schema:      schema.From(ref),
		})
	}
}
func PathParam[T any](name string, description string) strut.OpConfig {
	var ref T
	return func(op *swag.Operation) {
		op.Parameters = append(op.Parameters, swag.Param{
			Name:        name,
			In:          "path",
			Description: description,
			Schema:      schema.From(ref),
			Required:    true,
		})
	}
}

func CookieParam[T any](name string, description string) strut.OpConfig {
	var ref T
	return func(op *swag.Operation) {
		op.Parameters = append(op.Parameters, swag.Param{
			Name:        name,
			In:          "cookie",
			Description: description,
			Schema:      schema.From(ref),
		})
	}
}
func HeaderParam[T any](name string, description string) strut.OpConfig {
	var ref T
	return func(op *swag.Operation) {
		op.Parameters = append(op.Parameters, swag.Param{
			Name:        name,
			In:          "header",
			Description: description,
			Schema:      schema.From(ref),
		})
	}
}

func Deprecated() strut.OpConfig {
	return func(op *swag.Operation) {
		op.Deprecated = true
	}
}

func Tags(tags ...string) strut.OpConfig {
	return func(op *swag.Operation) {
		op.Tags = append(op.Tags, tags...)
	}
}

func Operation(op *swag.Operation) strut.OpConfig {
	return func(o *swag.Operation) {
		*o = *op
	}
}

func Response(statusCode int, res *swag.OpResponse) strut.OpConfig {
	return func(op *swag.Operation) {
		if op.Responses == nil {
			op.Responses = map[string]*swag.OpResponse{}
		}
		op.Responses[fmt.Sprintf("%d", statusCode)] = res
	}
}

func ResponseDescription(code int, description string) strut.OpConfig {
	return func(op *swag.Operation) {
		if op.Responses == nil {
			op.Responses = map[string]*swag.OpResponse{}
		}

		statusCode := fmt.Sprintf("%d", code)
		if op.Responses[statusCode] == nil {
			op.Responses[statusCode] = &swag.OpResponse{}
		}
		op.Responses[statusCode].Description = description
	}
}

func RequestDescription(description string) strut.OpConfig {
	return func(op *swag.Operation) {
		if op.RequestBody == nil {
			op.RequestBody = &swag.RequestBody{}
		}
		op.RequestBody.Description = description
	}
}
