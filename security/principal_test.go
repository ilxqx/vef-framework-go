package security

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilxqx/vef-framework-go/constants"
)

type TestUserDetails struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	Age         int    `json:"age"`
}

type TestExternalAppDetails struct {
	AppId     string   `json:"appId"`
	AppSecret string   `json:"appSecret"`
	Scopes    []string `json:"scopes"`
}

func TestNewUser(t *testing.T) {
	t.Run("Create user without roles", func(t *testing.T) {
		user := NewUser("user123", "John Doe")
		assert.Equal(t, PrincipalTypeUser, user.Type)
		assert.Equal(t, "user123", user.Id)
		assert.Equal(t, "John Doe", user.Name)
		assert.Empty(t, user.Roles)
		assert.Nil(t, user.Details)
	})

	t.Run("Create user with roles", func(t *testing.T) {
		user := NewUser("user456", "Jane Smith", "admin", "editor")
		assert.Equal(t, PrincipalTypeUser, user.Type)
		assert.Equal(t, "user456", user.Id)
		assert.Equal(t, "Jane Smith", user.Name)
		assert.Equal(t, []string{"admin", "editor"}, user.Roles)
	})
}

func TestNewExternalApp(t *testing.T) {
	t.Run("Create external app without roles", func(t *testing.T) {
		app := NewExternalApp("app123", "Payment Service")
		assert.Equal(t, PrincipalTypeExternalApp, app.Type)
		assert.Equal(t, "app123", app.Id)
		assert.Equal(t, "Payment Service", app.Name)
		assert.Empty(t, app.Roles)
		assert.Nil(t, app.Details)
	})

	t.Run("Create external app with roles", func(t *testing.T) {
		app := NewExternalApp("app456", "Auth Service", "service", "trusted")
		assert.Equal(t, PrincipalTypeExternalApp, app.Type)
		assert.Equal(t, "app456", app.Id)
		assert.Equal(t, "Auth Service", app.Name)
		assert.Equal(t, []string{"service", "trusted"}, app.Roles)
	})
}

func TestPrincipalWithRoles(t *testing.T) {
	t.Run("Add roles to principal", func(t *testing.T) {
		user := NewUser("user123", "Test User")
		user.WithRoles("admin", "moderator")
		assert.Equal(t, []string{"admin", "moderator"}, user.Roles)
	})

	t.Run("Add roles multiple times", func(t *testing.T) {
		user := NewUser("user123", "Test User", "viewer")
		user.WithRoles("admin").WithRoles("editor")
		assert.Equal(t, []string{"viewer", "admin", "editor"}, user.Roles)
	})
}

func TestPrincipalSystem(t *testing.T) {
	t.Run("System principal has correct values", func(t *testing.T) {
		assert.Equal(t, PrincipalTypeSystem, PrincipalSystem.Type)
		assert.Equal(t, constants.OperatorSystem, PrincipalSystem.Id)
		assert.Equal(t, "系统", PrincipalSystem.Name)
	})
}

func TestPrincipalAnonymous(t *testing.T) {
	t.Run("Anonymous principal has correct values", func(t *testing.T) {
		assert.Equal(t, PrincipalTypeUser, PrincipalAnonymous.Type)
		assert.Equal(t, constants.OperatorAnonymous, PrincipalAnonymous.Id)
		assert.Equal(t, "匿名", PrincipalAnonymous.Name)
	})
}

func TestPrincipalJSONMarshal(t *testing.T) {
	t.Run("Marshal user with map details", func(t *testing.T) {
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

	t.Run("Marshal user without details", func(t *testing.T) {
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
	t.Run("Unmarshal user with map details", func(t *testing.T) {
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
		require.NoError(t, err, "Should unmarshal user with map details")

		assert.Equal(t, PrincipalTypeUser, principal.Type)
		assert.Equal(t, "user123", principal.Id)
		assert.Equal(t, "Test User", principal.Name)
		assert.Equal(t, []string{"admin", "editor"}, principal.Roles)

		detailsPtr, ok := principal.Details.(*map[string]any)
		require.True(t, ok, "Details should be a map")

		details := *detailsPtr
		assert.Equal(t, "test@example.com", details["email"])
		assert.Equal(t, float64(30), details["age"])
	})

	t.Run("Unmarshal user with struct details", func(t *testing.T) {
		originalType := userDetailsType
		defer func() { userDetailsType = originalType }()

		SetUserDetailsType[TestUserDetails]()

		jsonData := `{
			"type": "user",
			"id": "user456",
			"name": "Jane Doe",
			"roles": ["viewer"],
			"details": {
				"email": "jane@example.com",
				"phoneNumber": "+1234567890",
				"age": 25
			}
		}`

		var principal Principal

		err := json.Unmarshal([]byte(jsonData), &principal)
		require.NoError(t, err, "Should unmarshal user with struct details")

		details, ok := principal.Details.(*TestUserDetails)
		require.True(t, ok, "Details should be TestUserDetails")
		assert.Equal(t, "jane@example.com", details.Email)
		assert.Equal(t, "+1234567890", details.PhoneNumber)
		assert.Equal(t, 25, details.Age)
	})

	t.Run("Unmarshal external app with struct details", func(t *testing.T) {
		originalType := externalAppDetailsType
		defer func() { externalAppDetailsType = originalType }()

		SetExternalAppDetailsType[TestExternalAppDetails]()

		jsonData := `{
			"type": "external_app",
			"id": "app123",
			"name": "Auth Service",
			"roles": ["service"],
			"details": {
				"appId": "app_123456",
				"appSecret": "secret_abc",
				"scopes": ["read", "write"]
			}
		}`

		var principal Principal

		err := json.Unmarshal([]byte(jsonData), &principal)
		require.NoError(t, err, "Should unmarshal external app with struct details")

		details, ok := principal.Details.(*TestExternalAppDetails)
		require.True(t, ok, "Details should be TestExternalAppDetails")
		assert.Equal(t, "app_123456", details.AppId)
		assert.Equal(t, "secret_abc", details.AppSecret)
		assert.Equal(t, []string{"read", "write"}, details.Scopes)
	})

	t.Run("Unmarshal system principal", func(t *testing.T) {
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

	t.Run("Unmarshal invalid json", func(t *testing.T) {
		jsonData := `{invalid json}`

		var principal Principal

		err := json.Unmarshal([]byte(jsonData), &principal)
		assert.Error(t, err)
	})
}

func TestAttemptUnmarshalDetails(t *testing.T) {
	t.Run("Unmarshal user details from map", func(t *testing.T) {
		originalType := userDetailsType
		defer func() { userDetailsType = originalType }() // Reset at end

		SetUserDetailsType[TestUserDetails]()

		user := NewUser("user123", "Test User")
		detailsMap := map[string]any{
			"email":       "test@example.com",
			"phoneNumber": "+1234567890",
			"age":         30,
		}

		user.AttemptUnmarshalDetails(detailsMap)

		details, ok := user.Details.(*TestUserDetails)
		require.True(t, ok)
		assert.Equal(t, "test@example.com", details.Email)
		assert.Equal(t, "+1234567890", details.PhoneNumber)
		assert.Equal(t, 30, details.Age)
	})

	t.Run("Unmarshal external app details from map", func(t *testing.T) {
		originalType := externalAppDetailsType
		defer func() { externalAppDetailsType = originalType }()

		SetExternalAppDetailsType[TestExternalAppDetails]()

		app := NewExternalApp("app123", "Test App")
		detailsMap := map[string]any{
			"appId":     "app_123",
			"appSecret": "secret",
			"scopes":    []any{"read", "write"},
		}

		app.AttemptUnmarshalDetails(detailsMap)

		details, ok := app.Details.(*TestExternalAppDetails)
		require.True(t, ok, "Details should be TestExternalAppDetails")
		assert.Equal(t, "app_123", details.AppId)
		assert.Equal(t, "secret", details.AppSecret)
	})

	t.Run("Details type is map, keep as is", func(t *testing.T) {
		user := NewUser("user123", "Test User")
		detailsMap := map[string]any{
			"key": "value",
		}

		user.AttemptUnmarshalDetails(detailsMap)
		assert.Equal(t, detailsMap, user.Details, "Details should remain as map")
	})

	t.Run("Non-map details for user type", func(t *testing.T) {
		user := NewUser("user123", "Test User")
		stringDetails := "string details"

		user.AttemptUnmarshalDetails(stringDetails)
		assert.Equal(t, stringDetails, user.Details)
	})

	t.Run("System principal keeps details as is", func(t *testing.T) {
		principal := &Principal{
			Type: PrincipalTypeSystem,
			Id:   "system",
			Name: "System",
		}

		details := map[string]any{"key": "value"}
		principal.AttemptUnmarshalDetails(details)
		assert.Equal(t, details, principal.Details)
	})

	t.Run("Decode with partial fields creates struct", func(t *testing.T) {
		originalType := userDetailsType
		defer func() { userDetailsType = originalType }()

		SetUserDetailsType[TestUserDetails]()

		user := NewUser("user123", "Test User")
		partialDetails := map[string]any{
			"email":        "test@example.com",
			"invalidField": "value",
		}

		user.AttemptUnmarshalDetails(partialDetails)
		details, ok := user.Details.(*TestUserDetails)
		require.True(t, ok, "Details should be TestUserDetails")
		assert.Equal(t, "test@example.com", details.Email)
		assert.Equal(t, "", details.PhoneNumber, "Unset field should have zero value")
	})
}

func TestSetUserDetailsType(t *testing.T) {
	t.Run("Set valid struct type", func(t *testing.T) {
		originalType := userDetailsType
		defer func() { userDetailsType = originalType }()

		SetUserDetailsType[TestUserDetails]()
		assert.Equal(t, "TestUserDetails", userDetailsType.Name(), "Type name should be TestUserDetails")
	})

	t.Run("Panic on non-struct type", func(t *testing.T) {
		assert.Panics(t, func() {
			SetUserDetailsType[string]()
		}, "Should panic on non-struct type")
	})
}

func TestSetExternalAppDetailsType(t *testing.T) {
	t.Run("Set valid struct type", func(t *testing.T) {
		originalType := externalAppDetailsType
		defer func() { externalAppDetailsType = originalType }()

		SetExternalAppDetailsType[TestExternalAppDetails]()
		assert.Equal(t, "TestExternalAppDetails", externalAppDetailsType.Name(), "Type name should be TestExternalAppDetails")
	})

	t.Run("Panic on non-struct type", func(t *testing.T) {
		assert.Panics(t, func() {
			SetExternalAppDetailsType[int]()
		}, "Should panic on non-struct type")
	})
}

func TestPrincipalRoundTrip(t *testing.T) {
	t.Run("Marshal and unmarshal user", func(t *testing.T) {
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

	t.Run("Marshal and unmarshal external app", func(t *testing.T) {
		original := NewExternalApp("app123", "Auth Service", "service")
		original.Details = map[string]any{
			"appId":  "123",
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
