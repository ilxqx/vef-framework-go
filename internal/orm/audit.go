package orm

import (
	"reflect"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"

	"github.com/ilxqx/vef-framework-go/internal/orm/audit"
)

// autoColumns is the list of auto column handlers that are applied to all models.
// These handlers automatically manage audit fields like ID generation, timestamps, and user tracking.
var (
	autoColumns = audit.DefaultHandlers()
)

// processAutoColumns applies auto column handlers to a model before insert/update operations.
// It processes audit.Handler interfaces to automatically manage fields like IDs, timestamps, and user tracking.
func processAutoColumns(handlers []audit.Handler, query any, hasSet bool, table *schema.Table, modelValue any, mv reflect.Value) {
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

			processAutoColumns(handlers, query, hasSet, table, elem.Interface(), elem)
		}

		return
	}

	for _, handler := range handlers {
		if field, ok := table.FieldMap[handler.Name()]; ok {
			value := field.Value(mv)

			// Handle different query types and handler interfaces
			switch q := query.(type) {
			case *bun.InsertQuery:
				if insertHandler, ok := handler.(audit.InsertHandler); ok {
					insertHandler.OnInsert(q, table, field, modelValue, value)
				}
			case *bun.UpdateQuery:
				if updateHandler, ok := handler.(audit.UpdateHandler); ok {
					updateHandler.OnUpdate(q, hasSet, table, field, modelValue, value)
				} else if _, ok := handler.(audit.InsertHandler); ok {
					// Exclude insert-only handlers from update operations
					q.ExcludeColumn(field.Name)
				}
			}
		}
	}
}
