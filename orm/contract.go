package orm

import (
	"context"
	"database/sql"
	"reflect"
	"time"

	"github.com/ilxqx/vef-framework-go/mo"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// AuditConditionBuilder is a builder for audit conditions.
type AuditConditionBuilder interface {
	// CreatedByEquals is a condition that checks if the created by column is equal to a value.
	CreatedByEquals(createdBy string, alias ...string) ConditionBuilder
	// OrCreatedByEquals is an OR condition that checks if the created by column is equal to a value.
	OrCreatedByEquals(createdBy string, alias ...string) ConditionBuilder
	// CreatedByEqualsSubQuery is a condition that checks if the created by column is equal to a subquery.
	CreatedByEqualsSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// OrCreatedByEqualsSubQuery is a condition that checks if the created by column is equal to a subquery.
	OrCreatedByEqualsSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// CreatedByEqualsCurrent is a condition that checks if the created by column is equal to the current user.
	CreatedByEqualsCurrent(alias ...string) ConditionBuilder
	// OrCreatedByEqualsCurrent is a condition that checks if the created by column is equal to the current user.
	OrCreatedByEqualsCurrent(alias ...string) ConditionBuilder
	// CreatedByNotEquals is a condition that checks if the created by column is not equal to a value.
	CreatedByNotEquals(createdBy string, alias ...string) ConditionBuilder
	// OrCreatedByNotEquals is a condition that checks if the created by column is not equal to a value.
	OrCreatedByNotEquals(createdBy string, alias ...string) ConditionBuilder
	// CreatedByNotEqualsSubQuery is a condition that checks if the created by column is not equal to a subquery.
	CreatedByNotEqualsSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// OrCreatedByNotEqualsSubQuery is a condition that checks if the created by column is not equal to a subquery.
	OrCreatedByNotEqualsSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// CreatedByNotEqualsCurrent is a condition that checks if the created by column is not equal to the current user.
	CreatedByNotEqualsCurrent(alias ...string) ConditionBuilder
	// OrCreatedByNotEqualsCurrent is a condition that checks if the created by column is not equal to the current user.
	OrCreatedByNotEqualsCurrent(alias ...string) ConditionBuilder
	// CreatedByIn is a condition that checks if the created by column is in a list of values.
	CreatedByIn(createdBys []string, alias ...string) ConditionBuilder
	// OrCreatedByIn is a condition that checks if the created by column is in a list of values.
	OrCreatedByIn(createdBys []string, alias ...string) ConditionBuilder
	// CreatedByInSubQuery is a condition that checks if the created by column is in a subquery.
	CreatedByInSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// CreatedByNotIn is a condition that checks if the created by column is not in a list of values.
	CreatedByNotIn(createdBys []string, alias ...string) ConditionBuilder
	// OrCreatedByNotIn is a condition that checks if the created by column is not in a list of values.
	OrCreatedByNotIn(createdBys []string, alias ...string) ConditionBuilder
	// CreatedByNotInSubQuery is a condition that checks if the created by column is not in a subquery.
	CreatedByNotInSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// OrCreatedByNotInSubQuery is a condition that checks if the created by column is not in a subquery.
	OrCreatedByNotInSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// UpdatedByEquals is a condition that checks if the updated by column is equal to a value.
	UpdatedByEquals(updatedBy string, alias ...string) ConditionBuilder
	// UpdatedByEqualsSubQuery is a condition that checks if the updated by column is equal to a subquery.
	UpdatedByEqualsSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// OrUpdatedByEqualsSubQuery is a condition that checks if the updated by column is equal to a subquery.
	OrUpdatedByEqualsSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// UpdatedByEqualsCurrent is a condition that checks if the updated by column is equal to the current user.
	UpdatedByEqualsCurrent(alias ...string) ConditionBuilder
	// OrUpdatedByEqualsCurrent is a condition that checks if the updated by column is equal to the current user.
	OrUpdatedByEqualsCurrent(alias ...string) ConditionBuilder
	// UpdatedByNotEquals is a condition that checks if the updated by column is not equal to a value.
	UpdatedByNotEquals(updatedBy string, alias ...string) ConditionBuilder
	// UpdatedByNotEqualsSubQuery is a condition that checks if the updated by column is not equal to a subquery.
	UpdatedByNotEqualsSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// OrUpdatedByNotEqualsSubQuery is a condition that checks if the updated by column is not equal to a subquery.
	OrUpdatedByNotEqualsSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// UpdatedByNotEqualsCurrent is a condition that checks if the updated by column is not equal to the current user.
	UpdatedByNotEqualsCurrent(alias ...string) ConditionBuilder
	// OrUpdatedByNotEqualsCurrent is a condition that checks if the updated by column is not equal to the current user.
	OrUpdatedByNotEqualsCurrent(alias ...string) ConditionBuilder
	// UpdatedByIn is a condition that checks if the updated by column is in a list of values.
	UpdatedByIn(updatedBys []string, alias ...string) ConditionBuilder
	// UpdatedByInSubQuery is a condition that checks if the updated by column is in a subquery.
	UpdatedByInSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// OrUpdatedByInSubQuery is a condition that checks if the updated by column is in a subquery.
	OrUpdatedByInSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// UpdatedByNotIn is a condition that checks if the updated by column is not in a list of values.
	UpdatedByNotIn(updatedBys []string, alias ...string) ConditionBuilder
	// OrUpdatedByNotIn is a condition that checks if the updated by column is not in a list of values.
	OrUpdatedByNotIn(updatedBys []string, alias ...string) ConditionBuilder
	// UpdatedByNotInSubQuery is a condition that checks if the updated by column is not in a subquery.
	UpdatedByNotInSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// OrUpdatedByNotInSubQuery is a condition that checks if the updated by column is not in a subquery.
	OrUpdatedByNotInSubQuery(builder func(query Query), alias ...string) ConditionBuilder
	// CreatedAtGreaterThan is a condition that checks if the created at column is greater than a value.
	CreatedAtGreaterThan(createdAt time.Time, alias ...string) ConditionBuilder
	// OrCreatedAtGreaterThan is a condition that checks if the created at column is greater than a value.
	OrCreatedAtGreaterThan(createdAt time.Time, alias ...string) ConditionBuilder
	// CreatedAtGreaterThanOrEqual is a condition that checks if the created at column is greater than or equal to a value.
	CreatedAtGreaterThanOrEqual(createdAt time.Time, alias ...string) ConditionBuilder
	// OrCreatedAtGreaterThanOrEqual is a condition that checks if the created at column is greater than or equal to a value.
	OrCreatedAtGreaterThanOrEqual(createdAt time.Time, alias ...string) ConditionBuilder
	// CreatedAtLessThan is a condition that checks if the created at column is less than a value.
	CreatedAtLessThan(createdAt time.Time, alias ...string) ConditionBuilder
	// OrCreatedAtLessThan is a condition that checks if the created at column is less than a value.
	OrCreatedAtLessThan(createdAt time.Time, alias ...string) ConditionBuilder
	// CreatedAtLessThanOrEqual is a condition that checks if the created at column is less than or equal to a value.
	CreatedAtLessThanOrEqual(createdAt time.Time, alias ...string) ConditionBuilder
	// OrCreatedAtLessThanOrEqual is a condition that checks if the created at column is less than or equal to a value.
	OrCreatedAtLessThanOrEqual(createdAt time.Time, alias ...string) ConditionBuilder
	// CreatedAtBetween is a condition that checks if the created at column is between two values.
	CreatedAtBetween(start, end time.Time, alias ...string) ConditionBuilder
	// OrCreatedAtBetween is a condition that checks if the created at column is between two values.
	OrCreatedAtBetween(start, end time.Time, alias ...string) ConditionBuilder
	// CreatedAtNotBetween is a condition that checks if the created at column is not between two values.
	CreatedAtNotBetween(start, end time.Time, alias ...string) ConditionBuilder
	// OrCreatedAtNotBetween is a condition that checks if the created at column is not between two values.
	OrCreatedAtNotBetween(start, end time.Time, alias ...string) ConditionBuilder
	// UpdatedAtGreaterThan is a condition that checks if the updated at column is greater than a value.
	UpdatedAtGreaterThan(updatedAt time.Time, alias ...string) ConditionBuilder
	// OrUpdatedAtGreaterThan is a condition that checks if the updated at column is greater than a value.
	OrUpdatedAtGreaterThan(updatedAt time.Time, alias ...string) ConditionBuilder
	// UpdatedAtGreaterThanOrEqual is a condition that checks if the updated at column is greater than or equal to a value.
	UpdatedAtGreaterThanOrEqual(updatedAt time.Time, alias ...string) ConditionBuilder
	// OrUpdatedAtGreaterThanOrEqual is a condition that checks if the updated at column is greater than or equal to a value.
	OrUpdatedAtGreaterThanOrEqual(updatedAt time.Time, alias ...string) ConditionBuilder
	// UpdatedAtLessThan is a condition that checks if the updated at column is less than a value.
	UpdatedAtLessThan(updatedAt time.Time, alias ...string) ConditionBuilder
	// OrUpdatedAtLessThan is a condition that checks if the updated at column is less than a value.
	OrUpdatedAtLessThan(updatedAt time.Time, alias ...string) ConditionBuilder
	// UpdatedAtLessThanOrEqual is a condition that checks if the updated at column is less than or equal to a value.
	UpdatedAtLessThanOrEqual(updatedAt time.Time, alias ...string) ConditionBuilder
	// OrUpdatedAtLessThanOrEqual is a condition that checks if the updated at column is less than or equal to a value.
	OrUpdatedAtLessThanOrEqual(updatedAt time.Time, alias ...string) ConditionBuilder
	// UpdatedAtBetween is a condition that checks if the updated at column is between two values.
	UpdatedAtBetween(start, end time.Time, alias ...string) ConditionBuilder
	// OrUpdatedAtBetween is a condition that checks if the updated at column is between two values.
	OrUpdatedAtBetween(start, end time.Time, alias ...string) ConditionBuilder
	// UpdatedAtNotBetween is a condition that checks if the updated at column is not between two values.
	UpdatedAtNotBetween(start, end time.Time, alias ...string) ConditionBuilder
	// OrUpdatedAtNotBetween is a condition that checks if the updated at column is not between two values.
	OrUpdatedAtNotBetween(start, end time.Time, alias ...string) ConditionBuilder
}

// PKConditionBuilder is a builder for primary key conditions.
type PKConditionBuilder interface {
	// PKEquals is a condition that checks if the primary key is equal to a value.
	PKEquals(pk any, alias ...string) ConditionBuilder
	// OrPKEquals is a condition that checks if the primary key is equal to a value.
	OrPKEquals(pk any, alias ...string) ConditionBuilder
	// PKNotEquals is a condition that checks if the primary key is not equal to a value.
	PKNotEquals(pk any, alias ...string) ConditionBuilder
	// OrPKNotEquals is a condition that checks if the primary key is not equal to a value.
	OrPKNotEquals(pk any, alias ...string) ConditionBuilder
	// PKIn is a condition that checks if the primary key is in a list of values.
	PKIn(pks any, alias ...string) ConditionBuilder
	// OrPKIn is a condition that checks if the primary key is in a list of values.
	OrPKIn(pks any, alias ...string) ConditionBuilder
	// PKNotIn is a condition that checks if the primary key is not in a list of values.
	PKNotIn(pks any, alias ...string) ConditionBuilder
	// OrPKNotIn is a condition that checks if the primary key is not in a list of values.
	OrPKNotIn(pks any, alias ...string) ConditionBuilder
}

// ConditionBuilder is a builder for conditions.
type ConditionBuilder interface {
	Applier[ConditionBuilder]
	AuditConditionBuilder
	PKConditionBuilder
	// Equals is a condition that checks if a column is equal to a value.
	Equals(column string, value any) ConditionBuilder
	// OrEquals is a condition that checks if a column is equal to a value.
	OrEquals(column string, value any) ConditionBuilder
	// EqualsColumn is a condition that checks if a column is equal to another column.
	EqualsColumn(column1, column2 string) ConditionBuilder
	// OrEqualsColumn is a condition that checks if a column is equal to another column.
	OrEqualsColumn(column1, column2 string) ConditionBuilder
	// EqualsSubQuery is a condition that checks if a column is equal to a subquery.
	EqualsSubQuery(column string, builder func(query Query)) ConditionBuilder
	// OrEqualsSubQuery is a condition that checks if a column is equal to a subquery.
	OrEqualsSubQuery(column string, builder func(query Query)) ConditionBuilder
	// EqualsExpr is a condition that checks if a column is equal to an expression.
	EqualsExpr(column, expr string, args ...any) ConditionBuilder
	// OrEqualsExpr is a condition that checks if a column is equal to an expression.
	OrEqualsExpr(column, expr string, args ...any) ConditionBuilder
	// NotEquals is a condition that checks if a column is not equal to a value.
	NotEquals(column string, value any) ConditionBuilder
	// OrNotEquals is a condition that checks if a column is not equal to a value.
	OrNotEquals(column string, value any) ConditionBuilder
	// NotEqualsColumn is a condition that checks if a column is not equal to another column.
	NotEqualsColumn(column1, column2 string) ConditionBuilder
	// OrNotEqualsColumn is a condition that checks if a column is not equal to another column.
	OrNotEqualsColumn(column1, column2 string) ConditionBuilder
	// NotEqualsSubQuery is a condition that checks if a column is not equal to a subquery.
	NotEqualsSubQuery(column string, builder func(query Query)) ConditionBuilder
	// OrNotEqualsSubQuery is a condition that checks if a column is not equal to a subquery.
	OrNotEqualsSubQuery(column string, builder func(query Query)) ConditionBuilder
	// NotEqualsExpr is a condition that checks if a column is not equal to an expression.
	NotEqualsExpr(column, expr string, args ...any) ConditionBuilder
	// OrNotEqualsExpr is a condition that checks if a column is not equal to an expression.
	OrNotEqualsExpr(column, expr string, args ...any) ConditionBuilder
	// GreaterThan is a condition that checks if a column is greater than a value.
	GreaterThan(column string, value any) ConditionBuilder
	// OrGreaterThan is a condition that checks if a column is greater than a value.
	OrGreaterThan(column string, value any) ConditionBuilder
	// GreaterThanColumn is a condition that checks if a column is greater than another column.
	GreaterThanColumn(column1, column2 string) ConditionBuilder
	// OrGreaterThanColumn is a condition that checks if a column is greater than another column.
	OrGreaterThanColumn(column1, column2 string) ConditionBuilder
	// GreaterThanSubQuery is a condition that checks if a column is greater than a subquery.
	GreaterThanSubQuery(column string, builder func(query Query)) ConditionBuilder
	// OrGreaterThanSubQuery is a condition that checks if a column is greater than a subquery.
	OrGreaterThanSubQuery(column string, builder func(query Query)) ConditionBuilder
	// GreaterThanExpr is a condition that checks if a column is greater than an expression.
	GreaterThanExpr(column, expr string, args ...any) ConditionBuilder
	// OrGreaterThanExpr is a condition that checks if a column is greater than an expression.
	OrGreaterThanExpr(column, expr string, args ...any) ConditionBuilder
	// GreaterThanOrEqual is a condition that checks if a column is greater than or equal to a value.
	GreaterThanOrEqual(column string, value any) ConditionBuilder
	// OrGreaterThanOrEqual is a condition that checks if a column is greater than or equal to a value.
	OrGreaterThanOrEqual(column string, value any) ConditionBuilder
	// GreaterThanOrEqualColumn is a condition that checks if a column is greater than or equal to another column.
	GreaterThanOrEqualColumn(column1, column2 string) ConditionBuilder
	// OrGreaterThanOrEqualColumn is a condition that checks if a column is greater than or equal to another column.
	OrGreaterThanOrEqualColumn(column1, column2 string) ConditionBuilder
	// GreaterThanOrEqualSubQuery is a condition that checks if a column is greater than or equal to a subquery.
	GreaterThanOrEqualSubQuery(column string, builder func(query Query)) ConditionBuilder
	// OrGreaterThanOrEqualSubQuery is a condition that checks if a column is greater than or equal to a subquery.
	OrGreaterThanOrEqualSubQuery(column string, builder func(query Query)) ConditionBuilder
	// GreaterThanOrEqualExpr is a condition that checks if a column is greater than or equal to an expression.
	GreaterThanOrEqualExpr(column, expr string, args ...any) ConditionBuilder
	// OrGreaterThanOrEqualExpr is a condition that checks if a column is greater than or equal to an expression.
	OrGreaterThanOrEqualExpr(column, expr string, args ...any) ConditionBuilder
	// LessThan is a condition that checks if a column is less than a value.
	LessThan(column string, value any) ConditionBuilder
	// OrLessThan is a condition that checks if a column is less than a value.
	OrLessThan(column string, value any) ConditionBuilder
	// LessThanColumn is a condition that checks if a column is less than another column.
	LessThanColumn(column1, column2 string) ConditionBuilder
	// OrLessThanColumn is a condition that checks if a column is less than another column.
	OrLessThanColumn(column1, column2 string) ConditionBuilder
	// LessThanSubQuery is a condition that checks if a column is less than a subquery.
	LessThanSubQuery(column string, builder func(query Query)) ConditionBuilder
	// OrLessThanSubQuery is a condition that checks if a column is less than a subquery.
	OrLessThanSubQuery(column string, builder func(query Query)) ConditionBuilder
	// LessThanExpr is a condition that checks if a column is less than an expression.
	LessThanExpr(column, expr string, args ...any) ConditionBuilder
	// OrLessThanExpr is a condition that checks if a column is less than an expression.
	OrLessThanExpr(column, expr string, args ...any) ConditionBuilder
	// LessThanOrEqual is a condition that checks if a column is less than or equal to a value.
	LessThanOrEqual(column string, value any) ConditionBuilder
	// OrLessThanOrEqual is a condition that checks if a column is less than or equal to a value.
	OrLessThanOrEqual(column string, value any) ConditionBuilder
	// LessThanOrEqualColumn is a condition that checks if a column is less than or equal to another column.
	LessThanOrEqualColumn(column1, column2 string) ConditionBuilder
	// OrLessThanOrEqualColumn is a condition that checks if a column is less than or equal to another column.
	OrLessThanOrEqualColumn(column1, column2 string) ConditionBuilder
	// LessThanOrEqualSubQuery is a condition that checks if a column is less than or equal to a subquery.
	LessThanOrEqualSubQuery(column string, builder func(query Query)) ConditionBuilder
	// OrLessThanOrEqualSubQuery is a condition that checks if a column is less than or equal to a subquery.
	OrLessThanOrEqualSubQuery(column string, builder func(query Query)) ConditionBuilder
	// LessThanOrEqualExpr is a condition that checks if a column is less than or equal to an expression.
	LessThanOrEqualExpr(column, expr string, args ...any) ConditionBuilder
	// OrLessThanOrEqualExpr is a condition that checks if a column is less than or equal to an expression.
	OrLessThanOrEqualExpr(column, expr string, args ...any) ConditionBuilder
	// Between is a condition that checks if a column is between two values.
	Between(column string, start, end any) ConditionBuilder
	// OrBetween is a condition that checks if a column is between two values.
	OrBetween(column string, start, end any) ConditionBuilder
	// BetweenExpr is a condition that checks if a column is between an expression and a value.
	BetweenExpr(column, expr string, args ...any) ConditionBuilder
	// OrBetweenExpr is a condition that checks if a column is between an expression and a value.
	OrBetweenExpr(column, expr string, args ...any) ConditionBuilder
	// NotBetween is a condition that checks if a column is not between two values.
	NotBetween(column string, start, end any) ConditionBuilder
	// OrNotBetween is a condition that checks if a column is not between two values.
	OrNotBetween(column string, start, end any) ConditionBuilder
	// NotBetweenExpr is a condition that checks if a column is not between an expression and a value.
	NotBetweenExpr(column, expr string, args ...any) ConditionBuilder
	// OrNotBetweenExpr is a condition that checks if a column is not between an expression and a value.
	OrNotBetweenExpr(column, expr string, args ...any) ConditionBuilder
	// In is a condition that checks if a column is in a list of values.
	In(column string, values any) ConditionBuilder
	// OrIn is a condition that checks if a column is in a list of values.
	OrIn(column string, values any) ConditionBuilder
	// InSubQuery is a condition that checks if a column is in a subquery.
	InSubQuery(column string, builder func(query Query)) ConditionBuilder
	// OrInSubQuery is a condition that checks if a column is in a subquery.
	OrInSubQuery(column string, builder func(query Query)) ConditionBuilder
	// InExpr is a condition that checks if a column is in an expression.
	InExpr(column, expr string, args ...any) ConditionBuilder
	// OrInExpr is a condition that checks if a column is in an expression.
	OrInExpr(column, expr string, args ...any) ConditionBuilder
	// NotIn is a condition that checks if a column is not in a list of values.
	NotIn(column string, values any) ConditionBuilder
	// OrNotIn is a condition that checks if a column is not in a list of values.
	OrNotIn(column string, values any) ConditionBuilder
	// NotInSubQuery is a condition that checks if a column is not in a subquery.
	NotInSubQuery(column string, builder func(query Query)) ConditionBuilder
	// OrNotInSubQuery is a condition that checks if a column is not in a subquery.
	OrNotInSubQuery(column string, builder func(query Query)) ConditionBuilder
	// NotInExpr is a condition that checks if a column is not in an expression.
	NotInExpr(column, expr string, args ...any) ConditionBuilder
	// OrNotInExpr is a condition that checks if a column is not in an expression.
	OrNotInExpr(column, expr string, args ...any) ConditionBuilder
	// IsNull is a condition that checks if a column is null.
	IsNull(column string) ConditionBuilder
	// OrIsNull is a condition that checks if a column is null.
	OrIsNull(column string) ConditionBuilder
	// IsNullSubQuery is a condition that checks if a column is null.
	IsNullSubQuery(builder func(query Query)) ConditionBuilder
	// OrIsNullSubQuery is a condition that checks if a column is null.
	OrIsNullSubQuery(builder func(query Query)) ConditionBuilder
	// IsNullExpr is a condition that checks if a column is null.
	IsNullExpr(expr string, args ...any) ConditionBuilder
	// OrIsNullExpr is a condition that checks if a column is null.
	OrIsNullExpr(expr string, args ...any) ConditionBuilder
	// IsNotNull is a condition that checks if a column is not null.
	IsNotNull(column string) ConditionBuilder
	// OrIsNotNull is a condition that checks if a column is not null.
	OrIsNotNull(column string) ConditionBuilder
	// IsNotNullSubQuery is a condition that checks if a column is not null.
	IsNotNullSubQuery(builder func(query Query)) ConditionBuilder
	// OrIsNotNullSubQuery is a condition that checks if a column is not null.
	OrIsNotNullSubQuery(builder func(query Query)) ConditionBuilder
	// IsNotNullExpr is a condition that checks if a column is not null.
	IsNotNullExpr(expr string, args ...any) ConditionBuilder
	// OrIsNotNullExpr is a condition that checks if a column is not null.
	OrIsNotNullExpr(expr string, args ...any) ConditionBuilder
	// IsTrue is a condition that checks if a column is true.
	IsTrue(column string) ConditionBuilder
	// OrIsTrue is a condition that checks if a column is true.
	OrIsTrue(column string) ConditionBuilder
	// IsTrueSubQuery is a condition that checks if a column is true.
	IsTrueSubQuery(builder func(query Query)) ConditionBuilder
	// OrIsTrueSubQuery is a condition that checks if a column is true.
	OrIsTrueSubQuery(builder func(query Query)) ConditionBuilder
	// IsTrueExpr is a condition that checks if a column is true.
	IsTrueExpr(expr string, args ...any) ConditionBuilder
	// OrIsTrueExpr is a condition that checks if a column is true.
	OrIsTrueExpr(expr string, args ...any) ConditionBuilder
	// IsFalse is a condition that checks if a column is false.
	IsFalse(column string) ConditionBuilder
	// OrIsFalse is a condition that checks if a column is false.
	OrIsFalse(column string) ConditionBuilder
	// IsFalseSubQuery is a condition that checks if a column is false.
	IsFalseSubQuery(builder func(query Query)) ConditionBuilder
	// OrIsFalseSubQuery is a condition that checks if a column is false.
	OrIsFalseSubQuery(builder func(query Query)) ConditionBuilder
	// IsFalseExpr is a condition that checks if a column is false.
	IsFalseExpr(expr string, args ...any) ConditionBuilder
	// OrIsFalseExpr is a condition that checks if a column is false.
	OrIsFalseExpr(expr string, args ...any) ConditionBuilder
	// Contains is a condition that checks if a column contains a value.
	Contains(column, value string) ConditionBuilder
	// OrContains is a condition that checks if a column contains a value.
	OrContains(column, value string) ConditionBuilder
	// ContainsAny is a condition that checks if a column contains any of the values.
	ContainsAny(column string, values []string) ConditionBuilder
	// OrContainsAny is a condition that checks if a column contains any of the values.
	OrContainsAny(column string, values []string) ConditionBuilder
	// ContainsIgnoreCase is a condition that checks if a column contains a value, ignoring case.
	ContainsIgnoreCase(column, value string) ConditionBuilder
	// OrContainsIgnoreCase is a condition that checks if a column contains a value, ignoring case.
	OrContainsIgnoreCase(column, value string) ConditionBuilder
	// ContainsAnyIgnoreCase is a condition that checks if a column contains any of the values, ignoring case.
	ContainsAnyIgnoreCase(column string, values []string) ConditionBuilder
	// OrContainsAnyIgnoreCase is a condition that checks if a column contains any of the values, ignoring case.
	OrContainsAnyIgnoreCase(column string, values []string) ConditionBuilder
	// NotContains is a condition that checks if a column does not contain a value.
	NotContains(column, value string) ConditionBuilder
	// OrNotContains is a condition that checks if a column does not contain a value.
	OrNotContains(column, value string) ConditionBuilder
	// NotContainsAny is a condition that checks if a column does not contain any of the values.
	NotContainsAny(column string, values []string) ConditionBuilder
	// OrNotContainsAny is a condition that checks if a column does not contain any of the values.
	OrNotContainsAny(column string, values []string) ConditionBuilder
	// NotContainsIgnoreCase is a condition that checks if a column does not contain a value, ignoring case.
	NotContainsIgnoreCase(column, value string) ConditionBuilder
	// OrNotContainsIgnoreCase is a condition that checks if a column does not contain a value, ignoring case.
	OrNotContainsIgnoreCase(column, value string) ConditionBuilder
	// NotContainsAnyIgnoreCase is a condition that checks if a column does not contain any of the values, ignoring case.
	NotContainsAnyIgnoreCase(column string, values []string) ConditionBuilder
	// OrNotContainsAnyIgnoreCase is a condition that checks if a column does not contain any of the values, ignoring case.
	OrNotContainsAnyIgnoreCase(column string, values []string) ConditionBuilder
	// StartsWith is a condition that checks if a column starts with a value.
	StartsWith(column, value string) ConditionBuilder
	// OrStartsWith is a condition that checks if a column starts with a value.
	OrStartsWith(column, value string) ConditionBuilder
	// StartsWithAny is a condition that checks if a column starts with any of the values.
	StartsWithAny(column string, values []string) ConditionBuilder
	// OrStartsWithAny is a condition that checks if a column starts with any of the values.
	OrStartsWithAny(column string, values []string) ConditionBuilder
	// StartsWithIgnoreCase is a condition that checks if a column starts with a value, ignoring case.
	StartsWithIgnoreCase(column, value string) ConditionBuilder
	// OrStartsWithIgnoreCase is a condition that checks if a column starts with a value, ignoring case.
	OrStartsWithIgnoreCase(column, value string) ConditionBuilder
	// StartsWithAnyIgnoreCase is a condition that checks if a column starts with any of the values, ignoring case.
	StartsWithAnyIgnoreCase(column string, values []string) ConditionBuilder
	// OrStartsWithAnyIgnoreCase is a condition that checks if a column starts with any of the values, ignoring case.
	OrStartsWithAnyIgnoreCase(column string, values []string) ConditionBuilder
	// NotStartsWith is a condition that checks if a column does not start with a value.
	NotStartsWith(column, value string) ConditionBuilder
	// OrNotStartsWith is a condition that checks if a column does not start with a value.
	OrNotStartsWith(column, value string) ConditionBuilder
	// NotStartsWithAny is a condition that checks if a column does not start with any of the values.
	NotStartsWithAny(column string, values []string) ConditionBuilder
	// OrNotStartsWithAny is a condition that checks if a column does not start with any of the values.
	OrNotStartsWithAny(column string, values []string) ConditionBuilder
	// NotStartsWithIgnoreCase is a condition that checks if a column does not start with a value, ignoring case.
	NotStartsWithIgnoreCase(column, value string) ConditionBuilder
	// OrNotStartsWithIgnoreCase is a condition that checks if a column does not start with a value, ignoring case.
	OrNotStartsWithIgnoreCase(column, value string) ConditionBuilder
	// NotStartsWithAnyIgnoreCase is a condition that checks if a column does not start with any of the values, ignoring case.
	NotStartsWithAnyIgnoreCase(column string, values []string) ConditionBuilder
	// OrNotStartsWithAnyIgnoreCase is a condition that checks if a column does not start with any of the values, ignoring case.
	OrNotStartsWithAnyIgnoreCase(column string, values []string) ConditionBuilder
	// EndsWith is a condition that checks if a column ends with a value.
	EndsWith(column, value string) ConditionBuilder
	// OrEndsWith is a condition that checks if a column ends with a value.
	OrEndsWith(column, value string) ConditionBuilder
	// EndsWithAny is a condition that checks if a column ends with any of the values.
	EndsWithAny(column string, values []string) ConditionBuilder
	// OrEndsWithAny is a condition that checks if a column ends with any of the values.
	OrEndsWithAny(column string, values []string) ConditionBuilder
	// EndsWithIgnoreCase is a condition that checks if a column ends with a value, ignoring case.
	EndsWithIgnoreCase(column, value string) ConditionBuilder
	// OrEndsWithIgnoreCase is a condition that checks if a column ends with a value, ignoring case.
	OrEndsWithIgnoreCase(column, value string) ConditionBuilder
	// EndsWithAnyIgnoreCase is a condition that checks if a column ends with any of the values, ignoring case.
	EndsWithAnyIgnoreCase(column string, values []string) ConditionBuilder
	// OrEndsWithAnyIgnoreCase is a condition that checks if a column ends with any of the values, ignoring case.
	OrEndsWithAnyIgnoreCase(column string, values []string) ConditionBuilder
	// NotEndsWith is a condition that checks if a column does not end with a value.
	NotEndsWith(column, value string) ConditionBuilder
	// OrNotEndsWith is a condition that checks if a column does not end with a value.
	OrNotEndsWith(column, value string) ConditionBuilder
	// NotEndsWithAny is a condition that checks if a column does not end with any of the values.
	NotEndsWithAny(column string, values []string) ConditionBuilder
	// OrNotEndsWithAny is a condition that checks if a column does not end with any of the values.
	OrNotEndsWithAny(column string, values []string) ConditionBuilder
	// NotEndsWithIgnoreCase is a condition that checks if a column does not end with a value, ignoring case.
	NotEndsWithIgnoreCase(column, value string) ConditionBuilder
	// OrNotEndsWithIgnoreCase is a condition that checks if a column does not end with a value, ignoring case.
	OrNotEndsWithIgnoreCase(column, value string) ConditionBuilder
	// NotEndsWithAnyIgnoreCase is a condition that checks if a column does not end with any of the values, ignoring case.
	NotEndsWithAnyIgnoreCase(column string, values []string) ConditionBuilder
	// OrNotEndsWithAnyIgnoreCase is a condition that checks if a column does not end with any of the values, ignoring case.
	OrNotEndsWithAnyIgnoreCase(column string, values []string) ConditionBuilder
	// Expr is a condition that checks if an expression is true.
	Expr(expr string, args ...any) ConditionBuilder
	// OrExpr is a condition that checks if an expression is true.
	OrExpr(expr string, args ...any) ConditionBuilder
	// Group is a condition that checks if a group of conditions are true.
	Group(builder func(cb ConditionBuilder)) ConditionBuilder
	// OrGroup is a condition that checks if a group of conditions are true.
	OrGroup(builder func(cb ConditionBuilder)) ConditionBuilder
}

// Executor is an interface that defines the methods for executing a query.
type Executor interface {
	// Exec executes a query and returns the result.
	Exec(ctx context.Context, dest ...any) (sql.Result, error)
	// Scan scans the result into a slice of any type.
	Scan(ctx context.Context, dest ...any) error
}

// QueryExecutor is an interface that defines the methods for executing a query.
type QueryExecutor interface {
	Executor
	// Rows returns the result as a sql.Rows.
	Rows(ctx context.Context) (*sql.Rows, error)
	// ScanAndCount scans the result into a slice of any type and returns the count of the result.
	ScanAndCount(ctx context.Context, dest ...any) (int64, error)
	// Count returns the count of the result.
	Count(ctx context.Context) (int64, error)
	// Exists returns true if the result exists.
	Exists(ctx context.Context) (bool, error)
}

// CTE is an interface that defines the methods for creating a common table expression.
type CTE[T Executor] interface {
	// With creates a common table expression.
	With(name string, builder func(query Query)) T
	// WithValues creates a common table expression with values.
	WithValues(name string, model any, withOrder ...bool) T
	// WithRecursive creates a recursive common table expression.
	WithRecursive(name string, builder func(query Query)) T
}

// Selectable is an interface that defines the methods for selecting columns.
type Selectable[T Executor] interface {
	// SelectAll selects all columns.
	SelectAll() T
	// Select selects specific columns.
	Select(columns ...string) T
	// Exclude excludes specific columns.
	Exclude(columns ...string) T
	// ExcludeAll excludes all columns.
	ExcludeAll() T
}

// Source is an interface that defines the methods for selecting a table.
type Source[T Executor] interface {
	// Model selects a model.
	Model(model any) T
	// Table selects a table.
	Table(name string) T
	// TableAs selects a table with an alias.
	TableAs(name, alias string) T
	// TableExpr selects a table with an expression.
	TableExpr(expr string, args ...any) T
	// TableExprAs selects a table with an expression and an alias.
	TableExprAs(expr string, alias string, args ...any) T
	// TableSubQuery selects a subquery.
	TableSubQuery(builder func(query Query)) T
	// TableSubQueryAs selects a subquery with an alias.
	TableSubQueryAs(builder func(query Query), alias string) T
}

// Joinable is an interface that defines the methods for joining a table.
type Joinable[T Executor] interface {
	// Join joins a table.
	Join(model any, builder func(cb ConditionBuilder)) T
	// JoinAs joins a table with an alias.
	JoinAs(model any, alias string, builder func(cb ConditionBuilder)) T
	// JoinTable joins a table.
	JoinTable(name string, builder func(cb ConditionBuilder)) T
	// JoinTableAs joins a table with an alias.
	JoinTableAs(name, alias string, builder func(cb ConditionBuilder)) T
	// JoinSubQuery joins a subquery.
	JoinSubQuery(builder func(query Query), conditionBuilder func(cb ConditionBuilder)) T
	// JoinSubQueryAs joins a subquery with an alias.
	JoinSubQueryAs(builder func(query Query), alias string, conditionBuilder func(cb ConditionBuilder)) T
	// JoinExpr joins an expression.
	JoinExpr(expr string, builder func(cb ConditionBuilder), args ...any) T
	// JoinExprAs joins an expression with an alias.
	JoinExprAs(expr, alias string, builder func(cb ConditionBuilder), args ...any) T
}

// Filterable is an interface that defines the methods for filtering a query.
type Filterable[T Executor] interface {
	// Where adds a where clause to the query.
	Where(builder func(cb ConditionBuilder)) T
	// WherePK adds a where clause to the query using the primary key.
	WherePK(columns ...string) T
	// WhereDeleted adds a where clause to the query using the deleted column.
	WhereDeleted() T
	// WhereAllWithDeleted adds a where clause to the query using the all with deleted column.
	WhereAllWithDeleted() T
}

// Orderable is an interface that defines the methods for ordering a query.
type Orderable[T Executor] interface {
	// OrderBy orders the query by a column.
	OrderBy(columns ...string) T
	// OrderByNullsFirst orders the query by a column, nulls first.
	OrderByNullsFirst(columns ...string) T
	// OrderByNullsLast orders the query by a column, nulls last.
	OrderByNullsLast(columns ...string) T
	// OrderByDesc orders the query by a column in descending order.
	OrderByDesc(columns ...string) T
	// OrderByDescNullsFirst orders the query by a column in descending order, nulls first.
	OrderByDescNullsFirst(columns ...string) T
	// OrderByDescNullsLast orders the query by a column in descending order, nulls last.
	OrderByDescNullsLast(columns ...string) T
	// OrderByExpr orders the query by an expression.
	OrderByExpr(expr string, args ...any) T
}

// Limitable is an interface that defines the methods for limiting a query.
type Limitable[T Executor] interface {
	// Limit limits the number of rows returned by the query.
	Limit(limit int) T
}

// ColumnSettable is an interface that defines the methods for setting a column.
type ColumnSettable[T Executor] interface {
	// Column sets a column to a value.
	Column(name string, value any) T
	// ColumnExpr sets a column to an expression.
	ColumnExpr(name, expr string, args ...any) T
}

// Returnable is an interface that defines the methods for returning a query.
type Returnable[T Executor] interface {
	// Returning returns the query with the specified columns.
	Returning(columns ...string) T
	// ReturningAll returns the query with all columns.
	ReturningAll() T
	// ReturningNull returns the query with null columns.
	ReturningNull() T
}

// ApplyFunc is a function that applies a shared operation.
type ApplyFunc[T any] func(T) T

// Applier is an interface that defines the methods for applying shared operations.
type Applier[T any] interface {
	// Apply applies shared operations.
	Apply(fns ...ApplyFunc[T]) T
}

// Query is an interface that defines the methods for querying a database.
type Query interface {
	QueryExecutor
	CTE[Query]
	Selectable[Query]
	Source[Query]
	Joinable[Query]
	Filterable[Query]
	Orderable[Query]
	Limitable[Query]
	Applier[Query]
	// SelectAs selects a column with an alias.
	SelectAs(column, alias string) Query
	// SelectModelColumns selects the columns of a model.
	// By default, all columns of the Model are selected if the select-related methods is not called.
	SelectModelColumns() Query
	// SelectModelPKs selects the primary keys of a model.
	SelectModelPKs() Query
	// SelectExpr selects a column with an expression.
	SelectExpr(expr string, args ...any) Query
	// SelectExprAs selects a column with an expression and an alias.
	SelectExprAs(expr string, alias string, args ...any) Query
	// Distinct returns a distinct query.
	Distinct() Query
	// DistinctOn returns a distinct query on an expression.
	DistinctOn(expr string, args ...any) Query
	// LeftJoin joins a table.
	LeftJoin(model any, builder func(cb ConditionBuilder)) Query
	// LeftJoinAs joins a table with an alias.
	LeftJoinAs(model any, alias string, builder func(cb ConditionBuilder)) Query
	// LeftJoinTable joins a table.
	LeftJoinTable(name string, builder func(cb ConditionBuilder)) Query
	// LeftJoinTableAs joins a table with an alias.
	LeftJoinTableAs(name, alias string, builder func(cb ConditionBuilder)) Query
	// LeftJoinSubQuery joins a subquery.
	LeftJoinSubQuery(builder func(query Query), conditionBuilder func(cb ConditionBuilder)) Query
	// LeftJoinSubQueryAs joins a subquery with an alias.
	LeftJoinSubQueryAs(builder func(query Query), alias string, conditionBuilder func(cb ConditionBuilder)) Query
	// LeftJoinExpr joins an expression.
	LeftJoinExpr(expr string, builder func(cb ConditionBuilder), args ...any) Query
	// LeftJoinExprAs joins an expression with an alias.
	LeftJoinExprAs(expr, alias string, builder func(cb ConditionBuilder), args ...any) Query
	// RightJoin joins a table.
	RightJoin(model any, builder func(cb ConditionBuilder)) Query
	RightJoinAs(model any, alias string, builder func(cb ConditionBuilder)) Query
	// RightJoinTable joins a table.
	RightJoinTable(name string, builder func(cb ConditionBuilder)) Query
	// RightJoinTableAs joins a table with an alias.
	RightJoinTableAs(name, alias string, builder func(cb ConditionBuilder)) Query
	// RightJoinSubQuery joins a subquery.
	RightJoinSubQuery(builder func(query Query), conditionBuilder func(cb ConditionBuilder)) Query
	// RightJoinSubQueryAs joins a subquery with an alias.
	RightJoinSubQueryAs(builder func(query Query), alias string, conditionBuilder func(cb ConditionBuilder)) Query
	// RightJoinExpr joins an expression.
	RightJoinExpr(expr string, builder func(cb ConditionBuilder), args ...any) Query
	// RightJoinExprAs joins an expression with an alias.
	RightJoinExprAs(expr, alias string, builder func(cb ConditionBuilder), args ...any) Query
	// ModelRelation joins a model relation.
	ModelRelation(relation ...ModelRelation) Query
	// Relation joins a relation.
	Relation(name string, apply ...func(query Query)) Query
	// GroupBy groups the query by a column.
	GroupBy(columns ...string) Query
	// GroupByExpr groups the query by an expression.
	GroupByExpr(expr string, args ...any) Query
	// Having adds a having clause to the query.
	Having(builder func(cb ConditionBuilder)) Query
	// Offset adds an offset to the query.
	Offset(offset int) Query
	// Paginate paginates the query.
	Paginate(pageable mo.Pageable, defaultAlias ...string) Query
	// ForShare adds a for share lock to the query.
	ForShare(tables ...string) Query
	// ForShareNoWait adds a for share no wait lock to the query.
	ForShareNoWait(tables ...string) Query
	// ForShareSkipLocked adds a for share skip locked lock to the query.
	ForShareSkipLocked(tables ...string) Query
	// ForUpdate adds a for update lock to the query.
	ForUpdate(tables ...string) Query
	// ForUpdateNoWait adds a for update no wait lock to the query.
	ForUpdateNoWait(tables ...string) Query
	// ForUpdateSkipLocked adds a for update skip locked lock to the query.
	ForUpdateSkipLocked(tables ...string) Query
	// Union joins a query.
	Union(builder func(query Query)) Query
	// UnionAll joins a query.
	UnionAll(builder func(query Query)) Query
	// Intersect joins a query.
	Intersect(builder func(query Query)) Query
	// IntersectAll joins a query.
	IntersectAll(builder func(query Query)) Query
	// Except joins a query.
	Except(builder func(query Query)) Query
	// ExceptAll joins a query.
	ExceptAll(builder func(query Query)) Query
}

// Create is an interface that defines the methods for creating a query.
type Create interface {
	Executor
	CTE[Create]
	Source[Create]
	Selectable[Create]
	ColumnSettable[Create]
	Returnable[Create]
	Applier[Create]
	// OnConflict adds an on conflict clause to the query.
	OnConflict(columns ...string) Create
	// OnConflictConstraint adds an on conflict constraint to the query.
	OnConflictConstraint(constraint string) Create
	// OnConflictDoUpdate adds an on conflict do update clause to the query.
	OnConflictDoUpdate() Create
	// OnConflictDoNothing adds an on conflict do nothing clause to the query.
	OnConflictDoNothing() Create
	// Set sets a column to a value on conflict.
	Set(name string, value ...any) Create
	// SetExpr sets a column to an expression on conflict.
	SetExpr(name, expr string, args ...any) Create
	// Where adds a where clause to the query.
	Where(builder func(cb ConditionBuilder)) Create
}

// Update is an interface that defines the methods for updating a query.
type Update interface {
	Executor
	CTE[Update]
	Source[Update]
	Joinable[Update]
	Selectable[Update]
	Filterable[Update]
	Orderable[Update]
	Limitable[Update]
	ColumnSettable[Update]
	Returnable[Update]
	Applier[Update]
	// Set sets a column to a value.
	Set(name string, value any) Update
	// SetExpr sets a column to an expression.
	SetExpr(name, expr string, args ...any) Update
	// OmitZero adds an omit zero clause to the query.
	OmitZero() Update
	// Bulk adds a bulk clause to the query.
	Bulk() Update
}

// Delete is an interface that defines the methods for deleting a query.
type Delete interface {
	Executor
	CTE[Delete]
	Source[Delete]
	Filterable[Delete]
	Orderable[Delete]
	Limitable[Delete]
	Returnable[Delete]
	Applier[Delete]
	// ForceDelete adds a force delete clause to the query.
	ForceDelete() Delete
}

// RawQuery is an interface that defines the methods for a raw SQL query.
type RawQuery interface {
	Executor
}

// Db is an interface that defines the methods for a database.
type Db interface {
	// NewQuery creates a new query.
	NewQuery() Query
	// NewCreate creates a new create.
	NewCreate() Create
	// NewUpdate creates a new update.
	NewUpdate() Update
	// NewDelete creates a new delete.
	NewDelete() Delete
	// NewRawQuery creates a new raw query.
	NewRawQuery(query string, args ...any) RawQuery
	// RunInTx runs a transaction.
	RunInTx(ctx context.Context, fn func(ctx context.Context, tx Db) error) error
	// RunInReadOnlyTx runs a read only transaction.
	RunInReadOnlyTx(ctx context.Context, fn func(ctx context.Context, tx Db) error) error
	// WithNamedArg returns a new Db with the named arg.
	WithNamedArg(name string, value any) Db
	// ModelPKs returns the primary keys of a model.
	ModelPKs(model any) map[string]any
	// Schema returns the schema of a table.
	Schema(model any) *schema.Table
}

// AutoColumn is an interface that can be implemented by a struct to automatically create or update a column.
type AutoColumn interface {
	// Name returns the name of the column.
	Name() string
}

// CreateAutoColumn is an interface that can be implemented by a struct to automatically create a column.
type CreateAutoColumn interface {
	AutoColumn
	// OnCreate is called when the column is created.
	OnCreate(query *bun.InsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value)
}

// UpdateAutoColumn is an interface that can be implemented by a struct to automatically update a column.
type UpdateAutoColumn interface {
	CreateAutoColumn
	// OnUpdate is called when the column is updated.
	OnUpdate(query *bun.UpdateQuery, hasSet bool, table *schema.Table, field *schema.Field, model any, value reflect.Value)
}
