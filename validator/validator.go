package validator

import (
	"errors"
	"reflect"

	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/ilxqx/vef-framework-go/result"

	zhLocale "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	v "github.com/go-playground/validator/v10"
	zhTranslation "github.com/go-playground/validator/v10/translations/zh"
	"github.com/samber/lo"
)

const (
	tagLabel = "label" // tagLabel is the struct tag name for field labels
)

var (
	logger     = log.Named("validator") // logger is the validator module logger
	translator ut.Translator            // translator handles error message translations
	validator  *v.Validate              // validator is the main validation instance
)

func init() {
	zh := zhLocale.New()
	universalTranslator := ut.New(zh, zh)

	translator, _ = universalTranslator.GetTranslator("zh")
	validator = v.New(v.WithRequiredStructEnabled())

	// RegisterDefaultTranslations registers Chinese translations
	if err := zhTranslation.RegisterDefaultTranslations(validator, translator); err != nil {
		logger.Panicf("Failed to register default translations: %v", err)
	}

	// RegisterTagNameFunc sets custom field name function
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
func RegisterValidationRules(rules ...ValidationRule) {
	for _, rule := range rules {
		rule.register(validator)
	}
}

// TypeFunc defines a custom type function for validation.
type TypeFunc = func(field reflect.Value) any

// RegisterTypeFunc registers a custom type function for specified types.
func RegisterTypeFunc(fn TypeFunc, types ...any) {
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

// RegisterNullJSONTypeFunc registers a type function for null.JSON[T] types.
func RegisterNullJSONTypeFunc[T any]() {
	validator.RegisterCustomTypeFunc(
		func(field reflect.Value) any {
			if nv, ok := field.Interface().(null.JSON[T]); ok && nv.Valid {
				return nv.V.Unwrap()
			}

			return nil
		},
		null.JSON[T]{},
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
