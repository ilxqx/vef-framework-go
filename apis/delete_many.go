package apis

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/go-streams"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/storage"
)

type deleteManyApi[TModel any] struct {
	ApiBuilder[DeleteManyApi[TModel]]

	preDeleteMany    PreDeleteManyProcessor[TModel]
	postDeleteMany   PostDeleteManyProcessor[TModel]
	dataPermDisabled bool
}

// Provide generates the final Api specification for batch model deletion.
// Returns a complete api.Spec that can be registered with the router.
func (d *deleteManyApi[TModel]) Provide() api.Spec {
	return d.Build(d.deleteMany)
}

func (d *deleteManyApi[TModel]) WithPreDeleteMany(processor PreDeleteManyProcessor[TModel]) DeleteManyApi[TModel] {
	d.preDeleteMany = processor

	return d
}

func (d *deleteManyApi[TModel]) WithPostDeleteMany(processor PostDeleteManyProcessor[TModel]) DeleteManyApi[TModel] {
	d.postDeleteMany = processor

	return d
}

func (d *deleteManyApi[TModel]) DisableDataPerm() DeleteManyApi[TModel] {
	d.dataPermDisabled = true

	return d
}

func (d *deleteManyApi[TModel]) deleteMany(db orm.Db, sc storage.Service, publisher event.Publisher) (func(ctx fiber.Ctx, db orm.Db, params DeleteManyParams) error, error) {
	promoter := storage.NewPromoter[TModel](sc, publisher)
	schema := db.TableOf((*TModel)(nil))
	pks := db.ModelPkFields((*TModel)(nil))

	if len(pks) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrModelNoPrimaryKey, schema.Name)
	}

	return func(ctx fiber.Ctx, db orm.Db, params DeleteManyParams) error {
		if len(params.Pks) == 0 {
			return result.Ok().Response(ctx)
		}

		models := make([]TModel, len(params.Pks))

		if err := streams.Range(0, len(params.Pks)).ForEachErr(func(i int) error {
			pkValue := params.Pks[i]
			modelValue := reflect.ValueOf(&models[i]).Elem()

			// Try to interpret pkValue as a map first (works for both single and composite Pks)
			if pkMap, ok := pkValue.(map[string]any); ok {
				if err := streams.FromSlice(pks).ForEachErr(func(pk *orm.PkField) error {
					value, ok := pkMap[pk.Name]
					if !ok {
						return result.Err(i18n.T("primary_key_required", map[string]any{"field": pk.Name}))
					}

					return pk.Set(modelValue, value)
				}); err != nil {
					return err
				}
			} else {
				// Direct value format - only valid for single primary key
				if len(pks) != 1 {
					return result.Err(i18n.T("composite_primary_key_requires_map"))
				}

				if err := pks[0].Set(modelValue, pkValue); err != nil {
					return err
				}
			}

			query := db.NewSelect().Model(&models[i]).WherePk()
			if !d.dataPermDisabled {
				if err := ApplyDataPermission(query, ctx); err != nil {
					return err
				}
			}

			return query.Scan(ctx.Context(), &models[i])
		}); err != nil {
			return err
		}

		return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
			query := tx.NewDelete().Model(&models)
			if d.preDeleteMany != nil {
				if err := d.preDeleteMany(models, query, ctx, tx); err != nil {
					return err
				}
			}

			if _, err := query.WherePk().Exec(txCtx); err != nil {
				return err
			}

			if d.postDeleteMany != nil {
				if err := d.postDeleteMany(models, ctx, tx); err != nil {
					return err
				}
			}

			if cleanupErr := batchCleanup(txCtx, promoter, models); cleanupErr != nil {
				return fmt.Errorf("delete succeeded but cleanup files failed: %w", cleanupErr)
			}

			return result.Ok().Response(ctx)
		})
	}, nil
}
