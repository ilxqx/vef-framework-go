package security_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/guregu/null/v6"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"

	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/app"
	appTest "github.com/ilxqx/vef-framework-go/internal/app/test"
	isecurity "github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/result"
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

// MockUserInfoLoader is a mock implementation of security.UserInfoLoader for testing.
type MockUserInfoLoader struct {
	mock.Mock
}

func (m *MockUserInfoLoader) LoadUserInfo(ctx context.Context, principal *security.Principal, params map[string]any) (*security.UserInfo, error) {
	args := m.Called(ctx, principal, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*security.UserInfo), args.Error(1)
}

// AuthResourceTestSuite is the test suite for AuthResource.
type AuthResourceTestSuite struct {
	suite.Suite

	ctx            context.Context
	app            *app.App
	stop           func()
	userLoader     *MockUserLoader
	userInfoLoader *MockUserInfoLoader
	jwtSecret      string
	testUser       *security.Principal
}

// SetupSuite runs once before all tests in the suite.
func (suite *AuthResourceTestSuite) SetupSuite() {
	suite.T().Log("Setting up AuthResourceTestSuite - initializing test app and mocks")

	suite.ctx = context.Background()
	suite.jwtSecret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	suite.testUser = security.NewUser("user001", "Test User", "admin", "user")
	suite.testUser.Details = map[string]any{
		"email":  "test@example.com",
		"phone":  "1234567890",
		"status": "active",
	}

	suite.userLoader = new(MockUserLoader)
	suite.userInfoLoader = new(MockUserInfoLoader)

	suite.setupTestApp()

	suite.T().Log("AuthResourceTestSuite setup complete - test app and mocks ready")
}

// TearDownSuite runs once after all tests in the suite.
func (suite *AuthResourceTestSuite) TearDownSuite() {
	suite.T().Log("Tearing down AuthResourceTestSuite")

	if suite.stop != nil {
		suite.stop()
	}

	suite.T().Log("AuthResourceTestSuite teardown complete")
}

// SetupTest runs before each test.
func (suite *AuthResourceTestSuite) SetupTest() {
	// Clear only the calls history, keep the expectations
	suite.userLoader.Calls = nil
	suite.userInfoLoader.Calls = nil
}

func (suite *AuthResourceTestSuite) setupTestApp() {
	// Hash the password for test user
	hashedPassword, err := security.HashPassword("password123")
	suite.Require().NoError(err)

	suite.app, suite.stop = appTest.NewTestApp(
		suite.T(),
		// Provide mock user loader
		fx.Supply(
			fx.Annotate(
				suite.userLoader,
				fx.As(new(security.UserLoader)),
			),
		),
		// Provide mock user info loader
		fx.Supply(
			fx.Annotate(
				suite.userInfoLoader,
				fx.As(new(security.UserInfoLoader)),
			),
		),
		// Replace security config with test values
		fx.Replace(
			&config.DatasourceConfig{
				Type: "sqlite",
			},
			&config.SecurityConfig{
				TokenExpires: 24 * time.Hour,
			},
			&security.JwtConfig{
				Secret:   suite.jwtSecret,
				Audience: "test-app",
			},
		),
		// Setup default mock responses
		fx.Invoke(func() {
			// Default LoadByUsername behavior
			suite.userLoader.On("LoadByUsername", mock.Anything, "testuser").
				Return(suite.testUser, hashedPassword, nil).
				Maybe()

			// Default LoadById behavior
			suite.userLoader.On("LoadById", mock.Anything, "user001").
				Return(suite.testUser, nil).
				Maybe()

			// User not found cases
			suite.userLoader.On("LoadByUsername", mock.Anything, "nonexistent").
				Return(nil, "", nil).
				Maybe()

			suite.userLoader.On("LoadById", mock.Anything, "nonexistent").
				Return(nil, nil).
				Maybe()
		}),
	)
}

// Helper methods for making API requests and reading responses

func (suite *AuthResourceTestSuite) makeApiRequest(body api.Request) *http.Response {
	jsonBody, err := encoding.ToJson(body)
	suite.Require().NoError(err, "Should encode request to JSON")

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *AuthResourceTestSuite) makeApiRequestWithToken(body api.Request, token string) *http.Response {
	jsonBody, err := encoding.ToJson(body)
	suite.Require().NoError(err, "Should encode request to JSON")

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, constants.AuthSchemeBearer+" "+token)

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err, "Api request should not fail")

	return resp
}

func (suite *AuthResourceTestSuite) readBody(resp *http.Response) result.Result {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	suite.Require().NoError(err, "Should read response body")
	res, err := encoding.FromJson[result.Result](string(body))
	suite.Require().NoError(err, "Should decode response JSON")

	return *res
}

func (suite *AuthResourceTestSuite) readDataAsMap(data any) map[string]any {
	m, ok := data.(map[string]any)
	suite.Require().True(ok, "Data should be a map")

	return m
}

// Test Cases

// TestLoginSuccess tests successful login with valid credentials.
func (suite *AuthResourceTestSuite) TestLoginSuccess() {
	suite.T().Log("Testing successful login")

	resp := suite.makeApiRequest(api.Request{
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
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Login should succeed")
	suite.Equal(i18n.T(result.OkMessage), body.Message, "Should return success message")

	data := suite.readDataAsMap(body.Data)
	suite.Contains(data, "accessToken", "Response should contain access token")
	suite.Contains(data, "refreshToken", "Response should contain refresh token")
	suite.NotEmpty(data["accessToken"], "Access token should not be empty")
	suite.NotEmpty(data["refreshToken"], "Refresh token should not be empty")

	suite.userLoader.AssertCalled(suite.T(), "LoadByUsername", mock.Anything, "testuser")
}

// TestLoginInvalidCredentials tests login failures with invalid credentials.
func (suite *AuthResourceTestSuite) TestLoginInvalidCredentials() {
	suite.T().Log("Testing login with invalid credentials")

	suite.Run("WrongPassword", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        isecurity.AuthTypePassword,
				"principal":   "testuser",
				"credentials": "wrongpassword",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Login should fail with wrong password")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code, "Should return credentials invalid error")
	})

	suite.Run("UserNotFound", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        isecurity.AuthTypePassword,
				"principal":   "nonexistent",
				"credentials": "password123",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Login should fail with non-existent user")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code, "Should return credentials invalid error")
	})
}

// TestLoginMissingParameters tests login failures with missing or invalid parameters.
func (suite *AuthResourceTestSuite) TestLoginMissingParameters() {
	suite.T().Log("Testing login with missing parameters")

	suite.Run("MissingUsername", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        isecurity.AuthTypePassword,
				"credentials": "password123",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Login should fail without username")
		suite.Equal(result.ErrCodePrincipalInvalid, body.Code, "Should return principal invalid error")
	})

	suite.Run("MissingPassword", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":      isecurity.AuthTypePassword,
				"principal": "testuser",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Login should fail without password")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code, "Should return credentials invalid error")
	})

	suite.Run("EmptyPassword", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        isecurity.AuthTypePassword,
				"principal":   "testuser",
				"credentials": "",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Login should fail with empty password")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code, "Should return credentials invalid error")
	})

	suite.Run("LoaderRecordNotFoundError", func() {
		username := "loaderNotFound"
		suite.userLoader.On("LoadByUsername", mock.Anything, username).
			Return((*security.Principal)(nil), "", result.ErrRecordNotFound).
			Once()

		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        isecurity.AuthTypePassword,
				"principal":   username,
				"credentials": "password123",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Login should fail when loader reports record not found")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code, "Should return credentials invalid error")
		suite.userLoader.AssertExpectations(suite.T())
	})

	suite.Run("LoaderUnexpectedError", func() {
		username := "loaderUnexpected"
		suite.userLoader.On("LoadByUsername", mock.Anything, username).
			Return((*security.Principal)(nil), "", errors.New("loader failure")).
			Once()

		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        isecurity.AuthTypePassword,
				"principal":   username,
				"credentials": "password123",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Login should fail when loader returns unexpected error")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code, "Should return credentials invalid error")
		suite.userLoader.AssertExpectations(suite.T())
	})
}

// TestRefreshSuccess tests successful token refresh.
func (suite *AuthResourceTestSuite) TestRefreshSuccess() {
	suite.T().Log("Testing successful token refresh")

	loginResp := suite.makeApiRequest(api.Request{
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
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk(), "Login should succeed")

	tokens := suite.readDataAsMap(loginBody.Data)
	refreshToken := tokens["refreshToken"].(string)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "refresh",
			Version:  "v1",
		},
		Params: map[string]any{
			"refreshToken": refreshToken,
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Refresh should succeed")
	suite.Equal(i18n.T(result.OkMessage), body.Message, "Should return success message")

	data := suite.readDataAsMap(body.Data)
	suite.Contains(data, "accessToken", "Response should contain access token")
	suite.Contains(data, "refreshToken", "Response should contain refresh token")
	suite.NotEmpty(data["accessToken"], "Access token should not be empty")
	suite.NotEmpty(data["refreshToken"], "Refresh token should not be empty")

	suite.NotEqual(tokens["accessToken"], data["accessToken"], "New access token should be different")

	suite.userLoader.AssertCalled(suite.T(), "LoadById", mock.Anything, "user001")
}

// TestRefreshInvalidToken tests refresh failures with invalid tokens.
func (suite *AuthResourceTestSuite) TestRefreshInvalidToken() {
	suite.T().Log("Testing refresh with invalid tokens")

	suite.Run("InvalidToken", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "refresh",
				Version:  "v1",
			},
			Params: map[string]any{
				"refreshToken": "invalid.token.here",
			},
		})

		suite.True(resp.StatusCode == 200 || resp.StatusCode == 401,
			"Should return 200 or 401, got %d", resp.StatusCode)

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Refresh should fail with invalid token")
		suite.Equal(result.ErrCodeTokenInvalid, body.Code, "Should return token invalid error")
	})

	suite.Run("EmptyToken", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "refresh",
				Version:  "v1",
			},
			Params: map[string]any{
				"refreshToken": "",
			},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Refresh should fail with empty token")
		suite.Equal(result.ErrCodePrincipalInvalid, body.Code, "Should return principal invalid error")
	})

	suite.Run("MissingToken", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "refresh",
				Version:  "v1",
			},
			Params: map[string]any{},
		})

		suite.Equal(200, resp.StatusCode, "Should return 200 OK")

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Refresh should fail without token")
		suite.Equal(result.ErrCodePrincipalInvalid, body.Code, "Should return principal invalid error")
	})
}

// TestRefreshWithAccessToken tests that refresh fails when using an access token instead of refresh token.
func (suite *AuthResourceTestSuite) TestRefreshWithAccessToken() {
	suite.T().Log("Testing refresh with access token (should fail)")

	loginResp := suite.makeApiRequest(api.Request{
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
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk(), "Login should succeed")

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "refresh",
			Version:  "v1",
		},
		Params: map[string]any{
			"refreshToken": accessToken,
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Refresh should fail with access token")
	suite.Equal(result.ErrCodeTokenInvalid, body.Code, "Should return token invalid error")
}

// TestRefreshUserNotFound tests refresh failure when user is not found.
func (suite *AuthResourceTestSuite) TestRefreshUserNotFound() {
	suite.T().Log("Testing refresh when user is not found")

	loginResp := suite.makeApiRequest(api.Request{
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
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk(), "Login should succeed")

	tokens := suite.readDataAsMap(loginBody.Data)
	refreshToken := tokens["refreshToken"].(string)

	prevExpected := append([]*mock.Call(nil), suite.userLoader.ExpectedCalls...)
	defer func() { suite.userLoader.ExpectedCalls = prevExpected }()

	call := suite.userLoader.On("LoadById", mock.Anything, mock.Anything).Return((*security.Principal)(nil), nil).Once()
	if n := len(suite.userLoader.ExpectedCalls); n > 1 {
		last := suite.userLoader.ExpectedCalls[n-1]
		suite.userLoader.ExpectedCalls = append([]*mock.Call{last}, suite.userLoader.ExpectedCalls[:n-1]...)
		_ = call
	}

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "refresh",
			Version:  "v1",
		},
		Params: map[string]any{
			"refreshToken": refreshToken,
		},
	})

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Refresh should fail when user not found")
	suite.Equal(result.ErrCodeRecordNotFound, body.Code, "Should return record not found error")
}

// TestLogoutSuccess tests successful logout.
func (suite *AuthResourceTestSuite) TestLogoutSuccess() {
	suite.T().Log("Testing successful logout")

	loginResp := suite.makeApiRequest(api.Request{
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
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk(), "Login should succeed")

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)

	resp := suite.makeApiRequestWithToken(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "logout",
			Version:  "v1",
		},
	}, accessToken)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Logout should succeed")
	suite.Equal(i18n.T(result.OkMessage), body.Message, "Should return success message")
}

// TestLoginAndRefreshFlow tests the complete login and refresh flow.
func (suite *AuthResourceTestSuite) TestLoginAndRefreshFlow() {
	suite.T().Log("Testing complete login and refresh flow")

	loginResp := suite.makeApiRequest(api.Request{
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
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk(), "Login should succeed")

	tokens1 := suite.readDataAsMap(loginBody.Data)

	refreshResp1 := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "refresh",
			Version:  "v1",
		},
		Params: map[string]any{
			"refreshToken": tokens1["refreshToken"],
		},
	})

	refreshBody1 := suite.readBody(refreshResp1)
	suite.True(refreshBody1.IsOk(), "First refresh should succeed")

	tokens2 := suite.readDataAsMap(refreshBody1.Data)

	suite.NotEqual(tokens1["accessToken"], tokens2["accessToken"], "New access token should be different")
	suite.NotEqual(tokens1["refreshToken"], tokens2["refreshToken"], "New refresh token should be different")

	refreshResp2 := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "refresh",
			Version:  "v1",
		},
		Params: map[string]any{
			"refreshToken": tokens2["refreshToken"],
		},
	})

	refreshBody2 := suite.readBody(refreshResp2)
	suite.True(refreshBody2.IsOk(), "Second refresh should succeed")

	tokens3 := suite.readDataAsMap(refreshBody2.Data)

	suite.NotEqual(tokens2["accessToken"], tokens3["accessToken"], "Tokens should keep changing")
	suite.NotEqual(tokens2["refreshToken"], tokens3["refreshToken"], "Tokens should keep changing")

	logoutResp := suite.makeApiRequestWithToken(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "logout",
			Version:  "v1",
		},
	}, tokens3["accessToken"].(string))

	logoutBody := suite.readBody(logoutResp)
	suite.True(logoutBody.IsOk(), "Logout should succeed")
}

// TestTokenDetails tests token structure and format.
func (suite *AuthResourceTestSuite) TestTokenDetails() {
	suite.T().Log("Testing token details and format")

	loginResp := suite.makeApiRequest(api.Request{
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
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk(), "Login should succeed")

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)
	refreshToken := tokens["refreshToken"].(string)

	suite.NotEmpty(accessToken, "Access token should not be empty")
	suite.NotEmpty(refreshToken, "Refresh token should not be empty")

	suite.NotEqual(accessToken, refreshToken, "Tokens should be different")

	suite.Equal(3, len(strings.Split(accessToken, ".")), "Access token should be JWT format (3 parts)")
	suite.Equal(3, len(strings.Split(refreshToken, ".")), "Refresh token should be JWT format (3 parts)")
}

// TestGetUserInfoSuccess tests successful retrieval of user information.
func (suite *AuthResourceTestSuite) TestGetUserInfoSuccess() {
	suite.T().Log("Testing successful get user info")

	loginResp := suite.makeApiRequest(api.Request{
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
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk(), "Login should succeed")

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)

	avatarURL := "https://example.com/avatar.jpg"
	expectedUserInfo := &security.UserInfo{
		Id:     "user001",
		Name:   "Test User",
		Gender: security.GenderMale,
		Avatar: null.StringFrom(avatarURL),
		PermTokens: []string{
			"user:read",
			"user:write",
			"order:read",
		},
		Menus: []security.UserMenu{
			{
				Type: security.UserMenuTypeDirectory,
				Path: "/system",
				Name: "System Management",
				Icon: null.StringFrom("setting"),
				Children: []security.UserMenu{
					{
						Type: security.UserMenuTypeMenu,
						Path: "/system/users",
						Name: "User Management",
						Icon: null.StringFrom("user"),
					},
				},
			},
		},
	}

	suite.userInfoLoader.On("LoadUserInfo", mock.Anything, mock.MatchedBy(func(p *security.Principal) bool {
		return p.Id == "user001"
	}), mock.Anything).Return(expectedUserInfo, nil).Once()

	resp := suite.makeApiRequestWithToken(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "get_user_info",
			Version:  "v1",
		},
	}, accessToken)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Get user info should succeed")
	suite.Equal(i18n.T(result.OkMessage), body.Message, "Should return success message")

	data := suite.readDataAsMap(body.Data)
	suite.Equal("user001", data["id"], "User ID should match")
	suite.Equal("Test User", data["name"], "User name should match")
	suite.Equal("male", data["gender"], "Gender should match")
	suite.Equal(avatarURL, data["avatar"], "Avatar URL should match")

	permTokens, ok := data["permTokens"].([]any)
	suite.True(ok, "Permission tokens should be an array")
	suite.Len(permTokens, 3, "Should have 3 permission tokens")
	suite.Contains(permTokens, "user:read", "Should contain user:read permission")
	suite.Contains(permTokens, "user:write", "Should contain user:write permission")
	suite.Contains(permTokens, "order:read", "Should contain order:read permission")

	menus, ok := data["menus"].([]any)
	suite.True(ok, "Menus should be an array")
	suite.Len(menus, 1, "Should have 1 menu")

	firstMenu := menus[0].(map[string]any)
	suite.Equal("directory", firstMenu["type"], "Menu type should be directory")
	suite.Equal("/system", firstMenu["path"], "Menu path should match")
	suite.Equal("System Management", firstMenu["name"], "Menu name should match")
	suite.Equal("setting", firstMenu["icon"], "Menu icon should match")

	children, ok := firstMenu["children"].([]any)
	suite.True(ok, "Children should be an array")
	suite.Len(children, 1, "Should have 1 child menu")

	suite.userInfoLoader.AssertExpectations(suite.T())
}

// TestGetUserInfoUnauthenticated tests get user info without authentication.
func (suite *AuthResourceTestSuite) TestGetUserInfoUnauthenticated() {
	suite.T().Log("Testing get user info without authentication")

	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "get_user_info",
			Version:  "v1",
		},
	})

	suite.Equal(401, resp.StatusCode, "Should return 401 Unauthorized")
}

// TestGetUserInfoLoaderError tests get user info when loader returns an error.
func (suite *AuthResourceTestSuite) TestGetUserInfoLoaderError() {
	suite.T().Log("Testing get user info with loader error")

	loginResp := suite.makeApiRequest(api.Request{
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
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk(), "Login should succeed")

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)

	suite.userInfoLoader.On("LoadUserInfo", mock.Anything, mock.MatchedBy(func(p *security.Principal) bool {
		return p.Id == "user001"
	}), mock.Anything).Return((*security.UserInfo)(nil), errors.New("database connection failed")).Once()

	resp := suite.makeApiRequestWithToken(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "get_user_info",
			Version:  "v1",
		},
	}, accessToken)

	suite.Equal(500, resp.StatusCode, "Should return 500 Internal Server Error")

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Get user info should fail when loader returns error")

	suite.userInfoLoader.AssertExpectations(suite.T())
}

// TestGetUserInfoWithEmptyMenus tests get user info with empty menus and permissions.
func (suite *AuthResourceTestSuite) TestGetUserInfoWithEmptyMenus() {
	suite.T().Log("Testing get user info with empty menus")

	loginResp := suite.makeApiRequest(api.Request{
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
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk(), "Login should succeed")

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)

	expectedUserInfo := &security.UserInfo{
		Id:         "user001",
		Name:       "Test User",
		Gender:     security.GenderUnknown,
		PermTokens: []string{},
		Menus:      []security.UserMenu{},
	}

	suite.userInfoLoader.On("LoadUserInfo", mock.Anything, mock.MatchedBy(func(p *security.Principal) bool {
		return p.Id == "user001"
	}), mock.Anything).Return(expectedUserInfo, nil).Once()

	resp := suite.makeApiRequestWithToken(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "get_user_info",
			Version:  "v1",
		},
	}, accessToken)

	suite.Equal(200, resp.StatusCode, "Should return 200 OK")

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Get user info should succeed")

	data := suite.readDataAsMap(body.Data)
	suite.Equal("user001", data["id"], "User ID should match")
	suite.Equal("Test User", data["name"], "User name should match")
	suite.Equal("unknown", data["gender"], "Gender should be unknown")
	suite.Nil(data["avatar"], "Avatar should be null when not set")

	permTokens, ok := data["permTokens"].([]any)
	suite.True(ok, "Permission tokens should be an array")
	suite.Len(permTokens, 0, "Permission tokens should be empty")

	menus, ok := data["menus"].([]any)
	suite.True(ok, "Menus should be an array")
	suite.Len(menus, 0, "Menus should be empty")

	suite.userInfoLoader.AssertExpectations(suite.T())
}

// TestAuthResourceSuite runs the test suite.
func TestAuthResourceSuite(t *testing.T) {
	suite.Run(t, new(AuthResourceTestSuite))
}
