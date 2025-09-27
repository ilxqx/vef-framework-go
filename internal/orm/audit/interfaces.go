package audit

import (
	"reflect"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// Handler is the base interface for all auto column handlers.
// It provides the column name that the handler manages.
type Handler interface {
	// Name returns the name of the column this handler manages.
	Name() string
}

// InsertHandler is an interface for handlers that automatically manage columns during insert operations.
// Handlers implementing this interface will be called before insert operations to set column values.
type InsertHandler interface {
	Handler
	// OnInsert is called when a new record is being inserted.
	// It allows the handler to automatically set or modify column values.
	OnInsert(query *bun.InsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value)
}

// UpdateHandler is an interface for handlers that manage columns during both insert and update operations.
// It extends InsertHandler to also handle update scenarios with additional context.
type UpdateHandler interface {
	InsertHandler
	// OnUpdate is called when an existing record is being updated.
	// The hasSet parameter indicates whether any SET clauses have been explicitly added to the update query.
	OnUpdate(query *bun.UpdateQuery, hasSet bool, table *schema.Table, field *schema.Field, model any, value reflect.Value)
}
