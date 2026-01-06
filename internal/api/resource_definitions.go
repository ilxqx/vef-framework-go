package api

import (
	"github.com/ilxqx/go-streams"

	"github.com/ilxqx/vef-framework-go/api"
)

type ResourceDefinition interface {
	Register(manager api.Manager) error
}

type compositeResourceDefinition struct {
	definitions []ResourceDefinition
}

func (c compositeResourceDefinition) Register(manager api.Manager) error {
	if err := streams.FromSlice(c.definitions).ForEachErr(func(definition ResourceDefinition) error {
		return definition.Register(manager)
	}); err != nil {
		return err
	}

	return nil
}

type simpleResourceDefinition struct {
	apis []*api.Definition
}

func (s simpleResourceDefinition) Register(manager api.Manager) error {
	if err := streams.FromSlice(s.apis).ForEachErr(func(apiDef *api.Definition) error {
		return manager.Register(apiDef)
	}); err != nil {
		return err
	}

	return nil
}
