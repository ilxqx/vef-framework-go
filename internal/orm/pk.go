package orm

import (
	"fmt"
	"reflect"

	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

// PkField describes a model's primary key field with common aliases.
// It provides helpers to get/set the Pk value on a concrete model instance.
type PkField struct {
	// Field is the Go struct field name, e.g. "UserId".
	Field string
	// Column is the database column name as defined in schema, e.g. "user_id".
	Column string
	// Name is the lower camel-case alias, usually used in params or Api payloads, e.g. "userId".
	Name string

	f *schema.Field
}

// Value returns the primary key value from the given model instance.
// The model must be a pointer to struct; otherwise an error is returned.
// It leverages bun's schema.Field to read the concrete field value safely.
func (p *PkField) Value(model any) (any, error) {
	value, err := p.validateModel(model)
	if err != nil {
		return nil, err
	}

	return p.f.Value(value).Interface(), nil
}

// Set writes the provided value into the model's primary key field.
// The model must be a pointer to struct. Basic kinds supported:
// - string (and *string)
// - int/int32/int64 (and their pointer forms)
// For unsupported kinds, an error is returned.
func (p *PkField) Set(model, value any) error {
	modelValue, err := p.validateModel(model)
	if err != nil {
		return err
	}

	pkValue := p.f.Value(modelValue)
	switch kind := p.f.IndirectType.Kind(); kind {
	case reflect.String:
		v, err := cast.ToStringE(value)
		if err != nil {
			return err
		}

		if p.f.IsPtr {
			pkValue.Set(reflect.ValueOf(&v))
		} else {
			pkValue.SetString(v)
		}

	case reflect.Int, reflect.Int32, reflect.Int64:
		v, err := cast.ToInt64E(value)
		if err != nil {
			return err
		}

		if p.f.IsPtr {
			pkValue.Set(reflect.ValueOf(&v))
		} else {
			pkValue.SetInt(v)
		}

	default:
		return fmt.Errorf("%w: %s", ErrPrimaryKeyUnsupportedType, kind)
	}

	return nil
}

func (p *PkField) validateModel(model any) (reflect.Value, error) {
	if value, ok := model.(reflect.Value); ok {
		if value.Kind() == reflect.Pointer {
			value = value.Elem()
			if value.Kind() != reflect.Struct {
				return reflect.Value{}, ErrModelMustBePointerToStruct
			}
		} else {
			if value.Kind() != reflect.Struct || !value.CanAddr() {
				return reflect.Value{}, ErrModelMustBePointerToStruct
			}
		}

		return value, nil
	}

	value := reflect.ValueOf(model)
	if value.Kind() != reflect.Pointer {
		return reflect.Value{}, ErrModelMustBePointerToStruct
	}

	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return reflect.Value{}, ErrModelMustBePointerToStruct
	}

	return value, nil
}

// NewPkField constructs a PkField helper from a bun schema.Field.
// Field is the Go struct field name; Column is the DB column name;
// Name is a lower-camel alias commonly used in params or Api payloads.
func NewPkField(field *schema.Field) *PkField {
	return &PkField{
		Field:  field.GoName,
		Column: field.Name,
		Name:   lo.CamelCase(field.Name),
		f:      field,
	}
}

// parsePkColumnsAndValues parses the primary key columns and values from the given table and primary key value.
func parsePkColumnsAndValues(method string, table *schema.Table, pk any, alias ...string) (*pkColumns, *pkValues) {
	if table == nil {
		panic(fmt.Sprintf(
			"method %s failed: table schema is nil. "+
				"This usually happens when: "+
				"1) Model() method was not called before %s, "+
				"2) Model() was called with plain nil or a value that is not a struct pointer (or slice pointer of struct), "+
				"3) Table()/TableExpr() was used without binding a model via Model(). "+
				"Please ensure you call Model() with a valid struct pointer (or slice pointer of struct) before using %s.",
			method, method, method,
		))
	}

	pks := table.PKs
	if len(pks) == 0 {
		panic(
			fmt.Sprintf("table %s has no primary key", table.Name),
		)
	}

	aliasToUse := table.SQLAlias
	if len(alias) > 0 {
		aliasToUse = bun.Safe(alias[0])
	}

	columns := make([]bun.Safe, 0, len(pks))
	for _, p := range pks {
		columns = append(columns, p.SQLName)
	}

	var values []any

	pkv := reflect.ValueOf(pk)
	if pkv.Kind() == reflect.Slice {
		values = make([]any, 0, pkv.Len())
		for i := range pkv.Len() {
			values = append(values, pkv.Index(i).Interface())
		}
	} else {
		values = []any{pk}
	}

	return &pkColumns{alias: aliasToUse, columns: columns}, &pkValues{values: values}
}

type pkColumns struct {
	alias   bun.Safe
	columns []bun.Safe
}

func (p *pkColumns) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	cLen := len(p.columns)
	if cLen == 0 {
		return dialect.AppendNull(b), nil
	}

	if cLen > 1 {
		b = append(b, constants.ByteLeftParenthesis)
	}

	for i, column := range p.columns {
		if i > 0 {
			b = append(b, constants.CommaSpace...)
		}

		if b, err = p.alias.AppendQuery(fmter, b); err != nil {
			return
		}

		b = append(b, constants.ByteDot)

		if b, err = column.AppendQuery(fmter, b); err != nil {
			return
		}
	}

	if cLen > 1 {
		b = append(b, constants.ByteRightParenthesis)
	}

	return b, nil
}

type pkValues struct {
	values []any
}

func (p *pkValues) AppendQuery(formatter schema.Formatter, b []byte) ([]byte, error) {
	vLen := len(p.values)
	if vLen == 0 {
		return dialect.AppendNull(b), nil
	}

	if vLen == 1 {
		b = formatter.AppendValue(b, reflect.ValueOf(p.values[0]))

		return b, nil
	}

	return bun.In(p.values).AppendQuery(formatter, b)
}
