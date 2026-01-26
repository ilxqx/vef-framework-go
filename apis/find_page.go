package apis

import (
	"reflect"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/go-streams"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/page"
	"github.com/ilxqx/vef-framework-go/result"
)

type findPageApi[TModel, TSearch any] struct {
	Find[TModel, TSearch, []TModel, FindPage[TModel, TSearch]]

	defaultPageSize int
}

func (a *findPageApi[TModel, TSearch]) Provide() []api.OperationSpec {
	return []api.OperationSpec{a.Build(a.findPage)}
}

// This value is used when the request's page size is zero or invalid.
func (a *findPageApi[TModel, TSearch]) WithDefaultPageSize(size int) FindPage[TModel, TSearch] {
	a.defaultPageSize = size

	return a
}

func (a *findPageApi[TModel, TSearch]) findPage(db orm.DB) (func(ctx fiber.Ctx, db orm.DB, transformer mold.Transformer, pageable page.Pageable, search TSearch, meta api.Meta) error, error) {
	if err := a.Setup(db, &FindApiConfig{
		QueryParts: &QueryPartsConfig{
			Condition:         []QueryPart{QueryRoot},
			Sort:              []QueryPart{QueryRoot},
			AuditUserRelation: []QueryPart{QueryRoot},
		},
	}); err != nil {
		return nil, err
	}

	return func(ctx fiber.Ctx, db orm.DB, transformer mold.Transformer, pageable page.Pageable, search TSearch, meta api.Meta) (err error) {
		pageable.Normalize(a.defaultPageSize)

		var (
			models []TModel
			query  = db.NewSelect().Model(&models).SelectModelColumns().Paginate(pageable)
			total  int64
		)

		if err = a.ConfigureQuery(query, search, meta, ctx, QueryRoot); err != nil {
			return err
		}

		if total, err = query.ScanAndCount(ctx.Context()); err != nil {
			return err
		}

		if total == 0 {
			return result.Ok(page.New(pageable, total, []any{})).Response(ctx)
		}

		if err := streams.Range(0, len(models)).ForEachErr(func(i int) error {
			return transformer.Struct(ctx.Context(), &models[i])
		}); err != nil {
			return err
		}

		processedModels := a.Process(models, search, ctx)
		if typedModels, ok := processedModels.([]TModel); ok {
			return result.Ok(page.New(pageable, total, typedModels)).Response(ctx)
		}

		modelsValue := reflect.Indirect(reflect.ValueOf(processedModels))
		if modelsValue.Kind() != reflect.Slice {
			return result.Err(
				i18n.T(ErrMessageProcessorMustReturnSlice, map[string]any{"type": reflect.TypeOf(processedModels).String()}),
				result.WithCode(ErrCodeProcessorInvalidReturn),
				result.WithStatus(fiber.StatusInternalServerError),
			)
		}

		items := make([]any, modelsValue.Len())
		for i := range items {
			items[i] = modelsValue.Index(i).Interface()
		}

		return result.Ok(page.New(pageable, total, items)).Response(ctx)
	}, nil
}
