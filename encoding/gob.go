package encoding

import (
	"bytes"
	"encoding/gob"
)

// ToGOB converts a struct value to a GOB byte slice.
func ToGOB(value any) ([]byte, error) {
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// FromGOB converts a GOB byte slice to a struct value.
func FromGOB[T any](data []byte) (*T, error) {
	var result T
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DecodeGOB decodes a GOB byte slice into the provided result pointer.
func DecodeGOB(data []byte, result any) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(result)
}
