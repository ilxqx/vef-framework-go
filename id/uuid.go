package id

import (
	"fmt"

	"github.com/google/uuid"
)

// DefaultUuidIdGenerator is the default UUID v7 generator instance.
// UUID v7 embeds a timestamp for natural ordering and includes random bits for uniqueness.
var DefaultUuidIdGenerator = NewUuidIdGenerator()

type uuidIdGenerator struct{}

// Generate creates a new UUID v7 as a 36-character hyphenated string.
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
func NewUuidIdGenerator() IdGenerator {
	return &uuidIdGenerator{}
}
