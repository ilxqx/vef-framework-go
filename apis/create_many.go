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

type createManyApi[TModel, TParams any] struct {
	Builder[CreateMany[TModel, TParams]]

	preCreateMany  PreCreateManyProcessor[TModel, TParams]
	postCreateMany PostCreateManyProcessor[TModel, TParams]
}

func (c *createManyApi[TModel, TParams]) Provide() []api.OperationSpec {
	return []api.OperationSpec{c.Build(c.createMany)}
}

func (c *createManyApi[TModel, TParams]) WithPreCreateMany(processor PreCreateManyProcessor[TModel, TParams]) CreateMany[TModel, TParams] {
	c.preCreateMany = processor

	return c
}

func (c *createManyApi[TModel, TParams]) WithPostCreateMany(processor PostCreateManyProcessor[TModel, TParams]) CreateMany[TModel, TParams] {
	c.postCreateMany = processor

	return c
}

func (c *createManyApi[TModel, TParams]) createMany(sc storage.Service, publisher event.Publisher) (func(ctx fiber.Ctx, db orm.DB, params CreateManyParams[TParams]) error, error) {
	promoter := storage.NewPromoter[TModel](sc, publisher)

	return func(ctx fiber.Ctx, db orm.DB, params CreateManyParams[TParams]) error {
		if len(params.List) == 0 {
			return result.Ok([]map[string]any{}).Response(ctx)
		}

		models := make([]TModel, len(params.List))
		for i := range params.List {
			if err := copier.Copy(&params.List[i], &models[i]); err != nil {
				return err
			}
		}

		return db.RunInTX(ctx.Context(), func(txCtx context.Context, tx orm.DB) error {
			query := tx.NewInsert().Model(&models)
			if c.preCreateMany != nil {
				if err := c.preCreateMany(models, params.List, query, ctx, tx); err != nil {
					return err
				}
			}

			for i := range models {
				if err := promoter.Promote(txCtx, &models[i], nil); err != nil {
					if rollbackErr := batchCleanup(txCtx, promoter, models[:i]); rollbackErr != nil {
						return fmt.Errorf("promote files for model %d failed: %w; rollback also failed: %w", i, err, rollbackErr)
					}

					return fmt.Errorf("promote files for model %d failed: %w", i, err)
				}
			}

			if _, err := query.Exec(txCtx); err != nil {
				if cleanupErr := batchCleanup(txCtx, promoter, models); cleanupErr != nil {
					return fmt.Errorf("batch create failed: %w; cleanup files also failed: %w", err, cleanupErr)
				}

				return err
			}

			if c.postCreateMany != nil {
				if err := c.postCreateMany(models, params.List, ctx, tx); err != nil {
					if cleanupErr := batchCleanup(txCtx, promoter, models); cleanupErr != nil {
						return fmt.Errorf("post-create-many failed: %w; cleanup files also failed: %w", err, cleanupErr)
					}

					return err
				}
			}

			pks := make([]map[string]any, len(models))
			for i := range models {
				pk, err := db.ModelPKs(&models[i])
				if err != nil {
					if cleanupErr := batchCleanup(txCtx, promoter, models); cleanupErr != nil {
						return fmt.Errorf("get primary keys failed: %w; cleanup files also failed: %w", err, cleanupErr)
					}

					return err
				}

				pks[i] = pk
			}

			return result.Ok(pks).Response(ctx)
		})
	}, nil
}
