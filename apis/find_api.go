package apis

import (
	"slices"

	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/sort"
)

// baseFindApi is the base implementation for all Find APIs.
// It provides a unified query configuration system using FindApiOptions.
type baseFindApi[TModel, TSearch, TProcessorIn, TApi any] struct {
	ApiBuilder[TApi]

	setupDone           bool
	dataPermDisabled    bool
	options             []*FindApiOption
	optionsByPart       map[QueryPart][]*FindApiOption
	auditUserModel      any
	auditUserNameColumn string
	defaultSort         []*sort.OrderSpec
	processor           Processor[TProcessorIn, TSearch]

	self TApi
}

// Setup initializes the FindApi with database and configuration.
// This method is called once in factory functions and is safe to call multiple times.
// Subsequent calls are no-ops.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) Setup(db orm.Db, config *FindApiConfig, opts ...*FindApiOption) error {
	if a.setupDone {
		return nil
	}
	defer func() { a.setupDone = true }()

	if config != nil && config.QueryParts != nil {
		var (
			qp    = config.QueryParts
			table = db.TableOf((*TModel)(nil))
		)

		opt := withSearchApplier[TSearch](
			lo.Ternary(qp.Condition != nil, qp.Condition, []QueryPart{QueryRoot})...,
		)
		a.options = append(a.options, opt)

		// Auto-apply data permission filtering unless disabled
		if !a.dataPermDisabled {
			opt := withDataPerm(lo.Ternary(qp.Condition != nil, qp.Condition, []QueryPart{QueryRoot})...)
			a.options = append(a.options, opt)
		}

		// Validate and apply audit user model if configured
		if a.auditUserModel != nil {
			pkLen := len(table.PKs)
			if pkLen == 0 {
				return ErrModelNoPrimaryKey
			}

			if pkLen > 1 {
				return ErrAuditUserCompositePk
			}

			opt := withAuditUserNames(
				a.auditUserModel,
				a.auditUserNameColumn,
				lo.Ternary(qp.AuditUserRelation != nil, qp.AuditUserRelation, []QueryPart{QueryRoot})...,
			)
			a.options = append(a.options, opt)
		}

		if a.defaultSort == nil {
			if len(table.PKs) == 1 {
				opt := withSort(
					[]*sort.OrderSpec{
						{
							Column:    table.PKs[0].Name,
							Direction: sort.OrderDesc,
						},
					},
					lo.Ternary(qp.Sort != nil, qp.Sort, []QueryPart{QueryRoot})...,
				)
				a.options = append(a.options, opt)
			} else {
				if field, ok := table.FieldMap[constants.ColumnCreatedAt]; ok {
					opt := withSort(
						[]*sort.OrderSpec{
							{
								Column:    field.Name,
								Direction: sort.OrderDesc,
							},
						},
						lo.Ternary(qp.Sort != nil, qp.Sort, []QueryPart{QueryRoot})...,
					)
					a.options = append(a.options, opt)
				}
			}
		} else if len(a.defaultSort) > 0 {
			// User-defined default sorting
			opt := withSort(a.defaultSort, lo.Ternary(qp.Sort != nil, qp.Sort, []QueryPart{QueryRoot})...)
			a.options = append(a.options, opt)
		}
	}

	a.options = append(a.options, opts...)

	// Pre-group options by QueryPart for efficient lookup in ConfigureQuery
	a.optionsByPart = make(map[QueryPart][]*FindApiOption)
	for _, opt := range a.options {
		for _, part := range opt.Parts {
			a.optionsByPart[part] = append(a.optionsByPart[part], opt)
		}
	}

	return nil
}

// ConfigureQuery applies all query configuration options for the specified query part.
// This is the core method that processes FindApiOptions and applies them to the query.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) ConfigureQuery(query orm.SelectQuery, search TSearch, ctx fiber.Ctx, part QueryPart) error {
	// Use pre-grouped options for efficient lookup
	// Track applied options to avoid duplicates when an option targets both specific part and QueryAll
	applied := make(map[*FindApiOption]bool)

	for _, opt := range a.optionsByPart[part] {
		if !applied[opt] {
			if err := opt.Applier(query, search, ctx); err != nil {
				return err
			}

			applied[opt] = true
		}
	}

	// Apply options for QueryAll (applies to all parts)
	for _, opt := range a.optionsByPart[QueryAll] {
		if !applied[opt] {
			if err := opt.Applier(query, search, ctx); err != nil {
				return err
			}

			applied[opt] = true
		}
	}

	return nil
}

// Process applies post-query processing to transform or enrich the query results.
// This method is called after data is fetched from the database but before returning to the client.
// If no Processor is configured via WithProcessor(), it returns the input unchanged.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) Process(input TProcessorIn, search TSearch, ctx fiber.Ctx) any {
	if a.processor == nil {
		return input
	}

	return a.processor(input, search, ctx)
}

// This function is called after data is fetched from the database but before returning to the client.
// Common use cases: data masking, computed fields, nested structure transformation, aggregation.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) WithProcessor(processor Processor[TProcessorIn, TSearch]) TApi {
	a.processor = processor

	return a.self
}

// WithOptions adds multiple FindApiOptions to the query configuration.
// This is useful for composing reusable option sets.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) WithOptions(opts ...*FindApiOption) TApi {
	a.options = append(a.options, opts...)

	return a.self
}

// WithSelect adds a column to the SELECT clause.
// Applies to all query parts by default (QueryAll) unless specific parts are provided.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) WithSelect(column string, parts ...QueryPart) TApi {
	a.options = append(a.options, withSelect(column, parts...))

	return a.self
}

// WithSelectAs adds a column with an alias to the SELECT clause.
// Applies to all query parts by default (QueryAll) unless specific parts are provided.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) WithSelectAs(column, alias string, parts ...QueryPart) TApi {
	a.options = append(a.options, withSelectAs(column, alias, parts...))

	return a.self
}

// This is applied when no dynamic sorting is provided in the request.
// The orderSpecs are stored and applied during Setup() to allow framework-level defaults.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) WithDefaultSort(orderSpecs ...*sort.OrderSpec) TApi {
	if len(orderSpecs) > 0 {
		a.defaultSort = slices.Clone(orderSpecs)
	} else {
		a.defaultSort = []*sort.OrderSpec{}
	}

	return a.self
}

// DisableDataPerm disables data permission filtering for this API.
// By default, data permission filtering is enabled (WithDataPerm is auto-applied in Setup).
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) DisableDataPerm() TApi {
	a.dataPermDisabled = true

	return a.self
}

// WithAuditUserNames configures audit user names to be fetched (created_by_name, updated_by_name).
// If nameColumn is provided, uses the first value; otherwise defaults to "name".
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) WithAuditUserNames(userModel any, nameColumn ...string) TApi {
	a.auditUserModel = userModel
	if len(nameColumn) > 0 {
		a.auditUserNameColumn = nameColumn[0]
	} else {
		a.auditUserNameColumn = defaultAuditUserNameColumn
	}

	return a.self
}

// WithCondition adds a WHERE condition using ConditionBuilder.
// Applies to root query only by default (QueryRoot) unless specific parts are provided.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) WithCondition(fn func(cb orm.ConditionBuilder), parts ...QueryPart) TApi {
	a.options = append(a.options, withCondition(fn, parts...))

	return a.self
}

// WithRelation adds a relation join to the query.
// Applies to all query parts by default (QueryAll) unless specific parts are provided.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) WithRelation(relation *orm.RelationSpec, parts ...QueryPart) TApi {
	a.options = append(a.options, withRelation(relation, parts...))

	return a.self
}

// WithQueryApplier adds a custom query applier function.
// Applies to root query only by default (QueryRoot) unless specific parts are provided.
func (a *baseFindApi[TModel, TSearch, TProcessorIn, TApi]) WithQueryApplier(applier func(query orm.SelectQuery, search TSearch, ctx fiber.Ctx) error, parts ...QueryPart) TApi {
	a.options = append(a.options, withQueryApplier(applier, parts...))

	return a.self
}
