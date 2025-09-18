package search

import (
	"reflect"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/mo"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/reflectx"
	"github.com/ilxqx/vef-framework-go/utils"
	"github.com/spf13/cast"

	"github.com/samber/lo"
)

var (
	logger    = log.Named("search")              // logger is the logger for the search package
	rangeType = reflect.TypeFor[mo.Range[int]]() // rangeType is the type for range values
)

// Search contains multiple conditions.
type Search struct {
	conditions []Condition // conditions contains all search conditions
}

// Condition is a condition item.
type Condition struct {
	Index    []int             // Index is the field index in the struct
	Alias    string            // Alias is the table alias for the condition
	Columns  []string          // Columns are the column names for the condition
	Operator Operator          // Operator is the comparison operator
	Args     map[string]string // Args contains additional arguments for the condition
}

// Apply applies the search to the condition builder.
func (f Search) Apply(cb orm.ConditionBuilder, target any, defaultAlias ...string) {
	value := reflect.Indirect(reflect.ValueOf(target))
	if value.Kind() != reflect.Struct {
		logger.Warnf("Invalid target type, expected struct, got %s", value.Type().Name())
		return
	}

	for _, c := range f.conditions {
		field := value.FieldByIndex(c.Index)
		if field.IsZero() && (!field.CanInt() && !field.CanUint() && !field.CanFloat() && field.Kind() != reflect.Bool) {
			// Skip non-numeric and non-boolean zero value.
			continue
		}

		alias := getColumnAlias(c.Alias, defaultAlias...)
		columns := make([]string, 0, len(c.Columns))
		for _, column := range c.Columns {
			columns = append(columns, utils.ColumnWithAlias(column, alias))
		}
		applyCondition(cb, c, columns, field)
	}
}

// getColumnAlias gets the alias of the column.
func getColumnAlias(alias string, defaultAlias ...string) string {
	if alias == constants.Empty {
		if len(defaultAlias) > 0 {
			return defaultAlias[0]
		}

		return constants.Empty
	}

	return alias
}

// applyCondition applies the condition to the query.
func applyCondition(cb orm.ConditionBuilder, c Condition, columns []string, value reflect.Value) {
	// Handle different conditions based on the operator.
	switch c.Operator {
	case Equals, NotEquals, GreaterThan, GreaterThanOrEqual, LessThan, LessThanOrEqual:
		applyComparisonCondition(cb, columns[0], c.Operator, value)
	case Between, NotBetween:
		applyBetweenCondition(cb, columns[0], c.Operator, value, c.Args)
	case In, NotIn:
		applyInCondition(cb, columns[0], value, c.Operator, c.Args)
	case IsNull, IsNotNull:
		applyNullCondition(cb, columns[0], value, c.Operator)
	case Contains, NotContains, StartsWith, NotStartsWith, EndsWith, NotEndsWith,
		ContainsIgnoreCase, NotContainsIgnoreCase, StartsWithIgnoreCase, NotStartsWithIgnoreCase,
		EndsWithIgnoreCase, NotEndsWithIgnoreCase:
		applyLikeCondition(cb, columns, value, c.Operator)
	}
}

// applyComparisonCondition applies the comparison operator to the condition builder.
func applyComparisonCondition(cb orm.ConditionBuilder, column string, operator Operator, value reflect.Value) {
	switch operator {
	case Equals:
		cb.Equals(column, value.Interface())
	case NotEquals:
		cb.NotEquals(column, value.Interface())
	case GreaterThan:
		cb.GreaterThan(column, value.Interface())
	case GreaterThanOrEqual:
		cb.GreaterThanOrEqual(column, value.Interface())
	case LessThan:
		cb.LessThan(column, value.Interface())
	case LessThanOrEqual:
		cb.LessThanOrEqual(column, value.Interface())
	}
}

// applyBetweenCondition applies the between operator to the condition builder.
func applyBetweenCondition(cb orm.ConditionBuilder, column string, operator Operator, value reflect.Value, conditionArgs map[string]string) {
	start, end, ok := getRangeValue(value, conditionArgs)
	if !ok {
		return
	}

	switch operator {
	case Between:
		cb.Between(column, start, end)
	case NotBetween:
		cb.NotBetween(column, start, end)
	}
}

// applyInCondition applies the in operator to the condition builder.
func applyInCondition(cb orm.ConditionBuilder, column string, value reflect.Value, operator Operator, conditionArgs map[string]string) {
	values := reflectx.ApplyIfString(
		value,
		func(s string) []any {
			var values []any
			delimiter := lo.CoalesceOrEmpty(conditionArgs[ArgDelimiter], constants.Comma)

			switch conditionArgs[ArgType] {
			case constants.TypeInt:
				for value := range strings.SplitSeq(s, delimiter) {
					values = append(values, cast.ToInt(value))
				}
			default:
				for value := range strings.SplitSeq(s, delimiter) {
					values = append(values, value)
				}
			}

			return values
		},
	)

	if len(values) == 0 {
		return
	}

	switch operator {
	case In:
		cb.In(column, values)
	case NotIn:
		cb.NotIn(column, values)
	}
}

// applyNullCondition applies the null operator to the condition builder.
// It checks if the value is a boolean and true, then applies IsNull or IsNotNull condition.
func applyNullCondition(cb orm.ConditionBuilder, column string, value reflect.Value, operator Operator) {
	value = reflect.Indirect(value)
	if value.Kind() != reflect.Bool || !value.Bool() {
		return
	}

	switch operator {
	case IsNull:
		cb.IsNull(column)
	case IsNotNull:
		cb.IsNotNull(column)
	}
}

// applyLikeCondition applies the like operator to the condition builder.
func applyLikeCondition(cb orm.ConditionBuilder, columns []string, value reflect.Value, operator Operator) {
	val := reflectx.ApplyIfString(value, func(s string) string { return s })

	if len(columns) > 1 {
		applyMultiColumnLikeCondition(cb, columns, val, operator)
		return
	}

	applySingleColumnLikeCondition(cb, columns[0], val, operator)
}

// applyMultiColumnLikeCondition applies like condition for multiple columns with OR logic.
func applyMultiColumnLikeCondition(cb orm.ConditionBuilder, columns []string, val string, operator Operator) {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, col := range columns {
			applyLikeOperation(cb, col, val, operator, true)
		}
	})
}

// applySingleColumnLikeCondition applies like condition for a single column.
func applySingleColumnLikeCondition(cb orm.ConditionBuilder, column, val string, operator Operator) {
	applyLikeOperation(cb, column, val, operator, false)
}

// applyLikeOperation applies the specific like operation on a column.
func applyLikeOperation(cb orm.ConditionBuilder, column, val string, operator Operator, useOr bool) {
	switch operator {
	case Contains:
		if useOr {
			cb.OrContains(column, val)
		} else {
			cb.Contains(column, val)
		}
	case ContainsIgnoreCase:
		if useOr {
			cb.OrContainsIgnoreCase(column, val)
		} else {
			cb.ContainsIgnoreCase(column, val)
		}
	case NotContains:
		if useOr {
			cb.OrNotContains(column, val)
		} else {
			cb.NotContains(column, val)
		}
	case NotContainsIgnoreCase:
		if useOr {
			cb.OrNotContainsIgnoreCase(column, val)
		} else {
			cb.NotContainsIgnoreCase(column, val)
		}
	case StartsWith:
		if useOr {
			cb.OrStartsWith(column, val)
		} else {
			cb.StartsWith(column, val)
		}
	case StartsWithIgnoreCase:
		if useOr {
			cb.OrStartsWithIgnoreCase(column, val)
		} else {
			cb.StartsWithIgnoreCase(column, val)
		}
	case NotStartsWith:
		if useOr {
			cb.OrNotStartsWith(column, val)
		} else {
			cb.NotStartsWith(column, val)
		}
	case NotStartsWithIgnoreCase:
		if useOr {
			cb.OrNotStartsWithIgnoreCase(column, val)
		} else {
			cb.NotStartsWithIgnoreCase(column, val)
		}
	case EndsWith:
		if useOr {
			cb.OrEndsWith(column, val)
		} else {
			cb.EndsWith(column, val)
		}
	case EndsWithIgnoreCase:
		if useOr {
			cb.OrEndsWithIgnoreCase(column, val)
		} else {
			cb.EndsWithIgnoreCase(column, val)
		}
	case NotEndsWith:
		if useOr {
			cb.OrNotEndsWith(column, val)
		} else {
			cb.NotEndsWith(column, val)
		}
	case NotEndsWithIgnoreCase:
		if useOr {
			cb.OrNotEndsWithIgnoreCase(column, val)
		} else {
			cb.NotEndsWithIgnoreCase(column, val)
		}
	}
}
