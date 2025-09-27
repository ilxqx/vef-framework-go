package orm

import (
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/uptrace/bun/schema"
)

type filterClause struct {
	condition schema.QueryAppender
}

func (f *filterClause) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	b = append(b, " FILTER (WHERE "...)
	if b, err = f.condition.AppendQuery(fmter, b); err != nil {
		return
	}
	b = append(b, constants.ByteRightParenthesis)

	return b, nil
}

func newFilterClause(condition schema.QueryAppender) *filterClause {
	return &filterClause{condition: condition}
}
