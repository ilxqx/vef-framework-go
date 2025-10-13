package security

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	eventPkg "github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/internal/event"
)

type CachedRolePermissionsLoaderTestSuite struct {
	suite.Suite

	ctx context.Context
	bus eventPkg.Bus
}

func (s *CachedRolePermissionsLoaderTestSuite) SetupSuite() {
	s.ctx = context.Background()

	s.bus = event.NewMemoryBus([]eventPkg.Middleware{})
	// Start the event bus
	err := s.bus.(interface{ Start() error }).Start()
	s.Require().NoError(err, "Failed to start event bus")
}

func (s *CachedRolePermissionsLoaderTestSuite) TestCachesResults() {
	mockLoader := new(MockRolePermissionsLoader)

	// Setup mock expectations - each role should be loaded only once
	adminPerms := map[string]DataScope{
		"test.read":  NewAllDataScope(),
		"test.write": NewAllDataScope(),
	}
	userPerms := map[string]DataScope{
		"test.read": NewSelfDataScope(""),
	}

	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms, nil).
		Once()
	mockLoader.On("LoadPermissions", mock.Anything, "user").
		Return(userPerms, nil).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.bus)

	// First call for admin - should call the underlying loader
	result, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().Equal(2, len(result))
	s.Require().Contains(result, "test.read")
	s.Require().Contains(result, "test.write")
	s.Require().NotContains(result, "test.delete")

	// Second call for admin - should use cache without calling underlying loader
	result2, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().Equal(2, len(result2))
	s.Require().Contains(result2, "test.read")
	s.Require().Contains(result2, "test.write")

	// First call for user - should call the underlying loader
	result3, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().Equal(1, len(result3))
	s.Require().Contains(result3, "test.read")
	s.Require().NotContains(result3, "test.write")

	// Second call for user - should use cache
	result4, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().Equal(1, len(result4))
	s.Require().Contains(result4, "test.read")

	// Verify mock was called exactly twice (once per role)
	mockLoader.AssertExpectations(s.T())
}

func (s *CachedRolePermissionsLoaderTestSuite) TestInvalidatesSpecificRoles() {
	mockLoader := new(MockRolePermissionsLoader)

	// First call expectations
	adminPerms1 := map[string]DataScope{
		"test.read":  NewAllDataScope(),
		"test.write": NewAllDataScope(),
	}
	userPerms := map[string]DataScope{
		"test.read": NewSelfDataScope(""),
	}

	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms1, nil).
		Once()
	mockLoader.On("LoadPermissions", mock.Anything, "user").
		Return(userPerms, nil).
		Once()

	// After invalidation, admin should be reloaded with new permissions
	adminPerms2 := map[string]DataScope{
		"test.read":   NewAllDataScope(),
		"test.write":  NewAllDataScope(),
		"test.delete": NewAllDataScope(),
	}
	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms2, nil).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.bus)

	// First call - loads from underlying loader
	result, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().Contains(result, "test.read")
	s.Require().Contains(result, "test.write")
	s.Require().NotContains(result, "test.delete")

	// Load user too
	resultUser, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().Contains(resultUser, "test.read")

	// Publish event to invalidate specific role (admin only)
	PublishRolePermissionsChangedEvent(s.bus, "admin")

	// Wait for event processing
	time.Sleep(10 * time.Millisecond)

	// Second call for admin - should reload from underlying loader with new permissions
	result2, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().Contains(result2, "test.read")
	s.Require().Contains(result2, "test.write")
	s.Require().Contains(result2, "test.delete")

	// User should still come from cache (not reloaded)
	resultUser2, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().Contains(resultUser2, "test.read")

	mockLoader.AssertExpectations(s.T())
}

func (s *CachedRolePermissionsLoaderTestSuite) TestInvalidatesAllRoles() {
	mockLoader := new(MockRolePermissionsLoader)

	// First call expectations
	adminPerms1 := map[string]DataScope{
		"test.read":  NewAllDataScope(),
		"test.write": NewAllDataScope(),
	}
	userPerms1 := map[string]DataScope{
		"test.read": NewSelfDataScope(""),
	}

	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms1, nil).
		Once()
	mockLoader.On("LoadPermissions", mock.Anything, "user").
		Return(userPerms1, nil).
		Once()

	// After invalidation, both roles should be reloaded with new permissions
	adminPerms2 := map[string]DataScope{
		"test.read":   NewAllDataScope(),
		"test.write":  NewAllDataScope(),
		"test.delete": NewAllDataScope(),
	}
	userPerms2 := map[string]DataScope{
		"test.read":   NewSelfDataScope(""),
		"test.update": NewSelfDataScope(""),
	}

	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms2, nil).
		Once()
	mockLoader.On("LoadPermissions", mock.Anything, "user").
		Return(userPerms2, nil).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.bus)

	// First call - loads from underlying loader
	result, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().Contains(result, "test.read")
	s.Require().Contains(result, "test.write")
	s.Require().NotContains(result, "test.delete")

	resultUser, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().Contains(resultUser, "test.read")
	s.Require().NotContains(resultUser, "test.update")

	// Publish event to invalidate all roles (empty roles slice)
	PublishRolePermissionsChangedEvent(s.bus)

	// Wait for event processing
	time.Sleep(10 * time.Millisecond)

	// Second call - should reload all roles from underlying loader
	result2, err := cachedLoader.LoadPermissions(s.ctx, "admin")
	s.Require().NoError(err)
	s.Require().Contains(result2, "test.read")
	s.Require().Contains(result2, "test.write")
	s.Require().Contains(result2, "test.delete")

	resultUser2, err := cachedLoader.LoadPermissions(s.ctx, "user")
	s.Require().NoError(err)
	s.Require().Contains(resultUser2, "test.read")
	s.Require().Contains(resultUser2, "test.update")

	mockLoader.AssertExpectations(s.T())
}

func (s *CachedRolePermissionsLoaderTestSuite) TestEmptyRole() {
	mockLoader := new(MockRolePermissionsLoader)

	// Mock for empty role should return empty map
	emptyMap := make(map[string]DataScope)
	mockLoader.On("LoadPermissions", mock.Anything, "").
		Return(emptyMap, nil).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.bus)

	result, err := cachedLoader.LoadPermissions(s.ctx, "")
	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Require().Empty(result)

	mockLoader.AssertExpectations(s.T())
}

func (s *CachedRolePermissionsLoaderTestSuite) TestLoaderError() {
	mockLoader := new(MockRolePermissionsLoader)
	expectedError := context.DeadlineExceeded
	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(nil, expectedError).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.bus)

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
	adminPerms := map[string]DataScope{
		"test.read":  NewAllDataScope(),
		"test.write": NewAllDataScope(),
	}

	// The mock should be called only once, even though we make multiple concurrent requests
	mockLoader.On("LoadPermissions", mock.Anything, "admin").
		Return(adminPerms, nil).
		Once()

	cachedLoader := NewCachedRolePermissionsLoader(mockLoader, s.bus)

	// Make multiple concurrent requests for the same role
	const numRequests = 10

	var wg sync.WaitGroup

	results := make([]map[string]DataScope, numRequests)
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
		s.Require().Contains(results[i], "test.read")
		s.Require().Contains(results[i], "test.write")
	}

	// The mock should have been called only once, proving that singleflight merged all requests
	mockLoader.AssertExpectations(s.T())
}

type MockRolePermissionsLoader struct {
	mock.Mock
}

func (m *MockRolePermissionsLoader) LoadPermissions(ctx context.Context, role string) (map[string]DataScope, error) {
	args := m.Called(ctx, role)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]DataScope), args.Error(1)
}

func TestCachedRolePermissionsLoaderSuite(t *testing.T) {
	suite.Run(t, new(CachedRolePermissionsLoaderTestSuite))
}
