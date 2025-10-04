package encoding

import (
	"encoding/hex"

	"github.com/ilxqx/vef-framework-go/constants"
)

// ToHex encodes binary data to a hexadecimal string.
func ToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// FromHex decodes a hexadecimal string to binary data.
func FromHex(s string) ([]byte, error) {
	if s == constants.Empty {
		return []byte{}, nil
	}

	return hex.DecodeString(s)
}
