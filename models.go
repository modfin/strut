package strut

import "gihub.com/modfin/strut/schema"

type Definition struct {
	OpenAPI    string          `json:"openapi,omitempty" yaml:"openapi,omitempty"`
	Info       Info            `json:"info,omitempty" yaml:"info,omitempty"`
	Paths      map[string]Path `json:"paths,omitempty" yaml:"paths,omitempty"`
	Components *Components     `json:"components,omitempty" yaml:"components,omitempty"`
	Servers    []Server        `json:"servers,omitempty" yaml:"servers,omitempty"`
}

type Info struct {
	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string `json:"version,omitempty" yaml:"version,omitempty"`
}

type Path struct {
	Summary     string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Parameters []Param `json:"parameters,omitempty" yaml:"parameters,omitempty"`

	Post *Operation `json:"post,omitempty" yaml:"post,omitempty"`

	//Get    *Operation `json:"get,omitempty" yaml:"get,omitempty"`
	//Put    *Operation `json:"put,omitempty" yaml:"put,omitempty"`
	//Delete *Operation `json:"delete,omitempty" yaml:"delete,omitempty"`
	//Patch  *Operation `json:"patch,omitempty" yaml:"patch,omitempty"`
	//Head   *Operation `json:"head,omitempty" yaml:"head,omitempty"`
}

// Operation represents an HTTP operation on a path
type Operation struct {
	Tags        []string             `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string               `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	OperationID string               `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  []Param              `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody         `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]*Response `json:"responses,omitempty" yaml:"responses,omitempty"`
	Deprecated  bool                 `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

// Param represents a parameter for an operation
type Param struct {
	Name            string       `json:"name,omitempty" yaml:"name,omitempty"`
	In              string       `json:"in,omitempty" yaml:"in,omitempty"` // e.g., "query", "header", "path", "cookie"
	Description     string       `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool         `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool         `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool         `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Schema          *schema.JSON `json:"schema,omitempty" yaml:"schema,omitempty"`
}

// RequestBody represents a request body
type RequestBody struct {
	Required    bool                 `json:"required,omitempty" yaml:"required,omitempty"`
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// Response represents a response from an API operation
type Response struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
	//Headers     map[string]Header    `json:"headers,omitempty" yaml:"headers,omitempty"`
	//Links       map[string]Link      `json:"links,omitempty" yaml:"links,omitempty"`
}

// MediaType represents a media type object
type MediaType struct {
	Schema   *schema.JSON        `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

// Header represents a header object
type Header struct {
	Description     string       `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool         `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool         `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool         `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string       `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         bool         `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool         `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *schema.JSON `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         interface{}  `json:"example,omitempty" yaml:"example,omitempty"`
}

// Example represents an example object
type Example struct {
	Summary       string      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string      `json:"description,omitempty" yaml:"description,omitempty"`
	Value         interface{} `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// Encoding represents an encoding object
type Encoding struct {
	ContentType   string            `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string            `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       bool              `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved bool              `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
}

// Link represents a link object
type Link struct {
	OperationRef string                 `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
	OperationID  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  interface{}            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	//Server       *Server                `json:"server,omitempty" yaml:"server,omitempty"`
}

// SecurityRequirement represents a security requirement
type SecurityRequirement map[string][]string

type Components struct {
	//SecuritySchemes map[string]SecurityScheme `json:"securitySchemes" yaml:"securitySchemes"`
	Schemas map[string]*schema.JSON `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	//Responses  map[string]Response     `json:"responses" yaml:"responses"`
	//Parameters map[string]Param        `json:"parameters" yaml:"parameters"`
	//Examples   map[string]Example      `json:"examples" yaml:"examples"`
}

// Server represents a server object
type Server struct {
	URL         string                    `json:"url" yaml:"url"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
}
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}
