package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/csv"
	"github.com/ilxqx/vef-framework-go/excel"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/tabular"
)

const (
	contentTypeExcel     = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	contentTypeCSV       = "text/csv; charset=utf-8"
	defaultFilenameExcel = "data.xlsx"
	defaultFilenameCSV   = "data.csv"
)

type exportAPI[TModel, TSearch any] struct {
	FindAPI[TModel, TSearch, []TModel, ExportAPI[TModel, TSearch]]

	format          TabularFormat
	excelOpts       []excel.ExportOption
	csvOpts         []csv.ExportOption
	preExport       PreExportProcessor[TModel, TSearch]
	filenameBuilder FilenameBuilder[TSearch]
}

// Provide generates the final API specification for export.
// Returns a complete api.Spec that can be registered with the router.
func (a *exportAPI[TModel, TSearch]) Provide() api.Spec {
	return a.FindAPI.Build(a.exportData)
}

// Build should not be called directly on concrete API types.
// Use Provide() to generate api.Spec with the correct handler instead.
func (a *exportAPI[TModel, TSearch]) Build(handler any) api.Spec {
	panic("apis: do not call FindAPI.Build on exportAPI; call Provide() instead")
}

func (a *exportAPI[TModel, TSearch]) Format(format TabularFormat) ExportAPI[TModel, TSearch] {
	a.format = format

	return a
}

func (a *exportAPI[TModel, TSearch]) ExcelOptions(opts ...excel.ExportOption) ExportAPI[TModel, TSearch] {
	a.excelOpts = opts

	return a
}

func (a *exportAPI[TModel, TSearch]) CSVOptions(opts ...csv.ExportOption) ExportAPI[TModel, TSearch] {
	a.csvOpts = opts

	return a
}

func (a *exportAPI[TModel, TSearch]) PreExport(processor PreExportProcessor[TModel, TSearch]) ExportAPI[TModel, TSearch] {
	a.preExport = processor

	return a
}

func (a *exportAPI[TModel, TSearch]) FilenameBuilder(builder FilenameBuilder[TSearch]) ExportAPI[TModel, TSearch] {
	a.filenameBuilder = builder

	return a
}

type exportParams struct {
	api.In

	Format TabularFormat `json:"format"` // Optional: override default format
}

func (a *exportAPI[TModel, TSearch]) exportData(db orm.Db) func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch, params exportParams) error {
	// Pre-compute schema information
	schema := db.TableOf((*TModel)(nil))

	// Pre-compute whether default ordering should be applied
	hasCreatedAt := schema.HasField(constants.ColumnCreatedAt)
	shouldApplyDefaultSort := !a.HasSortApplier() && hasCreatedAt

	// Pre-create exporters for both formats
	excelExporter := excel.NewExporterFor[TModel](a.excelOpts...)
	csvExporter := csv.NewExporterFor[TModel](a.csvOpts...)

	return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch, params exportParams) error {
		// Determine format: use param format if provided, otherwise use default
		format := lo.CoalesceOrEmpty(params.Format, a.format, FormatExcel)

		// Select pre-created exporter based on format
		var (
			exporter                     tabular.Exporter
			contentType, defaultFilename string
		)

		switch format {
		case FormatExcel:
			exporter = excelExporter
			contentType = contentTypeExcel
			defaultFilename = defaultFilenameExcel
		case FormatCSV:
			exporter = csvExporter
			contentType = contentTypeCSV
			defaultFilename = defaultFilenameCSV
		default:
			return result.Err(i18n.T("unsupported_export_format"))
		}

		var models []TModel

		query := a.BuildQuery(db, &models, search, ctx)

		if shouldApplyDefaultSort {
			// Add default ordering by created_at
			query.OrderByDesc(constants.ColumnCreatedAt)
		}

		// Execute query with safety limit
		if err := query.Limit(maxQueryLimit).Scan(ctx.Context()); err != nil {
			return err
		}

		// Apply transformation to each model
		if len(models) > 0 {
			for i := range models {
				if err := transformer.Struct(ctx.Context(), &models[i]); err != nil {
					return err
				}
			}
		}

		// Apply pre-export processor
		if a.preExport != nil {
			if err := a.preExport(models, search, ctx, db); err != nil {
				return err
			}
		}

		// Export to buffer
		buf, err := exporter.Export(models)
		if err != nil {
			return err
		}

		// Build filename
		filename := defaultFilename
		if a.filenameBuilder != nil {
			filename = a.filenameBuilder(search, ctx)
		}

		// Set response headers for file download
		ctx.Set(fiber.HeaderContentType, contentType)
		ctx.Set(fiber.HeaderContentDisposition, "attachment; filename="+filename)

		return ctx.Send(buf.Bytes())
	}
}
