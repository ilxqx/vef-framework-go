package apis

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/copier"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type createApi[TModel, TParams any] struct {
	ApiBuilder[CreateApi[TModel, TParams]]

	preCreate  PreCreateProcessor[TModel, TParams]
	postCreate PostCreateProcessor[TModel, TParams]
}

// Provide generates the final Api specification for model creation.
// Returns a complete api.Spec that can be registered with the router.
func (c *createApi[TModel, TParams]) Provide() api.Spec {
	return c.Build(c.create)
}

func (c *createApi[TModel, TParams]) WithPreCreate(processor PreCreateProcessor[TModel, TParams]) CreateApi[TModel, TParams] {
	c.preCreate = processor

	return c
}

func (c *createApi[TModel, TParams]) WithPostCreate(processor PostCreateProcessor[TModel, TParams]) CreateApi[TModel, TParams] {
	c.postCreate = processor

	return c
}

func (c *createApi[TModel, TParams]) create(ctx fiber.Ctx, db orm.Db, params TParams) error {
	var model TModel
	if err := copier.Copy(&params, &model); err != nil {
		return err
	}

	if c.preCreate != nil {
		if err := c.preCreate(&model, &params, ctx, db); err != nil {
			return err
		}
	}

	return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
		if _, err := tx.NewInsert().Model(&model).Exec(txCtx); err != nil {
			return err
		}

		if c.postCreate != nil {
			if err := c.postCreate(&model, &params, ctx, tx); err != nil {
				return err
			}
		}

		pks, err := db.ModelPks(&model)
		if err != nil {
			return err
		}

		return result.Ok(pks).Response(ctx)
	})
}
