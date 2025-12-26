package mcp_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/internal/app"
	"github.com/ilxqx/vef-framework-go/internal/apptest"
	isecurity "github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/password"
	"github.com/ilxqx/vef-framework-go/security"
)

// MockUserLoader is a mock implementation of security.UserLoader for testing.
type MockUserLoader struct {
	mock.Mock
}

func (m *MockUserLoader) LoadByUsername(ctx context.Context, username string) (*security.Principal, string, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}

	return args.Get(0).(*security.Principal), args.String(1), args.Error(2)
}

func (m *MockUserLoader) LoadById(ctx context.Context, id string) (*security.Principal, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*security.Principal), args.Error(1)
}

// McpTestSuite tests the MCP endpoint functionality with authentication.
type McpTestSuite struct {
	suite.Suite

	ctx        context.Context
	app        *app.App
	stop       func()
	userLoader *MockUserLoader
	jwtSecret  string
	testUser   *security.Principal
}

// SetupSuite runs once before all tests in the suite.
func (suite *McpTestSuite) SetupSuite() {
	suite.T().Log("Setting up McpTestSuite - initializing test app with MCP")

	suite.ctx = context.Background()
	suite.jwtSecret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	suite.testUser = security.NewUser("user001", "Test User", "admin", "user")
	suite.testUser.Details = map[string]any{
		"email": "test@example.com",
	}

	suite.userLoader = new(MockUserLoader)

	suite.setupTestApp()

	suite.T().Log("McpTestSuite setup complete - test app ready")
}

// TearDownSuite runs once after all tests in the suite.
func (suite *McpTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down McpTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("McpTestSuite teardown complete")
}

// SetupTest runs before each test.
func (suite *McpTestSuite) SetupTest() {
	suite.userLoader.Calls = nil
}

func (suite *McpTestSuite) setupTestApp() {
	hashedPassword, err := password.NewBcryptEncoder().Encode("password123")
	suite.Require().NoError(err)

	suite.app, suite.stop = apptest.NewTestApp(
		suite.T(),
		fx.Supply(
			fx.Annotate(
				suite.userLoader,
				fx.As(new(security.UserLoader)),
			),
		),
		fx.Replace(
			&config.DatasourceConfig{
				Type: "sqlite",
			},
			&config.McpConfig{
				Enabled:     true,
				RequireAuth: true,
			},
			&config.SecurityConfig{
				TokenExpires: 24 * time.Hour,
			},
			&security.JwtConfig{
				Secret:   suite.jwtSecret,
				Audience: "test-app",
			},
		),
		fx.Invoke(func() {
			suite.userLoader.On("LoadByUsername", mock.Anything, "testuser").
				Return(suite.testUser, hashedPassword, nil).
				Maybe()

			suite.userLoader.On("LoadById", mock.Anything, "user001").
				Return(suite.testUser, nil).
				Maybe()
		}),
	)
}

// Helper methods

func (suite *McpTestSuite) makeMcpRequest(body string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "MCP request should not fail")

	return resp
}

func (suite *McpTestSuite) makeMcpRequestWithToken(body string, token string) *http.Response {
	req := httptest.NewRequest(fiber.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set(fiber.HeaderAuthorization, constants.AuthSchemeBearer+" "+token)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err, "MCP request should not fail")

	return resp
}

func (suite *McpTestSuite) getAccessToken() string {
	hashedPassword, _ := password.NewBcryptEncoder().Encode("password123")

	suite.userLoader.On("LoadByUsername", mock.Anything, "testuser").
		Return(suite.testUser, hashedPassword, nil).
		Maybe()

	// Login to get access token using proper API request format
	loginRequest := api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"type":        isecurity.AuthTypePassword,
			"principal":   "testuser",
			"credentials": "password123",
		},
	}

	jsonBody, err := encoding.ToJson(loginRequest)
	suite.Require().NoError(err, "Should encode login request to JSON")

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req, 30*time.Second)
	suite.Require().NoError(err)
	suite.Require().Equal(200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	// Extract accessToken from response
	bodyStr := string(body)
	startIdx := strings.Index(bodyStr, `"accessToken":"`) + len(`"accessToken":"`)
	if startIdx < len(`"accessToken":"`) {
		suite.T().Logf("Login response: %s", bodyStr)
		suite.FailNow("Failed to get accessToken from login response")
	}
	endIdx := strings.Index(bodyStr[startIdx:], `"`) + startIdx

	return bodyStr[startIdx:endIdx]
}

func (suite *McpTestSuite) readBody(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	suite.Require().NoError(err, "Should read response body")

	return string(body)
}

// Test Cases

// TestMcpEndpointRequiresAuthentication tests that MCP endpoint requires authentication.
// It verifies that requests without valid tokens are rejected with 401 Unauthorized.
func (suite *McpTestSuite) TestMcpEndpointRequiresAuthentication() {
	suite.T().Log("Testing MCP endpoint authentication requirement")

	// Test 1: Request without any token
	suite.Run("RequestWithoutToken", func() {
		body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`

		resp := suite.makeMcpRequest(body)

		suite.Equal(401, resp.StatusCode, "Should return 401 Unauthorized without token")
		suite.T().Log("Request without token correctly rejected with 401")
	})

	// Test 2: Request with invalid token
	suite.Run("RequestWithInvalidToken", func() {
		body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`

		resp := suite.makeMcpRequestWithToken(body, "invalid.token.here")

		suite.Equal(401, resp.StatusCode, "Should return 401 Unauthorized with invalid token")
		suite.T().Log("Request with invalid token correctly rejected with 401")
	})

	// Test 3: Request with empty token
	suite.Run("RequestWithEmptyToken", func() {
		body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`

		resp := suite.makeMcpRequestWithToken(body, "")

		suite.Equal(401, resp.StatusCode, "Should return 401 Unauthorized with empty token")
		suite.T().Log("Request with empty token correctly rejected with 401")
	})
}

// TestMcpEndpointWithValidToken tests MCP endpoint accepts valid authentication.
// It verifies that authenticated requests are processed successfully.
func (suite *McpTestSuite) TestMcpEndpointWithValidToken() {
	suite.T().Log("Testing MCP endpoint with valid authentication")

	// Test 1: Initialize with valid JWT token
	suite.Run("InitializeWithValidToken", func() {
		token := suite.getAccessToken()
		suite.NotEmpty(token, "Should get valid access token")
		suite.T().Logf("Got access token: %s...", token[:20])

		body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`

		resp := suite.makeMcpRequestWithToken(body, token)

		// Step 1: Should not return 401 when token is valid
		suite.NotEqual(401, resp.StatusCode, "Should not return 401 with valid token")
		suite.Equal(200, resp.StatusCode, "Should return 200 OK with valid token")

		// Step 2: Validate response body
		responseBody := suite.readBody(resp)
		suite.T().Logf("MCP Response: %s", responseBody)

		// Step 3: Check for valid JSON-RPC response
		suite.Contains(responseBody, "jsonrpc", "Response should contain jsonrpc field")
		suite.Contains(responseBody, "result", "Response should contain result field")
		suite.Contains(responseBody, "serverInfo", "Response should contain serverInfo")
	})
}

// TestMcpEndpointMethods tests various MCP methods with authentication.
func (suite *McpTestSuite) TestMcpEndpointMethods() {
	suite.T().Log("Testing MCP endpoint methods")

	token := suite.getAccessToken()
	suite.NotEmpty(token, "Should get valid access token")

	suite.Run("ListTools", func() {
		// First initialize
		initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`
		initResp := suite.makeMcpRequestWithToken(initBody, token)
		suite.NotEqual(401, initResp.StatusCode, "Initialize should not return 401")

		// Then list tools
		body := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`

		resp := suite.makeMcpRequestWithToken(body, token)

		suite.NotEqual(401, resp.StatusCode, "ListTools should not return 401 with valid token")

		responseBody := suite.readBody(resp)
		suite.T().Logf("ListTools Response: %s", responseBody)
	})

	suite.Run("ListResources", func() {
		body := `{"jsonrpc":"2.0","id":3,"method":"resources/list","params":{}}`

		resp := suite.makeMcpRequestWithToken(body, token)

		suite.NotEqual(401, resp.StatusCode, "ListResources should not return 401 with valid token")
	})

	suite.Run("ListPrompts", func() {
		body := `{"jsonrpc":"2.0","id":4,"method":"prompts/list","params":{}}`

		resp := suite.makeMcpRequestWithToken(body, token)

		suite.NotEqual(401, resp.StatusCode, "ListPrompts should not return 401 with valid token")
	})
}

// TestMcpEndpointTokenExpiration tests that expired tokens are rejected.
// It verifies token expiration and signature validation.
func (suite *McpTestSuite) TestMcpEndpointTokenExpiration() {
	suite.T().Log("Testing MCP endpoint token expiration handling")

	// Test 1: Expired or malformed token
	suite.Run("ExpiredToken", func() {
		// Create a JWT with expired timestamp (exp: 1 = 1970-01-01)
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMDAxIiwiZXhwIjoxfQ.invalid"

		body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`

		resp := suite.makeMcpRequestWithToken(body, expiredToken)

		suite.Equal(401, resp.StatusCode, "Should return 401 with expired/invalid token")
		suite.T().Log("Expired token correctly rejected with 401")
	})
}

// TestMcpEndpointAuthorizationHeader tests different Authorization header formats.
func (suite *McpTestSuite) TestMcpEndpointAuthorizationHeader() {
	suite.T().Log("Testing MCP endpoint Authorization header formats")

	token := suite.getAccessToken()
	suite.NotEmpty(token, "Should get valid access token")

	suite.Run("BearerPrefix", func() {
		body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`

		req := httptest.NewRequest(fiber.MethodPost, "/mcp", strings.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)

		resp, err := suite.app.Test(req, 30*time.Second)
		suite.Require().NoError(err)

		suite.NotEqual(401, resp.StatusCode, "Should accept Bearer prefix")
	})

	suite.Run("LowerCaseBearer", func() {
		body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`

		req := httptest.NewRequest(fiber.MethodPost, "/mcp", strings.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set(fiber.HeaderAuthorization, "bearer "+token)

		resp, err := suite.app.Test(req, 30*time.Second)
		suite.Require().NoError(err)

		// SDK accepts lowercase bearer
		suite.NotEqual(401, resp.StatusCode, "Should accept lowercase bearer prefix")
	})

	suite.Run("NoPrefix", func() {
		body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}`

		req := httptest.NewRequest(fiber.MethodPost, "/mcp", strings.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set(fiber.HeaderAuthorization, token) // No "Bearer " prefix

		resp, err := suite.app.Test(req, 30*time.Second)
		suite.Require().NoError(err)

		suite.Equal(401, resp.StatusCode, "Should reject token without Bearer prefix")
	})
}

func TestMcpSuite(t *testing.T) {
	suite.Run(t, new(McpTestSuite))
}
