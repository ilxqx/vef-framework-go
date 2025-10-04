package search

import (
	"reflect"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/dbhelpers"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/monad"
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/spf13/cast"

	"github.com/samber/lo"
)

var (
	logger    = log.Named("search")
	rangeType = reflect.TypeFor[monad.Range[int]]()
)

// Search contains multiple conditions.
type Search struct {
	conditions []Condition
}

// Condition is a condition item.
type Condition struct {
	// Index is the field index in the struct
	Index []int
	// Alias is the table alias for the condition
	Alias string
	// Columns are the column names for the condition
	Columns []string
	// Operator is the comparison operator
	Operator Operator
	// Params contains additional parameters for the condition
	Params map[string]string
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
		if field.Kind() == reflect.Pointer && field.IsNil() {
			continue
		}

		fieldValue := field.Interface()
		switch nv := fieldValue.(type) {
		case null.String:
			if !nv.Valid {
				continue
			}
			fieldValue = nv.ValueOrZero()
		case null.Int:
			if !nv.Valid {
				continue
			}
			fieldValue = nv.ValueOrZero()
		case null.Int16:
			if !nv.Valid {
				continue
			}
			fieldValue = nv.ValueOrZero()
		case null.Int32:
			if !nv.Valid {
				continue
			}
			fieldValue = nv.ValueOrZero()
		case null.Float:
			if !nv.Valid {
				continue
			}
			fieldValue = nv.ValueOrZero()
		case null.Bool:
			if !nv.Valid {
				continue
			}
			fieldValue = nv.ValueOrZero()
		case null.Byte:
			if !nv.Valid {
				continue
			}
			fieldValue = nv.ValueOrZero()
		case null.DateTime:
			if !nv.Valid {
				continue
			}
			fieldValue = nv.ValueOrZero()
		case null.Date:
			if !nv.Valid {
				continue
			}
			fieldValue = nv.ValueOrZero()
		case null.Time:
			if !nv.Valid {
				continue
			}
			fieldValue = nv.ValueOrZero()
		case null.Decimal:
			if !nv.Valid {
				continue
			}
			fieldValue = nv.ValueOrZero()
		}

		alias := getColumnAlias(c.Alias, defaultAlias...)
		columns := make([]string, len(c.Columns))
		for i, column := range c.Columns {
			columns[i] = dbhelpers.ColumnWithAlias(column, alias)
		}
		applyCondition(cb, c, columns, fieldValue)
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
func applyCondition(cb orm.ConditionBuilder, c Condition, columns []string, value any) {
	// Handle different conditions based on the operator.
	switch c.Operator {
	case Equals, NotEquals, GreaterThan, GreaterThanOrEqual, LessThan, LessThanOrEqual:
		applyComparisonCondition(cb, columns[0], c.Operator, value)
	case Between, NotBetween:
		applyBetweenCondition(cb, columns[0], c.Operator, value, c.Params)
	case In, NotIn:
		applyInCondition(cb, columns[0], value, c.Operator, c.Params)
	case IsNull, IsNotNull:
		applyNullCondition(cb, columns[0], value, c.Operator)
	case Contains, NotContains, StartsWith, NotStartsWith, EndsWith, NotEndsWith,
		ContainsIgnoreCase, NotContainsIgnoreCase, StartsWithIgnoreCase, NotStartsWithIgnoreCase,
		EndsWithIgnoreCase, NotEndsWithIgnoreCase:
		applyLikeCondition(cb, columns, value, c.Operator)
	}
}

// applyComparisonCondition applies the comparison operator to the condition builder.
func applyComparisonCondition(cb orm.ConditionBuilder, column string, operator Operator, value any) {
	switch operator {
	case Equals:
		cb.Equals(column, value)
	case NotEquals:
		cb.NotEquals(column, value)
	case GreaterThan:
		cb.GreaterThan(column, value)
	case GreaterThanOrEqual:
		cb.GreaterThanOrEqual(column, value)
	case LessThan:
		cb.LessThan(column, value)
	case LessThanOrEqual:
		cb.LessThanOrEqual(column, value)
	}
}

// applyBetweenCondition applies the between operator to the condition builder.
func applyBetweenCondition(cb orm.ConditionBuilder, column string, operator Operator, value any, conditionParams map[string]string) {
	start, end, ok := getRangeValue(value, conditionParams)
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
func applyInCondition(cb orm.ConditionBuilder, column string, fieldValue any, operator Operator, conditionParams map[string]string) {
	var values []any

	// Handle string types first
	switch v := fieldValue.(type) {
	case string:
		values = parseStringInCondition(v, conditionParams)
	case *string:
		values = parseStringInCondition(*v, conditionParams)
	}

	// Handle slice types
	rv := reflect.Indirect(reflect.ValueOf(fieldValue))
	if rv.Kind() == reflect.Slice {
		for i := range rv.Len() {
			values = append(values, rv.Index(i).Interface())
		}
	}

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

// parseStringInCondition parses the string in condition.
func parseStringInCondition(slice string, conditionParams map[string]string) []any {
	var values []any
	if slice == constants.Empty {
		return values
	}

	delimiter := lo.CoalesceOrEmpty(conditionParams[ParamDelimiter], constants.Comma)
	for value := range strings.SplitSeq(slice, delimiter) {
		switch conditionParams[ParamType] {
		case constants.TypeInt:
			values = append(values, cast.ToInt(value))
		default:
			values = append(values, value)
		}
	}
	return values
}

// applyNullCondition applies the null operator to the condition builder.
// It checks if the value is a boolean and true, then applies IsNull or IsNotNull condition.
func applyNullCondition(cb orm.ConditionBuilder, column string, fieldValue any, operator Operator) {
	var shouldApply bool
	switch value := fieldValue.(type) {
	case bool:
		shouldApply = value
	case *bool:
		shouldApply = *value
	}

	switch operator {
	case IsNull:
		cb.ApplyIf(shouldApply, func(cb orm.ConditionBuilder) {
			cb.IsNull(column)
		})
	case IsNotNull:
		cb.ApplyIf(shouldApply, func(cb orm.ConditionBuilder) {
			cb.IsNotNull(column)
		})
	}
}

// applyLikeCondition applies the like operator to the condition builder.
func applyLikeCondition(cb orm.ConditionBuilder, columns []string, fieldValue any, operator Operator) {
	var content string
	switch value := fieldValue.(type) {
	case string:
		content = value
	case *string:
		content = *value
	}

	if content == constants.Empty {
		return
	}

	if len(columns) > 1 {
		applyMultiColumnLikeCondition(cb, columns, content, operator)
		return
	}
	applySingleColumnLikeCondition(cb, columns[0], content, operator)
}

// applyMultiColumnLikeCondition applies like condition for multiple columns with OR logic.
func applyMultiColumnLikeCondition(cb orm.ConditionBuilder, columns []string, content string, operator Operator) {
	cb.Group(func(cb orm.ConditionBuilder) {
		for _, col := range columns {
			applyLikeOperation(cb, col, content, operator, true)
		}
	})
}

// applySingleColumnLikeCondition applies like condition for a single column.
func applySingleColumnLikeCondition(cb orm.ConditionBuilder, column, content string, operator Operator) {
	applyLikeOperation(cb, column, content, operator, false)
}

// applyLikeOperation applies the specific like operation on a column.
func applyLikeOperation(cb orm.ConditionBuilder, column, content string, operator Operator, useOr bool) {
	switch operator {
	case Contains:
		if useOr {
			cb.OrContains(column, content)
		} else {
			cb.Contains(column, content)
		}
	case ContainsIgnoreCase:
		if useOr {
			cb.OrContainsIgnoreCase(column, content)
		} else {
			cb.ContainsIgnoreCase(column, content)
		}
	case NotContains:
		if useOr {
			cb.OrNotContains(column, content)
		} else {
			cb.NotContains(column, content)
		}
	case NotContainsIgnoreCase:
		if useOr {
			cb.OrNotContainsIgnoreCase(column, content)
		} else {
			cb.NotContainsIgnoreCase(column, content)
		}
	case StartsWith:
		if useOr {
			cb.OrStartsWith(column, content)
		} else {
			cb.StartsWith(column, content)
		}
	case StartsWithIgnoreCase:
		if useOr {
			cb.OrStartsWithIgnoreCase(column, content)
		} else {
			cb.StartsWithIgnoreCase(column, content)
		}
	case NotStartsWith:
		if useOr {
			cb.OrNotStartsWith(column, content)
		} else {
			cb.NotStartsWith(column, content)
		}
	case NotStartsWithIgnoreCase:
		if useOr {
			cb.OrNotStartsWithIgnoreCase(column, content)
		} else {
			cb.NotStartsWithIgnoreCase(column, content)
		}
	case EndsWith:
		if useOr {
			cb.OrEndsWith(column, content)
		} else {
			cb.EndsWith(column, content)
		}
	case EndsWithIgnoreCase:
		if useOr {
			cb.OrEndsWithIgnoreCase(column, content)
		} else {
			cb.EndsWithIgnoreCase(column, content)
		}
	case NotEndsWith:
		if useOr {
			cb.OrNotEndsWith(column, content)
		} else {
			cb.NotEndsWith(column, content)
		}
	case NotEndsWithIgnoreCase:
		if useOr {
			cb.OrNotEndsWithIgnoreCase(column, content)
		} else {
			cb.NotEndsWithIgnoreCase(column, content)
		}
	}
}
