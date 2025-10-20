package validator

import (
	"regexp"

	v "github.com/go-playground/validator/v10"
)

// Regex patterns for alphanum variations.
var (
	// AlphanumUsRegex validates strings containing only alphanumeric characters and underscores.
	alphanumUsRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	// AlphanumUsSlashRegex validates strings containing alphanumeric characters, underscores, and slashes.
	alphanumUsSlashRegex = regexp.MustCompile(`^[a-zA-Z0-9_/]+$`)
	// AlphanumUsDotRegex validates strings containing alphanumeric characters, underscores, and dots.
	alphanumUsDotRegex = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
)

// newAlphanumUsRule creates a validation rule for alphanumeric characters with underscores.
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

// newAlphanumUsSlashRule creates a validation rule for alphanumeric characters with underscores and slashes.
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

// newAlphanumUsDotRule creates a validation rule for alphanumeric characters with underscores and dots.
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
