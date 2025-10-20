package mold

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/mold"
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
)

// TranslateTransformer is a translator-based transformer that converts values to readable names
// Supports multiple translators and delegates to the appropriate one based on translation kind (from tag parameters).
type TranslateTransformer struct {
	logger      log.Logger
	translators []mold.Translator
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
	value := field.String()

	// Skip empty value or name processing
	if name == constants.Empty || value == constants.Empty {
		return nil
	}

	translatedField, ok := fl.SiblingField(name + translatedFieldNameSuffix)
	if !ok {
		return fmt.Errorf("%w: failed to get field %q for field %q with value %q", ErrTranslatedFieldNotFound, name+translatedFieldNameSuffix, name, value)
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

			if translatedField.CanSet() {
				translatedField.SetString(translated)
			} else {
				return fmt.Errorf("%w: field %q for field %q with value %q", ErrTranslatedFieldNotSettable, name+translatedFieldNameSuffix, name, value)
			}

			return nil
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
