package collector

import (
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/internal/log"
	"github.com/ilxqx/vef-framework-go/reflectx"
)

var (
	providerType = reflect.TypeFor[api.OperationsProvider]()
	logger       = log.Named("api.collector")
)

// EmbeddedProviderCollector collects API specs from embedded anonymous structs
// that implement the api.OperationsProvider interface.
type EmbeddedProviderCollector struct{}

// NewEmbeddedProviderCollector creates a new collector for embedded providers.
func NewEmbeddedProviderCollector() api.OperationsCollector {
	return new(EmbeddedProviderCollector)
}

// Collect gathers all API specs from embedded providers in the resource.
func (c *EmbeddedProviderCollector) Collect(resource api.Resource) []api.OperationSpec {
	var specs []api.OperationSpec

	visitor := reflectx.Visitor{
		VisitField: func(field reflect.StructField, fieldValue reflect.Value, _ int) reflectx.VisitAction {
			if !field.Anonymous {
				return reflectx.Continue
			}

			if !isProviderImplementation(fieldValue) {
				return reflectx.Continue
			}

			if provider, ok := fieldValue.Interface().(api.OperationsProvider); ok {
				ops := provider.Provide()

				specs = append(specs, ops...)
				if len(ops) > 0 {
					logger.Infof("Collected %d Api operations from embedded provider: %s",
						len(ops), field.Type.String())
				}
			}

			return reflectx.Continue
		},
	}

	reflectx.VisitOf(resource, visitor)

	return specs
}

func isProviderImplementation(value reflect.Value) bool {
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return false
		}

		value = value.Elem()
	}

	t := value.Type()

	return t.Implements(providerType) || reflect.PointerTo(t).Implements(providerType)
}
