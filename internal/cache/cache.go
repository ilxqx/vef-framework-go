package cache

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	"github.com/redis/go-redis/v9"

	"github.com/ilxqx/vef-framework-go/cache"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/log"
)

var logger = log.Named("cache")

// NewBadgerStore creates a new badger-based store.
func NewBadgerStore(cfg *config.LocalCacheConfig) (cache.Store, error) {
	// Configure Badger options
	badgerOpts := badger.DefaultOptions(constants.Empty)

	if cfg.InMemory {
		// Use pure in-memory mode
		badgerOpts = badgerOpts.WithInMemory(true)
		badgerOpts = badgerOpts.WithDir(constants.Empty)
		badgerOpts = badgerOpts.WithValueDir(constants.Empty)

		// Optimize for in-memory performance: prioritize speed over space
		badgerOpts = badgerOpts.WithCompression(options.None) // No compression for zero latency
		badgerOpts = badgerOpts.WithIndexCacheSize(100 << 20) // 100MB index cache for speed
		badgerOpts = badgerOpts.WithBlockCacheSize(50 << 20)  // 50MB block cache
	} else {
		if cfg.Directory == constants.Empty {
			return nil, ErrDirectoryRequired
		}

		badgerOpts = badgerOpts.WithDir(cfg.Directory)
		badgerOpts = badgerOpts.WithValueDir(cfg.Directory)

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

	store := &BadgerStore{
		db:  db,
		cfg: cfg,
	}

	// Start garbage collection for TTL entries (only for persistent mode)
	if !cfg.InMemory {
		go store.runGC()
	}

	return store, nil
}

// NewRedisStore creates a new redis-based store.
func NewRedisStore(client *redis.Client, cfg *config.RedisCacheConfig) cache.Store {
	return &RedisStore{
		client: client,
		cfg:    cfg,
	}
}
