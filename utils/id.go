package utils

import (
	nid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/xid"
)

// GenerateId generates a new unique ID.
func GenerateId() string {
	return xid.New().String()
}

// GenerateRandomId generates a new random ID.
func GenerateRandomId(length ...int) string {
	return nid.Must(length...)
}

// GenerateCustomizedRandomId generates a new random ID with a custom alphabet.
func GenerateCustomizedRandomId(alphabet string, length int) string {
	return nid.MustGenerate(alphabet, length)
}
