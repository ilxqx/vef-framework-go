package apis

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/copier"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type createAPI[TModel, TParams any] struct {
	APIBuilder[CreateAPI[TModel, TParams]]

	preCreate  PreCreateProcessor[TModel, TParams]
	postCreate PostCreateProcessor[TModel, TParams]
}

// Provide generates the final API specification for model creation.
// Returns a complete api.Spec that can be registered with the router.
func (c *createAPI[TModel, TParams]) Provide() api.Spec {
	return c.APIBuilder.Build(c.create)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (c *createAPI[TModel, TParams]) Build(handler any) api.Spec {
	panic("apis: do not call APIBuilder.Build on createAPI; call Provide() instead")
}

func (c *createAPI[TModel, TParams]) PreCreate(processor PreCreateProcessor[TModel, TParams]) CreateAPI[TModel, TParams] {
	c.preCreate = processor
	return c
}

func (c *createAPI[TModel, TParams]) PostCreate(processor PostCreateProcessor[TModel, TParams]) CreateAPI[TModel, TParams] {
	c.postCreate = processor
	return c
}

func (c *createAPI[TModel, TParams]) create(ctx fiber.Ctx, db orm.Db, params TParams) error {
	var model TModel
	if err := copier.Copy(&params, &model); err != nil {
		return err
	}

	if c.preCreate != nil {
		if err := c.preCreate(&model, &params, ctx, db); err != nil {
			return err
		}
	}

	return db.RunInTx(ctx, func(txCtx context.Context, tx orm.Db) error {
		if _, err := tx.NewInsert().Model(&model).Exec(txCtx); err != nil {
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
