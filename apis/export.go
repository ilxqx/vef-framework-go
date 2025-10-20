package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	"github.com/ilxqx/vef-framework-go/api"
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

type exportApi[TModel, TSearch any] struct {
	FindApi[TModel, TSearch, []TModel, ExportApi[TModel, TSearch]]

	format          TabularFormat
	excelOpts       []excel.ExportOption
	csvOpts         []csv.ExportOption
	preExport       PreExportProcessor[TModel, TSearch]
	filenameBuilder FilenameBuilder[TSearch]
}

// Provide generates the final Api specification for export.
// Returns a complete api.Spec that can be registered with the router.
func (a *exportApi[TModel, TSearch]) Provide() api.Spec {
	return a.Build(a.exportData)
}

func (a *exportApi[TModel, TSearch]) Format(format TabularFormat) ExportApi[TModel, TSearch] {
	a.format = format

	return a
}

func (a *exportApi[TModel, TSearch]) ExcelOptions(opts ...excel.ExportOption) ExportApi[TModel, TSearch] {
	a.excelOpts = opts

	return a
}

func (a *exportApi[TModel, TSearch]) CSVOptions(opts ...csv.ExportOption) ExportApi[TModel, TSearch] {
	a.csvOpts = opts

	return a
}

func (a *exportApi[TModel, TSearch]) PreExport(processor PreExportProcessor[TModel, TSearch]) ExportApi[TModel, TSearch] {
	a.preExport = processor

	return a
}

func (a *exportApi[TModel, TSearch]) FilenameBuilder(builder FilenameBuilder[TSearch]) ExportApi[TModel, TSearch] {
	a.filenameBuilder = builder

	return a
}

type exportParams struct {
	api.In

	Format TabularFormat `json:"format"` // Optional: override default format
}

func (a *exportApi[TModel, TSearch]) exportData(db orm.Db) (func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch, params exportParams) error, error) {
	if err := a.Init(db); err != nil {
		return nil, err
	}

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

		query := a.BuildQuery(db, &models, search, ctx).SelectModelColumns()
		a.ApplyDefaultSort(query)

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
	}, nil
}
