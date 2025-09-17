package cache

import (
	"bytes"
	"encoding/gob"

	"github.com/goccy/go-json"
)

// Serializer handles serialization/deserialization of cache values.
type Serializer[T any] interface {
	Serialize(value T) ([]byte, error)
	Deserialize(data []byte) (T, error)
}

// gobSerializer implements Serializer using gob encoding.
type gobSerializer[T any] struct{}

func (s gobSerializer[T]) Serialize(value T) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s gobSerializer[T]) Deserialize(data []byte) (T, error) {
	var value T
	buf := bytes.NewReader(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&value)

	return value, err
}

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
func (s jsonSerializer[T]) Deserialize(data []byte) (T, error) {
	var value T
	err := json.Unmarshal(data, &value)
	return value, err
}

// NewGobSerializer creates a new GOB-based serializer.
// GOB provides efficient binary serialization but is Go-specific.
//
// Best for:
//   - Go-only applications
//   - Maximum performance
//   - Complex Go types (interfaces, channels, functions)
func NewGobSerializer[T any]() Serializer[T] {
	return gobSerializer[T]{}
}

// NewJSONSerializer creates a new JSON-based serializer.
// JSON provides human-readable format and cross-language compatibility.
//
// Best for:
//   - Cross-language compatibility
//   - Debugging (human-readable format)
//   - Simple data types (structs, primitives, slices, maps)
//
// Note: JSON has limitations compared to GOB:
//   - Cannot serialize unexported fields
//   - Limited type information preservation
//   - May lose precision for some numeric types
func NewJSONSerializer[T any]() Serializer[T] {
	return jsonSerializer[T]{}
}
