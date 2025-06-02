package schema_test

import (
	"github.com/modfin/strut/schema"
	"reflect"
	"testing"
)

func TestFrom_MultipleTagsOnSameField(t *testing.T) {
	// Test handling of multiple validation tags on the same field
	type MultiTagStruct struct {
		Name string `json:"name" json-min-length:"5" json-max-length:"50" json-pattern:"^[a-zA-Z]+$" json-description:"User's full name"`
	}

	minLength := 5
	maxLength := 50
	pattern := "^[a-zA-Z]+$"
	description := "User's full name"

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"name": {
				Type:        schema.String,
				MinLength:   &minLength,
				MaxLength:   &maxLength,
				Pattern:     &pattern,
				Description: description,
			},
		},
		Required: []string{"name"},
	}

	result := schema.From(MultiTagStruct{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_ZeroValues(t *testing.T) {
	// Test handling of zero values in fields
	type ZeroValueStruct struct {
		ID        int     `json:"id"`
		Name      string  `json:"name"`
		IsActive  bool    `json:"is_active"`
		Score     float64 `json:"score"`
		EmptyList []int   `json:"empty_list"`
	}

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"id":        {Type: schema.Integer},
			"name":      {Type: schema.String},
			"is_active": {Type: schema.Boolean},
			"score":     {Type: schema.Number},
			"empty_list": {
				Type:  schema.Array,
				Items: &schema.JSON{Type: schema.Integer},
			},
		},
		Required: []string{"id", "name", "is_active", "score", "empty_list"},
	}

	result := schema.From(ZeroValueStruct{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_NestedMapWithArrays(t *testing.T) {
	// Test handling of complex nested structures with maps and arrays
	type NestedMapArrayStruct struct {
		Data map[string][]int `json:"data"`
	}

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"data": {
				Type: schema.Object,
				AdditionalProperties: &schema.JSON{
					Type:  schema.Array,
					Items: &schema.JSON{Type: schema.Integer},
				},
				Properties: map[string]*schema.JSON{},
			},
		},
		Required: []string{"data"},
	}

	result := schema.From(NestedMapArrayStruct{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_MapWithEnumKeys(t *testing.T) {
	// Test handling of maps with specific key types
	type MapWithEnumKeysStruct struct {
		Settings map[string]string `json:"settings" json-enum:"theme,language,timezone"`
	}

	// Note: The schema package doesn't currently support enum validation for map keys,
	// but this test ensures it doesn't break when such tags are provided
	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"settings": {
				Type:                 schema.Object,
				AdditionalProperties: &schema.JSON{Type: schema.String},
				Properties:           map[string]*schema.JSON{},
			},
		},
		Required: []string{"settings"},
	}

	result := schema.From(MapWithEnumKeysStruct{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_ArrayWithMultipleValidations(t *testing.T) {
	// Test handling of arrays with multiple validation constraints
	type ArrayValidationStruct struct {
		Tags []string `json:"tags" json-min-items:"1" json-max-items:"10" json-enum:"work,personal,important"`
	}

	minItems := 1
	maxItems := 10

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"tags": {
				Type:     schema.Array,
				MinItems: &minItems,
				MaxItems: &maxItems,
				Items: &schema.JSON{
					Type: schema.String,
					Enum: []interface{}{"work", "personal", "important"},
				},
			},
		},
		Required: []string{"tags"},
	}

	result := schema.From(ArrayValidationStruct{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_MultipleStructsWithSameField(t *testing.T) {
	// Test handling of multiple structs with the same field name but different types
	type Struct1 struct {
		Value int `json:"value"`
	}

	type Struct2 struct {
		Value string `json:"value"`
	}

	expected1 := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"value": {Type: schema.Integer},
		},
		Required: []string{"value"},
	}

	expected2 := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"value": {Type: schema.String},
		},
		Required: []string{"value"},
	}

	result1 := schema.From(Struct1{})
	result2 := schema.From(Struct2{})

	if !reflect.DeepEqual(result1, expected1) {
		t.Errorf("Expected %+v, got %+v", expected1, result1)
	}

	if !reflect.DeepEqual(result2, expected2) {
		t.Errorf("Expected %+v, got %+v", expected2, result2)
	}
}

func TestFrom_PointerToMap(t *testing.T) {
	// Test handling of pointer to map
	type PointerMapStruct struct {
		Settings *map[string]string `json:"settings"`
	}

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"settings": {
				Type:                 schema.Object,
				Nullable:             true,
				AdditionalProperties: &schema.JSON{Type: schema.String},
				Properties:           map[string]*schema.JSON{},
			},
		},
		Required: []string{"settings"},
	}

	result := schema.From(PointerMapStruct{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

func TestFrom_PointerToArray(t *testing.T) {
	// Test handling of pointer to array
	type PointerArrayStruct struct {
		Items *[]int `json:"items"`
	}

	expected := &schema.JSON{
		Type: schema.Object,
		Properties: map[string]*schema.JSON{
			"items": {
				Type:     schema.Array,
				Nullable: true,
				Items:    &schema.JSON{Type: schema.Integer},
			},
		},
		Required: []string{"items"},
	}

	result := schema.From(PointerArrayStruct{})
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}
