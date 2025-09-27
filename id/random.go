package id

import nid "github.com/matoous/go-nanoid/v2"

// randomIdGenerator implements IdGenerator using customizable random generation.
// It allows full control over the character set and length of generated IDs.
type randomIdGenerator struct {
	alphabet string // The character set to use for generation
	length   int    // The length of generated IDs
}

// Generate creates a new random ID using the configured alphabet and length.
// The randomness is cryptographically secure and the distribution is uniform.
func (g *randomIdGenerator) Generate() string {
	return nid.MustGenerate(g.alphabet, g.length)
}

// NewRandomIdGenerator creates a new random ID generator with custom alphabet and length.
// This generator is useful when you need specific requirements:
//   - Custom character sets (e.g., only numbers, only letters)
//   - Specific length requirements
//   - Avoiding confusing characters (e.g., 0, O, I, l)
//   - Custom encoding schemes
//
// Parameters:
//   - alphabet: The character set to use (must not be empty)
//   - length: The length of generated IDs (must be positive)
//
// Common alphabets:
//   - Numbers only: "0123456789"
//   - Hex: "0123456789abcdef"
//   - Base62: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
//   - URL-safe: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-"
//
// Example:
//
//	// Numeric IDs only
//	numGen := NewRandomIdGenerator("0123456789", 12)
//	numId := numGen.Generate()  // Returns something like "847392759284"
//
//	// Short hex IDs
//	hexGen := NewRandomIdGenerator("0123456789abcdef", 8)
//	hexId := hexGen.Generate()  // Returns something like "a3f7b2e9"
func NewRandomIdGenerator(alphabet string, length int) IdGenerator {
	return &randomIdGenerator{
		alphabet: alphabet,
		length:   length,
	}
}
