package cache

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

var defaultKeyBuilder = NewPrefixKeyBuilder(constants.Empty)

// Key builds a key with the default key builder.
func Key(keyParts ...string) string {
	return defaultKeyBuilder.Build(keyParts...)
}

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

	key := strings.Join(keyParts, k.separator)

	var sb strings.Builder
	sb.Grow(len(k.prefix) + len(k.separator) + len(key))
	_, _ = sb.WriteString(k.prefix)
	_, _ = sb.WriteString(k.separator)
	_, _ = sb.WriteString(key)

	return sb.String()
}
