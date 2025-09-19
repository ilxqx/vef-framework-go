package orm

import (
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// baseConditionBuilder is a builder for building base conditions.
type baseConditionBuilder struct {
	*richConditionBuilder
	condition []schema.QueryWithSep
}

// newCommonConditionBuilder creates a new baseConditionBuilder.
func newCommonConditionBuilder(table *orm.Table, subQueryBuilder func(builder func(query orm.Query)) *bun.SelectQuery) *baseConditionBuilder {
	richCb := &richConditionBuilder{
		table:           table,
		subQueryBuilder: subQueryBuilder,
		and: func(query string, args ...any) {
		},
		or: func(query string, args ...any) {
		},
		group: func(sep string, builder func(orm.ConditionBuilder)) {
		},
	}

	builder := &baseConditionBuilder{richConditionBuilder: richCb}

	richCb.and = builder.and
	richCb.or = builder.or
	richCb.group = builder.group

	return builder
}

func (cb *baseConditionBuilder) addCondition(on ...schema.QueryWithSep) {
	cb.condition = append(cb.condition, on...)
}

func (cb *baseConditionBuilder) and(query string, args ...any) {
	cb.addCondition(schema.SafeQueryWithSep(query, args, orm.SeparatorAnd))
}

func (cb *baseConditionBuilder) or(query string, args ...any) {
	cb.addCondition(schema.SafeQueryWithSep(query, args, orm.SeparatorOr))
}

func (cb *baseConditionBuilder) group(sep string, builder func(orm.ConditionBuilder)) {
	saved := cb.condition
	cb.condition = nil

	builder(cb)

	on := cb.condition
	cb.condition = saved
	cb.addGroup(sep, on)
}

func (cb *baseConditionBuilder) addGroup(sep string, on []schema.QueryWithSep) {
	if len(on) == 0 {
		return
	}

	cb.addCondition(schema.SafeQueryWithSep(constants.Empty, nil, sep))
	cb.addCondition(schema.SafeQueryWithSep(constants.Empty, nil, constants.LeftParenthesis))
	on[0].Sep = constants.Empty
	cb.addCondition(on...)
	cb.addCondition(schema.SafeQueryWithSep(constants.Empty, nil, constants.LeftParenthesis))
}

func (cb *baseConditionBuilder) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	if len(cb.condition) == 0 {
		return b, nil
	}

	for i, on := range cb.condition {
		if i > 0 {
			b = append(b, on.Sep...)
		}

		b, err = on.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}
