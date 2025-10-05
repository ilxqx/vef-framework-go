package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"

	"github.com/ilxqx/vef-framework-go/cache"
	"github.com/ilxqx/vef-framework-go/constants"
)

// badgerStore implements the Store interface using Badger as the storage backend.
type badgerStore struct {
	db      *badger.DB
	options badgerOptions
}

// badgerOptions for cache configuration.
type badgerOptions struct {
	InMemory   bool          // Whether to use in-memory storage
	Directory  string        // Directory path for persistent storage
	DefaultTTL time.Duration // Default TTL for cache entries
}

// createBadgerStore creates a new badger-based store.
func createBadgerStore(opts badgerOptions) (cache.Store, error) {
	// Configure Badger options
	badgerOpts := badger.DefaultOptions(constants.Empty)

	if opts.InMemory {
		// Use pure in-memory mode
		badgerOpts = badgerOpts.WithInMemory(true)
		badgerOpts = badgerOpts.WithDir(constants.Empty)
		badgerOpts = badgerOpts.WithValueDir(constants.Empty)

		// Optimize for in-memory performance: prioritize speed over space
		badgerOpts = badgerOpts.WithCompression(options.None) // No compression for zero latency
		badgerOpts = badgerOpts.WithIndexCacheSize(100 << 20) // 100MB index cache for speed
		badgerOpts = badgerOpts.WithBlockCacheSize(50 << 20)  // 50MB block cache
	} else {
		if opts.Directory == constants.Empty {
			return nil, ErrDirectoryRequired
		}

		badgerOpts = badgerOpts.WithDir(opts.Directory)
		badgerOpts = badgerOpts.WithValueDir(opts.Directory)

		// Optimize for persistent storage: balance performance and space
		badgerOpts = badgerOpts.WithCompression(options.Snappy) // Light compression for disk space
		badgerOpts = badgerOpts.WithIndexCacheSize(50 << 20)    // 50MB index cache (smaller)
		badgerOpts = badgerOpts.WithBlockCacheSize(25 << 20)    // 25MB block cache (smaller)
	}

	// Common optimizations for both modes
	badgerOpts = badgerOpts.WithLoggingLevel(badger.WARNING)

	// Open the database
	db, err := badger.Open(badgerOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger database: %w", err)
	}

	store := &badgerStore{
		db:      db,
		options: opts,
	}

	// Start garbage collection for TTL entries (only for persistent mode)
	if !opts.InMemory {
		go store.runGC()
	}

	return store, nil
}

// runGC runs garbage collection to clean up expired entries.
// This is only necessary for persistent mode as in-memory mode doesn't have value logs.
// The GC will automatically stop when the database is closed (ErrRejected).
func (s *badgerStore) runGC() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Run garbage collection to reclaim disk space from deleted/expired entries
		err := s.db.RunValueLogGC(0.5) // Use recommended 0.5 threshold
		if err != nil {
			if errors.Is(err, badger.ErrNoRewrite) {
				// No cleanup was needed, continue
				continue
			}

			if errors.Is(err, badger.ErrRejected) {
				// Database is closed or another GC is running, stop GC
				logger.Info("garbage collection stopped because database is closed or another GC is running")

				return
			}

			// Other errors are not critical, just continue
			logger.Errorf("failed to run garbage collection: %v", err)

			continue
		}
	}
}

// Name returns the name of the store.
func (s *badgerStore) Name() string {
	return "badger"
}

// Get retrieves raw bytes by key.
func (s *badgerStore) Get(ctx context.Context, key string) ([]byte, bool) {
	var (
		data  []byte
		found bool
	)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil // Not an error, just not found
			}

			return err
		}

		return item.Value(func(val []byte) error {
			// Make a copy of the value since Badger reuses the slice
			data = make([]byte, len(val))
			copy(data, val)

			found = true

			return nil
		})
	})
	if err != nil {
		logger.Errorf("failed to get value: %v", err)

		return nil, false
	}

	return data, found
}

// Set stores raw bytes with the given key and optional TTL.
func (s *badgerStore) Set(ctx context.Context, key string, data []byte, ttl ...time.Duration) error {
	return s.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), data)

		// Set TTL if provided
		if len(ttl) > 0 && ttl[0] > 0 {
			entry = entry.WithTTL(ttl[0])
		} else if s.options.DefaultTTL > 0 {
			entry = entry.WithTTL(s.options.DefaultTTL)
		}

		return txn.SetEntry(entry)
	})
}

// Contains checks if a key exists in the cache.
func (s *badgerStore) Contains(ctx context.Context, key string) bool {
	err := s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))

		return err
	})

	return err == nil
}

// Delete removes a key from the cache.
func (s *badgerStore) Delete(ctx context.Context, key string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// Clear removes all entries from the cache with the given prefix.
func (s *badgerStore) Clear(ctx context.Context, prefix string) error {
	if prefix == constants.Empty {
		// If no prefix, clear all entries
		return s.db.DropAll()
	}

	// With prefix, we need to find and delete all matching keys
	prefixBytes := []byte(prefix)

	return s.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		var keysToDelete [][]byte
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			key := make([]byte, len(it.Item().Key()))
			copy(key, it.Item().Key())
			keysToDelete = append(keysToDelete, key)
		}

		// Delete all matching keys
		for _, key := range keysToDelete {
			if err := txn.Delete(key); err != nil {
				return err
			}
		}

		return nil
	})
}

// Keys returns all keys in the cache, filtered by prefix.
func (s *badgerStore) Keys(ctx context.Context, prefix string) ([]string, error) {
	var (
		keys        []string
		prefixBytes []byte
	)

	if prefix != constants.Empty {
		prefixBytes = []byte(prefix)
	}

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need keys

		it := txn.NewIterator(opts)
		defer it.Close()

		if prefixBytes != nil {
			for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
				item := it.Item()
				key := string(item.Key())
				keys = append(keys, key)
			}
		} else {
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				key := string(item.Key())
				keys = append(keys, key)
			}
		}

		return nil
	})

	return keys, err
}

// processIteratorItem processes a single item from the badger iterator
// and calls the callback with the key-value pair.
// Returns false if the callback wants to stop iteration, true otherwise.
func (s *badgerStore) processIteratorItem(item *badger.Item, callback func(key string, data []byte) bool) (bool, error) {
	key := string(item.Key())

	var data []byte

	err := item.Value(func(val []byte) error {
		// Make a copy of the value since Badger reuses the slice
		data = make([]byte, len(val))
		copy(data, val)

		return nil
	})
	if err != nil {
		return false, err
	}

	// Call the callback, return false if it wants to stop iteration
	return callback(key, data), nil
}

// ForEach iterates over all key-value pairs in the cache, filtered by prefix.
func (s *badgerStore) ForEach(ctx context.Context, prefix string, callback func(key string, data []byte) bool) error {
	var prefixBytes []byte

	if prefix != constants.Empty {
		prefixBytes = []byte(prefix)
	}

	return s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10

		it := txn.NewIterator(opts)
		defer it.Close()

		if prefixBytes != nil {
			for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
				shouldContinue, err := s.processIteratorItem(it.Item(), callback)
				if err != nil {
					return err
				}

				if !shouldContinue {
					break
				}
			}
		} else {
			for it.Rewind(); it.Valid(); it.Next() {
				shouldContinue, err := s.processIteratorItem(it.Item(), callback)
				if err != nil {
					return err
				}

				if !shouldContinue {
					break
				}
			}
		}

		return nil
	})
}

// Size returns the number of entries in the cache, filtered by prefix.
func (s *badgerStore) Size(ctx context.Context, prefix string) (int64, error) {
	var (
		count       int64
		prefixBytes []byte
	)

	if prefix != constants.Empty {
		prefixBytes = []byte(prefix)
	}

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		if prefixBytes != nil {
			for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
				count++
			}
		} else {
			for it.Rewind(); it.Valid(); it.Next() {
				count++
			}
		}

		return nil
	})

	return count, err
}

// Close closes the cache and releases resources.
// The GC goroutine will automatically stop when the database is closed.
func (s *badgerStore) Close(_ context.Context) error {
	return s.db.Close()
}
