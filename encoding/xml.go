package encoding

import (
	"encoding/xml"

	"github.com/ilxqx/vef-framework-go/constants"
)

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

// DecodeXML decodes an XML string into the provided result pointer.
func DecodeXML(value string, result any) error {
	return xml.Unmarshal([]byte(value), result)
}
