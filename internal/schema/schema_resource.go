package schema

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/result"
	"github.com/ilxqx/vef-framework-go/schema"
)

// NewResource creates a new schema inspection resource.
func NewResource(service schema.Service) api.Resource {
	return &Resource{
		service: service,
		Resource: api.NewResource(
			"sys/schema",
			api.WithApis(
				api.Spec{Action: "list_tables", Public: isSchemaApiPublic, Limit: api.RateLimit{Max: 60}},
				api.Spec{Action: "get_table_schema", Public: isSchemaApiPublic, Limit: api.RateLimit{Max: 60}},
				api.Spec{Action: "list_views", Public: isSchemaApiPublic, Limit: api.RateLimit{Max: 60}},
				api.Spec{Action: "list_triggers", Public: isSchemaApiPublic, Limit: api.RateLimit{Max: 60}},
			),
		),
	}
}

// Resource handles schema inspection Api endpoints.
type Resource struct {
	api.Resource

	service schema.Service
}

// ListTables returns all tables in the current database/schema.
func (r *Resource) ListTables(ctx fiber.Ctx) error {
	tables, err := r.service.ListTables(ctx.Context())
	if err != nil {
		return err
	}

	return result.Ok(tables).Response(ctx)
}

// GetTableSchemaParams contains parameters for getting table schema details.
type GetTableSchemaParams struct {
	api.P

	Name string `json:"name" validate:"required"`
}

// GetTableSchema returns detailed structure information about a specific table.
func (r *Resource) GetTableSchema(ctx fiber.Ctx, params GetTableSchemaParams) error {
	table, err := r.service.GetTableSchema(ctx.Context(), params.Name)
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			return result.Err(
				i18n.T("schema_table_not_found"),
				result.WithCode(result.ErrCodeSchemaTableNotFound),
			)
		}

		return err
	}

	return result.Ok(table).Response(ctx)
}

// ListViews returns all views in the current database/schema.
func (r *Resource) ListViews(ctx fiber.Ctx) error {
	views, err := r.service.ListViews(ctx.Context())
	if err != nil {
		return err
	}

	return result.Ok(views).Response(ctx)
}

// ListTriggers returns all triggers in the current database/schema.
func (r *Resource) ListTriggers(ctx fiber.Ctx) error {
	triggers, err := r.service.ListTriggers(ctx.Context())
	if err != nil {
		return err
	}

	return result.Ok(triggers).Response(ctx)
}
