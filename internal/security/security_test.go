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

	"github.com/ilxqx/vef-framework-go"
	"github.com/ilxqx/vef-framework-go/api"
	"github.com/ilxqx/vef-framework-go/config"
	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/encoding"
	"github.com/ilxqx/vef-framework-go/i18n"
	"github.com/ilxqx/vef-framework-go/internal/app"
	appTest "github.com/ilxqx/vef-framework-go/internal/app/test"
	"github.com/ilxqx/vef-framework-go/internal/security"
	"github.com/ilxqx/vef-framework-go/result"
	securityPkg "github.com/ilxqx/vef-framework-go/security"
)

// MockUserLoader is a mock implementation of security.UserLoader for testing.
type MockUserLoader struct {
	mock.Mock
}

func (m *MockUserLoader) LoadByUsername(ctx context.Context, username string) (*securityPkg.Principal, string, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}

	return args.Get(0).(*securityPkg.Principal), args.String(1), args.Error(2)
}

func (m *MockUserLoader) LoadById(ctx context.Context, id string) (*securityPkg.Principal, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*securityPkg.Principal), args.Error(1)
}

// MockUserInfoLoader is a mock implementation of security.UserInfoLoader for testing.
type MockUserInfoLoader struct {
	mock.Mock
}

func (m *MockUserInfoLoader) LoadUserInfo(ctx context.Context, principal *securityPkg.Principal, params map[string]any) (*securityPkg.UserInfo, error) {
	args := m.Called(ctx, principal, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*securityPkg.UserInfo), args.Error(1)
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
	testUser       *securityPkg.Principal
}

// SetupSuite runs once before all tests in the suite.
func (suite *AuthResourceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.jwtSecret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	// Create test user principal
	suite.testUser = securityPkg.NewUser("user001", "Test User", "admin", "user")
	suite.testUser.Details = map[string]any{
		"email":  "test@example.com",
		"phone":  "1234567890",
		"status": "active",
	}

	// Create mock user loader
	suite.userLoader = new(MockUserLoader)
	// Create mock user info loader
	suite.userInfoLoader = new(MockUserInfoLoader)

	// Setup test app
	suite.setupTestApp()
}

// TearDownSuite runs once after all tests in the suite.
func (suite *AuthResourceTestSuite) TearDownSuite() {
	if suite.stop != nil {
		suite.stop()
	}
}

// SetupTest runs before each test.
func (suite *AuthResourceTestSuite) SetupTest() {
	// Clear only the calls history, keep the expectations
	suite.userLoader.Calls = nil
	suite.userInfoLoader.Calls = nil
}

func (suite *AuthResourceTestSuite) setupTestApp() {
	// Hash the password for test user
	hashedPassword, err := securityPkg.HashPassword("password123")
	suite.Require().NoError(err)

	suite.app, suite.stop = appTest.NewTestApp(
		suite.T(),
		// Provide the auth resource
		vef.ProvideApiResource(security.NewAuthResource, ``, ``, `optional:"true"`),
		// Provide mock user loader
		fx.Supply(
			fx.Annotate(
				suite.userLoader,
				fx.As(new(securityPkg.UserLoader)),
			),
		),
		// Provide mock user info loader
		fx.Supply(
			fx.Annotate(
				suite.userInfoLoader,
				fx.As(new(securityPkg.UserInfoLoader)),
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
			&securityPkg.JwtConfig{
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

// Helper methods

func (suite *AuthResourceTestSuite) makeApiRequest(body api.Request) *http.Response {
	jsonBody, err := encoding.ToJSON(body)
	suite.Require().NoError(err)

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	return resp
}

func (suite *AuthResourceTestSuite) makeApiRequestWithToken(body api.Request, token string) *http.Response {
	jsonBody, err := encoding.ToJSON(body)
	suite.Require().NoError(err)

	req := httptest.NewRequest(fiber.MethodPost, "/api", strings.NewReader(jsonBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, constants.AuthSchemeBearer+" "+token)

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	return resp
}

func (suite *AuthResourceTestSuite) readBody(resp *http.Response) result.Result {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	suite.Require().NoError(err)
	res, err := encoding.FromJSON[result.Result](string(body))
	suite.Require().NoError(err)

	return *res
}

func (suite *AuthResourceTestSuite) readDataAsMap(data any) map[string]any {
	m, ok := data.(map[string]any)
	suite.Require().True(ok, "Expected data to be a map")

	return m
}

// Test Cases

func (suite *AuthResourceTestSuite) TestLoginSuccess() {
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"type":        security.AuthTypePassword,
			"principal":   "testuser",
			"credentials": "password123",
		},
	})

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful login")
	suite.Equal(i18n.T(result.OkMessage), body.Message)

	// Verify tokens are returned
	data := suite.readDataAsMap(body.Data)
	suite.Contains(data, "accessToken")
	suite.Contains(data, "refreshToken")
	suite.NotEmpty(data["accessToken"])
	suite.NotEmpty(data["refreshToken"])

	// Verify mock was called
	suite.userLoader.AssertCalled(suite.T(), "LoadByUsername", mock.Anything, "testuser")
}

func (suite *AuthResourceTestSuite) TestLoginInvalidCredentials() {
	suite.Run("WrongPassword", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        security.AuthTypePassword,
				"principal":   "testuser",
				"credentials": "wrongpassword",
			},
		})

		suite.Equal(200, resp.StatusCode)

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Expected login to fail with wrong password")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code)
	})

	suite.Run("UserNotFound", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        security.AuthTypePassword,
				"principal":   "nonexistent",
				"credentials": "password123",
			},
		})

		suite.Equal(200, resp.StatusCode)

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Expected login to fail with non-existent user")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code)
	})
}

func (suite *AuthResourceTestSuite) TestLoginMissingParameters() {
	suite.Run("MissingUsername", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        security.AuthTypePassword,
				"credentials": "password123",
			},
		})

		suite.Equal(200, resp.StatusCode)

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Expected login to fail without username")
		suite.Equal(result.ErrCodePrincipalInvalid, body.Code)
	})

	suite.Run("MissingPassword", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":      security.AuthTypePassword,
				"principal": "testuser",
			},
		})

		suite.Equal(200, resp.StatusCode)

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Expected login to fail without password")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code)
	})

	suite.Run("EmptyPassword", func() {
		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        security.AuthTypePassword,
				"principal":   "testuser",
				"credentials": "",
			},
		})

		suite.Equal(200, resp.StatusCode)

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Expected login to fail with empty password")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code)
	})

	suite.Run("LoaderRecordNotFoundError", func() {
		username := "loaderNotFound"
		suite.userLoader.On("LoadByUsername", mock.Anything, username).
			Return((*securityPkg.Principal)(nil), "", result.ErrRecordNotFound).
			Once()

		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        security.AuthTypePassword,
				"principal":   username,
				"credentials": "password123",
			},
		})

		suite.Equal(200, resp.StatusCode)

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Expected login to fail when loader reports record not found")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code)
		suite.userLoader.AssertExpectations(suite.T())
	})

	suite.Run("LoaderUnexpectedError", func() {
		username := "loaderUnexpected"
		suite.userLoader.On("LoadByUsername", mock.Anything, username).
			Return((*securityPkg.Principal)(nil), "", errors.New("loader failure")).
			Once()

		resp := suite.makeApiRequest(api.Request{
			Identifier: api.Identifier{
				Resource: "security/auth",
				Action:   "login",
				Version:  "v1",
			},
			Params: map[string]any{
				"type":        security.AuthTypePassword,
				"principal":   username,
				"credentials": "password123",
			},
		})

		suite.Equal(200, resp.StatusCode)

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Expected login to fail when loader returns unexpected error")
		suite.Equal(result.ErrCodeCredentialsInvalid, body.Code)
		suite.userLoader.AssertExpectations(suite.T())
	})
}

func (suite *AuthResourceTestSuite) TestRefreshSuccess() {
	// First, login to get tokens
	loginResp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"type":        security.AuthTypePassword,
			"principal":   "testuser",
			"credentials": "password123",
		},
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk())

	tokens := suite.readDataAsMap(loginBody.Data)
	refreshToken := tokens["refreshToken"].(string)

	// Now test refresh
	// Note: In test mode (VEF_TEST_MODE=true), the refresh token has notBefore=0,
	// allowing immediate use. In production, notBefore would be accessTokenExpires/2.
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

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful refresh")
	suite.Equal(i18n.T(result.OkMessage), body.Message)

	// Verify new tokens are returned
	data := suite.readDataAsMap(body.Data)
	suite.Contains(data, "accessToken")
	suite.Contains(data, "refreshToken")
	suite.NotEmpty(data["accessToken"])
	suite.NotEmpty(data["refreshToken"])

	// Verify the new access token is different from the old one
	suite.NotEqual(tokens["accessToken"], data["accessToken"])

	// Verify LoadById was called
	suite.userLoader.AssertCalled(suite.T(), "LoadById", mock.Anything, "user001")
}

func (suite *AuthResourceTestSuite) TestRefreshInvalidToken() {
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

		// Authentication failures may return either 401 or 200 with error body
		suite.True(resp.StatusCode == 200 || resp.StatusCode == 401,
			"Expected status code 200 or 401, got %d", resp.StatusCode)

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Expected refresh to fail with invalid token")
		suite.Equal(result.ErrCodeTokenInvalid, body.Code)
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

		suite.Equal(200, resp.StatusCode)

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Expected refresh to fail with empty token")
		suite.Equal(result.ErrCodePrincipalInvalid, body.Code)
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

		suite.Equal(200, resp.StatusCode)

		body := suite.readBody(resp)
		suite.False(body.IsOk(), "Expected refresh to fail without token")
		suite.Equal(result.ErrCodePrincipalInvalid, body.Code)
	})
}

func (suite *AuthResourceTestSuite) TestRefreshWithAccessToken() {
	// Login to get tokens
	loginResp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"type":        security.AuthTypePassword,
			"principal":   "testuser",
			"credentials": "password123",
		},
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk())

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)

	// Try to refresh with access token (should fail)
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

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Expected refresh to fail with access token")
	suite.Equal(result.ErrCodeTokenInvalid, body.Code)
}

func (suite *AuthResourceTestSuite) TestRefreshUserNotFound() {
	// In test mode, refresh token's notBefore is disabled, so we can refresh immediately.
	// This test verifies that when the user is not found during refresh, the Api returns the expected error.

	// Step 1: Login to obtain tokens
	loginResp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"type":        security.AuthTypePassword,
			"principal":   "testuser",
			"credentials": "password123",
		},
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk(), "expected login success before refresh test")

	tokens := suite.readDataAsMap(loginBody.Data)
	refreshToken := tokens["refreshToken"].(string)

	// Step 2: Simulate user deletion/not found for any user id used in the refresh token
	// Save current expectations and restore after this test to avoid side effects on other tests
	prevExpected := append([]*mock.Call(nil), suite.userLoader.ExpectedCalls...)
	defer func() { suite.userLoader.ExpectedCalls = prevExpected }()

	// Add an override for the next LoadById call and move it to the front so it matches first
	call := suite.userLoader.On("LoadById", mock.Anything, mock.Anything).Return((*securityPkg.Principal)(nil), nil).Once()
	// Reorder: move last added expectation to the front
	if n := len(suite.userLoader.ExpectedCalls); n > 1 {
		last := suite.userLoader.ExpectedCalls[n-1]
		suite.userLoader.ExpectedCalls = append([]*mock.Call{last}, suite.userLoader.ExpectedCalls[:n-1]...)
		// Ensure the pointer 'call' still refers to the correct entry (not strictly necessary for matching)
		_ = call
	}

	// Step 3: Attempt refresh, expect record not found
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

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "expected refresh to fail when user not found")
	suite.Equal(result.ErrCodeRecordNotFound, body.Code)
}

func (suite *AuthResourceTestSuite) TestLogoutSuccess() {
	// Note: Logout requires authentication because it's not marked as Public in auth_resource.go
	// First login to get an access token
	loginResp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"type":        security.AuthTypePassword,
			"principal":   "testuser",
			"credentials": "password123",
		},
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk())

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)

	// Now logout with the access token
	resp := suite.makeApiRequestWithToken(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "logout",
			Version:  "v1",
		},
	}, accessToken)

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful logout")
	suite.Equal(i18n.T(result.OkMessage), body.Message)
}

func (suite *AuthResourceTestSuite) TestLoginAndRefreshFlow() {
	// Step 1: Login
	loginResp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"type":        security.AuthTypePassword,
			"principal":   "testuser",
			"credentials": "password123",
		},
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk())

	tokens1 := suite.readDataAsMap(loginBody.Data)

	// Step 2: Refresh the token
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
	suite.True(refreshBody1.IsOk())

	tokens2 := suite.readDataAsMap(refreshBody1.Data)

	// Verify new tokens are different
	suite.NotEqual(tokens1["accessToken"], tokens2["accessToken"])
	suite.NotEqual(tokens1["refreshToken"], tokens2["refreshToken"])

	// Step 3: Refresh again with the new refresh token
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
	suite.True(refreshBody2.IsOk())

	tokens3 := suite.readDataAsMap(refreshBody2.Data)

	// Verify tokens keep changing
	suite.NotEqual(tokens2["accessToken"], tokens3["accessToken"])
	suite.NotEqual(tokens2["refreshToken"], tokens3["refreshToken"])

	// Step 4: Logout
	logoutResp := suite.makeApiRequestWithToken(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "logout",
			Version:  "v1",
		},
	}, tokens3["accessToken"].(string))

	logoutBody := suite.readBody(logoutResp)
	suite.True(logoutBody.IsOk())
}

func (suite *AuthResourceTestSuite) TestTokenDetails() {
	// Login to get tokens
	loginResp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"type":        security.AuthTypePassword,
			"principal":   "testuser",
			"credentials": "password123",
		},
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk())

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)
	refreshToken := tokens["refreshToken"].(string)

	// Verify tokens are non-empty
	suite.NotEmpty(accessToken)
	suite.NotEmpty(refreshToken)

	// Verify tokens are different
	suite.NotEqual(accessToken, refreshToken)

	// Verify tokens are Jwt format (3 parts separated by dots)
	suite.Equal(3, len(strings.Split(accessToken, ".")))
	suite.Equal(3, len(strings.Split(refreshToken, ".")))
}

func (suite *AuthResourceTestSuite) TestGetUserInfoSuccess() {
	// First, login to get an access token
	loginResp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"type":        security.AuthTypePassword,
			"principal":   "testuser",
			"credentials": "password123",
		},
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk())

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)

	// Setup mock user info
	avatarURL := "https://example.com/avatar.jpg"
	expectedUserInfo := &securityPkg.UserInfo{
		Id:     "user001",
		Name:   "Test User",
		Gender: securityPkg.GenderMale,
		Avatar: null.StringFrom(avatarURL),
		PermTokens: []string{
			"user:read",
			"user:write",
			"order:read",
		},
		Menus: []securityPkg.UserMenu{
			{
				Type: securityPkg.UserMenuTypeDirectory,
				Path: "/system",
				Name: "System Management",
				Icon: null.StringFrom("setting"),
				Children: []securityPkg.UserMenu{
					{
						Type: securityPkg.UserMenuTypeMenu,
						Path: "/system/users",
						Name: "User Management",
						Icon: null.StringFrom("user"),
					},
				},
			},
		},
	}

	suite.userInfoLoader.On("LoadUserInfo", mock.Anything, mock.MatchedBy(func(p *securityPkg.Principal) bool {
		return p.Id == "user001"
	}), mock.Anything).Return(expectedUserInfo, nil).Once()

	// Call get_user_info with access token
	resp := suite.makeApiRequestWithToken(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "get_user_info",
			Version:  "v1",
		},
	}, accessToken)

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful get_user_info")
	suite.Equal(i18n.T(result.OkMessage), body.Message)

	// Verify user info structure
	data := suite.readDataAsMap(body.Data)
	suite.Equal("user001", data["id"])
	suite.Equal("Test User", data["name"])
	suite.Equal("male", data["gender"])
	suite.Equal(avatarURL, data["avatar"])

	// Verify permission tokens
	permTokens, ok := data["permTokens"].([]any)
	suite.True(ok, "permTokens should be an array")
	suite.Len(permTokens, 3)
	suite.Contains(permTokens, "user:read")
	suite.Contains(permTokens, "user:write")
	suite.Contains(permTokens, "order:read")

	// Verify menus structure
	menus, ok := data["menus"].([]any)
	suite.True(ok, "menus should be an array")
	suite.Len(menus, 1)

	firstMenu := menus[0].(map[string]any)
	suite.Equal("directory", firstMenu["type"])
	suite.Equal("/system", firstMenu["path"])
	suite.Equal("System Management", firstMenu["name"])
	suite.Equal("setting", firstMenu["icon"])

	children, ok := firstMenu["children"].([]any)
	suite.True(ok, "children should be an array")
	suite.Len(children, 1)

	// Verify mock was called
	suite.userInfoLoader.AssertExpectations(suite.T())
}

func (suite *AuthResourceTestSuite) TestGetUserInfoUnauthenticated() {
	// Try to call get_user_info without authentication
	resp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "get_user_info",
			Version:  "v1",
		},
	})

	// Should be unauthorized
	suite.Equal(401, resp.StatusCode)
}

func (suite *AuthResourceTestSuite) TestGetUserInfoLoaderError() {
	// First, login to get an access token
	loginResp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"type":        security.AuthTypePassword,
			"principal":   "testuser",
			"credentials": "password123",
		},
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk())

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)

	// Setup mock to return error
	suite.userInfoLoader.On("LoadUserInfo", mock.Anything, mock.MatchedBy(func(p *securityPkg.Principal) bool {
		return p.Id == "user001"
	}), mock.Anything).Return((*securityPkg.UserInfo)(nil), errors.New("database connection failed")).Once()

	// Call get_user_info with access token
	resp := suite.makeApiRequestWithToken(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "get_user_info",
			Version:  "v1",
		},
	}, accessToken)

	// Unhandled errors return 500
	suite.Equal(500, resp.StatusCode)

	body := suite.readBody(resp)
	suite.False(body.IsOk(), "Expected get_user_info to fail when loader returns error")

	// Verify mock was called
	suite.userInfoLoader.AssertExpectations(suite.T())
}

func (suite *AuthResourceTestSuite) TestGetUserInfoWithEmptyMenus() {
	// First, login to get an access token
	loginResp := suite.makeApiRequest(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "login",
			Version:  "v1",
		},
		Params: map[string]any{
			"type":        security.AuthTypePassword,
			"principal":   "testuser",
			"credentials": "password123",
		},
	})

	loginBody := suite.readBody(loginResp)
	suite.True(loginBody.IsOk())

	tokens := suite.readDataAsMap(loginBody.Data)
	accessToken := tokens["accessToken"].(string)

	// Setup mock user info with empty menus and no optional fields
	expectedUserInfo := &securityPkg.UserInfo{
		Id:         "user001",
		Name:       "Test User",
		Gender:     securityPkg.GenderUnknown,
		PermTokens: []string{},
		Menus:      []securityPkg.UserMenu{},
	}

	suite.userInfoLoader.On("LoadUserInfo", mock.Anything, mock.MatchedBy(func(p *securityPkg.Principal) bool {
		return p.Id == "user001"
	}), mock.Anything).Return(expectedUserInfo, nil).Once()

	// Call get_user_info with access token
	resp := suite.makeApiRequestWithToken(api.Request{
		Identifier: api.Identifier{
			Resource: "security/auth",
			Action:   "get_user_info",
			Version:  "v1",
		},
	}, accessToken)

	suite.Equal(200, resp.StatusCode)

	body := suite.readBody(resp)
	suite.True(body.IsOk(), "Expected successful get_user_info")

	// Verify user info structure
	data := suite.readDataAsMap(body.Data)
	suite.Equal("user001", data["id"])
	suite.Equal("Test User", data["name"])
	suite.Equal("unknown", data["gender"])
	suite.Nil(data["avatar"], "avatar should be null when not set")

	// Verify empty arrays
	permTokens, ok := data["permTokens"].([]any)
	suite.True(ok, "permTokens should be an array")
	suite.Len(permTokens, 0)

	menus, ok := data["menus"].([]any)
	suite.True(ok, "menus should be an array")
	suite.Len(menus, 0)

	// Verify mock was called
	suite.userInfoLoader.AssertExpectations(suite.T())
}

// TestAuthResourceSuite runs the test suite.
func TestAuthResourceSuite(t *testing.T) {
	suite.Run(t, new(AuthResourceTestSuite))
}
