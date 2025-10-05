package apis

import (
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/search"
)

// NewAPIBuilder creates a new base API builder instance.
// This is the foundation for all CRUD API builders providing common configuration options.
func NewAPIBuilder[T any](self T, version ...string) APIBuilder[T] {
	return &baseAPIBuilder[T]{
		self: self,
		version: lo.TernaryF(
			len(version) > 0,
			func() string { return version[0] },
			func() string { return api.VersionV1 },
		),
	}
}

// NewCreateAPI creates a new CreateAPI instance.
func NewCreateAPI[TModel, TParams any](version ...string) CreateAPI[TModel, TParams] {
	api := new(createAPI[TModel, TParams])
	api.APIBuilder = NewAPIBuilder[CreateAPI[TModel, TParams]](api, version...)

	return api.Action(ActionCreate)
}

// NewUpdateAPI creates a new updateAPI instance.
func NewUpdateAPI[TModel, TParams any](version ...string) UpdateAPI[TModel, TParams] {
	api := new(updateAPI[TModel, TParams])
	api.APIBuilder = NewAPIBuilder[UpdateAPI[TModel, TParams]](api, version...)

	return api.Action(ActionUpdate)
}

// NewDeleteAPI creates a new deleteAPI instance.
func NewDeleteAPI[TModel any](version ...string) DeleteAPI[TModel] {
	api := new(deleteAPI[TModel])
	api.APIBuilder = NewAPIBuilder[DeleteAPI[TModel]](api, version...)

	return api.Action(ActionDelete)
}

// NewCreateManyAPI creates a new CreateManyAPI instance for batch creation.
func NewCreateManyAPI[TModel, TParams any](version ...string) CreateManyAPI[TModel, TParams] {
	api := new(createManyAPI[TModel, TParams])
	api.APIBuilder = NewAPIBuilder[CreateManyAPI[TModel, TParams]](api, version...)

	return api.Action(ActionCreateMany)
}

// NewUpdateManyAPI creates a new UpdateManyAPI instance for batch update.
func NewUpdateManyAPI[TModel, TParams any](version ...string) UpdateManyAPI[TModel, TParams] {
	api := new(updateManyAPI[TModel, TParams])
	api.APIBuilder = NewAPIBuilder[UpdateManyAPI[TModel, TParams]](api, version...)

	return api.Action(ActionUpdateMany)
}

// NewDeleteManyAPI creates a new DeleteManyAPI instance for batch deletion.
func NewDeleteManyAPI[TModel any](version ...string) DeleteManyAPI[TModel] {
	api := new(deleteManyAPI[TModel])
	api.APIBuilder = NewAPIBuilder[DeleteManyAPI[TModel]](api, version...)

	return api.Action(ActionDeleteMany)
}

func NewFindAPI[TModel, TSearch, TProcessor, TAPI any](self TAPI, version ...string) FindAPI[TModel, TSearch, TProcessor, TAPI] {
	return &baseFindAPI[TModel, TSearch, TProcessor, TAPI]{
		APIBuilder: NewAPIBuilder(self, version...),

		searchApplier: search.Applier[TSearch](),
		self:          self,
	}
}

// NewFindOneAPI creates a new FindOneAPI instance.
func NewFindOneAPI[TModel, TSearch any](version ...string) FindOneAPI[TModel, TSearch] {
	api := new(findOneAPI[TModel, TSearch])
	api.FindAPI = NewFindAPI[TModel, TSearch, TModel, FindOneAPI[TModel, TSearch]](api, version...)

	return api.Action(ActionFindOne)
}

// NewFindAllAPI creates a new FindAllAPI instance.
func NewFindAllAPI[TModel, TSearch any](version ...string) FindAllAPI[TModel, TSearch] {
	api := new(findAllAPI[TModel, TSearch])
	api.FindAPI = NewFindAPI[TModel, TSearch, []TModel, FindAllAPI[TModel, TSearch]](api, version...)

	return api.Action(ActionFindAll)
}

// NewFindPageAPI creates a new FindPageAPI instance.
func NewFindPageAPI[TModel, TSearch any](version ...string) FindPageAPI[TModel, TSearch] {
	api := new(findPageAPI[TModel, TSearch])
	api.FindAPI = NewFindAPI[TModel, TSearch, []TModel, FindPageAPI[TModel, TSearch]](api, version...)

	return api.Action(ActionFindPage)
}

// NewFindOptionsAPI creates a new FindOptionsAPI with the specified options.
func NewFindOptionsAPI[TModel, TSearch any](version ...string) FindOptionsAPI[TModel, TSearch] {
	api := &findOptionsAPI[TModel, TSearch]{
		fieldMapping: &OptionFieldMapping{
			LabelField: defaultLabelField,
			ValueField: defaultValueField,
		},
	}
	api.FindAPI = NewFindAPI[TModel, TSearch, []Option, FindOptionsAPI[TModel, TSearch]](api, version...)

	return api.Action(ActionFindOptions)
}

// NewFindTreeAPI creates a new FindTreeAPI for hierarchical data retrieval.
// The treeBuilder function converts flat database records into nested tree structures.
// Requires models to have id and parent_id fields for parent-child relationships.
func NewFindTreeAPI[TModel, TSearch any](treeBuilder func(flatModels []TModel) []TModel, version ...string) FindTreeAPI[TModel, TSearch] {
	api := &findTreeAPI[TModel, TSearch]{
		idField:       idField,
		parentIdField: parentIdField,
		treeBuilder:   treeBuilder,
	}
	api.FindAPI = NewFindAPI[TModel, TSearch, []TModel, FindTreeAPI[TModel, TSearch]](api, version...)

	return api.Action(ActionFindTree)
}

// NewFindTreeOptionsAPI creates a new FindTreeOptionsAPI with the specified options.
func NewFindTreeOptionsAPI[TModel, TSearch any](version ...string) FindTreeOptionsAPI[TModel, TSearch] {
	api := &findTreeOptionsAPI[TModel, TSearch]{
		fieldMapping: &TreeOptionFieldMapping{
			OptionFieldMapping: OptionFieldMapping{
				LabelField: defaultLabelField,
				ValueField: defaultValueField,
			},
			IdField:       idField,
			ParentIdField: parentIdField,
		},
	}
	api.FindAPI = NewFindAPI[TModel, TSearch, []TreeOption, FindTreeOptionsAPI[TModel, TSearch]](api, version...)

	return api.Action(ActionFindTreeOptions)
}

// NewExportAPI creates a new ExportAPI instance.
func NewExportAPI[TModel, TSearch any](version ...string) ExportAPI[TModel, TSearch] {
	api := new(exportAPI[TModel, TSearch])
	api.FindAPI = NewFindAPI[TModel, TSearch, []TModel, ExportAPI[TModel, TSearch]](api, version...)

	return api.Action(ActionExport)
}

// NewImportAPI creates a new ImportAPI instance.
func NewImportAPI[TModel, TSearch any](version ...string) ImportAPI[TModel, TSearch] {
	api := new(importAPI[TModel, TSearch])
	api.APIBuilder = NewAPIBuilder[ImportAPI[TModel, TSearch]](api, version...)

	return api.Action(ActionImport)
}
