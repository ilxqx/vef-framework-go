package validator

import (
	"fmt"
	"strings"

	ut "github.com/go-playground/universal-translator"
	v "github.com/go-playground/validator/v10"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/i18n"
)

// presetValidationRules contains all custom validation rules.
var presetValidationRules = []ValidationRule{
	newPhoneNumberRule(),     // Phone number validation rule
	newDecimalMinRule(),      // Decimal minimum value validation rule
	newDecimalMaxRule(),      // Decimal maximum value validation rule
	newAlphanumUsRule(),      // Alphanumeric with underscores validation rule
	newAlphanumUsSlashRule(), // Alphanumeric with underscores and slashes validation rule
	newAlphanumUsDotRule(),   // Alphanumeric with underscores and dots validation rule
}

// ValidationRule defines a custom validation rule with translation support.
type ValidationRule struct {
	RuleTag                  string                         // RuleTag is the unique identifier for the validation rule (used in struct tags)
	ErrMessageTemplate       string                         // ErrMessageTemplate is the error message template with placeholders like {0}, {1} (used as fallback)
	ErrMessageI18nKey        string                         // ErrMessageI18nKey is the i18n key for the error message (optional, takes precedence over ErrMessageTemplate)
	Validate                 func(fl v.FieldLevel) bool     // Validate performs the actual validation logic and returns true if valid
	ParseParam               func(fe v.FieldError) []string // ParseParam extracts parameters from FieldError for error message formatting
	CallValidationEvenIfNull bool                           // CallValidationEvenIfNull determines whether to validate nil/zero values
}

// register registers the validation rule to the validator.
func (vr ValidationRule) register(validator *v.Validate) error {
	if err := validator.RegisterValidation(vr.RuleTag, vr.Validate, vr.CallValidationEvenIfNull); err != nil {
		return fmt.Errorf("failed to register %q validation rule: %w", vr.RuleTag, err)
	}

	if err := validator.RegisterTranslation(
		vr.RuleTag,
		translator,
		func(t ut.Translator) error {
			return t.Add(vr.RuleTag, vr.ErrMessageTemplate, false)
		},
		func(t ut.Translator, fe v.FieldError) string {
			// Try i18n translation first if key is provided
			if vr.ErrMessageI18nKey != constants.Empty {
				msg := i18n.T(vr.ErrMessageI18nKey)
				// If translation found (not equal to key), replace placeholders
				if msg != vr.ErrMessageI18nKey {
					return vr.replacePlaceholders(msg, vr.ParseParam(fe))
				}
			}

			// Fallback to go-playground/validator translation
			msg, err := t.T(vr.RuleTag, vr.ParseParam(fe)...)
			if err != nil {
				logger.Errorf("Failed to translate %s: %v", vr.RuleTag, err)

				return vr.ErrMessageTemplate // fallback to template
			}

			return msg
		},
	); err != nil {
		return fmt.Errorf("failed to register %q validation rule: %w", vr.RuleTag, err)
	}

	return nil
}

// replacePlaceholders replaces {0}, {1}, {2}... In the message with the provided params.
func (vr ValidationRule) replacePlaceholders(message string, params []string) string {
	result := message
	for i, param := range params {
		placeholder := fmt.Sprintf("{%d}", i)
		result = strings.ReplaceAll(result, placeholder, param)
	}

	return result
}
