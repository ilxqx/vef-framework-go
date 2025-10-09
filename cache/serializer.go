package cache

import (
	"github.com/goccy/go-json"
)

// jsonSerializer implements Serializer using JSON encoding.
// It provides human-readable serialization format and cross-language compatibility.
type jsonSerializer[T any] struct{}

// Serialize converts a value to JSON bytes.
// Returns an error if the value cannot be marshaled to JSON.
func (s jsonSerializer[T]) Serialize(value T) ([]byte, error) {
	return json.Marshal(value)
}

// Deserialize converts JSON bytes back to a value of type T.
// Returns an error if the data is not valid JSON or cannot be unmarshaled to type T.
func (s jsonSerializer[T]) Deserialize(data []byte) (value T, err error) {
	err = json.Unmarshal(data, &value)

	return value, err
}

// newJSONSerializer creates a new JSON-based serializer.
func newJSONSerializer[T any]() Serializer[T] {
	return jsonSerializer[T]{}
}
