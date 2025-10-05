package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"

	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/testhelpers"
)

// RedisTestSuite is the test suite for redis package.
type RedisTestSuite struct {
	suite.Suite

	ctx            context.Context
	redisContainer *testhelpers.RedisContainer
}

// SetupSuite runs before all tests in the suite.
func (suite *RedisTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Start Redis container using testhelpers
	suite.redisContainer = testhelpers.NewRedisContainer(suite.ctx, &suite.Suite)
}

// TearDownSuite runs after all tests in the suite.
func (suite *RedisTestSuite) TearDownSuite() {
	if suite.redisContainer != nil {
		suite.redisContainer.Terminate(suite.ctx, &suite.Suite)
	}
}

// TestNewClient tests Redis client creation with default configuration.
func (suite *RedisTestSuite) TestNewClient() {
	config := &config.RedisConfig{
		Host:     "127.0.0.1",
		Port:     6379,
		Database: 0,
	}

	client := NewClient("test-app", config)
	suite.Require().NotNil(client)

	// Verify client options
	options := client.Options()
	suite.Equal("test-app", options.ClientName)
	suite.Equal(constants.VEFName, options.IdentitySuffix)
	suite.Equal(3, options.Protocol)
	suite.Equal("127.0.0.1:6379", options.Addr)
	suite.Equal(0, options.DB)
	suite.True(options.ContextTimeoutEnabled)

	// Verify pool configuration
	suite.Greater(options.PoolSize, 0)
	suite.Greater(options.PoolTimeout, time.Duration(0))
	suite.Greater(options.MaxRetries, 0)
	suite.Greater(options.MinIdleConns, 0)
	suite.Greater(options.ConnMaxLifetime, time.Duration(0))
	suite.Greater(options.ConnMaxIdleTime, time.Duration(0))

	suite.T().Logf("Redis client created with pool size: %d", options.PoolSize)
}

// TestNewClientWithCustomConfig tests Redis client with custom configuration.
func (suite *RedisTestSuite) TestNewClientWithCustomConfig() {
	config := &config.RedisConfig{
		Host:     "custom-host",
		Port:     6380,
		User:     "testuser",
		Password: "testpass",
		Database: 5,
		Network:  "tcp",
	}

	client := NewClient("custom-app", config)
	suite.Require().NotNil(client)

	options := client.Options()
	suite.Equal("custom-app", options.ClientName)
	suite.Equal("custom-host:6380", options.Addr)
	suite.Equal("testuser", options.Username)
	suite.Equal("testpass", options.Password)
	suite.Equal(5, options.DB)
	suite.Equal("tcp", options.Network)
}

// TestRedisConnection tests actual Redis connection using testcontainers.
func (suite *RedisTestSuite) TestRedisConnection() {
	// Use the pre-configured Redis container
	config := suite.redisContainer.RdsConfig

	client := NewClient("test-connection", config)
	suite.Require().NotNil(client)

	suite.T().Logf("Redis connection config: %+v", config)

	// Test connection
	err := client.Ping(suite.ctx).Err()
	suite.Require().NoError(err)

	// Test basic operations
	suite.testBasicRedisOperations(client)

	// Clean up
	suite.Require().NoError(client.Close())
}

// TestHealthCheck tests the health check functionality.
func (suite *RedisTestSuite) TestHealthCheck() {
	// Use the pre-configured Redis container
	config := suite.redisContainer.RdsConfig

	client := NewClient("test-health", config)
	suite.Require().NotNil(client)

	// Test health check
	err := HealthCheck(suite.ctx, client)
	suite.Require().NoError(err)

	// Clean up
	suite.Require().NoError(client.Close())
}

// TestHealthCheckFailure tests health check with invalid configuration.
func (suite *RedisTestSuite) TestHealthCheckFailure() {
	config := &config.RedisConfig{
		Host:     "invalid-host",
		Port:     9999,
		Database: 0,
	}

	client := NewClient("test-health-fail", config)
	suite.Require().NotNil(client)

	// Test health check failure
	err := HealthCheck(suite.ctx, client)
	suite.Error(err)

	// Clean up
	suite.Require().NoError(client.Close())
}

// TestBuildRedisAddr tests the Redis address building function.
func (suite *RedisTestSuite) TestBuildRedisAddr() {
	tests := []struct {
		name     string
		config   *config.RedisConfig
		expected string
	}{
		{
			name: "default host and port",
			config: &config.RedisConfig{
				Host: "",
				Port: 0,
			},
			expected: "127.0.0.1:6379",
		},
		{
			name: "custom host and port",
			config: &config.RedisConfig{
				Host: "redis.example.com",
				Port: 6380,
			},
			expected: "redis.example.com:6380",
		},
		{
			name: "custom host with default port",
			config: &config.RedisConfig{
				Host: "localhost",
				Port: 0,
			},
			expected: "localhost:6379",
		},
		{
			name: "default host with custom port",
			config: &config.RedisConfig{
				Host: "",
				Port: 6380,
			},
			expected: "127.0.0.1:6380",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			addr := buildRedisAddr(tt.config)
			suite.Equal(tt.expected, addr)
		})
	}
}

// TestGetPoolSize tests the pool size calculation.
func (suite *RedisTestSuite) TestGetPoolSize() {
	poolSize := getPoolSize()

	// Pool size should be at least 4 and at most 100
	suite.GreaterOrEqual(poolSize, 4)
	suite.LessOrEqual(poolSize, 100)

	suite.T().Logf("Calculated pool size: %d", poolSize)
}

// TestGetConnectionConfig tests the connection configuration.
func (suite *RedisTestSuite) TestGetConnectionConfig() {
	poolSize := 10
	poolTimeout, idleTimeout, maxRetries := getConnectionConfig(poolSize)

	// Verify reasonable values
	suite.GreaterOrEqual(poolTimeout, 1*time.Second)
	suite.LessOrEqual(poolTimeout, 5*time.Second)
	suite.Equal(5*time.Minute, idleTimeout)
	suite.Equal(3, maxRetries)

	suite.T().Logf("Connection config - Pool timeout: %v, Idle timeout: %v, Max retries: %d",
		poolTimeout, idleTimeout, maxRetries)
}

// testBasicRedisOperations performs basic Redis operations to verify functionality.
func (suite *RedisTestSuite) testBasicRedisOperations(client *redis.Client) {
	suite.T().Log("Testing basic Redis operations")

	// Test SET and GET
	err := client.Set(suite.ctx, "test_key", "test_value", 0).Err()
	suite.Require().NoError(err)

	val, err := client.Get(suite.ctx, "test_key").Result()
	suite.Require().NoError(err)
	suite.Equal("test_value", val)

	// Test DEL
	err = client.Del(suite.ctx, "test_key").Err()
	suite.Require().NoError(err)

	// Verify key is deleted
	_, err = client.Get(suite.ctx, "test_key").Result()
	suite.Error(err) // Should be redis.Nil error

	// Test HSET and HGET
	err = client.HSet(suite.ctx, "test_hash", "field1", "value1").Err()
	suite.Require().NoError(err)

	hashVal, err := client.HGet(suite.ctx, "test_hash", "field1").Result()
	suite.Require().NoError(err)
	suite.Equal("value1", hashVal)

	// Clean up
	err = client.Del(suite.ctx, "test_hash").Err()
	suite.Require().NoError(err)

	suite.T().Log("Basic Redis operations completed successfully")
}

// TestRedisSuite runs the test suite.
func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}
