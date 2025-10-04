package security

import (
	"encoding/json"
	"testing"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test user details struct
type TestUserDetails struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Age         int    `json:"age"`
}

// Test external app details struct
type TestExternalAppDetails struct {
	AppID     string   `json:"app_id"`
	AppSecret string   `json:"app_secret"`
	Scopes    []string `json:"scopes"`
}

func TestNewUser(t *testing.T) {
	t.Run("create user without roles", func(t *testing.T) {
		user := NewUser("user123", "John Doe")
		assert.Equal(t, PrincipalTypeUser, user.Type)
		assert.Equal(t, "user123", user.Id)
		assert.Equal(t, "John Doe", user.Name)
		assert.Empty(t, user.Roles)
		assert.Nil(t, user.Details)
	})

	t.Run("create user with roles", func(t *testing.T) {
		user := NewUser("user456", "Jane Smith", "admin", "editor")
		assert.Equal(t, PrincipalTypeUser, user.Type)
		assert.Equal(t, "user456", user.Id)
		assert.Equal(t, "Jane Smith", user.Name)
		assert.Equal(t, []string{"admin", "editor"}, user.Roles)
	})
}

func TestNewExternalApp(t *testing.T) {
	t.Run("create external app without roles", func(t *testing.T) {
		app := NewExternalApp("app123", "Payment Service")
		assert.Equal(t, PrincipalTypeExternalApp, app.Type)
		assert.Equal(t, "app123", app.Id)
		assert.Equal(t, "Payment Service", app.Name)
		assert.Empty(t, app.Roles)
		assert.Nil(t, app.Details)
	})

	t.Run("create external app with roles", func(t *testing.T) {
		app := NewExternalApp("app456", "Auth Service", "service", "trusted")
		assert.Equal(t, PrincipalTypeExternalApp, app.Type)
		assert.Equal(t, "app456", app.Id)
		assert.Equal(t, "Auth Service", app.Name)
		assert.Equal(t, []string{"service", "trusted"}, app.Roles)
	})
}

func TestPrincipalWithRoles(t *testing.T) {
	t.Run("add roles to principal", func(t *testing.T) {
		user := NewUser("user123", "Test User")
		user.WithRoles("admin", "moderator")
		assert.Equal(t, []string{"admin", "moderator"}, user.Roles)
	})

	t.Run("add roles multiple times", func(t *testing.T) {
		user := NewUser("user123", "Test User", "viewer")
		user.WithRoles("admin").WithRoles("editor")
		assert.Equal(t, []string{"viewer", "admin", "editor"}, user.Roles)
	})
}

func TestPrincipalSystem(t *testing.T) {
	t.Run("system principal has correct values", func(t *testing.T) {
		assert.Equal(t, PrincipalTypeSystem, PrincipalSystem.Type)
		assert.Equal(t, constants.OperatorSystem, PrincipalSystem.Id)
		assert.Equal(t, "系统", PrincipalSystem.Name)
	})
}

func TestPrincipalAnonymous(t *testing.T) {
	t.Run("anonymous principal has correct values", func(t *testing.T) {
		assert.Equal(t, PrincipalTypeUser, PrincipalAnonymous.Type)
		assert.Equal(t, constants.OperatorAnonymous, PrincipalAnonymous.Id)
		assert.Equal(t, "匿名", PrincipalAnonymous.Name)
	})
}

func TestPrincipalJSONMarshal(t *testing.T) {
	t.Run("marshal user with map details", func(t *testing.T) {
		user := NewUser("user123", "Test User", "admin")
		user.Details = map[string]any{
			"email": "test@example.com",
			"age":   30,
		}

		data, err := json.Marshal(user)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "user", result["type"])
		assert.Equal(t, "user123", result["id"])
		assert.Equal(t, "Test User", result["name"])
		assert.Contains(t, result, "details")
	})

	t.Run("marshal user without details", func(t *testing.T) {
		user := NewUser("user123", "Test User")

		data, err := json.Marshal(user)
		require.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "user", result["type"])
		assert.Nil(t, result["details"])
	})
}

func TestPrincipalJSONUnmarshal(t *testing.T) {
	t.Run("unmarshal user with map details", func(t *testing.T) {
		jsonData := `{
			"type": "user",
			"id": "user123",
			"name": "Test User",
			"roles": ["admin", "editor"],
			"details": {
				"email": "test@example.com",
				"age": 30
			}
		}`

		var principal Principal
		err := json.Unmarshal([]byte(jsonData), &principal)
		require.NoError(t, err)

		assert.Equal(t, PrincipalTypeUser, principal.Type)
		assert.Equal(t, "user123", principal.Id)
		assert.Equal(t, "Test User", principal.Name)
		assert.Equal(t, []string{"admin", "editor"}, principal.Roles)

		// Details will be a pointer to map[string]any when userDetailsType is map[string]any
		detailsPtr, ok := principal.Details.(*map[string]any)
		require.True(t, ok)
		details := *detailsPtr
		assert.Equal(t, "test@example.com", details["email"])
		assert.Equal(t, float64(30), details["age"])
	})

	t.Run("unmarshal user with struct details", func(t *testing.T) {
		// Set user details type
		originalType := userDetailsType
		defer func() { userDetailsType = originalType }() // Reset at end
		SetUserDetailsType[TestUserDetails]()

		jsonData := `{
			"type": "user",
			"id": "user456",
			"name": "Jane Doe",
			"roles": ["viewer"],
			"details": {
				"email": "jane@example.com",
				"phone_number": "+1234567890",
				"age": 25
			}
		}`

		var principal Principal
		err := json.Unmarshal([]byte(jsonData), &principal)
		require.NoError(t, err)

		details, ok := principal.Details.(*TestUserDetails)
		require.True(t, ok)
		assert.Equal(t, "jane@example.com", details.Email)
		assert.Equal(t, "+1234567890", details.PhoneNumber)
		assert.Equal(t, 25, details.Age)
	})

	t.Run("unmarshal external app with struct details", func(t *testing.T) {
		// Set external app details type
		originalType := externalAppDetailsType
		defer func() { externalAppDetailsType = originalType }() // Reset at end
		SetExternalAppDetailsType[TestExternalAppDetails]()

		jsonData := `{
			"type": "external_app",
			"id": "app123",
			"name": "Auth Service",
			"roles": ["service"],
			"details": {
				"app_id": "app_123456",
				"app_secret": "secret_abc",
				"scopes": ["read", "write"]
			}
		}`

		var principal Principal
		err := json.Unmarshal([]byte(jsonData), &principal)
		require.NoError(t, err)

		details, ok := principal.Details.(*TestExternalAppDetails)
		require.True(t, ok)
		assert.Equal(t, "app_123456", details.AppID)
		assert.Equal(t, "secret_abc", details.AppSecret)
		assert.Equal(t, []string{"read", "write"}, details.Scopes)
	})

	t.Run("unmarshal system principal", func(t *testing.T) {
		jsonData := `{
			"type": "system",
			"id": "system",
			"name": "系统",
			"details": null
		}`

		var principal Principal
		err := json.Unmarshal([]byte(jsonData), &principal)
		require.NoError(t, err)

		assert.Equal(t, PrincipalTypeSystem, principal.Type)
		assert.Equal(t, constants.OperatorSystem, principal.Id)
		assert.Nil(t, principal.Details)
	})

	t.Run("unmarshal invalid json", func(t *testing.T) {
		jsonData := `{invalid json}`

		var principal Principal
		err := json.Unmarshal([]byte(jsonData), &principal)
		assert.Error(t, err)
	})
}

func TestAttemptUnmarshalDetails(t *testing.T) {
	t.Run("unmarshal user details from map", func(t *testing.T) {
		originalType := userDetailsType
		defer func() { userDetailsType = originalType }() // Reset at end
		SetUserDetailsType[TestUserDetails]()

		user := NewUser("user123", "Test User")
		detailsMap := map[string]any{
			"email":        "test@example.com",
			"phone_number": "+1234567890",
			"age":          30,
		}

		user.AttemptUnmarshalDetails(detailsMap)

		details, ok := user.Details.(*TestUserDetails)
		require.True(t, ok)
		assert.Equal(t, "test@example.com", details.Email)
		assert.Equal(t, "+1234567890", details.PhoneNumber)
		assert.Equal(t, 30, details.Age)
	})

	t.Run("unmarshal external app details from map", func(t *testing.T) {
		originalType := externalAppDetailsType
		defer func() { externalAppDetailsType = originalType }() // Reset at end
		SetExternalAppDetailsType[TestExternalAppDetails]()

		app := NewExternalApp("app123", "Test App")
		detailsMap := map[string]any{
			"app_id":     "app_123",
			"app_secret": "secret",
			"scopes":     []any{"read", "write"},
		}

		app.AttemptUnmarshalDetails(detailsMap)

		details, ok := app.Details.(*TestExternalAppDetails)
		require.True(t, ok)
		assert.Equal(t, "app_123", details.AppID)
		assert.Equal(t, "secret", details.AppSecret)
	})

	t.Run("details type is map, keep as is", func(t *testing.T) {
		// Default type is already map[string]any, no need to set
		user := NewUser("user123", "Test User")
		detailsMap := map[string]any{
			"key": "value",
		}

		user.AttemptUnmarshalDetails(detailsMap)
		// When type is map[string]any, it's kept as-is
		assert.Equal(t, detailsMap, user.Details)
	})

	t.Run("non-map details for user type", func(t *testing.T) {
		user := NewUser("user123", "Test User")
		stringDetails := "string details"

		user.AttemptUnmarshalDetails(stringDetails)
		assert.Equal(t, stringDetails, user.Details)
	})

	t.Run("system principal keeps details as is", func(t *testing.T) {
		principal := &Principal{
			Type: PrincipalTypeSystem,
			Id:   "system",
			Name: "System",
		}

		details := map[string]any{"key": "value"}
		principal.AttemptUnmarshalDetails(details)
		assert.Equal(t, details, principal.Details)
	})

	t.Run("decode with partial fields creates struct", func(t *testing.T) {
		originalType := userDetailsType
		defer func() { userDetailsType = originalType }() // Reset at end
		SetUserDetailsType[TestUserDetails]()

		user := NewUser("user123", "Test User")
		// Details with only some fields that match TestUserDetails
		partialDetails := map[string]any{
			"email":         "test@example.com",
			"invalid_field": "value", // This field will be ignored
		}

		user.AttemptUnmarshalDetails(partialDetails)
		// Decoder succeeds and creates struct with matched fields
		details, ok := user.Details.(*TestUserDetails)
		require.True(t, ok)
		assert.Equal(t, "test@example.com", details.Email)
		assert.Equal(t, "", details.PhoneNumber) // Unset field has zero value
	})
}

func TestSetUserDetailsType(t *testing.T) {
	t.Run("set valid struct type", func(t *testing.T) {
		originalType := userDetailsType
		defer func() { userDetailsType = originalType }() // Reset at end
		SetUserDetailsType[TestUserDetails]()
		assert.Equal(t, "TestUserDetails", userDetailsType.Name())
	})

	t.Run("panic on non-struct type", func(t *testing.T) {
		assert.Panics(t, func() {
			SetUserDetailsType[string]()
		})
	})
}

func TestSetExternalAppDetailsType(t *testing.T) {
	t.Run("set valid struct type", func(t *testing.T) {
		originalType := externalAppDetailsType
		defer func() { externalAppDetailsType = originalType }() // Reset at end
		SetExternalAppDetailsType[TestExternalAppDetails]()
		assert.Equal(t, "TestExternalAppDetails", externalAppDetailsType.Name())
	})

	t.Run("panic on non-struct type", func(t *testing.T) {
		assert.Panics(t, func() {
			SetExternalAppDetailsType[int]()
		})
	})
}

func TestPrincipalRoundTrip(t *testing.T) {
	t.Run("marshal and unmarshal user", func(t *testing.T) {
		original := NewUser("user123", "Test User", "admin", "editor")
		original.Details = map[string]any{
			"email": "test@example.com",
			"age":   30,
		}

		// Marshal
		data, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal
		var restored Principal
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		assert.Equal(t, original.Type, restored.Type)
		assert.Equal(t, original.Id, restored.Id)
		assert.Equal(t, original.Name, restored.Name)
		assert.Equal(t, original.Roles, restored.Roles)
	})

	t.Run("marshal and unmarshal external app", func(t *testing.T) {
		original := NewExternalApp("app123", "Auth Service", "service")
		original.Details = map[string]any{
			"app_id": "123",
			"scopes": []string{"read", "write"},
		}

		// Marshal
		data, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal
		var restored Principal
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		assert.Equal(t, original.Type, restored.Type)
		assert.Equal(t, original.Id, restored.Id)
		assert.Equal(t, original.Name, restored.Name)
		assert.Equal(t, original.Roles, restored.Roles)
	})
}
