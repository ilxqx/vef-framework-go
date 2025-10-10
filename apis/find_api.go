package apis

import (
	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/orm"
)

type baseFindAPI[TModel, TSearch, TProcessorIn, TAPI any] struct {
	APIBuilder[TAPI]

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

	self TAPI
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) DisableDataPerm() TAPI {
	a.disableDataPerm = true

	return a.self
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) WithAuditUserNames(userModel any, nameColumn ...string) TAPI {
	a.auditUserModel = userModel
	if len(nameColumn) > 0 && nameColumn[0] != constants.Empty {
		a.auditUserNameColumn = nameColumn[0]
	} else {
		a.auditUserNameColumn = "name"
	}

	return a.self
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) QueryApplier(applier QueryApplier[TSearch]) TAPI {
	a.queryApplier = applier

	return a.self
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) FilterApplier(applier FilterApplier[TSearch]) TAPI {
	a.filterApplier = applier

	return a.self
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) SortApplier(applier SortApplier[TSearch]) TAPI {
	a.sortApplier = applier

	return a.self
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) Relations(relations ...orm.RelationSpec) TAPI {
	a.relations = append(a.relations, relations...)

	return a.self
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) Processor(processor Processor[TProcessorIn, TSearch]) TAPI {
	a.processor = processor

	return a.self
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) Init(db orm.Db) error {
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

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ShouldApplyDefaultSort() bool {
	return a.shouldApplyDefaultSort
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ApplyDefaultSort(query orm.SelectQuery) {
	if a.shouldApplyDefaultSort {
		query.OrderByDesc(constants.ColumnCreatedAt)
	}
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) BuildQuery(db orm.Db, model any, search TSearch, ctx fiber.Ctx) orm.SelectQuery {
	query := db.NewSelect()
	a.ConfigureQuery(query, model, search, ctx)

	return query
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ConfigureQuery(query orm.SelectQuery, model any, search TSearch, ctx fiber.Ctx) {
	// Set the model first (required for data permission)
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

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) Process(input TProcessorIn, search TSearch, ctx fiber.Ctx) any {
	if a.processor == nil {
		return input
	}

	return a.processor(input, search, ctx)
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ApplySort(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	applySort(query, a.sortApplier, search, ctx)
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ApplySearch(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	query.Where(func(cb orm.ConditionBuilder) {
		cb.Apply(a.searchApplier(search))
	})
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ApplyFilter(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	query.Where(func(cb orm.ConditionBuilder) {
		cb.Apply(a.filterApplier(search, ctx))
	})
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ApplyConditions(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	query.Where(func(cb orm.ConditionBuilder) {
		cb.Apply(a.searchApplier(search))

		if a.filterApplier != nil {
			cb.Apply(a.filterApplier(search, ctx))
		}
	})
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ApplyQuery(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	if a.queryApplier != nil {
		query.Apply(a.queryApplier(search, ctx))
	}
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ApplyRelations(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) {
	if len(a.relations) > 0 {
		query.JoinRelations(a.relations...)
	}
}

// ApplyAuditUserRelations applies RelationSpec configurations to query audit user names (created_by_name, updated_by_name).
// This method creates LEFT JOIN relations with the user model to populate audit user name fields.
// It automatically constructs two RelationSpecs: one for the creator and one for the updater.
// The method is called automatically during query building if WithAuditUserNames was configured.
func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ApplyAuditUserRelations(query orm.SelectQuery) {
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
func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ApplyDataPermission(query orm.SelectQuery, ctx fiber.Ctx) {
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
