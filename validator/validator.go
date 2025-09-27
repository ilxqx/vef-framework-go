package validator

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/ilxqx/vef-framework-go/result"

	enLocale "github.com/go-playground/locales/en"
	zhLocale "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	v "github.com/go-playground/validator/v10"
	enTranslation "github.com/go-playground/validator/v10/translations/en"
	zhTranslation "github.com/go-playground/validator/v10/translations/zh"
	"github.com/samber/lo"
)

const (
	// tagLabel is the struct tag name for field labels
	tagLabel = "label"
)

var (
	logger = log.Named("validator")
	// translator handles error message translations
	translator ut.Translator
	// validator is the main validation instance
	validator *v.Validate
)

func init() {
	// Initialize validator with i18n support based on environment configuration
	preferredLanguage := lo.CoalesceOrEmpty(os.Getenv(constants.EnvI18NLanguage), constants.DefaultI18NLanguage)
	localeTranslator := lo.TernaryF(
		preferredLanguage == constants.DefaultI18NLanguage,
		zhLocale.New,
		enLocale.New,
	)
	universalTranslator := ut.New(localeTranslator, localeTranslator)

	translator, _ = universalTranslator.GetTranslator(
		lo.Ternary(
			preferredLanguage == constants.DefaultI18NLanguage,
			"zh",
			"en",
		),
	)
	validator = v.New(v.WithRequiredStructEnabled())

	// Register translations
	if err := lo.TernaryF(
		preferredLanguage == constants.DefaultI18NLanguage,
		func() error {
			return zhTranslation.RegisterDefaultTranslations(validator, translator)
		},
		func() error {
			return enTranslation.RegisterDefaultTranslations(validator, translator)
		},
	); err != nil {
		panic(
			fmt.Errorf("failed to register default translations: %w", err),
		)
	}

	// Register field name function
	validator.RegisterTagNameFunc(func(field reflect.StructField) string {
		label := field.Tag.Get(tagLabel)
		if lo.IsEmpty(label) {
			return field.Name
		}

		return label
	})

	setup()
}

// RegisterValidationRules registers custom validation rules.
func RegisterValidationRules(rules ...ValidationRule) error {
	for _, rule := range rules {
		if err := rule.register(validator); err != nil {
			return err
		}
	}

	return nil
}

// CustomTypeFunc defines a custom type function for validation that extracts values from custom types.
type CustomTypeFunc = func(field reflect.Value) any

// RegisterTypeFunc registers a custom type function for specified types.
func RegisterTypeFunc(fn CustomTypeFunc, types ...any) {
	validator.RegisterCustomTypeFunc(fn, types...)
}

// RegisterNullValueTypeFunc registers a type function for null.Value[T] types.
func RegisterNullValueTypeFunc[T any]() {
	validator.RegisterCustomTypeFunc(
		func(field reflect.Value) any {
			if nv, ok := field.Interface().(null.Value[T]); ok && nv.Valid {
				return nv.V
			}

			return nil
		},
		null.Value[T]{},
	)
}

// Validate validates the value.
func Validate(value any) error {
	if err := validator.Struct(value); err != nil {
		var validationErrors v.ValidationErrors
		errors.As(err, &validationErrors)
		for _, validationError := range validationErrors {
			return result.ErrWithCode(result.ErrCodeBadRequest, validationError.Translate(translator))
		}
	}

	return nil
}
