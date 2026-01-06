package apis

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/go-streams"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/copier"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/storage"
)

type updateManyApi[TModel, TParams any] struct {
	ApiBuilder[UpdateManyApi[TModel, TParams]]

	preUpdateMany    PreUpdateManyProcessor[TModel, TParams]
	postUpdateMany   PostUpdateManyProcessor[TModel, TParams]
	dataPermDisabled bool
}

// Provide generates the final Api specification for batch model updates.
// Returns a complete api.Spec that can be registered with the router.
func (u *updateManyApi[TModel, TParams]) Provide() api.Spec {
	return u.Build(u.updateMany)
}

func (u *updateManyApi[TModel, TParams]) WithPreUpdateMany(processor PreUpdateManyProcessor[TModel, TParams]) UpdateManyApi[TModel, TParams] {
	u.preUpdateMany = processor

	return u
}

func (u *updateManyApi[TModel, TParams]) WithPostUpdateMany(processor PostUpdateManyProcessor[TModel, TParams]) UpdateManyApi[TModel, TParams] {
	u.postUpdateMany = processor

	return u
}

func (u *updateManyApi[TModel, TParams]) DisableDataPerm() UpdateManyApi[TModel, TParams] {
	u.dataPermDisabled = true

	return u
}

func (u *updateManyApi[TModel, TParams]) updateMany(db orm.Db, sc storage.Service, publisher event.Publisher) (func(ctx fiber.Ctx, db orm.Db, params UpdateManyParams[TParams]) error, error) {
	promoter := storage.NewPromoter[TModel](sc, publisher)
	schema := db.TableOf((*TModel)(nil))
	pks := db.ModelPkFields((*TModel)(nil))

	if len(pks) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrModelNoPrimaryKey, schema.Name)
	}

	return func(ctx fiber.Ctx, db orm.Db, params UpdateManyParams[TParams]) error {
		if len(params.List) == 0 {
			return result.Ok().Response(ctx)
		}

		oldModels := make([]TModel, len(params.List))
		models := make([]TModel, len(params.List))

		if err := streams.Range(0, len(params.List)).ForEachErr(func(i int) error {
			if err := copier.Copy(&params.List[i], &models[i]); err != nil {
				return err
			}

			modelValue := reflect.ValueOf(&models[i]).Elem()
			if err := streams.FromSlice(pks).ForEachErr(func(pk *orm.PkField) error {
				pkValue, err := pk.Value(modelValue)
				if err != nil {
					return err
				}

				if reflect.ValueOf(pkValue).IsZero() {
					return result.Err(i18n.T("primary_key_required", map[string]any{"field": pk.Name}))
				}

				return nil
			}); err != nil {
				return err
			}

			query := db.NewSelect().Model(&models[i]).WherePk()
			if !u.dataPermDisabled {
				if err := ApplyDataPermission(query, ctx); err != nil {
					return err
				}
			}

			if err := query.Scan(ctx.Context(), &oldModels[i]); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
			query := tx.NewUpdate().Model(&oldModels)

			if u.preUpdateMany != nil {
				if err := u.preUpdateMany(oldModels, models, params.List, query, ctx, tx); err != nil {
					return err
				}
			}

			if err := streams.Range(0, len(models)).ForEachErr(func(i int) error {
				return copier.Copy(&models[i], &oldModels[i], copier.WithIgnoreEmpty())
			}); err != nil {
				return err
			}

			if err := streams.Range(0, len(oldModels)).ForEachErr(func(i int) error {
				if err := promoter.Promote(txCtx, &oldModels[i], &models[i]); err != nil {
					if rollbackErr := batchRollback(txCtx, promoter, oldModels, models, i); rollbackErr != nil {
						return fmt.Errorf("promote files for model %d failed: %w; rollback also failed: %w", i, err, rollbackErr)
					}

					return fmt.Errorf("promote files for model %d failed: %w", i, err)
				}

				return nil
			}); err != nil {
				return err
			}

			if _, err := query.Bulk().Exec(txCtx); err != nil {
				if rollbackErr := batchRollback(txCtx, promoter, oldModels, models, len(oldModels)); rollbackErr != nil {
					return fmt.Errorf("batch update failed: %w; rollback files also failed: %w", err, rollbackErr)
				}

				return err
			}

			if u.postUpdateMany != nil {
				if err := u.postUpdateMany(oldModels, models, params.List, ctx, tx); err != nil {
					if rollbackErr := batchRollback(txCtx, promoter, oldModels, models, len(oldModels)); rollbackErr != nil {
						return fmt.Errorf("post-update-many failed: %w; rollback files also failed: %w", err, rollbackErr)
					}

					return err
				}
			}

			return result.Ok().Response(ctx)
		})
	}, nil
}
