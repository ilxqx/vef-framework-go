package orm

import (
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

type CriteriaBuilder struct {
	qb    QueryBuilder
	eb    ExprBuilder
	and   func(query string, args ...any)
	or    func(query string, args ...any)
	group func(sep string, builder func(ConditionBuilder))
}

func (cb *CriteriaBuilder) Apply(fns ...ApplyFunc[ConditionBuilder]) ConditionBuilder {
	for _, fn := range fns {
		if fn != nil {
			fn(cb)
		}
	}

	return cb
}

func (cb *CriteriaBuilder) ApplyIf(condition bool, fns ...ApplyFunc[ConditionBuilder]) ConditionBuilder {
	if condition {
		return cb.Apply(fns...)
	}

	return cb
}

func (cb *CriteriaBuilder) Equals(column string, value any) ConditionBuilder {
	cb.and("? = ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) OrEquals(column string, value any) ConditionBuilder {
	cb.or("? = ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) EqualsColumn(column1, column2 string) ConditionBuilder {
	cb.and("? = ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) OrEqualsColumn(column1, column2 string) ConditionBuilder {
	cb.or("? = ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) EqualsSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? = (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrEqualsSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? = (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) EqualsAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? = ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrEqualsAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? = ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) EqualsAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? = ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrEqualsAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? = ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) EqualsExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("? = ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrEqualsExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("? = ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) NotEquals(column string, value any) ConditionBuilder {
	cb.and("? <> ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) OrNotEquals(column string, value any) ConditionBuilder {
	cb.or("? <> ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) NotEqualsColumn(column1, column2 string) ConditionBuilder {
	cb.and("? <> ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) OrNotEqualsColumn(column1, column2 string) ConditionBuilder {
	cb.or("? <> ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) NotEqualsSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? <> (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrNotEqualsSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? <> (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) NotEqualsAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? <> ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrNotEqualsAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? <> ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) NotEqualsAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? <> ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrNotEqualsAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? <> ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) NotEqualsExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("? <> ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrNotEqualsExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("? <> ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) GreaterThan(column string, value any) ConditionBuilder {
	cb.and("? > ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThan(column string, value any) ConditionBuilder {
	cb.or("? > ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) GreaterThanColumn(column1, column2 string) ConditionBuilder {
	cb.and("? > ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThanColumn(column1, column2 string) ConditionBuilder {
	cb.or("? > ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) GreaterThanSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? > (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThanSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? > (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) GreaterThanAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? > ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThanAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? > ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) GreaterThanAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? > ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThanAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? > ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) GreaterThanExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("? > ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThanExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("? > ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) GreaterThanOrEqual(column string, value any) ConditionBuilder {
	cb.and("? >= ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThanOrEqual(column string, value any) ConditionBuilder {
	cb.or("? >= ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) GreaterThanOrEqualColumn(column1, column2 string) ConditionBuilder {
	cb.and("? >= ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThanOrEqualColumn(column1, column2 string) ConditionBuilder {
	cb.or("? >= ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) GreaterThanOrEqualSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? >= (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThanOrEqualSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? >= (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) GreaterThanOrEqualAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? >= ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThanOrEqualAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? >= ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) GreaterThanOrEqualAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? >= ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThanOrEqualAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? >= ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) GreaterThanOrEqualExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("? >= ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrGreaterThanOrEqualExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("? >= ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) LessThan(column string, value any) ConditionBuilder {
	cb.and("? < ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) OrLessThan(column string, value any) ConditionBuilder {
	cb.or("? < ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) LessThanColumn(column1, column2 string) ConditionBuilder {
	cb.and("? < ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) OrLessThanColumn(column1, column2 string) ConditionBuilder {
	cb.or("? < ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) LessThanSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? < (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrLessThanSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? < (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) LessThanAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? < ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrLessThanAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? < ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) LessThanAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? < ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrLessThanAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? < ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) LessThanExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("? < ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrLessThanExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("? < ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) LessThanOrEqual(column string, value any) ConditionBuilder {
	cb.and("? <= ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) OrLessThanOrEqual(column string, value any) ConditionBuilder {
	cb.or("? <= ?", cb.eb.Column(column), value)

	return cb
}

func (cb *CriteriaBuilder) LessThanOrEqualColumn(column1, column2 string) ConditionBuilder {
	cb.and("? <= ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) OrLessThanOrEqualColumn(column1, column2 string) ConditionBuilder {
	cb.or("? <= ?", cb.eb.Column(column1), cb.eb.Column(column2))

	return cb
}

func (cb *CriteriaBuilder) LessThanOrEqualSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? <= (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrLessThanOrEqualSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? <= (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) LessThanOrEqualAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? <= ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrLessThanOrEqualAny(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? <= ANY (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) LessThanOrEqualAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? <= ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrLessThanOrEqualAll(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? <= ALL (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) LessThanOrEqualExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("? <= ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrLessThanOrEqualExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("? <= ?", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) Between(column string, start, end any) ConditionBuilder {
	cb.and("? BETWEEN ? AND ?", cb.eb.Column(column), start, end)

	return cb
}

func (cb *CriteriaBuilder) OrBetween(column string, start, end any) ConditionBuilder {
	cb.or("? BETWEEN ? AND ?", cb.eb.Column(column), start, end)

	return cb
}

func (cb *CriteriaBuilder) BetweenExpr(column string, startB, endB func(ExprBuilder) any) ConditionBuilder {
	cb.and("? BETWEEN ? AND ?", cb.eb.Column(column), startB(cb.eb), endB(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrBetweenExpr(column string, startB, endB func(ExprBuilder) any) ConditionBuilder {
	cb.or("? BETWEEN ? AND ?", cb.eb.Column(column), startB(cb.eb), endB(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) NotBetween(column string, start, end any) ConditionBuilder {
	cb.and("? NOT BETWEEN ? AND ?", cb.eb.Column(column), start, end)

	return cb
}

func (cb *CriteriaBuilder) OrNotBetween(column string, start, end any) ConditionBuilder {
	cb.or("? NOT BETWEEN ? AND ?", cb.eb.Column(column), start, end)

	return cb
}

func (cb *CriteriaBuilder) NotBetweenExpr(column string, startB, endB func(ExprBuilder) any) ConditionBuilder {
	cb.and("? NOT BETWEEN ? AND ?", cb.eb.Column(column), startB(cb.eb), endB(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrNotBetweenExpr(column string, startB, endB func(ExprBuilder) any) ConditionBuilder {
	cb.or("? NOT BETWEEN ? AND ?", cb.eb.Column(column), startB(cb.eb), endB(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) In(column string, values any) ConditionBuilder {
	cb.and("? IN (?)", cb.eb.Column(column), bun.In(values))

	return cb
}

func (cb *CriteriaBuilder) OrIn(column string, values any) ConditionBuilder {
	cb.or("? IN (?)", cb.eb.Column(column), bun.In(values))

	return cb
}

func (cb *CriteriaBuilder) InSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? IN (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrInSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? IN (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) InExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("? IN (?)", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrInExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("? IN (?)", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) NotIn(column string, values any) ConditionBuilder {
	cb.and("? NOT IN (?)", cb.eb.Column(column), bun.In(values))

	return cb
}

func (cb *CriteriaBuilder) OrNotIn(column string, values any) ConditionBuilder {
	cb.or("? NOT IN (?)", cb.eb.Column(column), bun.In(values))

	return cb
}

func (cb *CriteriaBuilder) NotInSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.and("? NOT IN (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrNotInSubQuery(column string, builder func(query SelectQuery)) ConditionBuilder {
	cb.or("? NOT IN (?)", cb.eb.Column(column), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) NotInExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("? NOT IN (?)", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrNotInExpr(column string, builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("? NOT IN (?)", cb.eb.Column(column), builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) IsNull(column string) ConditionBuilder {
	cb.and("? IS NULL", cb.eb.Column(column))

	return cb
}

func (cb *CriteriaBuilder) OrIsNull(column string) ConditionBuilder {
	cb.or("? IS NULL", cb.eb.Column(column))

	return cb
}

func (cb *CriteriaBuilder) IsNullSubQuery(builder func(query SelectQuery)) ConditionBuilder {
	cb.and("(?) IS NULL", cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrIsNullSubQuery(builder func(query SelectQuery)) ConditionBuilder {
	cb.or("(?) IS NULL", cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) IsNullExpr(builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("(?) IS NULL", builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrIsNullExpr(builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("(?) IS NULL", builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) IsNotNull(column string) ConditionBuilder {
	cb.and("? IS NOT NULL", cb.eb.Column(column))

	return cb
}

func (cb *CriteriaBuilder) OrIsNotNull(column string) ConditionBuilder {
	cb.or("? IS NOT NULL", cb.eb.Column(column))

	return cb
}

func (cb *CriteriaBuilder) IsNotNullSubQuery(builder func(query SelectQuery)) ConditionBuilder {
	cb.and("(?) IS NOT NULL", cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrIsNotNullSubQuery(builder func(query SelectQuery)) ConditionBuilder {
	cb.or("(?) IS NOT NULL", cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) IsNotNullExpr(builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("(?) IS NOT NULL", builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrIsNotNullExpr(builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("(?) IS NOT NULL", builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) IsTrue(column string) ConditionBuilder {
	cb.and("? IS TRUE", cb.eb.Column(column))

	return cb
}

func (cb *CriteriaBuilder) OrIsTrue(column string) ConditionBuilder {
	cb.or("? IS TRUE", cb.eb.Column(column))

	return cb
}

func (cb *CriteriaBuilder) IsTrueSubQuery(builder func(query SelectQuery)) ConditionBuilder {
	cb.and("(?) IS TRUE", cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrIsTrueSubQuery(builder func(query SelectQuery)) ConditionBuilder {
	cb.or("(?) IS TRUE", cb.qb.BuildSubQuery(builder))

	return cb
}

// IsTrueExpr adds an IS TRUE check for a custom expression.
func (cb *CriteriaBuilder) IsTrueExpr(builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("(?) IS TRUE", builder(cb.eb))

	return cb
}

// OrIsTrueExpr adds an OR IS TRUE check for a custom expression.
func (cb *CriteriaBuilder) OrIsTrueExpr(builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("(?) IS TRUE", builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) IsFalse(column string) ConditionBuilder {
	cb.and("? IS FALSE", cb.eb.Column(column))

	return cb
}

func (cb *CriteriaBuilder) OrIsFalse(column string) ConditionBuilder {
	cb.or("? IS FALSE", cb.eb.Column(column))

	return cb
}

func (cb *CriteriaBuilder) IsFalseSubQuery(builder func(query SelectQuery)) ConditionBuilder {
	cb.and("(?) IS FALSE", cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrIsFalseSubQuery(builder func(query SelectQuery)) ConditionBuilder {
	cb.or("(?) IS FALSE", cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) IsFalseExpr(builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("(?) IS FALSE", builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrIsFalseExpr(builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("(?) IS FALSE", builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) Contains(column, value string) ConditionBuilder {
	cb.and("? LIKE ?", cb.eb.Column(column), FuzzyContains.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) OrContains(column, value string) ConditionBuilder {
	cb.or("? LIKE ?", cb.eb.Column(column), FuzzyContains.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) ContainsAny(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrContains(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrContainsAny(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrContains(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) ContainsIgnoreCase(column, value string) ConditionBuilder {
	// Use ILIKE on Postgres; fallback to LOWER(column) LIKE LOWER(value) on others
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? ILIKE ?", cb.eb.Column(column), FuzzyContains.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyContains.BuildPattern(value)),
			)
		},
	})
	cb.and("?", expr)

	return cb
}

func (cb *CriteriaBuilder) OrContainsIgnoreCase(column, value string) ConditionBuilder {
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? ILIKE ?", cb.eb.Column(column), FuzzyContains.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyContains.BuildPattern(value)),
			)
		},
	})
	cb.or("?", expr)

	return cb
}

func (cb *CriteriaBuilder) ContainsAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrContainsIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrContainsAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrContainsIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) NotContains(column, value string) ConditionBuilder {
	cb.and("? NOT LIKE ?", cb.eb.Column(column), FuzzyContains.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) OrNotContains(column, value string) ConditionBuilder {
	cb.or("? NOT LIKE ?", cb.eb.Column(column), FuzzyContains.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) NotContainsAny(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotContains(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrNotContainsAny(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotContains(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) NotContainsIgnoreCase(column, value string) ConditionBuilder {
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? NOT ILIKE ?", cb.eb.Column(column), FuzzyContains.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? NOT LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyContains.BuildPattern(value)),
			)
		},
	})
	cb.and("?", expr)

	return cb
}

func (cb *CriteriaBuilder) OrNotContainsIgnoreCase(column, value string) ConditionBuilder {
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? NOT ILIKE ?", cb.eb.Column(column), FuzzyContains.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? NOT LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyContains.BuildPattern(value)),
			)
		},
	})
	cb.or("?", expr)

	return cb
}

func (cb *CriteriaBuilder) NotContainsAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotContainsIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrNotContainsAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotContainsIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) StartsWith(column, value string) ConditionBuilder {
	cb.and("? LIKE ?", cb.eb.Column(column), FuzzyStarts.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) OrStartsWith(column, value string) ConditionBuilder {
	cb.or("? LIKE ?", cb.eb.Column(column), FuzzyStarts.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) StartsWithAny(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrStartsWith(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrStartsWithAny(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrStartsWith(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) StartsWithIgnoreCase(column, value string) ConditionBuilder {
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? ILIKE ?", cb.eb.Column(column), FuzzyStarts.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyStarts.BuildPattern(value)),
			)
		},
	})
	cb.and("?", expr)

	return cb
}

func (cb *CriteriaBuilder) OrStartsWithIgnoreCase(column, value string) ConditionBuilder {
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? ILIKE ?", cb.eb.Column(column), FuzzyStarts.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyStarts.BuildPattern(value)),
			)
		},
	})
	cb.or("?", expr)

	return cb
}

func (cb *CriteriaBuilder) StartsWithAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrStartsWithIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrStartsWithAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrStartsWithIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) NotStartsWith(column, value string) ConditionBuilder {
	cb.and("? NOT LIKE ?", cb.eb.Column(column), FuzzyStarts.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) OrNotStartsWith(column, value string) ConditionBuilder {
	cb.or("? NOT LIKE ?", cb.eb.Column(column), FuzzyStarts.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) NotStartsWithAny(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotStartsWith(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrNotStartsWithAny(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotStartsWith(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) NotStartsWithIgnoreCase(column, value string) ConditionBuilder {
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? NOT ILIKE ?", cb.eb.Column(column), FuzzyStarts.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? NOT LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyStarts.BuildPattern(value)),
			)
		},
	})
	cb.and("?", expr)

	return cb
}

func (cb *CriteriaBuilder) OrNotStartsWithIgnoreCase(column, value string) ConditionBuilder {
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? NOT ILIKE ?", cb.eb.Column(column), FuzzyStarts.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? NOT LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyStarts.BuildPattern(value)),
			)
		},
	})
	cb.or("?", expr)

	return cb
}

func (cb *CriteriaBuilder) NotStartsWithAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotStartsWithIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrNotStartsWithAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotStartsWithIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) EndsWith(column, value string) ConditionBuilder {
	cb.and("? LIKE ?", cb.eb.Column(column), FuzzyEnds.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) OrEndsWith(column, value string) ConditionBuilder {
	cb.or("? LIKE ?", cb.eb.Column(column), FuzzyEnds.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) EndsWithAny(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrEndsWith(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrEndsWithAny(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrEndsWith(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) EndsWithIgnoreCase(column, value string) ConditionBuilder {
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? ILIKE ?", cb.eb.Column(column), FuzzyEnds.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyEnds.BuildPattern(value)),
			)
		},
	})
	cb.and("?", expr)

	return cb
}

func (cb *CriteriaBuilder) OrEndsWithIgnoreCase(column, value string) ConditionBuilder {
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? ILIKE ?", cb.eb.Column(column), FuzzyEnds.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyEnds.BuildPattern(value)),
			)
		},
	})
	cb.or("?", expr)

	return cb
}

func (cb *CriteriaBuilder) EndsWithAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrEndsWithIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrEndsWithAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.OrEndsWithIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) NotEndsWith(column, value string) ConditionBuilder {
	cb.and("? NOT LIKE ?", cb.eb.Column(column), FuzzyEnds.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) OrNotEndsWith(column, value string) ConditionBuilder {
	cb.or("? NOT LIKE ?", cb.eb.Column(column), FuzzyEnds.BuildPattern(value))

	return cb
}

func (cb *CriteriaBuilder) NotEndsWithAny(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotEndsWith(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrNotEndsWithAny(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotEndsWith(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) NotEndsWithIgnoreCase(column, value string) ConditionBuilder {
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? NOT ILIKE ?", cb.eb.Column(column), FuzzyEnds.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? NOT LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyEnds.BuildPattern(value)),
			)
		},
	})
	cb.and("?", expr)

	return cb
}

func (cb *CriteriaBuilder) OrNotEndsWithIgnoreCase(column, value string) ConditionBuilder {
	expr := cb.eb.ExprByDialect(DialectExprs{
		Postgres: func() schema.QueryAppender {
			return cb.eb.Expr("? NOT ILIKE ?", cb.eb.Column(column), FuzzyEnds.BuildPattern(value))
		},
		Default: func() schema.QueryAppender {
			return cb.eb.Expr(
				"? NOT LIKE ?",
				cb.eb.Lower(cb.eb.Column(column)),
				cb.eb.Lower(FuzzyEnds.BuildPattern(value)),
			)
		},
	})
	cb.or("?", expr)

	return cb
}

func (cb *CriteriaBuilder) NotEndsWithAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.Group(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotEndsWithIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) OrNotEndsWithAnyIgnoreCase(column string, values []string) ConditionBuilder {
	cb.OrGroup(func(cb ConditionBuilder) {
		for _, value := range values {
			cb.NotEndsWithIgnoreCase(column, value)
		}
	})

	return cb
}

func (cb *CriteriaBuilder) Expr(builder func(ExprBuilder) any) ConditionBuilder {
	cb.and("?", builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) OrExpr(builder func(ExprBuilder) any) ConditionBuilder {
	cb.or("?", builder(cb.eb))

	return cb
}

func (cb *CriteriaBuilder) Group(builder func(ConditionBuilder)) ConditionBuilder {
	cb.group(separatorAnd, builder)

	return cb
}

func (cb *CriteriaBuilder) OrGroup(builder func(ConditionBuilder)) ConditionBuilder {
	cb.group(separatorOr, builder)

	return cb
}

func (cb *CriteriaBuilder) CreatedByEquals(createdBy string, alias ...string) ConditionBuilder {
	cb.and("? = ?", buildColumnExpr(constants.ColumnCreatedBy, alias...), createdBy)

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByEquals(createdBy string, alias ...string) ConditionBuilder {
	cb.or("? = ?", buildColumnExpr(constants.ColumnCreatedBy, alias...), createdBy)

	return cb
}

func (cb *CriteriaBuilder) CreatedByEqualsSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? = (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByEqualsSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? = (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) CreatedByEqualsAny(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? = ANY (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByEqualsAny(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? = ANY (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) CreatedByEqualsAll(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? = ALL (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByEqualsAll(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? = ALL (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) CreatedByEqualsCurrent(alias ...string) ConditionBuilder {
	cb.and("? = ?Operator", buildColumnExpr(constants.ColumnCreatedBy, alias...))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByEqualsCurrent(alias ...string) ConditionBuilder {
	cb.or("? = ?Operator", buildColumnExpr(constants.ColumnCreatedBy, alias...))

	return cb
}

func (cb *CriteriaBuilder) CreatedByNotEquals(createdBy string, alias ...string) ConditionBuilder {
	cb.and("? <> ?", buildColumnExpr(constants.ColumnCreatedBy, alias...), createdBy)

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByNotEquals(createdBy string, alias ...string) ConditionBuilder {
	cb.or("? <> ?", buildColumnExpr(constants.ColumnCreatedBy, alias...), createdBy)

	return cb
}

func (cb *CriteriaBuilder) CreatedByNotEqualsSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? <> (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByNotEqualsSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? <> (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) CreatedByNotEqualsAny(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? <> ANY (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByNotEqualsAny(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? <> ANY (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) CreatedByNotEqualsAll(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? <> ALL (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByNotEqualsAll(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? <> ALL (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) CreatedByNotEqualsCurrent(alias ...string) ConditionBuilder {
	cb.and("? <> ?Operator", buildColumnExpr(constants.ColumnCreatedBy, alias...))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByNotEqualsCurrent(alias ...string) ConditionBuilder {
	cb.or("? <> ?Operator", buildColumnExpr(constants.ColumnCreatedBy, alias...))

	return cb
}

func (cb *CriteriaBuilder) CreatedByIn(createdBys []string, alias ...string) ConditionBuilder {
	cb.and("? IN (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), bun.In(createdBys))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByIn(createdBys []string, alias ...string) ConditionBuilder {
	cb.or("? IN (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), bun.In(createdBys))

	return cb
}

func (cb *CriteriaBuilder) CreatedByInSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? IN (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByInSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? IN (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) CreatedByNotIn(createdBys []string, alias ...string) ConditionBuilder {
	cb.and("? NOT IN (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), bun.In(createdBys))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByNotIn(createdBys []string, alias ...string) ConditionBuilder {
	cb.or("? NOT IN (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), bun.In(createdBys))

	return cb
}

func (cb *CriteriaBuilder) CreatedByNotInSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? NOT IN (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrCreatedByNotInSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? NOT IN (?)", buildColumnExpr(constants.ColumnCreatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByEquals(updatedBy string, alias ...string) ConditionBuilder {
	cb.and("? = ?", buildColumnExpr(constants.ColumnUpdatedBy, alias...), updatedBy)

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByEquals(updatedBy string, alias ...string) ConditionBuilder {
	cb.or("? = ?", buildColumnExpr(constants.ColumnUpdatedBy, alias...), updatedBy)

	return cb
}

func (cb *CriteriaBuilder) UpdatedByEqualsSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? = (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByEqualsSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? = (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByEqualsAny(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? = ANY (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByEqualsAny(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? = ANY (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByEqualsAll(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? = ALL (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByEqualsAll(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? = ALL (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByEqualsCurrent(alias ...string) ConditionBuilder {
	cb.and("? = ?Operator", buildColumnExpr(constants.ColumnUpdatedBy, alias...))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByEqualsCurrent(alias ...string) ConditionBuilder {
	cb.or("? = ?Operator", buildColumnExpr(constants.ColumnUpdatedBy, alias...))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByNotEquals(updatedBy string, alias ...string) ConditionBuilder {
	cb.and("? <> ?", buildColumnExpr(constants.ColumnUpdatedBy, alias...), updatedBy)

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByNotEquals(updatedBy string, alias ...string) ConditionBuilder {
	cb.or("? <> ?", buildColumnExpr(constants.ColumnUpdatedBy, alias...), updatedBy)

	return cb
}

func (cb *CriteriaBuilder) UpdatedByNotEqualsSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? <> (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByNotEqualsSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? <> (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByNotEqualsAny(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? <> ANY (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByNotEqualsAny(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? <> ANY (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByNotEqualsAll(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? <> ALL (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByNotEqualsAll(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? <> ALL (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByNotEqualsCurrent(alias ...string) ConditionBuilder {
	cb.and("? <> ?Operator", buildColumnExpr(constants.ColumnUpdatedBy, alias...))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByNotEqualsCurrent(alias ...string) ConditionBuilder {
	cb.or("? <> ?Operator", buildColumnExpr(constants.ColumnUpdatedBy, alias...))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByIn(updatedBys []string, alias ...string) ConditionBuilder {
	cb.and("? IN (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), bun.In(updatedBys))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByIn(updatedBys []string, alias ...string) ConditionBuilder {
	cb.or("? IN (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), bun.In(updatedBys))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByInSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? IN (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByInSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? IN (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByNotIn(updatedBys []string, alias ...string) ConditionBuilder {
	cb.and("? NOT IN (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), bun.In(updatedBys))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByNotIn(updatedBys []string, alias ...string) ConditionBuilder {
	cb.or("? NOT IN (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), bun.In(updatedBys))

	return cb
}

func (cb *CriteriaBuilder) UpdatedByNotInSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.and("? NOT IN (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedByNotInSubQuery(builder func(SelectQuery), alias ...string) ConditionBuilder {
	cb.or("? NOT IN (?)", buildColumnExpr(constants.ColumnUpdatedBy, alias...), cb.qb.BuildSubQuery(builder))

	return cb
}

func (cb *CriteriaBuilder) CreatedAtGreaterThan(createdAt time.Time, alias ...string) ConditionBuilder {
	cb.and("? > ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), createdAt)

	return cb
}

func (cb *CriteriaBuilder) OrCreatedAtGreaterThan(createdAt time.Time, alias ...string) ConditionBuilder {
	cb.or("? > ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), createdAt)

	return cb
}

func (cb *CriteriaBuilder) CreatedAtGreaterThanOrEqual(createdAt time.Time, alias ...string) ConditionBuilder {
	cb.and("? >= ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), createdAt)

	return cb
}

func (cb *CriteriaBuilder) OrCreatedAtGreaterThanOrEqual(createdAt time.Time, alias ...string) ConditionBuilder {
	cb.or("? >= ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), createdAt)

	return cb
}

func (cb *CriteriaBuilder) CreatedAtLessThan(createdAt time.Time, alias ...string) ConditionBuilder {
	cb.and("? < ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), createdAt)

	return cb
}

func (cb *CriteriaBuilder) OrCreatedAtLessThan(createdAt time.Time, alias ...string) ConditionBuilder {
	cb.or("? < ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), createdAt)

	return cb
}

func (cb *CriteriaBuilder) CreatedAtLessThanOrEqual(createdAt time.Time, alias ...string) ConditionBuilder {
	cb.and("? <= ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), createdAt)

	return cb
}

func (cb *CriteriaBuilder) OrCreatedAtLessThanOrEqual(createdAt time.Time, alias ...string) ConditionBuilder {
	cb.or("? <= ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), createdAt)

	return cb
}

func (cb *CriteriaBuilder) CreatedAtBetween(start, end time.Time, alias ...string) ConditionBuilder {
	cb.and("? BETWEEN ? AND ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), start, end)

	return cb
}

func (cb *CriteriaBuilder) OrCreatedAtBetween(start, end time.Time, alias ...string) ConditionBuilder {
	cb.or("? BETWEEN ? AND ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), start, end)

	return cb
}

func (cb *CriteriaBuilder) CreatedAtNotBetween(start, end time.Time, alias ...string) ConditionBuilder {
	cb.and("? NOT BETWEEN ? AND ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), start, end)

	return cb
}

func (cb *CriteriaBuilder) OrCreatedAtNotBetween(start, end time.Time, alias ...string) ConditionBuilder {
	cb.or("? NOT BETWEEN ? AND ?", buildColumnExpr(constants.ColumnCreatedAt, alias...), start, end)

	return cb
}

func (cb *CriteriaBuilder) UpdatedAtGreaterThan(updatedAt time.Time, alias ...string) ConditionBuilder {
	cb.and("? > ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), updatedAt)

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedAtGreaterThan(updatedAt time.Time, alias ...string) ConditionBuilder {
	cb.or("? > ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), updatedAt)

	return cb
}

func (cb *CriteriaBuilder) UpdatedAtGreaterThanOrEqual(updatedAt time.Time, alias ...string) ConditionBuilder {
	cb.and("? >= ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), updatedAt)

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedAtGreaterThanOrEqual(updatedAt time.Time, alias ...string) ConditionBuilder {
	cb.or("? >= ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), updatedAt)

	return cb
}

func (cb *CriteriaBuilder) UpdatedAtLessThan(updatedAt time.Time, alias ...string) ConditionBuilder {
	cb.and("? < ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), updatedAt)

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedAtLessThan(updatedAt time.Time, alias ...string) ConditionBuilder {
	cb.or("? < ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), updatedAt)

	return cb
}

func (cb *CriteriaBuilder) UpdatedAtLessThanOrEqual(updatedAt time.Time, alias ...string) ConditionBuilder {
	cb.and("? <= ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), updatedAt)

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedAtLessThanOrEqual(updatedAt time.Time, alias ...string) ConditionBuilder {
	cb.or("? <= ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), updatedAt)

	return cb
}

func (cb *CriteriaBuilder) UpdatedAtBetween(start, end time.Time, alias ...string) ConditionBuilder {
	cb.and("? BETWEEN ? AND ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), start, end)

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedAtBetween(start, end time.Time, alias ...string) ConditionBuilder {
	cb.or("? BETWEEN ? AND ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), start, end)

	return cb
}

func (cb *CriteriaBuilder) UpdatedAtNotBetween(start, end time.Time, alias ...string) ConditionBuilder {
	cb.and("? NOT BETWEEN ? AND ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), start, end)

	return cb
}

func (cb *CriteriaBuilder) OrUpdatedAtNotBetween(start, end time.Time, alias ...string) ConditionBuilder {
	cb.or("? NOT BETWEEN ? AND ?", buildColumnExpr(constants.ColumnUpdatedAt, alias...), start, end)

	return cb
}

func (cb *CriteriaBuilder) PKEquals(pk any, alias ...string) ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("PKEquals", cb.qb.GetTable(), pk, alias...)
	cb.and("? = ?", pc, pv)

	return cb
}

func (cb *CriteriaBuilder) OrPKEquals(pk any, alias ...string) ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("OrPKEquals", cb.qb.GetTable(), pk, alias...)
	cb.or("? = ?", pc, pv)

	return cb
}

func (cb *CriteriaBuilder) PKNotEquals(pk any, alias ...string) ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("PKNotEquals", cb.qb.GetTable(), pk, alias...)
	cb.and("? <> ?", pc, pv)

	return cb
}

func (cb *CriteriaBuilder) OrPKNotEquals(pk any, alias ...string) ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("OrPKNotEquals", cb.qb.GetTable(), pk, alias...)
	cb.or("? <> ?", pc, pv)

	return cb
}

func (cb *CriteriaBuilder) PKIn(pks any, alias ...string) ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("PKIn", cb.qb.GetTable(), pks, alias...)
	cb.and("? IN (?)", pc, pv)

	return cb
}

func (cb *CriteriaBuilder) OrPKIn(pks any, alias ...string) ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("OrPKIn", cb.qb.GetTable(), pks, alias...)
	cb.or("? IN (?)", pc, pv)

	return cb
}

func (cb *CriteriaBuilder) PKNotIn(pks any, alias ...string) ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("PKNotIn", cb.qb.GetTable(), pks, alias...)
	cb.and("? NOT IN (?)", pc, pv)

	return cb
}

func (cb *CriteriaBuilder) OrPKNotIn(pks any, alias ...string) ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("OrPKNotIn", cb.qb.GetTable(), pks, alias...)
	cb.or("? NOT IN (?)", pc, pv)

	return cb
}
