package reflectx

import (
	"container/list"
	"reflect"

	"github.com/ilxqx/vef-framework-go/constants"
)

// TraversalMode defines the traversal strategy for visiting struct fields and methods.
type TraversalMode int

const (
	// DepthFirst traverses deeply into each branch before moving to siblings.
	DepthFirst TraversalMode = iota
	// BreadthFirst traverses all siblings at the current level before going deeper.
	BreadthFirst
)

// VisitAction represents the action to take after visiting a node.
type VisitAction int

const (
	// Continue with normal traversal.
	Continue VisitAction = iota
	// Stop traversal immediately.
	Stop
	// SkipChildren skips traversing into child nodes of current node.
	SkipChildren
)

// TagConfig configures which tagged fields should be recursively traversed.
type TagConfig struct {
	// Name is the tag key (e.g., "visit")
	Name string
	// Value is the tag value that triggers recursion (e.g., "dive")
	Value string
}

// VisitorConfig configures how the traversal should be performed.
type VisitorConfig struct {
	// TraversalMode specifies depth-first or breadth-first traversal
	TraversalMode TraversalMode
	// Recursive enables traversal into embedded structs and tagged fields
	Recursive bool
	// DiveTag configures which non-anonymous fields should be recursively traversed
	// Default is TagConfig{Name: "visit", Value: "dive"}
	DiveTag TagConfig
	// MaxDepth limits the maximum recursion depth (0 means no limit)
	MaxDepth int
}

// VisitorOption configures visitor behavior.
type VisitorOption func(*VisitorConfig)

// WithTraversalMode sets the traversal mode (DepthFirst or BreadthFirst).
func WithTraversalMode(mode TraversalMode) VisitorOption {
	return func(c *VisitorConfig) {
		c.TraversalMode = mode
	}
}

// WithDisableRecursive disables recursive traversal into embedded structs and tagged fields.
// By default, recursive traversal is enabled.
func WithDisableRecursive() VisitorOption {
	return func(c *VisitorConfig) {
		c.Recursive = false
	}
}

// WithDiveTag configures which non-anonymous fields should be recursively traversed.
func WithDiveTag(tagName, tagValue string) VisitorOption {
	return func(c *VisitorConfig) {
		c.DiveTag = TagConfig{Name: tagName, Value: tagValue}
	}
}

// WithMaxDepth limits the maximum recursion depth (0 means no limit).
func WithMaxDepth(maxDepth int) VisitorOption {
	return func(c *VisitorConfig) {
		c.MaxDepth = maxDepth
	}
}

// defaultVisitorConfig returns the default configuration for struct traversal.
func defaultVisitorConfig() VisitorConfig {
	return VisitorConfig{
		TraversalMode: DepthFirst,
		Recursive:     true,
		DiveTag:       TagConfig{Name: "visit", Value: "dive"},
		MaxDepth:      0,
	}
}

// Visitor defines callback functions for different types of nodes during traversal.
type Visitor struct {
	// VisitStruct is called when entering a struct type
	VisitStruct StructVisitor
	// VisitField is called for each struct field
	VisitField FieldVisitor
	// VisitMethod is called for each method (if not nil)
	VisitMethod MethodVisitor
}

// StructVisitor is called when entering a struct during traversal.
// Parameters: structType, structValue, depth in traversal tree.
type StructVisitor func(structType reflect.Type, structValue reflect.Value, depth int) VisitAction

// FieldVisitor is called for each field during traversal.
// Parameters: field metadata, field value, depth in traversal tree.
type FieldVisitor func(field reflect.StructField, fieldValue reflect.Value, depth int) VisitAction

// MethodVisitor is called for each method during traversal.
// Parameters: method metadata, bound method value (ready to call), depth in traversal tree.
type MethodVisitor func(method reflect.Method, methodValue reflect.Value, depth int) VisitAction

// TypeVisitor defines callback functions for type-only traversal (without values).
// This is useful for static analysis, pre-initialization, and type introspection.
type TypeVisitor struct {
	// VisitStructType is called when entering a struct type
	VisitStructType StructTypeVisitor
	// VisitFieldType is called for each struct field type
	VisitFieldType FieldTypeVisitor
	// VisitMethodType is called for each method type (if not nil)
	VisitMethodType MethodTypeVisitor
}

// StructTypeVisitor is called when entering a struct type during traversal.
// Parameters: structType, depth in traversal tree.
type StructTypeVisitor func(structType reflect.Type, depth int) VisitAction

// FieldTypeVisitor is called for each field type during traversal.
// Parameters: field metadata, depth in traversal tree.
type FieldTypeVisitor func(field reflect.StructField, depth int) VisitAction

// MethodTypeVisitor is called for each method type during traversal.
// Parameters: method metadata, receiver type, depth in traversal tree.
type MethodTypeVisitor func(method reflect.Method, receiverType reflect.Type, depth int) VisitAction

// VisitFor is a generic convenience function that visits a struct type T using type visitor callbacks.
// This eliminates the need to call reflect.TypeOf manually.
func VisitFor[T any](visitor TypeVisitor, opts ...VisitorOption) {
	VisitType(reflect.TypeFor[T](), visitor, opts...)
}

// VisitOf is a convenience function that visits a struct value using visitor callbacks.
// This eliminates the need to call reflect.ValueOf manually.
func VisitOf(value any, visitor Visitor, opts ...VisitorOption) {
	Visit(reflect.ValueOf(value), visitor, opts...)
}

// Visit traverses a struct using visitor callbacks with optional configuration.
// It supports both depth-first and breadth-first traversal with configurable recursion rules.
func Visit(target reflect.Value, visitor Visitor, opts ...VisitorOption) {
	config := defaultVisitorConfig()
	for _, opt := range opts {
		opt(&config)
	}
	// Ensure we have a valid target
	if !target.IsValid() {
		return
	}

	// Dereference pointers to get to the actual struct
	for target.Kind() == reflect.Pointer {
		if target.IsNil() {
			return
		}

		target = target.Elem()
	}

	// Only work with structs
	if target.Kind() != reflect.Struct {
		return
	}

	// Initialize traversal state
	visited := make(map[reflect.Type]bool)

	if config.TraversalMode == DepthFirst {
		visitDepthFirst(target, config, visitor, visited, 0)
	} else {
		visitBreadthFirst(target, config, visitor, visited)
	}
}

// VisitType traverses a struct type using type visitor callbacks with optional configuration.
// This is optimized for type-only analysis without creating value instances.
// It supports both depth-first and breadth-first traversal with configurable recursion rules.
func VisitType(targetType reflect.Type, visitor TypeVisitor, opts ...VisitorOption) {
	config := defaultVisitorConfig()
	for _, opt := range opts {
		opt(&config)
	}
	// Dereference pointer types to get to the actual struct type
	for targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	// Only work with structs
	if targetType.Kind() != reflect.Struct {
		return
	}

	// Initialize traversal state
	visited := make(map[reflect.Type]bool)

	if config.TraversalMode == DepthFirst {
		visitTypeDepthFirst(targetType, config, visitor, visited, 0)
	} else {
		visitTypeBreadthFirst(targetType, config, visitor, visited)
	}
}

// visitDepthFirst performs depth-first traversal of the struct.
func visitDepthFirst(target reflect.Value, config VisitorConfig, visitor Visitor, visited map[reflect.Type]bool, depth int) VisitAction {
	// Check max depth
	if config.MaxDepth > 0 && depth >= config.MaxDepth {
		return Continue
	}

	// Dereference pointers to get to the actual struct
	for target.Kind() == reflect.Pointer {
		if target.IsNil() {
			return Continue
		}

		target = target.Elem()
	}

	// Only work with structs
	if target.Kind() != reflect.Struct {
		return Continue
	}

	targetType := target.Type()

	// Prevent infinite recursion
	if visited[targetType] {
		return Continue
	}

	visited[targetType] = true

	// Visit the struct itself
	if visitor.VisitStruct != nil {
		if action := visitor.VisitStruct(targetType, target, depth); action != Continue {
			return action
		}
	}

	// Visit fields
	for i := 0; i < target.NumField(); i++ {
		field := target.Field(i)
		fieldType := targetType.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Visit the field
		if visitor.VisitField != nil {
			if action := visitor.VisitField(fieldType, field, depth); action == Stop {
				return Stop
			} else if action == SkipChildren {
				continue
			}
		}

		// Recursive traversal decision
		if config.Recursive && shouldRecurse(fieldType, config.DiveTag) {
			if action := visitDepthFirst(field, config, visitor, visited, depth+1); action == Stop {
				return Stop
			}
		}
	}

	// Visit methods if enabled
	if action := visitMethods(target, targetType, visitor.VisitMethod, depth); action == Stop {
		return Stop
	}

	return Continue
}

// visitBreadthFirst performs breadth-first traversal of the struct.
func visitBreadthFirst(target reflect.Value, config VisitorConfig, visitor Visitor, visited map[reflect.Type]bool) {
	type queueItem struct {
		value reflect.Value
		depth int
	}

	queue := list.New()
	queue.PushBack(queueItem{target, 0})

	for queue.Len() > 0 {
		item := queue.Remove(queue.Front()).(queueItem)
		current := item.value
		depth := item.depth

		// Check max depth
		if config.MaxDepth > 0 && depth >= config.MaxDepth {
			continue
		}

		// Dereference pointer if needed
		for current.Kind() == reflect.Pointer {
			if current.IsNil() {
				continue
			}

			current = current.Elem()
		}

		if current.Kind() != reflect.Struct {
			continue
		}

		currentType := current.Type()

		// Prevent infinite recursion
		if visited[currentType] {
			continue
		}

		visited[currentType] = true

		// Visit the struct
		if visitor.VisitStruct != nil {
			if action := visitor.VisitStruct(currentType, current, depth); action == Stop {
				return
			}
		}

		// Collect child nodes for next level
		var childNodes []queueItem

		// Visit fields and collect recursive candidates
		for i := 0; i < current.NumField(); i++ {
			field := current.Field(i)
			fieldType := currentType.Field(i)

			// Skip unexported fields
			if !field.CanInterface() {
				continue
			}

			// Visit the field
			skipChildren := false

			if visitor.VisitField != nil {
				if action := visitor.VisitField(fieldType, field, depth); action == Stop {
					return
				} else if action == SkipChildren {
					skipChildren = true
				}
			}

			// Add to next level if should recurse
			if !skipChildren && config.Recursive && shouldRecurse(fieldType, config.DiveTag) {
				childNodes = append(childNodes, queueItem{field, depth + 1})
			}
		}

		// Add child nodes to queue
		for _, child := range childNodes {
			queue.PushBack(child)
		}

		// Visit methods if enabled
		if action := visitMethods(current, currentType, visitor.VisitMethod, depth); action == Stop {
			return
		}
	}
}

// visitTypeDepthFirst performs depth-first traversal of struct types.
func visitTypeDepthFirst(targetType reflect.Type, config VisitorConfig, visitor TypeVisitor, visited map[reflect.Type]bool, depth int) VisitAction {
	// Check max depth
	if config.MaxDepth > 0 && depth >= config.MaxDepth {
		return Continue
	}

	// Prevent infinite recursion
	if visited[targetType] {
		return Continue
	}

	visited[targetType] = true

	// Visit the struct type itself
	if visitor.VisitStructType != nil {
		if action := visitor.VisitStructType(targetType, depth); action != Continue {
			return action
		}
	}

	// Visit field types
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Visit the field type
		if visitor.VisitFieldType != nil {
			if action := visitor.VisitFieldType(field, depth); action == Stop {
				return Stop
			} else if action == SkipChildren {
				continue
			}
		}

		// Recursive traversal decision
		if config.Recursive && shouldRecurse(field, config.DiveTag) {
			fieldType := Indirect(field.Type)
			if action := visitTypeDepthFirst(fieldType, config, visitor, visited, depth+1); action == Stop {
				return Stop
			}
		}
	}

	// Visit method types if enabled
	if action := visitMethodTypes(targetType, visitor.VisitMethodType, depth); action == Stop {
		return Stop
	}

	return Continue
}

// visitTypeBreadthFirst performs breadth-first traversal of struct types.
func visitTypeBreadthFirst(targetType reflect.Type, config VisitorConfig, visitor TypeVisitor, visited map[reflect.Type]bool) {
	type queueItem struct {
		structType reflect.Type
		depth      int
	}

	queue := list.New()
	queue.PushBack(queueItem{targetType, 0})

	for queue.Len() > 0 {
		item := queue.Remove(queue.Front()).(queueItem)
		current := Indirect(item.structType)
		depth := item.depth

		// Check max depth
		if config.MaxDepth > 0 && depth >= config.MaxDepth {
			continue
		}

		if current.Kind() != reflect.Struct {
			continue
		}

		// Prevent infinite recursion
		if visited[current] {
			continue
		}

		visited[current] = true

		// Visit the struct type
		if visitor.VisitStructType != nil {
			if action := visitor.VisitStructType(current, depth); action == Stop {
				return
			}
		}

		// Collect child types for next level
		var childTypes []queueItem

		// Visit field types and collect recursive candidates
		for i := 0; i < current.NumField(); i++ {
			field := current.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			// Visit the field type
			skipChildren := false

			if visitor.VisitFieldType != nil {
				if action := visitor.VisitFieldType(field, depth); action == Stop {
					return
				} else if action == SkipChildren {
					skipChildren = true
				}
			}

			// Add to next level if should recurse
			if !skipChildren && config.Recursive && shouldRecurse(field, config.DiveTag) {
				fieldType := field.Type
				childTypes = append(childTypes, queueItem{fieldType, depth + 1})
			}
		}

		// Add child types to queue
		for _, child := range childTypes {
			queue.PushBack(child)
		}

		// Visit method types if enabled
		if action := visitMethodTypes(current, visitor.VisitMethodType, depth); action == Stop {
			return
		}
	}
}

// visitMethods visits all methods on a struct value (including pointer receiver methods).
func visitMethods(target reflect.Value, targetType reflect.Type, visitor MethodVisitor, depth int) VisitAction {
	if visitor == nil {
		return Continue
	}

	// Use pointer type to access all methods (value + pointer receivers)
	var ptrTarget reflect.Value
	if target.CanAddr() {
		ptrTarget = target.Addr()
	} else {
		// Create a pointer to a copy if original is not addressable
		targetCopy := reflect.New(targetType)
		targetCopy.Elem().Set(target)
		ptrTarget = targetCopy
	}

	ptrType := ptrTarget.Type()
	for i := 0; i < ptrTarget.NumMethod(); i++ {
		method := ptrType.Method(i)
		methodValue := ptrTarget.Method(i)

		if action := visitor(method, methodValue, depth); action == Stop {
			return Stop
		}
	}

	return Continue
}

// visitMethodTypes visits all method types on a struct type (including pointer receiver methods).
func visitMethodTypes(targetType reflect.Type, visitor MethodTypeVisitor, depth int) VisitAction {
	if visitor == nil {
		return Continue
	}

	// Use pointer type to access all methods (value + pointer receivers)
	ptrType := reflect.PointerTo(targetType)
	for i := 0; i < ptrType.NumMethod(); i++ {
		method := ptrType.Method(i)
		if action := visitor(method, ptrType, depth); action == Stop {
			return Stop
		}
	}

	return Continue
}

// shouldRecurse determines if a field should be recursively traversed.
func shouldRecurse(field reflect.StructField, diveTag TagConfig) bool {
	// Always recurse into anonymous embedded fields
	if field.Anonymous {
		return Indirect(field.Type).Kind() == reflect.Struct
	}

	// Check for dive tag on non-anonymous fields
	if diveTag.Name != constants.Empty && diveTag.Value != constants.Empty && field.Tag.Get(diveTag.Name) == diveTag.Value {
		return Indirect(field.Type).Kind() == reflect.Struct
	}

	return false
}
