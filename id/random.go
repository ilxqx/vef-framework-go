package id

import nid "github.com/matoous/go-nanoid/v2"

type randomIDGenerator struct {
	alphabet string
	length   int
}

// Generate creates a new random ID using the configured alphabet and length.
func (g *randomIDGenerator) Generate() string {
	return nid.MustGenerate(g.alphabet, g.length)
}

// NewRandomIDGenerator creates a new random ID generator with custom alphabet and length.
func NewRandomIDGenerator(alphabet string, length int) IDGenerator {
	return &randomIDGenerator{
		alphabet: alphabet,
		length:   length,
	}
}
