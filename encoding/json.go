package encoding

import (
	"encoding/json"

	"github.com/ilxqx/vef-framework-go/constants"
)

// ToJson converts a struct value to a JSON string.
func ToJson(value any) (string, error) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return constants.Empty, err
	}

	return string(jsonBytes), nil
}

// FromJson converts a JSON string to a struct value.
func FromJson[T any](value string) (*T, error) {
	var result T
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeJson decodes a JSON string into the provided result pointer.
func DecodeJson(value string, result any) error {
	return json.Unmarshal([]byte(value), result)
}
