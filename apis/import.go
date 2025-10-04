package apis

import (
	"context"
	"mime/multipart"

	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/excel"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

type importAPI[TModel, TSearch any] struct {
	APIBuilder[ImportAPI[TModel, TSearch]]

	importerOpts []excel.ImportOption
	preImport    PreImportProcessor[TModel, TSearch]
	postImport   PostImportProcessor[TModel, TSearch]
}

// Provide generates the final API specification for import.
// Returns a complete api.Spec that can be registered with the router.
func (i *importAPI[TModel, TSearch]) Provide() api.Spec {
	return i.APIBuilder.Build(i.importData)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (i *importAPI[TModel, TSearch]) Build(handler any) api.Spec {
	panic("apis: do not call APIBuilder.Build on importAPI; call Provide() instead")
}

func (i *importAPI[TModel, TSearch]) ImportOptions(opts ...excel.ImportOption) ImportAPI[TModel, TSearch] {
	i.importerOpts = opts
	return i
}

func (i *importAPI[TModel, TSearch]) PreImport(processor PreImportProcessor[TModel, TSearch]) ImportAPI[TModel, TSearch] {
	i.preImport = processor
	return i
}

func (i *importAPI[TModel, TSearch]) PostImport(processor PostImportProcessor[TModel, TSearch]) ImportAPI[TModel, TSearch] {
	i.postImport = processor
	return i
}

type importParams struct {
	api.In

	File *multipart.FileHeader
}

func (i *importAPI[TModel, TSearch]) importData() func(ctx fiber.Ctx, db orm.Db, search TSearch, params importParams) error {
	// Create importer with configured options
	importer := excel.NewImporterFor[TModel](i.importerOpts...)

	return func(ctx fiber.Ctx, db orm.Db, search TSearch, params importParams) error {
		// Import requests must use multipart/form-data format
		if webhelpers.IsJSON(ctx) {
			return result.Err(i18n.T("import_requires_multipart"))
		}

		if params.File == nil {
			return result.Err(i18n.T("import_requires_file"))
		}

		// Open uploaded file
		file, err := params.File.Open()
		if err != nil {
			return result.Err(i18n.T("file_open_failed"))
		}
		defer file.Close()

		// Import data from file
		modelsAny, importErrors, err := importer.Import(file)
		if err != nil {
			return err
		}

		// Type assert to slice of models
		models := modelsAny.([]TModel)

		// Return errors if any
		if len(importErrors) > 0 {
			return result.Result{
				Code:    result.ErrCodeDefault,
				Message: i18n.T("import_validation_failed"),
				Data: fiber.Map{
					"errors": importErrors,
				},
			}.Response(ctx)
		}

		// Apply pre-import processor
		if i.preImport != nil {
			if err := i.preImport(models, search, ctx, db); err != nil {
				return err
			}
		}

		// Save models to database in a transaction
		return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
			if len(models) > 0 {
				if _, err := tx.NewInsert().Model(&models).Exec(txCtx); err != nil {
					return err
				}
			}

			// Apply post-import processor
			if i.postImport != nil {
				if err := i.postImport(models, search, ctx, tx); err != nil {
					return err
				}
			}

			return result.Ok(fiber.Map{
				"total": len(models),
			}).Response(ctx)
		})
	}
}
