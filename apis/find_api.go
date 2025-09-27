package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/orm"
)

type baseFindAPI[TModel, TSearch, TProcessorIn, TAPI any] struct {
	APIBuilder[TAPI]

	searchApplier SearchApplier[TSearch]
	filterApplier FilterApplier[TSearch]
	queryApplier  QueryApplier[TSearch]
	sortApplier   SortApplier[TSearch]
	relations     []orm.ModelRelation
	processor     Processor[TProcessorIn, TSearch]

	self TAPI
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

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) Relations(relations ...orm.ModelRelation) TAPI {
	a.relations = append(a.relations, relations...)
	return a.self
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) Processor(processor Processor[TProcessorIn, TSearch]) TAPI {
	a.processor = processor
	return a.self
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) BuildQuery(db orm.Db, model any, search TSearch, ctx fiber.Ctx) orm.SelectQuery {
	query := db.NewSelect()
	a.ConfigureQuery(query, model, search, ctx)
	return query
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) ConfigureQuery(query orm.SelectQuery, model any, search TSearch, ctx fiber.Ctx) {
	a.ApplyConditions(query.Model(model), search, ctx)
	a.ApplyRelations(query, search, ctx)
	a.ApplyQuery(query, search, ctx)
	a.ApplySort(query, search, ctx)
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) Process(input TProcessorIn, search TSearch, ctx fiber.Ctx) any {
	if a.processor == nil {
		return input
	}

	return a.processor(input, search, ctx)
}

func (a *baseFindAPI[TModel, TSearch, TProcessorIn, TAPI]) HasSortApplier() bool {
	return a.sortApplier != nil
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
		query.ModelRelations(a.relations...)
	}
}

func applySort[TSearch any](query orm.SelectQuery, sortApplier SortApplier[TSearch], search TSearch, ctx fiber.Ctx) {
	query.ApplyIf(sortApplier != nil, func(query orm.SelectQuery) {
		sortApplier(search, ctx)(newSorter(query))
	})
}
