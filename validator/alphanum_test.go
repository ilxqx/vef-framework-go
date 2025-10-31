package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlphanumUs(t *testing.T) {
	type testStruct struct {
		Value string `validate:"alphanum_us"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"lowercaseLetters", "abc", false},
		{"uppercaseLetters", "ABC", false},
		{"mixedCase", "AbCdEf", false},
		{"numbers", "123", false},
		{"alphanumeric", "abc123", false},
		{"withUnderscores", "abc_123", false},
		{"multipleUnderscores", "abc__123__def", false},
		{"leadingUnderscore", "_abc", false},
		{"trailingUnderscore", "abc_", false},
		{"onlyUnderscores", "___", false},
		{"snakeCase", "get_user_info", false},
		{"withSpace", "abc 123", true},
		{"withDash", "abc-123", true},
		{"withSlash", "abc/123", true},
		{"withDot", "abc.123", true},
		{"withSpecialChars", "abc@123", true},
		{"emptyString", "", true},
		{"chineseCharacters", "中文", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Value: tt.value}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err, "Should return validation error for value: %q", tt.value)
			} else {
				assert.NoError(t, err, "Should not return validation error for value: %q", tt.value)
			}
		})
	}
}

func TestAlphanumUsSlash(t *testing.T) {
	type testStruct struct {
		Value string `validate:"alphanum_us_slash"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"lowercaseLetters", "abc", false},
		{"uppercaseLetters", "ABC", false},
		{"numbers", "123", false},
		{"alphanumeric", "abc123", false},
		{"withUnderscores", "abc_123", false},
		{"withSlashes", "abc/123", false},
		{"resourcePath", "sys/user", false},
		{"nestedPath", "auth/get_user_info", false},
		{"multipleSlashes", "a/b/c/d", false},
		{"leadingSlash", "/abc", false},
		{"trailingSlash", "abc/", false},
		{"mixed", "sys_module/user_info", false},
		{"withSpace", "abc 123", true},
		{"withDash", "abc-123", true},
		{"withDot", "abc.123", true},
		{"withSpecialChars", "abc@123", true},
		{"emptyString", "", true},
		{"chineseCharacters", "中文", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Value: tt.value}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err, "Should return validation error for value: %q", tt.value)
			} else {
				assert.NoError(t, err, "Should not return validation error for value: %q", tt.value)
			}
		})
	}
}

func TestAlphanumUsDot(t *testing.T) {
	type testStruct struct {
		Value string `validate:"alphanum_us_dot"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"lowercaseLetters", "abc", false},
		{"uppercaseLetters", "ABC", false},
		{"numbers", "123", false},
		{"alphanumeric", "abc123", false},
		{"withUnderscores", "abc_123", false},
		{"withDots", "abc.123", false},
		{"fileName", "config.yaml", false},
		{"moduleName", "com.example.app", false},
		{"versionNumber", "v1.2.3", false},
		{"multipleDots", "a.b.c.d", false},
		{"leadingDot", ".hidden", false},
		{"trailingDot", "file.", false},
		{"mixed", "app_config.prod.yaml", false},
		{"withSpace", "abc 123", true},
		{"withDash", "abc-123", true},
		{"withSlash", "abc/123", true},
		{"withSpecialChars", "abc@123", true},
		{"emptyString", "", true},
		{"chineseCharacters", "中文", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Value: tt.value}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err, "Should return validation error for value: %q", tt.value)
			} else {
				assert.NoError(t, err, "Should not return validation error for value: %q", tt.value)
			}
		})
	}
}

func TestAlphanumRulesCombined(t *testing.T) {
	type testStruct struct {
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
			name:     "allValid",
			action:   "get_user_info",
			resource: "sys/user",
			fileName: "config.yaml",
			wantErr:  false,
		},
		{
			name:     "invalidActionWithSlash",
			action:   "get/user",
			resource: "sys/user",
			fileName: "config.yaml",
			wantErr:  true,
		},
		{
			name:     "invalidResourceWithDot",
			action:   "get_user",
			resource: "sys.user",
			fileName: "config.yaml",
			wantErr:  true,
		},
		{
			name:     "invalidFilenameWithSlash",
			action:   "get_user",
			resource: "sys/user",
			fileName: "config/app.yaml",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{
				Action:   tt.action,
				Resource: tt.resource,
				FileName: tt.fileName,
			}

			err := Validate(&s)
			if tt.wantErr {
				assert.Error(t, err, "Should return validation error")
			} else {
				assert.NoError(t, err, "Should not return validation error")
			}
		})
	}
}
