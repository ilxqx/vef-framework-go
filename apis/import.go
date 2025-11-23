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

type importApi[TModel any] struct {
	ApiBuilder[ImportApi[TModel]]

	defaultFormat TabularFormat
	excelOpts     []excel.ImportOption
	csvOpts       []csv.ImportOption
	preImport     PreImportProcessor[TModel]
	postImport    PostImportProcessor[TModel]
}

// Provide generates the final Api specification for import.
// Returns a complete api.Spec that can be registered with the router.
func (i *importApi[TModel]) Provide() api.Spec {
	return i.Build(i.importData)
}

func (i *importApi[TModel]) WithDefaultFormat(format TabularFormat) ImportApi[TModel] {
	i.defaultFormat = format

	return i
}

func (i *importApi[TModel]) WithExcelOptions(opts ...excel.ImportOption) ImportApi[TModel] {
	i.excelOpts = opts

	return i
}

func (i *importApi[TModel]) WithCsvOptions(opts ...csv.ImportOption) ImportApi[TModel] {
	i.csvOpts = opts

	return i
}

func (i *importApi[TModel]) WithPreImport(processor PreImportProcessor[TModel]) ImportApi[TModel] {
	i.preImport = processor

	return i
}

func (i *importApi[TModel]) WithPostImport(processor PostImportProcessor[TModel]) ImportApi[TModel] {
	i.postImport = processor

	return i
}

type importParams struct {
	api.P

	File *multipart.FileHeader `json:"file"`
}

type importConfig struct {
	api.M

	Format TabularFormat `json:"format"`
}

func (i *importApi[TModel]) importData() func(ctx fiber.Ctx, db orm.Db, logger log.Logger, config importConfig, params importParams) error {
	excelImporter := excel.NewImporterFor[TModel](i.excelOpts...)
	csvImporter := csv.NewImporterFor[TModel](i.csvOpts...)

	return func(ctx fiber.Ctx, db orm.Db, logger log.Logger, config importConfig, params importParams) error {
		// Import requests must use multipart/form-data format
		if webhelpers.IsJson(ctx) {
			return result.Err(i18n.T("import_requires_multipart"))
		}

		if params.File == nil {
			return result.Err(i18n.T("import_requires_file"))
		}

		var (
			format   = lo.CoalesceOrEmpty(config.Format, i.defaultFormat, FormatExcel)
			importer tabular.Importer
		)

		switch format {
		case FormatExcel:
			importer = excelImporter
		case FormatCsv:
			importer = csvImporter
		default:
			return result.Err(i18n.T("unsupported_import_format"))
		}

		file, err := params.File.Open()
		if err != nil {
			return result.Err(i18n.T("file_open_failed"))
		}

		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				logger.Errorf("failed to close file: %v", closeErr)
			}
		}()

		modelsAny, importErrors, err := importer.Import(file)
		if err != nil {
			return err
		}

		models := modelsAny.([]TModel)

		if len(importErrors) > 0 {
			return result.Result{
				Code:    result.ErrCodeDefault,
				Message: i18n.T("import_validation_failed"),
				Data: fiber.Map{
					"errors": importErrors,
				},
			}.Response(ctx)
		}

		return db.RunInTx(ctx.Context(), func(txCtx context.Context, tx orm.Db) error {
			query := tx.NewInsert().Model(&models)
			if i.preImport != nil {
				if err := i.preImport(models, query, ctx, tx); err != nil {
					return err
				}
			}

			if len(models) > 0 {
				if _, err := query.Exec(txCtx); err != nil {
					return err
				}
			}

			if i.postImport != nil {
				if err := i.postImport(models, ctx, tx); err != nil {
					return err
				}
			}

			return result.Ok(
				fiber.Map{
					"total": len(models),
				}).
				Response(ctx)
		})
	}
}
