package id

import nid "github.com/matoous/go-nanoid/v2"

type randomIdGenerator struct {
	alphabet string
	length   int
}

// Generate creates a new random ID using the configured alphabet and length.
func (g *randomIdGenerator) Generate() string {
	return nid.MustGenerate(g.alphabet, g.length)
}

// NewRandomIdGenerator creates a new random ID generator with custom alphabet and length.
func NewRandomIdGenerator(alphabet string, length int) IdGenerator {
	return &randomIdGenerator{
		alphabet: alphabet,
		length:   length,
	}
}
