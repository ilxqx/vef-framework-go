package apis

import (
	"github.com/gofiber/fiber/v3"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/excel"
	"github.com/ilxqx/vef-framework-go/mold"
	"github.com/ilxqx/vef-framework-go/orm"
)

type exportAPI[TModel, TSearch any] struct {
	FindAPI[TModel, TSearch, []TModel, ExportAPI[TModel, TSearch]]

	exporterOpts    []excel.ExportOption
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

func (a *exportAPI[TModel, TSearch]) ExportOptions(opts ...excel.ExportOption) ExportAPI[TModel, TSearch] {
	a.exporterOpts = opts
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

func (a *exportAPI[TModel, TSearch]) exportData(db orm.Db) func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error {
	// Pre-compute schema information
	schema := db.TableOf((*TModel)(nil))

	// Pre-compute whether default ordering should be applied
	hasCreatedAt := schema.HasField(constants.ColumnCreatedAt)
	shouldApplyDefaultSort := !a.HasSortApplier() && hasCreatedAt

	// Create exporter with configured options
	exporter := excel.NewExporterFor[TModel](a.exporterOpts...)

	return func(ctx fiber.Ctx, db orm.Db, transformer mold.Transformer, search TSearch) error {
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
		filename := "data.xlsx"
		if a.filenameBuilder != nil {
			filename = a.filenameBuilder(search, ctx)
		}

		// Set response headers for file download
		ctx.Set(fiber.HeaderContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		ctx.Set(fiber.HeaderContentDisposition, "attachment; filename="+filename)

		return ctx.Send(buf.Bytes())
	}
}
