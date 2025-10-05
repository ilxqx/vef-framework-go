package tabular

import (
	"reflect"

	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/uptrace/bun"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/reflectx"
	"github.com/ilxqx/vef-framework-go/strhelpers"
)

var baseModelType = reflect.TypeFor[bun.BaseModel]()

// parseStruct parses the tabular columns from a struct using visitor pattern.
func parseStruct(t reflect.Type) []*Column {
	t = reflectx.Indirect(t)
	if t.Kind() != reflect.Struct {
		logger.Warnf("Invalid value type, expected struct, got %s", t.Name())

		return nil
	}

	columns := make([]*Column, 0)
	columnOrder := 0

	visitor := reflectx.TypeVisitor{
		VisitFieldType: func(field reflect.StructField, depth int) reflectx.VisitAction {
			if field.Anonymous && field.Type == baseModelType {
				return reflectx.SkipChildren
			}

			tag, hasTag := field.Tag.Lookup(TagTabular)

			// If has tag, parse it
			if hasTag {
				// Skip ignored fields
				if tag == IgnoreField {
					return reflectx.SkipChildren
				}

				// Skip dive fields - visitor will handle recursion automatically
				if tag == AttrDive {
					return reflectx.Continue
				}

				// Parse tag attributes
				attrs := strhelpers.ParseTagAttrs(tag)

				// Build column from attributes
				column := buildColumn(field, attrs, columnOrder)
				columns = append(columns, column)
				columnOrder++
			} else {
				// No tag: use default configuration
				column := buildColumn(field, make(map[string]string), columnOrder)
				columns = append(columns, column)
				columnOrder++
			}

			return reflectx.SkipChildren
		},
	}

	reflectx.VisitType(
		t, visitor,
		reflectx.WithDiveTag(TagTabular, AttrDive),
		reflectx.WithTraversalMode(reflectx.DepthFirst),
	)

	return columns
}

// buildColumn builds a Column from a struct field and attributes.
func buildColumn(field reflect.StructField, attrs map[string]string, autoOrder int) *Column {
	// Get column name - support default value (name=用户ID or just 用户ID)
	name := attrs[AttrName]
	if name == constants.Empty {
		// Use default value from ParseTagAttrs (when tag is just "用户ID" without key)
		name = lo.CoalesceOrEmpty(attrs[strhelpers.TagAttrDefaultKey], field.Name)
	}

	// Parse width
	var width float64
	if widthStr := attrs[AttrWidth]; widthStr != constants.Empty {
		width = cast.ToFloat64(widthStr)
	}

	// Parse order - if not specified, use auto-incrementing order
	order := autoOrder
	if orderStr := attrs[AttrOrder]; orderStr != constants.Empty {
		order = cast.ToInt(orderStr)
	}

	return &Column{
		Index:     field.Index,
		Name:      name,
		Width:     width,
		Order:     order,
		Default:   attrs[AttrDefault],
		Format:    attrs[AttrFormat],
		Formatter: attrs[AttrFormatter],
		Parser:    attrs[AttrParser],
	}
}
