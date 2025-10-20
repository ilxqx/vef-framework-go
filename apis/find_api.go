package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/orm"
)

type baseFindApi[TModel, TSearch, TProcessorIn, TApi any] struct {
	ApiBuilder[TApi]

	searchApplier       SearchApplier[TSearch]
	filterApplier       FilterApplier[TSearch]
	queryApplier        QueryApplier[TSearch]
	sortApplier         SortApplier[TSearch]
	relations           []orm.RelationSpec
	processor           Processor[TProcessorIn, TSearch]
	disableDataPerm     bool
	auditUserModel      any    // User model for joining audit user names
	auditUserNameColumn string // Column name for user name, default "name"

	// Cached/pre-computed data
	shouldApplyDefaultSort bool // Whether default created_at ordering should be applied
	initialized            bool // Whether Init has been called

	self TApi
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) DisableDataPerm() TApi {
	a.disableDataPerm = true

	return a.self
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) WithAuditUserNames(userModel any, nameColumn ...string) TApi {
	a.auditUserModel = userModel
	if len(nameColumn) > 0 && nameColumn[0] != constants.Empty {
		a.auditUserNameColumn = nameColumn[0]
	} else {
		a.auditUserNameColumn = "name"
	}

	return a.self
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) QueryApplier(applier QueryApplier[TSearch]) TApi {
	a.queryApplier = applier

	return a.self
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) FilterApplier(applier FilterApplier[TSearch]) TApi {
	a.filterApplier = applier

	return a.self
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) SortApplier(applier SortApplier[TSearch]) TApi {
	a.sortApplier = applier

	return a.self
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) Relations(relations ...orm.RelationSpec) TApi {
	a.relations = append(a.relations, relations...)

	return a.self
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) Processor(processor Processor[TProcessorIn, TSearch]) TApi {
	a.processor = processor

	return a.self
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) Init(db orm.Db) error {
	if a.initialized {
		return nil
	}
	defer func() { a.initialized = true }()

	// Pre-compute default sort configuration
	table := db.TableOf((*TModel)(nil))
	hasCreatedAt := table.HasField(constants.ColumnCreatedAt)
	a.shouldApplyDefaultSort = a.sortApplier == nil && hasCreatedAt

	// Validate audit user model if configured
	if a.auditUserModel != nil {
		pkLen := len(db.TableOf(a.auditUserModel).PKs)
		if pkLen == 0 {
			return ErrModelNoPrimaryKey
		}

		if pkLen > 1 {
			return ErrAuditUserCompositePK
		}
	}

	return nil
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ShouldApplyDefaultSort() bool {
	return a.shouldApplyDefaultSort
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ApplyDefaultSort(query orm.SelectQuery) {
	if a.shouldApplyDefaultSort {
		query.OrderByDesc(constants.ColumnCreatedAt)
	}
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) BuildQuery(db orm.Db, model any, search TSearch, ctx fiber.Ctx) orm.SelectQuery {
	query := db.NewSelect()
	a.ConfigureQuery(query, model, search, ctx)

	return query
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ConfigureQuery(query orm.SelectQuery, model any, search TSearch, ctx fiber.Ctx) {
	query.Model(model)

	// Apply data permission if available
	a.ApplyDataPermission(query, ctx)

	// Apply other query conditions
	a.ApplyConditions(query, search, ctx)
	a.ApplyRelations(query, search, ctx)

	// Apply audit user relations if configured
	a.ApplyAuditUserRelations(query)

	a.ApplyQuery(query, search, ctx)
	a.ApplySort(query, search, ctx)
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) Process(input TProcessorIn, search TSearch, ctx fiber.Ctx) any {
	if a.processor == nil {
		return input
	}

	return a.processor(input, search, ctx)
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ApplySort(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	applySort(query, a.sortApplier, search, ctx)
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ApplySearch(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	query.Where(func(cb orm.ConditionBuilder) {
		cb.Apply(a.searchApplier(search))
	})
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ApplyFilter(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	query.Where(func(cb orm.ConditionBuilder) {
		cb.Apply(a.filterApplier(search, ctx))
	})
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ApplyConditions(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	query.Where(func(cb orm.ConditionBuilder) {
		cb.Apply(a.searchApplier(search))

		if a.filterApplier != nil {
			cb.Apply(a.filterApplier(search, ctx))
		}
	})
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ApplyQuery(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	if a.queryApplier != nil {
		query.Apply(a.queryApplier(search, ctx))
	}
}

func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ApplyRelations(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	if len(a.relations) > 0 {
		query.JoinRelations(a.relations...)
	}
}

// ApplyAuditUserRelations applies RelationSpec configurations to query audit user names (created_by_name, updated_by_name).
// This method creates LEFT JOIN relations with the user model to populate audit user name fields.
// It automatically constructs two RelationSpecs: one for the creator and one for the updater.
// The method is called automatically during query building if WithAuditUserNames was configured.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ApplyAuditUserRelations(query orm.SelectQuery) {
	// Skip if audit user model is not configured
	if a.auditUserModel == nil {
		return
	}

	// Create RelationSpecs for creator and updater
	relations := []orm.RelationSpec{
		{
			Model:         a.auditUserModel,
			Alias:         "creator",
			JoinType:      orm.LeftJoin,
			ForeignColumn: "created_by",
			SelectedColumns: []orm.ColumnInfo{
				{
					Name:  a.auditUserNameColumn,
					Alias: constants.ColumnCreatedByName,
				},
			},
		},
		{
			Model:         a.auditUserModel,
			Alias:         "updater",
			JoinType:      orm.LeftJoin,
			ForeignColumn: "updated_by",
			SelectedColumns: []orm.ColumnInfo{
				{
					Name:  a.auditUserNameColumn,
					Alias: constants.ColumnUpdatedByName,
				},
			},
		},
	}

	query.JoinRelations(relations...)
}

// ApplyDataPermission applies data permission filtering to the query.
// This method retrieves the DataPermissionApplier from context and applies it to the query.
// It should be called after Model() is set but before other conditions are applied.
// Data permission filtering is enabled by default and can be disabled via DisableDataPerm().
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ApplyDataPermission(query orm.SelectQuery, ctx fiber.Ctx) {
	// Skip if data permission is explicitly disabled
	if a.disableDataPerm {
		return
	}

	applier := contextx.DataPermApplier(ctx)
	if applier == nil {
		return
	}

	if err := applier.Apply(query); err != nil {
		// Log error but don't fail the request
		// The error is already logged by the applier
		contextx.Logger(ctx).Errorf("Failed to apply data permission: %v", err)
	}
}

func applySort[TSearch any](query orm.SelectQuery, sortApplier SortApplier[TSearch], search TSearch, ctx fiber.Ctx) {
	query.ApplyIf(sortApplier != nil, func(query orm.SelectQuery) {
		sortApplier(search, ctx)(newSorter(query))
	})
}
