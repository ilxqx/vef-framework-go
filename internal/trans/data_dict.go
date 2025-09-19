package trans

import (
	"context"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/trans"
)

// dataDictTransformer is a data dictionary transformer that converts code values to readable names
// Supports specifying dictionary type through tag parameters, e.g.: trans:"dict=user_status"
type dataDictTransformer struct {
	logger   log.Logger
	resolver trans.DataDictNameResolver
}

// Tag returns the transformer tag name "dict"
func (*dataDictTransformer) Tag() string {
	return "dict"
}

// Transform executes data dictionary transformation logic
// Gets code value from field and converts it to corresponding name through resolver
func (t *dataDictTransformer) Transform(ctx context.Context, fl trans.FieldLevel) error {
	field := fl.Field()
	value := field.String()

	// Skip if resolver is nil
	if t.resolver == nil {
		t.logger.Warnf("Ignore dict transformation for code '%s' because DataDictNameResolver is nil, please provide one in the container", value)
		return nil
	}

	// Skip empty value processing
	if value == constants.Empty {
		return nil
	}

	// Get dictionary key to distinguish different types of data dictionaries
	dictKey := fl.Param()
	if dictKey == constants.Empty {
		t.logger.Warnf("Ignore dict transformation for code '%s' because dict key is empty", value)
		return nil
	}

	// Resolve code value to name, set to field if found
	if name := t.resolver.Resolve(ctx, dictKey, value); name.Valid {
		field.SetString(name.String)
	}
	return nil
}

// newDataDictTransformer creates a data dictionary transformer instance
func newDataDictTransformer(resolver trans.DataDictNameResolver) trans.FieldTransformer {
	return &dataDictTransformer{
		logger:   logger.Named("data_dict"),
		resolver: resolver,
	}
}
