package orm

import (
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/uptrace/bun/schema"
)

// ClauseConditionBuilder is responsible for collecting and grouping condition clauses, and rendering them.
type ClauseConditionBuilder struct {
	*CriteriaBuilder

	conditions []schema.QueryWithSep
}

// newConditionBuilder creates a new ClauseConditionBuilder.
func newConditionBuilder(qb QueryBuilder) *ClauseConditionBuilder {
	cb := &CriteriaBuilder{
		qb: qb,
		eb: qb.ExprBuilder(),
		and: func(string, ...any) {
		},
		or: func(string, ...any) {
		},
		group: func(string, func(ConditionBuilder)) {
		},
	}

	builder := &ClauseConditionBuilder{CriteriaBuilder: cb}
	cb.and = builder.And
	cb.or = builder.Or
	cb.group = builder.BuildGroup
	return builder
}

func (cb *ClauseConditionBuilder) AppendConditions(conditions ...schema.QueryWithSep) {
	cb.conditions = append(cb.conditions, conditions...)
}

func (cb *ClauseConditionBuilder) And(query string, args ...any) {
	cb.AppendConditions(schema.SafeQueryWithSep(query, args, separatorAnd))
}

func (cb *ClauseConditionBuilder) Or(query string, args ...any) {
	cb.AppendConditions(schema.SafeQueryWithSep(query, args, separatorOr))
}

func (cb *ClauseConditionBuilder) BuildGroup(sep string, builder func(ConditionBuilder)) {
	saved := cb.conditions
	cb.conditions = nil

	builder(cb)

	on := cb.conditions
	cb.conditions = saved
	cb.AppendGroup(sep, on)
}

func (cb *ClauseConditionBuilder) AppendGroup(sep string, conditions []schema.QueryWithSep) {
	if len(conditions) == 0 {
		return
	}

	cb.AppendConditions(schema.SafeQueryWithSep(constants.Empty, nil, sep))
	cb.AppendConditions(schema.SafeQueryWithSep(constants.Empty, nil, constants.LeftParenthesis))
	conditions[0].Sep = constants.Empty
	cb.AppendConditions(conditions...)
	cb.AppendConditions(schema.SafeQueryWithSep(constants.Empty, nil, constants.RightParenthesis))
}

func (cb *ClauseConditionBuilder) AppendQuery(fmter schema.Formatter, b []byte) (_ []byte, err error) {
	if len(cb.conditions) == 0 {
		return b, nil
	}

	for i, condition := range cb.conditions {
		if i > 0 {
			b = append(b, condition.Sep...)
		}

		b, err = condition.AppendQuery(fmter, b)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}
