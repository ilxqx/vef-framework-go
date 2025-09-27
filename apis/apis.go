package apis

import "github.com/ilxqx/vef-framework-go/search"

// NewAPIBuilder creates a new base API builder instance.
// This is the foundation for all CRUD API builders providing common configuration options.
func NewAPIBuilder[T any](self T) APIBuilder[T] {
	return &baseAPIBuilder[T]{
		self: self,
	}
}

// NewCreateAPI creates a new CreateAPI instance.
func NewCreateAPI[TModel, TParams any]() CreateAPI[TModel, TParams] {
	api := new(createAPI[TModel, TParams])
	api.APIBuilder = NewAPIBuilder(api).Action(ActionCreate)
	return api
}

// NewUpdateAPI creates a new updateAPI instance.
func NewUpdateAPI[TModel, TParams any]() UpdateAPI[TModel, TParams] {
	api := new(updateAPI[TModel, TParams])
	api.APIBuilder = NewAPIBuilder(api).Action(ActionUpdate)
	return api
}

// NewDeleteAPI creates a new deleteAPI instance.
func NewDeleteAPI[TModel any]() DeleteAPI[TModel] {
	api := new(deleteAPI[TModel])
	api.APIBuilder = NewAPIBuilder(api).Action(ActionDelete)
	return api
}

func NewFindAPI[TModel, TSearch, TProcessor, TAPI any](self TAPI) FindAPI[TModel, TSearch, TProcessor, TAPI] {
	return &baseFindAPI[TModel, TSearch, TProcessor, TAPI]{
		APIBuilder:    NewAPIBuilder(self),
		searchApplier: search.Applier[TSearch](),
		self:          self,
	}
}

// NewFindOneAPI creates a new FindOneAPI instance.
func NewFindOneAPI[TModel, TSearch any]() FindOneAPI[TModel, TSearch] {
	api := new(findOneAPI[TModel, TSearch])
	api.FindAPI = NewFindAPI[TModel, TSearch, TModel, FindOneAPI[TModel, TSearch]](api).Action(ActionFindOne)
	return api
}

// NewFindAllAPI creates a new FindAllAPI instance.
func NewFindAllAPI[TModel, TSearch any]() FindAllAPI[TModel, TSearch] {
	api := new(findAllAPI[TModel, TSearch])
	api.FindAPI = NewFindAPI[TModel, TSearch, []TModel, FindAllAPI[TModel, TSearch]](api).Action(ActionFindAll)
	return api
}

// NewFindPageAPI creates a new FindPageAPI instance.
func NewFindPageAPI[TModel, TSearch any]() FindPageAPI[TModel, TSearch] {
	api := new(findPageAPI[TModel, TSearch])
	api.FindAPI = NewFindAPI[TModel, TSearch, []TModel, FindPageAPI[TModel, TSearch]](api).Action(ActionFindPage)
	return api
}

// NewFindOptionsAPI creates a new FindOptionsAPI with the specified options.
func NewFindOptionsAPI[TModel, TSearch any]() FindOptionsAPI[TModel, TSearch] {
	api := &findOptionsAPI[TModel, TSearch]{
		defaultConfig: &OptionsConfig{
			LabelField: defaultLabelField,
			ValueField: defaultValueField,
		},
	}
	api.FindAPI = NewFindAPI[TModel, TSearch, []Option, FindOptionsAPI[TModel, TSearch]](api).Action(ActionFindOptions)
	return api
}

// NewFindTreeAPI creates a new FindTreeAPI for hierarchical data retrieval.
// The treeBuilder function converts flat database records into nested tree structures.
// Requires models to have id and parent_id fields for parent-child relationships.
func NewFindTreeAPI[TModel, TSearch any](treeBuilder func(flatModels []TModel) []TModel) FindTreeAPI[TModel, TSearch] {
	api := &findTreeAPI[TModel, TSearch]{
		idField:       idField,
		parentIdField: parentIdField,
		treeBuilder:   treeBuilder,
	}
	api.FindAPI = NewFindAPI[TModel, TSearch, []TModel, FindTreeAPI[TModel, TSearch]](api).Action(ActionFindTree)
	return api
}

// NewFindTreeOptionsAPI creates a new FindTreeOptionsAPI with the specified options.
func NewFindTreeOptionsAPI[TModel, TSearch any]() FindTreeOptionsAPI[TModel, TSearch] {
	api := &findTreeOptionsAPI[TModel, TSearch]{
		defaultConfig: &TreeOptionsConfig{
			OptionsConfig: OptionsConfig{
				LabelField: defaultLabelField,
				ValueField: defaultValueField,
			},
			IdField:       idField,
			ParentIdField: parentIdField,
		},
	}
	api.FindAPI = NewFindAPI[TModel, TSearch, []TreeOption, FindTreeOptionsAPI[TModel, TSearch]](api).Action(ActionFindTreeOptions)
	return api
}
