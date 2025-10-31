package encoding

import (
	"bytes"
	"encoding/gob"
)

// ToGob converts a struct value to a GOB byte slice.
func ToGob(value any) ([]byte, error) {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)

	if err := encoder.Encode(value); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// FromGob converts a GOB byte slice to a struct value.
func FromGob[T any](data []byte) (*T, error) {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	var result T
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeGob decodes a GOB byte slice into the provided result pointer.
func DecodeGob(data []byte, result any) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	return decoder.Decode(result)
}
