package apis

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/copier"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/storage"
)

type updateApi[TModel, TParams any] struct {
	Builder[Update[TModel, TParams]]

	preUpdate        PreUpdateProcessor[TModel, TParams]
	postUpdate       PostUpdateProcessor[TModel, TParams]
	dataPermDisabled bool
}

func (u *updateApi[TModel, TParams]) Provide() []api.OperationSpec {
	return []api.OperationSpec{u.Build(u.update)}
}

func (u *updateApi[TModel, TParams]) WithPreUpdate(processor PreUpdateProcessor[TModel, TParams]) Update[TModel, TParams] {
	u.preUpdate = processor

	return u
}

func (u *updateApi[TModel, TParams]) WithPostUpdate(processor PostUpdateProcessor[TModel, TParams]) Update[TModel, TParams] {
	u.postUpdate = processor

	return u
}

func (u *updateApi[TModel, TParams]) DisableDataPerm() Update[TModel, TParams] {
	u.dataPermDisabled = true

	return u
}

func (u *updateApi[TModel, TParams]) update(db orm.DB, sc storage.Service, publisher event.Publisher) (func(ctx fiber.Ctx, db orm.DB, params TParams) error, error) {
	promoter := storage.NewPromoter[TModel](sc, publisher)
	schema := db.TableOf((*TModel)(nil))
	pks := db.ModelPKFields((*TModel)(nil))

	if len(pks) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrModelNoPrimaryKey, schema.Name)
	}

	return func(ctx fiber.Ctx, db orm.DB, params TParams) error {
		var (
			oldModel   TModel
			model      TModel
			modelValue = reflect.ValueOf(&model).Elem()
		)

		if err := copier.Copy(&params, &model); err != nil {
			return err
		}

		for _, pk := range pks {
			pkValue, err := pk.Value(modelValue)
			if err != nil {
				return err
			}

			if reflect.ValueOf(pkValue).IsZero() {
				return result.Err(i18n.T("primary_key_required", map[string]any{"field": pk.Name}))
			}
		}

		query := db.NewSelect().Model(&model).WherePK()
		if !u.dataPermDisabled {
			if err := ApplyDataPermission(query, ctx); err != nil {
				return err
			}
		}

		if err := query.Scan(ctx.Context(), &oldModel); err != nil {
			return err
		}

		return db.RunInTX(ctx.Context(), func(txCtx context.Context, tx orm.DB) error {
			query := tx.NewUpdate().Model(&oldModel)
			if u.preUpdate != nil {
				if err := u.preUpdate(&oldModel, &model, &params, query, ctx, tx); err != nil {
					return err
				}
			}

			if err := copier.Copy(&model, &oldModel, copier.WithIgnoreEmpty()); err != nil {
				return err
			}

			if err := promoter.Promote(txCtx, &oldModel, &model); err != nil {
				return fmt.Errorf("promote files failed: %w", err)
			}

			if _, err := query.WherePK().Exec(txCtx); err != nil {
				if cleanupErr := promoter.Promote(txCtx, &model, &oldModel); cleanupErr != nil {
					return fmt.Errorf("update failed: %w; rollback files also failed: %w", err, cleanupErr)
				}

				return err
			}

			if u.postUpdate != nil {
				if err := u.postUpdate(&oldModel, &model, &params, ctx, tx); err != nil {
					if cleanupErr := promoter.Promote(txCtx, &model, &oldModel); cleanupErr != nil {
						return fmt.Errorf("post-update failed: %w; rollback files also failed: %w", err, cleanupErr)
					}

					return err
				}
			}

			return result.Ok().Response(ctx)
		})
	}, nil
}
