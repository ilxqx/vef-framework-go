package schema

import (
	"context"
	"database/sql"
	"fmt"

	as "ariga.io/atlas/sql/schema"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/schema"
)

// DefaultService is the default implementation of schema.Service.
type DefaultService struct {
	inspector Inspector
}

// NewService creates a new schema service.
func NewService(db *sql.DB, dsConfig *config.DatasourceConfig) (schema.Service, error) {
	inspector, err := NewInspector(db, dsConfig.Type, dsConfig.Schema)
	if err != nil {
		return nil, err
	}

	return &DefaultService{
		inspector: inspector,
	}, nil
}

// ListTables returns all tables in the current database/schema.
func (s *DefaultService) ListTables(ctx context.Context) ([]schema.Table, error) {
	result, err := s.inspector.InspectSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect schema: %w", err)
	}

	tables := make([]schema.Table, len(result.Tables))
	for i, t := range result.Tables {
		table := schema.Table{
			Name: t.Name,
		}
		if t.Schema != nil {
			table.Schema = t.Schema.Name
		}

		for _, attr := range t.Attrs {
			if comment, ok := attr.(*as.Comment); ok {
				table.Comment = comment.Text

				break
			}
		}

		tables[i] = table
	}

	return tables, nil
}

// GetTableSchema returns detailed structure information about a specific table.
func (s *DefaultService) GetTableSchema(ctx context.Context, name string) (*schema.TableSchema, error) {
	table, err := s.inspector.InspectTable(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect table: %w", err)
	}

	return s.convertTable(table), nil
}

// convertTable converts an Atlas table to a schema.TableSchema.
func (s *DefaultService) convertTable(t *as.Table) *schema.TableSchema {
	info := schema.TableSchema{
		Name:    t.Name,
		Columns: make([]schema.Column, len(t.Columns)),
	}

	if t.Schema != nil {
		info.Schema = t.Schema.Name
	}

	// Build primary key column set for quick lookup
	pkColumns := make(map[string]bool)
	if t.PrimaryKey != nil {
		pkCols := make([]string, len(t.PrimaryKey.Parts))
		for i, part := range t.PrimaryKey.Parts {
			if part.C != nil {
				pkColumns[part.C.Name] = true
				pkCols[i] = part.C.Name
			}
		}

		if len(pkCols) > 0 {
			info.PrimaryKey = &schema.PrimaryKey{
				Name:    t.PrimaryKey.Name,
				Columns: pkCols,
			}
		}
	}

	// Convert columns
	for i, col := range t.Columns {
		colInfo := schema.Column{
			Name:         col.Name,
			Type:         col.Type.Raw,
			Nullable:     col.Type.Null,
			IsPrimaryKey: pkColumns[col.Name],
		}

		if col.Default != nil {
			if raw, ok := col.Default.(*as.RawExpr); ok {
				colInfo.Default = raw.X
			}
		}

		for _, attr := range col.Attrs {
			if comment, ok := attr.(*as.Comment); ok {
				colInfo.Comment = comment.Text
			}
		}
		// AutoIncrement detection is database-specific, check via type string
		if hasAutoIncrement(col) {
			colInfo.IsAutoIncrement = true
		}

		info.Columns[i] = colInfo
	}

	// Convert indexes and unique keys
	for _, idx := range t.Indexes {
		columns := make([]string, len(idx.Parts))
		for i, part := range idx.Parts {
			if part.C != nil {
				columns[i] = part.C.Name
			}
		}

		if idx.Unique {
			info.UniqueKeys = append(info.UniqueKeys, schema.UniqueKey{
				Name:    idx.Name,
				Columns: columns,
			})
		} else {
			info.Indexes = append(info.Indexes, schema.Index{
				Name:    idx.Name,
				Columns: columns,
			})
		}
	}

	// Convert foreign keys
	for _, fk := range t.ForeignKeys {
		fkInfo := schema.ForeignKey{
			Name:       fk.Symbol,
			Columns:    make([]string, len(fk.Columns)),
			RefColumns: make([]string, len(fk.RefColumns)),
		}

		if fk.RefTable != nil {
			fkInfo.RefTable = fk.RefTable.Name
		}

		for i, col := range fk.Columns {
			fkInfo.Columns[i] = col.Name
		}

		for i, col := range fk.RefColumns {
			fkInfo.RefColumns[i] = col.Name
		}

		fkInfo.OnUpdate = referentialActionToString(fk.OnUpdate)
		fkInfo.OnDelete = referentialActionToString(fk.OnDelete)

		info.ForeignKeys = append(info.ForeignKeys, fkInfo)
	}

	// Convert table attributes (comment, check constraints)
	for _, attr := range t.Attrs {
		switch a := attr.(type) {
		case *as.Comment:
			info.Comment = a.Text
		case *as.Check:
			info.Checks = append(info.Checks, schema.Check{
				Name: a.Name,
				Expr: a.Expr,
			})
		}
	}

	return &info
}

// referentialActionToString converts a referential action to string.
func referentialActionToString(action as.ReferenceOption) string {
	switch action {
	case as.Cascade:
		return "CASCADE"
	case as.SetNull:
		return "SET NULL"
	case as.SetDefault:
		return "SET DEFAULT"
	case as.Restrict:
		return "RESTRICT"
	case as.NoAction:
		return "NO ACTION"
	default:
		return constants.Empty
	}
}

// hasAutoIncrement checks if a column has auto-increment attribute.
func hasAutoIncrement(col *as.Column) bool {
	for _, attr := range col.Attrs {
		// Check attribute type name for auto increment indicators
		typeName := fmt.Sprintf("%T", attr)
		if typeName == "*mysql.AutoIncrement" || typeName == "*sqlite.AutoIncrement" {
			return true
		}
	}
	// PostgreSQL uses SERIAL types which show up in the type raw string
	if col.Type != nil && col.Type.Raw != constants.Empty {
		raw := col.Type.Raw
		if raw == "serial" || raw == "bigserial" || raw == "smallserial" ||
			raw == "SERIAL" || raw == "BIGSERIAL" || raw == "SMALLSERIAL" {

			return true
		}
	}

	return false
}

// ListViews returns all views in the current database/schema.
func (s *DefaultService) ListViews(ctx context.Context) ([]schema.View, error) {
	views, err := s.inspector.InspectViews(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect views: %w", err)
	}

	result := make([]schema.View, len(views))
	for i, v := range views {
		view := schema.View{
			Name:         v.Name,
			Definition:   v.Def,
			Materialized: v.Materialized(),
		}
		if v.Schema != nil {
			view.Schema = v.Schema.Name
		}
		// Extract column names
		view.Columns = make([]string, len(v.Columns))
		for j, col := range v.Columns {
			view.Columns[j] = col.Name
		}
		// Extract comment
		for _, attr := range v.Attrs {
			if comment, ok := attr.(*as.Comment); ok {
				view.Comment = comment.Text

				break
			}
		}

		result[i] = view
	}

	return result, nil
}

// ListTriggers returns all triggers in the current database/schema.
func (s *DefaultService) ListTriggers(ctx context.Context) ([]schema.Trigger, error) {
	triggers, err := s.inspector.InspectTriggers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect triggers: %w", err)
	}

	result := make([]schema.Trigger, len(triggers))
	for i, t := range triggers {
		trigger := schema.Trigger{
			Name:       t.Name,
			ActionTime: string(t.ActionTime),
			ForEachRow: t.For == as.TriggerForRow,
			Body:       t.Body,
		}
		if t.Table != nil {
			trigger.Table = t.Table.Name
		}

		if t.View != nil {
			trigger.View = t.View.Name
		}
		// Extract event names
		trigger.Events = make([]string, len(t.Events))
		for j, event := range t.Events {
			trigger.Events[j] = event.Name
		}

		result[i] = trigger
	}

	return result, nil
}
