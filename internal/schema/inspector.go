package schema

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/sqlite"
	"github.com/samber/lo"

	as "ariga.io/atlas/sql/schema"

	"github.com/ilxqx/vef-framework-go/constants"
)

var ErrUnsupportedDBType = errors.New("unsupported database type")

type AtlasInspector struct {
	inspector as.Inspector
	schema    string
}

// NewInspector creates a new Atlas Inspector for the given database connection.
func NewInspector(db *sql.DB, dbType constants.DBType, schemaName string) (Inspector, error) {
	var inspector as.Inspector
	var schema string
	var err error

	switch dbType {
	case constants.Postgres:
		inspector, err = postgres.Open(db)
		schema = lo.CoalesceOrEmpty(schemaName, "public")

	case constants.MySQL:
		inspector, err = mysql.Open(db)
		// For MySQL, schema is the database name, which is already set in the connection
		schema = constants.Empty

	case constants.SQLite:
		inspector, err = sqlite.Open(db)
		schema = "main"

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDBType, dbType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open %s inspector: %w", dbType, err)
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
