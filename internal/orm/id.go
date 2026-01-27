package orm

import (
	"reflect"

	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/id"
)

// IDHandler implements InsertHandler for automatically generating unique primary key IDs.
// It uses Snowflake algorithm to generate distributed unique IDs in Base36 format.
type IDHandler struct{}

// OnInsert automatically generates a unique ID for string primary key fields that are zero-valued.
// It only applies to primary key fields of string type that haven't been explicitly set.
func (*IDHandler) OnInsert(_ *BunInsertQuery, _ *schema.Table, field *schema.Field, _ any, value reflect.Value) {
	if field.IsPK && field.IndirectType.Kind() == reflect.String && value.IsZero() {
		value.SetString(id.Generate())
	}
}

// Name returns the column name for the ID field.
func (*IDHandler) Name() string {
	return constants.ColumnID
}
