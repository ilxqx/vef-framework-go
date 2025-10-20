package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlphanumUs(t *testing.T) {
	type TestStruct struct {
		Value string `validate:"alphanum_us"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Valid cases
		{"lowercase letters", "abc", false},
		{"uppercase letters", "ABC", false},
		{"mixed case", "AbCdEf", false},
		{"numbers", "123", false},
		{"alphanumeric", "abc123", false},
		{"with underscores", "abc_123", false},
		{"multiple underscores", "abc__123__def", false},
		{"leading underscore", "_abc", false},
		{"trailing underscore", "abc_", false},
		{"only underscores", "___", false},
		{"snake_case", "get_user_info", false},

		// Invalid cases
		{"with space", "abc 123", true},
		{"with dash", "abc-123", true},
		{"with slash", "abc/123", true},
		{"with dot", "abc.123", true},
		{"with special chars", "abc@123", true},
		{"empty string", "", true},
		{"chinese characters", "中文", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{Value: tt.value}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err, "Expected validation error for value: %q", tt.value)
			} else {
				assert.NoError(t, err, "Expected no validation error for value: %q", tt.value)
			}
		})
	}
}

func TestAlphanumUsSlash(t *testing.T) {
	type TestStruct struct {
		Value string `validate:"alphanum_us_slash"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Valid cases
		{"lowercase letters", "abc", false},
		{"uppercase letters", "ABC", false},
		{"numbers", "123", false},
		{"alphanumeric", "abc123", false},
		{"with underscores", "abc_123", false},
		{"with slashes", "abc/123", false},
		{"resource path", "sys/user", false},
		{"nested path", "auth/get_user_info", false},
		{"multiple slashes", "a/b/c/d", false},
		{"leading slash", "/abc", false},
		{"trailing slash", "abc/", false},
		{"mixed", "sys_module/user_info", false},

		// Invalid cases
		{"with space", "abc 123", true},
		{"with dash", "abc-123", true},
		{"with dot", "abc.123", true},
		{"with special chars", "abc@123", true},
		{"empty string", "", true},
		{"chinese characters", "中文", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{Value: tt.value}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err, "Expected validation error for value: %q", tt.value)
			} else {
				assert.NoError(t, err, "Expected no validation error for value: %q", tt.value)
			}
		})
	}
}

func TestAlphanumUsDot(t *testing.T) {
	type TestStruct struct {
		Value string `validate:"alphanum_us_dot"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Valid cases
		{"lowercase letters", "abc", false},
		{"uppercase letters", "ABC", false},
		{"numbers", "123", false},
		{"alphanumeric", "abc123", false},
		{"with underscores", "abc_123", false},
		{"with dots", "abc.123", false},
		{"file name", "config.yaml", false},
		{"module name", "com.example.app", false},
		{"version number", "v1.2.3", false},
		{"multiple dots", "a.b.c.d", false},
		{"leading dot", ".hidden", false},
		{"trailing dot", "file.", false},
		{"mixed", "app_config.prod.yaml", false},

		// Invalid cases
		{"with space", "abc 123", true},
		{"with dash", "abc-123", true},
		{"with slash", "abc/123", true},
		{"with special chars", "abc@123", true},
		{"empty string", "", true},
		{"chinese characters", "中文", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{Value: tt.value}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err, "Expected validation error for value: %q", tt.value)
			} else {
				assert.NoError(t, err, "Expected no validation error for value: %q", tt.value)
			}
		})
	}
}

func TestAlphanumRulesCombined(t *testing.T) {
	type TestStruct struct {
		Action   string `validate:"alphanum_us" label:"操作"`
		Resource string `validate:"alphanum_us_slash" label:"资源"`
		FileName string `validate:"alphanum_us_dot" label:"文件名"`
	}

	tests := []struct {
		name     string
		action   string
		resource string
		fileName string
		wantErr  bool
	}{
		{
			name:     "all valid",
			action:   "get_user_info",
			resource: "sys/user",
			fileName: "config.yaml",
			wantErr:  false,
		},
		{
			name:     "invalid action with slash",
			action:   "get/user",
			resource: "sys/user",
			fileName: "config.yaml",
			wantErr:  true,
		},
		{
			name:     "invalid resource with dot",
			action:   "get_user",
			resource: "sys.user",
			fileName: "config.yaml",
			wantErr:  true,
		},
		{
			name:     "invalid filename with slash",
			action:   "get_user",
			resource: "sys/user",
			fileName: "config/app.yaml",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := TestStruct{
				Action:   tt.action,
				Resource: tt.resource,
				FileName: tt.fileName,
			}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
