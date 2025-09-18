package search

import (
	"reflect"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/reflectx"
	"github.com/ilxqx/vef-framework-go/utils"
	"github.com/samber/lo"
)

// NewFromType creates a new search from a type.
func NewFromType(t reflect.Type) Search {
	t = reflectx.Indirect(t)
	if t.Kind() != reflect.Struct {
		logger.Warnf("Invalid value type, expected struct, got %s", t.Name())
		return Search{}
	}

	return Search{conditions: parseStruct(t)}
}

// New creates a new search from a struct.
func New[T any]() Search {
	return NewFromType(reflect.TypeFor[T]())
}

// parseStruct parses the search conditions from a struct.
func parseStruct(t reflect.Type) []Condition {
	conditions := make([]Condition, 0)
	for i := range t.NumField() {
		field := t.Field(i)
		fieldType := reflectx.Indirect(field.Type)

		if tag, ok := field.Tag.Lookup(TagSearch); ok {
			attrs := utils.ParseTagAttrs(tag)
			// Handle dive field.
			if _, ok := attrs[AttrDive]; ok {
				if fieldType.Kind() == reflect.Struct {
					conditions = append(conditions, parseStruct(fieldType)...)
				} else {
					logger.Warnf("Invalid dive field type, expected struct, got %s", fieldType.Name())
				}
				continue
			}

			// Handle filter field.
			conditions = append(conditions, buildCondition(field, attrs))
		}
	}

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

	operator := lo.ValueOr(attrs, AttrOperator, lo.ValueOr(attrs, AttrDefault, string(Equals)))

	return Condition{
		Index:    field.Index,
		Alias:    attrs[AttrAlias],
		Columns:  columns,
		Operator: Operator(operator),
		Args: lo.TernaryF(
			attrs[AttrArgs] == constants.Empty,
			func() map[string]string {
				return make(map[string]string)
			},
			func() map[string]string {
				return utils.ParseQueryString(attrs[AttrArgs])
			},
		),
	}
}
