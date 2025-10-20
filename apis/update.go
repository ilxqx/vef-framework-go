package apis

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/copier"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type updateApi[TModel, TParams any] struct {
	ApiBuilder[UpdateApi[TModel, TParams]]

	preUpdate       PreUpdateProcessor[TModel, TParams]
	postUpdate      PostUpdateProcessor[TModel, TParams]
	disableDataPerm bool
}

// Provide generates the final Api specification for model updates.
// Returns a complete api.Spec that can be registered with the router.
func (u *updateApi[TModel, TParams]) Provide() api.Spec {
	return u.Build(u.update)
}

func (u *updateApi[TModel, TParams]) PreUpdate(processor PreUpdateProcessor[TModel, TParams]) UpdateApi[TModel, TParams] {
	u.preUpdate = processor

	return u
}

func (u *updateApi[TModel, TParams]) PostUpdate(processor PostUpdateProcessor[TModel, TParams]) UpdateApi[TModel, TParams] {
	u.postUpdate = processor

	return u
}

func (u *updateApi[TModel, TParams]) DisableDataPerm() UpdateApi[TModel, TParams] {
	u.disableDataPerm = true

	return u
}

func (u *updateApi[TModel, TParams]) update(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, params TParams) error, error) {
	// Pre-compute schema information
	schema := db.TableOf((*TModel)(nil))
	// Pre-compute primary key fields
	pks := db.ModelPkFields((*TModel)(nil))

	// Validate schema has primary keys
	if len(pks) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrModelNoPrimaryKey, schema.Name)
	}

	return func(ctx fiber.Ctx, db orm.Db, params TParams) error {
		var (
			oldModel   TModel
			model      TModel
			modelValue = reflect.ValueOf(&model).Elem()
		)

		if err := copier.Copy(&params, &model); err != nil {
			return err
		}

		// Validate primary key is not empty
		for _, pk := range pks {
			pkValue, err := pk.Value(modelValue)
			if err != nil {
				return err
			}

			if reflect.ValueOf(pkValue).IsZero() {
				return result.Err(i18n.T("primary_key_required", map[string]any{"field": pk.Name}))
			}
		}

		// Build query with data permission filtering
		query := db.NewSelect().Model(&model).WherePk()
		if !u.disableDataPerm {
			if err := applyDataPermission(query, ctx); err != nil {
				return err
			}
		}

		if err := query.Scan(ctx.Context(), &oldModel); err != nil {
			return err
		}

		if u.preUpdate != nil {
			if err := u.preUpdate(&oldModel, &model, &params, ctx, db); err != nil {
				return err
			}
		}

		if err := copier.Copy(&model, &oldModel, copier.WithIgnoreEmpty()); err != nil {
			return err
		}

		return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
			if _, err := tx.NewUpdate().Model(&oldModel).WherePk().Exec(txCtx); err != nil {
				return err
			}

			if u.postUpdate != nil {
				if err := u.postUpdate(&oldModel, &model, &params, ctx, tx); err != nil {
					return err
				}
			}

			return result.Ok().Response(ctx)
		})
	}, nil
}
