package encoding

import (
	"encoding/base64"

	"github.com/ilxqx/vef-framework-go/constants"
)

// ToBase64 encodes binary data to a base64 string using standard encoding.
func ToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// FromBase64 decodes a base64 string to binary data using standard encoding.
func FromBase64(s string) ([]byte, error) {
	if s == constants.Empty {
		return nil, nil
	}

	return base64.StdEncoding.DecodeString(s)
}

// ToBase64URL encodes binary data to a base64 URL-safe string.
func ToBase64URL(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// FromBase64URL decodes a base64 URL-safe string to binary data.
func FromBase64URL(s string) ([]byte, error) {
	if s == constants.Empty {
		return nil, nil
	}

	return base64.URLEncoding.DecodeString(s)
}
