package with

import (
	"fmt"
	"gihub.com/modfin/strut"
	"gihub.com/modfin/strut/schema"
)

func Description(description string) strut.OpConfig {
	return func(op *strut.Operation) {
		op.Description = description
	}
}
func OperationId(operationId string) strut.OpConfig {
	return func(op *strut.Operation) {
		op.OperationID = operationId
	}
}

func Summary(summary string) strut.OpConfig {
	return func(op *strut.Operation) {
		op.Summary = summary
	}
}

func Param(param ...strut.Param) strut.OpConfig {
	return func(op *strut.Operation) {
		op.Parameters = append(op.Parameters, param...)
	}
}

func QueryParam[T any](name string, description string) strut.OpConfig {
	var ref T
	return func(op *strut.Operation) {
		op.Parameters = append(op.Parameters, strut.Param{
			Name:        name,
			In:          "query",
			Description: description,
			Schema:      schema.From(ref),
		})
	}
}
func PathParam[T any](name string, description string) strut.OpConfig {
	var ref T
	return func(op *strut.Operation) {
		op.Parameters = append(op.Parameters, strut.Param{
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
	return func(op *strut.Operation) {
		op.Parameters = append(op.Parameters, strut.Param{
			Name:        name,
			In:          "cookie",
			Description: description,
			Schema:      schema.From(ref),
		})
	}
}
func HeaderParam[T any](name string, description string) strut.OpConfig {
	var ref T
	return func(op *strut.Operation) {
		op.Parameters = append(op.Parameters, strut.Param{
			Name:        name,
			In:          "header",
			Description: description,
			Schema:      schema.From(ref),
		})
	}
}

func Deprecated() strut.OpConfig {
	return func(op *strut.Operation) {
		op.Deprecated = true
	}
}

func Tags(tags ...string) strut.OpConfig {
	return func(op *strut.Operation) {
		op.Tags = append(op.Tags, tags...)
	}
}

func Operation(op *strut.Operation) strut.OpConfig {
	return func(o *strut.Operation) {
		*o = *op
	}
}

func Response(statusCode int, res *strut.Response) strut.OpConfig {
	return func(op *strut.Operation) {
		if op.Responses == nil {
			op.Responses = map[string]*strut.Response{}
		}
		op.Responses[fmt.Sprintf("%d", statusCode)] = res
	}
}

func ResponseDescription(description string) strut.OpConfig {
	return func(op *strut.Operation) {
		if op.Responses == nil {
			op.Responses = map[string]*strut.Response{}
		}

		if op.Responses["200"] == nil {
			op.Responses["200"] = &strut.Response{}
		}
		op.Responses["200"].Description = description
	}
}

func RequestDescription(description string) strut.OpConfig {
	return func(op *strut.Operation) {
		if op.RequestBody == nil {
			op.RequestBody = &strut.RequestBody{}
		}
		op.RequestBody.Description = description
	}
}
