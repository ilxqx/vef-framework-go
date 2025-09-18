package utils

import (
	"encoding/xml"

	"github.com/goccy/go-json"
	"github.com/ilxqx/vef-framework-go/constants"
)

// ToJSON converts a struct value to a JSON string.
func ToJSON(value any) (string, error) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return constants.Empty, err
	}

	return string(jsonBytes), nil
}

// FromJSON converts a JSON string to a struct value.
func FromJSON[T any](value string) (*T, error) {
	var result T
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ToXML converts a struct value to an XML string.
func ToXML(value any) (string, error) {
	xmlBytes, err := xml.Marshal(value)
	if err != nil {
		return constants.Empty, err
	}

	return string(xmlBytes), nil
}

// FromXML converts an XML string to a struct value.
func FromXML[T any](value string) (*T, error) {
	var result T
	if err := xml.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
