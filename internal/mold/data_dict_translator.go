package mold

import (
	"context"
	"errors"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/mold"
)

const (
	dictKeyPrefix = "dict:"
)

// ErrDataDictResolverNotConfigured is returned when DataDictResolver is not provided.
var ErrDataDictResolverNotConfigured = errors.New("data dictionary resolver is not configured, please provide one in the container")

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
		return constants.Empty, ErrDataDictResolverNotConfigured
	}

	// Extract the dictionary key from the value (remove "dict:" prefix)
	dictKey := kind[len(dictKeyPrefix):]

	result, err := t.resolver.Resolve(ctx, dictKey, value)
	if err != nil {
		t.logger.Errorf("Failed to resolve dictionary %q for code %q: %v", dictKey, value, err)

		return constants.Empty, err
	}

	return result, nil
}

// NewDataDictTranslator creates a data dictionary translator instance.
func NewDataDictTranslator(resolver mold.DataDictResolver) mold.Translator {
	return &DataDictTranslator{
		logger:   logger.Named("data_dict"),
		resolver: resolver,
	}
}
