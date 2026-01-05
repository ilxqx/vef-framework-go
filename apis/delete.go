package apis

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/go-streams"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/storage"
)

type deleteApi[TModel any] struct {
	ApiBuilder[DeleteApi[TModel]]

	preDelete        PreDeleteProcessor[TModel]
	postDelete       PostDeleteProcessor[TModel]
	dataPermDisabled bool
}

// Provide generates the final Api specification for model deletion.
// Returns a complete api.Spec that can be registered with the router.
func (d *deleteApi[TModel]) Provide() api.Spec {
	return d.Build(d.delete)
}

func (d *deleteApi[TModel]) WithPreDelete(processor PreDeleteProcessor[TModel]) DeleteApi[TModel] {
	d.preDelete = processor

	return d
}

func (d *deleteApi[TModel]) WithPostDelete(processor PostDeleteProcessor[TModel]) DeleteApi[TModel] {
	d.postDelete = processor

	return d
}

func (d *deleteApi[TModel]) DisableDataPerm() DeleteApi[TModel] {
	d.dataPermDisabled = true

	return d
}

func (d *deleteApi[TModel]) delete(db orm.Db, sc storage.Service, publisher event.Publisher) (func(ctx fiber.Ctx, db orm.Db) error, error) {
	promoter := storage.NewPromoter[TModel](sc, publisher)
	schema := db.TableOf((*TModel)(nil))
	pks := db.ModelPkFields((*TModel)(nil))

	if len(pks) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrModelNoPrimaryKey, schema.Name)
	}

	return func(ctx fiber.Ctx, db orm.Db) error {
		var (
			model      TModel
			modelValue = reflect.ValueOf(&model).Elem()
			req        = contextx.ApiRequest(ctx)
		)

		if err := streams.FromSlice(pks).ForEachErr(func(pk *orm.PkField) error {
			value, ok := req.Params[pk.Name]
			if !ok {
				return result.Err(i18n.T("primary_key_required", map[string]any{"field": pk.Name}))
			}

			return pk.Set(modelValue, value)
		}); err != nil {
			return err
		}

		query := db.NewSelect().Model(&model).WherePk()
		if !d.dataPermDisabled {
			if err := ApplyDataPermission(query, ctx); err != nil {
				return err
			}
		}

		if err := query.Scan(ctx.Context(), &model); err != nil {
			return err
		}

		return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
			query := tx.NewDelete().Model(&model)
			if d.preDelete != nil {
				if err := d.preDelete(&model, query, ctx, tx); err != nil {
					return err
				}
			}

			if _, err := query.WherePk().Exec(txCtx); err != nil {
				return err
			}

			if d.postDelete != nil {
				if err := d.postDelete(&model, ctx, tx); err != nil {
					return err
				}
			}

			if err := promoter.Promote(txCtx, nil, &model); err != nil {
				return fmt.Errorf("delete succeeded but cleanup files failed: %w", err)
			}

			return result.Ok().Response(ctx)
		})
	}, nil
}
