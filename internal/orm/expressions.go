package orm

import (
	"reflect"

	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

var (
	bytesType         = reflect.TypeFor[[]byte]()
	queryAppenderType = reflect.TypeFor[schema.QueryAppender]()
)

type Expressions struct {
	exprs []any
	sep   any
}

func (e *Expressions) AppendQuery(gen schema.QueryGen, b []byte) ([]byte, error) {
	var appendExprs func(b []byte, slice reflect.Value) ([]byte, error)

	appendExprs = func(b []byte, slice reflect.Value) (_ []byte, err error) {
		sliceLen := slice.Len()
		if sliceLen == 0 {
			return dialect.AppendNull(b), nil
		}

		for i := range sliceLen {
			if i > 0 {
				// Handle separator
				switch sep := e.sep.(type) {
				case string:
					b = append(b, sep...)
				case schema.QueryAppender:
					if b, err = sep.AppendQuery(gen, b); err != nil {
						return
					}
				default:
					b = gen.AppendValue(b, reflect.ValueOf(sep))
				}
			}

			expr := slice.Index(i)
			if expr.Type().Implements(queryAppenderType) {
				appender := expr.Interface().(schema.QueryAppender)
				if b, err = appender.AppendQuery(gen, b); err != nil {
					return
				}
			}

			if expr.Kind() == reflect.Slice && expr.Type() != bytesType {
				b = append(b, constants.ByteLeftParenthesis)
				if b, err = appendExprs(b, expr); err != nil {
					return
				}

				b = append(b, constants.ByteRightParenthesis)
			} else {
				b = gen.AppendValue(b, expr)
			}
		}

		return b, nil
	}

	return appendExprs(b, reflect.ValueOf(e.exprs))
}

func newExpressions(sep any, exprs ...any) *Expressions {
	return &Expressions{
		exprs: exprs,
		sep:   sep,
	}
}
