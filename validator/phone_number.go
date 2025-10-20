package validator

import (
	"regexp"

	v "github.com/go-playground/validator/v10"
)

// phoneNumberRegex validates 11-digit phone numbers.
var phoneNumberRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// newPhoneNumberRule creates a new validation rule for phone numbers.
func newPhoneNumberRule() ValidationRule {
	return ValidationRule{
		RuleTag:                  "phone_number",
		ErrMessageTemplate:       "{0}格式不正确",
		ErrMessageI18nKey:        "validator_phone_number",
		CallValidationEvenIfNull: false,
		Validate: func(fl v.FieldLevel) bool {
			return phoneNumberRegex.MatchString(fl.Field().String())
		},
		ParseParam: func(fe v.FieldError) []string {
			return []string{fe.Field()}
		},
	}
}
