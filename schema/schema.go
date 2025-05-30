package schema

//https://platform.openai.com/docs/guides/structured-outputs#supported-schemas
//String
//Number
//Boolean
//Integer
//Object
//Array
//Enum
//anyOf

// https://ai.google.dev/gemini-api/docs/structured-output?lang=python
// int
// float
// bool
// str (or enum)
// list[AllowedType]
// dict[str, AllowedType]
//
// anyOf
// enum
// format
// items
// maximum
// minimum
// maxItems
// minItems
// nullable
// properties
// propertyOrdering*
// required

type JSONType string

const (
	Object  JSONType = "object"
	Array   JSONType = "array"
	String  JSONType = "string"
	Number  JSONType = "number"
	Integer JSONType = "integer"
	Boolean JSONType = "boolean"
)

type JSON struct {
	// Only relevant when defining a custom schema, i.e. not using schema.From
	Ref  string           `json:"$ref,omitempty" yaml:"$ref,omitempty"`   // #/$defs/... etc, overrides everything else
	Defs map[string]*JSON `json:"$defs,omitempty" yaml:"$defs,omitempty"` // for $ref

	// JSON Metadata
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Type System
	Type     JSONType `json:"type,omitempty" yaml:"type,omitempty"`
	Nullable bool     `json:"nullable,omitempty" yaml:"nullable,omitempty"`

	// Combinators
	Properties           map[string]*JSON `json:"properties,omitempty" yaml:"properties,omitempty"`                     // for Object
	AdditionalProperties *JSON            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"` // for Map[string]someting...
	Items                *JSON            `json:"items,omitempty" yaml:"items,omitempty"`                               // for Array

	// Validation
	Enum     []interface{} `json:"enum,omitempty" yaml:"enum,omitempty"`
	Required []string      `json:"required,omitempty" yaml:"required,omitempty"`

	/// Number Validation
	Maximum          *float64 `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`

	/// String Validation
	MaxLength *int    `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength *int    `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern   *string `json:"pattern,omitempty" yaml:"pattern,omitempty"` // regular expression
	// Format is used to specify a format for the string, e.g. "date-time", "email", etc.
	Format *string `json:"format,omitempty" yaml:"format,omitempty"` // e.g. "date-time", "email", etc.

	// Array Validation
	MaxItems *int `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems *int `json:"minItems,omitempty" yaml:"minItems,omitempty"`
}
