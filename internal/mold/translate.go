package mold

import (
	"context"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/mold"
)

const (
	descFieldNameSuffix = "Desc"
)

// TranslateTransformer is a translator-based transformer that converts values to readable names
// Supports multiple translators and delegates to the appropriate one based on translation kind (from tag parameters)
type TranslateTransformer struct {
	logger      log.Logger
	translators []mold.Translator
}

// Tag returns the transformer tag name "translate"
func (*TranslateTransformer) Tag() string {
	return "translate"
}

// Transform executes translation transformation logic
// Gets translation kind from tag parameter and field value, then converts through appropriate translator
func (t *TranslateTransformer) Transform(ctx context.Context, fl mold.FieldLevel) error {
	name := fl.Name()
	field := fl.Field()
	value := field.String()

	// Skip empty value or name processing
	if name == constants.Empty || value == constants.Empty {
		return nil
	}

	descField, ok := fl.SiblingField(name + descFieldNameSuffix)
	if !ok {
		t.logger.Warnf("Ignore translation for field '%s' with value '%s' because target desc field '%s%s' is not found", name, value, name, descFieldNameSuffix)
		return nil
	}

	kind := fl.Param()
	if kind == constants.Empty {
		t.logger.Warnf("Ignore translation for field '%s' with value '%s' because translation kind parameter is empty", name, value)
		return nil
	}

	// Find the translator that supports the translation kind
	for _, translator := range t.translators {
		if translator.Supports(kind) {
			translated, err := translator.Translate(ctx, kind, value)
			if err != nil {
				return err
			}
			if descField.CanSet() {
				descField.SetString(translated)
			} else {
				t.logger.Warnf("Ignore translation for field '%s' with value '%s' because target field '%s%s' is not settable", name, value, name, descFieldNameSuffix)
			}
			return nil
		}
	}

	t.logger.Warnf("Ignore translation for field '%s' with value '%s' because no translator supports kind '%s'", name, value, kind)
	return nil
}

// NewTranslateTransformer creates a translate transformer instance
func NewTranslateTransformer(translators []mold.Translator) mold.FieldTransformer {
	return &TranslateTransformer{
		logger:      logger.Named("translate"),
		translators: translators,
	}
}
