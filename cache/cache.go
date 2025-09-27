package cache

import "github.com/samber/lo"

// New creates a new cache instance with the given store and uses gob serialization.
func New[T any](name string, store Store) Cache[T] {
	serializer := lo.TernaryF(
		store.Name() == "redis",
		NewJSONSerializer[T],
		NewGobSerializer[T],
	)

	return &cacheAdapter[T]{
		store:      store,
		serializer: serializer,
		keyBuilder: NewPrefixKeyBuilder(cacheKeyPrefix + name),
	}
}

// NewWithSerializer creates a new cache instance with a custom serializer.
func NewWithSerializer[T any](name string, store Store, serializer Serializer[T]) Cache[T] {
	return &cacheAdapter[T]{
		store:      store,
		serializer: serializer,
		keyBuilder: NewPrefixKeyBuilder(cacheKeyPrefix + name),
	}
}
