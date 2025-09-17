package validator

import (
	v "github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// newDecimalMinRule creates a new validation rule for decimal minimum value.
func newDecimalMinRule() ValidationRule {
	return newDecimalComparisonRule("dec_min", "{0}最小只能为{1}", func(dec, threshold decimal.Decimal) bool {
		return dec.GreaterThanOrEqual(threshold)
	})
}

// newDecimalMaxRule creates a new validation rule for decimal maximum value.
func newDecimalMaxRule() ValidationRule {
	return newDecimalComparisonRule("dec_max", "{0}必须小于或等于{1}", func(dec, threshold decimal.Decimal) bool {
		return dec.LessThanOrEqual(threshold)
	})
}

// newDecimalComparisonRule creates a validation rule for decimal comparison operations.
func newDecimalComparisonRule(ruleTag, errMessageTemplate string, compare func(decimal.Decimal, decimal.Decimal) bool) ValidationRule {
	return ValidationRule{
		RuleTag:                  ruleTag,
		ErrMessageTemplate:       errMessageTemplate,
		CallValidationEvenIfNull: false,
		Validate: func(fl v.FieldLevel) bool {
			dec, ok := fl.Field().Interface().(decimal.Decimal)
			if !ok {
				logger.Warnf("[%s] %s requires a decimal.Decimal, but got %s", fl.FieldName(), ruleTag, fl.Field().Type().String())
				return false
			}

			threshold, err := decimal.NewFromString(fl.Param())
			if err != nil {
				logger.Warnf("[%s] Failed to parse the param of %s: %v", fl.FieldName(), ruleTag, err)
				return false
			}

			return compare(dec, threshold)
		},
		ParseParam: func(fe v.FieldError) []string {
			return []string{fe.Field(), fe.Param()}
		},
	}
}
