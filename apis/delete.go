package apis

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type deleteAPI[TModel any] struct {
	APIBuilder[DeleteAPI[TModel]]

	preDelete  PreDeleteProcessor[TModel]
	postDelete PostDeleteProcessor[TModel]
}

// Provide generates the final API specification for model deletion.
// Returns a complete api.Spec that can be registered with the router.
func (d *deleteAPI[TModel]) Provide() api.Spec {
	return d.APIBuilder.Build(d.delete)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (d *deleteAPI[TModel]) Build(handler any) api.Spec {
	panic("apis: do not call APIBuilder.Build on deleteAPI; call Provide() instead")
}

func (d *deleteAPI[TModel]) PreDelete(processor PreDeleteProcessor[TModel]) DeleteAPI[TModel] {
	d.preDelete = processor

	return d
}

func (d *deleteAPI[TModel]) PostDelete(processor PostDeleteProcessor[TModel]) DeleteAPI[TModel] {
	d.postDelete = processor

	return d
}

func (d *deleteAPI[TModel]) delete(db orm.Db) (func(ctx fiber.Ctx, db orm.Db) error, error) {
	// Pre-compute schema information
	schema := db.TableOf((*TModel)(nil))
	// Pre-compute primary key fields
	pks := db.ModelPKFields((*TModel)(nil))

	// Validate schema has primary keys
	if len(pks) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrModelNoPrimaryKey, schema.Name)
	}

	return func(ctx fiber.Ctx, db orm.Db) error {
		var (
			model      TModel
			modelValue = reflect.ValueOf(&model).Elem()
			req        = contextx.APIRequest(ctx)
		)

		// Extract and set primary key values using pre-computed metadata
		for _, pk := range pks {
			value, ok := req.Params[pk.Name]
			if !ok {
				return result.Err(i18n.T("primary_key_required", map[string]any{"field": pk.Name}))
			}

			if err := pk.Set(modelValue, value); err != nil {
				return err
			}
		}

		// Load the existing model
		if err := db.NewSelect().Model(&model).WherePK().Scan(ctx.Context(), &model); err != nil {
			return err
		}

		// Execute pre-delete hook if configured
		if d.preDelete != nil {
			if err := d.preDelete(&model, ctx, db); err != nil {
				return err
			}
		}

		// Execute delete operation within transaction
		return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
			if _, err := tx.NewDelete().Model(&model).WherePK().Exec(txCtx); err != nil {
				return err
			}

			// Execute post-delete hook if configured
			if d.postDelete != nil {
				if err := d.postDelete(&model, ctx, tx); err != nil {
					return err
				}
			}

			return result.Ok().Response(ctx)
		})
	}, nil
}
