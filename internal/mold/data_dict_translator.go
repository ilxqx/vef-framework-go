package mold

import (
	"context"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/mold"
)

const (
	dictKeyPrefix = "dict:"
)

// DataDictTranslator is a data dictionary translator that converts code values to readable names.
type DataDictTranslator struct {
	logger   log.Logger
	resolver mold.DataDictResolver
}

func (t *DataDictTranslator) Supports(kind string) bool {
	return strings.HasPrefix(kind, dictKeyPrefix)
}

func (t *DataDictTranslator) Translate(ctx context.Context, kind, value string) (string, error) {
	// Skip if resolver is nil
	if t.resolver == nil {
		t.logger.Warnf("Ignore dict translation for value '%s' because DataDictResolver is nil, please provide one in the container", value)

		return constants.Empty, nil
	}

	// Extract the dictionary key from the value (remove "dict:" prefix)
	dictKey := kind[len(dictKeyPrefix):]

	result := t.resolver.Resolve(ctx, dictKey, value)

	return result.ValueOrZero(), nil
}

// NewDataDictTranslator creates a data dictionary translator instance.
func NewDataDictTranslator(resolver mold.DataDictResolver) mold.Translator {
	return &DataDictTranslator{
		logger:   logger.Named("data_dict"),
		resolver: resolver,
	}
}
