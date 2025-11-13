package api

import "github.com/ilxqx/vef-framework-go/api"

type ResourceDefinition interface {
	Register(manager api.Manager) error
}

type compositeResourceDefinition struct {
	definitions []ResourceDefinition
}

func (c compositeResourceDefinition) Register(manager api.Manager) error {
	for _, definition := range c.definitions {
		if err := definition.Register(manager); err != nil {
			return err
		}
	}

	return nil
}

type simpleResourceDefinition struct {
	apis []*api.Definition
}

func (s simpleResourceDefinition) Register(manager api.Manager) error {
	for _, apiDef := range s.apis {
		if err := manager.Register(apiDef); err != nil {
			return err
		}
	}

	return nil
}
