package security

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryNonceStore(t *testing.T) {
	t.Run("CreatesValidStore", func(t *testing.T) {
		store := NewMemoryNonceStore()

		assert.NotNil(t, store, "Store should not be nil")
		_, ok := store.(*MemoryNonceStore)
		assert.True(t, ok, "Should return *MemoryNonceStore")
	})

	t.Run("ImplementsNonceStoreInterface", func(*testing.T) {
		_ = NewMemoryNonceStore()
	})
}

func TestMemoryNonceStore_Exists(t *testing.T) {
	ctx := context.Background()

	t.Run("NonExistentNonce", func(t *testing.T) {
		store := NewMemoryNonceStore()

		exists, err := store.Exists(ctx, "test-app", "test-nonce")

		require.NoError(t, err, "Should check existence without error")
		assert.False(t, exists, "Nonce should not exist initially")
	})

	t.Run("ExistingNonce", func(t *testing.T) {
		store := NewMemoryNonceStore()

		err := store.Store(ctx, "test-app", "test-nonce", 5*time.Minute)
		require.NoError(t, err, "Should store nonce without error")

		exists, err := store.Exists(ctx, "test-app", "test-nonce")

		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "Nonce should exist after storing")
	})

	t.Run("EmptyAppID", func(t *testing.T) {
		store := NewMemoryNonceStore()

		err := store.Store(ctx, "", "test-nonce", 5*time.Minute)
		require.NoError(t, err, "Should store nonce with empty appID without error")

		exists, err := store.Exists(ctx, "", "test-nonce")

		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "Should handle empty appID")
	})

	t.Run("EmptyNonce", func(t *testing.T) {
		store := NewMemoryNonceStore()

		err := store.Store(ctx, "test-app", "", 5*time.Minute)
		require.NoError(t, err, "Should store empty nonce without error")

		exists, err := store.Exists(ctx, "test-app", "")

		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "Should handle empty nonce")
	})
}

func TestMemoryNonceStore_Store(t *testing.T) {
	ctx := context.Background()

	t.Run("StoreNewNonce", func(t *testing.T) {
		store := NewMemoryNonceStore()

		err := store.Store(ctx, "test-app", "test-nonce", 5*time.Minute)

		require.NoError(t, err, "Should store new nonce without error")

		exists, err := store.Exists(ctx, "test-app", "test-nonce")
		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "Stored nonce should exist")
	})

	t.Run("StoreDuplicateNonce", func(t *testing.T) {
		store := NewMemoryNonceStore()

		err := store.Store(ctx, "test-app", "test-nonce", 5*time.Minute)
		require.NoError(t, err, "Should store first nonce without error")

		err = store.Store(ctx, "test-app", "test-nonce", 5*time.Minute)
		require.NoError(t, err, "Should store duplicate nonce without error")

		exists, err := store.Exists(ctx, "test-app", "test-nonce")
		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "Duplicate nonce should still exist")
	})

	t.Run("StoreWithZeroTTL", func(t *testing.T) {
		store := NewMemoryNonceStore()

		err := store.Store(ctx, "test-app", "test-nonce", 0)

		require.NoError(t, err, "Should store nonce with zero TTL without error")
	})

	t.Run("StoreWithNegativeTTL", func(t *testing.T) {
		store := NewMemoryNonceStore()

		err := store.Store(ctx, "test-app", "test-nonce", -1*time.Minute)

		require.NoError(t, err, "Should store nonce with negative TTL without error")
	})
}

func TestMemoryNonceStore_DifferentApps(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryNonceStore()

	t.Run("SameNonceDifferentApps", func(t *testing.T) {
		err := store.Store(ctx, "test-app-1", "shared-nonce", 5*time.Minute)
		require.NoError(t, err, "Should store nonce for first app without error")

		exists, err := store.Exists(ctx, "test-app-2", "shared-nonce")
		require.NoError(t, err, "Should check existence without error")
		assert.False(t, exists, "Same nonce for different app should not exist")

		exists, err = store.Exists(ctx, "test-app-1", "shared-nonce")
		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "Original app should still have the nonce")
	})

	t.Run("DifferentNoncesSameApp", func(t *testing.T) {
		err := store.Store(ctx, "test-app-3", "nonce-1", 5*time.Minute)
		require.NoError(t, err, "Should store first nonce without error")

		exists, err := store.Exists(ctx, "test-app-3", "nonce-2")
		require.NoError(t, err, "Should check existence without error")
		assert.False(t, exists, "Different nonce for same app should not exist")

		exists, err = store.Exists(ctx, "test-app-3", "nonce-1")
		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "Original nonce should still exist")
	})

	t.Run("MultipleAppsMultipleNonces", func(t *testing.T) {
		apps := []string{"app-a", "app-b", "app-c"}
		nonces := []string{"nonce-x", "nonce-y", "nonce-z"}

		for _, app := range apps {
			for _, nonce := range nonces {
				err := store.Store(ctx, app, nonce, 5*time.Minute)
				require.NoError(t, err, "Should store nonce %s for app %s without error", nonce, app)
			}
		}

		for _, app := range apps {
			for _, nonce := range nonces {
				exists, err := store.Exists(ctx, app, nonce)
				require.NoError(t, err, "Should check existence without error")
				assert.True(t, exists, "Nonce %s for app %s should exist", nonce, app)
			}
		}
	})
}

func TestMemoryNonceStore_TTLExpiration(t *testing.T) {
	ctx := context.Background()

	t.Run("NonceExpiresAfterTTL", func(t *testing.T) {
		store := NewMemoryNonceStore()

		err := store.Store(ctx, "test-app", "expiring-nonce", 50*time.Millisecond)
		require.NoError(t, err, "Should store nonce without error")

		exists, err := store.Exists(ctx, "test-app", "expiring-nonce")
		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "Nonce should exist immediately after storing")

		time.Sleep(100 * time.Millisecond)

		exists, err = store.Exists(ctx, "test-app", "expiring-nonce")
		require.NoError(t, err, "Should check existence without error")
		assert.False(t, exists, "Nonce should not exist after TTL expiration")
	})

	t.Run("DifferentTTLsForDifferentNonces", func(t *testing.T) {
		store := NewMemoryNonceStore()

		err := store.Store(ctx, "test-app", "short-ttl", 50*time.Millisecond)
		require.NoError(t, err, "Should store short TTL nonce without error")

		err = store.Store(ctx, "test-app", "long-ttl", 5*time.Minute)
		require.NoError(t, err, "Should store long TTL nonce without error")

		time.Sleep(100 * time.Millisecond)

		exists, err := store.Exists(ctx, "test-app", "short-ttl")
		require.NoError(t, err, "Should check existence without error")
		assert.False(t, exists, "Short TTL nonce should have expired")

		exists, err = store.Exists(ctx, "test-app", "long-ttl")
		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "Long TTL nonce should still exist")
	})

	t.Run("NonceExistsBeforeTTL", func(t *testing.T) {
		store := NewMemoryNonceStore()

		err := store.Store(ctx, "test-app", "test-nonce", 1*time.Second)
		require.NoError(t, err, "Should store nonce without error")

		time.Sleep(100 * time.Millisecond)

		exists, err := store.Exists(ctx, "test-app", "test-nonce")
		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "Nonce should still exist before TTL expires")
	})
}

func TestMemoryNonceStore_MultipleNonces(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryNonceStore()

	t.Run("StoreAndVerifyMultipleNonces", func(t *testing.T) {
		nonces := []string{"nonce-1", "nonce-2", "nonce-3", "nonce-4", "nonce-5"}

		for _, nonce := range nonces {
			err := store.Store(ctx, "test-app", nonce, 5*time.Minute)
			require.NoError(t, err, "Should store nonce %s without error", nonce)
		}

		for _, nonce := range nonces {
			exists, err := store.Exists(ctx, "test-app", nonce)
			require.NoError(t, err, "Should check existence without error")
			assert.True(t, exists, "Nonce %s should exist", nonce)
		}

		exists, err := store.Exists(ctx, "test-app", "non-existent")
		require.NoError(t, err, "Should check existence without error")
		assert.False(t, exists, "Non-stored nonce should not exist")
	})

	t.Run("LargeNumberOfNonces", func(t *testing.T) {
		largeStore := NewMemoryNonceStore()

		for i := 0; i < 1000; i++ {
			nonce := "nonce-" + string(rune('a'+i%26)) + string(rune('0'+i%10))
			err := largeStore.Store(ctx, "test-app", nonce, 5*time.Minute)
			require.NoError(t, err, "Should store nonce without error")
		}

		exists, err := largeStore.Exists(ctx, "test-app", "nonce-a0")
		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "First nonce should exist after storing many nonces")
	})
}

func TestMemoryNonceStore_KeyFormat(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryNonceStore()

	t.Run("SpecialCharactersInAppID", func(t *testing.T) {
		specialAppIDs := []string{
			"app:with:colons",
			"app-with-dashes",
			"app_with_underscores",
			"app.with.dots",
			"app/with/slashes",
		}

		for _, appID := range specialAppIDs {
			err := store.Store(ctx, appID, "test-nonce", 5*time.Minute)
			require.NoError(t, err, "Should store nonce for appID %s without error", appID)

			exists, err := store.Exists(ctx, appID, "test-nonce")
			require.NoError(t, err, "Should check existence without error")
			assert.True(t, exists, "Should handle appID: %s", appID)
		}
	})

	t.Run("SpecialCharactersInNonce", func(t *testing.T) {
		specialNonces := []string{
			"nonce:with:colons",
			"nonce-with-dashes",
			"nonce_with_underscores",
			"nonce.with.dots",
			"nonce/with/slashes",
		}

		for _, nonce := range specialNonces {
			err := store.Store(ctx, "test-app", nonce, 5*time.Minute)
			require.NoError(t, err, "Should store nonce %s without error", nonce)

			exists, err := store.Exists(ctx, "test-app", nonce)
			require.NoError(t, err, "Should check existence without error")
			assert.True(t, exists, "Should handle nonce: %s", nonce)
		}
	})

	t.Run("UnicodeCharacters", func(t *testing.T) {
		err := store.Store(ctx, "应用", "随机数", 5*time.Minute)
		require.NoError(t, err, "Should store nonce with unicode characters without error")

		exists, err := store.Exists(ctx, "应用", "随机数")
		require.NoError(t, err, "Should check existence without error")
		assert.True(t, exists, "Should handle unicode characters")
	})
}

func TestMemoryNonceStore_Concurrency(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryNonceStore()

	t.Run("ConcurrentStoreAndExists", func(t *testing.T) {
		var wg sync.WaitGroup

		numGoroutines := 100

		for i := range numGoroutines {
			id := i
			wg.Go(func() {
				appID := "test-app"
				nonce := "nonce-" + string(rune('a'+id%26))

				err := store.Store(ctx, appID, nonce, 5*time.Minute)
				assert.NoError(t, err, "Should store nonce without error in goroutine %d", id)

				_, err = store.Exists(ctx, appID, nonce)
				assert.NoError(t, err, "Should check existence without error in goroutine %d", id)
			})
		}

		wg.Wait()
	})

	t.Run("ConcurrentDifferentApps", func(t *testing.T) {
		var wg sync.WaitGroup

		numApps := 50

		for i := range numApps {
			id := i
			wg.Go(func() {
				appID := "app-" + string(rune('a'+id%26))
				nonce := "nonce-" + string(rune('0'+id%10))

				err := store.Store(ctx, appID, nonce, 5*time.Minute)
				assert.NoError(t, err, "Should store nonce without error for app %s", appID)

				exists, err := store.Exists(ctx, appID, nonce)
				assert.NoError(t, err, "Should check existence without error for app %s", appID)
				assert.True(t, exists, "Nonce should exist for app %s", appID)
			})
		}

		wg.Wait()
	})
}

func TestMemoryNonceStore_ContextHandling(t *testing.T) {
	t.Run("CancelledContext", func(t *testing.T) {
		store := NewMemoryNonceStore()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := store.Store(ctx, "test-app", "test-nonce", 5*time.Minute)
		require.NoError(t, err, "Should store nonce with canceled context without error")

		exists, err := store.Exists(ctx, "test-app", "test-nonce")
		require.NoError(t, err, "Should check existence with canceled context without error")
		assert.True(t, exists, "Nonce should exist even with canceled context")
	})

	t.Run("TimeoutContext", func(t *testing.T) {
		store := NewMemoryNonceStore()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(1 * time.Millisecond)

		err := store.Store(ctx, "test-app", "test-nonce", 5*time.Minute)
		require.NoError(t, err, "Should store nonce with timeout context without error")
	})
}
