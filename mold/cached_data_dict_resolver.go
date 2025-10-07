package mold

import (
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"

	"github.com/ilxqx/vef-framework-go/cache"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/internal/log"
	logPkg "github.com/ilxqx/vef-framework-go/log"
)

const (
	// EventTypeDataDictChanged represents the event type used to invalidate cached dictionary values.
	eventTypeDataDictChanged = "translate.data_dict.changed"
)

// DataDictLoaderFunc allows using a plain function as a DataDictLoader.
type DataDictLoaderFunc func(ctx context.Context, key string) (map[string]string, error)

// Load executes the wrapped function.
func (f DataDictLoaderFunc) Load(ctx context.Context, key string) (map[string]string, error) {
	return f(ctx, key)
}

// DataDictChangedEvent is emitted whenever dictionary entries need to be invalidated.
type DataDictChangedEvent struct {
	event.BaseEvent

	// Keys lists the affected dictionary keys. When empty, all cached dictionaries should be cleared.
	Keys []string `json:"keys"`
}

// PublishDataDictChangedEvent publishes a dictionary invalidation event.
// When no keys are provided, subscribers are expected to clear their entire cache.
func PublishDataDictChangedEvent(publisher event.Publisher, keys ...string) {
	publisher.Publish(&DataDictChangedEvent{
		BaseEvent: event.NewBaseEvent(eventTypeDataDictChanged),

		Keys: keys,
	})
}

// CachedDataDictResolver adds caching and event-based invalidation around a DataDictLoader implementation.
type CachedDataDictResolver struct {
	loader     DataDictLoader
	cache      cache.Cache[map[string]string]
	keyBuilder cache.KeyBuilder
	logger     logPkg.Logger
	group      singleflight.Group // Prevents cache stampede
}

// NewCachedDataDictResolver constructs a caching resolver for dictionary lookups.
//
// loader: Mandatory dictionary loader implementation (must not be nil).
// store:  Backing cache store instance (must not be nil).
// bus:    Event subscriber used to listen for invalidation events (must not be nil).
//
// Panics if any of the required parameters (loader, store, bus) is nil.
func NewCachedDataDictResolver(
	loader DataDictLoader,
	store cache.Store,
	bus event.Subscriber,
) DataDictResolver {
	if loader == nil {
		panic("NewCachedDataDictResolver requires a non-nil DataDictLoader, but got nil")
	}

	if store == nil {
		panic("NewCachedDataDictResolver requires a non-nil cache.Store, but got nil")
	}

	if bus == nil {
		panic("NewCachedDataDictResolver requires a non-nil event.Subscriber, but got nil")
	}

	resolver := &CachedDataDictResolver{
		loader:     loader,
		cache:      cache.New[map[string]string](cache.Key("translate", "data_dict"), store),
		keyBuilder: cache.NewPrefixKeyBuilder("dict"),
		logger:     log.Named("translate:cached_data_dict_resolver"),
	}

	bus.Subscribe(eventTypeDataDictChanged, resolver.handleInvalidation)

	return resolver
}

// Resolve finds the dictionary display name for the provided key/code combination.
// Returns the translated name and an error if resolution fails.
// Returns empty string without error if the key or code is empty, or if the entry is not found.
func (r *CachedDataDictResolver) Resolve(ctx context.Context, key, code string) (string, error) {
	if key == constants.Empty || code == constants.Empty {
		return constants.Empty, nil
	}

	entries, err := r.getEntries(ctx, key)
	if err != nil {
		return constants.Empty, fmt.Errorf("failed to load dictionary '%s': %w", key, err)
	}

	name, ok := entries[code]
	if !ok {
		return constants.Empty, nil
	}

	return name, nil
}

func (r *CachedDataDictResolver) getEntries(ctx context.Context, key string) (map[string]string, error) {
	cacheKey := r.keyBuilder.Build(key)

	// First, try to get from cache
	if cached, found := r.cache.Get(ctx, cacheKey); found {
		return cached, nil
	}

	// Use singleflight to prevent cache stampede
	// Use dictionary key as the singleflight key to ensure concurrent requests for the same dictionary are merged
	result, err, _ := r.group.Do(key, func() (any, error) {
		// Double-check cache after acquiring the singleflight lock
		// Another goroutine might have populated the cache while we were waiting
		if cached, found := r.cache.Get(ctx, cacheKey); found {
			return cached, nil
		}

		// Load from underlying loader
		entries, err := r.loader.Load(ctx, key)
		if err != nil {
			return nil, err
		}

		if entries == nil {
			entries = make(map[string]string)
		}

		// Cache the result
		if err := r.cache.Set(ctx, cacheKey, entries); err != nil {
			return nil, fmt.Errorf("failed to cache dictionary '%s': %w", key, err)
		}

		return entries, nil
	})
	if err != nil {
		return nil, err
	}

	return result.(map[string]string), nil
}

func (r *CachedDataDictResolver) handleInvalidation(ctx context.Context, evt event.Event) {
	changeEvent, ok := evt.(*DataDictChangedEvent)
	if !ok {
		r.logger.Errorf("Received invalid event type: %T", evt)

		return
	}

	if len(changeEvent.Keys) == 0 {
		if err := r.cache.Clear(ctx); err != nil {
			r.logger.Errorf("Failed to clear dictionary cache: %v", err)
		} else {
			r.logger.Info("Cleared all dictionary cache entries")
		}

		return
	}

	for _, dictKey := range changeEvent.Keys {
		cacheKey := r.keyBuilder.Build(dictKey)
		if err := r.cache.Delete(ctx, cacheKey); err != nil {
			r.logger.Errorf("Failed to delete cache for dictionary '%s': %v", dictKey, err)
		} else {
			r.logger.Infof("Cleared cache for dictionary '%s'", dictKey)
		}
	}
}
