package orm

import (
	"fmt"
	"reflect"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/schema"
)

// parsePKColumnsAndValues parses the primary key columns and values from the given table and primary key value.
func parsePKColumnsAndValues(method string, table *schema.Table, pk any, alias ...string) (*pkColumns, *pkValues) {
	if table == nil {
		panic(
			fmt.Sprintf("method %s must be called after Model method", method),
		)
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
	alias   bun.Safe   // alias is the table alias for the primary key columns
	columns []bun.Safe // columns contains the primary key column names
}

func (p *pkColumns) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	cLen := len(p.columns)
	if cLen == 0 {
		return dialect.AppendNull(b), nil
	}

	if cLen > 1 {
		b = append(b, '(')
	}

	for i, column := range p.columns {
		if i > 0 {
			b = append(b, constants.CommaSpace...)
		}

		b, err = p.alias.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}

		b = append(b, '.')

		b, err = column.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	if cLen > 1 {
		b = append(b, ')')
	}

	return b, nil
}

type pkValues struct {
	values []any // values contains the primary key values
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
