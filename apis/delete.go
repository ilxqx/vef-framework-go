package apis

import (
	"context"
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
	preDelete  PreDeleteProcessor[TModel] // Function to execute before deleting the model
	postDelete PostDeleteProcessor[TModel] // Function to execute after deleting the model
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
//   - ctx: The Fiber context containing the primary key parameters
//   - db: The database connection
//
// Returns an error if the operation fails, otherwise returns success via HTTP response.
func (d *DeleteAPI[TModel]) Delete(ctx fiber.Ctx, db orm.Db) error {
	var (
		model      TModel
		modelValue = reflect.ValueOf(&model).Elem()
		schema     = db.Schema(&model)
		req        = contextx.APIRequest(ctx)
	)

	if len(schema.PKs) == 0 {
		return result.Errf("model %s has no primary key", schema.Name)
	}

	for _, pk := range schema.PKs {
		var (
			pkName    = lo.CamelCase(pk.GoName)
			value, ok = req.Params[pkName]
			pkValue   = pk.Value(modelValue)
			kind      = pk.IndirectType.Kind()
		)

		if !ok {
			return result.Errf("主键参数 %s 必须", pkName)
		}

		switch kind {
		case reflect.String:
			if pk.IsPtr {
				pkValue.Set(reflect.ValueOf(&value))
			} else {
				pkValue.SetString(cast.ToString(value))
			}
		case reflect.Int, reflect.Int64:
			intValue := cast.ToInt64(value)
			if pk.IsPtr {
				pkValue.Set(reflect.ValueOf(&intValue))
			} else {
				pkValue.SetInt(intValue)
			}
		}
	}

	if err := db.NewQuery().Model(&model).WherePK().Scan(ctx, &model); err != nil {
		return err
	}

	if d.preDelete != nil {
		if err := d.preDelete(&model, ctx); err != nil {
			return err
		}
	}

	return db.RunInTx(ctx, func(txCtx context.Context, tx orm.Db) error {
		if _, err := tx.NewDelete().Model(&model).WherePK().Exec(txCtx); err != nil {
			return err
		}

		if d.postDelete != nil {
			if err := d.postDelete(&model, ctx, tx); err != nil {
				return err
			}
		}

		return result.Ok().Response(ctx)
	})
}

// NewDeleteAPI creates a new DeleteAPI instance.
// Use method chaining to configure pre/post processing hooks.
//
// Example:
//   api := NewDeleteAPI[User]().
//     WithPreDelete(checkPermissions).
//     WithPostDelete(auditDelete)
func NewDeleteAPI[TModel any]() *DeleteAPI[TModel] {
	return new(DeleteAPI[TModel])
}
