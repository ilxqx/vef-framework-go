package cache

import (
	"fmt"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

// KeyBuilder defines the interface for building cache keys with different naming strategies.
type KeyBuilder interface {
	// Build constructs a cache key from the given base key
	Build(keyParts ...string) string
}

// PrefixKeyBuilder implements KeyBuilder with prefix-based naming strategy.
type PrefixKeyBuilder struct {
	prefix    string
	separator string
}

// NewPrefixKeyBuilder creates a new prefix-based key builder with default ":" separator.
func NewPrefixKeyBuilder(prefix string) *PrefixKeyBuilder {
	return &PrefixKeyBuilder{
		prefix:    prefix,
		separator: constants.Colon,
	}
}

// NewPrefixKeyBuilderWithSeparator creates a new prefix-based key builder with custom separator.
func NewPrefixKeyBuilderWithSeparator(prefix, separator string) *PrefixKeyBuilder {
	return &PrefixKeyBuilder{
		prefix:    prefix,
		separator: separator,
	}
}

// Build constructs a cache key with prefix.
func (k *PrefixKeyBuilder) Build(keyParts ...string) string {
	if k.prefix == constants.Empty {
		return strings.Join(keyParts, k.separator)
	}

	return fmt.Sprintf("%s%s%s", k.prefix, k.separator, strings.Join(keyParts, k.separator))
}
