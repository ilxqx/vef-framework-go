package event

import (
	"encoding/json"
	"maps"
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/id"
)

// Event represents the base interface for all events in the system.
// All custom events should embed this interface to be compatible with the event bus.
type Event interface {
	// Id returns a unique identifier for this specific event instance.
	Id() string
	// Type returns a unique string identifier for the event type.
	// This is used for routing and filtering events.
	Type() string
	// Source returns the source that generated this event.
	Source() string
	// Time returns when the event occurred.
	Time() time.Time
	// Meta returns the metadata for the event.
	Meta() map[string]string
}

// BaseEvent provides a default implementation of the Event interface.
// Custom events can embed this struct to inherit the base functionality.
// Fields are unexported to prevent modification after creation.
type BaseEvent struct {
	typ    string
	id     string
	source string
	time   time.Time
	meta   map[string]string
}

func (e BaseEvent) Type() string {
	return e.typ
}

func (e BaseEvent) Time() time.Time {
	return e.time
}

func (e BaseEvent) Id() string {
	return e.id
}

func (e BaseEvent) Source() string {
	return e.source
}

func (e BaseEvent) Meta() map[string]string {
	// Return a copy to prevent external modification
	result := make(map[string]string, len(e.meta))
	maps.Copy(result, e.meta)

	return result
}

// baseEventOption defines an option for configuring BaseEvent creation.
type baseEventOption func(*BaseEvent)

// WithSource sets the event source.
func WithSource(source string) baseEventOption {
	return func(e *BaseEvent) {
		e.source = source
	}
}

// WithMeta adds a metadata key-value pair.
func WithMeta(key, value string) baseEventOption {
	return func(e *BaseEvent) {
		e.meta[key] = value
	}
}

// MarshalJSON implements custom JSON marshaling for BaseEvent.
func (e BaseEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type     string            `json:"type"`
		Id       string            `json:"id"`
		Source   string            `json:"source"`
		Time     time.Time         `json:"time"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}{
		Type:     e.typ,
		Id:       e.id,
		Source:   e.source,
		Time:     e.time,
		Metadata: e.meta,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for BaseEvent.
func (e *BaseEvent) UnmarshalJSON(data []byte) error {
	var temp struct {
		Type     string            `json:"type"`
		Id       string            `json:"id"`
		Source   string            `json:"source"`
		Time     time.Time         `json:"time"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	e.typ = temp.Type
	e.id = temp.Id
	e.source = temp.Source
	e.time = temp.Time

	e.meta = temp.Metadata
	if e.meta == nil {
		e.meta = make(map[string]string)
	}

	return nil
}

// NewBaseEvent creates a new BaseEvent with the specified type.
// It automatically generates a unique ID and sets the current time.
// Optional source and metadata can be set using WithSource and WithMeta options.
func NewBaseEvent(eventType string, opts ...baseEventOption) BaseEvent {
	event := BaseEvent{
		typ:    eventType,
		id:     generateEventId(),
		source: constants.Empty,
		time:   time.Now(),
		meta:   make(map[string]string),
	}

	// Apply all options
	for _, opt := range opts {
		opt(&event)
	}

	return event
}

// generateEventId creates a unique identifier for events.
func generateEventId() string {
	return id.GenerateUuid()
}
