package apis

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/contextx"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

// DeleteAPI provides delete functionality with pre/post processing hooks.
// It supports method chaining for configuration and handles transaction management automatically.
// The model is identified by its primary key(s) extracted from the request parameters.
//
// Type parameters:
//   - TModel: The database model type
type DeleteAPI[TModel any] struct {
	preDelete  PreDeleteProcessor[TModel]
	postDelete PostDeleteProcessor[TModel]
}

// WithPreDelete sets the pre-delete processor for the DeleteAPI.
// This processor is called before the model is deleted from the database.
// Returns the API instance for method chaining.
func (d *DeleteAPI[TModel]) WithPreDelete(processor PreDeleteProcessor[TModel]) *DeleteAPI[TModel] {
	d.preDelete = processor
	return d
}

// WithPostDelete sets the post-delete processor for the DeleteAPI.
// This processor is called after the model is successfully deleted within the same transaction.
// Returns the API instance for method chaining.
func (d *DeleteAPI[TModel]) WithPostDelete(processor PostDeleteProcessor[TModel]) *DeleteAPI[TModel] {
	d.postDelete = processor
	return d
}

// Delete deletes a model by primary key with pre/post processing hooks.
// The primary key values are extracted from the request parameters using camelCase field names.
// The operation is executed within a database transaction for data consistency.
//
// Parameters:
//   - db: The database connection for schema introspection
//
// Returns a handler function that processes delete requests.
func (d *DeleteAPI[TModel]) Delete(db orm.Db) (func(ctx fiber.Ctx, db orm.Db) error, error) {
	// Pre-compute schema information
	schema := db.Schema((*TModel)(nil))

	// Validate schema has primary keys
	if len(schema.PKs) == 0 {
		return nil, fmt.Errorf("model '%s' has no primary key", schema.Name)
	}

	// Pre-compute primary key metadata
	type pkInfo struct {
		field     *orm.Field
		paramName string
		kind      reflect.Kind
	}

	pkInfos := make([]pkInfo, len(schema.PKs))
	for i, pk := range schema.PKs {
		pkInfos[i] = pkInfo{
			field:     pk,
			paramName: lo.CamelCase(pk.GoName),
			kind:      pk.IndirectType.Kind(),
		}
	}

	return func(ctx fiber.Ctx, db orm.Db) error {
		var (
			model      TModel
			modelValue = reflect.ValueOf(&model).Elem()
			req        = contextx.APIRequest(ctx)
		)

		// Extract and set primary key values using pre-computed metadata
		for _, info := range pkInfos {
			value, ok := req.Params[info.paramName]
			if !ok {
				return result.Errf("primary key parameter '%s' is required", info.paramName)
			}

			pkValue := info.field.Value(modelValue)

			switch info.kind {
			case reflect.String:
				if info.field.IsPtr {
					pkValue.Set(reflect.ValueOf(&value))
				} else {
					pkValue.SetString(cast.ToString(value))
				}
			case reflect.Int, reflect.Int64:
				intValue := cast.ToInt64(value)
				if info.field.IsPtr {
					pkValue.Set(reflect.ValueOf(&intValue))
				} else {
					pkValue.SetInt(intValue)
				}
			}
		}

		// Load the existing model
		if err := db.NewQuery().Model(&model).WherePK().Scan(ctx, &model); err != nil {
			return err
		}

		// Execute pre-delete hook if configured
		if d.preDelete != nil {
			if err := d.preDelete(&model, ctx, db); err != nil {
				return err
			}
		}

		// Execute delete operation within transaction
		return db.RunInTx(ctx, func(txCtx context.Context, tx orm.Db) error {
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

// NewDeleteAPI creates a new DeleteAPI instance.
// Use method chaining to configure pre/post processing hooks.
//
// Example:
//
//	api := NewDeleteAPI[User]().
//	  WithPreDelete(checkPermissions).
//	  WithPostDelete(auditDelete)
func NewDeleteAPI[TModel any]() *DeleteAPI[TModel] {
	return new(DeleteAPI[TModel])
}
