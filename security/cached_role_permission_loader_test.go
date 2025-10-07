package security

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	cachePkg "github.com/ilxqx/vef-framework-go/cache"
	"github.com/ilxqx/vef-framework-go/config"
	eventPkg "github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/internal/cache"
	"github.com/ilxqx/vef-framework-go/internal/event"
	"github.com/ilxqx/vef-framework-go/set"
)

type CachedRolePermissionsLoaderTestSuite struct {
	suite.Suite

	ctx   context.Context
	store cachePkg.Store
	bus   eventPkg.Bus
}

func (s *CachedRolePermissionsLoaderTestSuite) SetupSuite() {
	s.ctx = context.Background()

	var err error

	s.store, err = cache.NewBadgerStore(&config.LocalCacheConfig{
		InMemory: true,
	})

	s.Require().NoError(err, "Failed to create badger store")

	s.bus = event.NewMemoryBus(s.ctx, []eventPkg.Middleware{})
	// Start the event bus
	err = s.bus.(interface{ Start() error }).Start()
	s.Require().NoError(err, "Failed to start event bus")
}

func (s *CachedRolePermissionsLoaderTestSuite) TearDownTest() {
	// Clear cache between tests to avoid interference
	_ = s.store.Clear(s.ctx, "")
}

func (s *CachedRolePermissionsLoaderTestSuite) TestCachesResults() {
	mockLoader := new(MockRolePermissionsLoader)

	// Setup mock expectations - each role should be loaded only once
	adminPerms := set.NewHashSetFromSlice([]string{"test.read", "test.write"})
	userPerms := set.NewHashSetFromSlice([]string{"test.read"})

	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms, nil).
		Once()
	mockLoader.On("LoadPermissions", mock.Anything, "user").
		Return(userPerms, nil).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.store, s.bus)

	// First call for admin - should call the underlying loader
	result, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().True(result.Contains("test.read"))
	s.Require().True(result.Contains("test.write"))
	s.Require().False(result.Contains("test.delete"))

	// Second call for admin - should use cache without calling underlying loader
	result2, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().True(result2.Contains("test.read"))
	s.Require().True(result2.Contains("test.write"))

	// First call for user - should call the underlying loader
	result3, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().True(result3.Contains("test.read"))
	s.Require().False(result3.Contains("test.write"))

	// Second call for user - should use cache
	result4, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().True(result4.Contains("test.read"))

	// Verify mock was called exactly twice (once per role)
	mockLoader.AssertExpectations(s.T())
}

func (s *CachedRolePermissionsLoaderTestSuite) TestInvalidatesSpecificRoles() {
	mockLoader := new(MockRolePermissionsLoader)

	// First call expectations
	adminPerms1 := set.NewHashSetFromSlice([]string{"test.read", "test.write"})
	userPerms := set.NewHashSetFromSlice([]string{"test.read"})

	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms1, nil).
		Once()
	mockLoader.On("LoadPermissions", mock.Anything, "user").
		Return(userPerms, nil).
		Once()

	// After invalidation, admin should be reloaded with new permissions
	adminPerms2 := set.NewHashSetFromSlice([]string{"test.read", "test.write", "test.delete"})
	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms2, nil).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.store, s.bus)

	// First call - loads from underlying loader
	result, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().True(result.Contains("test.read"))
	s.Require().True(result.Contains("test.write"))
	s.Require().False(result.Contains("test.delete"))

	// Load user too
	resultUser, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().True(resultUser.Contains("test.read"))

	// Publish event to invalidate specific role (admin only)
	PublishRolePermissionsChangedEvent(s.bus, "admin")

	// Wait for event processing
	time.Sleep(10 * time.Millisecond)

	// Second call for admin - should reload from underlying loader with new permissions
	result2, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().True(result2.Contains("test.read"))
	s.Require().True(result2.Contains("test.write"))
	s.Require().True(result2.Contains("test.delete"))

	// User should still come from cache (not reloaded)
	resultUser2, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().True(resultUser2.Contains("test.read"))

	mockLoader.AssertExpectations(s.T())
}

func (s *CachedRolePermissionsLoaderTestSuite) TestInvalidatesAllRoles() {
	mockLoader := new(MockRolePermissionsLoader)

	// First call expectations
	adminPerms1 := set.NewHashSetFromSlice([]string{"test.read", "test.write"})
	userPerms1 := set.NewHashSetFromSlice([]string{"test.read"})

	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms1, nil).
		Once()
	mockLoader.On("LoadPermissions", mock.Anything, "user").
		Return(userPerms1, nil).
		Once()

	// After invalidation, both roles should be reloaded with new permissions
	adminPerms2 := set.NewHashSetFromSlice([]string{"test.read", "test.write", "test.delete"})
	userPerms2 := set.NewHashSetFromSlice([]string{"test.read", "test.update"})

	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms2, nil).
		Once()
	mockLoader.On("LoadPermissions", mock.Anything, "user").
		Return(userPerms2, nil).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.store, s.bus)

	// First call - loads from underlying loader
	result, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().True(result.Contains("test.read"))
	s.Require().True(result.Contains("test.write"))
	s.Require().False(result.Contains("test.delete"))

	resultUser, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().True(resultUser.Contains("test.read"))
	s.Require().False(resultUser.Contains("test.update"))

	// Publish event to invalidate all roles (empty roles slice)
	PublishRolePermissionsChangedEvent(s.bus)

	// Wait for event processing
	time.Sleep(10 * time.Millisecond)

	// Second call - should reload all roles from underlying loader
	result2, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().True(result2.Contains("test.read"))
	s.Require().True(result2.Contains("test.write"))
	s.Require().True(result2.Contains("test.delete"))

	resultUser2, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().True(resultUser2.Contains("test.read"))
	s.Require().True(resultUser2.Contains("test.update"))

	mockLoader.AssertExpectations(s.T())
}

func (s *CachedRolePermissionsLoaderTestSuite) TestEmptyRole() {
	mockLoader := new(MockRolePermissionsLoader)

	// Mock for empty role should return empty set
	emptySet := set.NewHashSet[string]()
	mockLoader.On("LoadPermissions", mock.Anything, "").
		Return(emptySet, nil).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.store, s.bus)

	result, err := cachedLoader.LoadPermissions(s.ctx, "")
	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Require().True(result.IsEmpty())

	mockLoader.AssertExpectations(s.T())
}

func (s *CachedRolePermissionsLoaderTestSuite) TestLoaderError() {
	mockLoader := new(MockRolePermissionsLoader)
	expectedError := context.DeadlineExceeded
	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(nil, expectedError).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.store, s.bus)

	result, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().Error(err)
	s.Require().Equal(expectedError, err)
	s.Require().Nil(result)

	mockLoader.AssertExpectations(s.T())
}

// TestSingleflightMergesConcurrentRequests verifies that concurrent requests for the same role
// are merged by singleflight and only trigger one underlying load operation.
func (s *CachedRolePermissionsLoaderTestSuite) TestSingleflightMergesConcurrentRequests() {
	mockLoader := new(MockRolePermissionsLoader)

	// Setup mock to return permissions for the role
	adminPerms := set.NewHashSetFromSlice([]string{"test.read", "test.write"})

	// The mock should be called only once, even though we make multiple concurrent requests
	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms, nil).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.store, s.bus)

	// Make multiple concurrent requests for the same role
	const numRequests = 10

	var wg sync.WaitGroup

	results := make([]set.Set[string], numRequests)
	errors := make([]error, numRequests)

	for i := range numRequests {
		wg.Go(func() {
			results[i], errors[i] = cachedLoader.LoadPermissions(s.ctx, "admin")
		})
	}

	wg.Wait()

	// All requests should succeed
	for i := range numRequests {
		s.Require().NoError(errors[i], "Request %d should not error", i)
		s.Require().NotNil(results[i], "Request %d should return a result", i)
		s.Require().True(results[i].Contains("test.read"))
		s.Require().True(results[i].Contains("test.write"))
	}

	// The mock should have been called only once, proving that singleflight merged all requests
	mockLoader.AssertExpectations(s.T())
}

type MockRolePermissionsLoader struct {
	mock.Mock
}

func (m *MockRolePermissionsLoader) LoadPermissions(ctx context.Context, role string) (set.Set[string], error) {
	args := m.Called(ctx, role)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(set.Set[string]), args.Error(1)
}

func TestCachedRolePermissionsLoaderSuite(t *testing.T) {
	suite.Run(t, new(CachedRolePermissionsLoaderTestSuite))
}
