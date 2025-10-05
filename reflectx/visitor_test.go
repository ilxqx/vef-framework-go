package reflectx

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structures for visitor pattern testing.
type VisitorTestBase struct {
	BaseValue string
}

func (b VisitorTestBase) BaseMethod() string {
	return "base"
}

type VisitorTestEmbedded struct {
	VisitorTestBase

	EmbeddedValue int
	Services      *VisitorTestServices `visit:"dive"`
}

func (e VisitorTestEmbedded) EmbeddedMethod() string {
	return "embedded"
}

type VisitorTestServices struct {
	Logger VisitorTestLogger `visit:"dive"`
	Cache  *VisitorTestCache `visit:"dive"`
}

type VisitorTestLogger struct {
	Level string
}

type VisitorTestCache struct {
	Size int
}

type VisitorTestNested struct {
	VisitorTestEmbedded

	NestedValue bool
}

func TestVisit_DepthFirst(t *testing.T) {
	// Create test structure
	testStruct := VisitorTestNested{
		VisitorTestEmbedded: VisitorTestEmbedded{
			VisitorTestBase: VisitorTestBase{BaseValue: "test"},
			EmbeddedValue:   42,
			Services: &VisitorTestServices{
				Logger: VisitorTestLogger{Level: "info"},
				Cache:  &VisitorTestCache{Size: 100},
			},
		},
		NestedValue: true,
	}

	var (
		visitedStructs []string
		visitedFields  []string
		visitedMethods []string
	)

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
			visitedFields = append(visitedFields, field.Name)

			return Continue
		},
		VisitMethod: func(method reflect.Method, methodValue reflect.Value, depth int) VisitAction {
			visitedMethods = append(visitedMethods, method.Name)

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor)

	// Verify structs are visited in depth-first order
	expectedStructs := []string{"VisitorTestNested", "VisitorTestEmbedded", "VisitorTestBase", "VisitorTestServices", "VisitorTestLogger", "VisitorTestCache"}
	assert.Equal(t, expectedStructs, visitedStructs)

	// Verify key fields are visited
	assert.Contains(t, visitedFields, "NestedValue")
	assert.Contains(t, visitedFields, "EmbeddedValue")
	assert.Contains(t, visitedFields, "BaseValue")
	assert.Contains(t, visitedFields, "Services")
	assert.Contains(t, visitedFields, "Logger")
	assert.Contains(t, visitedFields, "Cache")

	// Verify methods are visited
	assert.Contains(t, visitedMethods, "BaseMethod")
	assert.Contains(t, visitedMethods, "EmbeddedMethod")
}

func TestVisit_BreadthFirst(t *testing.T) {
	testStruct := VisitorTestNested{
		VisitorTestEmbedded: VisitorTestEmbedded{
			VisitorTestBase: VisitorTestBase{BaseValue: "test"},
			EmbeddedValue:   42,
			Services: &VisitorTestServices{
				Logger: VisitorTestLogger{Level: "info"},
				Cache:  &VisitorTestCache{Size: 100},
			},
		},
		NestedValue: true,
	}

	var (
		visitedStructs []string
		depths         []int
	)

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())
			depths = append(depths, depth)

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor, WithTraversalMode(BreadthFirst))

	// Verify breadth-first ordering: structures at same depth should be visited together
	require.Len(t, visitedStructs, 6)
	require.Len(t, depths, 6)

	// Check that depths are generally non-decreasing (breadth-first property)
	assert.Equal(t, "VisitorTestNested", visitedStructs[0])
	assert.Equal(t, 0, depths[0])

	// Find structures at depth 1
	depth1Structs := []string{}

	for i, depth := range depths {
		if depth == 1 {
			depth1Structs = append(depth1Structs, visitedStructs[i])
		}
	}

	assert.Contains(t, depth1Structs, "VisitorTestEmbedded")
}

func TestVisit_MaxDepth(t *testing.T) {
	testStruct := VisitorTestNested{
		VisitorTestEmbedded: VisitorTestEmbedded{
			VisitorTestBase: VisitorTestBase{BaseValue: "test"},
			Services: &VisitorTestServices{
				Logger: VisitorTestLogger{Level: "info"},
			},
		},
	}

	var visitedStructs []string

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor, WithMaxDepth(2))

	// Should not visit deeper structures due to MaxDepth
	assert.NotContains(t, visitedStructs, "VisitorTestLogger")
}

func TestVisit_StopAction(t *testing.T) {
	testStruct := VisitorTestEmbedded{
		VisitorTestBase: VisitorTestBase{BaseValue: "test"},
		EmbeddedValue:   42,
	}

	var visitedFields []string

	visitor := Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
			visitedFields = append(visitedFields, field.Name)
			if field.Name == "EmbeddedValue" {
				return Stop // Stop traversal when we find EmbeddedValue
			}

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor)

	// Should stop after finding EmbeddedValue
	assert.Contains(t, visitedFields, "EmbeddedValue")
	// Verify it actually stopped by checking that later fields are not visited
	laterFieldFound := false
	embeddedValueIndex := -1

	for i, field := range visitedFields {
		if field == "EmbeddedValue" {
			embeddedValueIndex = i

			break
		}
	}

	for i := embeddedValueIndex + 1; i < len(visitedFields); i++ {
		if visitedFields[i] != "" {
			laterFieldFound = true
		}
	}

	assert.False(t, laterFieldFound, "Should not visit fields after Stop action")
}

func TestVisit_SkipChildrenAction(t *testing.T) {
	testStruct := VisitorTestEmbedded{
		VisitorTestBase: VisitorTestBase{BaseValue: "test"},
		Services: &VisitorTestServices{
			Logger: VisitorTestLogger{Level: "info"},
		},
	}

	var visitedStructs []string

	visitor := Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
			if field.Name == "Services" {
				return SkipChildren // Don't traverse into Services
			}

			return Continue
		},
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor)

	// Should not visit VisitorTestServices or its nested structures due to SkipChildren
	assert.NotContains(t, visitedStructs, "VisitorTestServices")
	assert.NotContains(t, visitedStructs, "VisitorTestLogger")
}

func TestVisit_TaggedFields(t *testing.T) {
	testStruct := VisitorTestEmbedded{
		Services: &VisitorTestServices{
			Logger: VisitorTestLogger{Level: "info"},
			Cache:  &VisitorTestCache{Size: 100},
		},
	}

	var visitedStructs []string

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor)

	// Should visit Services due to visit:"dive" tag
	assert.Contains(t, visitedStructs, "VisitorTestServices")
	// Should also visit Cache due to its visit:"dive" tag
	assert.Contains(t, visitedStructs, "VisitorTestCache")
}

func TestVisit_NoRecursion(t *testing.T) {
	testStruct := VisitorTestNested{
		VisitorTestEmbedded: VisitorTestEmbedded{
			VisitorTestBase: VisitorTestBase{BaseValue: "test"},
		},
	}

	var visitedStructs []string

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor, WithDisableRecursive())

	// Should only visit the top-level struct
	assert.Equal(t, []string{"VisitorTestNested"}, visitedStructs)
}

func TestVisit_NilPointer(t *testing.T) {
	var nilStruct *VisitorTestBase

	var visitedStructs []string

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
	}

	Visit(reflect.ValueOf(nilStruct), visitor)

	// Should not visit anything for nil pointer
	assert.Empty(t, visitedStructs)
}

func TestVisit_NonStruct(t *testing.T) {
	testValue := "not a struct"

	var visitedStructs []string

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
	}

	Visit(reflect.ValueOf(testValue), visitor)

	// Should not visit anything for non-struct types
	assert.Empty(t, visitedStructs)
}

// Tests for type-only visitor API

func TestVisitType_DepthFirst(t *testing.T) {
	var (
		visitedTypes   []string
		visitedFields  []string
		visitedMethods []string
	)

	visitor := TypeVisitor{
		VisitStructType: func(structType reflect.Type, depth int) VisitAction {
			visitedTypes = append(visitedTypes, structType.Name())

			return Continue
		},
		VisitFieldType: func(field reflect.StructField, depth int) VisitAction {
			visitedFields = append(visitedFields, field.Name)

			return Continue
		},
		VisitMethodType: func(method reflect.Method, receiverType reflect.Type, depth int) VisitAction {
			visitedMethods = append(visitedMethods, method.Name)

			return Continue
		},
	}

	VisitType(reflect.TypeOf(VisitorTestNested{}), visitor)

	// Verify types are visited in depth-first order
	expectedTypes := []string{"VisitorTestNested", "VisitorTestEmbedded", "VisitorTestBase", "VisitorTestServices", "VisitorTestLogger", "VisitorTestCache"}
	assert.Equal(t, expectedTypes, visitedTypes)

	// Verify key fields are visited
	assert.Contains(t, visitedFields, "NestedValue")
	assert.Contains(t, visitedFields, "EmbeddedValue")
	assert.Contains(t, visitedFields, "BaseValue")
	assert.Contains(t, visitedFields, "Services")
	assert.Contains(t, visitedFields, "Logger")
	assert.Contains(t, visitedFields, "Cache")

	// Verify methods are visited
	assert.Contains(t, visitedMethods, "BaseMethod")
	assert.Contains(t, visitedMethods, "EmbeddedMethod")
}

func TestVisitType_BreadthFirst(t *testing.T) {
	var (
		visitedTypes []string
		depths       []int
	)

	visitor := TypeVisitor{
		VisitStructType: func(structType reflect.Type, depth int) VisitAction {
			visitedTypes = append(visitedTypes, structType.Name())
			depths = append(depths, depth)

			return Continue
		},
	}

	VisitType(reflect.TypeOf(VisitorTestNested{}), visitor, WithTraversalMode(BreadthFirst))

	// Verify breadth-first ordering
	require.Len(t, visitedTypes, 6)
	require.Len(t, depths, 6)

	assert.Equal(t, "VisitorTestNested", visitedTypes[0])
	assert.Equal(t, 0, depths[0])

	// Find types at depth 1
	depth1Types := []string{}

	for i, depth := range depths {
		if depth == 1 {
			depth1Types = append(depth1Types, visitedTypes[i])
		}
	}

	assert.Contains(t, depth1Types, "VisitorTestEmbedded")
}

func TestVisitType_MaxDepth(t *testing.T) {
	var visitedTypes []string

	visitor := TypeVisitor{
		VisitStructType: func(structType reflect.Type, depth int) VisitAction {
			visitedTypes = append(visitedTypes, structType.Name())

			return Continue
		},
	}

	VisitType(reflect.TypeOf(VisitorTestNested{}), visitor, WithMaxDepth(2))

	// Should not visit deeper structures due to MaxDepth
	assert.NotContains(t, visitedTypes, "VisitorTestLogger")
}

func TestVisitType_StopAction(t *testing.T) {
	var visitedTypes []string

	visitor := TypeVisitor{
		VisitStructType: func(structType reflect.Type, depth int) VisitAction {
			visitedTypes = append(visitedTypes, structType.Name())
			if structType.Name() == "VisitorTestEmbedded" {
				return Stop // Stop traversal when we find VisitorTestEmbedded
			}

			return Continue
		},
	}

	VisitType(reflect.TypeOf(VisitorTestNested{}), visitor)

	// Should stop after finding VisitorTestEmbedded
	assert.Contains(t, visitedTypes, "VisitorTestEmbedded")
	assert.NotContains(t, visitedTypes, "VisitorTestBase")
}

func TestVisitType_SkipChildrenAction(t *testing.T) {
	var visitedTypes []string

	visitor := TypeVisitor{
		VisitFieldType: func(field reflect.StructField, depth int) VisitAction {
			if field.Name == "Services" {
				return SkipChildren // Don't traverse into Services
			}

			return Continue
		},
		VisitStructType: func(structType reflect.Type, depth int) VisitAction {
			visitedTypes = append(visitedTypes, structType.Name())

			return Continue
		},
	}

	VisitType(reflect.TypeOf(VisitorTestEmbedded{}), visitor)

	// Should not visit VisitorTestServices or its nested structures due to SkipChildren
	assert.NotContains(t, visitedTypes, "VisitorTestServices")
	assert.NotContains(t, visitedTypes, "VisitorTestLogger")
}

func TestVisitType_NonStruct(t *testing.T) {
	var visitedTypes []string

	visitor := TypeVisitor{
		VisitStructType: func(structType reflect.Type, depth int) VisitAction {
			visitedTypes = append(visitedTypes, structType.Name())

			return Continue
		},
	}

	VisitType(reflect.TypeOf("not a struct"), visitor)

	// Should not visit anything for non-struct types
	assert.Empty(t, visitedTypes)
}

func TestVisitType_PointerToStruct(t *testing.T) {
	var visitedTypes []string

	visitor := TypeVisitor{
		VisitStructType: func(structType reflect.Type, depth int) VisitAction {
			visitedTypes = append(visitedTypes, structType.Name())

			return Continue
		},
	}

	VisitType(reflect.TypeOf((*VisitorTestBase)(nil)), visitor)

	// Should visit the underlying struct type
	assert.Contains(t, visitedTypes, "VisitorTestBase")
}

func TestMethodVisitor_CallableMethodValue(t *testing.T) {
	testStruct := VisitorTestEmbedded{
		VisitorTestBase: VisitorTestBase{BaseValue: "test_value"},
	}

	var methodResults []string

	visitor := Visitor{
		VisitMethod: func(method reflect.Method, methodValue reflect.Value, depth int) VisitAction {
			if method.Name == "BaseMethod" {
				// methodValue 应该是可调用的方法值，已经绑定了接收者
				results := methodValue.Call(nil)
				if len(results) > 0 {
					methodResults = append(methodResults, results[0].String())
				}
			}

			return Continue
		},
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			// 只访问第一层结构，避免重复
			if depth > 0 {
				return SkipChildren
			}

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor)

	// 验证我们能够直接调用 methodValue
	assert.Len(t, methodResults, 1)
	assert.Equal(t, "base", methodResults[0])
}

func TestVisitor_NilCheckBehavior(t *testing.T) {
	testStruct := VisitorTestEmbedded{
		VisitorTestBase: VisitorTestBase{BaseValue: "test"},
	}

	var (
		visitedStructs []string
		visitedFields  []string
		visitedMethods []string
	)

	// Test with only struct visitor (no field or method visitors)
	visitor1 := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
		// VisitField and VisitMethod are nil - should not be called
	}

	Visit(reflect.ValueOf(testStruct), visitor1)

	// Should visit structs but not fields or methods
	assert.Contains(t, visitedStructs, "VisitorTestEmbedded")
	assert.Contains(t, visitedStructs, "VisitorTestBase")
	assert.Empty(t, visitedFields)
	assert.Empty(t, visitedMethods)

	// Reset and test with all visitors
	visitedStructs = nil
	visitedFields = nil
	visitedMethods = nil

	visitor2 := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
			visitedFields = append(visitedFields, field.Name)

			return Continue
		},
		VisitMethod: func(method reflect.Method, methodValue reflect.Value, depth int) VisitAction {
			visitedMethods = append(visitedMethods, method.Name)

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor2)

	// Should visit structs, fields, and methods
	assert.Contains(t, visitedStructs, "VisitorTestEmbedded")
	assert.Contains(t, visitedFields, "BaseValue")
	assert.Contains(t, visitedMethods, "BaseMethod")
}

func TestVisitFor_Generic(t *testing.T) {
	var visitedTypes []string

	visitor := TypeVisitor{
		VisitStructType: func(structType reflect.Type, depth int) VisitAction {
			visitedTypes = append(visitedTypes, structType.Name())

			return Continue
		},
	}

	// Use the generic convenience function
	VisitFor[VisitorTestNested](visitor)

	// Should visit all struct types in the hierarchy
	assert.Contains(t, visitedTypes, "VisitorTestNested")
	assert.Contains(t, visitedTypes, "VisitorTestEmbedded")
	assert.Contains(t, visitedTypes, "VisitorTestBase")
}

func TestVisitOf_Convenience(t *testing.T) {
	testStruct := VisitorTestEmbedded{
		VisitorTestBase: VisitorTestBase{BaseValue: "test"},
		EmbeddedValue:   42,
	}

	var (
		visitedStructs []string
		visitedFields  []string
	)

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
			visitedFields = append(visitedFields, field.Name)

			return Continue
		},
	}

	// Use the convenience function
	VisitOf(testStruct, visitor)

	// Should visit structs and fields
	assert.Contains(t, visitedStructs, "VisitorTestEmbedded")
	assert.Contains(t, visitedStructs, "VisitorTestBase")
	assert.Contains(t, visitedFields, "BaseValue")
	assert.Contains(t, visitedFields, "EmbeddedValue")
}

// Test edge cases and boundary conditions

func TestVisit_EmptyStruct(t *testing.T) {
	type EmptyStruct struct{}

	testStruct := EmptyStruct{}

	var (
		visitedStructs []string
		visitedFields  []string
		visitedMethods []string
	)

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
			visitedFields = append(visitedFields, field.Name)

			return Continue
		},
		VisitMethod: func(method reflect.Method, methodValue reflect.Value, depth int) VisitAction {
			visitedMethods = append(visitedMethods, method.Name)

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor)

	assert.Equal(t, []string{"EmptyStruct"}, visitedStructs)
	assert.Empty(t, visitedFields)
	assert.Empty(t, visitedMethods)
}

func TestVisit_UnexportedFields(t *testing.T) {
	type StructWithUnexportedFields struct {
		PublicField  string
		privateField int // Should be skipped
	}

	testStruct := StructWithUnexportedFields{
		PublicField:  "public",
		privateField: 42,
	}

	var visitedFields []string

	visitor := Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
			visitedFields = append(visitedFields, field.Name)

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor)

	assert.Equal(t, []string{"PublicField"}, visitedFields)
}

func TestVisit_MultiplePointerLevels(t *testing.T) {
	testStruct := VisitorTestBase{BaseValue: "test"}
	ptrToStruct := &testStruct
	ptrToPtr := &ptrToStruct

	var visitedStructs []string

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
	}

	Visit(reflect.ValueOf(ptrToPtr), visitor)

	require.Len(t, visitedStructs, 1)
	assert.Equal(t, "VisitorTestBase", visitedStructs[0])
}

func TestVisit_InvalidValue(t *testing.T) {
	var visitedStructs []string

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
	}

	// Test with zero value (invalid)
	var invalidValue reflect.Value
	Visit(invalidValue, visitor)

	assert.Empty(t, visitedStructs)
}

func TestVisit_CyclicReference(t *testing.T) {
	// Use struct with pointer to itself to test cycle detection
	type SelfReferencing struct {
		Value string
		Next  *SelfReferencing `visit:"dive"`
	}

	node1 := &SelfReferencing{Value: "node1"}
	node2 := &SelfReferencing{Value: "node2"}
	node1.Next = node2
	node2.Next = node1 // Create cycle

	var visitedStructs []string

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name()+"_"+structValue.FieldByName("Value").String())

			return Continue
		},
	}

	Visit(reflect.ValueOf(node1), visitor)

	// Should visit each instance, but prevent infinite recursion
	// Due to cycle detection, the same struct type should not cause infinite loop
	assert.NotEmpty(t, visitedStructs)
	// The exact behavior depends on implementation, but it shouldn't hang
	assert.True(t, len(visitedStructs) < 10, "Should not visit too many instances due to cycle detection")
}

func TestVisit_MethodsOnNonAddressableValue(t *testing.T) {
	// Create non-addressable value (result of function call)
	getValue := func() VisitorTestBase {
		return VisitorTestBase{BaseValue: "test"}
	}

	var visitedMethods []string

	visitor := Visitor{
		VisitMethod: func(method reflect.Method, methodValue reflect.Value, depth int) VisitAction {
			visitedMethods = append(visitedMethods, method.Name)

			return Continue
		},
	}

	Visit(reflect.ValueOf(getValue()), visitor)

	// Should be able to visit methods on non-addressable values
	// VisitorTestBase has BaseMethod defined
	assert.Contains(t, visitedMethods, "BaseMethod")
}

func TestVisit_MaxDepthZero(t *testing.T) {
	testStruct := VisitorTestNested{
		VisitorTestEmbedded: VisitorTestEmbedded{
			VisitorTestBase: VisitorTestBase{BaseValue: "base"},
			EmbeddedValue:   42,
		},
		NestedValue: true,
	}

	var visitedStructs []string

	visitor := Visitor{
		VisitStruct: func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
	}

	// MaxDepth 0 should visit root struct and embedded anonymous fields (depth 0)
	// but not dive into tagged fields (depth 1+)
	Visit(reflect.ValueOf(testStruct), visitor, WithMaxDepth(0))

	// VisitorTestNested embeds VisitorTestEmbedded anonymously,
	// and VisitorTestEmbedded embeds VisitorTestBase anonymously
	// All of these are visited at depth 0 due to anonymous embedding
	assert.Contains(t, visitedStructs, "VisitorTestNested")
	assert.Contains(t, visitedStructs, "VisitorTestEmbedded")
	assert.Contains(t, visitedStructs, "VisitorTestBase")
}

func TestVisitType_WithNilVisitors(t *testing.T) {
	var visitedStructs []string

	visitor := TypeVisitor{
		// Only set VisitStructType, leave others nil
		VisitStructType: func(structType reflect.Type, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
		// VisitFieldType and VisitMethodType are nil
	}

	VisitType(reflect.TypeFor[VisitorTestBase](), visitor)

	assert.Equal(t, []string{"VisitorTestBase"}, visitedStructs)
}
