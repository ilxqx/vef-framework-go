package orm

import (
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// richConditionBuilder is a builder for building rich conditions.
type richConditionBuilder struct {
	table           *schema.Table
	subQueryBuilder func(builder func(query orm.Query)) *bun.SelectQuery
	and             func(query string, args ...any)
	or              func(query string, args ...any)
	group           func(sep string, builder func(orm.ConditionBuilder))
}

func (cb *richConditionBuilder) Apply(fns ...orm.ApplyFunc[orm.ConditionBuilder]) orm.ConditionBuilder {
	for _, fn := range fns {
		if fn != nil {
			fn(cb)
		}
	}

	return cb
}

func (cb *richConditionBuilder) ApplyIf(condition bool, fns ...orm.ApplyFunc[orm.ConditionBuilder]) orm.ConditionBuilder {
	if condition {
		return cb.Apply(fns...)
	}
	return cb
}

func (cb *richConditionBuilder) Equals(column string, value any) orm.ConditionBuilder {
	cb.and("? = ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) OrEquals(column string, value any) orm.ConditionBuilder {
	cb.or("? = ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) EqualsColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.and("? = ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) OrEqualsColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.or("? = ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) EqualsSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("? = (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrEqualsSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("? = (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) EqualsExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.and("? = ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrEqualsExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.or("? = ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) NotEquals(column string, value any) orm.ConditionBuilder {
	cb.and("? <> ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) OrNotEquals(column string, value any) orm.ConditionBuilder {
	cb.or("? <> ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) NotEqualsColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.and("? <> ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) OrNotEqualsColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.or("? <> ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) NotEqualsSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("? <> (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrNotEqualsSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("? <> (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) NotEqualsExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.and("? <> ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrNotEqualsExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.or("? <> ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) GreaterThan(column string, value any) orm.ConditionBuilder {
	cb.and("? > ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) OrGreaterThan(column string, value any) orm.ConditionBuilder {
	cb.or("? > ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) GreaterThanColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.and("? > ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) OrGreaterThanColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.or("? > ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) GreaterThanSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("? > (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrGreaterThanSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("? > (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) GreaterThanExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.and("? > ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrGreaterThanExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.or("? > ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) GreaterThanOrEqual(column string, value any) orm.ConditionBuilder {
	cb.and("? >= ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) OrGreaterThanOrEqual(column string, value any) orm.ConditionBuilder {
	cb.or("? >= ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) GreaterThanOrEqualColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.and("? >= ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) OrGreaterThanOrEqualColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.or("? >= ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) GreaterThanOrEqualSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("? >= (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrGreaterThanOrEqualSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("? >= (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) GreaterThanOrEqualExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.and("? >= ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrGreaterThanOrEqualExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.or("? >= ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) LessThan(column string, value any) orm.ConditionBuilder {
	cb.and("? < ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) OrLessThan(column string, value any) orm.ConditionBuilder {
	cb.or("? < ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) LessThanColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.and("? < ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) OrLessThanColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.or("? < ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) LessThanSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("? < (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrLessThanSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("? < (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) LessThanExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.and("? < ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrLessThanExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.or("? < ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) LessThanOrEqual(column string, value any) orm.ConditionBuilder {
	cb.and("? <= ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) OrLessThanOrEqual(column string, value any) orm.ConditionBuilder {
	cb.or("? <= ?", parseColumnExpr(column), value)
	return cb
}

func (cb *richConditionBuilder) LessThanOrEqualColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.and("? <= ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) OrLessThanOrEqualColumn(column1 string, column2 string) orm.ConditionBuilder {
	cb.or("? <= ?", parseColumnExpr(column1), parseColumnExpr(column2))
	return cb
}

func (cb *richConditionBuilder) LessThanOrEqualSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("? <= (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrLessThanOrEqualSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("? <= (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) LessThanOrEqualExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.and("? <= ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrLessThanOrEqualExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.or("? <= ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) Between(column string, start any, end any) orm.ConditionBuilder {
	cb.and("? BETWEEN ? AND ?", parseColumnExpr(column), start, end)
	return cb
}

func (cb *richConditionBuilder) OrBetween(column string, start any, end any) orm.ConditionBuilder {
	cb.or("? BETWEEN ? AND ?", parseColumnExpr(column), start, end)
	return cb
}

func (cb *richConditionBuilder) BetweenExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.and("? BETWEEN ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrBetweenExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.or("? BETWEEN ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) NotBetween(column string, start any, end any) orm.ConditionBuilder {
	cb.and("? NOT BETWEEN ? AND ?", parseColumnExpr(column), start, end)
	return cb
}

func (cb *richConditionBuilder) OrNotBetween(column string, start any, end any) orm.ConditionBuilder {
	cb.or("? NOT BETWEEN ? AND ?", parseColumnExpr(column), start, end)
	return cb
}

func (cb *richConditionBuilder) NotBetweenExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.and("? NOT BETWEEN ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrNotBetweenExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.or("? NOT BETWEEN ?", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) In(column string, values any) orm.ConditionBuilder {
	cb.and("? IN (?)", parseColumnExpr(column), bun.In(values))
	return cb
}

func (cb *richConditionBuilder) OrIn(column string, values any) orm.ConditionBuilder {
	cb.or("? IN (?)", parseColumnExpr(column), bun.In(values))
	return cb
}

func (cb *richConditionBuilder) InSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("? IN (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrInSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("? IN (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) InExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.and("? IN (?)", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrInExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.or("? IN (?)", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) NotIn(column string, values any) orm.ConditionBuilder {
	cb.and("? NOT IN (?)", parseColumnExpr(column), bun.In(values))
	return cb
}

func (cb *richConditionBuilder) OrNotIn(column string, values any) orm.ConditionBuilder {
	cb.or("? NOT IN (?)", parseColumnExpr(column), bun.In(values))
	return cb
}

func (cb *richConditionBuilder) NotInSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("? NOT IN (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrNotInSubQuery(column string, builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("? NOT IN (?)", parseColumnExpr(column), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) NotInExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.and("? NOT IN (?)", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrNotInExpr(column string, expr string, args ...any) orm.ConditionBuilder {
	cb.or("? NOT IN (?)", parseColumnExpr(column), bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) IsNull(column string) orm.ConditionBuilder {
	cb.and("? IS NULL", parseColumnExpr(column))
	return cb
}

func (cb *richConditionBuilder) OrIsNull(column string) orm.ConditionBuilder {
	cb.or("? IS NULL", parseColumnExpr(column))
	return cb
}

func (cb *richConditionBuilder) IsNullSubQuery(builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("(?) IS NULL", cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrIsNullSubQuery(builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("(?) IS NULL", cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) IsNullExpr(expr string, args ...any) orm.ConditionBuilder {
	cb.and("? IS NULL", bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrIsNullExpr(expr string, args ...any) orm.ConditionBuilder {
	cb.or("? IS NULL", bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) IsNotNull(column string) orm.ConditionBuilder {
	cb.and("? IS NOT NULL", parseColumnExpr(column))
	return cb
}

func (cb *richConditionBuilder) OrIsNotNull(column string) orm.ConditionBuilder {
	cb.or("? IS NOT NULL", parseColumnExpr(column))
	return cb
}

func (cb *richConditionBuilder) IsNotNullSubQuery(builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("(?) IS NOT NULL", cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrIsNotNullSubQuery(builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("(?) IS NOT NULL", cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) IsNotNullExpr(expr string, args ...any) orm.ConditionBuilder {
	cb.and("? IS NOT NULL", bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrIsNotNullExpr(expr string, args ...any) orm.ConditionBuilder {
	cb.or("? IS NOT NULL", bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) IsTrue(column string) orm.ConditionBuilder {
	cb.and("? IS TRUE", parseColumnExpr(column))
	return cb
}

func (cb *richConditionBuilder) OrIsTrue(column string) orm.ConditionBuilder {
	cb.or("? IS TRUE", parseColumnExpr(column))
	return cb
}

func (cb *richConditionBuilder) IsTrueSubQuery(builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("(?) IS TRUE", cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrIsTrueSubQuery(builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("(?) IS TRUE", cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) IsTrueExpr(expr string, args ...any) orm.ConditionBuilder {
	cb.and("? IS TRUE", bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrIsTrueExpr(expr string, args ...any) orm.ConditionBuilder {
	cb.or("? IS TRUE", bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) IsFalse(column string) orm.ConditionBuilder {
	cb.and("? IS FALSE", parseColumnExpr(column))
	return cb
}

func (cb *richConditionBuilder) OrIsFalse(column string) orm.ConditionBuilder {
	cb.or("? IS FALSE", parseColumnExpr(column))
	return cb
}

func (cb *richConditionBuilder) IsFalseSubQuery(builder func(query orm.Query)) orm.ConditionBuilder {
	cb.and("(?) IS FALSE", cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrIsFalseSubQuery(builder func(query orm.Query)) orm.ConditionBuilder {
	cb.or("(?) IS FALSE", cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) IsFalseExpr(expr string, args ...any) orm.ConditionBuilder {
	cb.and("? IS FALSE", bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrIsFalseExpr(expr string, args ...any) orm.ConditionBuilder {
	cb.or("? IS FALSE", bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) Contains(column string, value string) orm.ConditionBuilder {
	cb.and("? LIKE ?", parseColumnExpr(column), constants.Percent+value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) OrContains(column string, value string) orm.ConditionBuilder {
	cb.or("? LIKE ?", parseColumnExpr(column), constants.Percent+value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) ContainsAny(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrContains(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrContainsAny(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrContains(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) ContainsIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.and("? LIKE LOWER(?)", parseColumnExpr(column), constants.Percent+value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) OrContainsIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.or("? LIKE LOWER(?)", parseColumnExpr(column), constants.Percent+value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) ContainsAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrContainsIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrContainsAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrContainsIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) NotContains(column string, value string) orm.ConditionBuilder {
	cb.and("? NOT LIKE ?", parseColumnExpr(column), constants.Percent+value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) OrNotContains(column string, value string) orm.ConditionBuilder {
	cb.or("? NOT LIKE ?", parseColumnExpr(column), constants.Percent+value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) NotContainsAny(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotContains(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrNotContainsAny(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotContains(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) NotContainsIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.and("? NOT LIKE LOWER(?)", parseColumnExpr(column), constants.Percent+value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) OrNotContainsIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.or("? NOT LIKE LOWER(?)", parseColumnExpr(column), constants.Percent+value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) NotContainsAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotContainsIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrNotContainsAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotContainsIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) StartsWith(column string, value string) orm.ConditionBuilder {
	cb.and("? LIKE ?", parseColumnExpr(column), value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) OrStartsWith(column string, value string) orm.ConditionBuilder {
	cb.or("? LIKE ?", parseColumnExpr(column), value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) StartsWithAny(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrStartsWith(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrStartsWithAny(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrStartsWith(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) StartsWithIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.and("? LIKE LOWER(?)", parseColumnExpr(column), value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) OrStartsWithIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.or("? LIKE LOWER(?)", parseColumnExpr(column), value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) StartsWithAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrStartsWithIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrStartsWithAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrStartsWithIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) NotStartsWith(column string, value string) orm.ConditionBuilder {
	cb.and("? NOT LIKE ?", parseColumnExpr(column), value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) OrNotStartsWith(column string, value string) orm.ConditionBuilder {
	cb.or("? NOT LIKE ?", parseColumnExpr(column), value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) NotStartsWithAny(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotStartsWith(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrNotStartsWithAny(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotStartsWith(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) NotStartsWithIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.and("? NOT LIKE LOWER(?)", parseColumnExpr(column), value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) OrNotStartsWithIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.or("? NOT LIKE LOWER(?)", parseColumnExpr(column), value+constants.Percent)
	return cb
}

func (cb *richConditionBuilder) NotStartsWithAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotStartsWithIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrNotStartsWithAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotStartsWithIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) EndsWith(column string, value string) orm.ConditionBuilder {
	cb.and("? LIKE ?", parseColumnExpr(column), constants.Percent+value)
	return cb
}

func (cb *richConditionBuilder) OrEndsWith(column string, value string) orm.ConditionBuilder {
	cb.or("? LIKE ?", parseColumnExpr(column), constants.Percent+value)
	return cb
}

func (cb *richConditionBuilder) EndsWithAny(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrEndsWith(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrEndsWithAny(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrEndsWith(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) EndsWithIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.and("? LIKE LOWER(?)", parseColumnExpr(column), constants.Percent+value)
	return cb
}

func (cb *richConditionBuilder) OrEndsWithIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.or("? LIKE LOWER(?)", parseColumnExpr(column), constants.Percent+value)
	return cb
}

func (cb *richConditionBuilder) EndsWithAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrEndsWithIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrEndsWithAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.OrEndsWithIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) NotEndsWith(column string, value string) orm.ConditionBuilder {
	cb.and("? NOT LIKE ?", parseColumnExpr(column), constants.Percent+value)
	return cb
}

func (cb *richConditionBuilder) OrNotEndsWith(column string, value string) orm.ConditionBuilder {
	cb.or("? NOT LIKE ?", parseColumnExpr(column), constants.Percent+value)
	return cb
}

func (cb *richConditionBuilder) NotEndsWithAny(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotEndsWith(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrNotEndsWithAny(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotEndsWith(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) NotEndsWithIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.and("? NOT LIKE LOWER(?)", parseColumnExpr(column), constants.Percent+value)
	return cb
}

func (cb *richConditionBuilder) OrNotEndsWithIgnoreCase(column string, value string) orm.ConditionBuilder {
	cb.or("? NOT LIKE LOWER(?)", parseColumnExpr(column), constants.Percent+value)
	return cb
}

func (cb *richConditionBuilder) NotEndsWithAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotEndsWithIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) OrNotEndsWithAnyIgnoreCase(column string, values []string) orm.ConditionBuilder {
	cb.OrGroup(func(cb orm.ConditionBuilder) {
		for _, value := range values {
			cb.NotEndsWithIgnoreCase(column, value)
		}
	})
	return cb
}

func (cb *richConditionBuilder) Expr(expr string, args ...any) orm.ConditionBuilder {
	cb.and("?", bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) OrExpr(expr string, args ...any) orm.ConditionBuilder {
	cb.or("?", bun.SafeQuery(expr, args...))
	return cb
}

func (cb *richConditionBuilder) Group(builder func(orm.ConditionBuilder)) orm.ConditionBuilder {
	cb.group(orm.SeparatorAnd, builder)
	return cb
}

func (cb *richConditionBuilder) OrGroup(builder func(orm.ConditionBuilder)) orm.ConditionBuilder {
	cb.group(orm.SeparatorOr, builder)
	return cb
}

func (cb *richConditionBuilder) CreatedByEquals(createdBy string, alias ...string) orm.ConditionBuilder {
	cb.and("? = ?", buildColumnExpr(orm.ColumnCreatedBy, alias...), createdBy)
	return cb
}

func (cb *richConditionBuilder) OrCreatedByEquals(createdBy string, alias ...string) orm.ConditionBuilder {
	cb.or("? = ?", buildColumnExpr(orm.ColumnCreatedBy, alias...), createdBy)
	return cb
}

func (cb *richConditionBuilder) CreatedByEqualsSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.and("? = (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrCreatedByEqualsSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.or("? = (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) CreatedByEqualsCurrent(alias ...string) orm.ConditionBuilder {
	cb.and("? = ?Operator", buildColumnExpr(orm.ColumnCreatedBy, alias...))
	return cb
}

func (cb *richConditionBuilder) OrCreatedByEqualsCurrent(alias ...string) orm.ConditionBuilder {
	cb.or("? = ?Operator", buildColumnExpr(orm.ColumnCreatedBy, alias...))
	return cb
}

func (cb *richConditionBuilder) CreatedByNotEquals(createdBy string, alias ...string) orm.ConditionBuilder {
	cb.and("? <> ?", buildColumnExpr(orm.ColumnCreatedBy, alias...), createdBy)
	return cb
}

func (cb *richConditionBuilder) OrCreatedByNotEquals(createdBy string, alias ...string) orm.ConditionBuilder {
	cb.or("? <> ?", buildColumnExpr(orm.ColumnCreatedBy, alias...), createdBy)
	return cb
}

func (cb *richConditionBuilder) CreatedByNotEqualsSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.and("? <> (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrCreatedByNotEqualsSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.or("? <> (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) CreatedByNotEqualsCurrent(alias ...string) orm.ConditionBuilder {
	cb.and("? <> ?Operator", buildColumnExpr(orm.ColumnCreatedBy, alias...))
	return cb
}

func (cb *richConditionBuilder) OrCreatedByNotEqualsCurrent(alias ...string) orm.ConditionBuilder {
	cb.or("? <> ?Operator", buildColumnExpr(orm.ColumnCreatedBy, alias...))
	return cb
}

func (cb *richConditionBuilder) CreatedByIn(createdBys []string, alias ...string) orm.ConditionBuilder {
	cb.and("? IN (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), bun.In(createdBys))
	return cb
}

func (cb *richConditionBuilder) OrCreatedByIn(createdBys []string, alias ...string) orm.ConditionBuilder {
	cb.or("? IN (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), bun.In(createdBys))
	return cb
}

func (cb *richConditionBuilder) CreatedByInSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.and("? IN (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrCreatedByInSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.or("? IN (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) CreatedByNotIn(createdBys []string, alias ...string) orm.ConditionBuilder {
	cb.and("? NOT IN (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), bun.In(createdBys))
	return cb
}

func (cb *richConditionBuilder) OrCreatedByNotIn(createdBys []string, alias ...string) orm.ConditionBuilder {
	cb.or("? NOT IN (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), bun.In(createdBys))
	return cb
}

func (cb *richConditionBuilder) CreatedByNotInSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.and("? NOT IN (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrCreatedByNotInSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.or("? NOT IN (?)", buildColumnExpr(orm.ColumnCreatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) UpdatedByEquals(updatedBy string, alias ...string) orm.ConditionBuilder {
	cb.and("? = ?", buildColumnExpr(orm.ColumnUpdatedBy, alias...), updatedBy)
	return cb
}

func (cb *richConditionBuilder) OrUpdatedByEquals(updatedBy string, alias ...string) orm.ConditionBuilder {
	cb.or("? = ?", buildColumnExpr(orm.ColumnUpdatedBy, alias...), updatedBy)
	return cb
}

func (cb *richConditionBuilder) UpdatedByEqualsSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.and("? = (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrUpdatedByEqualsSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.or("? = (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) UpdatedByEqualsCurrent(alias ...string) orm.ConditionBuilder {
	cb.and("? = ?Operator", buildColumnExpr(orm.ColumnUpdatedBy, alias...))
	return cb
}

func (cb *richConditionBuilder) OrUpdatedByEqualsCurrent(alias ...string) orm.ConditionBuilder {
	cb.or("? = ?Operator", buildColumnExpr(orm.ColumnUpdatedBy, alias...))
	return cb
}

func (cb *richConditionBuilder) UpdatedByNotEquals(updatedBy string, alias ...string) orm.ConditionBuilder {
	cb.and("? <> ?", buildColumnExpr(orm.ColumnUpdatedBy, alias...), updatedBy)
	return cb
}

func (cb *richConditionBuilder) OrUpdatedByNotEquals(updatedBy string, alias ...string) orm.ConditionBuilder {
	cb.or("? <> ?", buildColumnExpr(orm.ColumnUpdatedBy, alias...), updatedBy)
	return cb
}

func (cb *richConditionBuilder) UpdatedByNotEqualsSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.and("? <> (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrUpdatedByNotEqualsSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.or("? <> (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) UpdatedByNotEqualsCurrent(alias ...string) orm.ConditionBuilder {
	cb.and("? <> ?Operator", buildColumnExpr(orm.ColumnUpdatedBy, alias...))
	return cb
}

func (cb *richConditionBuilder) OrUpdatedByNotEqualsCurrent(alias ...string) orm.ConditionBuilder {
	cb.or("? <> ?Operator", buildColumnExpr(orm.ColumnUpdatedBy, alias...))
	return cb
}

func (cb *richConditionBuilder) UpdatedByIn(updatedBys []string, alias ...string) orm.ConditionBuilder {
	cb.and("? IN (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), bun.In(updatedBys))
	return cb
}

func (cb *richConditionBuilder) OrUpdatedByIn(updatedBys []string, alias ...string) orm.ConditionBuilder {
	cb.or("? IN (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), bun.In(updatedBys))
	return cb
}

func (cb *richConditionBuilder) UpdatedByInSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.and("? IN (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrUpdatedByInSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.or("? IN (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) UpdatedByNotIn(updatedBys []string, alias ...string) orm.ConditionBuilder {
	cb.and("? NOT IN (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), bun.In(updatedBys))
	return cb
}

func (cb *richConditionBuilder) OrUpdatedByNotIn(updatedBys []string, alias ...string) orm.ConditionBuilder {
	cb.or("? NOT IN (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), bun.In(updatedBys))
	return cb
}

func (cb *richConditionBuilder) UpdatedByNotInSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.and("? NOT IN (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) OrUpdatedByNotInSubQuery(builder func(orm.Query), alias ...string) orm.ConditionBuilder {
	cb.or("? NOT IN (?)", buildColumnExpr(orm.ColumnUpdatedBy, alias...), cb.subQueryBuilder(builder))
	return cb
}

func (cb *richConditionBuilder) CreatedAtGreaterThan(createdAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? > ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), createdAt)
	return cb
}

func (cb *richConditionBuilder) OrCreatedAtGreaterThan(createdAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? > ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), createdAt)
	return cb
}

func (cb *richConditionBuilder) CreatedAtGreaterThanOrEqual(createdAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? >= ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), createdAt)
	return cb
}

func (cb *richConditionBuilder) OrCreatedAtGreaterThanOrEqual(createdAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? >= ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), createdAt)
	return cb
}

func (cb *richConditionBuilder) CreatedAtLessThan(createdAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? < ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), createdAt)
	return cb
}

func (cb *richConditionBuilder) OrCreatedAtLessThan(createdAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? < ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), createdAt)
	return cb
}

func (cb *richConditionBuilder) CreatedAtLessThanOrEqual(createdAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? <= ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), createdAt)
	return cb
}

func (cb *richConditionBuilder) OrCreatedAtLessThanOrEqual(createdAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? <= ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), createdAt)
	return cb
}

func (cb *richConditionBuilder) CreatedAtBetween(start time.Time, end time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? BETWEEN ? AND ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), start, end)
	return cb
}

func (cb *richConditionBuilder) OrCreatedAtBetween(start time.Time, end time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? BETWEEN ? AND ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), start, end)
	return cb
}

func (cb *richConditionBuilder) CreatedAtNotBetween(start time.Time, end time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? NOT BETWEEN ? AND ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), start, end)
	return cb
}

func (cb *richConditionBuilder) OrCreatedAtNotBetween(start time.Time, end time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? NOT BETWEEN ? AND ?", buildColumnExpr(orm.ColumnCreatedAt, alias...), start, end)
	return cb
}

func (cb *richConditionBuilder) UpdatedAtGreaterThan(updatedAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? > ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), updatedAt)
	return cb
}

func (cb *richConditionBuilder) OrUpdatedAtGreaterThan(updatedAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? > ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), updatedAt)
	return cb
}

func (cb *richConditionBuilder) UpdatedAtGreaterThanOrEqual(updatedAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? >= ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), updatedAt)
	return cb
}

func (cb *richConditionBuilder) OrUpdatedAtGreaterThanOrEqual(updatedAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? >= ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), updatedAt)
	return cb
}

func (cb *richConditionBuilder) UpdatedAtLessThan(updatedAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? < ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), updatedAt)
	return cb
}

func (cb *richConditionBuilder) OrUpdatedAtLessThan(updatedAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? < ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), updatedAt)
	return cb
}

func (cb *richConditionBuilder) UpdatedAtLessThanOrEqual(updatedAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? <= ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), updatedAt)
	return cb
}

func (cb *richConditionBuilder) OrUpdatedAtLessThanOrEqual(updatedAt time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? <= ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), updatedAt)
	return cb
}

func (cb *richConditionBuilder) UpdatedAtBetween(start time.Time, end time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? BETWEEN ? AND ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), start, end)
	return cb
}

func (cb *richConditionBuilder) OrUpdatedAtBetween(start time.Time, end time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? BETWEEN ? AND ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), start, end)
	return cb
}

func (cb *richConditionBuilder) UpdatedAtNotBetween(start time.Time, end time.Time, alias ...string) orm.ConditionBuilder {
	cb.and("? NOT BETWEEN ? AND ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), start, end)
	return cb
}

func (cb *richConditionBuilder) OrUpdatedAtNotBetween(start time.Time, end time.Time, alias ...string) orm.ConditionBuilder {
	cb.or("? NOT BETWEEN ? AND ?", buildColumnExpr(orm.ColumnUpdatedAt, alias...), start, end)
	return cb
}

func (cb *richConditionBuilder) PKEquals(pk any, alias ...string) orm.ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("PKEquals", cb.table, pk, alias...)
	cb.and("? = ?", pc, pv)
	return cb
}

func (cb *richConditionBuilder) OrPKEquals(pk any, alias ...string) orm.ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("OrPKEquals", cb.table, pk, alias...)
	cb.or("? = ?", pc, pv)
	return cb
}

func (cb *richConditionBuilder) PKNotEquals(pk any, alias ...string) orm.ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("PKNotEquals", cb.table, pk, alias...)
	cb.and("? <> ?", pc, pv)
	return cb
}

func (cb *richConditionBuilder) OrPKNotEquals(pk any, alias ...string) orm.ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("OrPKNotEquals", cb.table, pk, alias...)
	cb.or("? <> ?", pc, pv)
	return cb
}

func (cb *richConditionBuilder) PKIn(pks any, alias ...string) orm.ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("PKIn", cb.table, pks, alias...)
	cb.and("? IN (?)", pc, pv)
	return cb
}

func (cb *richConditionBuilder) OrPKIn(pks any, alias ...string) orm.ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("OrPKIn", cb.table, pks, alias...)
	cb.or("? IN (?)", pc, pv)
	return cb
}

func (cb *richConditionBuilder) PKNotIn(pks any, alias ...string) orm.ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("PKNotIn", cb.table, pks, alias...)
	cb.and("? NOT IN (?)", pc, pv)
	return cb
}

func (cb *richConditionBuilder) OrPKNotIn(pks any, alias ...string) orm.ConditionBuilder {
	pc, pv := parsePKColumnsAndValues("OrPKNotIn", cb.table, pks, alias...)
	cb.or("? NOT IN (?)", pc, pv)
	return cb
}
