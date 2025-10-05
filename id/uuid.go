package id

import (
	"fmt"

	"github.com/google/uuid"
)

// DefaultUuidIdGenerator is the default UUID v7 generator instance.
// UUID v7 provides time-based ordering and follows RFC 4122 standards.
// It produces 36-character strings in the format: xxxxxxxx-xxxx-7xxx-xxxx-xxxxxxxxxxxx.
var DefaultUuidIdGenerator = NewUuidIdGenerator()

// uuidIdGenerator implements IdGenerator using UUID v7 algorithm.
// UUID v7 combines timestamp with random data for uniqueness and time-based sorting.
type uuidIdGenerator struct{}

// Generate creates a new UUID v7 as a 36-character hyphenated string.
// UUID v7 embeds a timestamp for natural ordering and includes random bits for uniqueness.
// The format follows RFC 4122 standard: xxxxxxxx-xxxx-7xxx-xxxx-xxxxxxxxxxxx
// where the '7' indicates version 7.
func (g *uuidIdGenerator) Generate() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic(
			fmt.Errorf("failed to generate uuid: %w", err),
		)
	}

	return id.String()
}

// NewUuidIdGenerator creates a new UUID v7 generator instance.
// UUID v7 is recommended when you need:
//   - Standards compliance (RFC 4122)
//   - Time-based sorting
//   - Compatibility with existing UUID-based systems
//   - Database-friendly format
//
// Note: UUID v7 is slightly slower than XID but provides better standards compliance.
//
// Example:
//
//	gen := NewUuidIdGenerator()
//	id := gen.Generate()  // Returns something like "018f4e42-832a-7123-9abc-def012345678"
func NewUuidIdGenerator() IdGenerator {
	return &uuidIdGenerator{}
}
