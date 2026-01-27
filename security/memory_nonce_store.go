package security

import (
	"context"
	"time"

	"github.com/ilxqx/vef-framework-go/cache"
	"github.com/ilxqx/vef-framework-go/constants"
)

// MemoryNonceStore implements NonceStore using an in-memory cache.
// This implementation is suitable for development and single-instance deployments.
// For distributed systems, use RedisNonceStore instead.
type MemoryNonceStore struct {
	cache cache.Cache[bool]
}

// NewMemoryNonceStore creates a new in-memory nonce store.
func NewMemoryNonceStore() NonceStore {
	return &MemoryNonceStore{
		cache: cache.NewMemory[bool](),
	}
}

// buildKey creates a unique cache key for the app-nonce combination.
func (*MemoryNonceStore) buildKey(appID, nonce string) string {
	return appID + constants.Colon + nonce
}

// Exists checks if a nonce has already been used for the given app.
func (m *MemoryNonceStore) Exists(ctx context.Context, appID, nonce string) (bool, error) {
	key := m.buildKey(appID, nonce)

	return m.cache.Contains(ctx, key), nil
}

// Store saves a nonce with the specified TTL.
func (m *MemoryNonceStore) Store(ctx context.Context, appID, nonce string, ttl time.Duration) error {
	key := m.buildKey(appID, nonce)

	return m.cache.Set(ctx, key, true, ttl)
}
