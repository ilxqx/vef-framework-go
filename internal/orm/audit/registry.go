package audit

// DefaultHandlers returns the list of default auto column handlers.
// These handlers automatically manage audit fields like ID generation, timestamps, and user tracking.
func DefaultHandlers() []Handler {
	return []Handler{
		&idGenerator{},
		&CreatedAtHandler{},
		&UpdatedAtHandler{},
		&CreatedByHandler{},
		&UpdatedByHandler{},
	}
}
