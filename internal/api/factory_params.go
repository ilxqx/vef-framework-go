package api

import (
	"fmt"
	"reflect"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/reflectx"
)

type FactoryResolverFunc func() reflect.Value

type FactoryParamResolverManager struct {
	resolvers map[reflect.Type]FactoryResolverFunc
}

// NewFactoryParamResolverManager creates a manager where user-provided resolvers
// override framework built-in resolvers when types overlap.
func NewFactoryParamResolverManager(resolvers []api.FactoryParamResolver) *FactoryParamResolverManager {
	merged := make(map[reflect.Type]FactoryResolverFunc, len(resolvers))

	for _, resolver := range resolvers {
		t := resolver.Type()
		merged[t] = resolver.Resolve
	}

	return &FactoryParamResolverManager{resolvers: merged}
}

// Resolve attempts exact type matching first, then falls back to field injection from target resource.
func (m *FactoryParamResolverManager) Resolve(
	target reflect.Value,
	paramType reflect.Type,
) (FactoryResolverFunc, error) {
	if resolver, ok := m.resolvers[paramType]; ok {
		return resolver, nil
	}

	// Fallback: allow factory functions to access resource fields
	if field := findFieldInStruct(target, paramType); field.IsValid() {
		return buildFactoryFieldResolver(field, paramType)
	}

	return nil, fmt.Errorf("%w: %s", ErrResolveFactoryParamType, paramType.String())
}

func buildFactoryFieldResolver(
	field reflect.Value,
	targetType reflect.Type,
) (FactoryResolverFunc, error) {
	converted, err := reflectx.ConvertValue(field, targetType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert field value: %w", err)
	}

	return func() reflect.Value { return converted }, nil
}
