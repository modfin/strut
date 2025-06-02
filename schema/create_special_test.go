package schema_test

import (
	"github.com/modfin/strut/schema"
	"reflect"
	"testing"
)

func TestFrom_StructWithAnonymousFields(t *testing.T) {
	// Test handling of structs with anonymous fields
	type StructWithAnonymousFields struct {
		string
		int
		bool
	}

	// Anonymous fields are not exported in JSON by default
	expected := &schema.JSON{
		Type:       schema.Object,
		Properties: map[string]*schema.JSON{},
	}

	result := schema.From(StructWithAnonymousFields{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_ArrayOfMaps(t *testing.T) {
	// Test handling of arrays of maps
	type ArrayOfMapsStruct struct {
		Configs []map[string]int `json:"configs"`
	}

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"configs": {
				Type: schema.Array,
				Items: &schema.JSON{
					Type:                 schema.Object,
					AdditionalProperties: &schema.JSON{Type: schema.Integer},
					Properties:           map[string]*schema.JSON{},
				},
			},
		},
		Required: []string{"configs"},
	}

	result := schema.From(ArrayOfMapsStruct{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_StructWithJsonNameOverrides(t *testing.T) {
	// Test handling of structs with JSON name overrides
	type StructWithJsonNameOverrides struct {
		UserName  string `json:"user_name"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"user_name":  {Type: schema.String},
			"first_name": {Type: schema.String},
			"last_name":  {Type: schema.String},
		},
		Required: []string{"user_name", "first_name", "last_name"},
	}

	result := schema.From(StructWithJsonNameOverrides{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_StructWithMultipleOmitEmpty(t *testing.T) {
	// Test handling of structs with multiple omitempty fields
	type StructWithMultipleOmitEmpty struct {
		ID        int     `json:"id"`
		Name      string  `json:"name,omitempty"`
		Email     string  `json:"email,omitempty"`
		Age       int     `json:"age,omitempty"`
		IsActive  bool    `json:"is_active,omitempty"`
		Score     float64 `json:"score,omitempty"`
	}

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"id":        {Type: schema.Integer},
			"name":      {Type: schema.String},
			"email":     {Type: schema.String},
			"age":       {Type: schema.Integer},
			"is_active": {Type: schema.Boolean},
			"score":     {Type: schema.Number},
		},
		Required: []string{"id"},
	}

	result := schema.From(StructWithMultipleOmitEmpty{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_StructWithMixedArrayTypes(t *testing.T) {
	// Test handling of structs with arrays of different types
	type StructWithMixedArrayTypes struct {
		StringArray []string  `json:"string_array"`
		IntArray    []int     `json:"int_array"`
		FloatArray  []float64 `json:"float_array"`
		BoolArray   []bool    `json:"bool_array"`
	}

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"string_array": {
				Type:  schema.Array,
				Items: &schema.JSON{Type: schema.String},
			},
			"int_array": {
				Type:  schema.Array,
				Items: &schema.JSON{Type: schema.Integer},
			},
			"float_array": {
				Type:  schema.Array,
				Items: &schema.JSON{Type: schema.Number},
			},
			"bool_array": {
				Type:  schema.Array,
				Items: &schema.JSON{Type: schema.Boolean},
			},
		},
		Required: []string{"string_array", "int_array", "float_array", "bool_array"},
	}

	result := schema.From(StructWithMixedArrayTypes{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_StructWithNestedAnonymousStruct(t *testing.T) {
	// Test handling of structs with nested anonymous structs
	type StructWithNestedAnonymousStruct struct {
		Name string `json:"name"`
		Data struct {
			Value int    `json:"value"`
			Key   string `json:"key"`
		} `json:"data"`
	}

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"name": {Type: schema.String},
			"data": {
				Type: schema.Object,
				Properties: map[string]*schema.JSON{
					"value": {Type: schema.Integer},
					"key":   {Type: schema.String},
				},
				Required: []string{"value", "key"},
			},
		},
		Required: []string{"name", "data"},
	}

	result := schema.From(StructWithNestedAnonymousStruct{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_StructWithNestedPointerToAnonymousStruct(t *testing.T) {
	// Test handling of structs with nested pointers to anonymous structs
	type StructWithNestedPointerToAnonymousStruct struct {
		Name string `json:"name"`
		Data *struct {
			Value int    `json:"value"`
			Key   string `json:"key"`
		} `json:"data"`
	}

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"name": {Type: schema.String},
			"data": {
				Type:     schema.Object,
				Nullable: true,
				Properties: map[string]*schema.JSON{
					"value": {Type: schema.Integer},
					"key":   {Type: schema.String},
				},
				Required: []string{"value", "key"},
			},
		},
		Required: []string{"name", "data"},
	}

	result := schema.From(StructWithNestedPointerToAnonymousStruct{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}
