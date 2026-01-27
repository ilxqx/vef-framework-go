package event

import (
	"encoding/json"
	"maps"
	"time"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/id"
)

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

func (e BaseEvent) ID() string {
	return e.id
}

func (e BaseEvent) Source() string {
	return e.source
}

func (e BaseEvent) Meta() map[string]string {
	result := make(map[string]string, len(e.meta))
	maps.Copy(result, e.meta)

	return result
}

// BaseEventOption defines an option for configuring BaseEvent creation.
type BaseEventOption func(*BaseEvent)

func WithSource(source string) BaseEventOption {
	return func(e *BaseEvent) {
		e.source = source
	}
}

// WithMeta adds a metadata key-value pair.
func WithMeta(key, value string) BaseEventOption {
	return func(e *BaseEvent) {
		e.meta[key] = value
	}
}

// MarshalJSON implements custom JSON marshaling for BaseEvent.
func (e BaseEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type     string            `json:"type"`
		ID       string            `json:"id"`
		Source   string            `json:"source"`
		Time     time.Time         `json:"time"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}{
		Type:     e.typ,
		ID:       e.id,
		Source:   e.source,
		Time:     e.time,
		Metadata: e.meta,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for BaseEvent.
func (e *BaseEvent) UnmarshalJSON(data []byte) error {
	var temp struct {
		Type     string            `json:"type"`
		ID       string            `json:"id"`
		Source   string            `json:"source"`
		Time     time.Time         `json:"time"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	e.typ = temp.Type
	e.id = temp.ID
	e.source = temp.Source
	e.time = temp.Time

	e.meta = temp.Metadata
	if e.meta == nil {
		e.meta = make(map[string]string)
	}

	return nil
}

// NewBaseEvent creates a new BaseEvent with the specified type.
// Optional source and metadata can be set using WithSource and WithMeta options.
func NewBaseEvent(eventType string, opts ...BaseEventOption) BaseEvent {
	event := BaseEvent{
		typ:    eventType,
		id:     id.GenerateUUID(),
		source: constants.Empty,
		time:   time.Now(),
		meta:   make(map[string]string),
	}

	for _, opt := range opts {
		opt(&event)
	}

	return event
}
