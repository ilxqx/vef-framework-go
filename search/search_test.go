package search

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ilxqx/vef-framework-go/monad"
)

// Test structs with various field types and conditions

type SimpleSearch struct {
	Name   string `search:"column=name,operator=contains"`
	Age    int    `search:"operator=gte"`
	Active bool   `search:"column=is_active"`
	Salary float64
}

type ComplexSearch struct {
	// String operators
	Title       string `search:"column=title,operator=eq"`
	Description string `search:"operator=contains"`
	Content     string `search:"operator=startsWith"`
	Tags        string `search:"operator=endsWith"`
	Category    string `search:"operator=iContains"`

	// Numeric operators
	MinPrice int     `search:"column=price,operator=gte"`
	MaxPrice int     `search:"column=price,operator=lte"`
	Rating   float64 `search:"operator=gt"`

	// Collection operators
	StatusList string `search:"column=status,operator=in,params=delimiter:|"`
	ExcludeIds string `search:"column=id,operator=notIn"`

	// Range operators
	PriceRange monad.Range[int] `search:"column=price,operator=between"`
	DateRange  string           `search:"column=created_at,operator=between,params=type:date,delimiter:-"`

	// Null check operators
	DeletedAt bool `search:"column=deleted_at,operator=isNull"`
	UpdatedAt bool `search:"column=updated_at,operator=isNotNull"`

	// Multiple columns
	SearchText string `search:"column=title|description|content,operator=contains"`
}

type NestedSearch struct {
	// Nested struct
	User   UserSearch `search:"dive"`
	Status string     `search:"operator:eq"`

	// Another nested struct
	Product ProductSearch `search:"dive"`

	// Regular field
	CreatedAt time.Time `search:"column=created_at,operator=gte"`
}

type UserSearch struct {
	Name     string `search:"column=user_name,operator=contains"`
	Email    string `search:"column=user_email,operator=eq"`
	IsActive bool   `search:"column=user_active"`
}

type ProductSearch struct {
	ProductName string         `search:"column=product_name"`
	Price       float64        `search:"column=product_price,operator=gte"`
	Category    CategorySearch `search:"dive"`
}

type CategorySearch struct {
	Name string `search:"column=category_name,operator=eq"`
	Code string `search:"column=category_code"`
}

type EdgeCaseSearch struct {
	// Field without search tag (will use default eq)
	NoTagField string

	// Field explicitly ignored
	IgnoredField string `search:"-"`

	// Field with empty tag
	EmptyTag string `search:""`

	// Field with only operator
	OnlyOperator string `search:"operator=contains"`

	// Field with custom alias
	CustomAlias string `search:"alias=t1,column=name"`

	// Field with params
	WithArgs string `search:"operator=in,params=delimiter:;,type:int"`

	// Field with default fallback (startsWith as default operator, no operator= specified)
	WithDefault string `search:"startsWith"`

	// Invalid dive field (should log warning)
	InvalidDive string `search:"dive"`
}

func TestNew(t *testing.T) {
	search := NewFor[SimpleSearch]()
	assert.NotNil(t, search.conditions)
	assert.Len(t, search.conditions, 4) // Name, Age, Active, Salary (all included now)
}

func TestNewFromType(t *testing.T) {
	search := New(reflect.TypeOf(SimpleSearch{}))
	assert.NotNil(t, search.conditions)
	assert.Len(t, search.conditions, 4) // All fields included now
}

func TestSimpleSearch(t *testing.T) {
	search := NewFor[SimpleSearch]()

	assert.Len(t, search.conditions, 4) // All fields included now

	// Create expected conditions
	expectedByColumn := map[string]struct {
		operator Operator
		alias    string
		params   map[string]string
	}{
		"name": {
			operator: Contains,
			alias:    "",
			params:   map[string]string{},
		},
		"age": {
			operator: GreaterThanOrEqual,
			alias:    "",
			params:   map[string]string{},
		},
		"is_active": {
			operator: Equals,
			alias:    "",
			params:   map[string]string{},
		},
		"salary": {
			operator: Equals, // Default operator for no-tag field
			alias:    "",
			params:   map[string]string{},
		},
	}

	// Verify each condition
	for _, condition := range search.conditions {
		assert.Len(t, condition.Columns, 1, "Each condition should have exactly one column")

		columnName := condition.Columns[0]
		expected, exists := expectedByColumn[columnName]
		assert.True(t, exists, "Unexpected column: %s", columnName)

		assert.Equal(t, expected.operator, condition.Operator, "Operator mismatch for column %s", columnName)
		assert.Equal(t, expected.alias, condition.Alias, "Alias mismatch for column %s", columnName)
		assert.Equal(t, expected.params, condition.Params, "Params mismatch for column %s", columnName)
	}
}

func TestComplexSearch(t *testing.T) {
	search := NewFor[ComplexSearch]()

	// Verify we have conditions for key fields
	assert.Greater(t, len(search.conditions), 5, "Should have multiple conditions")

	// Test a few key scenarios
	foundMultiColumn := false
	foundWithParams := false
	foundRangeOp := false

	for _, condition := range search.conditions {
		// Check for multi-column condition
		if len(condition.Columns) > 1 {
			foundMultiColumn = true

			assert.Equal(t, Contains, condition.Operator, "Multi-column should use contains")
		}

		// Check for condition with params
		if len(condition.Params) > 0 {
			foundWithParams = true
		}

		// Check for range operators
		if condition.Operator == Between || condition.Operator == NotBetween {
			foundRangeOp = true
		}
	}

	assert.True(t, foundMultiColumn, "Should have multi-column condition")
	assert.True(t, foundWithParams, "Should have condition with params")
	assert.True(t, foundRangeOp, "Should have range operator")
}

func TestNestedSearch(t *testing.T) {
	search := NewFor[NestedSearch]()

	// Should include fields from nested structs after recursion fix
	expectedColumns := []string{
		"user_name", "user_email", "user_active", // from UserSearch
		"status", "created_at", // from NestedSearch
		"product_name", "product_price", // from ProductSearch
		"category_name", "category_code", // from CategorySearch
	}

	assert.Len(t, search.conditions, len(expectedColumns), "Should have all nested fields")

	// Verify all expected columns are present
	for _, expectedCol := range expectedColumns {
		found := false

		for _, condition := range search.conditions {
			if len(condition.Columns) == 1 && condition.Columns[0] == expectedCol {
				found = true

				break
			}
		}

		assert.True(t, found, "Expected column not found: %s", expectedCol)
	}
}

func TestEdgeCases(t *testing.T) {
	search := NewFor[EdgeCaseSearch]()

	// Should ignore: IgnoredField (search:"-") and InvalidDive (dive on non-struct)
	// Should process: NoTagField, EmptyTag, OnlyOperator, CustomAlias, WithParams, WithDefault
	assert.Len(t, search.conditions, 6, "Should have exactly 6 conditions")

	foundWithAlias := false
	foundWithParams := false
	foundDefault := false

	for i, condition := range search.conditions {
		t.Logf("Condition %d: Columns=%v, Operator=%s, Alias=%s, Params=%v",
			i, condition.Columns, condition.Operator, condition.Alias, condition.Params)

		// Check for alias
		if condition.Alias == "t1" {
			foundWithAlias = true

			assert.Equal(t, []string{"name"}, condition.Columns)
		}

		// Check for params
		if len(condition.Params) > 0 {
			foundWithParams = true
		}

		// Check for default operator fallback
		// WithDefault field has `search:"default=startsWith"` which means startsWith is used as default operator
		if condition.Operator == "startsWith" {
			foundDefault = true
		}
	}

	assert.True(t, foundWithAlias, "Should have condition with alias")
	assert.True(t, foundWithParams, "Should have condition with params")
	assert.True(t, foundDefault, "Should have condition with default operator")
}

func TestOperatorShorthand(t *testing.T) {
	// Test the special operator shorthand syntax
	type ShorthandSearch struct {
		// Simple operator shorthand
		Name1 string `search:"eq"`
		Name2 string `search:"contains"`
		Name3 string `search:"startsWith"`

		// Operator shorthand with additional parameters
		Name4 string `search:"contains,column=title|description"`
		Name5 string `search:"in,column=status,params=delimiter:|"`
		Name6 string `search:"gte,column=price"`

		// Mixed: explicit operator= syntax
		Name7 string `search:"operator=endsWith,column=suffix"`
	}

	search := NewFor[ShorthandSearch]()

	expectedConditions := map[string]struct {
		operator Operator
		columns  []string
		params   map[string]string
	}{
		"name_1": { // snake_case conversion
			operator: "eq", // shorthand operator as-is
			columns:  []string{"name_1"},
			params:   map[string]string{},
		},
		"name_2": {
			operator: "contains",
			columns:  []string{"name_2"},
			params:   map[string]string{},
		},
		"name_3": {
			operator: "startsWith",
			columns:  []string{"name_3"},
			params:   map[string]string{},
		},
		"title": { // from first column name in tag
			operator: "contains",
			columns:  []string{"title", "description"},
			params:   map[string]string{},
		},
		"status": {
			operator: "in",
			columns:  []string{"status"},
			params:   map[string]string{"delimiter": "|"},
		},
		"price": {
			operator: "gte",
			columns:  []string{"price"},
			params:   map[string]string{},
		},
		"suffix": {
			operator: "endsWith", // explicit operator= syntax
			columns:  []string{"suffix"},
			params:   map[string]string{},
		},
	}

	assert.Len(t, search.conditions, len(expectedConditions))

	// Verify each condition
	for _, condition := range search.conditions {
		assert.Greater(t, len(condition.Columns), 0, "Should have at least one column")

		// Use first column as key for lookup
		key := condition.Columns[0]
		expected, exists := expectedConditions[key]
		assert.True(t, exists, "Unexpected column: %s", key)

		assert.Equal(t, expected.operator, condition.Operator, "Operator mismatch for %s", key)
		assert.Equal(t, expected.columns, condition.Columns, "Columns mismatch for %s", key)
		assert.Equal(t, expected.params, condition.Params, "Params mismatch for %s", key)
	}
}

func TestNewFromTypeWithNonStruct(t *testing.T) {
	// Test with non-struct types should return empty search
	search := New(reflect.TypeOf("string"))
	assert.Empty(t, search.conditions)

	search = New(reflect.TypeOf(42))
	assert.Empty(t, search.conditions)

	search = New(reflect.TypeOf([]string{}))
	assert.Empty(t, search.conditions)
}

func TestEmptyStruct(t *testing.T) {
	type EmptyStruct struct{}

	search := NewFor[EmptyStruct]()
	assert.Empty(t, search.conditions)
}

func TestStructWithoutSearchTags(t *testing.T) {
	type NoSearchTags struct {
		Name   string
		Age    int
		Active bool
	}

	search := NewFor[NoSearchTags]()
	// Should have 3 conditions with default settings (eq operator, snake_case column)
	assert.Len(t, search.conditions, 3)

	// Verify all fields use default operator
	for _, condition := range search.conditions {
		assert.Equal(t, Equals, condition.Operator)
	}
}

func TestDeepNestedStruct(t *testing.T) {
	type Level3 struct {
		Value string `search:"column=level3_value"`
	}

	type Level2 struct {
		Name   string `search:"column=level2_name"`
		Level3 Level3 `search:"dive"`
	}

	type Level1 struct {
		Title  string `search:"column=level1_title"`
		Level2 Level2 `search:"dive"`
	}

	search := NewFor[Level1]()

	expectedColumns := []string{"level1_title", "level2_name", "level3_value"}

	assert.Len(t, search.conditions, 3, "Should have all deeply nested fields")

	for _, expectedCol := range expectedColumns {
		found := false

		for _, condition := range search.conditions {
			if len(condition.Columns) == 1 && condition.Columns[0] == expectedCol {
				found = true

				break
			}
		}

		assert.True(t, found, "Expected column not found: %s", expectedCol)
	}
}

// TestNoTagStruct tests struct fields without search tags.
type TestNoTagStruct struct {
	Name   string
	Age    int
	Email  string
	Status int `search:"-"` // Explicitly ignored
}

func TestSearch_NoTags(t *testing.T) {
	search := NewFor[TestNoTagStruct]()

	// Should have 3 conditions (Status is ignored)
	assert.Len(t, search.conditions, 3)

	// Verify default operator (eq) and snake_case column names
	assert.Equal(t, Equals, search.conditions[0].Operator)
	assert.Equal(t, []string{"name"}, search.conditions[0].Columns)

	assert.Equal(t, Equals, search.conditions[1].Operator)
	assert.Equal(t, []string{"age"}, search.conditions[1].Columns)

	assert.Equal(t, Equals, search.conditions[2].Operator)
	assert.Equal(t, []string{"email"}, search.conditions[2].Columns)
}
