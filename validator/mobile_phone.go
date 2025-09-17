package validator

import (
	"regexp"

	v "github.com/go-playground/validator/v10"
)

var (
	mobilePhoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`) // mobilePhoneRegex validates Chinese mobile phone numbers
)

// newMobilePhoneRule creates a new validation rule for mobile phone.
func newMobilePhoneRule() ValidationRule {
	return ValidationRule{
		RuleTag:                  "mobile_phone",
		ErrMessageTemplate:       "{0}格式不正确",
		CallValidationEvenIfNull: false,
		Validate: func(fl v.FieldLevel) bool {
			return mobilePhoneRegex.MatchString(fl.Field().String())
		},
		ParseParam: func(fe v.FieldError) []string {
			return []string{fe.Field()}
		},
	}
}
