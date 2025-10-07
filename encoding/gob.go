package encoding

import (
	"bytes"
	"encoding/gob"
)

// ToGOB converts a struct value to a GOB byte slice.
func ToGOB(value any) ([]byte, error) {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)

	if err := encoder.Encode(value); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// FromGOB converts a GOB byte slice to a struct value.
func FromGOB[T any](data []byte) (*T, error) {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	var result T
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeGOB decodes a GOB byte slice into the provided result pointer.
func DecodeGOB(data []byte, result any) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	return decoder.Decode(result)
}
