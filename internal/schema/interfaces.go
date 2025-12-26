package schema

import (
	"context"

	as "ariga.io/atlas/sql/schema"
)

// Inspector wraps Atlas inspection capabilities for read-only schema inspection.
type Inspector interface {
	// InspectSchema inspects the current database schema.
	InspectSchema(ctx context.Context) (*as.Schema, error)
	// InspectTable inspects a specific table.
	InspectTable(ctx context.Context, name string) (*as.Table, error)
	// InspectViews inspects all views in the current database schema.
	InspectViews(ctx context.Context) ([]*as.View, error)
	// InspectTriggers inspects all triggers in the current database schema.
	InspectTriggers(ctx context.Context) ([]*as.Trigger, error)
}
