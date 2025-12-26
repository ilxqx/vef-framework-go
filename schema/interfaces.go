package schema

import "context"

// Service defines the interface for read-only database schema inspection.
type Service interface {
	// ListTables returns all tables in the current database/schema.
	ListTables(ctx context.Context) ([]Table, error)
	// GetTableSchema returns detailed structure information about a specific table.
	GetTableSchema(ctx context.Context, name string) (*TableSchema, error)
	// ListViews returns all views in the current database/schema.
	ListViews(ctx context.Context) ([]View, error)
	// ListTriggers returns all triggers in the current database/schema.
	ListTriggers(ctx context.Context) ([]Trigger, error)
}
