package security

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/security"
)

type MockExternalAppLoader struct {
	mock.Mock
}

func (m *MockExternalAppLoader) LoadByID(ctx context.Context, id string) (*security.Principal, string, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}

	return args.Get(0).(*security.Principal), args.String(1), args.Error(2)
}

type MockNonceStore struct {
	mock.Mock
}

func (m *MockNonceStore) Exists(ctx context.Context, appID, nonce string) (bool, error) {
	args := m.Called(ctx, appID, nonce)

	return args.Bool(0), args.Error(1)
}

func (m *MockNonceStore) Store(ctx context.Context, appID, nonce string, ttl time.Duration) error {
	args := m.Called(ctx, appID, nonce, ttl)

	return args.Error(0)
}

const testSecretHex = "af6675678bd81ad7c93c4a51d122ef61e9750fe5d42ceac1c33b293f36bc14c2"

// generateValidCredentials creates valid signature credentials for testing.
func generateValidCredentials(t *testing.T, appID, secret string) *security.SignatureCredentials {
	t.Helper()

	sig, err := security.NewSignature(secret, security.WithNonceStore(nil))
	require.NoError(t, err)

	result, err := sig.Sign(appID)
	require.NoError(t, err)

	return &security.SignatureCredentials{
		Timestamp: result.Timestamp,
		Nonce:     result.Nonce,
		Signature: result.Signature,
	}
}

func TestNewSignatureAuthenticator(t *testing.T) {
	t.Run("WithLoader", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		auth := NewSignatureAuthenticator(loader, nil)

		assert.NotNil(t, auth, "Authenticator should not be nil")
	})

	t.Run("WithoutLoader", func(t *testing.T) {
		auth := NewSignatureAuthenticator(nil, nil)

		assert.NotNil(t, auth, "Authenticator should not be nil even without loader")
	})

	t.Run("WithNonceStore", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)
		auth := NewSignatureAuthenticator(loader, nonceStore)

		assert.NotNil(t, auth, "Authenticator should not be nil")
	})
}

func TestSignatureAuthenticator_Supports(t *testing.T) {
	loader := new(MockExternalAppLoader)
	auth := NewSignatureAuthenticator(loader, nil)

	t.Run("SupportedKind", func(t *testing.T) {
		assert.True(t, auth.Supports(AuthKindSignature), "Should support signature kind")
	})

	t.Run("UnsupportedKinds", func(t *testing.T) {
		unsupportedKinds := []string{"password", "token", "jwt", "openapi", "bearer", ""}

		for _, kind := range unsupportedKinds {
			t.Run(kind, func(t *testing.T) {
				assert.False(t, auth.Supports(kind), "Should not support %q kind", kind)
			})
		}
	})
}

func TestSignatureAuthenticator_Authenticate(t *testing.T) {
	ctx := context.Background()

	t.Run("MissingAppID", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		auth := NewSignatureAuthenticator(loader, nil)

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:      AuthKindSignature,
			Principal: "",
			Credentials: &security.SignatureCredentials{
				Timestamp: time.Now().Unix(),
				Nonce:     "abcdefghijklmnop",
				Signature: "signature",
			},
		})

		assert.Error(t, err, "Should return error for empty appID")
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		auth := NewSignatureAuthenticator(loader, nil)

		testCases := []struct {
			name        string
			credentials any
		}{
			{"NilCredentials", nil},
			{"StringCredentials", "string credentials"},
			{"WrongStructType", struct{ Foo string }{Foo: "bar"}},
			{"IntCredentials", 12345},
			{"MapCredentials", map[string]any{"key": "value"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := auth.Authenticate(ctx, security.Authentication{
					Kind:        AuthKindSignature,
					Principal:   "app1",
					Credentials: tc.credentials,
				})
				assert.Error(t, err, "Should return error for invalid credentials type")
			})
		}
	})

	t.Run("AppNotFound", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		loader.On("LoadByID", mock.Anything, "app1").Return(nil, "", nil)

		auth := NewSignatureAuthenticator(loader, nil)

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:      AuthKindSignature,
			Principal: "app1",
			Credentials: &security.SignatureCredentials{
				Timestamp: time.Now().Unix(),
				Nonce:     "abcdefghijklmnop",
				Signature: "signature",
			},
		})

		assert.Error(t, err, "Should return error when app is not found")
		loader.AssertExpectations(t)
	})

	t.Run("LoaderReturnsError", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		expectedErr := errors.New("database connection failed")
		loader.On("LoadByID", mock.Anything, "app1").Return(nil, "", expectedErr)

		auth := NewSignatureAuthenticator(loader, nil)

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:      AuthKindSignature,
			Principal: "app1",
			Credentials: &security.SignatureCredentials{
				Timestamp: time.Now().Unix(),
				Nonce:     "abcdefghijklmnop",
				Signature: "signature",
			},
		})

		assert.ErrorIs(t, err, expectedErr, "Should return loader error")
		loader.AssertExpectations(t)
	})

	t.Run("EmptySecret", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, "", nil)

		auth := NewSignatureAuthenticator(loader, nil)

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:      AuthKindSignature,
			Principal: "app1",
			Credentials: &security.SignatureCredentials{
				Timestamp: time.Now().Unix(),
				Nonce:     "abcdefghijklmnop",
				Signature: "signature",
			},
		})

		assert.Error(t, err, "Should return error when secret is empty")
		loader.AssertExpectations(t)
	})
}

func TestSignatureAuthenticator_TimestampValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("ExpiredTimestamp", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)

		auth := NewSignatureAuthenticator(loader, nil)

		oldTimestamp := time.Now().Add(-10 * time.Minute).Unix()

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:      AuthKindSignature,
			Principal: "app1",
			Credentials: &security.SignatureCredentials{
				Timestamp: oldTimestamp,
				Nonce:     "abcdefghijklmnop",
				Signature: "0000000000000000000000000000000000000000000000000000000000000000",
			},
		})

		assert.Error(t, err, "Should return error for expired timestamp")
		loader.AssertExpectations(t)
	})

	t.Run("FutureTimestamp", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)

		auth := NewSignatureAuthenticator(loader, nil)

		futureTimestamp := time.Now().Add(10 * time.Minute).Unix()

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:      AuthKindSignature,
			Principal: "app1",
			Credentials: &security.SignatureCredentials{
				Timestamp: futureTimestamp,
				Nonce:     "abcdefghijklmnop",
				Signature: "0000000000000000000000000000000000000000000000000000000000000000",
			},
		})

		assert.Error(t, err, "Should return error for future timestamp")
		loader.AssertExpectations(t)
	})

	t.Run("ValidTimestampWithinTolerance", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)
		nonceStore.On("Exists", mock.Anything, "app1", mock.AnythingOfType("string")).Return(false, nil)
		nonceStore.On("Store", mock.Anything, "app1", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

		auth := NewSignatureAuthenticator(loader, nonceStore)

		credentials := generateValidCredentials(t, "app1", testSecretHex)

		result, err := auth.Authenticate(ctx, security.Authentication{
			Kind:        AuthKindSignature,
			Principal:   "app1",
			Credentials: credentials,
		})

		require.NoError(t, err, "Should authenticate with valid timestamp")
		assert.Equal(t, "app1", result.ID, "Principal ID should match")
		loader.AssertExpectations(t)
		nonceStore.AssertExpectations(t)
	})
}

func TestSignatureAuthenticator_SignatureValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("InvalidSignature", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)
		nonceStore.On("Exists", mock.Anything, "app1", mock.AnythingOfType("string")).Return(false, nil)

		auth := NewSignatureAuthenticator(loader, nonceStore)

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:      AuthKindSignature,
			Principal: "app1",
			Credentials: &security.SignatureCredentials{
				Timestamp: time.Now().Unix(),
				Nonce:     "abcdefghijklmnop",
				Signature: "0000000000000000000000000000000000000000000000000000000000000000",
			},
		})

		assert.Error(t, err, "Should return error for invalid signature")
		loader.AssertExpectations(t)
	})

	t.Run("MalformedSignature", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)
		nonceStore.On("Exists", mock.Anything, "app1", mock.AnythingOfType("string")).Return(false, nil)

		auth := NewSignatureAuthenticator(loader, nonceStore)

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:      AuthKindSignature,
			Principal: "app1",
			Credentials: &security.SignatureCredentials{
				Timestamp: time.Now().Unix(),
				Nonce:     "abcdefghijklmnop",
				Signature: "not-valid-hex",
			},
		})

		assert.Error(t, err, "Should return error for malformed signature")
		loader.AssertExpectations(t)
	})

	t.Run("WrongSecret", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		wrongSecret := "0000000000000000000000000000000000000000000000000000000000000000"
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, wrongSecret, nil)
		nonceStore.On("Exists", mock.Anything, "app1", mock.AnythingOfType("string")).Return(false, nil)

		auth := NewSignatureAuthenticator(loader, nonceStore)

		credentials := generateValidCredentials(t, "app1", testSecretHex)

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:        AuthKindSignature,
			Principal:   "app1",
			Credentials: credentials,
		})

		assert.Error(t, err, "Should return error when secret doesn't match")
		loader.AssertExpectations(t)
	})
}

func TestSignatureAuthenticator_NonceValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("NonceAlreadyUsed", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)
		nonceStore.On("Exists", mock.Anything, "app1", mock.AnythingOfType("string")).Return(true, nil)

		auth := NewSignatureAuthenticator(loader, nonceStore)

		credentials := generateValidCredentials(t, "app1", testSecretHex)

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:        AuthKindSignature,
			Principal:   "app1",
			Credentials: credentials,
		})

		assert.Error(t, err, "Should return error when nonce is already used")
		loader.AssertExpectations(t)
		nonceStore.AssertExpectations(t)
	})

	t.Run("NonceStoreExistsError", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)
		nonceStore.On("Exists", mock.Anything, "app1", mock.AnythingOfType("string")).Return(false, errors.New("redis connection failed"))

		auth := NewSignatureAuthenticator(loader, nonceStore)

		credentials := generateValidCredentials(t, "app1", testSecretHex)

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:        AuthKindSignature,
			Principal:   "app1",
			Credentials: credentials,
		})

		assert.Error(t, err, "Should return error when nonce store fails")
		loader.AssertExpectations(t)
		nonceStore.AssertExpectations(t)
	})

	t.Run("NonceStoreStoreError", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)
		nonceStore.On("Exists", mock.Anything, "app1", mock.AnythingOfType("string")).Return(false, nil)
		nonceStore.On("Store", mock.Anything, "app1", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(errors.New("redis write failed"))

		auth := NewSignatureAuthenticator(loader, nonceStore)

		credentials := generateValidCredentials(t, "app1", testSecretHex)

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:        AuthKindSignature,
			Principal:   "app1",
			Credentials: credentials,
		})

		assert.Error(t, err, "Should return error when nonce store write fails")
		loader.AssertExpectations(t)
		nonceStore.AssertExpectations(t)
	})
}

func TestSignatureAuthenticator_IPWhitelist(t *testing.T) {
	ctx := context.Background()

	t.Run("DisabledApp", func(t *testing.T) {
		loader := new(MockExternalAppLoader)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		principal.Details = &security.ExternalAppConfig{
			Enabled: false,
		}
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)

		auth := NewSignatureAuthenticator(loader, nil)

		_, err := auth.Authenticate(ctx, security.Authentication{
			Kind:      AuthKindSignature,
			Principal: "app1",
			Credentials: &security.SignatureCredentials{
				Timestamp: time.Now().Unix(),
				Nonce:     "abcdefghijklmnop",
				Signature: "signature",
			},
		})

		assert.Error(t, err, "Should return error for disabled app")
		loader.AssertExpectations(t)
	})

	t.Run("EnabledAppWithoutWhitelist", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		principal.Details = &security.ExternalAppConfig{
			Enabled:     true,
			IPWhitelist: "",
		}
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)
		nonceStore.On("Exists", mock.Anything, "app1", mock.AnythingOfType("string")).Return(false, nil)
		nonceStore.On("Store", mock.Anything, "app1", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

		auth := NewSignatureAuthenticator(loader, nonceStore)

		credentials := generateValidCredentials(t, "app1", testSecretHex)

		result, err := auth.Authenticate(ctx, security.Authentication{
			Kind:        AuthKindSignature,
			Principal:   "app1",
			Credentials: credentials,
		})

		require.NoError(t, err, "Should authenticate when app is enabled without whitelist")
		assert.Equal(t, "app1", result.ID, "Principal ID should match")
		loader.AssertExpectations(t)
		nonceStore.AssertExpectations(t)
	})

	t.Run("NoExternalAppConfig", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)
		nonceStore.On("Exists", mock.Anything, "app1", mock.AnythingOfType("string")).Return(false, nil)
		nonceStore.On("Store", mock.Anything, "app1", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

		auth := NewSignatureAuthenticator(loader, nonceStore)

		credentials := generateValidCredentials(t, "app1", testSecretHex)

		result, err := auth.Authenticate(ctx, security.Authentication{
			Kind:        AuthKindSignature,
			Principal:   "app1",
			Credentials: credentials,
		})

		require.NoError(t, err, "Should authenticate when no ExternalAppConfig is set")
		assert.Equal(t, "app1", result.ID, "Principal ID should match")
		loader.AssertExpectations(t)
		nonceStore.AssertExpectations(t)
	})
}

func TestSignatureAuthenticator_SuccessfulAuthentication(t *testing.T) {
	ctx := context.Background()

	t.Run("WithNonceStore", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)
		nonceStore.On("Exists", mock.Anything, "app1", mock.AnythingOfType("string")).Return(false, nil)
		nonceStore.On("Store", mock.Anything, "app1", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

		auth := NewSignatureAuthenticator(loader, nonceStore)

		credentials := generateValidCredentials(t, "app1", testSecretHex)

		result, err := auth.Authenticate(ctx, security.Authentication{
			Kind:        AuthKindSignature,
			Principal:   "app1",
			Credentials: credentials,
		})

		require.NoError(t, err, "Should authenticate successfully")
		assert.Equal(t, "app1", result.ID, "Principal ID should match")
		assert.Equal(t, "Test App", result.Name, "Principal name should match")
		assert.Equal(t, security.PrincipalTypeExternalApp, result.Type, "Principal type should be external app")

		loader.AssertExpectations(t)
		nonceStore.AssertExpectations(t)
	})

	t.Run("WithoutNonceStore", func(t *testing.T) {
		loader := new(MockExternalAppLoader)

		principal := security.NewExternalApp("app1", "Test App", "api_user")
		loader.On("LoadByID", mock.Anything, "app1").Return(principal, testSecretHex, nil)

		auth := NewSignatureAuthenticator(loader, nil)

		credentials := generateValidCredentials(t, "app1", testSecretHex)

		result, err := auth.Authenticate(ctx, security.Authentication{
			Kind:        AuthKindSignature,
			Principal:   "app1",
			Credentials: credentials,
		})

		require.NoError(t, err, "Should authenticate successfully without nonce store")
		assert.Equal(t, "app1", result.ID, "Principal ID should match")

		loader.AssertExpectations(t)
	})

	t.Run("MultipleApps", func(t *testing.T) {
		loader := new(MockExternalAppLoader)
		nonceStore := new(MockNonceStore)

		secret1 := testSecretHex
		secret2 := "bf7786789ce92be8d04d5b62e233ff72fa861ff6e53dfbc2d44c3a4a47cd25d3"

		principal1 := security.NewExternalApp("app1", "App One", "api_user")
		principal2 := security.NewExternalApp("app2", "App Two", "api_admin")

		loader.On("LoadByID", mock.Anything, "app1").Return(principal1, secret1, nil)
		loader.On("LoadByID", mock.Anything, "app2").Return(principal2, secret2, nil)
		nonceStore.On("Exists", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(false, nil)
		nonceStore.On("Store", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

		auth := NewSignatureAuthenticator(loader, nonceStore)

		// Authenticate app1
		creds1 := generateValidCredentials(t, "app1", secret1)
		result1, err := auth.Authenticate(ctx, security.Authentication{
			Kind:        AuthKindSignature,
			Principal:   "app1",
			Credentials: creds1,
		})
		require.NoError(t, err, "Should authenticate app1")
		assert.Equal(t, "app1", result1.ID)
		assert.Equal(t, "App One", result1.Name)

		// Authenticate app2
		creds2 := generateValidCredentials(t, "app2", secret2)
		result2, err := auth.Authenticate(ctx, security.Authentication{
			Kind:        AuthKindSignature,
			Principal:   "app2",
			Credentials: creds2,
		})
		require.NoError(t, err, "Should authenticate app2")
		assert.Equal(t, "app2", result2.ID)
		assert.Equal(t, "App Two", result2.Name)

		loader.AssertExpectations(t)
	})
}
