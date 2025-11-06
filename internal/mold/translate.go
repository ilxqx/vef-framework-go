package mold

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/ilxqx/vef-framework-go/reflectx"
)

const (
	translatedFieldNameSuffix = "Name"
)

var (
	// ErrTranslatedFieldNotFound is returned when the target translated field (e.g., StatusName) is not found.
	ErrTranslatedFieldNotFound = errors.New("target translated field not found")
	// ErrTranslationKindEmpty is returned when the translation kind parameter is missing.
	ErrTranslationKindEmpty = errors.New("translation kind parameter is empty")
	// ErrTranslatedFieldNotSettable is returned when the target translated field cannot be set.
	ErrTranslatedFieldNotSettable = errors.New("target translated field is not settable")
	// ErrNoTranslatorSupportsKind is returned when no translator supports the given kind.
	ErrNoTranslatorSupportsKind = errors.New("no translator supports the given kind")
	// ErrUnsupportedFieldType is returned when the field type is not supported for translation.
	ErrUnsupportedFieldType = errors.New("unsupported field type for translation")

	nullStringType = reflect.TypeFor[null.String]()
)

// TranslateTransformer is a translator-based transformer that converts values to readable names
// Supports multiple translators and delegates to the appropriate one based on translation kind (from tag parameters).
type TranslateTransformer struct {
	logger      log.Logger
	translators []mold.Translator
}

// extractStringValue extracts string value from supported field types: string, *string, null.String.
// Returns empty string and an error for unsupported types.
func extractStringValue(fieldName string, field reflect.Value) (string, error) {
	if !field.IsValid() {
		return constants.Empty, fmt.Errorf("%w: field %q is invalid", ErrUnsupportedFieldType, fieldName)
	}

	fieldType := field.Type()

	// Handle string
	if fieldType.Kind() == reflect.String {
		return field.String(), nil
	}

	// Handle *string
	if reflectx.Indirect(fieldType).Kind() == reflect.String {
		if field.IsNil() {
			return constants.Empty, nil
		}
		return field.Elem().String(), nil
	}

	// Handle null.String
	if fieldType == nullStringType {
		nullStr := field.Interface().(null.String)
		if !nullStr.Valid {
			return constants.Empty, nil
		}
		return nullStr.String, nil
	}

	return constants.Empty, fmt.Errorf("%w: field %q has unsupported type %v (only string, *string, null.String are supported)", ErrUnsupportedFieldType, fieldName, fieldType)
}

// setTranslatedValue sets the translated string value to the target field.
// Supports string, *string, and null.String types.
func setTranslatedValue(translatedField reflect.Value, translated, translatedFieldName string) error {
	translatedFieldType := translatedField.Type()

	if translatedFieldType.Kind() == reflect.String {
		translatedField.SetString(translated)
		return nil
	}

	if translatedFieldType.Kind() == reflect.Pointer {
		valueType := translatedFieldType.Elem()
		if valueType.Kind() == reflect.String {
			if translatedField.IsNil() {
				translatedField.Set(reflect.New(valueType))
			}
			translatedField.Elem().SetString(translated)
			return nil
		}
		return fmt.Errorf("%w: translated field %q has unsupported pointer type %v", ErrUnsupportedFieldType, translatedFieldName, translatedFieldType)
	}

	if translatedFieldType == nullStringType {
		translatedField.Set(reflect.ValueOf(null.StringFrom(translated)))
		return nil
	}

	return fmt.Errorf("%w: translated field %q has unsupported type %v", ErrUnsupportedFieldType, translatedFieldName, translatedFieldType)
}

// Tag returns the transformer tag name "translate".
func (*TranslateTransformer) Tag() string {
	return "translate"
}

// Transform executes translation transformation logic
// Gets translation kind from tag parameter and field value, then converts through appropriate translator.
func (t *TranslateTransformer) Transform(ctx context.Context, fl mold.FieldLevel) error {
	name := fl.Name()
	field := fl.Field()

	// Extract string value from supported field types
	value, err := extractStringValue(name, field)
	if err != nil {
		return err
	}

	// Skip empty value or name processing
	if name == constants.Empty || value == constants.Empty {
		return nil
	}

	translatedFieldName := name + translatedFieldNameSuffix
	translatedField, ok := fl.SiblingField(translatedFieldName)
	if !ok {
		return fmt.Errorf("%w: failed to get field %q for field %q with value %q", ErrTranslatedFieldNotFound, translatedFieldName, name, value)
	}

	kind := fl.Param()
	if kind == constants.Empty {
		return fmt.Errorf("%w: field %q with value %q", ErrTranslationKindEmpty, name, value)
	}

	// Find the translator that supports the translation kind
	for _, translator := range t.translators {
		if translator.Supports(kind) {
			translated, err := translator.Translate(ctx, kind, value)
			if err != nil {
				return err
			}

			if !translatedField.CanSet() {
				return fmt.Errorf("%w: field %q for field %q with value %q", ErrTranslatedFieldNotSettable, translatedFieldName, name, value)
			}

			return setTranslatedValue(translatedField, translated, translatedFieldName)
		}
	}

	if strings.HasSuffix(kind, constants.QuestionMark) {
		return nil
	}

	return fmt.Errorf("%w: kind %q for field %q with value %q", ErrNoTranslatorSupportsKind, kind, name, value)
}

// NewTranslateTransformer creates a translate transformer instance.
func NewTranslateTransformer(translators []mold.Translator) mold.FieldTransformer {
	return &TranslateTransformer{
		logger:      logger.Named("translate"),
		translators: translators,
	}
}
