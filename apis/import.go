package apis

import (
	"context"
	"mime/multipart"

	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/csv"
	"github.com/ilxqx/vef-framework-go/excel"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/tabular"
	"github.com/ilxqx/vef-framework-go/webhelpers"
)

type importApi[TModel, TSearch any] struct {
	ApiBuilder[ImportApi[TModel, TSearch]]

	format     TabularFormat
	excelOpts  []excel.ImportOption
	csvOpts    []csv.ImportOption
	preImport  PreImportProcessor[TModel, TSearch]
	postImport PostImportProcessor[TModel, TSearch]
}

// Provide generates the final Api specification for import.
// Returns a complete api.Spec that can be registered with the router.
func (i *importApi[TModel, TSearch]) Provide() api.Spec {
	return i.Build(i.importData)
}

func (i *importApi[TModel, TSearch]) Format(format TabularFormat) ImportApi[TModel, TSearch] {
	i.format = format

	return i
}

func (i *importApi[TModel, TSearch]) ExcelOptions(opts ...excel.ImportOption) ImportApi[TModel, TSearch] {
	i.excelOpts = opts

	return i
}

func (i *importApi[TModel, TSearch]) CSVOptions(opts ...csv.ImportOption) ImportApi[TModel, TSearch] {
	i.csvOpts = opts

	return i
}

func (i *importApi[TModel, TSearch]) PreImport(processor PreImportProcessor[TModel, TSearch]) ImportApi[TModel, TSearch] {
	i.preImport = processor

	return i
}

func (i *importApi[TModel, TSearch]) PostImport(processor PostImportProcessor[TModel, TSearch]) ImportApi[TModel, TSearch] {
	i.postImport = processor

	return i
}

type importParams struct {
	api.In

	Format TabularFormat         `json:"format"` // Optional: override default format
	File   *multipart.FileHeader `json:"file"`
}

func (i *importApi[TModel, TSearch]) importData() func(ctx fiber.Ctx, db orm.Db, logger log.Logger, search TSearch, params importParams) error {
	// Pre-create importers for both formats
	excelImporter := excel.NewImporterFor[TModel](i.excelOpts...)
	csvImporter := csv.NewImporterFor[TModel](i.csvOpts...)

	return func(ctx fiber.Ctx, db orm.Db, logger log.Logger, search TSearch, params importParams) error {
		// Import requests must use multipart/form-data format
		if webhelpers.IsJson(ctx) {
			return result.Err(i18n.T("import_requires_multipart"))
		}

		if params.File == nil {
			return result.Err(i18n.T("import_requires_file"))
		}

		// Determine format: use param format if provided, otherwise use default
		format := lo.CoalesceOrEmpty(params.Format, i.format, FormatExcel)

		// Select pre-created importer based on format
		var importer tabular.Importer

		switch format {
		case FormatExcel:
			importer = excelImporter
		case FormatCSV:
			importer = csvImporter
		default:
			return result.Err(i18n.T("unsupported_import_format"))
		}

		// Open uploaded file
		file, err := params.File.Open()
		if err != nil {
			return result.Err(i18n.T("file_open_failed"))
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				logger.Errorf("failed to close file: %v", closeErr)
			}
		}()

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
