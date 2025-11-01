package api

import "github.com/ilxqx/vef-framework-go/api"

// ResourceDefinition represents a collection of Api definitions that can be registered with a manager.
type ResourceDefinition interface {
	// Register registers all Api definitions with the given manager.
	// Returns an error if any Api registration fails (e.g., duplicate definitions).
	Register(manager api.Manager) error
}

// compositeResourceDefinition combines multiple resource definitions into one.
type compositeResourceDefinition struct {
	definitions []ResourceDefinition
}

// Register registers all contained resource definitions with the manager.
// Returns an error immediately if any registration fails.
func (c compositeResourceDefinition) Register(manager api.Manager) error {
	for _, definition := range c.definitions {
		if err := definition.Register(manager); err != nil {
			return err
		}
	}

	return nil
}

// simpleResourceDefinition contains a collection of Api definitions from a single resource.
type simpleResourceDefinition struct {
	apis []*api.Definition
}

// Register registers all Api definitions with the manager.
// Returns an error immediately if any registration fails.
func (s simpleResourceDefinition) Register(manager api.Manager) error {
	for _, api := range s.apis {
		if err := manager.Register(api); err != nil {
			return err
		}
	}

	return nil
}
