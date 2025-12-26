package schema

import (
	"context"
	"database/sql"
	"fmt"

	"ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/postgres"
	as "ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/sqlite"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/samber/lo"
)

type AtlasInspector struct {
	inspector as.Inspector
	schema    string
}

// NewInspector creates a new Atlas Inspector for the given database connection.
func NewInspector(db *sql.DB, dbType constants.DbType, schemaName string) (Inspector, error) {
	var (
		inspector as.Inspector
		schema    string
		err       error
	)

	switch dbType {
	case constants.DbPostgres:
		inspector, err = postgres.Open(db)
		if err != nil {
			return nil, fmt.Errorf("failed to open postgres inspector: %w", err)
		}
		schema = lo.CoalesceOrEmpty(schemaName, "public")

	case constants.DbMySQL:
		inspector, err = mysql.Open(db)
		if err != nil {
			return nil, fmt.Errorf("failed to open mysql inspector: %w", err)
		}
		// For MySQL, schema is the database name, which is already set in the connection
		schema = constants.Empty

	case constants.DbSQLite:
		inspector, err = sqlite.Open(db)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite inspector: %w", err)
		}
		schema = "main"

	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	return &AtlasInspector{
		inspector: inspector,
		schema:    schema,
	}, nil
}

func (i *AtlasInspector) InspectSchema(ctx context.Context) (*as.Schema, error) {
	return i.inspector.InspectSchema(ctx, i.schema, &as.InspectOptions{
		Mode: as.InspectTables,
	})
}

func (i *AtlasInspector) InspectTable(ctx context.Context, name string) (*as.Table, error) {
	schema, err := i.inspector.InspectSchema(ctx, i.schema, &as.InspectOptions{
		Tables: []string{name},
	})
	if err != nil {
		return nil, err
	}

	if len(schema.Tables) == 0 {
		return nil, ErrTableNotFound
	}

	return schema.Tables[0], nil
}

func (i *AtlasInspector) InspectViews(ctx context.Context) ([]*as.View, error) {
	schema, err := i.inspector.InspectSchema(ctx, i.schema, &as.InspectOptions{
		Mode: as.InspectViews,
	})
	if err != nil {
		return nil, err
	}

	return schema.Views, nil
}

func (i *AtlasInspector) InspectTriggers(ctx context.Context) ([]*as.Trigger, error) {
	// Triggers are attached to tables and views, so we need to inspect both
	schema, err := i.inspector.InspectSchema(ctx, i.schema, &as.InspectOptions{
		Mode: as.InspectTables | as.InspectViews | as.InspectTriggers,
	})
	if err != nil {
		return nil, err
	}

	var triggers []*as.Trigger

	// Collect triggers from tables
	for _, t := range schema.Tables {
		triggers = append(triggers, t.Triggers...)
	}

	// Collect triggers from views
	for _, v := range schema.Views {
		triggers = append(triggers, v.Triggers...)
	}

	return triggers, nil
}
