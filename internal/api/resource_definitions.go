package api

import "github.com/ilxqx/vef-framework-go/api"

// ResourceDefinition represents a collection of API definitions that can be registered with a manager.
type ResourceDefinition interface {
	// Register registers all API definitions with the given manager.
	Register(manager api.Manager)
}

// compositeResourceDefinition combines multiple resource definitions into one.
type compositeResourceDefinition struct {
	definitions []ResourceDefinition
}

// Register registers all contained resource definitions with the manager.
func (c compositeResourceDefinition) Register(manager api.Manager) {
	for _, definition := range c.definitions {
		definition.Register(manager)
	}
}

// simpleResourceDefinition contains a collection of API definitions from a single resource.
type simpleResourceDefinition struct {
	apis []*api.Definition
}

// Register registers all API definitions with the manager.
func (s simpleResourceDefinition) Register(manager api.Manager) {
	for _, api := range s.apis {
		manager.Register(api)
	}
}
