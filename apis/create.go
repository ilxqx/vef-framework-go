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

// CreateAPI provides create functionality with pre/post processing hooks.
// It supports method chaining for configuration and handles transaction management automatically.
//
// Type parameters:
//   - TModel: The database model type
//   - TParams: The input parameters type
type CreateAPI[TModel, TParams any] struct {
	preCreate  PreCreateProcessor[TModel, TParams]  // Function to execute before creating the model
	postCreate PostCreateProcessor[TModel, TParams] // Function to execute after creating the model
}

// WithPreCreate sets the pre-create processor for the CreateAPI.
// This processor is called before the model is saved to the database.
// Returns the API instance for method chaining.
func (c *CreateAPI[TModel, TParams]) WithPreCreate(processor PreCreateProcessor[TModel, TParams]) *CreateAPI[TModel, TParams] {
	c.preCreate = processor
	return c
}

// WithPostCreate sets the post-create processor for the CreateAPI.
// This processor is called after the model is successfully saved within the same transaction.
// Returns the API instance for method chaining.
func (c *CreateAPI[TModel, TParams]) WithPostCreate(processor PostCreateProcessor[TModel, TParams]) *CreateAPI[TModel, TParams] {
	c.postCreate = processor
	return c
}

// Create creates a new model with the provided parameters.
// It executes the entire operation within a database transaction for data consistency.
//
// Parameters:
//   - ctx: The Fiber context
//   - db: The database connection
//   - logger: The logger instance
//   - params: The input parameters for creating the model
//
// Returns an error if the operation fails, otherwise returns the primary key(s) of the created record via HTTP response.
func (c *CreateAPI[TModel, TParams]) Create(ctx fiber.Ctx, db orm.Db, logger log.Logger, params TParams) error {
	var model TModel
	if err := copier.Copy(&params, &model); err != nil {
		return err
	}

	if c.preCreate != nil {
		if err := c.preCreate(&model, &params, ctx); err != nil {
			return err
		}
	}

	return db.RunInTx(ctx, func(txCtx context.Context, tx orm.Db) error {
		if _, err := tx.NewCreate().Model(&model).Exec(txCtx); err != nil {
			if utils.IsDuplicateKeyError(err) {
				logger.Warnf("Record already exists: %v", err)
				return result.ErrRecordAlreadyExists
			}
			return err
		}

		if c.postCreate != nil {
			if err := c.postCreate(&model, &params, ctx, tx); err != nil {
				return err
			}
		}

		return result.Ok(db.ModelPKs(&model)).Response(ctx)
	})
}

// NewCreateAPI creates a new CreateAPI instance.
// Use method chaining to configure pre/post processing hooks.
//
// Example:
//   api := NewCreateAPI[User, CreateUserParams]().
//     WithPreCreate(validateUser).
//     WithPostCreate(sendWelcomeEmail)
func NewCreateAPI[TModel, TParams any]() *CreateAPI[TModel, TParams] {
	return new(CreateAPI[TModel, TParams])
}
