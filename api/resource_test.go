package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Version Tests ====================

func TestWithVersion_ValidFormats(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"single digit", "v1"},
		{"two digits", "v10"},
		{"three digits", "v100"},
		{"large version", "v999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := NewResource("test_resource", WithVersion(tt.version))
			assert.Equal(t, tt.version, resource.Version())
		})
	}
}

func TestWithVersion_InvalidFormats(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"no v prefix", "1"},
		{"uppercase V", "V1"},
		{"with dot", "v1.0"},
		{"with dash", "v1-beta"},
		{"with letters", "v1a"},
		{"only v", "v"},
		{"empty string", ""},
		{"version word", "version1"},
		{"decimal", "v1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				NewResource("test_resource", WithVersion(tt.version))
			})
		})
	}
}

func TestNewResource_DefaultVersion(t *testing.T) {
	resource := NewResource("test_resource")
	assert.Equal(t, VersionV1, resource.Version())
}

// ==================== Resource Name Tests ====================

func TestNewResource_ValidNames(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
	}{
		{"simple name", "user"},
		{"with underscore", "user_info"},
		{"multiple underscores", "get_user_info"},
		{"with number", "user2"},
		{"with namespace", "sys/user"},
		{"nested namespace", "sys/auth/user"},
		{"namespace with underscore", "sys/user_info"},
		{"complex", "auth/get_user_info"},
		{"number in segment", "api2/user_info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := NewResource(tt.resourceName)
			assert.Equal(t, tt.resourceName, resource.Name())
		})
	}
}

func TestNewResource_InvalidNames(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
	}{
		{"empty", ""},
		{"uppercase start", "User"},
		{"camelCase", "getUser"},
		{"PascalCase", "GetUser"},
		{"with dash", "user-info"},
		{"with space", "user info"},
		{"leading slash", "/user"},
		{"trailing slash", "user/"},
		{"trailing slash with namespace", "sys/user/"},
		{"consecutive slashes", "sys//user"},
		{"uppercase in namespace", "Sys/user"},
		{"camelCase in namespace", "sys/getUser"},
		{"leading underscore", "_user"},
		{"double underscore", "user__info"},
		{"only slash", "/"},
		{"only namespace", "sys/"},
		{"starts with number", "1user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				NewResource(tt.resourceName)
			})
		})
	}
}

// ==================== Action Name Tests ====================

func TestWithApis_ValidActionNames(t *testing.T) {
	tests := []struct {
		name   string
		action string
	}{
		{"simple action", "create"},
		{"with underscore", "find_page"},
		{"multiple underscores", "get_user_info"},
		{"with number", "create2"},
		{"long action", "find_all_active_users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := Spec{Action: tt.action}
			resource := NewResource("test_resource", WithApis(spec))
			require.Len(t, resource.Apis(), 1)
			assert.Equal(t, tt.action, resource.Apis()[0].Action)
		})
	}
}

func TestWithApis_InvalidActionNames(t *testing.T) {
	tests := []struct {
		name   string
		action string
	}{
		{"empty", ""},
		{"uppercase start", "Create"},
		{"camelCase", "findPage"},
		{"PascalCase", "FindPage"},
		{"with dash", "find-page"},
		{"with space", "find page"},
		{"with dot", "find.page"},
		{"with slash", "find/page"},
		{"leading underscore", "_create"},
		{"double underscore", "find__page"},
		{"trailing underscore", "create_"},
		{"starts with number", "1create"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				spec := Spec{Action: tt.action}
				NewResource("test_resource", WithApis(spec))
			})
		})
	}
}

func TestWithApis_MultipleSpecs(t *testing.T) {
	specs := []Spec{
		{Action: "create"},
		{Action: "find_page"},
		{Action: "update"},
	}

	resource := NewResource("test_resource", WithApis(specs...))
	assert.Len(t, resource.Apis(), 3)
}

func TestWithApis_MultipleSpecs_OneInvalid(t *testing.T) {
	assert.Panics(t, func() {
		specs := []Spec{
			{Action: "create"},
			{Action: "findPage"}, // Invalid camelCase
			{Action: "update"},
		}

		NewResource("test_resource", WithApis(specs...))
	})
}
