package validator

import (
	"fmt"

	ut "github.com/go-playground/universal-translator"
	v "github.com/go-playground/validator/v10"
)

// presetValidationRules contains all custom validation rules.
var presetValidationRules = []ValidationRule{
	newPhoneNumberRule(), // Phone number validation rule
	newDecimalMinRule(),  // Decimal minimum value validation rule
	newDecimalMaxRule(),  // Decimal maximum value validation rule
}

// ValidationRule defines a custom validation rule with translation support.
type ValidationRule struct {
	RuleTag                  string                         // RuleTag is the unique identifier for the validation rule (used in struct tags)
	ErrMessageTemplate       string                         // ErrMessageTemplate is the error message template with placeholders like {0}, {1}
	Validate                 func(fl v.FieldLevel) bool     // Validate performs the actual validation logic and returns true if valid
	ParseParam               func(fe v.FieldError) []string // ParseParam extracts parameters from FieldError for error message formatting
	CallValidationEvenIfNull bool                           // CallValidationEvenIfNull determines whether to validate nil/zero values
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
