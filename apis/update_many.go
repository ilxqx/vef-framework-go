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

type updateManyAPI[TModel, TParams any] struct {
	APIBuilder[UpdateManyAPI[TModel, TParams]]

	preUpdateMany  PreUpdateManyProcessor[TModel, TParams]
	postUpdateMany PostUpdateManyProcessor[TModel, TParams]
}

// Provide generates the final API specification for batch model updates.
// Returns a complete api.Spec that can be registered with the router.
func (u *updateManyAPI[TModel, TParams]) Provide() api.Spec {
	return u.APIBuilder.Build(u.updateMany)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (u *updateManyAPI[TModel, TParams]) Build(handler any) api.Spec {
	panic("apis: do not call APIBuilder.Build on updateManyAPI; call Provide() instead")
}

func (u *updateManyAPI[TModel, TParams]) PreUpdateMany(processor PreUpdateManyProcessor[TModel, TParams]) UpdateManyAPI[TModel, TParams] {
	u.preUpdateMany = processor
	return u
}

func (u *updateManyAPI[TModel, TParams]) PostUpdateMany(processor PostUpdateManyProcessor[TModel, TParams]) UpdateManyAPI[TModel, TParams] {
	u.postUpdateMany = processor
	return u
}

func (u *updateManyAPI[TModel, TParams]) updateMany(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, params UpdateManyParams[TParams]) error, error) {
	// Pre-compute schema information
	schema := db.TableOf((*TModel)(nil))
	// Pre-compute primary key fields
	pks := db.ModelPKFields((*TModel)(nil))

	// Validate schema has primary keys
	if len(pks) == 0 {
		return nil, fmt.Errorf("model '%s' has no primary key", schema.Name)
	}

	return func(ctx fiber.Ctx, db orm.Db, params UpdateManyParams[TParams]) error {
		if len(params.List) == 0 {
			return result.Ok().Response(ctx)
		}

		oldModels := make([]TModel, len(params.List))
		models := make([]TModel, len(params.List))

		// Copy params to models and validate primary keys
		for i := range params.List {
			if err := copier.Copy(&params.List[i], &models[i]); err != nil {
				return err
			}

			// Validate primary key is not empty
			modelValue := reflect.ValueOf(&models[i]).Elem()
			for _, pk := range pks {
				pkValue, err := pk.Value(modelValue)
				if err != nil {
					return err
				}

				if reflect.ValueOf(pkValue).IsZero() {
					return result.Err(i18n.T("primary_key_required", map[string]any{"field": pk.Name}))
				}
			}

			// Load existing model
			if err := db.NewSelect().Model(&models[i]).WherePK().Scan(ctx.Context(), &oldModels[i]); err != nil {
				return err
			}
		}

		if u.preUpdateMany != nil {
			if err := u.preUpdateMany(oldModels, models, params.List, ctx, db); err != nil {
				return err
			}
		}

		// Merge new values into old models
		for i := range models {
			if err := copier.Copy(&models[i], &oldModels[i], copier.WithIgnoreEmpty()); err != nil {
				return err
			}
		}

		return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
			for i := range oldModels {
				if _, err := tx.NewUpdate().Model(&oldModels[i]).WherePK().Exec(txCtx); err != nil {
					return err
				}
			}

			if u.postUpdateMany != nil {
				if err := u.postUpdateMany(oldModels, models, params.List, ctx, tx); err != nil {
					return err
				}
			}

			return result.Ok().Response(ctx)
		})
	}, nil
}
