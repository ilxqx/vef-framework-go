package orm

import (
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

type filterClause struct {
	condition schema.QueryAppender
}

func (f *filterClause) AppendQuery(gen schema.QueryGen, b []byte) (_ []byte, err error) {
	b = append(b, " FILTER (WHERE "...)
	if b, err = f.condition.AppendQuery(gen, b); err != nil {
		return
	}

	b = append(b, constants.ByteRightParenthesis)

	return b, nil
}

func newFilterClause(condition schema.QueryAppender) *filterClause {
	return &filterClause{condition: condition}
}
