package apis

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/copier"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/storage"
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

func (c *createApi[TModel, TParams]) create(sc storage.Service, publisher event.Publisher) (func(ctx fiber.Ctx, db orm.Db, params TParams) error, error) {
	promoter := storage.NewPromoter[TModel](sc, publisher)

	return func(ctx fiber.Ctx, db orm.Db, params TParams) error {
		var model TModel
		if err := copier.Copy(&params, &model); err != nil {
			return err
		}

		return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
			if c.preCreate != nil {
				if err := c.preCreate(&model, &params, ctx, db); err != nil {
					return err
				}
			}

			if err := promoter.Promote(txCtx, &model, nil); err != nil {
				return fmt.Errorf("promote files failed: %w", err)
			}

			if _, err := tx.NewInsert().Model(&model).Exec(txCtx); err != nil {
				if cleanupErr := promoter.Promote(txCtx, nil, &model); cleanupErr != nil {
					return fmt.Errorf("insert failed: %w; cleanup files also failed: %v", err, cleanupErr)
				}
				return err
			}

			if c.postCreate != nil {
				if err := c.postCreate(&model, &params, ctx, tx); err != nil {
					if cleanupErr := promoter.Promote(txCtx, nil, &model); cleanupErr != nil {
						return fmt.Errorf("post-create failed: %w; cleanup files also failed: %v", err, cleanupErr)
					}
					return err
				}
			}

			pks, err := db.ModelPks(&model)
			if err != nil {
				if cleanupErr := promoter.Promote(txCtx, nil, &model); cleanupErr != nil {
					return fmt.Errorf("get primary keys failed: %w; cleanup files also failed: %v", err, cleanupErr)
				}
				return err
			}

			return result.Ok(pks).Response(ctx)
		})
	}, nil
}
