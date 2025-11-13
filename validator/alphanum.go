package validator

import (
	"regexp"

	v "github.com/go-playground/validator/v10"
)

var (
	alphanumUsRegex      = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	alphanumUsSlashRegex = regexp.MustCompile(`^[a-zA-Z0-9_/]+$`)
	alphanumUsDotRegex   = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
)

func newAlphanumUsRule() ValidationRule {
	return ValidationRule{
		RuleTag:                  "alphanum_us",
		ErrMessageTemplate:       "{0}只能包含字母、数字和下划线",
		ErrMessageI18nKey:        "validator_alphanum_us",
		CallValidationEvenIfNull: false,
		Validate: func(fl v.FieldLevel) bool {
			return alphanumUsRegex.MatchString(fl.Field().String())
		},
		ParseParam: func(fe v.FieldError) []string {
			return []string{fe.Field()}
		},
	}
}

func newAlphanumUsSlashRule() ValidationRule {
	return ValidationRule{
		RuleTag:                  "alphanum_us_slash",
		ErrMessageTemplate:       "{0}只能包含字母、数字、下划线和斜线",
		ErrMessageI18nKey:        "validator_alphanum_us_slash",
		CallValidationEvenIfNull: false,
		Validate: func(fl v.FieldLevel) bool {
			return alphanumUsSlashRegex.MatchString(fl.Field().String())
		},
		ParseParam: func(fe v.FieldError) []string {
			return []string{fe.Field()}
		},
	}
}

func newAlphanumUsDotRule() ValidationRule {
	return ValidationRule{
		RuleTag:                  "alphanum_us_dot",
		ErrMessageTemplate:       "{0}只能包含字母、数字、下划线和点",
		ErrMessageI18nKey:        "validator_alphanum_us_dot",
		CallValidationEvenIfNull: false,
		Validate: func(fl v.FieldLevel) bool {
			return alphanumUsDotRegex.MatchString(fl.Field().String())
		},
		ParseParam: func(fe v.FieldError) []string {
			return []string{fe.Field()}
		},
	}
}
