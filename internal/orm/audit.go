package orm

import (
	"reflect"

	"github.com/uptrace/bun/schema"
)

// autoColumnHandlers is the list of auto column handlers that are applied to all models.
// These handlers automatically manage audit fields like ID generation, timestamps, and user tracking.
var (
	autoColumnHandlers = []ColumnHandler{
		&IdHandler{},
		&CreatedAtHandler{},
		&UpdatedAtHandler{},
		&CreatedByHandler{},
		&UpdatedByHandler{},
	}
)

// ColumnHandler is the base interface for all auto column handlers.
// It provides the column name that the handler manages.
type ColumnHandler interface {
	// Name returns the name of the column this handler manages.
	Name() string
}

// InsertColumnHandler is an interface for handlers that automatically manage columns during insert operations.
// Handlers implementing this interface will be called before insert operations to set column values.
type InsertColumnHandler interface {
	ColumnHandler
	// OnInsert is called when a new record is being inserted.
	// It allows the handler to automatically set or modify column values.
	OnInsert(query *BunInsertQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value)
}

// UpdateColumnHandler is an interface for handlers that manage columns during both insert and update operations.
// It extends InsertHandler to also handle update scenarios with additional context.
type UpdateColumnHandler interface {
	InsertColumnHandler
	// OnUpdate is called when an existing record is being updated.
	OnUpdate(query *BunUpdateQuery, table *schema.Table, field *schema.Field, model any, value reflect.Value)
}

// processAutoColumns applies auto column handlers to a model before insert/update operations.
// It processes audit.Handler interfaces to automatically manage fields like IDs, timestamps, and user tracking.
func processAutoColumns(query any, table *schema.Table, modelValue any, mv reflect.Value) {
	// Check if the value is valid and not nil
	if !mv.IsValid() || (mv.Kind() == reflect.Ptr && mv.IsNil()) {
		// For nil model values (like (*User)(nil) in update queries), skip audit processing
		// This is common in update queries where we only set specific fields
		return
	}

	// Handle slice values (batch operations) by processing each element
	if mv.Kind() == reflect.Slice {
		for i := 0; i < mv.Len(); i++ {
			elem := mv.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}

			processAutoColumns(query, table, elem.Interface(), elem)
		}

		return
	}

	for _, handler := range autoColumnHandlers {
		if field, ok := table.FieldMap[handler.Name()]; ok {
			value := field.Value(mv)

			// Handle different query types and handler interfaces
			switch q := query.(type) {
			case *BunInsertQuery:
				if insertHandler, ok := handler.(InsertColumnHandler); ok {
					insertHandler.OnInsert(q, table, field, modelValue, value)
				}
			case *BunUpdateQuery:
				if updateHandler, ok := handler.(UpdateColumnHandler); ok {
					updateHandler.OnUpdate(q, table, field, modelValue, value)
				}
			}
		}
	}
}
