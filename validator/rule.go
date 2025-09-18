package validator

import (
	"fmt"

	ut "github.com/go-playground/universal-translator"
	v "github.com/go-playground/validator/v10"
)

var (
	// presetValidationRules contains all custom validation rules
	presetValidationRules = []ValidationRule{
		newMobilePhoneRule(), // Mobile phone validation rule
		newDecimalMinRule(),  // Decimal minimum value validation rule
		newDecimalMaxRule(),  // Decimal maximum value validation rule
	}
)

// ValidationRule is the rule for validation.
type ValidationRule struct {
	RuleTag                  string                         // RuleTag is the tag of the validation rule
	ErrMessageTemplate       string                         // ErrMessageTemplate is the template of the error message
	Validate                 func(fl v.FieldLevel) bool     // Validate is the function to validate the value
	ParseParam               func(fe v.FieldError) []string // ParseParam is the function to parse the param
	CallValidationEvenIfNull bool                           // CallValidationEvenIfNull is the flag to call the validation function even if the value is nil
}

// register registers the validation rule to the validator.
func (vr ValidationRule) register(validator *v.Validate) error {
	if err := validator.RegisterValidation(vr.RuleTag, vr.Validate, vr.CallValidationEvenIfNull); err != nil {
		return fmt.Errorf("failed to register '%s' validation rule: %w", vr.RuleTag, err)
	}

	if err := validator.RegisterTranslation(
		vr.RuleTag,
		translator,
		func(t ut.Translator) error {
			return t.Add(vr.RuleTag, vr.ErrMessageTemplate, false)
		},
		func(t ut.Translator, fe v.FieldError) string {
			msg, err := t.T(vr.RuleTag, vr.ParseParam(fe)...)
			if err != nil {
				logger.Errorf("Failed to translate %s: %v", vr.RuleTag, err)
				return vr.ErrMessageTemplate // fallback to template
			}

			return msg
		},
	); err != nil {
		return fmt.Errorf("failed to register '%s' validation rule: %w", vr.RuleTag, err)
	}

	return nil
}
