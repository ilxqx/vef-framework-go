package reflectx

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	expectedStructs := []string{"VisitorTestNested", "VisitorTestEmbedded", "VisitorTestBase", "VisitorTestServices", "VisitorTestLogger", "VisitorTestCache"}
	assert.Equal(t, expectedStructs, visitedStructs, "Structs should be visited in depth-first order")

	assert.Contains(t, visitedFields, "NestedValue", "Should visit NestedValue field")
	assert.Contains(t, visitedFields, "EmbeddedValue", "Should visit EmbeddedValue field")
	assert.Contains(t, visitedFields, "BaseValue", "Should visit BaseValue field")
	assert.Contains(t, visitedFields, "Services", "Should visit Services field")
	assert.Contains(t, visitedFields, "Logger", "Should visit Logger field")
	assert.Contains(t, visitedFields, "Cache", "Should visit Cache field")

	assert.Contains(t, visitedMethods, "BaseMethod", "Should visit BaseMethod")
	assert.Contains(t, visitedMethods, "EmbeddedMethod", "Should visit EmbeddedMethod")
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

	require.Len(t, visitedStructs, 6, "Should visit 6 structs total")
	require.Len(t, depths, 6, "Should record 6 depths")

	assert.Equal(t, "VisitorTestNested", visitedStructs[0], "Root struct should be visited first")
	assert.Equal(t, 0, depths[0], "Root struct should be at depth 0")

	depth1Structs := []string{}

	for i, depth := range depths {
		if depth == 1 {
			depth1Structs = append(depth1Structs, visitedStructs[i])
		}
	}

	assert.Contains(t, depth1Structs, "VisitorTestEmbedded", "Embedded struct should be at depth 1")
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

	assert.NotContains(t, visitedStructs, "VisitorTestLogger", "Should not visit deeper structures due to MaxDepth")
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

	assert.Contains(t, visitedFields, "EmbeddedValue", "Should visit EmbeddedValue before stopping")

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

// Tests for type-only visitor Api

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
		VisitStructType: func(structType reflect.Type, depth int) VisitAction {
			visitedStructs = append(visitedStructs, structType.Name())

			return Continue
		},
		// VisitFieldType and VisitMethodType are nil
	}

	VisitType(reflect.TypeFor[VisitorTestBase](), visitor)

	assert.Equal(t, []string{"VisitorTestBase"}, visitedStructs)
}

// Tests for field index path tracking in embedded structures

func TestVisit_FieldIndexPath_AnonymousEmbedded(t *testing.T) {
	// Test that anonymous embedded fields have correct index paths
	testStruct := VisitorTestNested{
		VisitorTestEmbedded: VisitorTestEmbedded{
			VisitorTestBase: VisitorTestBase{BaseValue: "test"},
			EmbeddedValue:   42,
		},
		NestedValue: true,
	}

	fieldIndexMap := make(map[string][]int)

	visitor := Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
			fieldIndexMap[field.Name] = field.Index

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor)

	// Verify nested field index paths
	// BaseValue is in VisitorTestBase (embedded in VisitorTestEmbedded, which is embedded in VisitorTestNested)
	assert.NotNil(t, fieldIndexMap["BaseValue"])
	assert.Equal(t, []int{0, 0, 0}, fieldIndexMap["BaseValue"], "BaseValue should have path [0,0,0]")

	// EmbeddedValue is in VisitorTestEmbedded (embedded in VisitorTestNested)
	assert.NotNil(t, fieldIndexMap["EmbeddedValue"])
	assert.Equal(t, []int{0, 1}, fieldIndexMap["EmbeddedValue"], "EmbeddedValue should have path [0,1]")

	// NestedValue is a direct field of VisitorTestNested
	assert.NotNil(t, fieldIndexMap["NestedValue"])
	assert.Equal(t, []int{1}, fieldIndexMap["NestedValue"], "NestedValue should have path [1]")
}

func TestVisitType_FieldIndexPath_TaggedDive(t *testing.T) {
	// Test that non-anonymous fields with dive tag have correct index paths
	fieldIndexMap := make(map[string][]int)

	visitor := TypeVisitor{
		VisitFieldType: func(field reflect.StructField, depth int) VisitAction {
			fieldIndexMap[field.Name] = field.Index

			return Continue
		},
	}

	VisitType(reflect.TypeOf(VisitorTestEmbedded{}), visitor)

	// Services field is at [2] in VisitorTestEmbedded
	assert.NotNil(t, fieldIndexMap["Services"])
	assert.Equal(t, []int{2}, fieldIndexMap["Services"], "Services should have path [2]")

	// Logger is inside Services, which has dive tag
	assert.NotNil(t, fieldIndexMap["Logger"])
	assert.Equal(t, []int{2, 0}, fieldIndexMap["Logger"], "Logger should have path [2,0]")

	// Level is inside Logger
	assert.NotNil(t, fieldIndexMap["Level"])
	assert.Equal(t, []int{2, 0, 0}, fieldIndexMap["Level"], "Level should have path [2,0,0]")

	// Cache is inside Services
	assert.NotNil(t, fieldIndexMap["Cache"])
	assert.Equal(t, []int{2, 1}, fieldIndexMap["Cache"], "Cache should have path [2,1]")

	// Size is inside Cache
	assert.NotNil(t, fieldIndexMap["Size"])
	assert.Equal(t, []int{2, 1, 0}, fieldIndexMap["Size"], "Size should have path [2,1,0]")
}

func TestVisit_FieldIndexPath_CanAccessValues(t *testing.T) {
	// Test that index paths can be used to access actual field values
	testStruct := VisitorTestEmbedded{
		VisitorTestBase: VisitorTestBase{BaseValue: "base_value"},
		EmbeddedValue:   42,
		Services: &VisitorTestServices{
			Logger: VisitorTestLogger{Level: "debug"},
			Cache:  &VisitorTestCache{Size: 1024},
		},
	}

	type fieldInfo struct {
		index []int
		value reflect.Value
	}

	fieldMap := make(map[string]fieldInfo)

	visitor := Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
			fieldMap[field.Name] = fieldInfo{
				index: field.Index,
				value: fieldValue,
			}

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor)

	// Verify BaseValue
	info := fieldMap["BaseValue"]
	assert.Equal(t, []int{0, 0}, info.index)
	assert.Equal(t, "base_value", info.value.String())
	// Access via index path should match
	actualValue := reflect.ValueOf(testStruct).FieldByIndex(info.index)
	assert.Equal(t, "base_value", actualValue.String())

	// Verify EmbeddedValue
	info = fieldMap["EmbeddedValue"]
	assert.Equal(t, []int{1}, info.index)
	assert.Equal(t, int64(42), info.value.Int())
	actualValue = reflect.ValueOf(testStruct).FieldByIndex(info.index)
	assert.Equal(t, int64(42), actualValue.Int())

	// Verify Level (deeply nested)
	info = fieldMap["Level"]
	assert.Equal(t, []int{2, 0, 0}, info.index)
	assert.Equal(t, "debug", info.value.String())
	actualValue = reflect.ValueOf(testStruct).FieldByIndex(info.index)
	assert.Equal(t, "debug", actualValue.String())

	// Verify Size (through pointer)
	info = fieldMap["Size"]
	assert.Equal(t, []int{2, 1, 0}, info.index)
	assert.Equal(t, int64(1024), info.value.Int())
	actualValue = reflect.ValueOf(testStruct).FieldByIndex(info.index)
	assert.Equal(t, int64(1024), actualValue.Int())
}

func TestVisitType_FieldIndexPath_AllTraversalModes(t *testing.T) {
	// Test that all four traversal modes produce correct index paths
	testCases := []struct {
		name      string
		mode      TraversalMode
		useValue  bool
		fieldName string
		expected  []int
	}{
		{"TypeDepthFirst - BaseValue", DepthFirst, false, "BaseValue", []int{0, 0}},
		{"TypeDepthFirst - EmbeddedValue", DepthFirst, false, "EmbeddedValue", []int{1}},
		{"TypeDepthFirst - Level", DepthFirst, false, "Level", []int{2, 0, 0}},
		{"TypeBreadthFirst - BaseValue", BreadthFirst, false, "BaseValue", []int{0, 0}},
		{"TypeBreadthFirst - EmbeddedValue", BreadthFirst, false, "EmbeddedValue", []int{1}},
		{"TypeBreadthFirst - Level", BreadthFirst, false, "Level", []int{2, 0, 0}},
		{"ValueDepthFirst - BaseValue", DepthFirst, true, "BaseValue", []int{0, 0}},
		{"ValueDepthFirst - EmbeddedValue", DepthFirst, true, "EmbeddedValue", []int{1}},
		{"ValueDepthFirst - Level", DepthFirst, true, "Level", []int{2, 0, 0}},
		{"ValueBreadthFirst - BaseValue", BreadthFirst, true, "BaseValue", []int{0, 0}},
		{"ValueBreadthFirst - EmbeddedValue", BreadthFirst, true, "EmbeddedValue", []int{1}},
		{"ValueBreadthFirst - Level", BreadthFirst, true, "Level", []int{2, 0, 0}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var foundIndex []int

			if tc.useValue {
				testValue := VisitorTestEmbedded{
					VisitorTestBase: VisitorTestBase{BaseValue: "test"},
					EmbeddedValue:   42,
					Services: &VisitorTestServices{
						Logger: VisitorTestLogger{Level: "info"},
					},
				}

				visitor := Visitor{
					VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
						if field.Name == tc.fieldName {
							foundIndex = field.Index

							return Stop
						}

						return Continue
					},
				}

				Visit(reflect.ValueOf(testValue), visitor, WithTraversalMode(tc.mode))
			} else {
				visitor := TypeVisitor{
					VisitFieldType: func(field reflect.StructField, depth int) VisitAction {
						if field.Name == tc.fieldName {
							foundIndex = field.Index

							return Stop
						}

						return Continue
					},
				}

				VisitType(reflect.TypeOf(VisitorTestEmbedded{}), visitor, WithTraversalMode(tc.mode))
			}

			assert.Equal(t, tc.expected, foundIndex, "Field %s should have correct index path", tc.fieldName)
		})
	}
}

func TestVisit_FieldIndexPath_DeepNesting(t *testing.T) {
	// Test deeply nested structures (4+ levels)
	type Level4 struct {
		DeepValue string
	}

	type Level3 struct {
		Level4 `visit:"dive"`

		L3Value int
	}

	type Level2 struct {
		Level3 `visit:"dive"`

		L2Value bool
	}

	type Level1 struct {
		Level2

		L1Value float64
	}

	fieldIndexMap := make(map[string][]int)

	visitor := TypeVisitor{
		VisitFieldType: func(field reflect.StructField, depth int) VisitAction {
			fieldIndexMap[field.Name] = field.Index

			return Continue
		},
	}

	VisitType(reflect.TypeOf(Level1{}), visitor)

	// Verify deep nesting paths
	assert.Equal(t, []int{0, 0, 0, 0}, fieldIndexMap["DeepValue"], "DeepValue at level 4 should have 4-element path")
	assert.Equal(t, []int{0, 0, 1}, fieldIndexMap["L3Value"], "L3Value at level 3 should have 3-element path")
	assert.Equal(t, []int{0, 1}, fieldIndexMap["L2Value"], "L2Value at level 2 should have 2-element path")
	assert.Equal(t, []int{1}, fieldIndexMap["L1Value"], "L1Value at level 1 should have 1-element path")
}

func TestVisit_FieldIndexPath_MixedEmbedding(t *testing.T) {
	// Test mixed anonymous and tagged dive embedding
	type Inner struct {
		InnerField string
	}

	type Middle struct {
		Inner // Anonymous embedding - always recursed

		MiddleField int
		Tagged      Inner `visit:"dive"` // Non-anonymous with dive tag - only recursed with WithDiveTag
	}

	type Outer struct {
		Middle

		OuterField bool
	}

	fieldIndexMap := make(map[string][]int)

	// Test with dive tag enabled
	visitor := TypeVisitor{
		VisitFieldType: func(field reflect.StructField, depth int) VisitAction {
			fieldIndexMap[field.Name] = field.Index

			return Continue
		},
	}

	VisitType(reflect.TypeOf(Outer{}), visitor, WithDiveTag("visit", "dive"))

	// Anonymous InnerField should be at [0, 0, 0]
	anonymousInnerIndex, found := fieldIndexMap["InnerField"]
	assert.True(t, found, "Should find InnerField from anonymous embedding")
	assert.Equal(t, []int{0, 0, 0}, anonymousInnerIndex, "Anonymous InnerField should have path [0,0,0]")

	// MiddleField should be at [0, 1]
	assert.NotNil(t, fieldIndexMap["MiddleField"])
	assert.Equal(t, []int{0, 1}, fieldIndexMap["MiddleField"], "MiddleField should have path [0,1]")

	// OuterField should be at [1]
	assert.NotNil(t, fieldIndexMap["OuterField"])
	assert.Equal(t, []int{1}, fieldIndexMap["OuterField"], "OuterField should have path [1]")
}

func TestVisit_FieldIndexPath_PointerFields(t *testing.T) {
	// Test that index paths work correctly with pointer fields
	testStruct := VisitorTestEmbedded{
		Services: &VisitorTestServices{
			Cache: &VisitorTestCache{Size: 512},
		},
	}

	var (
		cacheFieldIndex []int
		sizeFieldIndex  []int
	)

	visitor := Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
			switch field.Name {
			case "Cache":
				cacheFieldIndex = field.Index
			case "Size":
				sizeFieldIndex = field.Index
			}

			return Continue
		},
	}

	Visit(reflect.ValueOf(testStruct), visitor)

	// Verify Cache pointer field index
	assert.Equal(t, []int{2, 1}, cacheFieldIndex, "Cache pointer field should have path [2,1]")

	// Verify Size field inside pointer
	assert.Equal(t, []int{2, 1, 0}, sizeFieldIndex, "Size inside pointer should have path [2,1,0]")

	// Verify we can access the value through the index path
	cacheValue := reflect.ValueOf(testStruct).FieldByIndex(cacheFieldIndex)
	assert.Equal(t, reflect.Ptr, cacheValue.Kind())
	assert.False(t, cacheValue.IsNil())

	sizeValue := reflect.ValueOf(testStruct).FieldByIndex(sizeFieldIndex)
	assert.Equal(t, int64(512), sizeValue.Int())
}

func TestVisitType_FieldIndexPath_Consistency(t *testing.T) {
	// Test that Type traversal and Value traversal produce the same index paths for non-nil fields
	testValue := VisitorTestNested{
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

	typeFieldIndices := make(map[string][]int)
	valueFieldIndices := make(map[string][]int)

	// Collect indices from Type traversal
	typeVisitor := TypeVisitor{
		VisitFieldType: func(field reflect.StructField, depth int) VisitAction {
			typeFieldIndices[field.Name] = field.Index

			return Continue
		},
	}
	VisitType(reflect.TypeOf(testValue), typeVisitor)

	// Collect indices from Value traversal
	valueVisitor := Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction {
			valueFieldIndices[field.Name] = field.Index

			return Continue
		},
	}
	Visit(reflect.ValueOf(testValue), valueVisitor)

	// Verify that all fields have the same indices in both traversals
	for fieldName, typeIndex := range typeFieldIndices {
		valueIndex, found := valueFieldIndices[fieldName]
		assert.True(t, found, "Field %s should be found in value traversal", fieldName)
		assert.Equal(t, typeIndex, valueIndex, "Field %s should have same index in both traversals", fieldName)
	}
}
