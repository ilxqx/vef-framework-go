package apis

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type deleteManyAPI[TModel any] struct {
	APIBuilder[DeleteManyAPI[TModel]]

	preDeleteMany   PreDeleteManyProcessor[TModel]
	postDeleteMany  PostDeleteManyProcessor[TModel]
	disableDataPerm bool
}

// Provide generates the final API specification for batch model deletion.
// Returns a complete api.Spec that can be registered with the router.
func (d *deleteManyAPI[TModel]) Provide() api.Spec {
	return d.APIBuilder.Build(d.deleteMany)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (d *deleteManyAPI[TModel]) Build(handler any) api.Spec {
	panic("apis: do not call APIBuilder.Build on deleteManyAPI; call Provide() instead")
}

func (d *deleteManyAPI[TModel]) PreDeleteMany(processor PreDeleteManyProcessor[TModel]) DeleteManyAPI[TModel] {
	d.preDeleteMany = processor

	return d
}

func (d *deleteManyAPI[TModel]) PostDeleteMany(processor PostDeleteManyProcessor[TModel]) DeleteManyAPI[TModel] {
	d.postDeleteMany = processor

	return d
}

func (d *deleteManyAPI[TModel]) DisableDataPerm() DeleteManyAPI[TModel] {
	d.disableDataPerm = true

	return d
}

func (d *deleteManyAPI[TModel]) deleteMany(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, params DeleteManyParams) error, error) {
	// Pre-compute schema information
	schema := db.TableOf((*TModel)(nil))
	// Pre-compute primary key fields
	pks := db.ModelPKFields((*TModel)(nil))

	// Validate schema has primary keys
	if len(pks) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrModelNoPrimaryKey, schema.Name)
	}

	return func(ctx fiber.Ctx, db orm.Db, params DeleteManyParams) error {
		if len(params.PKs) == 0 {
			return result.Ok().Response(ctx)
		}

		models := make([]TModel, len(params.PKs))

		// Process each primary key value
		for i, pkValue := range params.PKs {
			modelValue := reflect.ValueOf(&models[i]).Elem()

			// Try to interpret pkValue as a map first (works for both single and composite PKs)
			if pkMap, ok := pkValue.(map[string]any); ok {
				// Map format - set each PK field from the map
				for _, pk := range pks {
					value, ok := pkMap[pk.Name]
					if !ok {
						return result.Err(i18n.T("primary_key_required", map[string]any{"field": pk.Name}))
					}

					if err := pk.Set(modelValue, value); err != nil {
						return err
					}
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

			// Build query with data permission filtering
			query := db.NewSelect().Model(&models[i]).WherePK()
			if !d.disableDataPerm {
				if err := applyDataPermission(query, ctx); err != nil {
					return err
				}
			}

			// Load the existing model
			if err := query.Scan(ctx.Context(), &models[i]); err != nil {
				return err
			}
		}

		// Execute pre-delete hook if configured
		if d.preDeleteMany != nil {
			if err := d.preDeleteMany(models, ctx, db); err != nil {
				return err
			}
		}

		// Execute delete operation within transaction
		return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
			for i := range models {
				if _, err := tx.NewDelete().Model(&models[i]).WherePK().Exec(txCtx); err != nil {
					return err
				}
			}

			// Execute post-delete hook if configured
			if d.postDeleteMany != nil {
				if err := d.postDeleteMany(models, ctx, tx); err != nil {
					return err
				}
			}

			return result.Ok().Response(ctx)
		})
	}, nil
}
