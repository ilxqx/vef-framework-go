package security

import (
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"

	"github.com/ilxqx/vef-framework-go/cache"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/internal/log"
	logPkg "github.com/ilxqx/vef-framework-go/log"
	"github.com/ilxqx/vef-framework-go/set"
)

const (
	// EventTypeRolePermissionsChanged is the event type for role permissions changes.
	// When this event is published, the entire role permissions cache will be cleared.
	eventTypeRolePermissionsChanged = "security.role_permissions.changed"
)

// RolePermissionsChangedEvent is published when role permissions are modified.
type RolePermissionsChangedEvent struct {
	event.BaseEvent

	Roles []string `json:"roles"` // Affected role names (empty means all roles)
}

// PublishRolePermissionChangedEvent publishes a role permission changed event via the provided publisher.
// If no roles are specified, subscribers should interpret the event as affecting all roles.
func PublishRolePermissionsChangedEvent(publisher event.Publisher, roles ...string) {
	publisher.Publish(&RolePermissionsChangedEvent{
		BaseEvent: event.NewBaseEvent(eventTypeRolePermissionsChanged),

		Roles: roles,
	})
}

// CachedRolePermissionsLoader is a decorator that adds caching to a RolePermissionsLoader.
// It uses the cache system and event bus for automatic cache invalidation.
// It also prevents cache stampede using singleflight to ensure concurrent requests
// for the same role only trigger one underlying load operation.
type CachedRolePermissionsLoader struct {
	loader     RolePermissionsLoader
	cache      cache.Cache[[]string]
	keyBuilder cache.KeyBuilder
	logger     logPkg.Logger
	group      singleflight.Group // Prevents cache stampede
}

// NewCachedRolePermissionsLoader creates a new cached role permissions loader.
// It automatically subscribes to role permission change events to invalidate cache.
//
// loader: The underlying RolePermissionsLoader to decorate.
// store: The cache store injected via DI.
// eventBus: The event bus for listening to cache invalidation events.
func NewCachedRolePermissionsLoader(
	loader RolePermissionsLoader,
	store cache.Store,
	eventBus event.Subscriber,
) RolePermissionsLoader {
	// Create a dedicated cache instance for role permissions
	permissionCache := cache.New[[]string](cache.Key("security", "role_permissions"), store)

	cached := &CachedRolePermissionsLoader{
		loader:     loader,
		cache:      permissionCache,
		keyBuilder: cache.NewPrefixKeyBuilder("role"),
		logger:     log.Named("security:cached_role_permissions_loader"),
	}

	// Subscribe to role permission change events
	eventBus.Subscribe(eventTypeRolePermissionsChanged, cached.handlePermissionChanged)

	return cached
}

// handlePermissionChanged handles role permission change events.
// It clears the cache for affected roles or the entire cache if no specific roles are provided.
func (c *CachedRolePermissionsLoader) handlePermissionChanged(ctx context.Context, evt event.Event) {
	changeEvent, ok := evt.(*RolePermissionsChangedEvent)
	if !ok {
		c.logger.Errorf("Received invalid event type: %T", evt)

		return
	}

	// If no specific roles provided, clear all permissions cache
	if len(changeEvent.Roles) == 0 {
		if err := c.cache.Clear(ctx); err != nil {
			c.logger.Errorf("Failed to clear all role permissions cache: %v", err)
		} else {
			c.logger.Info("Cleared all role permissions cache")
		}

		return
	}

	// Clear cache for specific roles
	for _, role := range changeEvent.Roles {
		cacheKey := c.keyBuilder.Build(role)
		if err := c.cache.Delete(ctx, cacheKey); err != nil {
			c.logger.Errorf("Failed to delete cache for role %s: %v", role, err)
		} else {
			c.logger.Infof("Cleared cache for role: %s", role)
		}
	}
}

// LoadPermissions loads permissions for a single role, using cache when available.
// Cache key format: "role:{roleName}"
// Uses singleflight to prevent cache stampede for concurrent requests to the same role.
func (c *CachedRolePermissionsLoader) LoadPermissions(ctx context.Context, role string) (set.Set[string], error) {
	cacheKey := c.keyBuilder.Build(role)

	// First, try to get from cache
	if permissionsSlice, found := c.cache.Get(ctx, cacheKey); found {
		return set.NewHashSetFromSlice(permissionsSlice), nil
	}

	// Use singleflight to prevent cache stampede
	// Use role name as the key to ensure concurrent requests for the same role are merged
	result, err, _ := c.group.Do(role, func() (any, error) {
		// Double-check cache after acquiring the singleflight lock
		// Another goroutine might have populated the cache while we were waiting
		if permissionsSlice, found := c.cache.Get(ctx, cacheKey); found {
			return set.NewHashSetFromSlice(permissionsSlice), nil
		}

		// Load from underlying loader
		permissions, err := c.loader.LoadPermissions(ctx, role)
		if err != nil {
			return nil, err
		}

		// Store in cache (no TTL - invalidation is event-driven)
		if err := c.cache.Set(ctx, cacheKey, permissions.Values()); err != nil {
			return nil, fmt.Errorf("failed to cache permissions for role %s: %w", role, err)
		}

		return permissions, nil
	})
	if err != nil {
		return nil, err
	}

	return result.(set.Set[string]), nil
}
