package encoding

import (
	"encoding/xml"

	"github.com/ilxqx/vef-framework-go/constants"
)

// ToXml converts a struct value to an XML string.
func ToXml(value any) (string, error) {
	xmlBytes, err := xml.Marshal(value)
	if err != nil {
		return constants.Empty, err
	}

	return string(xmlBytes), nil
}

// FromXml converts an XML string to a struct value.
func FromXml[T any](value string) (*T, error) {
	var result T
	if err := xml.Unmarshal([]byte(value), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeXml decodes an XML string into the provided result pointer.
func DecodeXml(value string, result any) error {
	return xml.Unmarshal([]byte(value), result)
}
