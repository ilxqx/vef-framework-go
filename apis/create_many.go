package apis

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/copier"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
)

type createManyApi[TModel, TParams any] struct {
	ApiBuilder[CreateManyApi[TModel, TParams]]

	preCreateMany  PreCreateManyProcessor[TModel, TParams]
	postCreateMany PostCreateManyProcessor[TModel, TParams]
}

// Provide generates the final Api specification for batch model creation.
// Returns a complete api.Spec that can be registered with the router.
func (c *createManyApi[TModel, TParams]) Provide() api.Spec {
	return c.Build(c.createMany)
}

func (c *createManyApi[TModel, TParams]) PreCreateMany(processor PreCreateManyProcessor[TModel, TParams]) CreateManyApi[TModel, TParams] {
	c.preCreateMany = processor

	return c
}

func (c *createManyApi[TModel, TParams]) PostCreateMany(processor PostCreateManyProcessor[TModel, TParams]) CreateManyApi[TModel, TParams] {
	c.postCreateMany = processor

	return c
}

func (c *createManyApi[TModel, TParams]) createMany(ctx fiber.Ctx, db orm.Db, params CreateManyParams[TParams]) error {
	if len(params.List) == 0 {
		return result.Ok([]map[string]any{}).Response(ctx)
	}

	models := make([]TModel, len(params.List))
	for i := range params.List {
		if err := copier.Copy(&params.List[i], &models[i]); err != nil {
			return err
		}
	}

	if c.preCreateMany != nil {
		if err := c.preCreateMany(models, params.List, ctx, db); err != nil {
			return err
		}
	}

	return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
		if _, err := tx.NewInsert().Model(&models).Exec(txCtx); err != nil {
			return err
		}

		if c.postCreateMany != nil {
			if err := c.postCreateMany(models, params.List, ctx, tx); err != nil {
				return err
			}
		}

		// Return primary keys for all created models
		pks := make([]map[string]any, len(models))
		for i := range models {
			pk, err := db.ModelPks(&models[i])
			if err != nil {
				return err
			}

			pks[i] = pk
		}

		return result.Ok(pks).Response(ctx)
	})
}
