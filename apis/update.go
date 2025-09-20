package apis

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/copier"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/utils"
)

// UpdateAPI provides update functionality with pre/post processing hooks.
// It supports method chaining for configuration and handles transaction management automatically.
//
// Type parameters:
//   - TModel: The database model type
//   - TParams: The input parameters type
type UpdateAPI[TModel, TParams any] struct {
	preUpdate  PreUpdateProcessor[TModel, TParams]
	postUpdate PostUpdateProcessor[TModel, TParams]
}

// WithPreUpdate sets the pre-update processor for the UpdateAPI.
// This processor is called before the model is updated in the database.
// Returns the API instance for method chaining.
func (u *UpdateAPI[TModel, TParams]) WithPreUpdate(processor PreUpdateProcessor[TModel, TParams]) *UpdateAPI[TModel, TParams] {
	u.preUpdate = processor
	return u
}

// WithPostUpdate sets the post-update processor for the UpdateAPI.
// This processor is called after the model is successfully updated within the same transaction.
// Returns the API instance for method chaining.
func (u *UpdateAPI[TModel, TParams]) WithPostUpdate(processor PostUpdateProcessor[TModel, TParams]) *UpdateAPI[TModel, TParams] {
	u.postUpdate = processor
	return u
}

// Update updates an existing model with the provided parameters.
// It first fetches the existing model, applies the changes, and executes within a transaction.
//
// Parameters:
//   - ctx: The Fiber context
//   - db: The database connection
//   - logger: The logger instance
//   - params: The input parameters for updating the model
//
// Returns an error if the operation fails, otherwise returns success via HTTP response.
func (u *UpdateAPI[TModel, TParams]) Update(ctx fiber.Ctx, db orm.Db, logger log.Logger, params TParams) error {
	var (
		oldModel TModel
		model    TModel
	)

	if err := copier.Copy(&params, &model); err != nil {
		return err
	}

	if err := db.NewQuery().Model(&model).WherePK().Scan(ctx, &oldModel); err != nil {
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
			if utils.IsDuplicateKeyError(err) {
				logger.Warnf("Record already exists: %v", err)
				return result.ErrRecordAlreadyExists
			}
			return err
		}

		if u.postUpdate != nil {
			if err := u.postUpdate(&oldModel, &model, &params, ctx, tx); err != nil {
				return err
			}
		}

		return result.Ok().Response(ctx)
	})
}

// NewUpdateAPI creates a new UpdateAPI instance.
// Use method chaining to configure pre/post processing hooks.
//
// Example:
//
//	api := NewUpdateAPI[User, UpdateUserParams]().
//	  WithPreUpdate(validateChanges).
//	  WithPostUpdate(auditUpdate)
func NewUpdateAPI[TModel, TParams any]() *UpdateAPI[TModel, TParams] {
	return new(UpdateAPI[TModel, TParams])
}
