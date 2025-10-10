package mold

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	eventPkg "github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/internal/event"
)

type CachedDataDictResolverTestSuite struct {
	suite.Suite

	ctx context.Context
	bus eventPkg.Bus
}

func (s *CachedDataDictResolverTestSuite) SetupSuite() {
	s.ctx = context.Background()

	s.bus = event.NewMemoryBus(s.ctx, []eventPkg.Middleware{})

	err := s.bus.Start()
	s.Require().NoError(err)
}

func (s *CachedDataDictResolverTestSuite) TearDownSuite() {
	_ = s.bus.Shutdown(context.Background())
}

func (s *CachedDataDictResolverTestSuite) newResolver(loader DataDictLoader) DataDictResolver {
	return NewCachedDataDictResolver(loader, s.bus)
}

func (s *CachedDataDictResolverTestSuite) TestCachesEntries() {
	loader := new(MockDataDictLoader)
	loader.On("Load", mock.Anything, "status").Return(map[string]string{
		"draft":     "草稿",
		"published": "已发布",
	}, nil).Once()

	resolver := s.newResolver(loader)

	result, err := resolver.Resolve(s.ctx, "status", "published")
	s.NoError(err)
	s.Equal("已发布", result)

	result2, err := resolver.Resolve(s.ctx, "status", "draft")
	s.NoError(err)
	s.Equal("草稿", result2)

	loader.AssertExpectations(s.T())
}

func (s *CachedDataDictResolverTestSuite) TestInvalidatesSpecificKeys() {
	loader := new(MockDataDictLoader)
	loader.On("Load", mock.Anything, "status").Return(map[string]string{
		"draft": "草稿",
	}, nil).Once()
	loader.On("Load", mock.Anything, "status").Return(map[string]string{
		"draft":    "草稿",
		"archived": "已归档",
	}, nil).Once()

	resolver := s.newResolver(loader)

	first, err := resolver.Resolve(s.ctx, "status", "draft")
	s.NoError(err)
	s.Equal("草稿", first)

	PublishDataDictChangedEvent(s.bus, "status")
	time.Sleep(10 * time.Millisecond)

	second, err := resolver.Resolve(s.ctx, "status", "archived")
	s.NoError(err)
	s.Equal("已归档", second)

	loader.AssertExpectations(s.T())
}

func (s *CachedDataDictResolverTestSuite) TestInvalidatesAllKeys() {
	loader := new(MockDataDictLoader)
	loader.On("Load", mock.Anything, "status").Return(map[string]string{
		"draft": "草稿",
	}, nil).Once()
	loader.On("Load", mock.Anything, "category").Return(map[string]string{
		"news": "新闻",
	}, nil).Once()
	loader.On("Load", mock.Anything, "status").Return(map[string]string{
		"draft":     "草稿",
		"published": "已发布",
	}, nil).Once()

	resolver := s.newResolver(loader)

	firstStatus, err := resolver.Resolve(s.ctx, "status", "draft")
	s.NoError(err)
	s.Equal("草稿", firstStatus)

	firstCategory, err := resolver.Resolve(s.ctx, "category", "news")
	s.NoError(err)
	s.Equal("新闻", firstCategory)

	PublishDataDictChangedEvent(s.bus)
	time.Sleep(10 * time.Millisecond)

	updatedStatus, err := resolver.Resolve(s.ctx, "status", "published")
	s.NoError(err)
	s.Equal("已发布", updatedStatus)

	loader.AssertExpectations(s.T())
}

func (s *CachedDataDictResolverTestSuite) TestLoaderError() {
	loader := new(MockDataDictLoader)
	expectedErr := context.DeadlineExceeded
	loader.On("Load", mock.Anything, "status").Return(map[string]string(nil), expectedErr).Once()

	resolver := s.newResolver(loader)

	result, err := resolver.Resolve(s.ctx, "status", "draft")
	s.Error(err)
	s.True(errors.Is(err, expectedErr), "Error should wrap the original error")
	s.Contains(err.Error(), "failed to load dictionary 'status'")
	s.Equal("", result)

	loader.AssertExpectations(s.T())
}

func (s *CachedDataDictResolverTestSuite) TestEmptyKeyOrCode() {
	loader := new(MockDataDictLoader)

	resolver := s.newResolver(loader)

	result1, err1 := resolver.Resolve(s.ctx, "", "code")
	s.NoError(err1)
	s.Equal("", result1)

	result2, err2 := resolver.Resolve(s.ctx, "key", "")
	s.NoError(err2)
	s.Equal("", result2)

	loader.AssertExpectations(s.T())
}

func (s *CachedDataDictResolverTestSuite) TestCodeNotFound() {
	loader := new(MockDataDictLoader)
	loader.On("Load", mock.Anything, "status").Return(map[string]string{
		"draft":     "草稿",
		"published": "已发布",
	}, nil).Once()

	resolver := s.newResolver(loader)

	result, err := resolver.Resolve(s.ctx, "status", "archived")
	s.NoError(err)
	s.Equal("", result)

	loader.AssertExpectations(s.T())
}

func (s *CachedDataDictResolverTestSuite) TestPanicsWhenLoaderIsNil() {
	s.Panics(func() {
		NewCachedDataDictResolver(nil, s.bus)
	}, "Expected panic when loader is nil")
}

func (s *CachedDataDictResolverTestSuite) TestPanicsWhenBusIsNil() {
	loader := new(MockDataDictLoader)
	s.Panics(func() {
		NewCachedDataDictResolver(loader, nil)
	}, "Expected panic when bus is nil")
}

func (s *CachedDataDictResolverTestSuite) TestNilCacheCreatesDefault() {
	loader := new(MockDataDictLoader)
	loader.On("Load", mock.Anything, "status").Return(map[string]string{
		"draft": "草稿",
	}, nil).Once()

	resolver := NewCachedDataDictResolver(loader, s.bus)

	result, err := resolver.Resolve(s.ctx, "status", "draft")
	s.NoError(err)
	s.Equal("草稿", result)

	loader.AssertExpectations(s.T())
}

// TestSingleflightMergesConcurrentRequests verifies that concurrent requests for the same dictionary key
// are merged by singleflight and only trigger one underlying load operation.
func (s *CachedDataDictResolverTestSuite) TestSingleflightMergesConcurrentRequests() {
	loader := new(MockDataDictLoader)

	// Setup mock to return dictionary data for the key
	dictData := map[string]string{
		"draft":     "草稿",
		"published": "已发布",
		"archived":  "已归档",
	}

	// The mock should be called only once, even though we make multiple concurrent requests
	loader.On("Load", mock.Anything, "status").
		Return(dictData, nil).
		Once()

	resolver := s.newResolver(loader)

	// Make multiple concurrent requests for the same dictionary key
	const numRequests = 10

	var wg sync.WaitGroup

	results := make([]string, numRequests)
	errors := make([]error, numRequests)

	for i := range numRequests {
		wg.Go(func() {
			// Different codes but same dictionary key
			codes := []string{"draft", "published", "archived"}
			code := codes[i%len(codes)]
			results[i], errors[i] = resolver.Resolve(s.ctx, "status", code)
		})
	}

	wg.Wait()

	// All requests should succeed
	for i := range numRequests {
		s.NoError(errors[i], "Request %d should not error", i)
		s.NotEmpty(results[i], "Request %d should return a result", i)

		// Verify the result matches expected value
		codes := []string{"draft", "published", "archived"}
		expectedCode := codes[i%len(codes)]
		expectedValue := dictData[expectedCode]
		s.Equal(expectedValue, results[i], "Request %d should return correct value", i)
	}

	// The mock should have been called only once, proving that singleflight merged all requests
	loader.AssertExpectations(s.T())
}

type MockDataDictLoader struct {
	mock.Mock
}

func (m *MockDataDictLoader) Load(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]string), args.Error(1)
}

func TestCachedDataDictResolverSuite(t *testing.T) {
	suite.Run(t, new(CachedDataDictResolverTestSuite))
}
