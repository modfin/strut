package strut_test

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/modfin/strut"
	"github.com/modfin/strut/schema"
	"github.com/modfin/strut/swag"
	"github.com/modfin/strut/with"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Complex test models for OpenAPI schema validation
type Address struct {
	Street     string  `json:"street" json-description:"Street name"`
	Number     int     `json:"number" json-description:"Street number" json-minimum:"1"`
	City       string  `json:"city" json-description:"City name"`
	State      string  `json:"state,omitempty" json-description:"State or province"`
	PostalCode string  `json:"postal_code" json-description:"Postal Status"`
	Country    string  `json:"country" json-description:"Country name"`
	Latitude   float64 `json:"latitude,omitempty" json-description:"Latitude coordinate" json-minimum:"-90" json-maximum:"90"`
	Longitude  float64 `json:"longitude,omitempty" json-description:"Longitude coordinate" json-minimum:"-180" json-maximum:"180"`
}

type Contact struct {
	Email     string `json:"email" json-description:"Email address" json-format:"email"`
	Phone     string `json:"phone,omitempty" json-description:"Phone number"`
	FirstName string `json:"first_name" json-description:"First name"`
	LastName  string `json:"last_name" json-description:"Last name"`
}

type OrderItem struct {
	ProductID   string  `json:"product_id" json-description:"Product identifier"`
	Name        string  `json:"name" json-description:"Product name"`
	Quantity    int     `json:"quantity" json-description:"Quantity ordered" json-minimum:"1"`
	UnitPrice   float64 `json:"unit_price" json-description:"Price per unit" json-minimum:"0"`
	TotalPrice  float64 `json:"total_price" json-description:"Total price for this item" json-minimum:"0"`
	Description string  `json:"description,omitempty" json-description:"Product description"`
}

type PaymentMethod string

const (
	CreditCard PaymentMethod = "credit_card"
	DebitCard  PaymentMethod = "debit_card"
	PayPal     PaymentMethod = "paypal"
	BankWire   PaymentMethod = "bank_wire"
)

type OrderStatus string

const (
	Pending   OrderStatus = "pending"
	Confirmed OrderStatus = "confirmed"
	Shipped   OrderStatus = "shipped"
	Delivered OrderStatus = "delivered"
	Cancelled OrderStatus = "cancelled"
)

type Order struct {
	ID            string            `json:"id" json-description:"Order identifier"`
	CustomerID    string            `json:"customer_id" json-description:"Customer identifier"`
	Contact       Contact           `json:"contact" json-description:"Contact information"`
	Items         []OrderItem       `json:"items" json-description:"Order items" json-min-items:"1"`
	TotalAmount   float64           `json:"total_amount" json-description:"Total order amount" json-minimum:"0"`
	ShippingAddr  Address           `json:"shipping_address" json-description:"Shipping address"`
	BillingAddr   *Address          `json:"billing_address,omitempty" json-description:"Billing address (if different from shipping)"`
	Status        OrderStatus       `json:"Status" json-description:"Order Status" json-enum:"pending,confirmed,shipped,delivered,cancelled"`
	PaymentMethod PaymentMethod     `json:"payment_method" json-description:"Payment method" json-enum:"credit_card,debit_card,paypal,bank_wire"`
	Notes         string            `json:"notes,omitempty" json-description:"Additional notes"`
	Tags          []string          `json:"tags,omitempty" json-description:"Order tags"`
	Metadata      map[string]string `json:"metadata,omitempty" json-description:"Additional metadata"`
}

type CreateOrderRequest struct {
	CustomerID    string            `json:"customer_id" json-description:"Customer identifier"`
	Contact       Contact           `json:"contact" json-description:"Contact information"`
	Items         []OrderItem       `json:"items" json-description:"Order items" json-min-items:"1"`
	ShippingAddr  Address           `json:"shipping_address" json-description:"Shipping address"`
	BillingAddr   *Address          `json:"billing_address,omitempty" json-description:"Billing address (if different from shipping)"`
	PaymentMethod PaymentMethod     `json:"payment_method" json-description:"Payment method" json-enum:"credit_card,debit_card,paypal,bank_wire"`
	Notes         string            `json:"notes,omitempty" json-description:"Additional notes"`
	Tags          []string          `json:"tags,omitempty" json-description:"Order tags"`
	Metadata      map[string]string `json:"metadata,omitempty" json-description:"Additional metadata"`
}

type OrderList struct {
	Orders []Order `json:"orders" json-description:"List of orders"`
	Total  int     `json:"total" json-description:"Total number of orders"`
	Page   int     `json:"page" json-description:"Current page number" json-minimum:"1"`
	Limit  int     `json:"limit" json-description:"Number of items per page" json-minimum:"1"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"Status" json-description:"New order Status" json-enum:"pending,confirmed,shipped,delivered,cancelled"`
	Notes  string      `json:"notes,omitempty" json-description:"Status change notes"`
}

// Handler functions
func GetOrders(ctx context.Context) (res OrderList, err error) {
	// Mock implementation
	return OrderList{
		Orders: []Order{
			{
				ID:         "order-1",
				CustomerID: "cust-1",
				Status:     Pending,
				Items: []OrderItem{
					{
						ProductID:  "prod-1",
						Name:       "Test Product",
						Quantity:   2,
						UnitPrice:  19.99,
						TotalPrice: 39.98,
					},
				},
				TotalAmount: 39.98,
				Contact: Contact{
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
				},
				ShippingAddr: Address{
					Street:     "Main St",
					Number:     123,
					City:       "Anytown",
					PostalCode: "12345",
					Country:    "USA",
				},
				PaymentMethod: CreditCard,
			},
		},
		Total: 1,
		Page:  1,
		Limit: 10,
	}, nil
}

func GetOrderByID(ctx context.Context) (res Order, err error) {
	// Mock implementation
	return Order{
		ID:         "order-1",
		CustomerID: "cust-1",
		Status:     Pending,
		Items: []OrderItem{
			{
				ProductID:  "prod-1",
				Name:       "Test Product",
				Quantity:   2,
				UnitPrice:  19.99,
				TotalPrice: 39.98,
			},
		},
		TotalAmount: 39.98,
		Contact: Contact{
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
		},
		ShippingAddr: Address{
			Street:     "Main St",
			Number:     123,
			City:       "Anytown",
			PostalCode: "12345",
			Country:    "USA",
		},
		PaymentMethod: CreditCard,
	}, nil
}

func CreateOrder(ctx context.Context, req CreateOrderRequest) (res Order, err error) {
	// Mock implementation
	totalAmount := 0.0
	for _, item := range req.Items {
		totalAmount += item.TotalPrice
	}

	return Order{
		ID:            "new-order",
		CustomerID:    req.CustomerID,
		Contact:       req.Contact,
		Items:         req.Items,
		TotalAmount:   totalAmount,
		ShippingAddr:  req.ShippingAddr,
		BillingAddr:   req.BillingAddr,
		Status:        Pending,
		PaymentMethod: req.PaymentMethod,
		Notes:         req.Notes,
		Tags:          req.Tags,
		Metadata:      req.Metadata,
	}, nil
}

func UpdateOrderStatus(ctx context.Context, req UpdateOrderStatusRequest) (res Order, err error) {
	// Mock implementation
	order, _ := GetOrderByID(ctx)
	order.Status = req.Status
	order.Notes = req.Notes
	return order, nil
}

// Test helper to create a test server with complex models
func setupComplexTestServer(t *testing.T) (*strut.Strut, *httptest.Server) {
	r := chi.NewRouter()
	s := strut.New(slog.Default(), r).
		Title("Order Management API").
		Description("API for managing customer orders").
		Version("1.0.0").
		AddServer("https://api.example.com/v1", "Production server").
		AddServer("https://staging-api.example.com/v1", "Staging server")

	// Register endpoints
	strut.Get(s, "/orders", GetOrders,
		with.OperationId("get-orders"),
		with.Description("List all orders with pagination"),
		with.Summary("List orders"),
		with.Tags("orders"),
		with.QueryParam[int]("page", "Page number"),
		with.QueryParam[int]("limit", "Items per page"),
		with.QueryParam[string]("Status", "Filter by order Status"),
		with.QueryParam[string]("customer_id", "Filter by customer ID"),
		with.ResponseDescription(200, "Paginated list of orders"),
	)

	strut.Get(s, "/orders/{id}", GetOrderByID,
		with.OperationId("get-order-by-id"),
		with.Description("Get a specific order by its ID"),
		with.Summary("Get order details"),
		with.Tags("orders"),
		with.PathParam[string]("id", "Order ID"),
		with.ResponseDescription(200, "Order details"),
		with.Response(404, swag.ResponseOf[strut.Error]("Order not found")),
	)

	strut.Post(s, "/orders", CreateOrder,
		with.OperationId("create-order"),
		with.Description("Create a new order"),
		with.Summary("Create order"),
		with.Tags("orders"),
		with.RequestDescription("Order to create"),
		with.ResponseDescription(200, "Created order"),
		with.Response(400, swag.ResponseOf[strut.Error]("Invalid order data")),
	)

	strut.Put(s, "/orders/{id}/Status", UpdateOrderStatus,
		with.OperationId("update-order-Status"),
		with.Description("Update the Status of an existing order"),
		with.Summary("Update order Status"),
		with.Tags("orders"),
		with.PathParam[string]("id", "Order ID"),
		with.RequestDescription("New order Status"),
		with.ResponseDescription(200, "Updated order"),
		with.Response(404, swag.ResponseOf[strut.Error]("Order not found")),
	)

	r.Get("/.well-known/openapi.yaml", s.SchemaHandlerYAML)
	r.Get("/.well-known/openapi.json", s.SchemaHandlerJSON)

	ts := httptest.NewServer(r)

	return s, ts
}

// Tests
func TestComplexOpenAPISpec(t *testing.T) {
	_, ts := setupComplexTestServer(t)
	defer ts.Close()

	// Get the OpenAPI spec in JSON format
	resp, err := http.Get(ts.URL + "/.well-known/openapi.json")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var spec map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&spec)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Basic validation
	assert.Equal(t, "3.0.3", spec["openapi"])

	info, ok := spec["info"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Order Management API", info["title"])
	assert.Equal(t, "API for managing customer orders", info["description"])
	assert.Equal(t, "1.0.0", info["version"])

	// Validate servers
	servers, ok := spec["servers"].([]interface{})
	require.True(t, ok)
	require.Len(t, servers, 2)

	server1 := servers[0].(map[string]interface{})
	assert.Equal(t, "https://api.example.com/v1", server1["url"])
	assert.Equal(t, "Production server", server1["description"])

	server2 := servers[1].(map[string]interface{})
	assert.Equal(t, "https://staging-api.example.com/v1", server2["url"])
	assert.Equal(t, "Staging server", server2["description"])
}

func TestComplexSchemaGeneration(t *testing.T) {
	_, ts := setupComplexTestServer(t)
	defer ts.Close()

	// Get the OpenAPI spec in YAML format for better readability in logs
	resp, err := http.Get(ts.URL + "/.well-known/openapi.yaml")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Parse the YAML
	var spec map[string]interface{}
	err = yaml.Unmarshal(body, &spec)
	require.NoError(t, err)

	// Get components section
	components, ok := spec["components"].(map[string]interface{})
	require.True(t, ok)

	schemas, ok := components["schemas"].(map[string]interface{})
	require.True(t, ok)

	// Check that our complex models are in the schemas
	assert.Contains(t, schemas, "strut_test_Order")
	//assert.Contains(t, schemas, "strut_test_Address") // Not unwinded
	//assert.Contains(t, schemas, "strut_test_Contact") // Not unwinded
	assert.Contains(t, schemas, "strut_test_OrderList")
	assert.Contains(t, schemas, "strut_test_CreateOrderRequest")
	assert.Contains(t, schemas, "strut_test_UpdateOrderStatusRequest")

	// Validate Order schema
	orderSchema, ok := schemas["strut_test_Order"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "object", orderSchema["type"])

	orderProps, ok := orderSchema["properties"].(map[string]interface{})
	require.True(t, ok)

	// Check nested object (Contact)
	contact, ok := orderProps["contact"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Contact information", contact["description"])

	// Check array of objects (Items)
	items, ok := orderProps["items"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "array", items["type"])
	assert.Equal(t, 1, items["minItems"])

	// Check enum field (Status)
	status, ok := orderProps["Status"].(map[string]interface{})
	require.True(t, ok)
	statusEnum, ok := status["enum"].([]interface{})
	require.True(t, ok)
	assert.Len(t, statusEnum, 5)
	assert.Contains(t, statusEnum, "pending")
	assert.Contains(t, statusEnum, "confirmed")
	assert.Contains(t, statusEnum, "shipped")
	assert.Contains(t, statusEnum, "delivered")
	assert.Contains(t, statusEnum, "cancelled")

	// Check nullable field (BillingAddr)
	billingAddr, ok := orderProps["billing_address"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, billingAddr["nullable"])

	// Check map field (Metadata)
	metadata, ok := orderProps["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "object", metadata["type"])
	assert.NotNil(t, metadata["additionalProperties"])
}

// // Nested schemas are not unwinded into separate structs.. It seems to enterprise and DRY fanboy
//func TestNestedObjectSchemas(t *testing.T) {
//	s, _ := setupComplexTestServer(t)
//
//	// Get the schema for Address
//	addrSchema, ok := s.Definition.Components.Schemas["strut_test_Address"]
//	fmt.Printf("%#v\n", s.Definition.Components.Schemas)
//	require.True(t, ok)
//
//	// Validate properties
//	props := addrSchema.Properties
//
//	// String property
//	assert.Equal(t, schema.String, props["street"].Type)
//	assert.Equal(t, "Street name", props["street"].Description)
//
//	// Integer property with minimum
//	assert.Equal(t, schema.Integer, props["number"].Type)
//	assert.Equal(t, 1.0, *props["number"].Minimum)
//
//	// Float property with range
//	assert.Equal(t, schema.Number, props["latitude"].Type)
//	assert.Equal(t, -90.0, *props["latitude"].Minimum)
//	assert.Equal(t, 90.0, *props["latitude"].Maximum)
//
//	// Required fields
//	assert.Contains(t, addrSchema.Required, "street")
//	assert.Contains(t, addrSchema.Required, "number")
//	assert.Contains(t, addrSchema.Required, "city")
//	assert.Contains(t, addrSchema.Required, "postal_code")
//	assert.Contains(t, addrSchema.Required, "country")
//	assert.NotContains(t, addrSchema.Required, "state")     // Optional
//	assert.NotContains(t, addrSchema.Required, "latitude")  // Optional
//	assert.NotContains(t, addrSchema.Required, "longitude") // Optional
//}

func TestArraySchemas(t *testing.T) {
	s, _ := setupComplexTestServer(t)

	// Get the schema for Order
	orderSchema, ok := s.Definition.Components.Schemas["strut_test_Order"]
	require.True(t, ok)

	// Check the items array
	itemsProp := orderSchema.Properties["items"]
	assert.Equal(t, schema.Array, itemsProp.Type)
	assert.Equal(t, 1, *itemsProp.MinItems)

	// Check the items schema
	itemsSchema := itemsProp.Items
	assert.Equal(t, schema.Object, itemsSchema.Type)

	// Check the properties of the items
	itemProps := itemsSchema.Properties
	assert.Equal(t, schema.String, itemProps["product_id"].Type)
	assert.Equal(t, schema.Integer, itemProps["quantity"].Type)
	assert.Equal(t, 1.0, *itemProps["quantity"].Minimum)
	assert.Equal(t, schema.Number, itemProps["unit_price"].Type)
	assert.Equal(t, 0.0, *itemProps["unit_price"].Minimum)
}

func TestEnumSchemas(t *testing.T) {
	s, _ := setupComplexTestServer(t)

	// Get the schema for Order
	orderSchema, ok := s.Definition.Components.Schemas["strut_test_Order"]
	require.True(t, ok)

	// Check the Status enum
	statusProp := orderSchema.Properties["Status"]
	assert.Equal(t, schema.String, statusProp.Type)
	assert.Len(t, statusProp.Enum, 5)
	assert.Contains(t, statusProp.Enum, "pending")
	assert.Contains(t, statusProp.Enum, "confirmed")
	assert.Contains(t, statusProp.Enum, "shipped")
	assert.Contains(t, statusProp.Enum, "delivered")
	assert.Contains(t, statusProp.Enum, "cancelled")

	// Check the payment_method enum
	paymentProp := orderSchema.Properties["payment_method"]
	assert.Equal(t, schema.String, paymentProp.Type)
	assert.Len(t, paymentProp.Enum, 4)
	assert.Contains(t, paymentProp.Enum, "credit_card")
	assert.Contains(t, paymentProp.Enum, "debit_card")
	assert.Contains(t, paymentProp.Enum, "paypal")
	assert.Contains(t, paymentProp.Enum, "bank_wire")
}

func TestMapSchemas(t *testing.T) {
	s, _ := setupComplexTestServer(t)

	// Get the schema for Order
	orderSchema, ok := s.Definition.Components.Schemas["strut_test_Order"]
	require.True(t, ok)

	// Check the metadata map
	metadataProp := orderSchema.Properties["metadata"]
	assert.Equal(t, schema.Object, metadataProp.Type)
	assert.NotNil(t, metadataProp.AdditionalProperties)
	assert.Equal(t, schema.String, metadataProp.AdditionalProperties.Type)
}

func TestPathAndQueryParameters(t *testing.T) {
	s, _ := setupComplexTestServer(t)

	// Check path parameters for /orders/{id}
	pathItem, ok := s.Definition.Paths["/orders/{id}"]
	require.True(t, ok)
	require.NotNil(t, pathItem.Get)

	var idParamFound bool
	for _, param := range pathItem.Get.Parameters {
		if param.Name == "id" && param.In == "path" {
			idParamFound = true
			assert.True(t, param.Required)
			assert.Equal(t, "Order ID", param.Description)
		}
	}
	assert.True(t, idParamFound, "Path parameter 'id' not found")

	// Check query parameters for /orders
	pathItem, ok = s.Definition.Paths["/orders"]
	require.True(t, ok)
	require.NotNil(t, pathItem.Get)

	expectedParams := map[string]struct{}{
		"page":        {},
		"limit":       {},
		"Status":      {},
		"customer_id": {},
	}

	for _, param := range pathItem.Get.Parameters {
		if param.In == "query" {
			_, exists := expectedParams[param.Name]
			assert.True(t, exists, "Unexpected query parameter: %s", param.Name)
			delete(expectedParams, param.Name)

			// Check specific parameter types
			if param.Name == "page" || param.Name == "limit" {
				assert.Equal(t, schema.Integer, param.Schema.Type)
			} else {
				assert.Equal(t, schema.String, param.Schema.Type)
			}
		}
	}

	assert.Empty(t, expectedParams, "Missing query parameters: %v", expectedParams)
}

func TestComplexResponseStatusCodes(t *testing.T) {
	s, _ := setupComplexTestServer(t)

	// Check responses for POST /orders
	pathItem, ok := s.Definition.Paths["/orders"]
	require.True(t, ok)
	require.NotNil(t, pathItem.Post)

	// Should have 200 and 400 responses
	assert.Contains(t, pathItem.Post.Responses, "200")
	assert.Contains(t, pathItem.Post.Responses, "400")

	// Check 400 response
	badRequestResp := pathItem.Post.Responses["400"]
	assert.Equal(t, "Invalid order data", badRequestResp.Description)

	// Check content type
	content, ok := badRequestResp.Content["application/json"]
	require.True(t, ok)
	assert.Contains(t, content.Schema.Properties["message"].Description, "Error message")

}

func TestOpenAPISpecForAITools(t *testing.T) {
	_, ts := setupComplexTestServer(t)
	defer ts.Close()

	// Get the OpenAPI spec in JSON format
	resp, err := http.Get(ts.URL + "/.well-known/openapi.json")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Print the OpenAPI spec for manual inspection
	// This is useful for AI tools to understand the API structure
	//t.Logf("OpenAPI Spec JSON: %s", string(body))

	// Validate that the spec is valid JSON
	var spec map[string]interface{}
	err = json.Unmarshal(body, &spec)
	require.NoError(t, err)

	// Validate that the spec has all required OpenAPI fields
	requiredFields := []string{"openapi", "info", "paths"}
	for _, field := range requiredFields {
		assert.Contains(t, spec, field, "Missing required OpenAPI field: %s", field)
	}

	// Validate that all operations have operationIds
	// This is important for AI tools to reference operations
	paths, ok := spec["paths"].(map[string]interface{})
	require.True(t, ok)

	for path, pathObj := range paths {
		pathItem, ok := pathObj.(map[string]interface{})
		if !ok {
			continue
		}

		for method, opObj := range pathItem {
			if method == "parameters" {
				continue
			}

			op, ok := opObj.(map[string]interface{})
			if !ok {
				continue
			}

			assert.Contains(t, op, "operationId", "Missing operationId for %s %s", method, path)
		}
	}
}
