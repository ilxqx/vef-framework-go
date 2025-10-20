package apis

import (
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/search"
)

// NewApiBuilder creates a new base Api builder instance.
// This is the foundation for all CRUD Api builders providing common configuration options.
func NewApiBuilder[T any](self T, version ...string) ApiBuilder[T] {
	return &baseApiBuilder[T]{
		self: self,
		version: lo.TernaryF(
			len(version) > 0,
			func() string { return version[0] },
			func() string { return api.VersionV1 },
		),
	}
}

// NewCreateApi creates a new CreateApi instance.
func NewCreateApi[TModel, TParams any](version ...string) CreateApi[TModel, TParams] {
	api := new(createApi[TModel, TParams])
	api.ApiBuilder = NewApiBuilder[CreateApi[TModel, TParams]](api, version...)

	return api.Action(ActionCreate)
}

// NewUpdateApi creates a new UpdateApi instance.
func NewUpdateApi[TModel, TParams any](version ...string) UpdateApi[TModel, TParams] {
	api := new(updateApi[TModel, TParams])
	api.ApiBuilder = NewApiBuilder[UpdateApi[TModel, TParams]](api, version...)

	return api.Action(ActionUpdate)
}

// NewDeleteApi creates a new DeleteApi instance.
func NewDeleteApi[TModel any](version ...string) DeleteApi[TModel] {
	api := new(deleteApi[TModel])
	api.ApiBuilder = NewApiBuilder[DeleteApi[TModel]](api, version...)

	return api.Action(ActionDelete)
}

// NewCreateManyApi creates a new CreateManyApi instance for batch creation.
func NewCreateManyApi[TModel, TParams any](version ...string) CreateManyApi[TModel, TParams] {
	api := new(createManyApi[TModel, TParams])
	api.ApiBuilder = NewApiBuilder[CreateManyApi[TModel, TParams]](api, version...)

	return api.Action(ActionCreateMany)
}

// NewUpdateManyApi creates a new UpdateManyApi instance for batch update.
func NewUpdateManyApi[TModel, TParams any](version ...string) UpdateManyApi[TModel, TParams] {
	api := new(updateManyApi[TModel, TParams])
	api.ApiBuilder = NewApiBuilder[UpdateManyApi[TModel, TParams]](api, version...)

	return api.Action(ActionUpdateMany)
}

// NewDeleteManyApi creates a new DeleteManyApi instance for batch deletion.
func NewDeleteManyApi[TModel any](version ...string) DeleteManyApi[TModel] {
	api := new(deleteManyApi[TModel])
	api.ApiBuilder = NewApiBuilder[DeleteManyApi[TModel]](api, version...)

	return api.Action(ActionDeleteMany)
}

func NewFindApi[TModel, TSearch, TProcessor, TApi any](self TApi, version ...string) FindApi[TModel, TSearch, TProcessor, TApi] {
	return &baseFindApi[TModel, TSearch, TProcessor, TApi]{
		ApiBuilder: NewApiBuilder(self, version...),

		searchApplier: search.Applier[TSearch](),
		self:          self,
	}
}

// NewFindOneApi creates a new FindOneApi instance.
func NewFindOneApi[TModel, TSearch any](version ...string) FindOneApi[TModel, TSearch] {
	api := new(findOneApi[TModel, TSearch])
	api.FindApi = NewFindApi[TModel, TSearch, TModel, FindOneApi[TModel, TSearch]](api, version...)

	return api.Action(ActionFindOne)
}

// NewFindAllApi creates a new FindAllApi instance.
func NewFindAllApi[TModel, TSearch any](version ...string) FindAllApi[TModel, TSearch] {
	api := new(findAllApi[TModel, TSearch])
	api.FindApi = NewFindApi[TModel, TSearch, []TModel, FindAllApi[TModel, TSearch]](api, version...)

	return api.Action(ActionFindAll)
}

// NewFindPageApi creates a new FindPageApi instance.
func NewFindPageApi[TModel, TSearch any](version ...string) FindPageApi[TModel, TSearch] {
	api := new(findPageApi[TModel, TSearch])
	api.FindApi = NewFindApi[TModel, TSearch, []TModel, FindPageApi[TModel, TSearch]](api, version...)

	return api.Action(ActionFindPage)
}

// NewFindOptionsApi creates a new FindOptionsApi with the specified options.
func NewFindOptionsApi[TModel, TSearch any](version ...string) FindOptionsApi[TModel, TSearch] {
	api := &findOptionsApi[TModel, TSearch]{
		columnMapping: &OptionColumnMapping{
			LabelColumn: defaultLabelColumn,
			ValueColumn: defaultValueColumn,
		},
	}
	api.FindApi = NewFindApi[TModel, TSearch, []Option, FindOptionsApi[TModel, TSearch]](api, version...)

	return api.Action(ActionFindOptions)
}

// NewFindTreeApi creates a new FindTreeApi for hierarchical data retrieval.
// The treeBuilder function converts flat database records into nested tree structures.
// Requires models to have id and parent_id columns for parent-child relationships.
func NewFindTreeApi[TModel, TSearch any](treeBuilder func(flatModels []TModel) []TModel, version ...string) FindTreeApi[TModel, TSearch] {
	api := &findTreeApi[TModel, TSearch]{
		idColumn:       idColumn,
		parentIdColumn: parentIdColumn,
		treeBuilder:    treeBuilder,
	}
	api.FindApi = NewFindApi[TModel, TSearch, []TModel, FindTreeApi[TModel, TSearch]](api, version...)

	return api.Action(ActionFindTree)
}

// NewFindTreeOptionsApi creates a new FindTreeOptionsApi with the specified options.
func NewFindTreeOptionsApi[TModel, TSearch any](version ...string) FindTreeOptionsApi[TModel, TSearch] {
	api := &findTreeOptionsApi[TModel, TSearch]{
		columnMapping: &TreeOptionColumnMapping{
			OptionColumnMapping: OptionColumnMapping{
				LabelColumn: defaultLabelColumn,
				ValueColumn: defaultValueColumn,
			},
			IdColumn:       idColumn,
			ParentIdColumn: parentIdColumn,
		},
	}
	api.FindApi = NewFindApi[TModel, TSearch, []TreeOption, FindTreeOptionsApi[TModel, TSearch]](api, version...)

	return api.Action(ActionFindTreeOptions)
}

// NewExportApi creates a new ExportApi instance.
func NewExportApi[TModel, TSearch any](version ...string) ExportApi[TModel, TSearch] {
	api := new(exportApi[TModel, TSearch])
	api.FindApi = NewFindApi[TModel, TSearch, []TModel, ExportApi[TModel, TSearch]](api, version...)

	return api.Action(ActionExport)
}

// NewImportApi creates a new ImportApi instance.
func NewImportApi[TModel, TSearch any](version ...string) ImportApi[TModel, TSearch] {
	api := new(importApi[TModel, TSearch])
	api.ApiBuilder = NewApiBuilder[ImportApi[TModel, TSearch]](api, version...)

	return api.Action(ActionImport)
}
