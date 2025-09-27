package search

import (
	"reflect"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/reflectx"
	"github.com/ilxqx/vef-framework-go/strhelpers"
	"github.com/samber/lo"
)

// New creates a Search instance by parsing struct fields with search tags from the given reflect.Type.
// Returns an empty Search if the type is not a struct.
func New(typ reflect.Type) Search {
	typ = reflectx.Indirect(typ)
	if typ.Kind() != reflect.Struct {
		logger.Warnf("Invalid value type, expected struct, got %s", typ.Name())
		return Search{}
	}

	return Search{conditions: parseStruct(typ)}
}

// NewFor creates a Search instance by parsing struct fields with search tags from type T.
// This is a generic convenience function that calls New with reflect.TypeFor[T]().
func NewFor[T any]() Search {
	return New(reflect.TypeFor[T]())
}

// parseStruct parses the search conditions from a struct using visitor pattern.
func parseStruct(t reflect.Type) []Condition {
	conditions := make([]Condition, 0)

	visitor := reflectx.TypeVisitor{
		VisitFieldType: func(field reflect.StructField, depth int) reflectx.VisitAction {
			if tag, ok := field.Tag.Lookup(TagSearch); ok {
				// Skip dive fields - visitor will handle recursion automatically
				if tag == AttrDive {
					return reflectx.Continue
				}

				attrs := strhelpers.ParseTagAttrs(tag)
				// Handle regular search fields
				conditions = append(conditions, buildCondition(field, attrs))
			}

			return reflectx.SkipChildren
		},
	}

	reflectx.VisitType(
		t, visitor,
		reflectx.WithDiveTag(TagSearch, AttrDive),
		reflectx.WithTraversalMode(reflectx.DepthFirst),
	)

	return conditions
}

// buildCondition builds a condition from a struct field and attributes.
func buildCondition(field reflect.StructField, attrs map[string]string) Condition {
	var (
		column  = attrs[AttrColumn]
		columns []string
	)
	if column == constants.Empty {
		columns = []string{lo.SnakeCase(field.Name)}
	} else {
		columns = strings.Split(column, constants.Pipe)
	}

	operator := attrs[AttrOperator]
	if operator == constants.Empty {
		if defaultOp := attrs[AttrDefault]; defaultOp != constants.Empty {
			operator = defaultOp
		} else {
			operator = string(Equals)
		}
	}

	return Condition{
		Index:    field.Index,
		Alias:    attrs[AttrAlias],
		Columns:  columns,
		Operator: Operator(operator),
		Params: lo.TernaryF(
			attrs[AttrParams] == constants.Empty,
			func() map[string]string {
				return make(map[string]string)
			},
			func() map[string]string {
				return strhelpers.ParseTagArgs(attrs[AttrParams])
			},
		),
	}
}
