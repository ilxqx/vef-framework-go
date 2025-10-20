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

type deleteApi[TModel any] struct {
	ApiBuilder[DeleteApi[TModel]]

	preDelete       PreDeleteProcessor[TModel]
	postDelete      PostDeleteProcessor[TModel]
	disableDataPerm bool
}

// Provide generates the final Api specification for model deletion.
// Returns a complete api.Spec that can be registered with the router.
func (d *deleteApi[TModel]) Provide() api.Spec {
	return d.Build(d.delete)
}

func (d *deleteApi[TModel]) PreDelete(processor PreDeleteProcessor[TModel]) DeleteApi[TModel] {
	d.preDelete = processor

	return d
}

func (d *deleteApi[TModel]) PostDelete(processor PostDeleteProcessor[TModel]) DeleteApi[TModel] {
	d.postDelete = processor

	return d
}

func (d *deleteApi[TModel]) DisableDataPerm() DeleteApi[TModel] {
	d.disableDataPerm = true

	return d
}

func (d *deleteApi[TModel]) delete(db orm.Db) (func(ctx fiber.Ctx, db orm.Db) error, error) {
	// Pre-compute schema information
	schema := db.TableOf((*TModel)(nil))
	// Pre-compute primary key fields
	pks := db.ModelPkFields((*TModel)(nil))

	// Validate schema has primary keys
	if len(pks) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrModelNoPrimaryKey, schema.Name)
	}

	return func(ctx fiber.Ctx, db orm.Db) error {
		var (
			model      TModel
			modelValue = reflect.ValueOf(&model).Elem()
			req        = contextx.ApiRequest(ctx)
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

		// Build query with data permission filtering
		query := db.NewSelect().Model(&model).WherePk()
		if !d.disableDataPerm {
			if err := applyDataPermission(query, ctx); err != nil {
				return err
			}
		}

		// Load the existing model
		if err := query.Scan(ctx.Context(), &model); err != nil {
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
			if _, err := tx.NewDelete().Model(&model).WherePk().Exec(txCtx); err != nil {
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
