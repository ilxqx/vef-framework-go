package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithVersion_ValidFormats tests version validation with valid format inputs.
func TestWithVersion_ValidFormats(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"SingleDigit", "v1"},
		{"TwoDigits", "v10"},
		{"ThreeDigits", "v100"},
		{"LargeVersion", "v999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := NewResource("test_resource", WithVersion(tt.version))
			assert.Equal(t, tt.version, resource.Version(), "Version should match the provided value")
		})
	}
}

// TestWithVersion_InvalidFormats tests version validation rejects invalid format inputs.
func TestWithVersion_InvalidFormats(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"NoVPrefix", "1"},
		{"UppercaseV", "V1"},
		{"WithDot", "v1.0"},
		{"WithDash", "v1-beta"},
		{"WithLetters", "v1a"},
		{"OnlyV", "v"},
		{"EmptyString", ""},
		{"VersionWord", "version1"},
		{"Decimal", "v1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				NewResource("test_resource", WithVersion(tt.version))
			}, "Should panic for invalid version format")
		})
	}
}

// TestNewResource_DefaultVersion tests default version assignment.
func TestNewResource_DefaultVersion(t *testing.T) {
	resource := NewResource("test_resource")
	assert.Equal(t, VersionV1, resource.Version(), "Default version should be v1")
}

// TestNewResource_ValidNames tests resource name validation with valid inputs.
func TestNewResource_ValidNames(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
	}{
		{"SimpleName", "user"},
		{"WithUnderscore", "user_info"},
		{"MultipleUnderscores", "get_user_info"},
		{"WithNumber", "user2"},
		{"WithNamespace", "sys/user"},
		{"NestedNamespace", "sys/auth/user"},
		{"NamespaceWithUnderscore", "sys/user_info"},
		{"Complex", "auth/get_user_info"},
		{"NumberInSegment", "api2/user_info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := NewResource(tt.resourceName)
			assert.Equal(t, tt.resourceName, resource.Name(), "Resource name should match the provided value")
		})
	}
}

// TestNewResource_InvalidNames tests resource name validation rejects invalid inputs.
func TestNewResource_InvalidNames(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string
	}{
		{"Empty", ""},
		{"UppercaseStart", "User"},
		{"CamelCase", "getUser"},
		{"PascalCase", "GetUser"},
		{"WithDash", "user-info"},
		{"WithSpace", "user info"},
		{"LeadingSlash", "/user"},
		{"TrailingSlash", "user/"},
		{"TrailingSlashWithNamespace", "sys/user/"},
		{"ConsecutiveSlashes", "sys//user"},
		{"UppercaseInNamespace", "Sys/user"},
		{"CamelCaseInNamespace", "sys/getUser"},
		{"LeadingUnderscore", "_user"},
		{"DoubleUnderscore", "user__info"},
		{"OnlySlash", "/"},
		{"OnlyNamespace", "sys/"},
		{"StartsWithNumber", "1user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				NewResource(tt.resourceName)
			}, "Should panic for invalid resource name format")
		})
	}
}

// TestWithApis_ValidActionNames tests action name validation with valid inputs.
func TestWithApis_ValidActionNames(t *testing.T) {
	tests := []struct {
		name   string
		action string
	}{
		{"SimpleAction", "create"},
		{"WithUnderscore", "find_page"},
		{"MultipleUnderscores", "get_user_info"},
		{"WithNumber", "create2"},
		{"LongAction", "find_all_active_users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := Spec{Action: tt.action}
			resource := NewResource("test_resource", WithApis(spec))
			require.Len(t, resource.Apis(), 1, "Should have exactly one API spec")
			assert.Equal(t, tt.action, resource.Apis()[0].Action, "Action name should match the provided value")
		})
	}
}

// TestWithApis_InvalidActionNames tests action name validation rejects invalid inputs.
func TestWithApis_InvalidActionNames(t *testing.T) {
	tests := []struct {
		name   string
		action string
	}{
		{"Empty", ""},
		{"UppercaseStart", "Create"},
		{"CamelCase", "findPage"},
		{"PascalCase", "FindPage"},
		{"WithDash", "find-page"},
		{"WithSpace", "find page"},
		{"WithDot", "find.page"},
		{"WithSlash", "find/page"},
		{"LeadingUnderscore", "_create"},
		{"DoubleUnderscore", "find__page"},
		{"TrailingUnderscore", "create_"},
		{"StartsWithNumber", "1create"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				spec := Spec{Action: tt.action}
				NewResource("test_resource", WithApis(spec))
			}, "Should panic for invalid action name format")
		})
	}
}

// TestWithApis_MultipleSpecs tests registration of multiple API specs.
func TestWithApis_MultipleSpecs(t *testing.T) {
	specs := []Spec{
		{Action: "create"},
		{Action: "find_page"},
		{Action: "update"},
	}

	resource := NewResource("test_resource", WithApis(specs...))
	assert.Len(t, resource.Apis(), 3, "Should have three API specs registered")
}

// TestWithApis_MultipleSpecs_OneInvalid tests that one invalid spec in a batch causes panic.
func TestWithApis_MultipleSpecs_OneInvalid(t *testing.T) {
	assert.Panics(t, func() {
		specs := []Spec{
			{Action: "create"},
			{Action: "findPage"}, // Invalid camelCase
			{Action: "update"},
		}

		NewResource("test_resource", WithApis(specs...))
	}, "Should panic when any spec has an invalid action name")
}
