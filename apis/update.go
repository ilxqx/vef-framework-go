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

type updateAPI[TModel, TParams any] struct {
	APIBuilder[UpdateAPI[TModel, TParams]]

	preUpdate  PreUpdateProcessor[TModel, TParams]
	postUpdate PostUpdateProcessor[TModel, TParams]
}

// Provide generates the final API specification for model updates.
// Returns a complete api.Spec that can be registered with the router.
func (u *updateAPI[TModel, TParams]) Provide() api.Spec {
	return u.APIBuilder.Build(u.update)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (u *updateAPI[TModel, TParams]) Build(handler any) api.Spec {
	panic("apis: do not call APIBuilder.Build on updateAPI; call Provide() instead")
}

func (u *updateAPI[TModel, TParams]) PreUpdate(processor PreUpdateProcessor[TModel, TParams]) UpdateAPI[TModel, TParams] {
	u.preUpdate = processor
	return u
}

func (u *updateAPI[TModel, TParams]) PostUpdate(processor PostUpdateProcessor[TModel, TParams]) UpdateAPI[TModel, TParams] {
	u.postUpdate = processor
	return u
}

func (u *updateAPI[TModel, TParams]) update(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, params TParams) error, error) {
	// Pre-compute schema information
	schema := db.Schema((*TModel)(nil))
	// Pre-compute primary key fields
	pks := db.ModelPKFields((*TModel)(nil))

	// Validate schema has primary keys
	if len(pks) == 0 {
		return nil, fmt.Errorf("model '%s' has no primary key", schema.Name)
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

		if err := db.NewSelect().Model(&model).WherePK().Scan(ctx, &oldModel); err != nil {
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

		return db.RunInTx(ctx, func(txCtx context.Context, tx orm.Db) error {
			if _, err := tx.NewUpdate().Model(&oldModel).WherePK().Exec(txCtx); err != nil {
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
